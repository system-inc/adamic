package lower

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestWorkerPrivateParserProofAndMutants(t *testing.T) {
	t.Parallel()
	bytes, err := os.ReadFile("../oracle/testdata/concurrency/accepted/task_private_parser.a")
	if err != nil {
		t.Fatal(err)
	}
	source := string(bytes)
	if _, err := lowerSource(t, source); err != nil {
		t.Fatal(err)
	}
	mutations := map[string]string{
		"shared receiver":          strings.Replace(strings.Replace(source, "const items:", "const sharedParser = new Parser('shared');\nconst items:", 1), "const parser = new Parser(text);", "const parser = sharedParser;", 1),
		"publication into capture": strings.Replace(strings.Replace(source, "const items:", "let leaked: Parser | undefined;\nconst items:", 1), "const parser = new Parser(text);", "const parser = new Parser(text); leaked = parser;", 1),
		"constructor publication":  strings.Replace(strings.Replace(source, "const items:", "let leaked: Parser | undefined;\nconst items:", 1), "this.scanner = new Scanner(text);", "leaked = this; this.scanner = new Scanner(text);", 1),
		"method publication":       strings.Replace(strings.Replace(source, "const items:", "let leaked: Scanner | undefined;\nconst items:", 1), "this.scanner.scan();", "leaked = this.scanner; this.scanner.scan();", 1),
	}
	for name, mutant := range mutations {
		t.Run(name, func(t *testing.T) {
			_, err := lowerSource(t, mutant)
			var refused *Refused
			if !errors.As(err, &refused) {
				t.Fatalf("ownership mutant must be Refused, got %v", err)
			}
			t.Log(refused)
		})
	}
}

func TestWorkerPrivateSharedChildAndLoopOrigins(t *testing.T) {
	t.Parallel()
	for _, body := range []string{
		"const local = new Reader(items); local.change(); return 0;",
		"const local = new Reader([]); let target: readonly number[] = []; for(let index=0;index<3;index++) { const mutable = target as number[]; mutable.push(index); target=items; } return local.size();",
		"const local = new Reader([]); let target: readonly number[] = []; if(flag) { target=items; } const mutable = target as number[]; mutable.push(1); return local.size();",
	} {
		source := "import { parallelMap } from 'adamic'; class Reader { readonly data: readonly number[]; constructor(data: readonly number[]) { this.data=data; } change(): void { const mutable = this.data as number[]; mutable.push(1); } size(): number { return this.data.length; } } const items: readonly number[]=[1,2]; console.log(parallelMap(items, flag => { " + body + " }).join(','));"
		_, err := lowerSource(t, source)
		var refused *Refused
		if !errors.As(err, &refused) || !strings.Contains(refused.What, "shared") {
			t.Errorf("shared origin must prevent mutation, got %v", err)
		}
	}
}

func TestWorkerPrivateUnknownDispatchAndHiddenExpressions(t *testing.T) {
	t.Parallel()
	for _, body := range []string{
		"const parser = new Parser(text); const method = parser.read; return method();",
		"const parser = new Parser(text); const ignored = `${count++}`; return parser.read();",
		"const parser = new Parser(text); publish(parser); return parser.read();",
	} {
		source := "import { parallelMap } from 'adamic'; let count=0; let leaked: Parser | undefined; class Parser { value=0; constructor(text: string) { this.value=text.length; } read(): number { this.value++; return this.value; } } function publish(item: Parser): void { leaked=item; } const items: readonly string[]=['first','second']; console.log(parallelMap(items,text=>{ " + body + " }).join(','));"
		_, err := lowerSource(t, source)
		var refused *Refused
		if !errors.As(err, &refused) {
			t.Errorf("hidden effect or unknown dispatch must refuse: %s got %v", body, err)
		}
	}
	source := "import { parallelMap } from 'adamic'; class Recursive { child = new Recursive(); read(): number { return 1; } } const items: readonly number[]=[1]; parallelMap(items,()=>new Recursive().read());"
	_, err := lowerSource(t, source)
	var refused *Refused
	if !errors.As(err, &refused) {
		t.Fatalf("recursive construction must refuse without panic: %v", err)
	}
}
