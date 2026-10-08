package oracle

import "testing"

func TestCheckedViewNodeArrayRecords(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, stdout, diagnostic string }{
		{"node-array-reads", "2:3:9:true\nfirst\n", ""},
		{"node-array-lazy", "2:3\nfirst\n", ""},
		{"node-array-properties-copy", "3:12\n", ""},
		{"node-array-own-mutation", "4:11:true\n", ""},
		{"node-array-missing-own", "", "field read failed: block(raw).statements.pos is not initialized; expected number, found missing"},
		{"node-array-wrong-own", "", "field read failed: block(raw).statements.pos is not a number; expected number, found string"},
		{"node-array-wrong-reference", "", "element read failed: statements[0] expected Statement, found string"},
		{"node-array-wrong-element", "", "element read failed: statements[0] expected Statement, found number"},
		{"node-array-wrong-field", "", "field read failed: statement.text is not a string; expected string, found number"},
		{"node-array-wrong-array", "", "field read failed: block(raw).statements is not a NodeArray<Statement>; expected NodeArray<Statement>, found object"},
		{"node-array-write-literal", "", "field write failed: property 'pos' on array at node-array-write-literal.a:11 has no compatible declared slot"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, path := interfaceFixture(t, "lane2/"+probe.name)
			want := run{stdout: []byte(probe.stdout)}
			// Unsafe source casts execute on Node too; their checked failure is the
			// language's additional contract, pinned below for both backends.
			source := onNode(t, path)
			if probe.diagnostic == "" {
				if difference := disagreement(want, source); difference != "" {
					t.Fatal("Node: " + difference)
				}
			} else {
				sourceOutput := map[string]string{
					"node-array-missing-own":     "undefined\n",
					"node-array-wrong-own":       "wrong\n",
					"node-array-wrong-reference": "present\n",
					"node-array-wrong-element":   "undefined\n",
					"node-array-wrong-field":     "9\n",
					"node-array-wrong-array":     "undefined\n",
					"node-array-write-literal":   "4\n",
				}[probe.name]
				if source.exitCode != 0 || string(source.stdout) != sourceOutput || len(source.stderr) != 0 {
					t.Fatalf("Node control: %#v", source)
				}
				want.exitCode = 70
				want.stderr = []byte("adamic: panic: " + probe.diagnostic + "\n")
			}
			actual, _ := nativelyUncached(t, program)
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s; got %#v", difference, got)
				}
			}
		})
	}
}
