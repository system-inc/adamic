package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestNullableObjectSharedStorageStaysNotYet(t *testing.T) {
	t.Parallel()
	for _, source := range []string{
		`interface Item {readonly name:string} const old:readonly (Item|null)[]=[null]; const view:readonly (Item|null|undefined)[]=old;`,
		`interface Item {readonly name:string} const old:{readonly value:Item|null}={value:null}; const view:{readonly value:Item|null|undefined}=old;`,
		`interface Item {readonly name:string} function old():Item|null{return null} const view:()=>Item|null|undefined=old;`,
	} {
		_, err := lowerSource(t, source)
		if err == nil || !(strings.Contains(err.Error(), "sharing nullable object storage") || strings.Contains(err.Error(), "spreading an array of other elements")) {
			t.Fatalf("want explicit storage-view stop, got %v", err)
		}
	}
}

func TestNullableObjectUnionUsesGraphOwnership(t *testing.T) {
	t.Parallel()
	const source = `interface Link {next:Link|null|undefined} const link:Link={next:null}; link.next=link;`
	t.Run("untagged lookup stays NotYet", func(t *testing.T) {
		_, err := lowerSource(t, source+`interface Item {readonly name:string} const old:readonly (Item|null)[]=[null]; console.log(String(old[0]===null));`)
		var pending *NotYet
		if !errors.As(err, &pending) || pending.What != "a nullable object lookup without separate null/undefined tags" {
			t.Fatalf("want main's nullable lookup tag refusal, got %v", err)
		}
		t.Logf("tag hazard remains refused: %s", pending.What)
	})
	t.Run("graph ownership agrees with Node", func(t *testing.T) {
		graphAccepted(t, source)
		graphAccepted(t, source+`console.log(String(link.next===link));`)
		t.Log("original and observable cycle agree with Node, JavaScript, sanitized native, and leak checks")
	})
}

func TestNullableObjectSpreadKeepsNarrowerLiteralRefused(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `interface Item {readonly name:string} const old:readonly (Item|null)[]=[null]; const copy:readonly (Item|null|undefined)[]=[...old];`)
	var refusal *Refused
	if !errors.As(err, &refusal) || refusal.What != "union-typed array built from a narrower literal" || refusal.Fix != "declare the literal's element type explicitly, or build the array with a mixed literal" {
		t.Fatalf("want the newer narrow-literal layout refusal with its fix, got %v", err)
	}
}
