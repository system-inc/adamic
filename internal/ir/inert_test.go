package ir

import (
	"reflect"
	"testing"
)

// madeUpExpression is an operation the IR doesn't have, standing for one added later that no proof
// has a case for.
type madeUpExpression struct{}

func (madeUpExpression) Type() Type { return Number }

// Inert fails closed: an operation not on its list, as one added later would be, isn't inert
// (#qk8rztp S2c-2b, Compiler's and Adamic's ruling). Mutant: Inert's last line made `return true`.
func TestInertFailsClosedOnAMadeUpExpression(t *testing.T) {
	t.Parallel()
	if Inert(madeUpExpression{}) {
		t.Fatal("a made-up expression is inert; Inert's default must be closed")
	}
	for _, expression := range []Expression{Call{}, CallClosure{}, GeneratorNext{}, Await{}, Yield{}, DynamicKeyedProperty{}, ReadTextFile{}, NodeFSFile{}} {
		if Inert(expression) {
			t.Errorf("%T is inert, but it runs program code or reaches the host", expression)
		}
	}
	if !Inert(Read{}) || !Inert(Binary{}) {
		t.Fatal("a read and a binary operation must be inert")
	}
}

// Every name on Inert's list is an expression the IR declares, so a rename can't leave a stale entry
// that names nothing while the renamed node falls to the closed default unnoticed.
func TestInertNamesOnlyExpressions(t *testing.T) {
	t.Parallel()
	declared := map[string]bool{}
	for _, value := range gobTypes {
		if _, isExpression := value.(Expression); isExpression {
			declared[reflect.TypeOf(value).Name()] = true
		}
	}
	for name := range inertExpressions {
		if !declared[name] {
			t.Errorf("Inert lists %s, which no IR expression is named", name)
		}
	}
}
