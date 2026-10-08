package oracle

import "testing"

// The fourth ranked group of lane 2 array contracts: CaseBlock.clauses, ParsedCommandLine.fileNames,
// DiagnosticMessageChain.next, FlowLabel.antecedent and CommaListExpression.elements, each modeled
// on tsc 6.0.3's declaration with a synthetic tagged root. Writes through the mutable views keep
// the original array's element contract. An array written into a slot that held undefined, and a
// push of records outside the certified flat scalar subset, remain named runtime refusals.
func TestCheckedViewRanked4ArrayContracts(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, node, diagnostic string }{
		{"ranked4-case-lazy", "2:3\n", ""},
		{"ranked4-case-nested", "undefined\n", "field read failed: clause.expression is not a Expression; expected Expression, found number"},
		{"ranked4-case-wrong-array", "undefined\n", "field read failed: caseBlock(raw).clauses is not a NodeArray<CaseOrDefaultClause>; expected NodeArray<CaseOrDefaultClause>, found number"},
		{"ranked4-case-wrong-tag", "3\n", "field read failed: clauses[0].kind expected \"case\" | \"default\", found string other"},
		{"ranked4-case", "1:3:default\n", ""},
		{"ranked4-chain-absent", "absent\n", ""},
		{"ranked4-chain-deep", "7\n", "field read failed: next[0]!.next![0]!.messageText is not a string; expected string, found number"},
		{"ranked4-chain-push-wrong", "2:6\n", "element write failed: <array write> expected { kind: \"chain\"; messageText: number; category: number; code: number; }, found { kind: \"chain\"; messageText: string; category: number; code: number; }"},
		{"ranked4-chain-push", "2:added\n", ""},
		{"ranked4-chain-set-wrong", "replaced1\n", "element write failed: <array write> expected { kind: \"chain\"; messageText: number; category: number; code: number; }, found { kind: \"chain\"; messageText: string; category: number; code: number; }"},
		{"ranked4-chain-set", "1:replaced\n", ""},
		{"ranked4-chain-undefined", "absent\n", ""},
		{"ranked4-chain-wrong-array", "5\n", "field read failed: chain(raw).next matches no member of DiagnosticMessageChain[] | undefined; expected DiagnosticMessageChain[] | undefined, found string"},
		{"ranked4-chain", "1:inner:leaf\n", ""},
		{"ranked4-comma-map-wrong", "4,six\n", "field read failed: element.pos is not a number; expected number, found string"},
		{"ranked4-comma-map", "4,6\n", ""},
		{"ranked4-comma-wrong-element", "undefined\n", "element read failed: elements[0] expected Expression, found number"},
		{"ranked4-comma-wrong-field", "four\n", "field read failed: elements[0]!.pos is not a number; expected number, found string"},
		{"ranked4-comma", "1:4\n", ""},
		{"ranked4-files-assign", "2:z.ts,y.ts\n", ""},
		{"ranked4-files-join-wrong", "1,2\n", "element read failed: names[element] expected string, found number"},
		{"ranked4-files-push-wrong", "3\n", "element read failed: <array write> expected string, found number"},
		{"ranked4-files-push", "2:a.ts,c.ts\n", ""},
		{"ranked4-files-set-wrong", "c.ts1\n", "element read failed: <array write> expected string, found number"},
		{"ranked4-files-wrong-element", "undefined\n", "element read failed: names[0] expected string, found number"},
		{"ranked4-files", "2:a.ts,b.ts\n", ""},
		{"ranked4-label-assign-from-array", "2\n", ""},
		{"ranked4-label-assign-undefined", "cleared\n", ""},
		{"ranked4-label-assign", "1:1\n", "field read failed: <write>.antecedent is not a array; expected array, found nullish"},
		{"ranked4-label-push", "2\n", "element write failed: <array write> expected object, found uncertified source element contract"},
		{"ranked4-label-undefined", "absent\n", ""},
		{"ranked4-label-wrong-element", "2:undefined\n", "field read failed: antecedent[0].kind expected \"label\" | \"start\", found string block"},
		{"ranked4-label", "1:1\n", ""},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, path := interfaceFixture(t, "lane2/"+probe.name)
			truth := onNode(t, path)
			want := run{stdout: []byte(probe.node)}
			if difference := viewReadDisagreement(want, truth, program); difference != "" {
				t.Fatal("Node: " + difference)
			}
			if probe.diagnostic != "" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: " + probe.diagnostic + "\n")}
			}
			actual, binary := nativelyUncached(t, program)
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := viewReadDisagreement(want, got, program); difference != "" {
					t.Fatalf("%s; stderr %q; stdout %q; exit %d", difference, got.stderr, got.stdout, got.exitCode)
				}
			}
			if probe.diagnostic == "" {
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
		})
	}
}
