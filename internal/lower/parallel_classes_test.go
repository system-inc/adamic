package lower

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestTaskReadonlyClassWriteMutants(t *testing.T) {
	t.Parallel()
	for _, fixture := range []string{"fields", "items"} {
		t.Run(fixture, func(t *testing.T) {
			bytes, err := os.ReadFile("../oracle/testdata/concurrency/accepted/task_class_" + fixture + ".a")
			if err != nil {
				t.Fatal(err)
			}
			source := string(bytes)
			if _, err := lowerSource(t, source); err != nil {
				t.Fatal(err)
			}
			if fixture == "fields" {
				source = strings.Replace(source, "readonly offsets: readonly number[]", "readonly offsets: number[]", 1)
				source = strings.Replace(source, "offsets: readonly number[])", "offsets: number[])", 1)
				source = strings.Replace(source, "return this.name", "this.offsets.push(index); return this.name", 1)
			} else {
				source = strings.Replace(source, "readonly start", "start", 1)
				source = strings.Replace(source, "return this.path", "this.start++; return this.path", 1)
			}
			_, err = lowerSource(t, source)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(refused.What, "mutable") {
				t.Fatalf("write-through-this mutant must fail the task Shareable proof, got %v", err)
			}
			t.Log(refused)
		})
	}
}

func TestTaskReadonlyClassEffectsRemainRefused(t *testing.T) {
	t.Parallel()
	for _, body := range []string{
		"count++; return this.value;",
		"const alias = this; return alias.value;",
		"const nested = () => this.value; return nested();",
		"return consume(this);",
	} {
		source := "import { parallelMap } from 'adamic'; let count = 0; function consume(item: Readonly<{value: number}>): number { return item.value; } class Reader { readonly value = 4; read(): number { " + body + " } } const reader = new Reader(); const items: readonly number[] = [1,2]; console.log(parallelMap(items, () => reader.read()).join(','));"
		_, err := lowerSource(t, source)
		var refused *Refused
		if !errors.As(err, &refused) || !strings.Contains(refused.What, "parallelMap work") {
			t.Errorf("body %s: want task refusal, got %v", body, err)
		}
	}
}

func TestTaskClassDispatchRemainsClosed(t *testing.T) {
	t.Parallel()
	for _, extra := range []string{
		"reader.read = () => { count++; return 0; };",
		"class Derived extends Reader { read(): number { count++; return 0; } }",
		"const Derived = class extends Reader { read(): number { count++; return 0; } };",
	} {
		source := "import { parallelMap } from 'adamic'; let count = 0; class Reader { readonly value = 4; read(): number { return this.value; } } const reader = new Reader(); " + extra + " const items: readonly number[] = [1,2]; console.log(parallelMap(items, () => reader.read()).join(','));"
		_, err := lowerSource(t, source)
		var refused *Refused
		if !errors.As(err, &refused) || !strings.Contains(refused.What, "dispatch hierarchy") {
			t.Errorf("dispatch mutant %s: got %v", extra, err)
		}
	}
}
