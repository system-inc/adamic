package worker

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestWasmPurityRefusesHostOperations(test *testing.T) {
	test.Parallel()
	builds := 0
	buildWithError := BuildWith("testdata/wasm/host_cwd.a", filepath.Join(test.TempDir(), "worker"), Options{
		Wasm: []string{"hostDirectory"},
		BuildWasm: func(
			string,
			string,
			string,
			[]string,
		) error {
			builds++
			return fmt.Errorf("unexpected Wasm build")
		},
	})
	for _, want := range []string{"host_cwd.a:3:", "hostDirectory", "impure: host operation host_cwd", "fix:"} {
		if buildWithError == nil || !strings.Contains(buildWithError.Error(), want) || builds != 0 {
			test.Fatalf(
				"expected located host refusal before Wasm build (%s), builds=%d: %v",
				want,
				builds,
				buildWithError,
			)
		}
	}
}

func TestWasmPurityRefusesRealpathOperations(test *testing.T) {
	test.Parallel()
	builds := 0
	buildWithError := BuildWith("testdata/wasm/host_realpath.a", filepath.Join(test.TempDir(), "worker"), Options{
		Wasm: []string{"hostPath"},
		BuildWasm: func(
			string,
			string,
			string,
			[]string,
		) error {
			builds++
			return fmt.Errorf("unexpected Wasm build")
		},
	})
	for _, want := range []string{"host_realpath.a:3:", "hostPath", "impure: host operation host_realpath", "fix:"} {
		if buildWithError == nil || !strings.Contains(buildWithError.Error(), want) || builds != 0 {
			test.Fatalf(
				"expected located host refusal before Wasm build (%s), builds=%d: %v",
				want,
				builds,
				buildWithError,
			)
		}
	}
}

func TestWasmPurityRefusesStdoutWrites(test *testing.T) {
	test.Parallel()
	builds := 0
	buildWithError := BuildWith("testdata/wasm/host_stdout_write.a", filepath.Join(test.TempDir(), "worker"), Options{
		Wasm: []string{"hostWrite"},
		BuildWasm: func(
			string,
			string,
			string,
			[]string,
		) error {
			builds++
			return fmt.Errorf("unexpected Wasm build")
		},
	})
	for _, want := range []string{"hostWrite", "impure: host operation host_stdout_write", "fix:"} {
		if buildWithError == nil || !strings.Contains(buildWithError.Error(), want) || builds != 0 {
			test.Fatalf("want %s before build, got %v", want, buildWithError)
		}
	}
}

func TestWasmPurityRefusesStdoutTTYReads(test *testing.T) {
	test.Parallel()
	builds := 0
	buildWithError := BuildWith("testdata/wasm/host_stdout_istty.a", filepath.Join(test.TempDir(), "worker"), Options{
		Wasm: []string{"hostTerminal"},
		BuildWasm: func(
			string,
			string,
			string,
			[]string,
		) error {
			builds++
			return fmt.Errorf("unexpected Wasm build")
		},
	})
	for _, want := range []string{"hostTerminal", "impure: host operation host_stdout_is_tty", "fix:"} {
		if buildWithError == nil || !strings.Contains(buildWithError.Error(), want) || builds != 0 {
			test.Fatalf("want %s before build, got %v", want, buildWithError)
		}
	}
}

func TestWasmPurityRefusesStdoutColumnsReads(test *testing.T) {
	test.Parallel()
	builds := 0
	buildWithError := BuildWith("testdata/wasm/host_stdout_columns.a", filepath.Join(test.TempDir(), "worker"), Options{
		Wasm: []string{"hostTerminal"},
		BuildWasm: func(
			string,
			string,
			string,
			[]string,
		) error {
			builds++
			return fmt.Errorf("unexpected Wasm build")
		},
	})
	for _, want := range []string{"hostTerminal", "impure: host operation host_stdout_columns", "fix:"} {
		if buildWithError == nil || !strings.Contains(buildWithError.Error(), want) || builds != 0 {
			test.Fatalf("want %s before build, got %v", want, buildWithError)
		}
	}
}

func TestWasmPurityRefusesStatPolling(test *testing.T) {
	test.Parallel()
	for _, operation := range []string{"watch_file", "unwatch_file"} {
		builds := 0
		buildWithError := BuildWith(
			"testdata/wasm/host_"+operation+".a",
			filepath.Join(test.TempDir(), "worker"),
			Options{
				Wasm: []string{"hostPolling"},
				BuildWasm: func(
					string,
					string,
					string,
					[]string,
				) error {
					builds++
					return fmt.Errorf("unexpected Wasm build")
				},
			},
		)
		if buildWithError == nil ||
			builds != 0 ||
			!strings.Contains(buildWithError.Error(), "impure: host operation host_"+operation) {
			test.Fatalf("want located host polling refusal before build, got %v", buildWithError)
		}
	}
}
