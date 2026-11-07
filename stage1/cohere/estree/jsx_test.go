package estree

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/stage1/cohere/estree/oracles"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func jsxCases() []string                 { return oracles.JSXCases() }
func jsxManifest(test *testing.T) string { return jsxManifestCases(test, jsxCases()) }
func jsxManifestCases(test *testing.T, cases []string) string {
	test.Helper()
	dir := test.TempDir()
	var list strings.Builder
	for i, text := range cases {
		path := filepath.Join(dir, fmt.Sprintf("%04d.tsx", i))
		if writeError := os.WriteFile(path, []byte(text), 0644); writeError != nil {
			test.Fatal(writeError)
		}
		list.WriteString(path + "\n")
	}
	path := filepath.Join(dir, "manifest")
	if writeError := os.WriteFile(path, []byte(list.String()), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	return path
}

func TestJSXOriginalLibraries(test *testing.T) {
	test.Parallel()
	library := os.Getenv("ADAMIC_ESTREE_LIBRARY")
	if library == "" {
		test.Skip(
			"set ADAMIC_ESTREE_LIBRARY to an npm install of @typescript-eslint/typescript-estree@8.65.0, typescript@6.0.3 and prettier@3.9.6; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it",
		)
	}
	oracle := goOracle(test)
	list := jsxManifest(test)
	script, _ := filepath.Abs("testdata/library.mjs")
	for _, mode := range []string{"raw", "postprocessed"} {
		flag := "--raw-json"
		if mode == "postprocessed" {
			flag = "--json"
		}
		wants := strings.Split(strings.TrimSpace(string(execute(test, "", oracle, flag, list))), "\n")
		gots := strings.Split(strings.TrimSpace(string(execute(test, "", "node", script, library, mode, list))), "\n")
		if len(wants) != 9 || len(gots) != 9 {
			test.Fatal("JSX library record count")
		}
		for i := range wants {
			var want, got map[string]any
			if decodeError := json.Unmarshal([]byte(wants[i]), &want); decodeError != nil {
				test.Fatal(decodeError)
			}
			if decodeError := json.Unmarshal([]byte(gots[i]), &got); decodeError != nil {
				test.Fatal(decodeError)
			}
			if i == 0 {
				expected := "Type argument list cannot be empty."
				if mode != "raw" {
					expected += " (1:11)"
				}
				if len(got) != 1 || got["error"] != expected || !strings.Contains(wants[i], `"params":[]`) {
					test.Fatal("empty JSX type argument gap changed")
				}
				test.Log(mode + ": original rejects empty JSX type arguments; Go accepts")
				continue
			}
			if mode != "raw" && (i == 4 || i == 6) {
				init := want["ast"].(map[string]any)["body"].([]any)[0].(map[string]any)["declarations"].([]any)[0].(map[string]any)["init"].(map[string]any)
				var value map[string]any
				var before, after string
				if i == 4 {
					value = init["openingElement"].(map[string]any)["attributes"].([]any)[0].(map[string]any)["value"].(map[string]any)
					before = "&😀"
					after = "&amp;😀"
				} else {
					value = init["children"].([]any)[0].(map[string]any)
					before = " // text /* not comment */ &&unknown;&#1114112; "
					after = " // text /* not comment */ &amp;&unknown;&#1114112; "
				}
				if value["value"] != before {
					test.Fatal("entity gap changed")
				}
				value["value"] = after
				test.Logf("postprocessed case %d: only ampersand spelling differs", i)
			}
			if mode != "raw" && i == 8 {
				normalized := jsxManifestCases(test, []string{strings.ReplaceAll(jsxCases()[i], "\r\n", "\n")})
				data := execute(test, "", oracle, flag, normalized)
				if decodeError := json.Unmarshal(bytes.TrimSpace(data), &want); decodeError != nil {
					test.Fatal(decodeError)
				}
				test.Log("postprocessed CRLF: library exactly matches Go on LF-normalized input")
			}
			if !bytes.Equal(mustJSON(test, want), mustJSON(test, got)) {
				test.Fatalf(
					"%s JSX %d differs beyond documented deltas: Go %s; library %s",
					mode,
					i,
					mustJSON(test, want),
					mustJSON(test, got),
				)
			}
		}
	}
}

const testJSXMutantShards = 32

// Shard counts stay fixed as cases grow; keys are this file, mode and source-case identity.
// ADAMIC_TEST_SHARD=i/n selects shards; unset runs every JSX mutant case.
// jsxMutantPath is the port with JSX's optional flag inverted.
func jsxMutantPath(test *testing.T) string {
	test.Helper()
	return miscMutant(test, "jsxConvert.ts", "boolValue(source.optional)", "boolValue(!source.optional)")
}

// Not parallel: miscBuild writes the shared adamic-build and adamic/runtime cache directories
func TestJSXMutant(test *testing.T) {
	finishSetup := miscStart(test)
	cases := jsxCases()
	ids := miscIDs("stage1/cohere/estree/jsx_test.go:jsx-mutant", cases)
	oracle := miscOracle(test)
	path := jsxMutantPath(test)
	binary, _ := miscBuild(test, path)
	answers := miscTextAnswers(test, oracle, cases, ".tsx")
	finishSetup()
	miscRunShards(test, testJSXMutantShards, ids, func(test *testing.T, i int) {
		list := jsxManifestCases(test, cases[i:i+1])
		want := answers[i].Data
		for name, got := range map[string][]byte{
			"Node":   onNode(test, path, "--manifest", list),
			"native": execute(test, "", binary, "--manifest", list),
		} {
			if miscCompareError := miscCompare(want, got, true); miscCompareError != nil {
				test.Fatalf("%s %s: %v", ids[i], name, miscCompareError)
			}
		}
	})
}

func TestJSXMutantPlantedSurvivor(test *testing.T) {
	test.Parallel()
	miscPlantedProof(test, testJSXMutantShards, miscIDs(
		"stage1/cohere/estree/jsx_test.go:jsx-mutant",
		jsxCases(),
	), func(planted bool) error {
		got := []byte("disagree")
		if planted {
			got = []byte("agree")
		}
		return miscCompare([]byte("agree"), got, true)
	})
}
