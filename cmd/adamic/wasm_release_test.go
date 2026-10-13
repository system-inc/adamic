package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Not parallel: calls the command's shared checker and diagnostics.
func TestWASIReleaseSections(t *testing.T) {
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	for _, count := range []bool{false, true} {
		output := filepath.Join(t.TempDir(), "hello.wasm")
		arguments := []string{"build", "--target", "wasm32-wasi", "../../internal/load/testdata/0.1/compile/01_hello.ts", "-o", output}
		if count {
			arguments = append(arguments, "--count")
		}
		if code := run(arguments); code != 0 {
			t.Fatalf("build (count=%t): %d", count, code)
		}
		module, err := os.ReadFile(output)
		if err != nil {
			t.Fatal(err)
		}
		if len(module) < 8 || !bytes.Equal(module[:8], []byte{0, 'a', 's', 'm', 1, 0, 0, 0}) {
			t.Fatal("invalid wasm header")
		}
		reader := bytes.NewReader(module[8:])
		var sections []string
		for reader.Len() > 0 {
			id, err := reader.ReadByte()
			if err != nil {
				t.Fatal(err)
			}
			size, err := binary.ReadUvarint(reader)
			if err != nil || size > uint64(reader.Len()) {
				t.Fatalf("invalid section size: %d, %v", size, err)
			}
			payload := make([]byte, int(size))
			if _, err := io.ReadFull(reader, payload); err != nil {
				t.Fatal(err)
			}
			if id != 0 {
				continue
			}
			custom := bytes.NewReader(payload)
			size, err = binary.ReadUvarint(custom)
			if err != nil || size > uint64(custom.Len()) {
				t.Fatalf("invalid custom section name: %d, %v", size, err)
			}
			name := make([]byte, int(size))
			if _, err := io.ReadFull(custom, name); err != nil {
				t.Fatal(err)
			}
			sections = append(sections, string(name))
		}
		t.Logf("count=%t custom sections: %v; bytes=%d", count, sections, len(module))
		hasName := false
		for _, name := range sections {
			if strings.HasPrefix(name, ".debug_") {
				t.Errorf("count=%t: debug section %s remains", count, name)
			}
			hasName = hasName || name == "name"
		}
		if !hasName {
			t.Errorf("count=%t: name section missing", count)
		}
	}
}
