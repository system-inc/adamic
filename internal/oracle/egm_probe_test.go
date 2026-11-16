package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// Throwaway probe for #egmxf20: which roads still put a lying tag into a union's tag slot.
func TestEgmProbe(t *testing.T) {
	directory, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/egm_probe"))
	if err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	for _, name := range names {
		path := filepath.Join(directory, name)
		short := func(b []byte) string {
			s := strings.ReplaceAll(string(b), "\n", "|")
			if len(s) > 110 {
				s = s[:110]
			}
			return s
		}
		program, err := lowered(t, path)
		if err != nil {
			message := err.Error()
			if index := strings.Index(message, name); index >= 0 {
				message = message[index:]
			}
			if len(message) > 150 {
				message = message[:150]
			}
			fmt.Printf("PROBE %s STOP %s\n", name, message)
			continue
		}
		node := onNode(t, path)
		js := onJavaScriptBackend(t, program)
		native, _ := natively(t, program)
		fmt.Printf("PROBE %s LOWERS node=%d:%s js=%d:%s native=%d:%s err=%s\n", name, node.exitCode, short(node.stdout), js.exitCode, short(js.stdout), native.exitCode, short(native.stdout), short(native.stderr))
	}
}
