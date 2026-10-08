package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
	"testing"
)

func TestCheckedViewNominalArrayReadMutant(t *testing.T)     { nominalArrayMutant(t, "read") }
func TestCheckedViewNominalArrayWriteMutant(t *testing.T)    { nominalArrayMutant(t, "write") }
func TestCheckedViewNominalArrayProducerMutant(t *testing.T) { nominalArrayMutant(t, "producer") }
func nominalArrayMutant(t *testing.T, site string) {
	write := site == "write"
	fixture := "entry-nominal-array-" + site
	program, path := interfaceFixture(t, "nullish/maps/"+fixture)
	truth := onNode(t, path)
	fake := -1
	for i, local := range program.Locals {
		if local.Name == "fake" {
			fake = i
		}
	}
	if fake < 0 {
		t.Fatal("missing fake")
	}
	changed := false
	for i, statement := range program.Main {
		if write {
			if evaluation, ok := statement.(ir.Evaluate); ok {
				if push, ok := evaluation.Value.(ir.ArrayPush); ok {
					push.Value = ir.Read{Local: fake, Of: ir.Object}
					evaluation.Value = push
					program.Main[i] = evaluation
					changed = true
				}
			}
			continue
		}
		declaration, ok := statement.(ir.Declare)
		if !ok || program.Locals[declaration.Local].Name != "item" {
			continue
		}
		array, ok := declaration.Value.(ir.ArrayLiteral)
		if !ok {
			t.Fatal("missing array literal")
		}
		array.Elements[0] = ir.Read{Local: fake, Of: ir.Object}
		declaration.Value = array
		program.Main[i] = declaration
		changed = true
	}
	if !changed {
		t.Fatal("mutation missed")
	}
	native, _ := nativelyUncached(t, program)
	for _, got := range []run{native, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || !strings.Contains(string(got.stderr), "class identity") || (site == "read" && !strings.Contains(string(got.stderr), "field read failed:")) || (site == "producer" && !strings.Contains(string(got.stderr), "Map nominal producer failed:")) {
			t.Fatalf("array lookalike ran on: %#v", got)
		}
	}
	t.Logf("Node=%q; lookalike caught at %s", truth.stdout, site)
}
