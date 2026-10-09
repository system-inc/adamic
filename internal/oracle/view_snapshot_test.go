package oracle

import "testing"

func init() {
	additionalFixtureCounts = append(additionalFixtureCounts, func(t *testing.T) []string {
		return []string{counted(t, interfaceFixturePath("v2/snapshot-maybe-boolean"), false, nil, false, false), counted(t, interfaceFixturePath("v2/snapshot-maybe-number"), false, nil, false, false)}
	})
}

func checkViewSnapshot(t *testing.T, name string) {
	t.Helper()
	program, path := interfaceFixture(t, "v2/snapshot-maybe-"+name)
	node := onNode(t, path)
	t.Logf("Node: exit=%d stdout=%q stderr=%q", node.exitCode, node.stdout, node.stderr)
	sanitized, binary := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "native-sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
		t.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
		if difference := disagreement(node, got); difference != "" {
			t.Errorf("%s: %s", backend, difference)
		}
	}
	if !t.Failed() {
		if report := leaksUncached(t, program, binary); report != "" {
			t.Fatal(report)
		}
	}
}

func TestViewSnapshotMaybeBoolean(t *testing.T) {
	t.Parallel()
	checkViewSnapshot(t, "boolean")
}

func TestViewSnapshotMaybeNumber(t *testing.T) {
	t.Parallel()
	checkViewSnapshot(t, "number")
}
