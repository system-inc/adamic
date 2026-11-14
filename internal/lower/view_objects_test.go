package lower

import (
	"github.com/system-inc/adamic/internal/ir"
	"strings"
	"testing"
)

func TestViewObjectContractsAreAvailableToEraser(t *testing.T) {
	t.Parallel()
	// .ts, where the cast stays a checked view: in .a a downcast whose checks no proof erases is refused (#m8dvvmf).
	program, err := lowerSourceAs(t, `interface Base {readonly kind:'box'|'other'}
 interface Box extends Base {readonly kind:'box';readonly child:{readonly ready:boolean}}
 function visit(node:Base):boolean {return (node as Box).child.ready}
 const raw={kind:'box' as const,child:{ready:true}}; console.log(String(visit(raw)));`, ".ts")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, function := range program.Functions {
		walk(function.Body, func(node any) bool {
			if property, ok := node.(ir.Property); ok && property.Name == "child" && property.View != "" {
				if property.ViewContract == 0 {
					t.Fatal("object read has no semantic contract for the eraser")
				}
				contract := program.ViewContracts[int(property.ViewContract)-1]
				if contract.Kind != ir.ViewObject || len(contract.Fields) != 1 || contract.Fields[0].Name != "ready" {
					t.Fatalf("incomplete child contract: %#v", contract)
				}
				field := program.ViewContracts[int(contract.Fields[0].Contract)-1]
				if field.Kind != ir.ViewScalar || field.Of != ir.Boolean {
					t.Fatalf("missing scalar child contract: %#v", field)
				}
				found = true
			}
			return true
		})
	}
	if !found {
		t.Fatal("no transitive object read found")
	}
}

// Rule C's source-slot certificate (#2msw6t8): a plain write of view.child lowers, and raw's reads of the
// slot, at a type that doesn't hold the { ready } the view wrote, are checked under their receiver's key,
// so raw.child.other stops loudly where it would have read past the new object (internal/oracle's
// TestSourceSlotCertificate runs it). A destructuring write keeps the stop. .ts, since .a refuses the
// downcast first (#m8dvvmf).
func TestViewObjectWritesNeedSourceCertificate(t *testing.T) {
	t.Parallel()
	source := func(assignment string) string {
		return `interface Base {readonly kind:'box'|'other'}
  interface Box extends Base {readonly kind:'box';child:{ready:boolean}}
  const raw={kind:'box' as const,child:{ready:true,other:42}};
  const held:Base=raw;const view=held as Box;` + assignment + `console.log(String(raw.child.other));`
	}
	program, err := lowerSourceAs(t, source(`view.child={ready:true};`), ".ts")
	if err != nil {
		t.Fatal(err)
	}
	reads := map[string]bool{}
	inspect := func(node any) bool {
		if read, ok := node.(ir.Property); ok && (read.Name == "child" || read.Name == "other") {
			reads[read.Name] = true
			if !program.CheckedRead(read) {
				t.Errorf("raw's read of %s, which may hold the view's narrower write, is unchecked", read.Name)
			}
		}
		return true
	}
	walk(program.Main, inspect)
	for _, function := range program.Functions {
		walk(function.Body, inspect)
	}
	if !reads["child"] || !reads["other"] {
		t.Fatalf("raw.child.other's reads not found: %v", reads)
	}
	_, err = lowerSourceAs(t, source(`({child:view.child}={child:{ready:true}});`), ".ts")
	if err == nil || !strings.Contains(err.Error(), "source-slot type certificate") {
		t.Fatalf("a destructuring write through the view: want the certificate's stop, got %v", err)
	}
}

// The declared part of W (#2msw6t8): the view writes a value raw's type holds, so the writes alone would
// leave raw.child unchecked, but Box declares child narrower than raw reads it, so the read is checked
// under its receiver's key, as certifySourceSlots writes it. No road reaches an absent slot at a type
// narrower than a declaration's, so this program agrees with Node either way (internal/oracle's
// declared_narrower.ts), and this key is the only witness. Mutant: the declared part dropped
// (sourceSlotDeclared not called), and the read goes unchecked.
func TestViewObjectWriteHeldToDeclaredTypes(t *testing.T) {
	t.Parallel()
	program, err := lowerSourceAs(t, `interface Base { readonly kind: "box" | "other" }
interface Box extends Base { readonly kind: "box"; child: { ready: boolean } }
const raw = { kind: "box" as const, child: { ready: true, other: 42 } };
const held: Base = raw;
const view = held as Box;
const fuller = { ready: false, other: 7 };
view.child = fuller;
console.log(String(raw.child.other));
`, ".ts")
	if err != nil {
		t.Fatal(err)
	}
	reads := 0
	inspect := func(node any) bool {
		if read, ok := node.(ir.Property); ok && read.Name == "child" {
			reads++
			if key := ir.CheckedFieldKey(read.ViewReceiverTypeID, read.Name); !program.CheckedFields[key] {
				t.Errorf("raw.child, read at a type Box's declaration of child doesn't fit, isn't checked: no %q in CheckedFields", key)
			}
		}
		return true
	}
	walk(program.Main, inspect)
	for _, function := range program.Functions {
		walk(function.Body, inspect)
	}
	if reads != 1 {
		t.Fatalf("%d reads of child, want raw.child's one", reads)
	}
}
