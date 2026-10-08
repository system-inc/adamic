package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestNullableObjectSharedStorageStaysNotYet(t *testing.T) {
	for _, source := range []string{
		`interface Item {readonly name:string} const old:readonly (Item|null)[]=[null]; const view:readonly (Item|null|undefined)[]=old;`,
		`interface Item {readonly name:string} const old:{readonly value:Item|null}={value:null}; const view:{readonly value:Item|null|undefined}=old;`,
		`interface Item {readonly name:string} function old():Item|null{return null} const view:()=>Item|null|undefined=old;`,
		`interface Item {readonly name:string} const old:readonly (Item|null)[]=[null]; const copy:readonly (Item|null|undefined)[]=[...old];`,
	} {
		_, err := lowerSource(t, source)
		if err == nil || !(strings.Contains(err.Error(), "sharing nullable object storage") || strings.Contains(err.Error(), "spreading an array of other elements")) {
			t.Fatalf("want explicit storage-view stop, got %v", err)
		}
	}
}

func TestNullableObjectUnionKeepsStrongCyclesRefused(t *testing.T) {
	_, err := lowerSource(t, `interface Link {next:Link|null|undefined} const link:Link={next:null}; link.next=link;`)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "cycle-capable") {
		t.Fatalf("want strong cycle refusal, got %v", err)
	}
}
