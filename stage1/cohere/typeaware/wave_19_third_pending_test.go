package typeaware

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestWave19ThirdProductionPending(t *testing.T) {
	directory := os.Getenv("ADAMIC_WAVE19_THIRD_RELEASE_ARTIFACTS")
	if directory == "" {
		t.Skip("run after TestWave19ThirdReleasedHandles with persistent artifacts")
	}
	repository, e := filepath.Abs("../../..")
	if e != nil {
		t.Fatal(e)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	source, e := os.ReadFile(filepath.Join(directory, "released.a"))
	if e != nil {
		t.Fatal(e)
	}
	entry := h.write("live-production.a", strings.Replace(string(source), "tsgoRelease(program);", "", 1)+"tsgoRelease(program);")
	archive := h.archive("production-checker", "", false)
	binary := h.build(filepath.Join(directory, "adamic"), "production-pending", entry, archive, false)
	questions := []struct{ question, kind, end, start string }{{"wave19-type-signatures\n1", "CallExpression", "21", "14"}, {"wave19-generic-call", "CallExpression", "21", "14"}, {"wave19-type-members\n1\nthen", "CallExpression", "21", "14"}, {"wave19-heritage-members", "MethodDeclaration", "13", "8"}}
	for _, q := range questions {
		got := h.run("production-pending-run", exec.Command(binary, filepath.Join(directory, "tsconfig.json"), filepath.Join(directory, "probe.a"), q.question, q.kind, q.end, q.start))
		exit, ok := got.err.(*exec.ExitError)
		if !ok || exit.ExitCode() != 70 || !strings.Contains(string(got.stderr), "unsupported checker question: "+strings.Split(q.question, "\n")[0]) {
			t.Fatalf("production refused differently: %s", got.stderr)
		}
	}
	t.Log("unmodified production archive refuses all four unregistered questions with panic 70")
}
