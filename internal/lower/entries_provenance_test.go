package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestEntriesAllocationProof(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		source string
		proven bool
	}{
		{`const source={value:1, hidden:2}; const view:{readonly value:number}=source; Object.entries(view);`, true},
		{`let source={value:1, hidden:2}; const view:{readonly value:number}=source; Object.entries(view);`, true},
		{`let source:{readonly value:number}={value:1}; const other={value:2, hidden:'bad'}; source=other; Object.entries(source);`, false},
		{`let source:{readonly value:number}={value:1}; const other={value:2, hidden:'bad'}; [source]=[other]; Object.entries(source);`, false},
		{`function entries(source:{readonly value:number}):void {Object.entries(source);} entries({value:1});`, false},
		{`const source={value:1, hidden:'bad'}; const view:{readonly value:number}=source; Object.entries(view);`, false},
	} {
		t.Run(probe.source, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSource(t, probe.source)
			if err != nil {
				t.Fatal(err)
			}
			found := 0
			inspect := func(value any) bool {
				if call, ok := value.(ir.ObjectCall); ok && call.Method == "entries" {
					found++
					if call.Checked == probe.proven {
						t.Errorf("proven=%t, checked=%t", probe.proven, call.Checked)
					}
				}
				return true
			}
			walk(program.Main, inspect)
			for _, function := range program.Functions {
				walk(function.Body, inspect)
			}
			if found != 1 {
				t.Fatalf("found %d entries calls", found)
			}
		})
	}
}

// Index annotations do not establish an allocation or a named field slot.
func TestEntriesRecordBoundaries(t *testing.T) {
	for _, probe := range []struct{ source, reason string }{
		{`const base={visible:1}; const source:{[key:string]:number;visible:number}={...base}; source['late']=2; Object.entries(source);`, "spread into record storage"},
		{`function write(value:{[key:string]:number}):void {value['late']=2;} const source={visible:1}; Object.entries(source);`, "closed const literal origin"},
		{`const source:{[key:string]:number;visible:number}={visible:1}; source['late']=2; console.log(source.late.toString()); Object.entries(source);`, "indexed record field read"},
	} {
		t.Run(probe.reason, func(t *testing.T) {
			_, err := lowerSource(t, probe.source)
			var stopped *NotYet
			if !errors.As(err, &stopped) || !strings.Contains(stopped.What, probe.reason) {
				t.Fatalf("want explicit NotYet %q, got %v", probe.reason, err)
			}
		})
	}
}
