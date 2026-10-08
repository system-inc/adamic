package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The complete original receiver currently refuses a reached element field.
func TestCheckedViewRanked24AmdDependencyFrontier(t *testing.T) {
	declarations := os.Getenv("ADAMIC_ARRAY23_ORIGINAL_DECLS")
	if declarations == "" {
		t.Skip("complete original declarations required")
	}
	file := filepath.Join(t.TempDir(), "amd-dependency-frontier.a")
	source := fmt.Sprintf(`import type { SourceFile } from %q;
interface Base { readonly kind: number; }
function read(base: Base): void {
 const viewed = base as SourceFile;
 console.log(viewed.amdDependencies[0]!.path);
}
const raw = {kind: 308, amdDependencies: [{path: 'a'}]};
read(raw);
`, filepath.ToSlash(filepath.Join(declarations, "compiler/types.d.ts")))
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	if diff := disagreement(run{stdout: []byte("a\n")}, onNode(t, file)); diff != "" {
		t.Fatal(diff)
	}
	_, err := lowered(t, file)
	const reason = "Adamic 0.1 refuses checked view read of field path with unsupported intersection contract; prove or implement the intersection contract before reading this field"
	if err == nil || !strings.Contains(err.Error(), reason) {
		t.Fatalf("intersection boundary changed: %v", err)
	}
	t.Log(reason)
}
