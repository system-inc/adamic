package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestGeneratorNextPromiseNeedsResumeProof(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `function* f():Generator<number,number,number>{const n=yield 1;return n;}const g=f();g.next();g.next();`)
	var gap *NotYet
	if !errors.As(err, &gap) || !strings.Contains(gap.What, "next argument is not proved present") {
		t.Fatalf("got %v", err)
	}
}
func TestGeneratorPrimitiveThrowRemainsExplicit(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `function* f():Generator<number,number,undefined>{yield 1;return 2;}const g=f();g.next();g.throw(3);`)
	var gap *NotYet
	if !errors.As(err, &gap) || !strings.Contains(gap.What, "represented Error completion protocol") {
		t.Fatalf("got %v", err)
	}
}
func TestGeneratorOpaqueReceiverRemainsExplicit(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `function consume(g:Generator<number,number,undefined>):void{g.next();}function* f():Generator<number,number,undefined>{yield 1;return 2;}consume(f());`)
	var gap *NotYet
	if !errors.As(err, &gap) || !strings.Contains(gap.What, "proved generator factory origin") {
		t.Fatalf("got %v", err)
	}
}
func TestGeneratorNestedFrameCaptureRemainsExplicit(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `function* f():Generator<number,number,undefined>{const x=1;const read=()=>x;yield read();return 2;}const g=f();g.next();`)
	var gap *NotYet
	if !errors.As(err, &gap) || !strings.Contains(gap.What, "nested capture x needs a weak frame reference") {
		t.Fatalf("got %v", err)
	}
}
func TestGeneratorEmptyBodyUsesOwnedFrame(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `function* empty(){return 1;}empty();const expression=function*(){return 2;};expression();const source={*empty(){return 3;}};source.empty();`)
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, function := range program.Functions {
		if function.Generator != nil {
			count++
		}
	}
	if count != 3 {
		t.Fatalf("got %d generator factories, want 3", count)
	}
}

func TestGeneratorProtocolReplacementRemainsExplicit(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `function* f():Generator<number,number,undefined>{yield 1;return 2;}const g=f();g.next=()=>({value:3,done:false});g.next();`)
	var gap *NotYet
	var refused *Refused
	if errors.As(err, &gap) && strings.Contains(gap.What, "protocol replacement") {
		return
	}
	if errors.As(err, &refused) && strings.Contains(refused.What, "method") {
		return
	}
	t.Fatalf("got %v, want an explicit method-replacement stop", err)
}
