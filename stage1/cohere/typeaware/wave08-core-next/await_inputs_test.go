package wave08core_test

import (
	"context"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	bridge "github.com/system-inc/adamic/bridge/tsgo/checker"
	wave08core "github.com/system-inc/adamic/stage1/cohere/typeaware/wave08-core-next"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestAwaitRawContractAndTypeEdges(t *testing.T) {
	root := t.TempDir()
	config := filepath.Join(root, "tsconfig.json")
	source := filepath.Join(root, "source.ts")
	for path, text := range map[string]string{
		config: `{"compilerOptions":{"strict":true,"target":"ES2022","lib":["ES2022"],"types":[]},"include":["*.ts"]}`,
		source: `const callback: () => Promise<number> = async () => 1;`,
	} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	program, err := bridge.Open(config, []string{source})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Compiler.GetSourceFile(source)
	c, release := program.Compiler.GetTypeCheckerForFile(context.Background(), file)
	defer release()
	ids := map[*checker.Type]uint64{}
	types := map[uint64]*checker.Type{}
	id := func(typ *checker.Type) uint64 {
		if typ == nil {
			return 0
		}
		if ids[typ] == 0 {
			ids[typ] = uint64(len(ids) + 1)
			types[ids[typ]] = typ
		}
		return ids[typ]
	}
	var fn *ast.Node
	var walk func(*ast.Node) bool
	walk = func(node *ast.Node) bool {
		if node.Kind == ast.KindArrowFunction {
			fn = node
		}
		node.ForEachChild(walk)
		return false
	}
	walk(file.AsNode())
	if fn == nil {
		t.Fatal("arrow missing")
	}
	fields := wave08core.AwaitContractFields(c, fn, id)
	if len(fields) < 5 || fields[2] == "0" || fields[3] != "0" || fields[4] != "0" {
		t.Fatalf("contextual/raw ancestry differs: %v", fields)
	}
	rootID, _ := strconv.ParseUint(fields[2], 10, 64)
	signatures := checker.Checker_getSignaturesOfType(c, types[rootID], checker.SignatureKindCall)
	if len(signatures) != 1 {
		t.Fatal("signature missing")
	}
	returns := checker.Checker_getReturnTypeOfSignature(c, signatures[0])
	raw := wave08core.AwaitTypeFields(c, fn, returns, "", false, id)
	if len(raw) < 12 || raw[len(raw)-2] == "0" {
		t.Fatalf("then-property edge missing: %v", raw)
	}
	union := c.GetUnionType([]*checker.Type{returns, checker.Checker_numberType(c)})
	raw = wave08core.AwaitTypeFields(c, fn, union, "", false, id)
	if raw[2] != strconv.FormatUint(uint64(union.Flags()), 10) || raw[3] != "2" {
		t.Fatalf("union edges differ: %v", raw)
	}
	if checker.TypeFlagsUnion != 134217728 {
		t.Fatal("native union flag is not pinned Go value")
	}
	t.Log("contextual signature, then property and two union branches are raw checker edges; pinned union flag 134217728")
}
