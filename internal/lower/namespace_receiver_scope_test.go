package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestNamespaceReceiverArrowInheritsScope(test *testing.T) {
	test.Parallel()
	_, err := lowerSource(
		test,
		"namespace N {export const x=7; export function read(this:{readonly x:number}):number {return (()=>this.x)();}}",
	)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "different receivers") {
		test.Fatalf("namespace arrow receiver escaped its scope check: %v", err)
	}
}
