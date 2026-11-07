package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestNullableObjectSharedStorageStaysNotYet(test *testing.T) {
	test.Parallel()
	for _, source := range []string{
		`interface Item {readonly name:string} const old:readonly (Item|null)[]=[null]; const view:readonly (Item|null|undefined)[]=old;`,
		`interface Item {readonly name:string} const old:{readonly value:Item|null}={value:null}; const view:{readonly value:Item|null|undefined}=old;`,
		`interface Item {readonly name:string} function old():Item|null{return null} const view:()=>Item|null|undefined=old;`,
	} {
		_, err := lowerSource(test, source)
		if err == nil || !(strings.Contains(err.Error(), "sharing nullable object storage") ||
			strings.Contains(err.Error(), "spreading an array of other elements")) {
			test.Fatalf("want explicit storage-view stop, got %v", err)
		}
	}
}

func TestNullableObjectUnionUsesGraphOwnership(test *testing.T) {
	test.Parallel()
	const source = `interface Link {next:Link|null|undefined} const link:Link={next:null}; link.next=link;`
	test.Run("untagged lookup stays NotYet", func(test *testing.T) {
		_, err := lowerSource(test, source+
			`interface Item {readonly name:string} const old:readonly (Item|null)[]=[null]; console.log(String(old[0]===null));`)
		var pending *NotYet
		if !errors.As(err, &pending) ||
			pending.What != "a nullable object lookup without separate null/undefined tags" {
			test.Fatalf("want main's nullable lookup tag refusal, got %v", err)
		}
		test.Logf("tag hazard remains refused: %s", pending.What)
	})
	test.Run("graph ownership agrees with Node", func(test *testing.T) {
		graphAccepted(test, source)
		graphAccepted(test, source+`console.log(String(link.next===link));`)
		test.Log("original and observable cycle agree with Node, JavaScript, sanitized native, and leak checks")
	})
}

func TestNullableObjectSpreadKeepsNarrowerLiteralRefused(test *testing.T) {
	test.Parallel()
	_, err := lowerSource(
		test,
		`interface Item {readonly name:string} const old:readonly (Item|null)[]=[null]; const copy:readonly (Item|null|undefined)[]=[...old];`,
	)
	var refusal *Refused
	if !errors.As(err, &refusal) ||
		refusal.What != "union-typed array built from a narrower literal" ||
		refusal.Fix != "declare the literal's element type explicitly, or build the array with a mixed literal" {
		test.Fatalf("want the newer narrow-literal layout refusal with its fix, got %v", err)
	}
}
