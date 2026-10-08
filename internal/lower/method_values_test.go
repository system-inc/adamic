package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func TestMethodValuesProof(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, source string
		refused      bool
	}{
		{"free arrow", "const object = { read(): number { const f = () => 3; return f(); } }; const f = object.read; console.log(`${f()}`);", false},
		{"lexical arrow", "const object = { value: 3, read(): number { const f = () => this.value; return f(); } }; const f = object.read;", true},
		{"unrelated same name", "const object = { read(): number { return 3; } }; const other = { value: 4, read(): number { return this.value; } }; const f = object.read; console.log(`${f()} ${other.read()}`);", false},
		{"interface free", "interface Reader { read(): number; } const object: Reader = { read(): number { return 3; } }; const f = object.read; console.log(`${f()}`);", false},
		{"interface reads this", "interface Reader { read(): number; } class Counter { value=3; read(): number { return this.value; } } const object: Reader = new Counter(); const f = object.read;", true},
		{"override reads this", "class Base { read(): number { return 3; } } class Child extends Base { value = 4; override read(): number { return this.value; } } const object: Base = new Child(); const f = object.read;", true},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var refusal *Refused
			if probe.refused {
				if !errors.As(err, &refusal) || !strings.Contains(err.Error(), "a method read as a value") || !strings.Contains(err.Error(), ".bind(") {
					t.Fatalf("want receiver refusal and bind fix, got %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestMethodValuesSelectiveChecks(t *testing.T) {
	t.Parallel()
	source := `class Counter { value=1; read(): number { return this.value; } quiet(): number { return this.value; } } const object=new Counter(); const read=object.read; console.log(` + "`${object.quiet()}`" + `);`
	path := filepath.Join(t.TempDir(), "main.ts")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := Lower(context.Background(), loaded)
	if err != nil {
		t.Fatal(err)
	}
	checks := 0
	for _, function := range program.Functions {
		if function.Name == "extracted_this" {
			checks++
		}
		if strings.HasSuffix(function.Name, "quiet") && function.MayThrow {
			t.Fatalf("unextracted method %s pays for a receiver check", function.Name)
		}
	}
	if checks != 1 {
		t.Fatalf("want one receiver check, got %d", checks)
	}
	walk(program, func(node any) bool {
		if property, ok := node.(ir.Property); ok && property.Extracted && property.Bound != nil {
			t.Fatal("unbound extraction acquired a receiver")
		}
		return true
	})
}

func TestMethodBindCycleIsRefused(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `class Counter { value=1; saved: (() => number) | undefined = undefined; read(): number {return this.value;} } const object=new Counter(); object.saved=object.read.bind(object);`)
	var refusal *Refused
	if !errors.As(err, &refusal) || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("want bound receiver cycle refusal, got %v", err)
	}
}
