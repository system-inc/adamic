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
func TestWasiReleaseSections(test *testing.T) {
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		test.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	for _, count := range []bool{false, true} {
		output := filepath.Join(test.TempDir(), "hello.wasm")
		arguments := []string{
			"build",
			"--target",
			"wasm32-wasi",
			"../../internal/load/testdata/0.1/compile/01_hello.ts",
			"-o",
			output,
		}
		if count {
			arguments = append(arguments, "--count")
		}
		if code := run(arguments); code != 0 {
			test.Fatalf("build (count=%t): %d", count, code)
		}
		module, readError := os.ReadFile(output)
		if readError != nil {
			test.Fatal(readError)
		}
		if len(module) < 8 || !bytes.Equal(module[:8], []byte{0, 'a', 's', 'm', 1, 0, 0, 0}) {
			test.Fatal("invalid wasm header")
		}
		reader := bytes.NewReader(module[8:])
		var sections []string
		for reader.Len() > 0 {
			id, readByteError := reader.ReadByte()
			if readByteError != nil {
				test.Fatal(readByteError)
			}
			size, readUvarintError := binary.ReadUvarint(reader)
			if readUvarintError != nil || size > uint64(reader.Len()) {
				test.Fatalf("invalid section size: %d, %v", size, readUvarintError)
			}
			payload := make([]byte, int(size))
			if _, readError := io.ReadFull(reader, payload); readError != nil {
				test.Fatal(readError)
			}
			if id != 0 {
				continue
			}
			custom := bytes.NewReader(payload)
			size, readUvarintError = binary.ReadUvarint(custom)
			if readUvarintError != nil || size > uint64(custom.Len()) {
				test.Fatalf("invalid custom section name: %d, %v", size, readUvarintError)
			}
			name := make([]byte, int(size))
			if _, readError := io.ReadFull(custom, name); readError != nil {
				test.Fatal(readError)
			}
			sections = append(sections, string(name))
		}
		test.Logf("count=%t custom sections: %v; bytes=%d", count, sections, len(module))
		hasName := false
		for _, name := range sections {
			if strings.HasPrefix(name, ".debug_") {
				test.Errorf("count=%t: debug section %s remains", count, name)
			}
			hasName = hasName || name == "name"
		}
		if !hasName {
			test.Errorf("count=%t: name section missing", count)
		}
	}
}
