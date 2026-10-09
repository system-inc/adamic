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

// Each gate-discoverable leaf runs one complete prefix range. Build products are
// shared by input hashes; comparisons and sanitizer execution always run.
func runDecodeASCIIUnit(t *testing.T, target string, index int) {
	t.Helper()
	defer checkGrainBudget(t)()
	if target == "wasi" && os.Getenv("ADAMIC_TEST_WASI") != "1" {
		t.Skip("set ADAMIC_TEST_WASI=1")
	}
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	run := func(t *testing.T, name string, args ...string) []byte {
		t.Helper()
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
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
	piece := unitRanges(decodePrefixes, decodePrefixUnitSize)[index]
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
}

func TestDecodeASCIIUnit00(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 0)
}

func TestDecodeASCIIUnit01(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 1)
}

func TestDecodeASCIIUnit02(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 2)
}

func TestDecodeASCIIUnit03(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 3)
}

func TestDecodeASCIIUnit04(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 4)
}

func TestDecodeASCIIUnit05(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 5)
}

func TestDecodeASCIIUnit06(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 6)
}

func TestDecodeASCIIUnit07(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 7)
}

func TestDecodeASCIIUnit08(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 8)
}

func TestDecodeASCIIUnit09(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 9)
}

func TestDecodeASCIIUnit10(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 10)
}

func TestDecodeASCIIUnit11(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 11)
}

func TestDecodeASCIIUnit12(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 12)
}

func TestDecodeASCIIUnit13(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 13)
}

func TestDecodeASCIIUnit14(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 14)
}

func TestDecodeASCIIUnit15(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 15)
}

func TestDecodeASCIIUnit16(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 16)
}

func TestDecodeASCIIUnit17(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 17)
}

func TestDecodeASCIIUnit18(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 18)
}

func TestDecodeASCIIUnit19(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 19)
}

func TestDecodeASCIIUnit20(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 20)
}

func TestDecodeASCIIUnit21(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 21)
}

func TestDecodeASCIIUnit22(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 22)
}

func TestDecodeASCIIUnit23(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 23)
}

func TestDecodeASCIIUnit24(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 24)
}

func TestDecodeASCIIUnit25(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 25)
}

func TestDecodeASCIIUnit26(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 26)
}

func TestDecodeASCIIUnit27(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 27)
}

func TestDecodeASCIIUnit28(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 28)
}

func TestDecodeASCIIUnit29(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 29)
}

func TestDecodeASCIIUnit30(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 30)
}

func TestDecodeASCIIUnit31(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 31)
}

func TestDecodeASCIIUnit32(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 32)
}

func TestDecodeASCIIUnit33(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 33)
}

func TestDecodeASCIIUnit34(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 34)
}

func TestDecodeASCIIUnit35(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 35)
}

func TestDecodeASCIIUnit36(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 36)
}

func TestDecodeASCIIUnit37(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 37)
}

func TestDecodeASCIIUnit38(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 38)
}

func TestDecodeASCIIUnit39(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 39)
}

func TestDecodeASCIIUnit40(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 40)
}

func TestDecodeASCIIUnit41(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 41)
}

func TestDecodeASCIIUnit42(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 42)
}

func TestDecodeASCIIUnit43(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 43)
}

func TestDecodeASCIIUnit44(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "native", 44)
}

func TestDecodeASCIIWASIUnit00(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 0)
}

func TestDecodeASCIIWASIUnit01(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 1)
}

func TestDecodeASCIIWASIUnit02(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 2)
}

func TestDecodeASCIIWASIUnit03(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 3)
}

func TestDecodeASCIIWASIUnit04(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 4)
}

func TestDecodeASCIIWASIUnit05(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 5)
}

func TestDecodeASCIIWASIUnit06(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 6)
}

func TestDecodeASCIIWASIUnit07(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 7)
}

func TestDecodeASCIIWASIUnit08(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 8)
}

func TestDecodeASCIIWASIUnit09(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 9)
}

func TestDecodeASCIIWASIUnit10(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 10)
}

func TestDecodeASCIIWASIUnit11(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 11)
}

func TestDecodeASCIIWASIUnit12(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 12)
}

func TestDecodeASCIIWASIUnit13(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 13)
}

func TestDecodeASCIIWASIUnit14(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 14)
}

func TestDecodeASCIIWASIUnit15(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 15)
}

func TestDecodeASCIIWASIUnit16(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 16)
}

func TestDecodeASCIIWASIUnit17(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 17)
}

func TestDecodeASCIIWASIUnit18(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 18)
}

func TestDecodeASCIIWASIUnit19(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 19)
}

func TestDecodeASCIIWASIUnit20(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 20)
}

func TestDecodeASCIIWASIUnit21(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 21)
}

func TestDecodeASCIIWASIUnit22(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 22)
}

func TestDecodeASCIIWASIUnit23(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 23)
}

func TestDecodeASCIIWASIUnit24(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 24)
}

func TestDecodeASCIIWASIUnit25(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 25)
}

func TestDecodeASCIIWASIUnit26(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 26)
}

func TestDecodeASCIIWASIUnit27(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 27)
}

func TestDecodeASCIIWASIUnit28(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 28)
}

func TestDecodeASCIIWASIUnit29(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 29)
}

func TestDecodeASCIIWASIUnit30(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 30)
}

func TestDecodeASCIIWASIUnit31(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 31)
}

func TestDecodeASCIIWASIUnit32(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 32)
}

func TestDecodeASCIIWASIUnit33(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 33)
}

func TestDecodeASCIIWASIUnit34(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 34)
}

func TestDecodeASCIIWASIUnit35(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 35)
}

func TestDecodeASCIIWASIUnit36(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 36)
}

func TestDecodeASCIIWASIUnit37(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 37)
}

func TestDecodeASCIIWASIUnit38(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 38)
}

func TestDecodeASCIIWASIUnit39(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 39)
}

func TestDecodeASCIIWASIUnit40(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 40)
}

func TestDecodeASCIIWASIUnit41(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 41)
}

func TestDecodeASCIIWASIUnit42(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 42)
}

func TestDecodeASCIIWASIUnit43(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 43)
}

func TestDecodeASCIIWASIUnit44(t *testing.T) {
	t.Parallel()
	runDecodeASCIIUnit(t, "wasi", 44)
}
