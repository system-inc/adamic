package lower

import (
	"errors"
	"strings"
	"testing"
)

// An ordinary function's receiver is dynamic even when an arrow in its body uses it.
func TestReceiverCaptureKeepsDynamicThisRefused(t *testing.T) {
	t.Parallel()
	for _, body := range []string{"return this.x;", "return (() => this.x)();"} {
		t.Run(body, func(t *testing.T) {
			_, err := lowerSource(t, `const value: (this: {x: number}) => number = function(): number {`+body+`};`)
			var refused *Refused
			if !errors.As(err, &refused) || refused.What != "this in a function expression" {
				t.Fatalf("want dynamic function-expression receiver refused, got %v", err)
			}
		})
	}
}

// Reading a method as a value does not capture its receiver in JavaScript.
func TestReceiverCaptureDetachedMethodStaysRefused(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"compare", "toString"} {
		t.Run(name, func(t *testing.T) {
			_, err := lowerSource(t, `class Receiver { x = 3; `+name+`(): number { return this.x; } }
const receiver = new Receiver(); const callback = receiver.`+name+`; console.log(String(callback()));`)
			if err == nil || !strings.Contains(err.Error(), "would lose its object, and this with it") {
				t.Fatalf("want detached receiver refused, got %v", err)
			}
		})
	}
}
