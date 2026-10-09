package typescript

import (
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"

	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// switchExhaustivenessFile is the fixture name every case in this file runs under.
//
// A TypeScript extension is required rather than incidental: nearly every case declares a union
// type or an enum, so a `.js` name would make most of these parse errors rather than rule inputs.
const switchExhaustivenessFile = "switchExhaustiveness.ts"

// TestSwitchExhaustivenessCheckStaysSilent carries forty-nine of tsgolint's fifty-four valid cases.
//
// Extracted from tsgolint's own test file by parsing it with go/ast rather than by transcribing it,
// because a fixture written by hand encodes the same belief as the port and would pass for exactly
// the reason the code would be wrong. Every source string and every option value below came out of
// that parse.
//
// FIVE of upstream's valid cases are excluded, and TestSwitchExhaustivenessCheckExclusionsAreStated
// names each one and why, so that this is a stated trim rather than a quiet one.
//
// The set is more pointed than its size suggests, because the OPTIONS are half the corpus: the same
// switch appears here as silent and in the invalid table as reporting, differing only in which
// option was set. A port that ignored options entirely would fail both tables rather than neither.
func TestSwitchExhaustivenessCheckStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		options    any
	}{
		{"upstream valid 0 [defaults]",
			`
type Day =
  | 'Monday'
  | 'Tuesday'
  | 'Wednesday'
  | 'Thursday'
  | 'Friday'
  | 'Saturday'
  | 'Sunday';

const day = 'Monday' as Day;
let result = 0;

switch (day) {
  case 'Monday': {
    result = 1;
    break;
  }
  case 'Tuesday': {
    result = 2;
    break;
  }
  case 'Wednesday': {
    result = 3;
    break;
  }
  case 'Thursday': {
    result = 4;
    break;
  }
  case 'Friday': {
    result = 5;
    break;
  }
  case 'Saturday': {
    result = 6;
    break;
  }
  case 'Sunday': {
    result = 7;
    break;
  }
}
    `,
			nil},
		{"upstream valid 1 [defaults]",
			`
type Num = 0 | 1 | 2;

function test(value: Num): number {
  switch (value) {
    case 0:
      return 0;
    case 1:
      return 1;
    case 2:
      return 2;
  }
}
    `,
			nil},
		{"upstream valid 2 [defaults]",
			`
type Bool = true | false;

function test(value: Bool): number {
  switch (value) {
    case true:
      return 1;
    case false:
      return 0;
  }
}
    `,
			nil},
		{"upstream valid 3 [defaults]",
			`
type Mix = 0 | 1 | 'two' | 'three' | true;

function test(value: Mix): number {
  switch (value) {
    case 0:
      return 0;
    case 1:
      return 1;
    case 'two':
      return 2;
    case 'three':
      return 3;
    case true:
      return 4;
  }
}
    `,
			nil},
		{"upstream valid 4 [defaults]",
			`
type A = 'a';
type B = 'b';
type C = 'c';
type Union = A | B | C;

function test(value: Union): number {
  switch (value) {
    case 'a':
      return 1;
    case 'b':
      return 2;
    case 'c':
      return 3;
  }
}
    `,
			nil},
		{"upstream valid 5 [defaults]",
			`
const A = 'a';
const B = 1;
const C = true;

type Union = typeof A | typeof B | typeof C;

function test(value: Union): number {
  switch (value) {
    case 'a':
      return 1;
    case 1:
      return 2;
    case true:
      return 3;
  }
}
    `,
			nil},
		{"upstream valid 6 [considerDefault=true]",
			`
type Day =
  | 'Monday'
  | 'Tuesday'
  | 'Wednesday'
  | 'Thursday'
  | 'Friday'
  | 'Saturday'
  | 'Sunday';

const day = 'Monday' as Day;
let result = 0;

switch (day) {
  case 'Monday': {
    result = 1;
    break;
  }
  default: {
    result = 42;
  }
}
      `,
			SwitchExhaustivenessCheckOptions{ConsiderDefaultExhaustiveForUnions: type_checking.Ref(true)}},
		{"upstream valid 7 [defaults]",
			`
const day = 'Monday' as string;
let result = 0;

switch (day) {
  case 'Monday': {
    result = 1;
    break;
  }
  case 'Tuesday': {
    result = 2;
    break;
  }
}
    `,
			nil},
		{"upstream valid 8 [defaults]",
			`
enum Enum {
  A,
  B,
}

function test(value: Enum): number {
  switch (value) {
    case Enum.A:
      return 1;
    case Enum.B:
      return 2;
  }
}
    `,
			nil},
		{"upstream valid 9 [defaults]",
			`
type ObjectUnion = { a: 1 } | { b: 2 };

function test(value: ObjectUnion): number {
  switch (value.a) {
    case 1:
      return 1;
  }
}
    `,
			nil},
		{"upstream valid 10 [allowDefault=true, requireDefault=true]",
			`
declare const value: number;
switch (value) {
  case 0:
    return 0;
  case 1:
    return 1;
  default:
    return -1;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(true), RequireDefaultForNonUnion: type_checking.Ref(true)}},
		{"upstream valid 11 [allowDefault=false, requireDefault=false]",
			`
declare const value: string;
switch (value) {
  case 'foo':
    return 0;
  case 'bar':
    return 1;
  default:
    return -1;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(false)}},
		{"upstream valid 12 [allowDefault=false, requireDefault=false]",
			`
declare const value: number;
switch (value) {
  case 0:
    return 0;
  case 1:
    return 1;
  default:
    return -1;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(false)}},
		{"upstream valid 13 [allowDefault=false, requireDefault=false]",
			`
declare const value: bigint;
switch (value) {
  case 0:
    return 0;
  case 1:
    return 1;
  default:
    return -1;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(false)}},
		{"upstream valid 14 [allowDefault=false, requireDefault=false]",
			`
declare const value: symbol;
const foo = Symbol('foo');
switch (value) {
  case foo:
    return 0;
  default:
    return -1;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(false)}},
		{"upstream valid 15 [allowDefault=false, requireDefault=false]",
			`
declare const value: 0 | 1 | number;
switch (value) {
  case 0:
    return 0;
  case 1:
    return 1;
  default:
    return -1;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(false)}},
		{"upstream valid 16 [allowDefault=false, requireDefault=true]",
			`
declare const value: 'literal';
switch (value) {
  case 'literal':
    return 0;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(true)}},
		{"upstream valid 17 [allowDefault=false, requireDefault=true]",
			`
declare const value: null;
switch (value) {
  case null:
    return 0;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(true)}},
		{"upstream valid 18 [allowDefault=false, requireDefault=true]",
			`
declare const value: undefined;
switch (value) {
  case undefined:
    return 0;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(true)}},
		{"upstream valid 19 [allowDefault=false, requireDefault=true]",
			`
declare const value: null | undefined;
switch (value) {
  case null:
    return 0;
  case undefined:
    return 0;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(true)}},
		{"upstream valid 20 [allowDefault=false, requireDefault=true]",
			`
declare const value: 'literal' & { _brand: true };
switch (value) {
  case 'literal':
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(true)}},
		{"upstream valid 21 [allowDefault=false, requireDefault=true]",
			`
declare const value: ('literal' & { _brand: true }) | 1;
switch (value) {
  case 'literal':
    break;
  case 1:
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(true)}},
		{"upstream valid 22 [allowDefault=false, requireDefault=true]",
			`
declare const value: (1 & { _brand: true }) | 'literal' | null;
switch (value) {
  case 'literal':
    break;
  case 1:
    break;
  case null:
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(true)}},
		{"upstream valid 23 [allowDefault=true, requireDefault=false]",
			`
declare const value: '1' | '2' | number;
switch (value) {
  case '1':
    break;
  case '2':
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(true), RequireDefaultForNonUnion: type_checking.Ref(false)}},
		{"upstream valid 24 [allowDefault=true, requireDefault=false]",
			`
declare const value: '1' | '2' | number;
switch (value) {
  case '1':
    break;
  case '2':
    break;
  default:
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(true), RequireDefaultForNonUnion: type_checking.Ref(false)}},
		{"upstream valid 25 [allowDefault=false, requireDefault=false]",
			`
declare const value: '1' | '2' | number;
switch (value) {
  case '1':
    break;
  case '2':
    break;
  default:
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(false)}},
		{"upstream valid 26 [allowDefault=true, requireDefault=false]",
			`
declare const value: '1' | '2' | (number & { foo: 'bar' });
switch (value) {
  case '1':
    break;
  case '2':
    break;
  default:
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(true), RequireDefaultForNonUnion: type_checking.Ref(false)}},
		{"upstream valid 27 [allowDefault=true, requireDefault=true]",
			`
declare const value: '1' | '2' | number;
switch (value) {
  case '1':
    break;
  case '2':
    break;
  default:
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(true), RequireDefaultForNonUnion: type_checking.Ref(true)}},
		{"upstream valid 28 [allowDefault=true, considerDefault=true, requireDefault=false]",
			`
declare const value: number | null | undefined;
switch (value) {
  case null:
    break;
  case undefined:
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(true), ConsiderDefaultExhaustiveForUnions: type_checking.Ref(true), RequireDefaultForNonUnion: type_checking.Ref(false)}},
		{"upstream valid 29 [allowDefault=false, considerDefault=true, requireDefault=false]",
			`
declare const value: '1' | '2' | number;
switch (value) {
  case '1':
    break;
  default:
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), ConsiderDefaultExhaustiveForUnions: type_checking.Ref(true), RequireDefaultForNonUnion: type_checking.Ref(false)}},
		{"upstream valid 30 [allowDefault=true, requireDefault=false]",
			`
declare const value: (string & { foo: 'bar' }) | '1';
switch (value) {
  case '1':
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(true), RequireDefaultForNonUnion: type_checking.Ref(false)}},
		{"upstream valid 31 [allowDefault=false, requireDefault=true]",
			`
const a = Symbol('a');
declare const value: typeof a | 2;
switch (value) {
  case a:
    break;
  case 2:
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(true)}},
		{"upstream valid 32 [allowDefault=false, requireDefault=false]",
			`
declare const value: string | number;
switch (value) {
  case 1:
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(false)}},
		{"upstream valid 33 [allowDefault=true, requireDefault=false]",
			`
declare const value: string | number;
switch (value) {
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(true), RequireDefaultForNonUnion: type_checking.Ref(false)}},
		{"upstream valid 34 [allowDefault=false, requireDefault=true]",
			`
declare const value: string | number;
switch (value) {
  default:
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(true)}},
		{"upstream valid 35 [allowDefault=false, requireDefault=false]",
			`
declare const value: number;
declare const a: number;
switch (value) {
  case a:
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(false)}},
		{"upstream valid 36 [allowDefault=true, requireDefault=false]",
			`
declare const value: bigint;
switch (value) {
  case 10n:
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(true), RequireDefaultForNonUnion: type_checking.Ref(false)}},
		{"upstream valid 37 [allowDefault=true, requireDefault=false]",
			`
declare const value: symbol;
const a = Symbol('a');
switch (value) {
  case a:
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(true), RequireDefaultForNonUnion: type_checking.Ref(false)}},
		{"upstream valid 38 [allowDefault=true, requireDefault=true]",
			`
declare const value: symbol;
const a = Symbol('a');
switch (value) {
  case a:
    break;
  default:
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(true), RequireDefaultForNonUnion: type_checking.Ref(true)}},
		{"upstream valid 39 [allowDefault=true, requireDefault=true]",
			`
const a = Symbol('a');
declare const value: typeof a | string;
switch (value) {
  case a:
    break;
  default:
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(true), RequireDefaultForNonUnion: type_checking.Ref(true)}},
		{"upstream valid 40 [allowDefault=true, considerDefault=true, requireDefault=true]",
			`
const a = Symbol('a');
declare const value: typeof a | string;
switch (value) {
  default:
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(true), ConsiderDefaultExhaustiveForUnions: type_checking.Ref(true), RequireDefaultForNonUnion: type_checking.Ref(true)}},
		{"upstream valid 41 [allowDefault=false, considerDefault=true, requireDefault=true]",
			`
declare const value: boolean | 1;
switch (value) {
  case 1:
    break;
  default:
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), ConsiderDefaultExhaustiveForUnions: type_checking.Ref(true), RequireDefaultForNonUnion: type_checking.Ref(true)}},
		{"upstream valid 42 [allowDefault=true, requireDefault=false]",
			`
declare const value: boolean | 1;
switch (value) {
  case 1:
    break;
  case true:
    break;
  case false:
    break;
  default:
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(true), RequireDefaultForNonUnion: type_checking.Ref(false)}},
		{"upstream valid 43 [allowDefault=true, requireDefault=false]",
			`
enum Aaa {
  Foo,
  Bar,
}
declare const value: Aaa | 1;
switch (value) {
  case 1:
    break;
  case Aaa.Foo:
    break;
  case Aaa.Bar:
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(true), RequireDefaultForNonUnion: type_checking.Ref(false)}},
		{"upstream valid 44 [considerDefault=true, requireDefault=true]",
			`
declare const literal: 'a' | 'b';
switch (literal) {
  case 'a':
    break;
  case 'b':
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{ConsiderDefaultExhaustiveForUnions: type_checking.Ref(true), RequireDefaultForNonUnion: type_checking.Ref(true)}},
		{"upstream valid 45 [considerDefault=true]",
			`
declare const literal: 'a' | 'b';
switch (literal) {
  case 'a':
    break;
  default:
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{ConsiderDefaultExhaustiveForUnions: type_checking.Ref(true)}},
		{"upstream valid 46 [allowDefault=false]",
			`
declare const literal: 'a' | 'b';
switch (literal) {
  case 'a':
    break;
  case 'b':
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false)}},
		{"upstream valid 47 [considerDefault=true]",
			`
enum MyEnum {
  Foo = 'Foo',
  Bar = 'Bar',
  Baz = 'Baz',
}

declare const myEnum: MyEnum;

switch (myEnum) {
  case MyEnum.Foo:
    break;
  case MyEnum.Bar:
    break;
  default: {
    break;
  }
}
      `,
			SwitchExhaustivenessCheckOptions{ConsiderDefaultExhaustiveForUnions: type_checking.Ref(true)}},
		{"upstream valid 48 [considerDefault=true]",
			`
declare const value: boolean;
switch (value) {
  case false:
    break;
  default: {
    break;
  }
}
      `,
			SwitchExhaustivenessCheckOptions{ConsiderDefaultExhaustiveForUnions: type_checking.Ref(true)}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(
				t, SwitchExhaustivenessCheck, switchExhaustivenessFile, testCase.sourceText, testCase.options,
			))
		})
	}
}

// TestSwitchExhaustivenessCheckFires carries forty-seven of tsgolint's fifty-two invalid cases,
// asserting fifty-one findings in upstream's own order. The fifty-second case needs a second file
// and has its own test below.
//
// FOUR of these inputs report TWICE on one switch statement, once for a missing literal branch and
// once for the missing default, because `checkSwitchExhaustive` and `checkSwitchNoUnionDefaultCase`
// are independent checks over one shared metadata value and both can fire on the same node. A
// fixture asserting a single id per input would have passed while silently dropping half the
// findings, which is why this asserts an ordered list rather than a presence.
//
// `addMissingCases` never appears here, and cannot: it is tsgolint's suggestion message and
// tsgolint ships no suggestions. Only `switchIsNotExhaustive` and `dangerousDefaultCase` are
// reachable, which TestSwitchExhaustivenessCheckShipsNoRepairs pins from the other direction.
//
// Two of these sources are written as interpreted Go strings rather than raw ones, because the
// TypeScript they contain has BACKTICKS in an enum member name and a Go raw literal cannot hold
// one. Upstream writes them as concatenations for the same reason. They are here rather than
// dropped because an enum key that is a template literal, or one containing an escaped quote, is
// exactly the shape a name-based implementation would mishandle and a type-identity-based one
// handles for free.
func TestSwitchExhaustivenessCheckFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		options    any
		wantIds    []string
	}{
		{"upstream invalid 0 [allowDefault=false, requireDefault=true]",
			`
declare const value: 'literal';
switch (value) {
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(true)},
			[]string{"switchIsNotExhaustive"}},
		{"upstream invalid 1 [allowDefault=false, requireDefault=true]",
			`
declare const value: 'literal' & { _brand: true };
switch (value) {
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(true)},
			[]string{"switchIsNotExhaustive"}},
		{"upstream invalid 2 [allowDefault=false, requireDefault=true]",
			`
declare const value: ('literal' & { _brand: true }) | 1;
switch (value) {
  case 'literal':
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(true)},
			[]string{"switchIsNotExhaustive"}},
		{"upstream invalid 3 [allowDefault=true, requireDefault=false]",
			`
declare const value: '1' | '2' | number;
switch (value) {
  case '1':
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(true), RequireDefaultForNonUnion: type_checking.Ref(false)},
			[]string{"switchIsNotExhaustive"}},
		{"upstream invalid 4 [allowDefault=true, requireDefault=true]",
			`
declare const value: '1' | '2' | number;
switch (value) {
  case '1':
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(true), RequireDefaultForNonUnion: type_checking.Ref(true)},
			[]string{"switchIsNotExhaustive", "switchIsNotExhaustive"}},
		{"upstream invalid 5 [allowDefault=true, requireDefault=true]",
			`
declare const value: (string & { foo: 'bar' }) | '1';
switch (value) {
  case '1':
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(true), RequireDefaultForNonUnion: type_checking.Ref(true)},
			[]string{"switchIsNotExhaustive"}},
		{"upstream invalid 6 [allowDefault=false, requireDefault=true]",
			`
declare const value: (string & { foo: 'bar' }) | '1' | 1 | null | undefined;
switch (value) {
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(true)},
			[]string{"switchIsNotExhaustive", "switchIsNotExhaustive"}},
		{"upstream invalid 7 [allowDefault=false, requireDefault=true]",
			`
declare const value: string | number;
switch (value) {
  case 1:
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(true)},
			[]string{"switchIsNotExhaustive"}},
		{"upstream invalid 8 [allowDefault=false, requireDefault=true]",
			`
declare const value: number;
declare const a: number;
switch (value) {
  case a:
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(true)},
			[]string{"switchIsNotExhaustive"}},
		{"upstream invalid 9 [allowDefault=false, requireDefault=true]",
			`
declare const value: bigint;
switch (value) {
  case 10n:
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(true)},
			[]string{"switchIsNotExhaustive"}},
		{"upstream invalid 10 [allowDefault=false, requireDefault=true]",
			`
declare const value: symbol;
const a = Symbol('a');
switch (value) {
  case a:
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(true)},
			[]string{"switchIsNotExhaustive"}},
		{"upstream invalid 11 [allowDefault=false, requireDefault=true]",
			`
const a = Symbol('aa');
const b = Symbol('bb');
declare const value: typeof a | typeof b | 1;
switch (value) {
  case 1:
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(true)},
			[]string{"switchIsNotExhaustive"}},
		{"upstream invalid 12 [allowDefault=false, requireDefault=true]",
			`
const a = Symbol('a');
declare const value: typeof a | string;
switch (value) {
  case a:
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(true)},
			[]string{"switchIsNotExhaustive"}},
		{"upstream invalid 13 [allowDefault=false, requireDefault=false]",
			`
declare const value: boolean;
switch (value) {
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(false)},
			[]string{"switchIsNotExhaustive"}},
		{"upstream invalid 14 [allowDefault=false, requireDefault=true]",
			`
declare const value: boolean | 1;
switch (value) {
  case false:
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(true)},
			[]string{"switchIsNotExhaustive"}},
		{"upstream invalid 15 [allowDefault=false, requireDefault=true]",
			`
declare const value: boolean | number;
switch (value) {
  case 1:
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(true)},
			[]string{"switchIsNotExhaustive", "switchIsNotExhaustive"}},
		{"upstream invalid 16 [allowDefault=false, requireDefault=true]",
			`
declare const value: object;
switch (value) {
  case 1:
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(true)},
			[]string{"switchIsNotExhaustive"}},
		{"upstream invalid 17 [allowDefault=true, requireDefault=true]",
			`
enum Aaa {
  Foo,
  Bar,
}
declare const value: Aaa | 1 | string;
switch (value) {
  case 1:
    break;
  case Aaa.Foo:
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(true), RequireDefaultForNonUnion: type_checking.Ref(true)},
			[]string{"switchIsNotExhaustive", "switchIsNotExhaustive"}},
		{"upstream invalid 18 [defaults]",
			`
type Day =
  | 'Monday'
  | 'Tuesday'
  | 'Wednesday'
  | 'Thursday'
  | 'Friday'
  | 'Saturday'
  | 'Sunday';

const day = 'Monday' as Day;
let result = 0;

switch (day) {
  case 'Monday': {
    result = 1;
    break;
  }
}
      `,
			nil,
			[]string{"switchIsNotExhaustive"}},
		{"upstream invalid 19 [defaults]",
			`
enum Enum {
  A,
  B,
}

function test(value: Enum): number {
  switch (value) {
    case Enum.A:
      return 1;
  }
}
      `,
			nil,
			[]string{"switchIsNotExhaustive"}},
		{"upstream invalid 20 [defaults]",
			`
type A = 'a';
type B = 'b';
type C = 'c';
type Union = A | B | C;

function test(value: Union): number {
  switch (value) {
    case 'a':
      return 1;
  }
}
      `,
			nil,
			[]string{"switchIsNotExhaustive"}},
		{"upstream invalid 21 [defaults]",
			`
const A = 'a';
const B = 1;
const C = true;

type Union = typeof A | typeof B | typeof C;

function test(value: Union): number {
  switch (value) {
    case 'a':
      return 1;
  }
}
      `,
			nil,
			[]string{"switchIsNotExhaustive"}},
		{"upstream invalid 22 [defaults]",
			`
type DiscriminatedUnion = { type: 'A'; a: 1 } | { type: 'B'; b: 2 };

function test(value: DiscriminatedUnion): number {
  switch (value.type) {
    case 'A':
      return 1;
  }
}
      `,
			nil,
			[]string{"switchIsNotExhaustive"}},
		{"upstream invalid 23 [defaults]",
			`
type Day =
  | 'Monday'
  | 'Tuesday'
  | 'Wednesday'
  | 'Thursday'
  | 'Friday'
  | 'Saturday'
  | 'Sunday';

const day = 'Monday' as Day;

switch (day) {
}
      `,
			nil,
			[]string{"switchIsNotExhaustive"}},
		{"upstream invalid 24 [defaults]",
			`
const a = Symbol('a');
const b = Symbol('b');
const c = Symbol('c');

type T = typeof a | typeof b | typeof c;

function test(value: T): number {
  switch (value) {
    case a:
      return 1;
  }
}
      `,
			nil,
			[]string{"switchIsNotExhaustive"}},
		{"upstream invalid 25 [defaults]",
			`
type T = 1 | 2;

function test(value: T): number {
  switch (value) {
    case 1:
      return 1;
  }
}
      `,
			nil,
			[]string{"switchIsNotExhaustive"}},
		{"upstream invalid 26 [defaults]",
			`
type T = 1 | 2;

function test(value: T): number {
  switch (value) {
  }
}
      `,
			nil,
			[]string{"switchIsNotExhaustive"}},
		{"upstream invalid 27 [defaults]",
			`
export enum Enum {
  'test-test' = 'test-test',
  'test' = 'test',
}

function test(arg: Enum): string {
  switch (arg) {
  }
}
      `,
			nil,
			[]string{"switchIsNotExhaustive"}},
		{"upstream invalid 28 [defaults]",
			`
export enum Enum {
  '' = 'test-test',
  'test' = 'test',
}

function test(arg: Enum): string {
  switch (arg) {
  }
}
      `,
			nil,
			[]string{"switchIsNotExhaustive"}},
		{"upstream invalid 29 [defaults]",
			`
export enum Enum {
  '9test' = 'test-test',
  'test' = 'test',
}

function test(arg: Enum): string {
  switch (arg) {
  }
}
      `,
			nil,
			[]string{"switchIsNotExhaustive"}},
		{"upstream invalid 30 [allowDefault=true, requireDefault=true]",
			`
const value: number = Math.floor(Math.random() * 3);
switch (value) {
  case 0:
    return 0;
  case 1:
    return 1;
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(true), RequireDefaultForNonUnion: type_checking.Ref(true)},
			[]string{"switchIsNotExhaustive"}},
		{"upstream invalid 31 [defaults]",
			"\n        enum Enum {\n          'a' = 1,\n          [`key-with\n\n          new-line`] = 2,\n        }\n\n        declare const a: Enum;\n\n        switch (a) {\n        }\n      ",
			nil,
			[]string{"switchIsNotExhaustive"}},
		{"upstream invalid 32 [defaults]",
			"\n        enum Enum {\n          'a' = 1,\n          \"'a' `b` \\\"c\\\"\" = 2,\n        }\n\n        declare const a: Enum;\n\n        switch (a) {}\n      ",
			nil,
			[]string{"switchIsNotExhaustive"}},
		{"upstream invalid 33 [allowDefault=false, requireDefault=false]",
			`
type MyUnion = 'foo' | 'bar' | 'baz';

declare const myUnion: MyUnion;

switch (myUnion) {
  case 'foo':
  case 'bar':
  case 'baz': {
    break;
  }
  default: {
    break;
  }
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(false)},
			[]string{"dangerousDefaultCase"}},
		{"upstream invalid 34 [allowDefault=false, requireDefault=false]",
			`
enum MyEnum {
  Foo = 'Foo',
  Bar = 'Bar',
  Baz = 'Baz',
}

declare const myEnum: MyEnum;

switch (myEnum) {
  case MyEnum.Foo:
  case MyEnum.Bar:
  case MyEnum.Baz: {
    break;
  }
  default: {
    break;
  }
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(false)},
			[]string{"dangerousDefaultCase"}},
		{"upstream invalid 35 [allowDefault=false, requireDefault=false]",
			`
enum MyEnum {
  Foo,
  Bar,
  Baz,
}

declare const myEnum: MyEnum;

switch (myEnum) {
  case MyEnum.Foo:
  case MyEnum.Bar:
  case MyEnum.Baz: {
    break;
  }
  default: {
    break;
  }
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(false)},
			[]string{"dangerousDefaultCase"}},
		{"upstream invalid 36 [allowDefault=false, requireDefault=false]",
			`
declare const myBoolean: boolean;

switch (myBoolean) {
  case true:
  case false: {
    break;
  }
  default: {
    break;
  }
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(false)},
			[]string{"dangerousDefaultCase"}},
		{"upstream invalid 37 [allowDefault=false, requireDefault=false]",
			`
declare const myValue: undefined;

switch (myValue) {
  case undefined: {
    break;
  }

  default: {
    break;
  }
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(false)},
			[]string{"dangerousDefaultCase"}},
		{"upstream invalid 38 [allowDefault=false, requireDefault=false]",
			`
declare const myValue: null;

switch (myValue) {
  case null: {
    break;
  }

  default: {
    break;
  }
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(false)},
			[]string{"dangerousDefaultCase"}},
		{"upstream invalid 39 [allowDefault=false, requireDefault=false]",
			`
declare const myValue: 'foo' | boolean | undefined | null;

switch (myValue) {
  case 'foo':
  case true:
  case false:
  case undefined:
  case null: {
    break;
  }

  default: {
    break;
  }
}
      `,
			SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false), RequireDefaultForNonUnion: type_checking.Ref(false)},
			[]string{"dangerousDefaultCase"}},
		{"upstream invalid 40 [considerDefault=false]",
			`
declare const literal: 'a' | 'b';

switch (literal) {
  case 'a':
    break;
  default:
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{ConsiderDefaultExhaustiveForUnions: type_checking.Ref(false)},
			[]string{"switchIsNotExhaustive"}},
		{"upstream invalid 41 [defaults]",
			`
declare const literal: 'a' | 'b';

switch (literal) {
  case 'a':
    break;
}
`,
			nil,
			[]string{"switchIsNotExhaustive"}},
		{"upstream invalid 42 [considerDefault=false]",
			`
declare const literal: 'a' | 'b';

switch (literal) {
  default:
  case 'a':
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{ConsiderDefaultExhaustiveForUnions: type_checking.Ref(false)},
			[]string{"switchIsNotExhaustive"}},
		{"upstream invalid 43 [considerDefault=false]",
			`
declare const literal: 'a' | 'b' | 'c';

switch (literal) {
  case 'a':
    break;
  default:
    break;
}
      `,
			SwitchExhaustivenessCheckOptions{ConsiderDefaultExhaustiveForUnions: type_checking.Ref(false)},
			[]string{"switchIsNotExhaustive"}},
		{"upstream invalid 44 [considerDefault=false]",
			`
enum MyEnum {
  Foo = 'Foo',
  Bar = 'Bar',
  Baz = 'Baz',
}

declare const myEnum: MyEnum;

switch (myEnum) {
  case MyEnum.Foo:
    break;
  default: {
    break;
  }
}
      `,
			SwitchExhaustivenessCheckOptions{ConsiderDefaultExhaustiveForUnions: type_checking.Ref(false)},
			[]string{"switchIsNotExhaustive"}},
		{"upstream invalid 45 [considerDefault=false]",
			`
declare const value: boolean;
switch (value) {
  default: {
    break;
  }
}
      `,
			SwitchExhaustivenessCheckOptions{ConsiderDefaultExhaustiveForUnions: type_checking.Ref(false)},
			[]string{"switchIsNotExhaustive"}},
		{"upstream invalid 50 [defaults]",
			`
        export namespace A {
          export enum B {
            C,
            D,
          }
        }
        declare const foo: A.B;
        switch (foo) {
          case A.B.C: {
            break;
          }
        }
      `,
			nil,
			[]string{"switchIsNotExhaustive"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(
				t, SwitchExhaustivenessCheck, switchExhaustivenessFile, testCase.sourceText, testCase.options,
			), testCase.wantIds...)
		})
	}
}

// TestSwitchExhaustivenessCheckResolvesAnEnumAcrossAModuleBoundary is upstream's one case that a
// single-file harness cannot pose at all.
//
// The discriminant's type is `A.B`, a namespaced enum declared in ANOTHER file, and the single case
// covers `A.B.C` while `A.B.D` is missing. Answering that needs the import resolved, the namespace
// entered, and the enum's members enumerated through the checker, so a finding here proves cross-file
// type resolution rather than a lucky local guess.
//
// Dropped instead, it would have been the quietest kind of loss: the file still parses, the switch
// still looks like every other switch in the table, and nothing would have been red.
func TestSwitchExhaustivenessCheckResolvesAnEnumAcrossAModuleBoundary(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTypedFiles(t, SwitchExhaustivenessCheck, map[string]string{
		// Byte for byte tsgolint's own fixture module at
		// internal/rules/fixtures/switch-exhaustiveness-check.ts, which its test file imports by
		// that path.
		"switch-exhaustiveness-check.ts": `
export namespace A {
  export enum B {
    C,
    D,
  }
}
`,
		switchExhaustivenessFile: `
        import { A } from './switch-exhaustiveness-check';
        declare const foo: A.B;
        switch (foo) {
          case A.B.C: {
            break;
          }
        }
      `,
	}, switchExhaustivenessFile)

	rule_testing.ExpectFindings(t, result, "switchIsNotExhaustive")
}

// TestSwitchExhaustivenessCheckReadsTheDefaultCaseComment replays the comment that stands in for a
// default clause through typescript-eslint 8.71.0 (#6esg2nx).
//
// tsgolint declared `defaultCaseCommentPattern` and never read it, and this test used to pin that
// no-op, as the tripwire for the day the option came alive. It came alive here, ported from 8.71's
// getCommentDefaultCase, so the tripwire is now the replay: the first six rows are every row in
// upstream's v8.71.0 test file that carries a comment or the pattern, the six tsgolint skipped, and
// the rest are edge rows for each shape the comment lookup has to get right. Each was run through the
// installed rule on a typed scratch project and the findings written here, id and span.
func TestSwitchExhaustivenessCheckReadsTheDefaultCaseComment(t *testing.T) {
	t.Parallel()

	cases := []struct {
		options string
		code    string
		want    []string
	}{
		{"{\"requireDefaultForNonUnion\":true}", "\ndeclare const value: number;\nswitch (value) {\n  case 0:\n    break;\n  case 1:\n    break;\n  // no default\n}\n      ", nil},
		{"{\"considerDefaultExhaustiveForUnions\":true}", "\ndeclare const value: 'a' | 'b';\nswitch (value) {\n  case 'a':\n    break;\n  // no default\n}\n      ", nil},
		{"{\"considerDefaultExhaustiveForUnions\":true,\"defaultCaseCommentPattern\":\"^skip\\\\sdefault\"}", "\ndeclare const value: 'a' | 'b';\nswitch (value) {\n  case 'a':\n    break;\n  // skip default\n}\n      ", nil},
		{"{\"allowDefaultCaseForExhaustiveSwitch\":false}", "\ndeclare const myValue: 'a' | 'b';\nswitch (myValue) {\n  case 'a':\n    return 'a';\n  case 'b':\n    return 'b';\n  // no default\n}\n      ", []string{"dangerousDefaultCase // no default"}},
		{"{\"considerDefaultExhaustiveForUnions\":false}", "\ndeclare const literal: 'a' | 'b' | 'c';\n\nswitch (literal) {\n  case 'a':\n    break;\n  // no default\n}\n      ", []string{"switchIsNotExhaustive literal"}},
		{"{\"considerDefaultExhaustiveForUnions\":false,\"defaultCaseCommentPattern\":\"^skip\\\\sdefault\"}", "\ndeclare const literal: 'a' | 'b' | 'c';\n\nswitch (literal) {\n  case 'a':\n    break;\n  // skip default\n}\n      ", []string{"switchIsNotExhaustive literal"}},
		{"{\"requireDefaultForNonUnion\":true}", "declare const v: number;\nswitch (v) {\n  case 0:\n    break;\n  // no default\n}", nil},
		{"{\"requireDefaultForNonUnion\":true}", "declare const v: number;\nswitch (v) {\n  case 0:\n    break;\n  // No Default\n}", nil},
		{"{\"requireDefaultForNonUnion\":true}", "declare const v: number;\nswitch (v) {\n  case 0:\n    break;\n  /* no default */\n}", nil},
		{"{\"requireDefaultForNonUnion\":true}", "declare const v: number;\nswitch (v) {\n  case 0:\n    break; // no default\n}", nil},
		{"{\"requireDefaultForNonUnion\":true}", "declare const v: number;\nswitch (v) {\n  case 0: {\n    break;\n    // no default\n  }\n}", []string{"switchIsNotExhaustive v"}},
		{"{\"requireDefaultForNonUnion\":true}", "declare const v: number;\nswitch (v) {\n  case 0:\n    break;\n  // no default\n  // something else\n}", []string{"switchIsNotExhaustive v"}},
		{"{\"requireDefaultForNonUnion\":true}", "declare const v: number;\nswitch (v) {\n  case 0:\n    break;\n  // something else\n  // no default\n}", nil},
		{"{\"requireDefaultForNonUnion\":true}", "declare const v: number;\nswitch (v) {\n  // no default\n}", []string{"switchIsNotExhaustive v"}},
		{"{\"requireDefaultForNonUnion\":true}", "declare const v: number;\nswitch (v) {\n  // no default\n  case 0:\n    break;\n}", []string{"switchIsNotExhaustive v"}},
		{"{\"requireDefaultForNonUnion\":true}", "declare const v: number;\nswitch (v) {\n  case 0:\n    break;\n  // no default here\n}", []string{"switchIsNotExhaustive v"}},
		{"{\"requireDefaultForNonUnion\":true,\"defaultCaseCommentPattern\":\"^skip\\\\sdefault\"}", "declare const v: number;\nswitch (v) {\n  case 0:\n    break;\n  // skip default\n}", nil},
		{"{\"requireDefaultForNonUnion\":true,\"defaultCaseCommentPattern\":\"^skip\\\\sdefault\"}", "declare const v: number;\nswitch (v) {\n  case 0:\n    break;\n  // no default\n}", []string{"switchIsNotExhaustive v"}},
		{"{\"requireDefaultForNonUnion\":true,\"defaultCaseCommentPattern\":\"^skip\\\\sdefault\"}", "declare const v: number;\nswitch (v) {\n  case 0:\n    break;\n  // SKIP default\n}", []string{"switchIsNotExhaustive v"}},
		{"{\"considerDefaultExhaustiveForUnions\":true}", "declare const literal: 'a' | 'b';\nswitch (literal) {\n  case 'a':\n    break;\n  // no default\n}", nil},
		{"{}", "declare const literal: 'a' | 'b';\nswitch (literal) {\n  case 'a':\n    break;\n  // no default\n}", []string{"switchIsNotExhaustive literal"}},
		{"{\"allowDefaultCaseForExhaustiveSwitch\":false}", "declare const literal: 'a' | 'b';\nswitch (literal) {\n  case 'a':\n    break;\n  case 'b':\n    break;\n  // no default\n}", []string{"dangerousDefaultCase // no default"}},
		{"{\"allowDefaultCaseForExhaustiveSwitch\":false}", "declare const literal: 'a' | 'b';\nswitch (literal) {\n  case 'a':\n    break;\n  case 'b':\n    break;\n  default:\n    break;\n  // no default\n}", []string{"dangerousDefaultCase default:\n    break;"}},
		{"{\"allowDefaultCaseForExhaustiveSwitch\":false}", "declare const literal: 'a' | 'b';\nswitch (literal) {\n  case 'a':\n    break;\n  case 'b':\n    break;\n  /* no default */\n}", []string{"dangerousDefaultCase /* no default */"}},
		{"{\"requireDefaultForNonUnion\":true}", "declare const v: number;\nswitch (v) {\n  case 0:\n    break;\n  //no default\n}", nil},
		{"{\"requireDefaultForNonUnion\":true}", "declare const v: number;\nswitch (v) {\n  case 0:\n    break;\n  /**\n   * no default\n   */\n}", []string{"switchIsNotExhaustive v"}},
		{"{\"requireDefaultForNonUnion\":true}", "declare const v: number;\nswitch (v) {\n  case 0:\n    break;\n  /*\n    no default\n  */\n}", nil},
		{"{\"requireDefaultForNonUnion\":true}", "declare const v: number;\nswitch (v) {\n  case 0:\n  // no default\n}", nil},
		{"{\"requireDefaultForNonUnion\":true,\"defaultCaseCommentPattern\":\"\"}", "declare const v: number;\nswitch (v) {\n  case 0:\n    break;\n  // skip\n}", nil},
		{"{\"requireDefaultForNonUnion\":true}", "declare const v: number;\nswitch (v) {\n  case 0:\n    break;\n}\n// no default", []string{"switchIsNotExhaustive v"}},
		{"{\"requireDefaultForNonUnion\":true}", "declare const v: number;\ndeclare const w: number;\nswitch (v) {\n  case 0:\n    switch (w) {\n      case 1:\n        break;\n      // no default\n    }\n}", []string{"switchIsNotExhaustive v"}},
	}
	if len(cases) != 31 {
		t.Fatalf("%d rows, and 31 were replayed", len(cases))
	}

	for _, testCase := range cases {
		t.Run(testCase.options+" "+testCase.code, func(t *testing.T) {
			t.Parallel()

			var options SwitchExhaustivenessCheckOptions
			if err := rule.UnmarshalOptions([]byte(testCase.options), &options); err != nil {
				t.Fatalf("decoding %s: %v", testCase.options, err)
			}
			// The harness writes the file trimmed, so positions are in the trimmed text.
			source := strings.TrimSpace(testCase.code) + "\n"
			result := rule_testing.RunTypedWithOptions(t, SwitchExhaustivenessCheck, switchExhaustivenessFile, source, options)
			var got []string
			for _, diagnostic := range result.Diagnostics {
				got = append(got, diagnostic.Message.Id+" "+source[diagnostic.Range.Pos():diagnostic.Range.End()])
			}
			if strings.Join(got, " | ") != strings.Join(testCase.want, " | ") {
				t.Fatalf("reported %q, typescript-eslint reports %q", got, testCase.want)
			}
		})
	}
}

// TestSwitchExhaustivenessCheckShipsNoRepairs pins the absence of fixes and suggestions.
//
// This is the one rule in the tsgolint family whose repair would WRITE NEW CODE — adding a missing
// `case` clause rather than deleting or rewriting an existing one — so the porting brief expected
// repairs here and warned that `rule_testing` applies fixes but not suggestions. It ships neither.
// `buildAddMissingCasesMessage` is declared and never called, every one of upstream's expected
// suggestion outputs is commented out under `TODO(port): add support for suggestions`, and
// `checkSwitchNoUnionDefaultCase` carries `// TODO(port): missing suggestion` at the report site.
//
// Asserted rather than assumed, because "no repairs" is the kind of claim that stays true by
// accident until a sync makes it false. If a later tsgolint implements them, a rule that suddenly
// rewrites source would otherwise land here unannounced, and the whole imported corpus above would
// stay green while it did — every case in it asserts message ids only.
func TestSwitchExhaustivenessCheckShipsNoRepairs(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTypedWithOptions(
		t, SwitchExhaustivenessCheck, switchExhaustivenessFile, `
declare const literal: 'a' | 'b' | 'c';
switch (literal) {
  case 'a':
    break;
}
`, nil)

	rule_testing.ExpectFindings(t, result, "switchIsNotExhaustive")

	for _, diagnostic := range result.Diagnostics {
		if len(diagnostic.Fixes) != 0 {
			t.Fatalf("tsgolint ships no fixes for this rule, but %q proposed %d", diagnostic.Message.Id, len(diagnostic.Fixes))
		}
		if len(diagnostic.Suggestions) != 0 {
			t.Fatalf("tsgolint ships no suggestions for this rule, but %q proposed %d", diagnostic.Message.Id, len(diagnostic.Suggestions))
		}
	}
}

// TestSwitchExhaustivenessCheckReportsTheDiscriminantNotTheStatement asserts the SPAN.
//
// `upstream.Adapt` was passing raw `node.Loc`, which includes leading trivia, so an indented
// finding began at the end of the previous line and carried a newline and the indentation into its
// rendered text. Fixed in 8bdd70b, and invisible to every message-id fixture in this file: the id
// and the count are both correct either way, and only slicing the source with the finding's own
// range shows it. It reaches a user as a caret pointing at the wrong line.
//
// The anchor matters as much as the trim. Two of this rule's three checks report on
// `node.Expression`, the DISCRIMINANT, not on the switch statement and not on the case clause. A
// port that reported the statement would produce a finding spanning the entire switch body, which
// no id assertion could see, so this pins the exact text.
func TestSwitchExhaustivenessCheckReportsTheDiscriminantNotTheStatement(t *testing.T) {
	t.Parallel()

	const sourceText = `
declare const someDiscriminant: 'a' | 'b';
function run() {
  switch (someDiscriminant) {
    case 'a':
      break;
  }
}
`

	result := rule_testing.RunTypedWithOptions(t, SwitchExhaustivenessCheck, switchExhaustivenessFile, sourceText, nil)
	rule_testing.ExpectFindings(t, result, "switchIsNotExhaustive")

	// The harness trims the source before writing it, so offsets are against the trimmed text.
	trimmed := strings.TrimSpace(sourceText)
	finding := result.Diagnostics[0]
	reported := trimmed[finding.Range.Pos():finding.Range.End()]

	if reported != "someDiscriminant" {
		t.Fatalf("expected the finding to span the discriminant, got %q", reported)
	}
}

// TestSwitchExhaustivenessCheckReportsTheDefaultClauseForADangerousDefault pins the OTHER anchor.
//
// `checkSwitchUnnecessaryDefaultCase` reports on the default CLAUSE rather than on the
// discriminant, so this rule has two distinct report anchors and an id assertion cannot tell them
// apart. Pinning only the discriminant would leave the second one free to point anywhere.
//
// The span is the clause including its body, not just the `default` keyword, because upstream
// reports the whole `CaseOrDefaultClause` node.
func TestSwitchExhaustivenessCheckReportsTheDefaultClauseForADangerousDefault(t *testing.T) {
	t.Parallel()

	const sourceText = `
declare const literal: 'a' | 'b';
switch (literal) {
  case 'a':
    break;
  case 'b':
    break;
  default:
    break;
}
`

	result := rule_testing.RunTypedWithOptions(
		t, SwitchExhaustivenessCheck, switchExhaustivenessFile, sourceText,
		SwitchExhaustivenessCheckOptions{AllowDefaultCaseForExhaustiveSwitch: type_checking.Ref(false)},
	)
	rule_testing.ExpectFindings(t, result, "dangerousDefaultCase")

	trimmed := strings.TrimSpace(sourceText)
	finding := result.Diagnostics[0]
	reported := trimmed[finding.Range.Pos():finding.Range.End()]

	if !strings.HasPrefix(reported, "default:") {
		t.Fatalf("expected the finding to start at the default clause, got %q", reported)
	}
	if strings.Contains(reported, "case 'a'") {
		t.Fatalf("the finding span reached back over the earlier cases: %q", reported)
	}
}

// TestSwitchExhaustivenessCheckDeclaresItNeedsTheChecker pins the declarations AND the nil guard.
//
// The standing advice is that a type-aware listener opens with `if ctx.TypeChecker == nil { return }`.
// While this rule was adapted that could not be followed: the listener was upstream's vendored code,
// and editing it was exactly what would have turned a future re-sync from a diff into a merge, so
// the declaration `upstream.Adapt` set unconditionally was the only guard available. Absorbing the
// rule made the listener ours to edit and made the nil case REACHABLE, so both halves are asserted.
//
// The vacuous direction is the dangerous one and it is worse for this rule than for its siblings.
// Handed no checker the rule does not panic, it reports nothing — and forty-nine of the
// ninety-seven imported cases in this file are CLEAN cases, which would all still pass. Half the
// suite would keep proving something and half would be proving nothing, which reads as a healthy
// suite.
func TestSwitchExhaustivenessCheckDeclaresItNeedsTheChecker(t *testing.T) {
	t.Parallel()

	if !SwitchExhaustivenessCheck.NeedsTypeChecker {
		t.Fatal("the rule must declare NeedsTypeChecker: without a checker it reports nothing rather than failing, so every clean fixture would pass vacuously")
	}

	// The guard the absorption made possible. Driving the listener with a checker-less Context must
	// return before `getSwitchMetadata` reaches the checker, and this is the only path that reaches
	// that branch, since registration always supplies one.
	typed := rule_testing.RunTyped(t, SwitchExhaustivenessCheck, switchExhaustivenessFile,
		"declare const day: 'a' | 'b';\nswitch (day) {\n  case 'a':\n    break;\n}\n")
	if len(typed.Diagnostics) != 1 {
		t.Fatalf("the typed harness found %d findings, want one", len(typed.Diagnostics))
	}

	listeners := SwitchExhaustivenessCheck.Run(rule.Context{SourceFile: typed.SourceFile}, nil)
	listener, hasListener := listeners[ast.KindSwitchStatement]
	if !hasListener {
		t.Fatal("the rule stopped listening on switch statements")
	}
	for _, statement := range typed.SourceFile.Statements.Nodes {
		if statement.Kind == ast.KindSwitchStatement {
			listener(statement)
		}
	}
}

// TestSwitchExhaustivenessCheckOptionsBindFromCamelCaseJson pins the decoder.
//
// The registration uses `rule.DecodeOptionsInto` on upstream's own struct, which works only because
// `encoding/json` matches field names case-insensitively: nothing in this tree declares that
// `allowDefaultCaseForExhaustiveSwitch` should reach `AllowDefaultCaseForExhaustiveSwitch`, and
// upstream's struct carries no JSON tags. That is a property of the standard library, so it is
// measured here rather than trusted.
//
// The distinction that makes this worth a test rather than a comment is the DEFAULT. Two of the
// three live options default to a value that is not the zero value, so the fields must stay
// POINTERS: a plain `bool` could not separate "the user wrote false" from "the user wrote nothing",
// and `allowDefaultCaseForExhaustiveSwitch` defaults to TRUE. Flattening them would silently invert
// that option for every user who did not set it.
func TestSwitchExhaustivenessCheckOptionsBindFromCamelCaseJson(t *testing.T) {
	t.Parallel()

	decode := rule.DecodeOptionsInto[SwitchExhaustivenessCheckOptions]()

	decoded, err := decode([]byte(`{"allowDefaultCaseForExhaustiveSwitch":false,"requireDefaultForNonUnion":true}`))
	if err != nil {
		t.Fatalf("decoding options: %v", err)
	}

	options, isOptions := decoded.(SwitchExhaustivenessCheckOptions)
	if !isOptions {
		t.Fatalf("expected %T, got %T", SwitchExhaustivenessCheckOptions{}, decoded)
	}

	if options.AllowDefaultCaseForExhaustiveSwitch == nil || *options.AllowDefaultCaseForExhaustiveSwitch {
		t.Fatal("allowDefaultCaseForExhaustiveSwitch did not bind to false")
	}
	if options.RequireDefaultForNonUnion == nil || !*options.RequireDefaultForNonUnion {
		t.Fatal("requireDefaultForNonUnion did not bind to true")
	}
	// The unset option must stay nil rather than becoming false, because the rule's own defaulting
	// block reads exactly that distinction and `considerDefaultExhaustiveForUnions` happens to
	// default to false only by coincidence of it being the safer direction.
	if options.ConsiderDefaultExhaustiveForUnions != nil {
		t.Fatal("an option nobody configured must stay nil so the rule can apply its own default")
	}

	// And the option that binds and then does nothing, which is the trap this rule carries. It must
	// still decode cleanly, because refusing it would reject configs upstream accepts.
	withPattern, err := decode([]byte(`{"defaultCaseCommentPattern":"^skip"}`))
	if err != nil {
		t.Fatalf("decoding the inert option: %v", err)
	}
	if patternOptions, _ := withPattern.(SwitchExhaustivenessCheckOptions); patternOptions.DefaultCaseCommentPattern == nil {
		t.Fatal("defaultCaseCommentPattern did not bind")
	}
}

// TestSwitchExhaustivenessCheckExclusionsAreStated names every upstream case this file does not
// carry, so that the trim is a decision on the record rather than a silent shortfall.
//
// Nine of upstream's hundred and six cases are excluded, in two groups, and neither group is a
// judgment of mine about what matters.
//
// SIX are tsgolint's own `Skip: true` cases, every one of them a `DefaultCaseCommentPattern` case
// marked `TODO(port): add support for DefaultCaseCommentPattern`. The feature exists here now
// (#6esg2nx), and those six are carried, from typescript-eslint 8.71.0's test file and replayed
// through its rule, in TestSwitchExhaustivenessCheckReadsTheDefaultCaseComment rather than here.
//
// THREE need `noUncheckedIndexedAccess`, which changes `x[0]` on a `string[]` from `string` to
// `string | undefined` and is therefore the entire point of those cases. `rule_testing`'s tsconfig is a
// hardcoded constant with no per-test override, so under our config the discriminant is plain
// `string`, no undefined member is ever missing, and all three would assert the OPPOSITE of
// upstream while looking like ordinary passes. That is exactly the shape the porting brief warns
// about, arriving through a compiler flag rather than through `lib`. Dropping them is the honest
// answer; keeping them under a config that changes their meaning is the failure this whole exercise
// exists to prevent.
//
// The three are recorded here as source rather than as prose so a later reader can restore them
// the moment `rule_testing` grows a per-test tsconfig, and so the claim that they behave differently
// under our config is checkable rather than asserted. Each one is run below: upstream expects the
// first two to be SILENT and the third to REPORT, and under our tsconfig the first two are silent
// for the wrong reason and the third is silent outright.
func TestSwitchExhaustivenessCheckExclusionsAreStated(t *testing.T) {
	t.Parallel()

	// Upstream's invalid case: with noUncheckedIndexedAccess, `x[0]` is `string | undefined`, the
	// switch covers only `'hi'`, and `undefined` is a missing branch. Without the flag the
	// discriminant is plain `string`, which contains a non-literal type and has no missing literal
	// branches at all, so nothing reports.
	reportsOnlyUnderNoUncheckedIndexedAccess := rule_testing.RunTypedWithOptions(
		t, SwitchExhaustivenessCheck, switchExhaustivenessFile, `
function foo(x: string[]) {
  switch (x[0]) {
    case 'hi':
      break;
  }
}
`, nil)

	rule_testing.ExpectClean(t, reportsOnlyUnderNoUncheckedIndexedAccess)

	// Upstream's first valid case for the same flag passes here too, and that is the QUIET half of
	// the danger: it is silent for a different reason than upstream's. With the flag, `x[0]` is
	// `string | undefined` and `case undefined` is what covers the undefined member. Without it,
	// `x[0]` is plain `string`, there is no undefined member to cover, and the `case undefined` arm
	// is doing nothing. Carried into the table above it would have looked like one more imported
	// case while proving nothing about the behavior it was written for.
	rule_testing.ExpectClean(t, rule_testing.RunTypedWithOptions(
		t, SwitchExhaustivenessCheck, switchExhaustivenessFile, `
function foo(x: string[]) {
  switch (x[0]) {
    case 'hi':
      break;
    case undefined:
      break;
  }
}
`, nil))

	// Upstream's second valid case is the LOUD half, and it is the one that justifies this whole
	// test existing rather than a paragraph of prose.
	//
	// Upstream expects SILENCE. Under our tsconfig it REPORTS. With the flag, `const a = x[0]` is
	// `string | undefined`, the `typeof a === 'string'` early return narrows `a` to `undefined` at
	// the switch, and `case a` therefore covers the `undefined` member of `y`. Without the flag `a`
	// is plain `string`, the early return narrows it to `never`, `case a` covers nothing, and the
	// `undefined` member of `string | undefined` is reported missing.
	//
	// So this case does not merely prove something different under our config, it asserts the exact
	// OPPOSITE of upstream. Dropped into the valid table unchanged it would have been a red test
	// telling a true story badly; dropped into the invalid table it would have been a GREEN test
	// encoding a behavior upstream does not have. Measured rather than reasoned about, and recorded
	// here in the direction our harness actually produces.
	rule_testing.ExpectFindings(t, rule_testing.RunTypedWithOptions(
		t, SwitchExhaustivenessCheck, switchExhaustivenessFile, `
function foo(x: string[], y: string | undefined) {
  const a = x[0];
  if (typeof a === 'string') {
    return;
  }
  switch (y) {
    case 'hi':
      break;
    case a:
      break;
  }
}
`, nil), "switchIsNotExhaustive")
}

// TestSwitchExhaustivenessCheckNamesTheMissingBranches pins the message text against
// typescript-eslint's, measured on the installed 8.67.0 build for every kind of member a switch can
// miss (#21011kd). tsgolint rendered the bare sentence for all seven; the list is what a reader acts on.
func TestSwitchExhaustivenessCheckNamesTheMissingBranches(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
		want   string
	}{
		{"string literals", "declare const day: 'a' | 'b' | 'c';\nswitch (day) { case 'a': break; }\n", `"b" | "c"`},
		{"number literals", "declare const count: 1 | 2 | 3;\nswitch (count) { case 1: break; }\n", `2 | 3`},
		{"a boolean", "declare const flag: boolean;\nswitch (flag) { case true: break; }\n", `false`},
		{"enum members", "enum Color { Red, Green, Blue }\ndeclare const color: Color;\nswitch (color) { case Color.Red: break; }\n", `Color.Green | Color.Blue`},
		{"null and undefined", "declare const maybe: 'x' | null | undefined;\nswitch (maybe) { case 'x': break; }\n", `undefined | null`},
		{"a unique symbol", "const first: unique symbol = Symbol('first');\nconst second: unique symbol = Symbol('second');\ndeclare const token: typeof first | typeof second;\nswitch (token) { case first: break; }\n", `typeof second`},
		{"an optional property", "declare const optional: { kind?: 'p' | 'q' };\nswitch (optional.kind) { case 'p': break; }\n", `undefined | "q"`},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, SwitchExhaustivenessCheck, switchExhaustivenessFile, testCase.source)
			rule_testing.ExpectFindings(t, result, "switchIsNotExhaustive")
			want := "Switch is not exhaustive. Cases not matched: " + testCase.want
			if got := result.Diagnostics[0].Message.Description; got != want {
				t.Fatalf("message: expected %q, got %q", want, got)
			}
		})
	}
}
