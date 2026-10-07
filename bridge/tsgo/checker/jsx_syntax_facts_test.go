package checker

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestJsxSyntaxFactsNumericBindingsAndNul(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	file := filepath.Join(dir, "input.tsx")
	config := filepath.Join(dir, "tsconfig.json")
	source := "const Foo=()=>null;const createElement=React.createElement;<Foo/>;<Missing/>;React.createElement('div',null,['\\u0000',1]);export {};"
	for path, text := range map[string]string{file: source, config: `{"compilerOptions":{"strict":true,"target":"ESNext","jsx":"preserve","types":[]},"files":["anchor.d.ts"]}`, filepath.Join(dir, "anchor.d.ts"): "export {};"} {
		if err := os.WriteFile(path, []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	program, err := Open(config, []string{file})
	if err != nil {
		t.Fatal(err)
	}
	wire, err := program.Inspect(file, 0, uint64(len(source)), "SourceFile", "jsx-syntax-facts")
	if err != nil {
		t.Fatal(err)
	}
	fields := decodedFields(t, wire)
	at := 0
	take := func() string {
		if at >= len(fields) {
			t.Fatal("truncated snapshot")
		}
		v := fields[at]
		at++
		return v
	}
	number := func() int {
		v, err := strconv.Atoi(take())
		if err != nil || v < 0 {
			t.Fatal("non-numeric syntax metadata")
		}
		return v
	}
	if number() != 1 || take() != "jsx-syntax-facts" {
		t.Fatal("wrong header")
	}
	roots, count := number(), number()
	if roots < 4 || count < roots {
		t.Fatal("missing listener nodes")
	}
	foundFoo, foundMissing, foundNul := false, false, false
	for id := 1; id < count; id++ {
		kind := number()
		start, end := number(), number()
		path := take()
		number()
		units := number()
		text := []int{}
		for i := 0; i < units; i++ {
			text = append(text, number())
		}
		for i := 0; i < 7; i++ {
			number()
		}
		for list := 0; list < 2; list++ {
			n := number()
			for i := 0; i < n; i++ {
				target := number()
				if target >= count {
					t.Fatal("dangling syntax identity")
				}
			}
		}
		resolved := number()
		declarations := number()
		for i := 0; i < declarations; i++ {
			if number() >= count {
				t.Fatal("dangling declaration identity")
			}
		}
		if path == file && (start > end || end > len(source)) {
			t.Fatal("invalid source byte span")
		}
		if kind == 79 && len(text) == 3 && text[0] == 'F' && text[1] == 'o' && text[2] == 'o' && resolved == 1 && declarations == 1 {
			foundFoo = true
		}
		if kind == 79 && len(text) == 7 && text[0] == 'M' && resolved == 0 {
			foundMissing = true
		}
		if kind == 10 && len(text) == 1 && text[0] == 0 {
			foundNul = true
		}
	}
	if at != len(fields) || !foundFoo || !foundMissing || !foundNul {
		t.Fatalf("snapshot incomplete: Foo=%v Missing=%v NUL=%v", foundFoo, foundMissing, foundNul)
	}
	for _, question := range []string{"jsx-syntax-facts\nextra"} {
		if _, err := program.Inspect(file, 0, uint64(len(source)), "SourceFile", question); err == nil {
			t.Fatal("accepted malformed question")
		}
	}
}
