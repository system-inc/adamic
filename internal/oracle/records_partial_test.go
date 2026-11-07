package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"path/filepath"
	"testing"
)

func init() {
	for _, name := range []string{"parser", "views", "undefined"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/records_partial_" + name + ".a", true, false})
	}
}
func TestPartialRecordJavaScriptAgreesWithNode(test *testing.T) {
	test.Parallel()
	for _, name := range []string{"parser", "views", "undefined"} {
		test.Run(name, func(test *testing.T) {
			path, absolutePathError := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/records_partial_"+name+".a"),
			)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			program, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			want, got := onNode(test, path), onJavaScriptBackend(test, program)
			if diff := disagreement(want, got); diff != "" {
				test.Fatalf("%s: Node %+v; lowered %+v", diff, want, got)
			}
		})
	}
}

// An absent optional field is not an own property with an undefined value.
func TestPartialRecordAbsentEntryMutant(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/records_partial_parser.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	changed := false
	for index, statement := range program.Main {
		declaration, ok := statement.(ir.Declare)
		if !ok {
			continue
		}
		literal, ok := declaration.Value.(ir.RecordLiteral)
		if !ok {
			continue
		}
		key := len(program.Strings)
		program.Strings = append(program.Strings, "first")
		literal.Entries = append(
			literal.Entries,
			ir.RecordEntry{Key: ir.StringConstant{Index: key}, Value: ir.MaybeOf{Of: literal.Element}},
		)
		declaration.Value = literal
		program.Main[index] = declaration
		changed = true
	}
	if !changed {
		test.Fatal("mutant changed no record literal")
	}
	want := onNode(test, path)
	native, _ := nativelyUncached(test, program)
	for name, got := range map[string]run{"javascript": onJavaScriptBackend(test, program), "native": native} {
		if got.exitCode != 0 || disagreement(want, got) != "stdout differs" {
			test.Fatalf("%s absent-entry mutant survived: Node %+v; mutant %+v", name, want, got)
		}
		test.Logf("%s absent-entry mutant caught by Node stdout", name)
	}
}
