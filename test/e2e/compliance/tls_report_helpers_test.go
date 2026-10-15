package test

import (
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
		{name: "recreated report stale timestamp", count: 1, checked: &metav1.Time{Time: after.Add(-time.Second)}, uid: "replacement"},
		{name: "recreated report failed check", count: 1, checked: &metav1.Time{Time: after.Add(time.Second)}, uid: "replacement", lastError: "timeout"},
		{name: "recreated report second check", count: 2, checked: &metav1.Time{Time: after.Add(time.Second)}, uid: "replacement", wantFresh: true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			report := previous
			report.Metadata.UID = scenario.uid
			report.Status = tlsReportStatus{CheckCount: scenario.count, LastCheckAt: scenario.checked, LastError: scenario.lastError}
			err := freshTLSReport(report, []tlsReport{previous}, after)
			if (err == nil) != scenario.wantFresh {
				t.Fatalf("fresh=%t: %v", scenario.wantFresh, err)
			}
			if err != nil && (!strings.Contains(err.Error(), "baselineCheckCount=") || !strings.Contains(err.Error(), "lastCheckAt=")) {
				t.Fatalf("missing freshness diagnostics: %v", err)
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
