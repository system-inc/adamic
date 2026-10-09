package typescript

import (
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const noUnusedPrivateClassMembersFile = "/repository/source/Privates.ts"

func noUnusedPrivateClassMembersCaseName(index int) string {
	return "case" + strconv.Itoa(index)
}

// TestNoUnusedPrivateClassMembersStaysSilent is upstream's passing cases verbatim.
//
// Every case was re-measured against the installed 8.67.0 build with the file written exactly as
// this harness writes it, so these are evidence rather than a label copied from an array named
// `valid`. The list includes the `#private` cases upstream carries alongside the keyword ones,
// which this rule must leave to the core rule of the same name.
//
// The shapes worth naming, because each is a class of false positive a plausible port ships:
// a use through a NON-this receiver typed as the class, a use through an alias of `this` and
// through an alias of that alias, a write to a setter (which calls a body and so is a read), a
// destructuring read, and a computed key that reads the member to decide which property to take.
func TestNoUnusedPrivateClassMembersStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []string{
		"class Foo {}\n",
		"class Foo {\n  publicMember = 42;\n}\n",
		"class Foo {\n  public publicMember = 42;\n}\n",
		"class Foo {\n  protected publicMember = 42;\n}\n",
		"class C {\n  #usedInInnerClass;\n\n  method(a) {\n    return class {\n      foo = a.#usedInInnerClass;\n    };\n  }\n}\n",
		"class C {\n  private accessor accessorMember = 42;\n\n  method() {\n    return this.accessorMember;\n  }\n}\n",
		"class C {\n  private static staticMember = 42;\n\n  static method() {\n    return this.staticMember;\n  }\n}\n",
		"class C {\n  private static staticMember = 42;\n\n  method() {\n    return C.staticMember;\n  }\n}\n",
		"class Test1 {\n  constructor(private parameterProperty: number) {}\n  method() {\n    return this.parameterProperty;\n  }\n}\n",
		"class Test1 {\n  constructor(private readonly parameterProperty: number) {}\n  method() {\n    return this.parameterProperty;\n  }\n}\n",
		"class Test1 {\n  constructor(private readonly parameterProperty: number = 1) {}\n  method() {\n    return this.parameterProperty;\n  }\n}\n",
		"class Foo {\n  private prop: number;\n\n  method(thing: Foo) {\n    return thing.prop;\n  }\n}\n",
		"class Decimal {\n  private padded(places: number): string {\n    return String(places);\n  }\n  render(places: number): string {\n    const rounded = this.round(places);\n    return rounded.padded(places);\n  }\n  round(places: number): Decimal {\n    return this;\n  }\n}\n",
		"class Foo {\n  private static staticProp: number;\n\n  method(thing: typeof Foo) {\n    return thing.staticProp;\n  }\n}\n",
		"class Foo {\n  private prop: number;\n\n  method() {\n    const self = this;\n    return self.prop;\n  }\n}\n",
		"class Foo {\n  #privateMember = 42;\n  method() {\n    return this.#privateMember;\n  }\n}\n",
		"class Foo {\n  private privateMember = 42;\n  method() {\n    return this.privateMember;\n  }\n}\n",
		"class Foo {\n  #privateMember = 42;\n  anotherMember = this.#privateMember;\n}\n",
		"class Foo {\n  private privateMember = 42;\n  anotherMember = this.privateMember;\n}\n",
		"class Foo {\n  #privateMember = 42;\n  foo() {\n    anotherMember = this.#privateMember;\n  }\n}\n",
		"class Foo {\n  private privateMember = 42;\n  foo() {\n    anotherMember = this.privateMember;\n  }\n}\n",
		"class C {\n  #privateMember;\n\n  foo() {\n    bar((this.#privateMember += 1));\n  }\n}\n",
		"class C {\n  private privateMember;\n\n  foo() {\n    bar((this.privateMember += 1));\n  }\n}\n",
		"class Foo {\n  #privateMember = 42;\n  method() {\n    return someGlobalMethod(this.#privateMember);\n  }\n}\n",
		"class Foo {\n  private privateMember = 42;\n  method() {\n    return someGlobalMethod(this.privateMember);\n  }\n}\n",
		"class C {\n  #privateMember;\n\n  foo() {\n    return class {};\n  }\n\n  bar() {\n    return this.#privateMember;\n  }\n}\n",
		"class C {\n  private privateMember;\n\n  foo() {\n    return class {};\n  }\n\n  bar() {\n    return this.privateMember;\n  }\n}\n",
		"class Foo {\n  #privateMember;\n  method() {\n    for (const bar in this.#privateMember) {\n    }\n  }\n}\n",
		"class Foo {\n  private privateMember;\n  method() {\n    for (const bar in this.privateMember) {\n    }\n  }\n}\n",
		"class Foo {\n  #privateMember;\n  method() {\n    for (const bar of this.#privateMember) {\n    }\n  }\n}\n",
		"class Foo {\n  private privateMember;\n  method() {\n    for (const bar of this.privateMember) {\n    }\n  }\n}\n",
		"class Foo {\n  #privateMember;\n  method() {\n    [bar = 1] = this.#privateMember;\n  }\n}\n",
		"class Foo {\n  private privateMember;\n  method() {\n    [bar = 1] = this.privateMember;\n  }\n}\n",
		"class Foo {\n  #privateMember;\n  method() {\n    [bar] = this.#privateMember;\n  }\n}\n",
		"class Foo {\n  private privateMember;\n  method() {\n    [bar] = this.privateMember;\n  }\n}\n",
		"class C {\n  #privateMember;\n\n  method() {\n    ({ [this.#privateMember]: a } = foo);\n  }\n}\n",
		"class C {\n  private privateMember;\n\n  method() {\n    ({ [this.privateMember]: a } = foo);\n  }\n}\n",
		"class C {\n  set #privateMember(value) {\n    doSomething(value);\n  }\n  get #privateMember() {\n    return something();\n  }\n  method() {\n    this.#privateMember += 1;\n  }\n}\n",
		"class C {\n  private set privateMember(value) {\n    doSomething(value);\n  }\n  private get privateMember() {\n    return something();\n  }\n  method() {\n    this.privateMember += 1;\n  }\n}\n",
		"class Foo {\n  set #privateMember(value) {}\n\n  method(a) {\n    [this.#privateMember] = a;\n  }\n}\n",
		"class Foo {\n  private set privateMember(value) {}\n\n  method(a) {\n    [this.privateMember] = a;\n  }\n}\n",
		"class C {\n  get #privateMember() {\n    return something();\n  }\n  set #privateMember(value) {\n    doSomething(value);\n  }\n  method() {\n    this.#privateMember += 1;\n  }\n}\n",
		"class C {\n  private get privateMember() {\n    return something();\n  }\n  private set privateMember(value) {\n    doSomething(value);\n  }\n  method() {\n    this.privateMember += 1;\n  }\n}\n",
		"class Foo {\n  private privateMember;\n  private privateMember2;\n\n  method() {\n    const { privateMember, privateMember2 } = this;\n    console.log(privateMember, privateMember2);\n  }\n}\n",
		"class Foo {\n  private static staticMember = 1;\n  static method() {\n    const { staticMember } = this;\n    console.log(staticMember);\n  }\n}\n",
		"class Foo {\n  private privateMember = 1;\n  method() {\n    const { privateMember } = this;\n  }\n}\n",
		"class Foo {\n  private privateMember = 1;\n  method() {\n    const { privateMember: privateMember2 } = this;\n  }\n}\n",
		"class Foo {\n  private privateMember = 1;\n\n  method() {\n    let privateMember;\n    ({ privateMember } = this);\n  }\n}\n",
		"class Foo {\n  private privateMember = 1;\n\n  method() {\n    const foo = ({ privateMember } = this) => {};\n  }\n}\n",
		"class Foo {\n  private privateMember;\n\n  method() {\n    const { privateMember: used } = this;\n  }\n}\n",
		"class Foo {\n  #privateMember() {\n    return 42;\n  }\n  anotherMethod() {\n    return this.#privateMember();\n  }\n}\n",
		"class Foo {\n  private privateMember() {\n    return 42;\n  }\n  anotherMethod() {\n    return this.privateMember();\n  }\n}\n",
		"class C {\n  set #privateMember(value) {\n    doSomething(value);\n  }\n\n  foo() {\n    this.#privateMember = 1;\n  }\n}\n",
		"class C {\n  private set privateMember(value) {\n    doSomething(value);\n  }\n\n  foo() {\n    this.privateMember = 1;\n  }\n}\n",
	}
	for index, sourceText := range cases {
		t.Run(noUnusedPrivateClassMembersCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoUnusedPrivateClassMembers,
				noUnusedPrivateClassMembersFile, sourceText))
		})
	}
}

// noUnusedPrivateClassMembersFinding is one expected finding with the layers it can be wrong at.
//
// One message id, so the id proves little. The rendered text names the member, and the SPAN is what
// separates a property from a constructor parameter property: upstream reports the whole parameter
// there, so the span covers the type annotation, which no id assertion can see.
type noUnusedPrivateClassMembersFinding struct {
	wantSpan    string
	wantMessage string
}

// TestNoUnusedPrivateClassMembersFires is upstream's reporting cases verbatim, with every finding's
// rendered text and span taken from the installed 8.67.0 build.
func TestNoUnusedPrivateClassMembersFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		sourceText   string
		wantFindings []noUnusedPrivateClassMembersFinding
	}{
		{
			sourceText: "class C {\n  #unusedInOuterClass;\n\n  foo() {\n    return class D {\n      #unusedInOuterClass;\n\n      bar() {\n        return this.#unusedInOuterClass;\n      }\n    };\n  }\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "#unusedInOuterClass",
					wantMessage: "Private class member '#unusedInOuterClass' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class C {\n  #unusedOnlyInSecondNestedClass;\n\n  foo() {\n    return class {\n      #unusedOnlyInSecondNestedClass;\n\n      bar() {\n        return this.#unusedOnlyInSecondNestedClass;\n      }\n    };\n  }\n\n  baz() {\n    return this.#unusedOnlyInSecondNestedClass;\n  }\n\n  bar() {\n    return class {\n      #unusedOnlyInSecondNestedClass;\n    };\n  }\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "#unusedOnlyInSecondNestedClass",
					wantMessage: "Private class member '#unusedOnlyInSecondNestedClass' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class C {\n  #usedOnlyInTheSecondInnerClass;\n\n  method(a) {\n    return class {\n      #usedOnlyInTheSecondInnerClass;\n\n      method2(b) {\n        foo = b.#usedOnlyInTheSecondInnerClass;\n      }\n\n      method3(b) {\n        foo = b.#usedOnlyInTheSecondInnerClass;\n      }\n    };\n  }\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "#usedOnlyInTheSecondInnerClass",
					wantMessage: "Private class member '#usedOnlyInTheSecondInnerClass' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class C {\n  private accessor accessorMember = 42;\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "accessorMember",
					wantMessage: "Private class member 'accessorMember' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class C {\n  private static staticMember = 42;\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "staticMember",
					wantMessage: "Private class member 'staticMember' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class Test1 {\n  constructor(private parameterProperty: number) {}\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "parameterProperty: number",
					wantMessage: "Private class member 'parameterProperty' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class Test1 {\n  constructor(private readonly parameterProperty: number) {}\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "parameterProperty: number",
					wantMessage: "Private class member 'parameterProperty' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class Test1 {\n  constructor(private readonly parameterProperty: number = 1) {}\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "parameterProperty: number",
					wantMessage: "Private class member 'parameterProperty' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class C {\n  private usedOutsideClass;\n}\n\nconst instance = new C();\nconsole.log(instance.usedOutsideClass);\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "usedOutsideClass",
					wantMessage: "Private class member 'usedOutsideClass' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class C {\n  private usedOutsideClass;\n}\n\nconst instance = new C();\nconsole.log(instance['usedOutsideClass']);\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "usedOutsideClass",
					wantMessage: "Private class member 'usedOutsideClass' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class Foo {\n  #privateMember = 5;\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "#privateMember",
					wantMessage: "Private class member '#privateMember' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class Foo {\n  private privateMember = 5;\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "privateMember",
					wantMessage: "Private class member 'privateMember' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class First {}\nclass Second {\n  #privateMember = 5;\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "#privateMember",
					wantMessage: "Private class member '#privateMember' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class First {}\nclass Second {\n  private privateMember = 5;\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "privateMember",
					wantMessage: "Private class member 'privateMember' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class First {\n  #privateMember = 5;\n}\nclass Second {}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "#privateMember",
					wantMessage: "Private class member '#privateMember' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class First {\n  private privateMember = 5;\n}\nclass Second {}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "privateMember",
					wantMessage: "Private class member 'privateMember' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class First {\n  #privateMember = 5;\n  #privateMember2 = 5;\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "#privateMember",
					wantMessage: "Private class member '#privateMember' is defined but never used.",
				},
				{
					wantSpan:    "#privateMember2",
					wantMessage: "Private class member '#privateMember2' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class First {\n  private privateMember = 5;\n  private privateMember2 = 5;\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "privateMember",
					wantMessage: "Private class member 'privateMember' is defined but never used.",
				},
				{
					wantSpan:    "privateMember2",
					wantMessage: "Private class member 'privateMember2' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class Foo {\n  #privateMember = 5;\n  method() {\n    this.#privateMember = 42;\n  }\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "#privateMember",
					wantMessage: "Private class member '#privateMember' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class Foo {\n  private privateMember = 5;\n  method() {\n    this.privateMember = 42;\n  }\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "privateMember",
					wantMessage: "Private class member 'privateMember' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class Foo {\n  #privateMember = 5;\n  method() {\n    this.#privateMember += 42;\n  }\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "#privateMember",
					wantMessage: "Private class member '#privateMember' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class Foo {\n  private privateMember = 5;\n  method() {\n    this.privateMember += 42;\n  }\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "privateMember",
					wantMessage: "Private class member 'privateMember' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class C {\n  #privateMember;\n\n  foo() {\n    this.#privateMember++;\n  }\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "#privateMember",
					wantMessage: "Private class member '#privateMember' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class C {\n  private privateMember;\n\n  foo() {\n    this.privateMember++;\n  }\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "privateMember",
					wantMessage: "Private class member 'privateMember' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class Foo {\n  #privateMember() {}\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "#privateMember",
					wantMessage: "Private class member '#privateMember' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class Foo {\n  private privateMember() {}\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "privateMember",
					wantMessage: "Private class member 'privateMember' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class Foo {\n  #privateMember() {}\n  #privateMemberUsed() {\n    return 42;\n  }\n  publicMethod() {\n    return this.#privateMemberUsed();\n  }\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "#privateMember",
					wantMessage: "Private class member '#privateMember' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class Foo {\n  private privateMember() {}\n  private privateMemberUsed() {\n    return 42;\n  }\n  publicMethod() {\n    return this.privateMemberUsed();\n  }\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "privateMember",
					wantMessage: "Private class member 'privateMember' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class Foo {\n  set #privateMember(value) {}\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "#privateMember",
					wantMessage: "Private class member '#privateMember' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class Foo {\n  private set privateMember(value) {}\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "privateMember",
					wantMessage: "Private class member 'privateMember' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class Foo {\n  #privateMember;\n  method() {\n    for (this.#privateMember in bar) {\n    }\n  }\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "#privateMember",
					wantMessage: "Private class member '#privateMember' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class Foo {\n  private privateMember;\n  method() {\n    for (this.privateMember in bar) {\n    }\n  }\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "privateMember",
					wantMessage: "Private class member 'privateMember' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class Foo {\n  #privateMember;\n  method() {\n    for (this.#privateMember of bar) {\n    }\n  }\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "#privateMember",
					wantMessage: "Private class member '#privateMember' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class Foo {\n  private privateMember;\n  method() {\n    for (this.privateMember of bar) {\n    }\n  }\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "privateMember",
					wantMessage: "Private class member 'privateMember' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class Foo {\n  #privateMember;\n  method() {\n    ({ x: this.#privateMember } = bar);\n  }\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "#privateMember",
					wantMessage: "Private class member '#privateMember' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class Foo {\n  private privateMember;\n  method() {\n    ({ x: this.privateMember } = bar);\n  }\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "privateMember",
					wantMessage: "Private class member 'privateMember' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class Foo {\n  #privateMember;\n  method() {\n    [...this.#privateMember] = bar;\n  }\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "#privateMember",
					wantMessage: "Private class member '#privateMember' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class Foo {\n  private privateMember;\n  method() {\n    [...this.privateMember] = bar;\n  }\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "privateMember",
					wantMessage: "Private class member 'privateMember' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class Foo {\n  #privateMember;\n  method() {\n    [this.#privateMember = 1] = bar;\n  }\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "#privateMember",
					wantMessage: "Private class member '#privateMember' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class Foo {\n  private privateMember;\n  method() {\n    [this.privateMember = 1] = bar;\n  }\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "privateMember",
					wantMessage: "Private class member 'privateMember' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class Foo {\n  #privateMember;\n  method() {\n    [this.#privateMember] = bar;\n  }\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "#privateMember",
					wantMessage: "Private class member '#privateMember' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class Foo {\n  private privateMember;\n  method() {\n    [this.privateMember] = bar;\n  }\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "privateMember",
					wantMessage: "Private class member 'privateMember' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class Foo {\n  private privateMember;\n\n  method() {\n    const { unused: privateMember } = this;\n  }\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "privateMember",
					wantMessage: "Private class member 'privateMember' is defined but never used.",
				},
			},
		},
		{
			sourceText: "const foo = 'bar';\nclass Foo {\n  private foo = 1;\n\n  method() {\n    const { [foo]: test } = this;\n  }\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "foo",
					wantMessage: "Private class member 'foo' is defined but never used.",
				},
			},
		},
		{
			sourceText: "const foo = 'bar';\nclass Foo {\n  private foo = 1;\n  private bar = 2;\n\n  method() {\n    const { [foo]: test } = this;\n  }\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "foo",
					wantMessage: "Private class member 'foo' is defined but never used.",
				},
				{
					wantSpan:    "bar",
					wantMessage: "Private class member 'bar' is defined but never used.",
				},
			},
		},
		{
			// Measured, not imported. `this` inside a nested class is the nested class, so neither
			// row reaches the outer member: the first reads the nested class's own `a`, the second
			// reads nothing the checker resolves. Both report, here and on the installed 8.67.0 build.
			// A receiver of the OUTER type does reach it from the nested class, which
			// TestNoUnusedPrivateClassMembersDivergesOnReceiverReach pins (#yfkhkvy).
			sourceText: "class C {\n  private a = 1;\n  m() {\n    return class D {\n      private a = 2;\n      n() { return this.a; }\n    };\n  }\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "a",
					wantMessage: "Private class member 'a' is defined but never used.",
				},
			},
		},
		{
			sourceText: "class C {\n  private a = 1;\n  m() {\n    return class D {\n      n() { return this.a; }\n    };\n  }\n}\n",
			wantFindings: []noUnusedPrivateClassMembersFinding{
				{
					wantSpan:    "a",
					wantMessage: "Private class member 'a' is defined but never used.",
				},
			},
		},
	}
	for index, testCase := range cases {
		t.Run(noUnusedPrivateClassMembersCaseName(index), func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, NoUnusedPrivateClassMembers,
				noUnusedPrivateClassMembersFile, testCase.sourceText)

			wantIds := make([]string, len(testCase.wantFindings))
			for position := range testCase.wantFindings {
				wantIds[position] = "unusedPrivateClassMember"
			}
			rule_testing.ExpectFindings(t, result, wantIds...)

			onDisk := strings.TrimSpace(testCase.sourceText) + "\n"
			for position, want := range testCase.wantFindings {
				diagnostic := result.Diagnostics[position]
				gotSpan := onDisk[diagnostic.Range.Pos():diagnostic.Range.End()]
				if gotSpan != want.wantSpan {
					t.Fatalf("finding %d span: expected %q, got %q", position, want.wantSpan, gotSpan)
				}
				if diagnostic.Message.Description != want.wantMessage {
					t.Fatalf("finding %d message: expected %q, got %q", position, want.wantMessage,
						diagnostic.Message.Description)
				}
			}
		})
	}
}

// TestNoUnusedPrivateClassMembersDivergesOnReceiverReach records where this port and upstream
// disagree, and which way.
//
// Upstream reaches a use through scope analysis and follows exactly ONE hop of `const self = this`.
// Measured on the installed 8.67.0 build across three depths:
//
//	const a = this; a.prop            clean
//	const a = this; const b = a; b.prop      REPORTS
//	const a = this; const b = a; const c = b; c.prop   REPORTS
//
// The member is genuinely read in all three, so the second and third are upstream false positives:
// it accuses working code of being dead, which for an unused-thing rule is the expensive direction.
// We resolve the property name through the checker instead of tracking aliases, so depth costs
// nothing and all three are clean here.
//
// The SAME divergence has a second face, found on the real tree rather than by construction. A
// receiver held in a LOCAL is not followed either:
//
//	class Foo { private helper() {...}  m() { const other = this.clone(); return other.helper(); } }
//
// Upstream reports `helper` there and is silent when the receiver is a PARAMETER annotated as the
// class, which its own corpus carries as a passing case. So the line it draws is not "only this"
// but "this, a one-hop alias, or an annotated parameter", and a local holding the same type falls
// outside it. Measured against the installed build both ways.
//
// That is not academic. The dry run over this tree found four such members with upstream and three
// with this rule, and the difference is one real method called as `rounded.toFixedPointStringPadded(...)`
// eight lines above its declaration. Upstream calls it dead; it is not.
//
// This is a deliberate decline to reproduce a defect, stated rather than silent. Both imported
// cases asserting such a report are removed from the firing table above and live here instead, at
// the verdict this rule actually produces. For an unused-thing rule the direction matters: a false
// positive asks a reader to delete working code, which is the expensive failure.
func TestNoUnusedPrivateClassMembersDivergesOnReceiverReach(t *testing.T) {
	t.Parallel()

	cases := []string{
		"class Foo {\n  private prop: number = 1;\n  method() {\n    const self1 = this;\n    return self1.prop;\n  }\n}\n",
		"class Foo {\n  private prop: number = 1;\n  method() {\n    const self1 = this;\n    const self2 = self1;\n    return self2.prop;\n  }\n}\n",
		"class Foo {\n  private prop: number = 1;\n  method() {\n    const a = this;\n    const b = a;\n    const c = b;\n    return c.prop;\n  }\n}\n",
	}
	cases = append(cases,
		// The foreign-receiver face of the same divergence, reduced from the real finding on
		// `Decimal.ts` that upstream reports and this rule does not.
		"class Foo {\n  private helper(): number {\n    return 1;\n  }\n  clone(): Foo {\n    return this;\n  }\n  method() {\n    const other = this.clone();\n    return other.helper();\n  }\n}\n",
		// A nested class is still inside the declaring body, so a receiver of the outer type reads
		// the member from there. Upstream reports the captured `this` and is silent on the
		// annotated parameter, since it reads a receiver's class from syntax; both are reads.
		"class C {\n  private a = 1;\n  m() {\n    const outer = this;\n    return class D {\n      n() { return outer.a; }\n    };\n  }\n}\n",
		"class Outer {\n  private secret = 1;\n  make() {\n    return class {\n      read(outer: Outer) { return outer.secret; }\n    };\n  }\n}\n",
	)
	for index, sourceText := range cases {
		t.Run(noUnusedPrivateClassMembersCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoUnusedPrivateClassMembers,
				noUnusedPrivateClassMembersFile, sourceText))
		})
	}

	// The control. Without it every row above is satisfied by a rule that cannot report at all.
	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoUnusedPrivateClassMembers,
		noUnusedPrivateClassMembersFile, "class Foo {\n  private prop: number = 1;\n}\n"),
		"unusedPrivateClassMember")
}

// TestNoUnusedPrivateClassMembersDivergesOnDeclareField pins the second place this port parts from
// upstream: a `declare` field is never reported, since it emits nothing and exists to make the class
// nominal. Measured on the installed 8.67.0 build: upstream reports the first and fourth cases and is
// silent on the other two, so those two rows are the divergence and the rest are agreement.
//
//	declare private brand, never read     upstream REPORTS, we are clean: the brand is the design
//	declare private brand, read           clean in both: the read keeps it alive either way
//	declare public brand                  clean in both: not private, so neither rule judges it
//	declare static private brand          upstream REPORTS, we are clean: it emits nothing either
//
// The firing half matters as much: a real `private` field beside the brand must still report, or the
// arm could be skipping the whole class rather than the one member.
func TestNoUnusedPrivateClassMembersDivergesOnDeclareField(t *testing.T) {
	t.Parallel()

	cases := []string{
		"class TypedBinding {\n  declare private readonly __brand: 'TypedBinding';\n}\n",
		"class TypedBinding {\n  declare private readonly __brand: 'TypedBinding';\n  read() {\n    return this.__brand;\n  }\n}\n",
		"class TypedBinding {\n  declare readonly __brand: 'TypedBinding';\n}\n",
		"class TypedBinding {\n  declare private static readonly __brand: 'TypedBinding';\n}\n",
	}
	for index, sourceText := range cases {
		t.Run(noUnusedPrivateClassMembersCaseName(index), func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoUnusedPrivateClassMembers,
				noUnusedPrivateClassMembersFile, sourceText))
		})
	}

	// The control: the brand stays silent and the real dead field beside it still reports, once.
	result := rule_testing.RunTyped(t, NoUnusedPrivateClassMembers, noUnusedPrivateClassMembersFile,
		"class TypedBinding {\n  declare private readonly __brand: 'TypedBinding';\n  private readonly unused: number = 1;\n}\n")
	rule_testing.ExpectFindings(t, result, "unusedPrivateClassMember")
	if description := result.Diagnostics[0].Message.Description; !strings.Contains(description, "'unused'") {
		t.Fatalf("expected the finding on 'unused', got %q", description)
	}
}
