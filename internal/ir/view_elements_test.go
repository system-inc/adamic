package ir

import (
	"bytes"
	"encoding/gob"
	"reflect"
	"testing"
)

func TestViewElementGobRoundTrip(t *testing.T) {
	t.Parallel()
	RegisterGob()
	want := ViewElement{Array: Read{Local: 3, Of: Array}, Index: NumberConstant{Value: -1}, Relative: true, Contract: 7, Where: "kids[0]", ViewType: "Node", Of: Object, Element: Object, Optional: true}
	var source Expression = want
	var encoded bytes.Buffer
	if err := gob.NewEncoder(&encoded).Encode(&source); err != nil {
		t.Fatal(err)
	}
	var got Expression
	if err := gob.NewDecoder(&encoded).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, source) || got.Type() != Object {
		t.Fatalf("element check lost its operand or contract: %#v", got)
	}
}
