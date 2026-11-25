package lower

import (
	"errors"
	"testing"
)

// The native leg is the plain backend probe table's receiver.ts row.
func TestSwitchNullishNumericReceiver(t *testing.T) {
	t.Parallel()
	source := `enum Kind { First, Second, Other }
interface Entry { kind: Kind }
function receiver(entry: Entry | undefined): string {
    switch (entry?.kind) {
        case Kind.First: return 'first';
        case Kind.Second: return 'second';
        default: return 'other';
    }
}
console.log(receiver(undefined));
console.log(receiver({ kind: Kind.First }));
console.log(receiver({ kind: Kind.Second }));
console.log(receiver({ kind: Kind.Other }));
`
	lowersAndAgreesWithNodeAs(t, source, ".ts")
}

// The native leg is the plain backend probe table's field.ts row.
func TestSwitchNullishNumericField(t *testing.T) {
	t.Parallel()
	source := `enum NewLineKind { CarriageReturnLineFeed, LineFeed }
interface Options { newLine?: NewLineKind | null }
function field(options: Options): string {
    switch (options.newLine) {
        case NewLineKind.CarriageReturnLineFeed: return 'crlf';
        case NewLineKind.LineFeed: return 'lf';
        case undefined: return 'missing';
        default: return 'null';
    }
}
console.log(field({}));
console.log(field({ newLine: NewLineKind.CarriageReturnLineFeed }));
console.log(field({ newLine: NewLineKind.LineFeed }));
console.log(field({ newLine: null }));
`
	lowersAndAgreesWithNodeAs(t, source, ".ts")
}

// The native leg is the plain backend probe table's called.ts row.
func TestSwitchNullishNumericCall(t *testing.T) {
	t.Parallel()
	source := `enum ModuleKind { CommonJS = 1, ESNext = 99 }
function format(choice: number): ModuleKind.ESNext | ModuleKind.CommonJS | undefined {
    if (choice === 0) return ModuleKind.ESNext;
    if (choice === 1) return ModuleKind.CommonJS;
    return undefined;
}
function called(choice: number): string {
    switch (format(choice)) {
        case ModuleKind.ESNext: return 'esnext';
        case ModuleKind.CommonJS: return 'commonjs';
        default: return 'other';
    }
}
console.log(called(0));
console.log(called(1));
console.log(called(2));
`
	lowersAndAgreesWithNodeAs(t, source, ".ts")
}

// The native leg is the plain backend probe table's number.ts row.
func TestSwitchNullishNumericNumber(t *testing.T) {
	t.Parallel()
	source := `function choose(value: number | undefined): string {
    switch (value) {
        case 0: return 'zero';
        default: return 'other';
    }
}
console.log(choose(Number('not a number')));
console.log(choose(-0));
console.log(choose(0));
console.log(choose(undefined));
`
	lowersAndAgreesWithNodeAs(t, source, ".ts")
}

// The native leg is the plain backend probe table's fallthrough.ts row.
func TestSwitchNullishNumericFallthrough(t *testing.T) {
	t.Parallel()
	source := `function choose(value: number | undefined): void {
    switch (value) {
        case undefined:
        case 0:
        case 1:
            console.log('shared');
            break;
        case 2:
            console.log('two');
        default:
            console.log('default');
        case 3:
            console.log('three');
    }
}
choose(undefined);
choose(0);
choose(1);
choose(2);
choose(3);
choose(4);

function middle(value: number | undefined): void {
    switch (value) {
        case 1: console.log('one'); break;
        default: console.log('middle');
        case 2: console.log('last');
    }
}
middle(undefined);
middle(2);
`
	lowersAndAgreesWithNodeAs(t, source, ".ts")
}

// The native leg is the plain backend probe table's exhaustive.ts row.
func TestSwitchNullishNumericExhaustive(t *testing.T) {
	t.Parallel()
	source := `enum Kind { First, Second }
function choose(value: Kind | undefined): void {
    switch (value) {
        case Kind.First: console.log('first'); return;
        case Kind.Second: console.log('second'); return;
    }
    console.log('out');
}
choose(undefined);
choose(Kind.First);
choose(Kind.Second);
`
	lowersAndAgreesWithNodeAs(t, source, ".ts")
}

// The native leg is the plain backend probe table's once.ts row.
func TestSwitchNullishNumericOnce(t *testing.T) {
	t.Parallel()
	source := `function observed(value: number | undefined): number | undefined {
    console.log('evaluated');
    return value;
}
function choose(value: number | undefined): void {
    switch (observed(value)) {
        case 0: console.log('zero'); break;
        case 1: console.log('one'); break;
        default: console.log('default');
    }
}
choose(undefined);
choose(1);
`
	lowersAndAgreesWithNodeAs(t, source, ".ts")
}

// The native leg is the plain backend probe table's label-order.ts row.
func TestSwitchNullishNumericLabelEvaluation(t *testing.T) {
	t.Parallel()
	source := `enum Kind { First, Second }
function labels(name: string): typeof Kind {
    console.log(name);
    return Kind;
}
function choose(value: Kind | undefined): void {
    switch (value) {
        case labels('first label').First: console.log('first body'); break;
        default: console.log('default body'); break;
        case labels('second label').Second: console.log('second body'); break;
    }
}
choose(undefined);
choose(Kind.First);
choose(Kind.Second);
`
	lowersAndAgreesWithNodeAs(t, source, ".ts")
}

// The native leg is the plain backend probe table's label-order-null.ts row.
func TestSwitchNullishNumericLabelEvaluationBoxed(t *testing.T) {
	t.Parallel()
	source := `enum Kind { First, Second }
function labels(name: string): typeof Kind {
    console.log(name);
    return Kind;
}
interface Options { kind?: Kind | null }
function choose(options: Options): void {
    switch (options.kind) {
        case labels('first label').First: console.log('first body'); break;
        default: console.log('default body'); break;
        case labels('second label').Second: console.log('second body'); break;
    }
}
choose({});
choose({ kind: null });
choose({ kind: Kind.First });
choose({ kind: Kind.Second });
`
	lowersAndAgreesWithNodeAs(t, source, ".ts")
}

// The native leg is the plain backend probe table's five control rows.
func TestSwitchNullishNumericControls(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"control-string", `interface Option { type: string | Map<string, number> }
function choose(option: Option): string {
    switch (option.type) {
        case 'string': return 'text';
        default: return 'other';
    }
}
console.log(choose({ type: 'string' }));
`},
		{"control-boolean", `enum Comparison { Less = -1, Equal = 0, Greater = 1 }
function choose(value: boolean | Comparison): string {
    switch (value) {
        case true: return 'true';
        case Comparison.Equal: return 'equal';
        default: return 'other';
    }
}
console.log(choose(true));
`},
		{"control-mixed", `function choose(value: number | string): string {
    switch (value) {
        case 0: return 'zero';
        default: return 'other';
    }
}
console.log(choose(0));
`},
		{"control-nonconstant-label", `function choose(value: number | undefined, limit: number): string {
    switch (value) {
        case limit: return 'limit';
        default: return 'other';
    }
}
console.log(choose(0, 0));
`},
		{"shadowed-undefined", `function choose(value: number | undefined, undefined: number): string {
    switch (value) {
        case undefined: return 'limit';
        default: return 'other';
    }
}
console.log(choose(0, 0));
`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, ".ts")
			var pending *NotYet
			if !errors.As(err, &pending) || pending.What != "a case whose type differs from the switch's" {
				t.Fatalf("want the existing case type stop, got %v", err)
			}
		})
	}
}
