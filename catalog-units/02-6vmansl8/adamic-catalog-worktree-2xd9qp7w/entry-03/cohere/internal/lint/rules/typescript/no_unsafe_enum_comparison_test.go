package typescript

import (
	"testing"

	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
)

// noUnsafeEnumComparisonFile is where the fixtures pretend to live.
//
// A `.ts` name: nothing in this corpus is JSX, and the rule has no file gate.
const noUnsafeEnumComparisonFile = "/repository/source/NoUnsafeEnumComparison.ts"

// The corpus is typescript-eslint's own, extracted from its RuleTester and verified byte for
// byte through a second independent extraction: 85 case bodies, zero mismatches.
//
// Every row runs through `RunTyped` rather than `Run`. This rule declares NeedsTypeChecker, and
// the plain harness hands it a nil checker -- which this rule GUARDS, so it would go completely
// silent and every StaysSilent case would pass vacuously while every Fires case failed in a way
// that looks like a rule bug.
type noUnsafeEnumComparisonCase struct {
	name   string
	source string
	ids    []string
}

var noUnsafeEnumComparisonFiresCases = []noUnsafeEnumComparisonCase{
	{
		name:   "invalid-0",
		source: "\nenum Fruit {\n  Apple,\n}\nFruit.Apple < 1;\n      ",
		ids:    []string{"mismatchedCondition"},
	},
	{
		name:   "invalid-1",
		source: "\nenum Fruit {\n  Apple,\n}\nFruit.Apple > 1;\n      ",
		ids:    []string{"mismatchedCondition"},
	},
	{
		name:   "invalid-2",
		source: "\nenum Fruit {\n  Apple,\n}\nFruit.Apple == 1;\n      ",
		ids:    []string{"mismatchedCondition"},
	},
	{
		name:   "invalid-3",
		source: "\nenum Fruit {\n  Apple,\n}\nFruit.Apple === 1;\n      ",
		ids:    []string{"mismatchedCondition"},
	},
	{
		name:   "invalid-4",
		source: "\nenum Fruit {\n  Apple,\n}\nFruit.Apple != 1;\n      ",
		ids:    []string{"mismatchedCondition"},
	},
	{
		name:   "invalid-5",
		source: "\nenum Fruit {\n  Apple,\n}\nFruit.Apple !== 1;\n      ",
		ids:    []string{"mismatchedCondition"},
	},
	{
		name:   "invalid-6",
		source: "\nenum Fruit {\n  Apple = 0,\n  Banana = 'banana',\n}\nFruit.Apple === 0;\n      ",
		ids:    []string{"mismatchedCondition"},
	},
	{
		name:   "invalid-7",
		source: "\nenum Fruit {\n  Apple = 0,\n  Banana = 'banana',\n}\nFruit.Banana === '';\n      ",
		ids:    []string{"mismatchedCondition"},
	},
	{
		name:   "invalid-8",
		source: "\nenum Vegetable {\n  Asparagus = 'asparagus',\n  Beet = 'beet',\n  Celery = 'celery',\n}\nVegetable.Asparagus === 'beet';\n      ",
		ids:    []string{"mismatchedCondition"},
	},
	{
		name:   "invalid-9",
		source: "\nenum Fruit {\n  Apple,\n  Banana,\n  Cherry,\n}\n1 === Fruit.Apple;\n      ",
		ids:    []string{"mismatchedCondition"},
	},
	{
		name:   "invalid-10",
		source: "\nenum Vegetable {\n  Asparagus = 'asparagus',\n  Beet = 'beet',\n  Celery = 'celery',\n}\n'beet' === Vegetable.Asparagus;\n      ",
		ids:    []string{"mismatchedCondition"},
	},
	{
		name:   "invalid-11",
		source: "\nenum Fruit {\n  Apple,\n  Banana,\n  Cherry,\n}\nconst fruit = Fruit.Apple;\nfruit === 1;\n      ",
		ids:    []string{"mismatchedCondition"},
	},
	{
		name:   "invalid-12",
		source: "\nenum Vegetable {\n  Asparagus = 'asparagus',\n  Beet = 'beet',\n  Celery = 'celery',\n}\nconst vegetable = Vegetable.Asparagus;\nvegetable === 'beet';\n      ",
		ids:    []string{"mismatchedCondition"},
	},
	{
		name:   "invalid-13",
		source: "\nenum Fruit {\n  Apple,\n  Banana,\n  Cherry,\n}\nconst fruit = Fruit.Apple;\n1 === fruit;\n      ",
		ids:    []string{"mismatchedCondition"},
	},
	{
		name:   "invalid-14",
		source: "\nenum Vegetable {\n  Asparagus = 'asparagus',\n  Beet = 'beet',\n  Celery = 'celery',\n}\nconst vegetable = Vegetable.Asparagus;\n'beet' === vegetable;\n      ",
		ids:    []string{"mismatchedCondition"},
	},
	{
		name:   "invalid-15",
		source: "\nenum Fruit {\n  Apple,\n  Banana,\n  Cherry,\n}\nenum Fruit2 {\n  Apple2,\n  Banana2,\n  Cherry2,\n}\nFruit.Apple === Fruit2.Apple2;\n      ",
		ids:    []string{"mismatchedCondition"},
	},
	{
		name:   "invalid-16",
		source: "\nenum Vegetable {\n  Asparagus = 'asparagus',\n  Beet = 'beet',\n  Celery = 'celery',\n}\nenum Vegetable2 {\n  Asparagus2 = 'asparagus2',\n  Beet2 = 'beet2',\n  Celery2 = 'celery2',\n}\nVegetable.Asparagus === Vegetable2.Asparagus2;\n      ",
		ids:    []string{"mismatchedCondition"},
	},
	{
		name:   "invalid-17",
		source: "\nenum Fruit {\n  Apple,\n  Banana,\n  Cherry,\n}\nenum Fruit2 {\n  Apple2,\n  Banana2,\n  Cherry2,\n}\nconst fruit = Fruit.Apple;\nfruit === Fruit2.Apple2;\n      ",
		ids:    []string{"mismatchedCondition"},
	},
	{
		name:   "invalid-18",
		source: "\nenum Vegetable {\n  Asparagus = 'asparagus',\n  Beet = 'beet',\n  Celery = 'celery',\n}\nenum Vegetable2 {\n  Asparagus2 = 'asparagus2',\n  Beet2 = 'beet2',\n  Celery2 = 'celery2',\n}\nconst vegetable = Vegetable.Asparagus;\nvegetable === Vegetable2.Asparagus2;\n      ",
		ids:    []string{"mismatchedCondition"},
	},
	{
		name:   "invalid-19",
		source: "\nenum Str {\n  A = 'a',\n}\nenum Num {\n  B = 1,\n}\nenum Mixed {\n  A = 'a',\n  B = 1,\n}\n\ndeclare const str: Str;\ndeclare const num: Num;\ndeclare const mixed: Mixed;\n\n// following are all errors because the value might be an enum value\nstr === 'a';\nnum === 1;\nmixed === 'a';\nmixed === 1;\n      ",
		ids:    []string{"mismatchedCondition", "mismatchedCondition", "mismatchedCondition", "mismatchedCondition"},
	},
	{
		name:   "invalid-20",
		source: "\nenum Fruit {\n  Apple = 'apple',\n}\ntype __String =\n  | (string & { __escapedIdentifier: void })\n  | (void & { __escapedIdentifier: void })\n  | Fruit;\ndeclare const weirdString: __String;\nweirdString === 'someArbitraryValue';\n      ",
		ids:    []string{"mismatchedCondition"},
	},
	{
		name:   "invalid-21",
		source: "\nenum Fruit {\n  Apple,\n}\n\ndeclare const fruit: Fruit;\n\nswitch (fruit) {\n  case 0: {\n    break;\n  }\n}\n      ",
		ids:    []string{"mismatchedCase"},
	},
	{
		name:   "invalid-22",
		source: "\nenum Fruit {\n  Apple,\n  Banana,\n}\n\ndeclare const fruit: Fruit;\n\nswitch (fruit) {\n  case Fruit.Apple: {\n    break;\n  }\n  case 1: {\n    break;\n  }\n}\n      ",
		ids:    []string{"mismatchedCase"},
	},
	{
		name:   "invalid-23",
		source: "\nenum Vegetable {\n  Asparagus = 'asparagus',\n}\n\ndeclare const vegetable: Vegetable;\n\nswitch (vegetable) {\n  case 'asparagus': {\n    break;\n  }\n}\n      ",
		ids:    []string{"mismatchedCase"},
	},
	{
		name:   "invalid-24",
		source: "\nenum Vegetable {\n  Asparagus = 'asparagus',\n  Beet = 'beet',\n}\n\ndeclare const vegetable: Vegetable;\n\nswitch (vegetable) {\n  case Vegetable.Asparagus: {\n    break;\n  }\n  case 'beet': {\n    break;\n  }\n}\n      ",
		ids:    []string{"mismatchedCase"},
	},
	{
		name:   "invalid-25",
		source: "\nenum Vegetable {\n  Asparagus = 'asparagus',\n  Beet = 'beet',\n}\n\ndeclare const vegetable: Vegetable;\n\nswitch (vegetable) {\n  case Vegetable.Asparagus: {\n    break;\n  }\n  case 'beet': {\n    break;\n  }\n  default: {\n    break;\n  }\n}\n      ",
		ids:    []string{"mismatchedCase"},
	},
	{
		name:   "invalid-26",
		source: "\nenum Str {\n  A = 'a',\n  B = 'b',\n}\ndeclare const str: Str;\nstr === 'b';\n      ",
		ids:    []string{"mismatchedCondition"},
	},
	{
		name:   "invalid-27",
		source: "\nenum Str {\n  A = 'a',\n  AB = 'ab',\n}\ndeclare const str: Str;\nstr === 'a' + 'b';\n      ",
		ids:    []string{"mismatchedCondition"},
	},
	{
		name:   "invalid-28",
		source: "\nenum Num {\n  A = 1,\n  B = 2,\n}\ndeclare const num: Num;\n1 === num;\n      ",
		ids:    []string{"mismatchedCondition"},
	},
	{
		name:   "invalid-29",
		source: "\nenum Num {\n  A = 1,\n  B = 2,\n}\ndeclare const num: Num;\n1 /* with */ === /* comment */ num;\n      ",
		ids:    []string{"mismatchedCondition"},
	},
	{
		name:   "invalid-30",
		source: "\nenum Num {\n  A = 1,\n  B = 2,\n}\ndeclare const num: Num;\n1 + 1 === num;\n      ",
		ids:    []string{"mismatchedCondition"},
	},
	{
		name:   "invalid-31",
		source: "\nenum Mixed {\n  A = 1,\n  B = 'b',\n}\ndeclare const mixed: Mixed;\nmixed === 1;\n      ",
		ids:    []string{"mismatchedCondition"},
	},
	{
		name:   "invalid-32",
		source: "\nenum Mixed {\n  A = 1,\n  B = 'b',\n}\ndeclare const mixed: Mixed;\nmixed === 'b';\n      ",
		ids:    []string{"mismatchedCondition"},
	},
	{
		name:   "invalid-33",
		source: "\nenum StringKey {\n  'test-key' /* with comment */ = 1,\n}\ndeclare const stringKey: StringKey;\nstringKey === 1;\n      ",
		ids:    []string{"mismatchedCondition"},
	},
	{
		name:   "invalid-34",
		source: "\nenum StringKey {\n  \"key-'with-single'-quotes\" = 1,\n}\ndeclare const stringKey: StringKey;\nstringKey === 1;\n      ",
		ids:    []string{"mismatchedCondition"},
	},
	{
		name:   "invalid-35",
		source: "\nenum StringKey {\n  'key-\"with-double\"-quotes' = 1,\n}\ndeclare const stringKey: StringKey;\nstringKey === 1;\n      ",
		ids:    []string{"mismatchedCondition"},
	},
	{
		name:   "invalid-36",
		source: "\nenum StringKey {\n  'key-`with-backticks`-quotes' = 1,\n}\ndeclare const stringKey: StringKey;\nstringKey === 1;\n      ",
		ids:    []string{"mismatchedCondition"},
	},
	{
		name:   "invalid-37",
		source: "\nenum Fruit {\n  Apple,\n}\ndeclare const foo: number & {};\nif (foo === Fruit.Apple) {\n}\n      ",
		ids:    []string{"mismatchedCondition"},
	},
	{
		name:   "invalid-38",
		source: "\nenum Fruit {\n  Apple,\n}\ndeclare const foo: number & { __someBrand: void };\nif (foo === Fruit.Apple) {\n}\n      ",
		ids:    []string{"mismatchedCondition"},
	},
	{
		name:   "invalid-39",
		source: "\nenum Vegetable {\n  Asparagus = 'asparagus',\n}\ndeclare const foo: string & {};\nif (foo === Vegetable.Asparagus) {\n}\n      ",
		ids:    []string{"mismatchedCondition"},
	},
	{
		name:   "invalid-40",
		source: "\nenum Vegetable {\n  Asparagus = 'asparagus',\n}\ndeclare const foo: string & { __someBrand: void };\nif (foo === Vegetable.Asparagus) {\n}\n      ",
		ids:    []string{"mismatchedCondition"},
	},
	{
		name:   "invalid-41",
		source: "\nenum NUMBER_ENUM {\n  First = 0,\n  Second = 1,\n}\n\ntype NumberUnion = 0 | 1;\n\ndeclare const numberUnion: NumberUnion;\n\nswitch (numberUnion) {\n  case NUMBER_ENUM.First:\n  case NUMBER_ENUM.Second:\n    break;\n}\n      ",
		ids:    []string{"mismatchedCase", "mismatchedCase"},
	},
	{
		name:   "invalid-42",
		source: "\nenum STRING_ENUM {\n  First = 'one',\n  Second = 'two',\n}\n\ntype StringUnion = 'one' | 'two';\n\ndeclare const stringUnion: StringUnion;\n\nswitch (stringUnion) {\n  case STRING_ENUM.First:\n  case STRING_ENUM.Second:\n    break;\n}\n      ",
		ids:    []string{"mismatchedCase", "mismatchedCase"},
	},
	{
		name:   "invalid-43",
		source: "\ndeclare const stringUnion: 'foo' | 'bar';\n\nenum StringEnum {\n  FOO = 'foo',\n  BAR = 'bar',\n}\n\ndeclare const stringEnum: StringEnum;\n\nstringUnion === stringEnum;\n      ",
		ids:    []string{"mismatchedCondition"},
	},
}

var noUnsafeEnumComparisonSilentCases = []noUnsafeEnumComparisonCase{
	{
		name:   "valid-0",
		source: "'a' > 'b';",
	},
	{
		name:   "valid-1",
		source: "'a' < 'b';",
	},
	{
		name:   "valid-2",
		source: "'a' == 'b';",
	},
	{
		name:   "valid-3",
		source: "'a' === 'b';",
	},
	{
		name:   "valid-4",
		source: "1 > 2;",
	},
	{
		name:   "valid-5",
		source: "1 < 2;",
	},
	{
		name:   "valid-6",
		source: "1 == 2;",
	},
	{
		name:   "valid-7",
		source: "1 === 2;",
	},
	{
		name:   "valid-8",
		source: "\nenum Fruit {\n  Apple,\n}\nFruit.Apple === ({} as any);\n    ",
	},
	{
		name:   "valid-9",
		source: "\nenum Fruit {\n  Apple,\n}\nFruit.Apple === undefined;\n    ",
	},
	{
		name:   "valid-10",
		source: "\nenum Fruit {\n  Apple,\n}\nFruit.Apple === null;\n    ",
	},
	{
		name:   "valid-11",
		source: "\nenum Fruit {\n  Apple,\n}\ndeclare const fruit: Fruit | -1;\nfruit === -1;\n    ",
	},
	{
		name:   "valid-12",
		source: "\nenum Fruit {\n  Apple,\n}\ndeclare const fruit: Fruit | number;\nfruit === -1;\n    ",
	},
	{
		name:   "valid-13",
		source: "\nenum Fruit {\n  Apple,\n}\ndeclare const fruit: Fruit | 'apple';\nfruit === 'apple';\n    ",
	},
	{
		name:   "valid-14",
		source: "\nenum Fruit {\n  Apple,\n}\ndeclare const fruit: Fruit | string;\nfruit === 'apple';\n    ",
	},
	{
		name:   "valid-15",
		source: "\nenum Fruit {\n  Apple = 'apple',\n}\ndeclare const fruit: Fruit | 'apple';\nfruit === 'apple';\n    ",
	},
	{
		name:   "valid-16",
		source: "\nenum Fruit {\n  Apple = 'apple',\n}\ndeclare const fruit: Fruit | string;\nfruit === 'apple';\n    ",
	},
	{
		name:   "valid-17",
		source: "\nenum Fruit {\n  Apple = 'apple',\n}\ndeclare const fruit: Fruit | 0;\nfruit === 0;\n    ",
	},
	{
		name:   "valid-18",
		source: "\nenum Fruit {\n  Apple = 'apple',\n}\ndeclare const fruit: Fruit | number;\nfruit === 0;\n    ",
	},
	{
		name:   "valid-19",
		source: "\nenum Fruit {\n  Apple,\n}\ndeclare const fruit: Fruit | 'apple';\nfruit === Math.random() > 0.5 ? 'apple' : Fruit.Apple;\n    ",
	},
	{
		name:   "valid-20",
		source: "\nenum Fruit {\n  Apple = 'apple',\n}\ndeclare const fruit: Fruit | 'apple';\nfruit === Math.random() > 0.5 ? 'apple' : Fruit.Apple;\n    ",
	},
	{
		name:   "valid-21",
		source: "\nenum Fruit {\n  Apple = 'apple',\n}\ndeclare const fruit: Fruit | string;\nfruit === Math.random() > 0.5 ? 'apple' : Fruit.Apple;\n    ",
	},
	{
		name:   "valid-22",
		source: "\nenum Fruit {\n  Apple = 'apple',\n}\ndeclare const fruit: Fruit | 0;\nfruit === Math.random() > 0.5 ? 0 : Fruit.Apple;\n    ",
	},
	{
		name:   "valid-23",
		source: "\nenum Fruit {\n  Apple = 'apple',\n}\ndeclare const fruit: Fruit | number;\nfruit === Math.random() > 0.5 ? 0 : Fruit.Apple;\n    ",
	},
	{
		name:   "valid-24",
		source: "\nenum Fruit {\n  Apple,\n  Banana,\n}\nFruit.Apple === Fruit.Banana;\n    ",
	},
	{
		name:   "valid-25",
		source: "\nenum Fruit {\n  Apple = 0,\n  Banana = 1,\n}\nFruit.Apple === Fruit.Banana;\n    ",
	},
	{
		name:   "valid-26",
		source: "\nenum Fruit {\n  Apple = 'apple',\n  Banana = 'banana',\n}\nFruit.Apple === Fruit.Banana;\n    ",
	},
	{
		name:   "valid-27",
		source: "\nenum Fruit {\n  Apple,\n  Banana,\n}\nconst fruit = Fruit.Apple;\nfruit === Fruit.Banana;\n    ",
	},
	{
		name:   "valid-28",
		source: "\nenum Vegetable {\n  Asparagus = 'asparagus',\n  Beet = 'beet',\n  Celery = 'celery',\n}\nconst vegetable = Vegetable.Asparagus;\nvegetable === Vegetable.Beet;\n    ",
	},
	{
		name:   "valid-29",
		source: "\nenum Fruit {\n  Apple,\n  Banana,\n  Cherry,\n}\nconst fruit1 = Fruit.Apple;\nconst fruit2 = Fruit.Banana;\nfruit1 === fruit2;\n    ",
	},
	{
		name:   "valid-30",
		source: "\nenum Vegetable {\n  Asparagus = 'asparagus',\n  Beet = 'beet',\n  Celery = 'celery',\n}\nconst vegetable1 = Vegetable.Asparagus;\nconst vegetable2 = Vegetable.Beet;\nvegetable1 === vegetable2;\n    ",
	},
	{
		name:   "valid-31",
		source: "\nenum Fruit {\n  Apple,\n  Banana,\n  Cherry,\n}\nenum Fruit2 {\n  Apple2,\n  Banana2,\n  Cherry2,\n}\ndeclare const left: number | Fruit;\ndeclare const right: number | Fruit2;\nleft === right;\n    ",
	},
	{
		name:   "valid-32",
		source: "\nenum Vegetable {\n  Asparagus = 'asparagus',\n  Beet = 'beet',\n  Celery = 'celery',\n}\nenum Vegetable2 {\n  Asparagus2 = 'asparagus2',\n  Beet2 = 'beet2',\n  Celery2 = 'celery2',\n}\ndeclare const left: string | Vegetable;\ndeclare const right: string | Vegetable2;\nleft === right;\n    ",
	},
	{
		name:   "valid-33",
		source: "\nenum Vegetable {\n  Asparagus = 'asparagus',\n  Beet = 'beet',\n  Celery = 'celery',\n}\nconst foo = {};\nconst vegetable = Vegetable.Asparagus;\nvegetable in foo;\n    ",
	},
	{
		name:   "valid-34",
		source: "\nenum Fruit {\n  Apple,\n  Banana,\n  Cherry,\n}\ndeclare const fruitOrBoolean: Fruit | boolean;\nfruitOrBoolean === true;\n    ",
	},
	{
		name:   "valid-35",
		source: "\nenum Str {\n  A = 'a',\n}\nenum Num {\n  B = 1,\n}\nenum Mixed {\n  A = 'a',\n  B = 1,\n}\n\ndeclare const str: Str;\ndeclare const strOrString: Str | string;\n\ndeclare const num: Num;\ndeclare const numOrNumber: Num | number;\n\ndeclare const mixed: Mixed;\ndeclare const mixedOrStringOrNumber: Mixed | string | number;\n\nfunction someFunction() {}\n\n// following are all ignored due to the presence of \"| string\" or \"| number\"\nstrOrString === 'a';\nnumOrNumber === 1;\nmixedOrStringOrNumber === 'a';\nmixedOrStringOrNumber === 1;\n\n// following are all ignored because the value can never be an enum value\nstr === 1;\nnum === 'a';\nstr === {};\nnum === {};\nmixed === {};\nstr === true;\nnum === true;\nmixed === true;\nstr === someFunction;\nnum === someFunction;\nmixed === someFunction;\n    ",
	},
	{
		name:   "valid-36",
		source: "\nenum Fruit {\n  Apple,\n}\n\nconst bitShift = 1 << Fruit.Apple;\n    ",
	},
	{
		name:   "valid-37",
		source: "\nenum Fruit {\n  Apple,\n}\n\nconst bitShift = 1 >> Fruit.Apple;\n    ",
	},
	{
		name:   "valid-38",
		source: "\nenum Fruit {\n  Apple,\n}\n\ndeclare const fruit: Fruit;\n\nswitch (fruit) {\n  case Fruit.Apple: {\n    break;\n  }\n}\n    ",
	},
	{
		name:   "valid-39",
		source: "\nenum Vegetable {\n  Asparagus = 'asparagus',\n}\n\ndeclare const vegetable: Vegetable;\n\nswitch (vegetable) {\n  case Vegetable.Asparagus: {\n    break;\n  }\n}\n    ",
	},
	{
		name:   "valid-40",
		source: "\nenum Vegetable {\n  Asparagus = 'asparagus',\n}\n\ndeclare const vegetable: Vegetable;\n\nswitch (vegetable) {\n  default: {\n    break;\n  }\n}\n    ",
	},
}

func TestNoUnsafeEnumComparisonFires(t *testing.T) {
	t.Parallel()
	for _, testCase := range noUnsafeEnumComparisonFiresCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnsafeEnumComparison, noUnsafeEnumComparisonFile, testCase.source)
			rule_testing.ExpectFindings(t, result, testCase.ids...)
		})
	}
}

func TestNoUnsafeEnumComparisonStaysSilent(t *testing.T) {
	t.Parallel()
	for _, testCase := range noUnsafeEnumComparisonSilentCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnsafeEnumComparison, noUnsafeEnumComparisonFile, testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// The two tests below exist because a mutation sweep found two discriminations upstream's own
// corpus cannot see. Both were settled against upstream's SOURCE for the helper they exercise, and
// each is written so the mutant that survived the corpus fails here.

// TestNoUnsafeEnumComparisonUnionsMustBeWhollyPrimitive covers the every-versus-some asymmetry.
//
// `isNumberLike` and `isStringLike` are EVERY union constituent over SOME intersection constituent.
// Flipping the union quantifier to `some` survives all 85 corpus cases, because no case writes a
// union with a mixed primitive kind: the closest are `Fruit | number` and `Fruit | string`, which
// gate three excuses before the quantifier is consulted.
//
// The distinguishing input is a union that is PARTLY string-like compared against a string enum.
// Under `every` it is not string-like, so the comparison is allowed; under `some` it reports.
// Measured both ways here: the shipped rule is clean and the mutant reports.
//
// The asymmetry is not arbitrary. A union is number-like only if every branch is, because a branch
// that is not gives the value a way to be something else. An intersection is number-like if any
// constituent is, because an intersection is all of its parts at once and `number & {}` is still a
// number.
func TestNoUnsafeEnumComparisonUnionsMustBeWhollyPrimitive(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		source  string
		reports bool
	}{
		{
			// The discriminator. A mixed union is not string-like, so this is safe.
			name:    "mixed-union-is-not-string-like",
			source:  "enum SE { FOO = 'foo' }\ndeclare const se: SE;\ndeclare const mixed: string | number;\nconst a = mixed === se;\n",
			reports: false,
		},
		{
			// The control. Without it, a rule that had stopped judging unions entirely would pass
			// the case above.
			name:    "wholly-string-union-is-string-like",
			source:  "enum SE { FOO = 'foo' }\ndeclare const se: SE;\ndeclare const allStrings: 'x' | 'y';\nconst b = allStrings === se;\n",
			reports: true,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnsafeEnumComparison, noUnsafeEnumComparisonFile,
				testCase.source)
			if testCase.reports {
				rule_testing.ExpectFindings(t, result, "mismatchedCondition")
				return
			}
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoUnsafeEnumComparisonGateOneIsACostGuard records an EQUIVALENT mutant rather than killing it.
//
// Upstream's first gate returns early when neither side carries an enum type. Removing it survives
// the whole corpus, and that is genuine equivalence rather than a fixture gap: with no enum on
// either side, `typeViolates` finds neither a number-valued nor a string-valued enum in either
// direction, so gate four answers false anyway. Both routes reach the same verdict for every
// possible input.
//
// It is kept because it is upstream's and because it is a real cost guard: without it, every `a === b`
// in the tree splits two types into union and intersection constituents to reach a conclusion that
// the first line could have stated. Asserted here so the equivalence is a recorded measurement
// rather than an argument, and so a future change that makes gate four reachable without an enum
// fails loudly.
func TestNoUnsafeEnumComparisonGateOneIsACostGuard(t *testing.T) {
	t.Parallel()
	const source = "declare const a: string;\ndeclare const b: number;\nconst c = a === b;\n"

	result := rule_testing.RunTyped(t, NoUnsafeEnumComparison, noUnsafeEnumComparisonFile, source)
	rule_testing.ExpectClean(t, result)

	// The control: the same shape WITH an enum on one side does report, so the clean verdict above
	// is a decision rather than the rule being inert on this file.
	const withEnum = "enum SE { FOO = 'foo' }\ndeclare const se: SE;\ndeclare const a: string;\nconst c = a === se;\n"
	controlResult := rule_testing.RunTyped(t, NoUnsafeEnumComparison, noUnsafeEnumComparisonFile, withEnum)
	rule_testing.ExpectFindings(t, controlResult, "mismatchedCondition")
}

// TestNoUnsafeEnumComparisonReadsUnionsConstituentwise pins which reading of `isTypeFlagSet` this
// rule needs, because the shelf offers only one of the two and upstream uses both.
//
// `enum-utils/shared.ts` imports an `isTypeFlagSet` from `../../util` that decomposes a union and
// ORs its constituents' flags, AND `tsutils.isTypeFlagSet` which reads a type's own flags. Our
// `type_checking.IsTypeFlagSet` is the second. Reaching for it by name at both of upstream's call
// sites would be wrong at one of them.
//
// This rule gets both right structurally rather than by choosing a helper: `getEnumTypes` iterates
// union constituents and tests each, which is what decomposing means, and `getEnumValueType` is
// only ever handed a constituent.
//
// The input below is the one that separates the readings. A union of two DIFFERENT enums has no
// enum-literal flag on the composite type -- our helper answers false for it -- while both of its
// constituents carry the flag. If this rule ever stops iterating and starts testing the whole type,
// that union stops being seen as carrying enums at all, gate one exits early, and the finding is
// silently lost.
func TestNoUnsafeEnumComparisonReadsUnionsConstituentwise(t *testing.T) {
	t.Parallel()
	// A union of two enums compared against a bare string. Every constituent is an enum literal, so
	// the comparison is unsafe and must report.
	const source = "enum A { X = 'x' }\nenum B { Y = 'y' }\ndeclare const either: A | B;\nconst c = either === 'x';\n"

	result := rule_testing.RunTyped(t, NoUnsafeEnumComparison, noUnsafeEnumComparisonFile, source)
	rule_testing.ExpectFindings(t, result, "mismatchedCondition")

	// The control: the same union compared against a member of one of its own enums is safe, so a
	// rule that had started reporting every union would fail here.
	const safe = "enum A { X = 'x' }\nenum B { Y = 'y' }\ndeclare const either: A | B;\nconst c = either === A.X;\n"
	controlResult := rule_testing.RunTyped(t, NoUnsafeEnumComparison, noUnsafeEnumComparisonFile, safe)
	rule_testing.ExpectClean(t, controlResult)
}
