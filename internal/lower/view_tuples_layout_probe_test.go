package lower

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Observe original optional shapes and the handed-off rest storage form before
// extending the one constructor. No admission or second tuple path is added.
func TestTupleOptionalRestOriginalLayoutProbe(t *testing.T) {
	root := os.Getenv("ADAMIC_TUPLE_ORIGINAL_DECLS")
	if root == "" {
		t.Skip("prepared original declarations required")
	}
	types := filepath.ToSlash(filepath.Join(root, "compiler/types.d.ts"))
	modules := filepath.ToSlash(filepath.Join(root, "compiler/moduleSpecifiers.d.ts"))
	sys := filepath.ToSlash(filepath.Join(root, "compiler/sys.d.ts"))
	for _, sample := range []struct {
		name, source string
		arity        int
		rest         bool
	}{
		{"watch", fmt.Sprintf("import type { FileWatcherCallback, DirectoryWatcherCallback } from %q; type Target=Parameters<FileWatcherCallback> | Parameters<DirectoryWatcherCallback>;", sys), -1, false},
		{"module-specifiers", fmt.Sprintf("import type { SourceFile, ModulePath, ModuleSpecifierCache } from %q; import type { ModuleSpecifierResult } from %q; type Target=readonly [kind?:ModuleSpecifierResult['kind'],specifiers?:readonly string[],moduleFile?:SourceFile,modulePaths?:readonly ModulePath[],cache?:ModuleSpecifierCache];", types, modules), 5, false},
		{"optional-map", "type Target=readonly [number,string?];", 2, false},
		{"rest-map", "type Target=readonly [number,...string[]];", 2, true},
	} {
		t.Run(sample.name, func(t *testing.T) {
			l, target, release := mixedUnionLowering(t, sample.source)
			defer release()
			members := []*checker.Type{target}
			if target.Flags()&checker.TypeFlagsUnion != 0 {
				members = target.Types()
			}
			for _, member := range members {
				if !checker.IsTupleType(member) {
					t.Fatalf("original shape is not a tuple: %s", l.checker.TypeToString(member))
				}
				flags := member.TargetTupleType().ElementFlags()
				arguments := l.checker.GetTypeArguments(member)
				if sample.arity >= 0 && len(arguments) != sample.arity {
					t.Fatalf("arity %d", len(arguments))
				}
				if len(arguments) != len(flags) {
					t.Fatal("position metadata differs")
				}
				for index, argument := range arguments {
					t.Logf("position %d: %s flag=%d", index, l.checker.TypeToString(argument), flags[index])
				}
				if sample.rest && (flags[len(flags)-1] != checker.ElementFlagsRest || l.checker.TypeToString(arguments[len(arguments)-1]) != "string") {
					t.Fatal("rest storage element changed")
				}
				if member.TargetTupleType().ElementFlags()[len(flags)-1] != checker.ElementFlagsRequired {
					id := l.tupleViewSlot(nil, member, func(*checker.Type) ir.ViewContractID { return 1 })
					if sample.rest && id != 0 {
						t.Fatal("rest source admission remains refused")
					}
					if !sample.rest && (id == 0 || !l.result.ViewContracts[id-1].TupleVariable) {
						t.Fatal("optional arity plan missing")
					}
				}
			}
			if sample.name == "watch" && len(members) != 2 {
				t.Fatal("original callback alternatives changed")
			}
		})
	}
}
