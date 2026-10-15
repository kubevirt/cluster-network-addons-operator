package test

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestTLSReportFreshness(t *testing.T) {
	t.Parallel()
	after := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	previous := tlsReport{Metadata: metav1.ObjectMeta{UID: "original"}, Spec: tlsReportSpec{Host: "metrics", Port: 8443}, Status: tlsReportStatus{CheckCount: 4}}
	for _, scenario := range []struct {
		name      string
		count     int64
		checked   *metav1.Time
		uid       types.UID
		lastError string
		wantFresh bool
	}{
		{name: "unchanged report", count: 4, checked: &metav1.Time{Time: after.Add(time.Second)}, uid: "original"},
		{name: "in flight pre rollout check", count: 5, checked: &metav1.Time{Time: after.Add(time.Second)}, uid: "original"},
		{name: "no timestamp", count: 6, uid: "original"},
		{name: "old timestamp", count: 6, checked: &metav1.Time{Time: after.Add(-time.Second)}, uid: "original"},
		{name: "timestamp at barrier", count: 6, checked: &metav1.Time{Time: after}, uid: "original"},
		{name: "failed check retains old curves", count: 6, checked: &metav1.Time{Time: after.Add(time.Second)}, uid: "original", lastError: "timeout"},
		{name: "two overlapping pre rollout checks", count: 6, checked: &metav1.Time{Time: after.Add(time.Second)}, uid: "original"},
		{name: "many old report completions", count: 20, checked: &metav1.Time{Time: after.Add(time.Second)}, uid: "original"},
		{name: "recreated report first check", count: 1, checked: &metav1.Time{Time: after.Add(time.Second)}, uid: "replacement", wantFresh: true},
		{name: "recreated report no checks", count: 0, checked: &metav1.Time{Time: after.Add(time.Second)}, uid: "replacement"},
		{name: "recreated report no timestamp", count: 1, uid: "replacement"},
		{name: "new scanner report with clock behind runner", count: 1, checked: &metav1.Time{Time: after.Add(-time.Hour)}, uid: "replacement", wantFresh: true},
		{name: "new scanner report with clock ahead of runner", count: 1, checked: &metav1.Time{Time: after.Add(time.Hour)}, uid: "replacement", wantFresh: true},
		{name: "future timestamp cannot freshen old report", count: 100, checked: &metav1.Time{Time: after.Add(time.Hour)}, uid: "original"},
		{name: "missing report identity", count: 1, checked: &metav1.Time{Time: after}},
		{name: "recreated report failed check", count: 1, checked: &metav1.Time{Time: after.Add(time.Second)}, uid: "replacement", lastError: "timeout"},
		{name: "recreated report second check", count: 2, checked: &metav1.Time{Time: after.Add(time.Second)}, uid: "replacement", wantFresh: true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			report := previous
			report.Metadata.UID = scenario.uid
			report.Status = tlsReportStatus{CheckCount: scenario.count, LastCheckAt: scenario.checked, LastError: scenario.lastError}
			err := freshTLSReport(report, []tlsReport{previous})
			if (err == nil) != scenario.wantFresh {
				t.Fatalf("fresh=%t: %v", scenario.wantFresh, err)
			}
			if err != nil && (!strings.Contains(err.Error(), "baselineCheckCount=") || !strings.Contains(err.Error(), "lastCheckAt=")) {
				t.Fatalf("missing freshness diagnostics: %v", err)
			}
		})
	}
}

func TestTLSScannerPauseAndRestore(t *testing.T) {
	t.Parallel()
	for _, scenario := range []struct {
		name      string
		failAt    int
		output    string
		wantCalls int
	}{
		{name: "drain and restore original replicas", output: `{"spec":{"replicas":2}}`, wantCalls: 6},
		{name: "get failure does not mutate scanner", failAt: 1, wantCalls: 1},
		{name: "decode failure does not mutate scanner", output: `invalid`, wantCalls: 1},
		{name: "stopped scanner is rejected", output: `{"spec":{"replicas":0}}`, wantCalls: 1},
		{name: "failed scale is restored", failAt: 2, output: `{"spec":{"replicas":2}}`, wantCalls: 4},
		{name: "failed zero rollout is restored", failAt: 3, output: `{"spec":{"replicas":2}}`, wantCalls: 5},
		{name: "failed pod drain is restored", failAt: 4, output: `{"spec":{"replicas":2}}`, wantCalls: 6},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			var commands [][]string
			run := func(arguments ...string) (string, string, error) {
				commands = append(commands, arguments)
				if len(commands) == scenario.failAt {
					return "", "injected failure", errors.New("kubectl failed")
				}
				return scenario.output, "", nil
			}
			restore, err := pauseTLSScanner(run)
			assertTLSScannerPauseResult(t, commands, restore, err, scenario.failAt == 0 && scenario.wantCalls > 1)
			assertTLSScannerRestoredCommands(t, commands, scenario.wantCalls)
		})
	}
}

func assertTLSScannerPauseResult(t *testing.T, commands [][]string, restore func() error, pauseError error, successful bool) {
	t.Helper()
	if !successful {
		if pauseError == nil || restore != nil {
			t.Fatalf("expected pause error with handled restoration, got restore=%v err=%v", restore != nil, pauseError)
		}
		return
	}
	if pauseError != nil || restore == nil {
		t.Fatalf("pause: %v", pauseError)
	}
	if err := restore(); err != nil {
		t.Fatalf("restore: %v", err)
	}
	if len(commands) < 4 {
		t.Fatalf("missing scanner pause commands: %v", commands)
	}
	if commands[2][0] != "rollout" || commands[3][0] != "wait" || !strings.Contains(strings.Join(commands[3], " "), "-l app.kubernetes.io/name=tls-compliance-operator,control-plane=controller-manager") {
		t.Fatalf("scanner must observe zero rollout and drain only its Pods: %v", commands)
	}
}

func assertTLSScannerRestoredCommands(t *testing.T, commands [][]string, wantCalls int) {
	t.Helper()
	if len(commands) != wantCalls {
		t.Fatalf("expected %d commands, got %v", wantCalls, commands)
	}
	if wantCalls > 1 && !strings.Contains(strings.Join(commands[len(commands)-2], " "), "--replicas=2") {
		t.Fatalf("original replica count not restored: %v", commands)
	}
	for _, command := range commands {
		if !strings.Contains(strings.Join(command, " "), "--request-timeout=15s") {
			t.Fatalf("unbounded API request: %v", command)
		}
	}
}

func TestTLSReportDeletionScope(t *testing.T) {
	t.Parallel()
	for _, scenario := range []struct {
		name    string
		reports []tlsReport
		wantErr bool
	}{
		{name: "only matched reports", reports: []tlsReport{
			{Metadata: metav1.ObjectMeta{Name: "cnao-metrics"}, Spec: tlsReportSpec{SourceNamespace: "cluster-network-addons"}},
			{Metadata: metav1.ObjectMeta{Name: "ipam-webhook"}, Spec: tlsReportSpec{SourceNamespace: "cluster-network-addons"}},
			{Metadata: metav1.ObjectMeta{Name: "kmp-metrics"}, Spec: tlsReportSpec{SourceNamespace: "cluster-network-addons"}},
			{Metadata: metav1.ObjectMeta{Name: "kmp-webhook"}, Spec: tlsReportSpec{SourceNamespace: "cluster-network-addons"}},
		}},
		{name: "empty list", wantErr: true},
		{name: "missing name", reports: []tlsReport{{Spec: tlsReportSpec{SourceNamespace: "cluster-network-addons"}}}, wantErr: true},
		{name: "foreign namespace", reports: []tlsReport{{Metadata: metav1.ObjectMeta{Name: "foreign"}, Spec: tlsReportSpec{SourceNamespace: "kube-system"}}}, wantErr: true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			var commands [][]string
			err := deleteTLSReports(func(arguments ...string) (string, string, error) {
				commands = append(commands, arguments)
				return "", "", nil
			}, scenario.reports)
			if (err != nil) != scenario.wantErr {
				t.Fatalf("wantErr=%t: %v", scenario.wantErr, err)
			}
			if scenario.wantErr {
				if len(commands) != 0 {
					t.Fatalf("invalid deletion mutated reports: %v", commands)
				}
				return
			}
			want := [][]string{{"delete", "tlscompliancereports", "cnao-metrics", "ipam-webhook", "kmp-metrics", "kmp-webhook", "--ignore-not-found=true", "--timeout=30s", "--request-timeout=15s"}}
			if !reflect.DeepEqual(commands, want) {
				t.Fatalf("deletion must name only the four matched reports: %v", commands)
			}
		})
	}
}

func TestTLSProfileRollout(t *testing.T) {
	t.Parallel()
	replicas := int32(1)
	ready := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Generation: 2},
		Spec:       appsv1.DeploymentSpec{Replicas: &replicas, Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{Args: []string{"--tls-group-preferences=CurveP521"}}}}}},
		Status:     appsv1.DeploymentStatus{ObservedGeneration: 2, Replicas: 1, UpdatedReplicas: 1, AvailableReplicas: 1},
	}
	for _, scenario := range []struct {
		name      string
		mutate    func(*appsv1.Deployment)
		wantReady bool
	}{
		{name: "complete rollout", mutate: func(*appsv1.Deployment) {}, wantReady: true},
		{name: "old profile still available", mutate: func(deployment *appsv1.Deployment) { deployment.Spec.Template.Spec.Containers[0].Args = nil }},
		{name: "generation unobserved", mutate: func(deployment *appsv1.Deployment) { deployment.Status.ObservedGeneration = 1 }},
		{name: "old replica remains", mutate: func(deployment *appsv1.Deployment) { deployment.Status.Replicas = 2 }},
		{name: "replica not updated", mutate: func(deployment *appsv1.Deployment) { deployment.Status.UpdatedReplicas = 0 }},
		{name: "replica unavailable", mutate: func(deployment *appsv1.Deployment) { deployment.Status.AvailableReplicas = 0 }},
		{name: "scaled down", mutate: func(deployment *appsv1.Deployment) { deployment.Spec.Replicas = new(int32) }},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			deployment := ready.DeepCopy()
			scenario.mutate(deployment)
			if err := tlsProfileRolledOut(*deployment, "--tls-group-preferences=CurveP521"); (err == nil) != scenario.wantReady {
				t.Fatalf("ready=%t: %v", scenario.wantReady, err)
			}
		})
	}
}

func TestTLSReportCopiesPreserveStrictExpectations(t *testing.T) {
	t.Parallel()
	status := tlsReportStatus{QuantumReady: true, NegotiatedCurves: map[string]string{"TLS 1.3": "X25519MLKEM768"}, Conditions: []metav1.Condition{{Type: "Available", Status: metav1.ConditionTrue}, {Type: "PQCCompliant", Status: metav1.ConditionTrue}}}
	reports := make([]tlsReport, 4)
	for reportIndex := range reports {
		reports[reportIndex].Status = cloneTLSReportStatus(status)
	}
	reports[3].Status.Conditions[1].Status = metav1.ConditionFalse
	reports[3].Status.NegotiatedCurves["TLS 1.3"] = "CurveP521"
	if reports[0].Status.Conditions[1].Status != metav1.ConditionTrue || status.NegotiatedCurves["TLS 1.3"] != "X25519MLKEM768" {
		t.Fatal("report expectations share mutable data")
	}
	normalized := normalizeConditions(reports)
	normalized[0].Status.Conditions[0].Status = metav1.ConditionFalse
	if reports[0].Status.Conditions[0].Status != metav1.ConditionTrue || !normalized[0].Status.QuantumReady || normalized[3].Status.NegotiatedCurves["TLS 1.3"] != "CurveP521" {
		t.Fatal("normalization aliases conditions or weakens strict TLS values")
	}
}
