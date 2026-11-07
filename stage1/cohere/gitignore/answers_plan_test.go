package gitignore

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestPortAnswerPreparedUnitPlan(test *testing.T) {
	test.Parallel()
	// An empty group has no owners and no missing cases.
	for _, shard := range portCaseShards(test, cases{}) {
		if len(shard.ids) != 0 {
			test.Fatal("empty group acquired a case")
		}
	}
	// Build a synthetic case in every comparison child and plant a failure in
	// each in turn. A failure must be visible in exactly that child's output.
	asked := cases{}
	for bucket := 0; bucket < portComparisonShards; bucket++ {
		for n := 0; ; n++ {
			name := fmt.Sprintf("witness-%d-%d", bucket, n)
			if portBucket("patterns/"+name) == bucket {
				asked.Patterns = append(asked.Patterns, patterns{Name: name, Queries: []query{{Path: "one"}}})
				break
			}
		}
	}
	shards := portCaseShards(test, asked)
	for target := range shards {
		caught := 0
		for child, shard := range shards {
			want, got := "", ""
			for i := range shard.ids {
				want += "answer\n"
				if child == target && i == 0 {
					got += "planted disagreement\n"
				} else {
					got += "answer\n"
				}
			}
			if firstDifference(got, want) != "" {
				if child != target {
					test.Fatalf("failure escaped child %d", target)
				}
				caught++
			}
		}
		if caught != 1 {
			test.Fatalf("child %d caught planted failure %d times", target, caught)
		}
	}
	file, parseFileError := parser.ParseFile(token.NewFileSet(), "answers_prepared_units_test.go", nil, 0)
	if parseFileError != nil {
		test.Fatal(parseFileError)
	}
	declared := map[string]bool{}
	for _, decl := range file.Decls {
		if f, ok := decl.(*ast.FuncDecl); ok &&
			strings.HasPrefix(f.Name.Name, "TestThePortAnswersAsGoCohereAndGitDo_") &&
			!strings.HasSuffix(f.Name.Name, "_Setup") {
			declared[f.Name.Name] = true
		}
	}
	for i := 0; i < testThePortAnswersAsGoCohereAndGitDoShards; i++ {
		name := fmt.Sprintf("TestThePortAnswersAsGoCohereAndGitDo_%03d", i)
		if !declared[name] {
			test.Fatalf("missing child %s", name)
		}
		delete(declared, name)
	}
	if len(declared) != 0 {
		test.Fatalf("unexpected children: %v", declared)
	}
}
