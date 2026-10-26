package worker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/native"
)

func workerJSONRefused(key string) bool {
	return key == "decode/json_decode_coverage_empty.a" || key == "decode/json_decode_empty_object.a"
}

func workerJSONKeys() []string {
	keys := []string{"decode/corpus", "encode/corpus", "decode/fast_risks.a"}
	for _, mode := range []string{"decode", "encode"} {
		// Reads: internal/oracle/testdata/json_*.a
		paths, err := filepath.Glob("../oracle/testdata/json_" + mode + "_*.a")
		if err != nil {
			panic(err)
		}
		for _, path := range paths {
			key := mode + "/" + filepath.Base(path)
			if !workerJSONRefused(key) {
				keys = append(keys, key)
			}
		}
	}
	return keys
}

// Only compilation is cached. Each behavior unit reruns Node, native, source
// oracle and planted runtime mutants with fresh observations.
func workerJSONProduct(t *testing.T, key string, build bool) string {
	t.Helper()
	mode, name, ok := strings.Cut(key, "/")
	if !ok {
		t.Fatal(key)
	}
	inputs, err := buildcache.GoBuildInInputs("", "adamic", "./cmd/adamic", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	inputs.Name = "worker-json-" + key
	inputs.Files = slices.Clone(inputs.Files)
	inputs.Flags = slices.Clone(inputs.Flags)
	inputs.Toolchain = slices.Clone(inputs.Toolchain)
	inputs.Files = append(inputs.Files, "internal/worker/json_products_test.go")
	fixture := "internal/oracle/testdata/" + name
	if key == "decode/fast_risks.a" {
		fixture = "internal/worker/testdata/json_decode/fast_risks.a"
	}
	if name == "corpus" {
		fixture = "internal/worker/testdata/json_" + mode + "/native.a"
		inputs.Files = append(inputs.Files, "internal/worker/testdata/json_"+mode+"/worker.a")
	}
	inputs.Files = append(inputs.Files, fixture)
	if name == "corpus" {
		inputs.Files = append(inputs.Files, "internal/worker/testdata/json_"+mode+"/cases.a")
	}
	request, err := native.RuntimeRequestFor(native.Options{Release: true})
	if err != nil {
		t.Fatal(err)
	}
	inputs.Flags = append(inputs.Flags, request.Portable...)
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool(request.Compiler, "--version"), buildcache.Tool(request.Archiver, "--version"), buildcache.Tool("node", "--version"))
	if !build {
		return findWorkerProduct(t, inputs)
	}
	compiler := workerCompilerProduct(t, nil, false)
	return buildcache.Product(t, inputs, func(directory string) error {
		root, err := filepath.Abs("../..")
		if err != nil {
			return err
		}
		source, err := os.ReadFile(filepath.Join(root, fixture))
		if err != nil {
			return err
		}
		entry := filepath.Join(directory, "entry.a")
		if name == "corpus" {
			entry = filepath.Join(root, "internal/worker/testdata/json_"+mode+"/worker.a")
		} else {
			source = append(source, []byte("\nimport type { HttpRequest, HttpResponse } from 'adamic/http';\nexport function handle(request: HttpRequest): HttpResponse { return {status:200,headers:[],body:request.body}; }\n")...)
			if err := os.WriteFile(entry, source, 0644); err != nil {
				return err
			}
		}
		for _, args := range [][]string{{"build", fixture, "-o", filepath.Join(directory, "native")}, {"worker", entry, "--out", filepath.Join(directory, "worker")}} {
			command := exec.CommandContext(t.Context(), compiler, args...)
			command.Dir = root
			if output, err := command.CombinedOutput(); err != nil {
				return fmt.Errorf("%v: %w\n%s", args, err, output)
			}
		}
		return nil
	})
}

func runWorkerJSONProductsShard(t *testing.T, shard int) {
	for _, key := range workerJSONKeys() {
		if familyOwns("TestProduct_WorkerJSON", shard, key) {
			t.Run(key, func(t *testing.T) { workerJSONProduct(t, key, true) })
		}
	}
}
