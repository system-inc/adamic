package estree

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInterfaceTypeMethodGap(t *testing.T) {
	main, _ := filepath.Abs("gaps/interfaceTypeMethod.ts")
	binary, script := build(t, main, true)
	for name, got := range map[string][]byte{"Node": onNode(t, main), "emitted": onNode(t, script)} {
		if string(got) != "1\n" {
			t.Fatalf("%s: %s", name, got)
		}
	}
	file, err := os.CreateTemp(t.TempDir(), "native-output")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(binary)
	cmd.Stdout = file
	cmd.Stderr = file
	err = cmd.Run()
	file.Close()
	data, _ := os.ReadFile(file.Name())
	if err == nil || !strings.Contains(string(data), "compiler bug: a method the checker proved is there is missing") {
		t.Fatalf("native: %v %s", err, data)
	}
	t.Log("Node and emitted JS print 1; sanitized native explicitly panics on the interface call with both arguments supplied")
	source, err := os.ReadFile(main)
	if err != nil {
		t.Fatal(err)
	}
	control := filepath.Join(t.TempDir(), "control.ts")
	os.WriteFile(control, []byte(strings.Replace(string(source), "type(minimum = 0, conditional = true)", "type(minimum: number, conditional: boolean)", 1)), 0644)
	controlBinary, controlScript := build(t, control, true)
	for name, got := range map[string][]byte{"Node": onNode(t, control), "emitted": onNode(t, controlScript), "native": execute(t, "", controlBinary)} {
		if string(got) != "1\n" {
			t.Fatalf("%s default-free control: %s", name, got)
		}
	}
	t.Log("removing concrete default parameters yields 1 in all builds; the expected-panic check catches that control")
}
