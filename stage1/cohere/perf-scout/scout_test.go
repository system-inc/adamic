package perfscout

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// Measurement inputs are mandatory: a missing artifact is a failure, never a skip.
func TestMeasurements(t *testing.T) {
	t.Parallel()
	directory := os.Getenv("ADAMIC_SCOUT_RESULTS")
	if directory == "" {
		t.Fatal("set ADAMIC_SCOUT_RESULTS to a completed scout.py run")
	}
	data, err := os.ReadFile(filepath.Join(directory, "measurements.json"))
	if err != nil {
		t.Fatal(err)
	}
	var report struct {
		Workloads map[string]struct {
			Rows    int
			Bytes   int
			SHA256  string
			Samples map[string][]struct {
				SHA256 string
				Bytes  int
			}
		}
	}
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"owned", "upstream", "compiler", "public", "repository", "printer"} {
		result, ok := report.Workloads[name]
		if !ok || result.Rows == 0 || result.Bytes == 0 {
			t.Fatalf("missing/empty %s", name)
		}
		for _, side := range []string{"Go", "native", "Node"} {
			samples := result.Samples[side]
			if len(samples) < 3 {
				t.Fatalf("%s %s: fewer than three rounds", name, side)
			}
			for _, sample := range samples {
				if sample.SHA256 != result.SHA256 || sample.Bytes != result.Bytes {
					t.Fatalf("%s %s differs", name, side)
				}
			}
		}
	}
}

// Successful wrong-answer mutations must trip the guard, rather than failing to run.
func TestGuardsCatchMutants(t *testing.T) {
	t.Parallel()
	command := exec.Command("python3", "-B", "-m", "unittest", "-v", "test_scout")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("%v\n%s", err, output)
	}
	if !bytes.Contains(output, []byte("OK")) {
		t.Fatal(string(output))
	}
	t.Log(string(output))
}

func TestEncodingAndMutant(t *testing.T) {
	t.Parallel()
	compiler := os.Getenv("ADAMIC_SCOUT_COMPILER")
	if compiler == "" {
		t.Fatal("ADAMIC_SCOUT_COMPILER required")
	}
	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	source, err := os.ReadFile("encode.ts")
	if err != nil {
		t.Fatal(err)
	}
	main, err := os.ReadFile("testdata/encodingMain.ts")
	if err != nil {
		t.Fatal(err)
	}
	want := []byte("\nhello\n\\u000a\n\\u000d\\u0009\n\\u0000\n\\u005c\n\\u007f\n\\u00e9\n\\ud83d\\ude00\n\\ud800\n\\udfff\na\\u005cb\\u000ac\n ~\n")
	for _, mutated := range []bool{false, true} {
		module := source
		if mutated {
			module = bytes.Replace(source, []byte(" || code === 92"), nil, 1)
		}
		if err := os.WriteFile(filepath.Join(directory, "encode.ts"), module, 0644); err != nil {
			t.Fatal(err)
		}
		mainPath := filepath.Join(directory, "main.ts")
		if err := os.WriteFile(mainPath, bytes.Replace(main, []byte("../encode.ts"), []byte("./encode.ts"), 1), 0644); err != nil {
			t.Fatal(err)
		}
		binary := filepath.Join(directory, "native")
		build := exec.Command(compiler, "build", mainPath, "-o", binary, "--sanitize")
		if output, err := build.CombinedOutput(); err != nil {
			t.Fatalf("build: %v\n%s", err, output)
		}
		for _, command := range []*exec.Cmd{exec.Command(binary), exec.Command("node", "--disable-warning=ExperimentalWarning", filepath.Join(root, "oracle/node.mjs"), mainPath)} {
			var diagnostics bytes.Buffer
			command.Stderr = &diagnostics
			output, err := command.Output()
			if err != nil || diagnostics.Len() != 0 {
				t.Fatalf("execution: %v\n%s", err, &diagnostics)
			}
			if mutated == bytes.Equal(output, want) {
				t.Fatalf("mutant=%t: unexpected bytes %q", mutated, output)
			}
		}
	}
}

func TestEncodingMeasurements(t *testing.T) {
	t.Parallel()
	directory := os.Getenv("ADAMIC_SCOUT_RESULTS")
	if directory == "" {
		t.Fatal("ADAMIC_SCOUT_RESULTS required")
	}
	data, err := os.ReadFile(filepath.Join(directory, "encoding-measurements.json"))
	if err != nil {
		t.Fatal(err)
	}
	var report struct {
		Workloads map[string]struct {
			SHA256  string
			Bytes   int
			Samples map[string][]struct {
				SHA256 string
				Bytes  int
			}
		}
		Profile struct{ Instructions int64 }
	}
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"owned", "upstream", "compiler", "public", "repository"} {
		result, ok := report.Workloads[name]
		if !ok || result.Bytes == 0 {
			t.Fatalf("missing %s", name)
		}
		for _, side := range []string{"Go", "before", "after", "Node_before", "Node_after"} {
			samples := result.Samples[side]
			if len(samples) != 5 {
				t.Fatalf("%s %s: expected five samples", name, side)
			}
			for _, sample := range samples {
				if sample.SHA256 != result.SHA256 || sample.Bytes != result.Bytes {
					t.Fatalf("%s %s: bytes differ", name, side)
				}
			}
		}
	}
	if report.Profile.Instructions == 0 {
		t.Fatal("missing optimized instruction profile")
	}
}

func TestFusedMeasurements(t *testing.T) {
	t.Parallel()
	directory := os.Getenv("ADAMIC_SCOUT_RESULTS")
	if directory == "" {
		t.Fatal("ADAMIC_SCOUT_RESULTS required")
	}
	data, err := os.ReadFile(filepath.Join(directory, "fused.json"))
	if err != nil {
		t.Fatal(err)
	}
	var report struct {
		Files, Rules, Bytes int
		Samples             map[string][]struct {
			SHA256 string
			Bytes  int
		}
	}
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if report.Files != 77 || report.Rules != 93 || report.Bytes == 0 {
		t.Fatal("incomplete fused inputs")
	}
	hash := ""
	for _, side := range []string{"Go", "before", "after", "Node_before", "Node_after"} {
		if len(report.Samples[side]) != 5 {
			t.Fatalf("%s: expected five rounds", side)
		}
		for _, sample := range report.Samples[side] {
			if hash == "" {
				hash = sample.SHA256
			}
			if sample.Bytes != report.Bytes || sample.SHA256 != hash {
				t.Fatalf("%s: output differs", side)
			}
		}
	}
}

func TestFusedInstructionProfiles(t *testing.T) {
	t.Parallel()
	directory := os.Getenv("ADAMIC_SCOUT_RESULTS")
	if directory == "" {
		t.Fatal("ADAMIC_SCOUT_RESULTS required")
	}
	data, err := os.ReadFile(filepath.Join(directory, "fused-profile.json"))
	if err != nil {
		t.Fatal(err)
	}
	var profiles map[string]json.RawMessage
	if err := json.Unmarshal(data, &profiles); err != nil {
		t.Fatal(err)
	}
	for _, side := range []string{"Go", "before", "after"} {
		var profile struct {
			Total   int64
			SelfAll map[string]int64 `json:"self_all"`
		}
		if err := json.Unmarshal(profiles[side], &profile); err != nil {
			t.Fatal(err)
		}
		var sum int64
		for _, count := range profile.SelfAll {
			sum += count
		}
		if profile.Total <= 0 || sum != profile.Total {
			t.Fatalf("%s: invalid instruction accounting", side)
		}
	}
}
