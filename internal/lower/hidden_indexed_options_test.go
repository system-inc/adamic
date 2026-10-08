package lower

import (
	"errors"
	"testing"
)

func TestHiddenIndexedOptionsKeepsUnsupportedReads(t *testing.T) {
	for _, source := range []string{
		`function read(value:{[key:string]:boolean}, key:string):boolean|undefined { return value[key]; }`,
		`const value={get first():boolean{return true;}}; const result=value["first"];`,
		`const value={first():boolean{return true;}}; const f=value["first"];`,
		`const value={toString:true}; const result=value["toString"];`,
	} {
		_, err := lowerSource(t, source)
		var notYet *NotYet
		var refused *Refused
		if !errors.As(err, &notYet) && !errors.As(err, &refused) {
			t.Fatalf("want unsupported indexed read to stop, got %v for %s", err, source)
		}
	}
}
