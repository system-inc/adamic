package lane

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
)

func TestProduct_Stage3LaneInstall(t *testing.T)  { t.Parallel(); laneProduct(t, 0) }
func TestProduct_Stage3LaneCompiler(t *testing.T) { t.Parallel(); laneProduct(t, 1) }
func TestProduct_Stage3LaneHarness(t *testing.T)  { t.Parallel(); laneProduct(t, 2) }

func laneTracksInputs(t *testing.T, index int) {
	t.Helper()
	root := repositoryRoot(t)
	inputs := laneInputs(t, root, index)
	scratch := t.TempDir()
	for _, name := range inputs.Files {
		source := filepath.Join(root, name)
		info, err := os.Lstat(source)
		if err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(scratch, name)
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			t.Fatal(err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			link, err := os.Readlink(source)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(link, target); err != nil {
				t.Fatal(err)
			}
		} else {
			data, err := os.ReadFile(source)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(target, data, info.Mode().Perm()); err != nil {
				t.Fatal(err)
			}
		}
	}
	before, err := buildcache.Key(scratch, inputs)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := buildcache.Key(root, inputs)
	if err != nil {
		t.Fatal(err)
	}
	if before != actual {
		t.Fatal("checkout location changed lane product key")
	}
	target := filepath.Join(scratch, "stage3/lane/products.py")
	if err := os.WriteFile(target, []byte("# changed lane recipe\n"), 0644); err != nil {
		t.Fatal(err)
	}
	after, err := buildcache.Key(scratch, inputs)
	if err != nil {
		t.Fatal(err)
	}
	if after == before {
		t.Fatal("changed lane worker reused product key")
	}
}
func TestStage3LaneInstallTracksInputs(t *testing.T)  { t.Parallel(); laneTracksInputs(t, 0) }
func TestStage3LaneCompilerTracksInputs(t *testing.T) { t.Parallel(); laneTracksInputs(t, 1) }
func TestStage3LaneHarnessTracksInputs(t *testing.T)  { t.Parallel(); laneTracksInputs(t, 2) }

type counts struct {
	Pass int `json:"pass"`
	Fail int `json:"fail"`
	Skip int `json:"skip"`
}
type unit struct {
	Name    string   `json:"name"`
	Command []string `json:"command"`
	Expect  counts   `json:"expect"`
}

func manifestUnit(path, name string) (unit, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return unit{}, err
	}
	var plan struct {
		Version int    `json:"version"`
		Units   []unit `json:"units"`
	}
	if err := json.Unmarshal(data, &plan); err != nil {
		return unit{}, err
	}
	if plan.Version != 1 || len(plan.Units) == 0 {
		return unit{}, fmt.Errorf("missing version 1 units")
	}
	var found []unit
	for _, row := range plan.Units {
		if row.Name == name {
			found = append(found, row)
		}
	}
	if len(found) != 1 || len(found[0].Command) == 0 {
		return unit{}, fmt.Errorf("want exactly one command for %s", name)
	}
	return found[0], nil
}

func runShard(t testing.TB, artifact, name, manifest string) {
	t.Helper()
	root := repositoryRoot(t)
	selected, err := manifestUnit(manifest, name)
	if err != nil {
		t.Fatal(err)
	}
	results := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Second)
	defer cancel()
	// The entire argv, including interpreter, is read from the manifest.
	command := boundedCommand(ctx, root, selected.Command[0], selected.Command[1:]...)
	command.Env = append(os.Environ(), "STAGE3_ARTIFACT="+artifact, "STAGE3_RESULTS="+results, "PYTHONDONTWRITEBYTECODE=1", "PYTHONPYCACHEPREFIX="+filepath.Join(results, "pycache"))
	output, err := command.CombinedOutput()
	t.Logf("%s", output)
	if err != nil {
		t.Fatalf("manifest command: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(results, name, "report.json"))
	if err != nil {
		t.Fatal(err)
	}
	var report struct {
		Status string `json:"status"`
		Counts counts `json:"counts"`
	}
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if report.Status != "pass" || report.Counts != selected.Expect {
		t.Fatalf("result %+v differs from manifest %+v", report, selected.Expect)
	}
}

func TestShard_Stage3LaneFile001(t *testing.T) {
	t.Parallel()
	runShard(t, laneProduct(t, 2), "stage3-lane-file-001", filepath.Join(repositoryRoot(t), "stage3/lane/shards.json"))
}

func TestShardManifestRejectsMissingAndEmpty(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "shards.json")
	if _, err := manifestUnit(path, "stage3-lane-file-001"); err == nil {
		t.Fatal("missing manifest passed")
	}
	if err := os.WriteFile(path, []byte(`{"version":1,"units":[]}`), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := manifestUnit(path, "stage3-lane-file-001"); err == nil {
		t.Fatal("empty manifest passed")
	}
}
