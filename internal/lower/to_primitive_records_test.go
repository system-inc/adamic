package lower

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// plainMethods asks whether any record may hold a function, by the record-making nodes it names:
// RecordLiteral, RecordCall, RecordCoalesce and a record ObjectLiteral (to_primitive.go). That list is
// kept by hand beside the IR (Lower's should on #09dvwny), so this holds it to where records are made:
// every native call of adamic_record_new sits in an emitter of a node the list names. A new emitter that
// makes a record goes red here until plainMethods judges its node too. The runtime makes none of its
// own; record.c only defines it. Mutant: a name dropped from recordMakers, which this names.
func TestPlainMethodsKnowEveryRecordMaker(t *testing.T) {
	t.Parallel()
	// Each emitter that makes a record, and the node plainMethods judges for it.
	recordMakers := map[string]string{
		"dictionaryLiteral": "ir.RecordLiteral",
		"recordLiteral":     "a record ir.ObjectLiteral",
	}
	const maker = "adamic_record_new("
	found := map[string]bool{}
	sources, err := filepath.Glob(filepath.Join("..", "native", "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range sources {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		total := strings.Count(string(source), maker)
		if total == 0 {
			continue
		}
		files := token.NewFileSet()
		file, err := parser.ParseFile(files, path, source, 0)
		if err != nil {
			t.Fatal(err)
		}
		inside := 0
		for _, declaration := range file.Decls {
			function, isFunction := declaration.(*ast.FuncDecl)
			if !isFunction {
				continue
			}
			body := string(source[files.Position(function.Pos()).Offset:files.Position(function.End()).Offset])
			count := strings.Count(body, maker)
			if count == 0 {
				continue
			}
			inside += count
			if _, known := recordMakers[function.Name.Name]; !known {
				t.Errorf("%s: %s makes a record, and plainMethods doesn't judge its node", path, function.Name.Name)
			}
			found[function.Name.Name] = true
		}
		if inside != total {
			t.Errorf("%s: %d of %d record makers sit outside any function", path, total-inside, total)
		}
	}
	for name, node := range recordMakers {
		if !found[name] {
			t.Errorf("%s (%s) no longer makes a record; drop it here and check plainMethods", name, node)
		}
	}
	runtime, err := filepath.Glob(filepath.Join("..", "native", "runtime", "*.c"))
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range runtime {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		calls := strings.Count(string(source), maker)
		if filepath.Base(path) == "record.c" {
			calls-- // its definition
		}
		if calls != 0 {
			t.Errorf("%s makes records of its own (%d), which plainMethods can't see", path, calls)
		}
	}
	if len(sources) == 0 || len(runtime) == 0 {
		t.Fatalf("read no emitter or runtime source (%d, %d)", len(sources), len(runtime))
	}
}
