package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

func higherRankSource(t *testing.T, extension string, source string) error {
	t.Helper()
	path := filepath.Join(t.TempDir(), "main"+extension)
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = Lower(context.Background(), program)
	return err
}

func checkHigherRank(t *testing.T, extension string, source string, names ...string) {
	t.Helper()
	err := higherRankSource(t, extension, source)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "higher-rank callable slot") {
		t.Fatalf("want higher-rank Refused, got %v", err)
	}
	for _, name := range names {
		if !strings.Contains(refused.What, name) {
			t.Fatalf("missing free binder/path %q: %v", name, err)
		}
	}
	if !strings.Contains(refused.Fix, "concrete callable slots") || !strings.Contains(refused.Where, "main"+extension+":") {
		t.Fatalf("missing path or fix: %v", err)
	}
	t.Log(err)
}

const higherRankProgram = `function identity<T>(value: T): T { return value; }
const box: { readonly run: <U, V>(value: U, extra: V) => U } = {run: identity};
console.log(box.run(4, "x") + " " + box.run("x", 4));`

func TestHigherRankSlotAdamic(t *testing.T) {
	t.Parallel()
	checkHigherRank(t, ".a", higherRankProgram, "box.run", "U", "V")
}
func TestHigherRankSlotTypeScript(t *testing.T) {
	t.Parallel()
	checkHigherRank(t, ".ts", higherRankProgram, "box.run", "U", "V")
}
func TestHigherRankParameter(t *testing.T) {
	t.Parallel()
	checkHigherRank(t, ".a", `function use(run: <Element>(value: Element) => Element): void { }`, "run", "Element")
}
func TestHigherRankContainer(t *testing.T) {
	t.Parallel()
	checkHigherRank(t, ".a", `function identity<T>(value: T): T { return value; } const items = [identity];`, "items", "T")
}
func TestGenericValueUnusedBinderIsRefused(t *testing.T) {
	t.Parallel()
	err := higherRankSource(t, ".a", `function number<T>(): number { return 4; } const run: () => number = number;`)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "type parameters: T") || !strings.Contains(refused.What, "run") {
		t.Fatalf("want unresolved binder and flow path: %v", err)
	}
}
func TestGenericValueDeclaredDefault(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNodeNative(t, `function number<T = number>(): number { return 4; } const run: () => number = number; console.log("" + run());`)
}
