package adaptedtree

import (
	"os/exec"
	"path/filepath"
	"testing"
)

// The JSDocArray deletion must preserve LocalsContainer when TypeScript strips
// internal declarations (#7ez1s16). The adapter probe replays the product's
// checkpoints and requires an orphaned @internal comment to be refused before
// any writes. Accepting the orphaned span makes the probe fail.
func TestJSDocArrayDeletionKeepsLocalsContainer(t *testing.T) {
	t.Parallel()
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal(err)
	}
	modules, err := filepath.Abs(filepath.Join(repositoryFromLane, "stage3/api/node_modules"))
	if err != nil {
		t.Fatal(err)
	}
	tree := filepath.Join(Product(t), "tree")
	probe := filepath.Join(repositoryFromLane, "stage3/adapt/99-correctness-tags-cache-table/deletion-test.cjs")
	command := exec.Command(node, probe, tree)
	command.Env = append(command.Environ(), "NODE_PATH="+modules)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("JSDocArray deletion probe: %v\n%s", err, output)
	}

	const want = `Clean after-checkpoint: five unchanged files; declaration stripping keeps LocalsContainer.
Before-checkpoint application reproduces all five after-checkpoint files.
Orphaned comment hides LocalsContainer and the adapter refuses before writing.
`
	if string(output) != want {
		t.Fatalf("JSDocArray deletion probe output:\n%s\nwant:\n%s", output, want)
	}
	t.Logf("\n%s", output)
}
