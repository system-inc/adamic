package estree

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func jsxCases() []string {
	return []string{"const x=<A<>/>;", "const x = <A/>;", "const x = <A> text </A>;", "const x = <><A/>{x}{/* comment */}</>;", "const x = <A.B.C ns:x='&amp;&#x1f600;' data-test={x} bool {...rest}><ns:a/>{...children}</A.B.C>;", "const x = <A<T,> x=\"\\n\"/>;", "const x = <A> // text /* not comment */ &amp;&unknown;&#1114112; </A>;", "let x = <A value={<B/>}/>;", "<A>\n é😀\r\n {x+1}\n</A>;"}
}
func jsxManifest(t *testing.T) string { return jsxManifestCases(t, jsxCases()) }
func jsxManifestCases(t *testing.T, cases []string) string {
	t.Helper()
	dir := t.TempDir()
	var list strings.Builder
	for i, text := range cases {
		path := filepath.Join(dir, fmt.Sprintf("%04d.tsx", i))
		if err := os.WriteFile(path, []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
		list.WriteString(path + "\n")
	}
	path := filepath.Join(dir, "manifest")
	if err := os.WriteFile(path, []byte(list.String()), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestJSXOriginalLibraries(t *testing.T) {
	t.Parallel()
	library := os.Getenv("ADAMIC_ESTREE_LIBRARY")
	if library == "" {
		t.Skip("set ADAMIC_ESTREE_LIBRARY to an npm install of @typescript-eslint/typescript-estree@8.65.0, typescript@6.0.3 and prettier@3.9.6; the gate skips this oracle until #xq2ecw6 (setup --gate-inputs) installs it")
	}
	oracle := goOracle(t)
	list := jsxManifest(t)
	script, _ := filepath.Abs("testdata/library.mjs")
	for _, mode := range []string{"raw", "postprocessed"} {
		flag := "--raw-json"
		if mode == "postprocessed" {
			flag = "--json"
		}
		wants := strings.Split(strings.TrimSpace(string(execute(t, "", oracle, flag, list))), "\n")
		gots := strings.Split(strings.TrimSpace(string(execute(t, "", "node", script, library, mode, list))), "\n")
		if len(wants) != 9 || len(gots) != 9 {
			t.Fatal("JSX library record count")
		}
		for i := range wants {
			var want, got map[string]any
			if err := json.Unmarshal([]byte(wants[i]), &want); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(gots[i]), &got); err != nil {
				t.Fatal(err)
			}
			if i == 0 {
				expected := "Type argument list cannot be empty."
				if mode != "raw" {
					expected += " (1:11)"
				}
				if len(got) != 1 || got["error"] != expected || !strings.Contains(wants[i], `"params":[]`) {
					t.Fatal("empty JSX type argument gap changed")
				}
				t.Log(mode + ": original rejects empty JSX type arguments; Go accepts")
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
					t.Fatal("entity gap changed")
				}
				value["value"] = after
				t.Logf("postprocessed case %d: only ampersand spelling differs", i)
			}
			if mode != "raw" && i == 8 {
				normalized := jsxManifestCases(t, []string{strings.ReplaceAll(jsxCases()[i], "\r\n", "\n")})
				data := execute(t, "", oracle, flag, normalized)
				if err := json.Unmarshal(bytes.TrimSpace(data), &want); err != nil {
					t.Fatal(err)
				}
				t.Log("postprocessed CRLF: library exactly matches Go on LF-normalized input")
			}
			if !bytes.Equal(mustJSON(t, want), mustJSON(t, got)) {
				t.Fatalf("%s JSX %d differs beyond documented deltas: Go %s; library %s", mode, i, mustJSON(t, want), mustJSON(t, got))
			}
		}
	}
}

const testJSXMutantShards = 32

// Shard counts stay fixed as cases grow; keys are this file, mode and source-case identity.
// ADAMIC_TEST_SHARD=i/n selects shards; unset runs every JSX mutant case.
// Not parallel: miscBuild writes the shared adamic-build and adamic/runtime cache directories
func TestJSXMutant(t *testing.T) {
	finishSetup := miscStart(t)
	cases := jsxCases()
	ids := miscIDs("stage1/cohere/estree/jsx_test.go:jsx-mutant", cases)
	oracle := miscOracle(t)
	path := miscMutant(t, "jsxConvert.ts", "boolValue(source.optional)", "boolValue(!source.optional)")
	binary, _ := miscBuild(t, path)
	answers := miscTextAnswers(t, oracle, cases, ".tsx")
	finishSetup()
	miscRunShards(t, testJSXMutantShards, ids, func(t *testing.T, i int) {
		list := jsxManifestCases(t, cases[i:i+1])
		want := answers[i].Data
		for name, got := range map[string][]byte{"Node": onNode(t, path, "--manifest", list), "native": execute(t, "", binary, "--manifest", list)} {
			if err := miscCompare(want, got, true); err != nil {
				t.Fatalf("%s %s: %v", ids[i], name, err)
			}
		}
	})
}

func TestJSXMutantPlantedSurvivor(t *testing.T) {
	t.Parallel()
	miscPlantedProof(t, testJSXMutantShards, miscIDs("stage1/cohere/estree/jsx_test.go:jsx-mutant", jsxCases()), func(planted bool) error {
		got := []byte("disagree")
		if planted {
			got = []byte("agree")
		}
		return miscCompare([]byte("agree"), got, true)
	})
}
