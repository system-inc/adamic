package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestV4EscapeAdapterConstructSignatureRefusal(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("../../review/compiler/views-v4/construct-misfit.a")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "construct.ts")
	if err := os.WriteFile(path, source, 0600); err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "escaped\nargument\nproducer\nbad\n" || len(truth.stderr) != 0 {
		t.Fatalf("Node: %#v", truth)
	}
	_, err = lowered(t, path)
	var refusal *lower.NotYet
	if !errors.As(err, &refusal) || !strings.Contains(refusal.Where, "construct.ts:7:") || refusal.What != "new an Identifier" || refusal.Fix != "construct the producer directly with its declared parameter types; checked construction through a view is not implemented" {
		t.Fatalf("construction needs NotYet, path and fix: %v", err)
	}
	t.Logf("Node succeeds; compiler refuses: %v", err)
}

func TestV4EscapeAdapterWithoutConstructSignatureRefusal(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("../../review/compiler/views-v4/construct-no-signature.a")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "no-construct.ts")
	if err := os.WriteFile(path, source, 0600); err != nil {
		t.Fatal(err)
	}
	truth := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle", "node.mjs"), path)
	if truth.exitCode != 0 || string(truth.stdout) != "producer\nbad\n" || len(truth.stderr) != 0 {
		t.Fatalf("Node: %#v", truth)
	}
	_, err = load.Load([]string{path})
	var refusal *load.CheckError
	if !errors.As(err, &refusal) || !strings.Contains(err.Error(), "no-construct.ts:5:37: error TS7009:") || !strings.Contains(err.Error(), "use a producer with a construct signature directly, or call the callable directly") {
		t.Fatalf("checker refusal needs path and fix: %v", err)
	}
	t.Logf("Node succeeds; checker refuses before lowering: %v", err)
}
