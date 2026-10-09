package core

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// The shapes api's 15 waiting sites take, reduced: declarations that reference each other, where no
// ordering removes the forward reference and the forward reference cannot run before its target.
func TestNoUseBeforeDefineExemptsAnUnorderableCycleThatCannotRunEarly(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		// GraphQlMiddleware's collectSelections and OrmEntityHydrator's hydrate pair: two function
		// declarations calling each other. Both are initialized with their bodies before anything runs.
		{"two function declarations calling each other", `
export function collectSelections(node: { children: unknown[] }): number {
    return visitSelection(node);
}
function visitSelection(node: { children: unknown[] }): number {
    return node.children.length === 0 ? 0 : collectSelections({ children: [] });
}
`},
		// Three functions in a ring: the cycle is transitive.
		{"three function declarations in a ring", `
export function first(): number { return second(); }
function second(): number { return third(); }
function third(): number { return first(); }
`},
		// The ORM relation fixtures: each entity's relation decorator names the other through a thunk,
		// which the decorator stores and calls after both classes exist.
		{"two entities whose relation thunks name each other", `
declare function OrmOneToMany(target: () => unknown, options?: object): (target: unknown, context: unknown) => void;
declare function OrmManyToOne(target: () => unknown): (target: unknown, context: unknown) => void;
export class Author {
    @OrmOneToMany(() => Post, { inverseSide: 'author' })
    posts?: unknown;
}
export class Post {
    @OrmManyToOne(() => Author)
    author?: unknown;
}
`},
		// The test fixtures declare their entities inside a test body, so the cycle is in a block.
		{"the same pair declared inside a function body", `
declare function OrmOneToMany(target: () => unknown): (target: unknown, context: unknown) => void;
export function buildSchema(): unknown[] {
    class Author {
        @OrmOneToMany(() => Post)
        posts?: unknown;
    }
    class Post {
        @OrmOneToMany(() => Author)
        authors?: unknown;
    }
    return [Author, Post];
}
`},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, NoUseBeforeDefine, "fixture.ts", testCase.source))
		})
	}
}

// What the exemption must not reach: a forward reference that can be reordered away, or one that can
// run before its target exists and so lands in the temporal dead zone.
func TestNoUseBeforeDefineStillReportsWhatCanRunEarlyOrBeReordered(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
	}{
		// No cycle: moving the declaration up fixes it.
		{"a function declaration called before it, with no cycle", `
export function run(): number { return helper(); }
function helper(): number { return 1; }
`},
		{"a decorator thunk naming a class that does not name back", `
declare function OrmOneToMany(target: () => unknown): (target: unknown, context: unknown) => void;
export class Author {
    @OrmOneToMany(() => Post)
    posts?: unknown;
}
export class Post {}
`},
		// In a cycle, but runs as the class is defined.
		{"a static initializer reading the other class of a cycle", `
export class Author {
    static related = Post.name;
}
export class Post {
    static related = Author.name;
}
`},
		// In a cycle, but an ordinary call may call its callback on the spot.
		{"a thunk passed to an ordinary call in a cycle", `
declare function register(target: () => unknown): unknown;
export class Author {
    static relation = register(() => Post);
}
export class Post {
    static relation = register(() => Author);
}
`},
		// In a cycle, but the callback takes parameters, so it is not the deferred-thunk idiom.
		{"a decorator callback with parameters in a cycle", `
declare function OrmOneToMany(target: (kind: string) => unknown): (target: unknown, context: unknown) => void;
export class Author {
    @OrmOneToMany((kind) => (kind.length > 0 ? Post : undefined))
    posts?: unknown;
}
export class Post {
    @OrmOneToMany(() => Author)
    authors?: unknown;
}
`},
		// In a cycle, but invoked immediately while the module loads.
		{"an immediately invoked function in a cycle", `
export const first = (() => second)();
export const second = (): unknown => first;
`},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, NoUseBeforeDefine, "fixture.ts", testCase.source),
				"usedBeforeDefined")
		})
	}
}
