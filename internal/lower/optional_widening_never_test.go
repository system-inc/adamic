package lower

import (
	"errors"
	"strings"
	"testing"
)

// A source field whose type is an intersection tsc reduces to never holds no value, so the optional-widening
// walk can't find a field it hides (#mhggcdw, c4v-7 finding 12: transformers/ts.ts:1054's
// ParameterPropertyDeclaration, an intersection one of whose fields conflicts). Its twin keeps the
// intersection inhabited, with two fields that don't conflict, and the view still hides a, so it stays
// refused. Read is readonly, so only the hiding rule judges the view, and it names kind, so tsc relates
// the inhabited twin to it. Mutant: optionalWidened's reduced-never intersection exemption removed, and
// the never pin goes back to "absent from structural source never".
func TestOptionalWideningNeverIntersection(t *testing.T) {
	t.Parallel()
	const view = "interface Read { readonly inner?: { readonly kind: string; readonly a?: number }; readonly name: string }\nfunction show(read: Read): void { console.log(read.name + ' ' + String(read.inner === undefined)); }\n"
	t.Run("never", func(t *testing.T) {
		t.Parallel()
		source := "type Kinded = { kind: 'a' } & { kind: 'b' };\ninterface Holder { inner?: Kinded; name: string }\n" + view + "const holder: Holder = { name: 'held' };\nshow(holder);\n"
		lowersAndAgreesWithNode(t, source)
	})
	t.Run("inhabited twin", func(t *testing.T) {
		t.Parallel()
		source := "type Kinded = { kind: 'a' } & { tag: 'b' };\ninterface Holder { inner?: Kinded; name: string }\n" + view + "const holder: Holder = { name: 'held' };\nshow(holder);\n"
		_, err := lowerSource(t, source)
		var refused *Refused
		if !errors.As(err, &refused) || !strings.Contains(err.Error(), "refuses optional property a in ") || !strings.Contains(err.Error(), "absent from structural source") {
			t.Errorf("got %v, want the optional-widening refusal of a", err)
		}
	})
}
