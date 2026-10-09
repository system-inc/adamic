package native

import (
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Not parallel: expands a large byte oracle corpus and runs complete runtime builds serially.
func TestDecodeASCII(t *testing.T) {
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	run := func(t *testing.T, name string, args ...string) []byte {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
		defer cancel()
		command := exec.CommandContext(ctx, name, args...)
		command.Dir = repository
		output, err := command.CombinedOutput()
		if err != nil {
			t.Fatalf("%s %v: %v\n%s", name, args, err, output)
		}
		return output
	}
	cache, err := newTestBuildCache(repository, "internal/native/decode_ascii")
	if err != nil {
		t.Fatal(err)
	}
	corpusDirectory, err := cache.Tree("decode-node-corpus", nil, func(destination string) error {
		t.Log(string(run(t, "node", "internal/native/decode_ascii/corpus.mjs", filepath.Join(destination, "corpus.bin"))))
		data, err := os.ReadFile(filepath.Join(destination, "corpus.bin"))
		if err != nil {
			return err
		}
		offsets := []int{0}
		for offset := 0; offset < len(data); {
			if offset+4 > len(data) {
				return fmt.Errorf("short decoder corpus header")
			}
			length := int(binary.LittleEndian.Uint16(data[offset:])) + int(binary.LittleEndian.Uint16(data[offset+2:]))
			offset += 4 + length
			if offset > len(data) {
				return fmt.Errorf("short decoder corpus record")
			}
			offsets = append(offsets, offset)
		}
		if len(offsets)-1 != decodePrefixes {
			return fmt.Errorf("decoder prefixes: got %d, want %d", len(offsets)-1, decodePrefixes)
		}
		for _, piece := range unitRanges(decodePrefixes, decodePrefixUnitSize) {
			path := filepath.Join(destination, piece.name()+".bin")
			if err := os.WriteFile(path, data[offsets[piece.first]:offsets[piece.last]], 0644); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	runtimeDirectory := filepath.Join(repository, "internal/native/runtime")
	if alternate := os.Getenv("ADAMIC_DECODE_RUNTIME"); alternate != "" {
		runtimeDirectory = alternate
	}
	sources, err := filepath.Glob(filepath.Join(runtimeDirectory, "*.c"))
	if err != nil {
		t.Fatal(err)
	}
	if len(sources) == 0 {
		t.Fatal("no runtime sources")
	}
	shard := currentTestShard(t)
	pieces := unitRanges(decodePrefixes, decodePrefixUnitSize)
	for targetIndex, target := range decodeTargets {
		hasPiece := false
		for index := range pieces {
			hasPiece = hasPiece || shard.owns(targetIndex*len(pieces)+index)
		}
		if !hasPiece {
			continue
		}
		t.Run(target, func(t *testing.T) {
			compiler := "clang"
			flags := []string{"-std=c11", "-Wall", "-Wextra", "-Werror", "-pedantic", "-O2", "-ffp-contract=off", "-fno-optimize-sibling-calls", "-I", runtimeDirectory}
			if target == "wasi" {
				if os.Getenv("ADAMIC_TEST_WASI") != "1" {
					t.Skip("set ADAMIC_TEST_WASI=1")
				}
				sysroot := os.Getenv("WASI_SYSROOT")
				if sysroot == "" {
					t.Fatal("WASI_SYSROOT missing")
				}
				compiler = filepath.Join(filepath.Dir(filepath.Dir(sysroot)), "bin/clang")
				flags = append(flags, "--target=wasm32-wasi", "--sysroot="+sysroot, "-DADAMIC_TARGET_WASI=1", "-Wl,-z,stack-size=1048576")
			} else {
				flags = append(flags, "-g", "-fsanitize=address,undefined", "-fno-sanitize-recover=all")
			}
			binary := filepath.Join(directory, "decode-"+target)
			args := append(flags, "internal/native/decode_ascii/probe.c", "internal/native/decode_ascii/baseline.c")
			args = append(args, sources...)
			args = append(args, "-lm", "-o", binary)
			command := exec.Command(compiler, args...)
			command.Dir = repository
			snapshot, err := readRuntime(os.DirFS(runtimeDirectory), ".")
			if err != nil {
				t.Fatal(err)
			}
			var contents [][]byte
			for _, file := range snapshot {
				contents = append(contents, []byte(file.name), file.contents)
			}
			if output, err := cache.Command(command, directory, contents...); err != nil {
				t.Fatalf("decoder build: %v\n%s", err, output)
			}
			for index, piece := range pieces {
				if !shard.owns(targetIndex*len(pieces) + index) {
					continue
				}
				t.Run(piece.name(), func(t *testing.T) {
					corpus := filepath.Join(corpusDirectory, piece.name()+".bin")
					var output []byte
					if target == "wasi" {
						output = run(t, "node", "--disable-warning=ExperimentalWarning", "oracle/wasi.mjs", binary, corpus)
					} else {
						output = run(t, binary, corpus)
					}
					want := fmt.Sprintf("prefixes=%d cases=%d Node and baseline identical", piece.last-piece.first, (piece.last-piece.first)*65*8*2)
					if !strings.Contains(string(output), want) {
						t.Fatalf("missing completed decoder partition %s: %s", piece.name(), output)
					}
					t.Log(string(output))
				})
			}
		})
	}
}
