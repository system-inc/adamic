package lint

import (
	"bytes"
	"os"
	"runtime"
	"testing"
)

// This harness selects split compilation explicitly. An inherited environment
// switch must never convert the independent whole-file witness into split.
func init() { os.Unsetenv("ADAMIC_NATIVE_SPLIT") }

func TestSplitWholeParity(t *testing.T) {
	for _, slug := range []string{"no-var", "no-empty", "eqeqeq"} {
		t.Run(slug, func(t *testing.T) {
			d := selectedDescriptor(t, slug)
			directory := selectedPort(t, d)
			rows, _ := ruleRows(t, d)
			for _, row := range stableRows(t, lintCapture(t, ".", d)) {
				if !bytes.HasSuffix([]byte(row), []byte("\tunsupported-recovery")) {
					rows = append(rows, row)
				}
			}
			path := manifest(t, rows)
			oracle := goOracleFrom(t, directory)
			split := buildPort(t, directory, true)
			whole := buildPortUncached(t, directory, true)
			want := execute(t, "", whole, "--manifest", path).output
			got := compareWithJavaScript(t, oracle, split, directory, path, emittedJavaScript(t, directory))
			if !bytes.Equal(got, want) {
				t.Fatal("split and whole-file outputs differ")
			}
			t.Logf("split and whole-file outputs byte-identical: %d bytes", len(want))
		})
	}
}

func TestWholeBuildSelection(t *testing.T) {
	if os.Getenv("ADAMIC_NATIVE_SPLIT") == "1" {
		t.Fatal("inherited split switch can replace the whole-file witness")
	}
	d := selectedDescriptor(t, "no-var")
	directory := selectedPort(t, d)
	t.Setenv("ADAMIC_GATE_UNCACHED", "0")
	if options := lintPortOptions(directory, true); !options.Split || options.Jobs != runtime.GOMAXPROCS(0) || !options.Sanitize {
		t.Fatalf("selected build options: %+v", options)
	}
	if options := lintPortOptions(".", true); options.Split || !options.Sanitize {
		t.Fatalf("full build options: %+v", options)
	}
	t.Setenv("ADAMIC_GATE_UNCACHED", "1")
	if options := lintPortOptions(directory, true); options.Split || !options.Sanitize {
		t.Fatalf("uncached build options: %+v", options)
	}
}
