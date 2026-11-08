package native

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/toolchain"
)

// decodeAsciiCorpus is Node's decoder corpus cut into the units' prefix ranges, one file per range, a product built
// ahead by TestProduct_DecodeASCIICorpus. Node writes all 92014 prefixes in about 20 s, which every unit used to wait
// on in a cache of its own, under the home directory when ADAMIC_BUILD_CACHE_DIR was unset, the first unit paying
// for it.
func decodeAsciiCorpus(test *testing.T) string {
	test.Helper()
	repository, absolutePathError := filepath.Abs("../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return buildcache.Product(test, buildcache.Inputs{
		Name: "decode-ascii-corpus",
		Files: []string{
			"internal/native/decode_ascii/corpus.mjs",
			"internal/native/decode_ascii_cache_test.go",
			"internal/native/split_units_test.go",
		},
		Flags:     []string{fmt.Sprintf("prefixes=%d", decodePrefixes), fmt.Sprintf("range=%d", decodePrefixUnitSize)},
		Toolchain: []string{buildcache.Tool("node", "--version")},
	}, func(directory string) error {
		corpus := filepath.Join(directory, "corpus.bin")
		command := toolchain.CommandIn(repository, "node", "internal/native/decode_ascii/corpus.mjs", corpus)
		output, combinedOutputError := command.CombinedOutput()
		if combinedOutputError != nil {
			return fmt.Errorf("node corpus.mjs: %w\n%s", combinedOutputError, output)
		}
		test.Log(string(output))
		data, readError := os.ReadFile(corpus)
		if readError != nil {
			return readError
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
			path := filepath.Join(directory, piece.name()+".bin")
			if writeError := os.WriteFile(
				path,
				data[offsets[piece.first]:offsets[piece.last]],
				0o644,
			); writeError != nil {
				return writeError
			}
		}
		// The units read only their ranges, so the whole corpus would double the product.
		return os.Remove(corpus)
	})
}

// decodeAsciiDecoder is the decoder probe for target, native under ASan and UBSan or wasi: the runtime's C with
// probe.c and baseline.c, a product built ahead by TestProduct_DecodeASCIINative and TestProduct_DecodeASCIIWASI. Every
// unit of a target runs its own range of the corpus through the same binary, which each used to wait on in the same
// cache as the corpus.
func decodeAsciiDecoder(test *testing.T, target string) string {
	test.Helper()
	repository, absolutePathError := filepath.Abs("../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	runtimeDirectory := "internal/native/runtime"
	inputs := buildcache.Inputs{
		Name: "decode-ascii-" + target,
		Files: []string{
			runtimeDirectory,
			"internal/native/decode_ascii/probe.c",
			"internal/native/decode_ascii/baseline.c",
			"internal/native/decode_ascii_cache_test.go",
		},
	}
	// A planted runtime (ADAMIC_DECODE_RUNTIME, as the cache mutants plant one) lies outside the tree, so its content
	// keys the decoder. It is built for this machine alone.
	if alternate := os.Getenv("ADAMIC_DECODE_RUNTIME"); alternate != "" {
		runtimeDirectory = alternate
		snapshot, readRuntimeError := readRuntime(os.DirFS(alternate), ".")
		if readRuntimeError != nil {
			test.Fatal(readRuntimeError)
		}
		for _, file := range snapshot {
			inputs.Flags = append(inputs.Flags, fmt.Sprintf("runtime %s %x", file.name, sha256.Sum256(file.contents)))
		}
	}
	runtimePath := runtimeDirectory
	if !filepath.IsAbs(runtimePath) {
		runtimePath = filepath.Join(repository, runtimePath)
	}
	sources, globError := filepath.Glob(filepath.Join(runtimePath, "*.c"))
	if globError != nil {
		test.Fatal(globError)
	}
	if len(sources) == 0 {
		test.Fatal("no runtime sources")
	}
	// clang reads the tree's files by their names in it, so a sanitized binary's globals name no checkout.
	var files []string
	for _, source := range sources {
		if !filepath.IsAbs(runtimeDirectory) {
			var relativePathError error
			if source, relativePathError = filepath.Rel(repository, source); relativePathError != nil {
				test.Fatal(relativePathError)
			}
		}
		files = append(files, source)
	}
	files = append(files, "internal/native/decode_ascii/probe.c", "internal/native/decode_ascii/baseline.c")
	compiler := "clang"
	flags := []string{
		"-std=c11",
		"-Wall",
		"-Wextra",
		"-Werror",
		"-pedantic",
		"-O2",
		"-ffp-contract=off",
		"-fno-optimize-sibling-calls",
		"-I",
		runtimeDirectory,
	}
	if target == "wasi" {
		sysroot := os.Getenv("WASI_SYSROOT")
		if sysroot == "" {
			test.Fatal("WASI_SYSROOT missing")
		}
		compiler = filepath.Join(filepath.Dir(filepath.Dir(sysroot)), "bin/clang")
		flags = append(
			flags,
			"--target=wasm32-wasi",
			"--sysroot="+sysroot,
			"-DADAMIC_TARGET_WASI=1",
			"-Wl,-z,stack-size=1048576",
		)
	} else {
		flags = append(flags, "-g", "-fsanitize=address,undefined", "-fno-sanitize-recover=all")
	}
	inputs.Flags = append(inputs.Flags, flags...)
	inputs.Toolchain = []string{buildcache.Tool(compiler, "--version")}
	directory := buildcache.Product(test, inputs, func(directory string) error {
		objectDirectory, temporaryDirectoryError := os.MkdirTemp("", "adamic-decode-")
		if temporaryDirectoryError != nil {
			return temporaryDirectoryError
		}
		defer os.RemoveAll(objectDirectory)
		// The checkout's directory is mapped out of the debug information, which would otherwise name it.
		compileFlags := prefixMaps(repository, "/adamic-decode")
		for _, flag := range flags {
			if !strings.HasPrefix(flag, "-Wl,") {
				compileFlags = append(compileFlags, flag)
			}
		}
		objects := make([]string, len(files))
		failures := make([]error, len(files))
		jobs := make(chan int)
		var workers sync.WaitGroup
		for worker := 0; worker < 4; worker++ {
			workers.Add(1)
			go func() {
				defer workers.Done()
				for index := range jobs {
					objects[index] = filepath.Join(objectDirectory, filepath.Base(files[index])+".o")
					command := toolchain.CommandIn(
						repository,
						compiler,
						append(append([]string(nil), compileFlags...), "-c", files[index], "-o", objects[index])...,
					)
					if output, combinedOutputError := command.CombinedOutput(); combinedOutputError != nil {
						failures[index] = fmt.Errorf("compile %s: %w\n%s", files[index], combinedOutputError, output)
					}
				}
			}()
		}
		for index := range files {
			jobs <- index
		}
		close(jobs)
		workers.Wait()
		for _, err := range failures {
			if err != nil {
				return err
			}
		}
		arguments := append(append([]string(nil), flags...), objects...)
		arguments = append(arguments, "-lm", "-o", filepath.Join(directory, "decode-"+target))
		command := toolchain.CommandIn(repository, compiler, arguments...)
		if output, combinedOutputError := command.CombinedOutput(); combinedOutputError != nil {
			return fmt.Errorf("decoder link: %w\n%s", combinedOutputError, output)
		}
		return nil
	})
	return filepath.Join(directory, "decode-"+target)
}

// TestProduct_DecodeAsciiCorpus builds the corpus every decode unit runs its range of, so the tree build builds it
// ahead.
func TestProduct_DecodeAsciiCorpus(test *testing.T) {
	test.Parallel()
	decodeAsciiCorpus(test)
}

// TestProduct_DecodeAsciiNative builds the sanitized native decoder the native units run, so the tree build builds it
// ahead.
func TestProduct_DecodeAsciiNative(test *testing.T) {
	test.Parallel()
	decodeAsciiDecoder(test, "native")
}

// TestProduct_DecodeAsciiwasi builds the WASI decoder the WASI units run, where they run, so the tree build builds it
// ahead.
func TestProduct_DecodeAsciiwasi(test *testing.T) {
	test.Parallel()
	if os.Getenv("ADAMIC_TEST_WASI") != "1" {
		test.Skip("set ADAMIC_TEST_WASI=1")
	}
	decodeAsciiDecoder(test, "wasi")
}
