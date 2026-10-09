package estree

import "testing"

func TestThreePortMutants(t *testing.T) {
	list := manifest(t, generated())
	want := execute(t, "", goOracle(t), "--manifest", list)
	for _, item := range []struct{ name, file, from, to string }{
		{"member-computed", "convert.ts", "boolValue(node.kind === 'ElementAccessExpression')", "boolValue(node.kind === 'PropertyAccessExpression')"},
		{"logical-rebalance", "postprocess.ts", "completed.set(id, this.rebalance(id));", "completed.set(id, id);"},
		{"merged-jsdoc-value", "postprocess.ts", "*//*", "*/ /*"},
	} {
		t.Run(item.name, func(t *testing.T) {
			path := mutantPort(t, item.file, item.from, item.to)
			got := onNode(t, path, "--manifest", list)
			diff := firstDifference(want, got)
			if diff == "" {
				t.Fatal("source Node mutant survived")
			}
			t.Logf("source Node finished; byte comparison caught %s", diff)
			binary, _ := build(t, path, true)
			got = execute(t, "", binary, "--manifest", list)
			diff = firstDifference(want, got)
			if diff == "" {
				t.Fatal("native mutant survived")
			}
			t.Logf("sanitized native finished; byte comparison caught %s", diff)
		})
	}
}
