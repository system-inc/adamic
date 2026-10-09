package native

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestBuildProductReproducible(t *testing.T) {
	const source = "#include \"adamic.h\"\nint main(int argc, char **argv) { (void)argv; volatile int value = 2147483647; if (argc > 1) return value + 1; return 0; }\n"
	var previous []byte
	for i := 0; i < 2; i++ {
		directory := t.TempDir()
		if err := BuildProduct(source, directory, Options{Sanitize: true}); err != nil {
			t.Fatal(err)
		}
		binary := filepath.Join(directory, "port")
		data, err := os.ReadFile(binary)
		if err != nil {
			t.Fatal(err)
		}
		if i > 0 && !bytes.Equal(previous, data) {
			t.Fatal("different product directories produced different binary bytes")
		}
		previous = data
		command := exec.Command(binary)
		command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0")
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("sanitized product: %v\n%s", err, output)
		}
		command = exec.Command(binary, "overflow")
		command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0")
		if output, err := command.CombinedOutput(); err == nil || !bytes.Contains(output, []byte("signed integer overflow")) {
			t.Fatalf("undefined-behavior sanitizer missing: %v\n%s", err, output)
		}
		entries, err := os.ReadDir(directory)
		if err != nil || len(entries) != 1 || entries[0].Name() != "port" {
			t.Fatalf("product contains build scratch: %v, %v", entries, err)
		}
	}
}
