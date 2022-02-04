package reporter

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/onsi/ginkgo/config"
	"github.com/onsi/ginkgo/types"
)

type KubernetesCNAOReporter struct {
	artifactsDir         string
	namespace            string
	previousDeviceStatus string
}

func New(artifactsDir string, namespace string) *KubernetesCNAOReporter {
	return &KubernetesCNAOReporter{
		artifactsDir: artifactsDir,
		namespace:    namespace,
	}
}

func (r *KubernetesCNAOReporter) SpecSuiteWillBegin(config config.GinkgoConfigType, summary *types.SuiteSummary) {
}

func (r *KubernetesCNAOReporter) BeforeSuiteDidRun(setupSummary *types.SetupSummary) {
	r.Cleanup()
}

func (r *KubernetesCNAOReporter) SpecWillRun(specSummary *types.SpecSummary) {
	if specSummary.Skipped() || specSummary.Pending() {
		return
	}

	r.storeStateBeforeEach()
}
func (r *KubernetesCNAOReporter) SpecDidComplete(specSummary *types.SpecSummary) {
	if specSummary.Skipped() || specSummary.Pending() {
		return
	}

	since := time.Now().Add(-specSummary.RunTime).Add(-5 * time.Second)
	name := strings.Join(specSummary.ComponentTexts[1:], " ")
	passed := specSummary.Passed()

	r.dumpStateAfterEach(name, since, passed)
}

func (r *KubernetesCNAOReporter) AfterSuiteDidRun(setupSummary *types.SetupSummary) {
}

func (r *KubernetesCNAOReporter) SpecSuiteDidEnd(summary *types.SuiteSummary) {
}

func (r *KubernetesCNAOReporter) storeStateBeforeEach() {
}

func runAndWait(funcs ...func()) {
	var wg sync.WaitGroup
	wg.Add(len(funcs))
	for _, f := range funcs {
		// You have to pass f to the goroutine, it's going to change
		// at the next loop iteration.
		go func(rf func()) {
			rf()
			wg.Done()
		}(f)
	}
	wg.Wait()
}

func (r *KubernetesCNAOReporter) dumpStateAfterEach(testName string, testStartTime time.Time, passed bool) {
	if passed {
		return
	}
	runAndWait(
		func() { r.logPods(testName, testStartTime) },
	)
}

// Cleanup cleans up the current content of the artifactsDir
func (r *KubernetesCNAOReporter) Cleanup() {
	// clean up artifacts from previous run
	if r.artifactsDir != "" {
		_, err := os.Stat(r.artifactsDir)
		if err != nil {
			if os.IsNotExist(err) {
				return
			} else {
				panic(err)
			}
		}
		names, err := ioutil.ReadDir(r.artifactsDir)
		if err != nil {
			panic(err)
		}
		for _, entery := range names {
			os.RemoveAll(path.Join([]string{r.artifactsDir, entery.Name()}...))
		}
	}
}

func (r *KubernetesCNAOReporter) logPods(testName string, sinceTime time.Time) error {

	// Let's print the pods logs to the GinkgoWriter so
	// we see the failure directly at prow junit output without opening files
	r.OpenTestLogFile("pods", testName, podLogsWriter(r.namespace, sinceTime))

	return nil
}

func (r *KubernetesCNAOReporter) OpenTestLogFile(logType string, testName string, cb func(f io.Writer), extraWriters ...io.Writer) {
	err := os.MkdirAll(r.artifactsDir, 0755)
	if err != nil {
		fmt.Println(err)
		return
	}

	name := fmt.Sprintf("%s/%s_%s.log", r.artifactsDir, testName, logType)
	fi, err := os.OpenFile(name, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func() {
		if err := fi.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	writers := []io.Writer{fi}
	if len(extraWriters) > 0 {
		writers = append(writers, extraWriters...)
	}
	cb(io.MultiWriter(writers...))
}
