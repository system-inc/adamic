# `nexus/correctness-no-collection-misuse`

| | |
|---|---|
| **Recommendation** | **Yes, at `error`.** Zero findings today; it guards four shapes that run without throwing and quietly do nothing |
| Findings | **ahra 0** (measured 2026-10-03; research count 0 for each shape) |
| Measured precision | no findings to read |
| Plugin | cohere-native, `nexus` |
| Auto-fixable | no |
| Needs type information | yes, every check rests on an operand's type; without a checker the rule registers nothing |

## What it checks

Four message ids, one rule, because each is a built-in collection used in a way its default-library type
says cannot work:

| Id | Shape | Why it is always wrong |
|---|---|---|
| `inOnArray` | `'Admin' in roles` on an array or tuple, the left side a string literal (or a union of them) that is neither a canonical index nor a property the array type declares | `in` asks for a property, never an element: always false |
| `impossibleSizeComparison` | `.length` of an array, tuple, string or typed array, or `.size` of a Map or Set, typed `number` alone, compared with a literal so that `< 0`, `<= -1`, `=== -1`, `== -1` (always false) or `>= 0`, `> -1`, `!== -1`, `!= -1` (always true), either way round | a length or size is never negative |
| `bracketAccessOnCollection` | `map[key]` read or written on a Map, ReadonlyMap, WeakMap, Set, ReadonlySet or WeakSet, the key typed so it cannot name a member (a string literal that is not one, a number, a boolean, a bigint, `null`, `undefined`) | brackets reach own properties, not entries |
| `objectMethodOnCollection` | `Object.keys`, `values`, `entries` or `getOwnPropertyNames` on one of those six, with `Object` and the method resolving to the default library | a collection's entries are not properties: always empty |

The collection types are matched by the symbol the default library declares, so a subclass (which may carry
fields of its own), a project class named `Map` and an intersection are left alone.

## Where it came from

The JavaScript catalog pass of the new-rules sweep (`#tevhg3f`, items 9 and 10), built in `#j03vwm6`. The
probes counted zero of every shape in ahra and Structure; the fixtures are synthetic.

## Existing rules checked

- `@typescript-eslint/no-for-in-array` (on) covers `for (const key in array)`, not the `in` operator.
- `@typescript-eslint/prefer-includes` (on) covers `indexOf(...)` against `-1` and `0` (`!== -1`, `>= 0`,
  `< 0`), a style rewrite, not `> 0` and not the `in` operator.
- `no-constant-binary-expression`, `@typescript-eslint/no-unnecessary-condition`: neither reasons about a
  length's range. TypeScript's own `noImplicitAny` rejects most bracket reads on a Map (`TS7053`); this
  catches them where that is off or suppressed.
- No unicorn or sonarjs plugin is ported in cohere (`no-in-misuse`, `no-collection-size-mischeck`,
  `no-impossible-length-comparison`, `no-collection-bracket-access`, `no-object-methods-with-collections`
  are the sources).

## What it declines

- **`indexOf(...) > 0`** (sonarjs `index-of-compare-to-positive-number`), listed in the task. Not exact:
  `> 0` also means "found, and not at the start". `fontSize.indexOf(fontSizeType) > 0` in scriptaculous
  (`~/Projects/yougetsignal/.../effects.js:430`) requires a number before the unit, and TypeScript's own
  `tools/scripts/gen/generatedFile.test.mts:75` asserts `calls.indexOf("build") > 0`, that build ran and was
  not first. No type tells the two intents apart. ahra has no site either way.
- `in` with a wide `string`, a number or a symbol on the left (`index in sparse` finds holes; a wide string
  may hold an index), a wide `string` or `symbol` key in brackets (it may name `size` or `Symbol.iterator`),
  `items?.length` (may be `undefined`), `ArrayLike` and plain `{ length: number }` objects (their length is
  whatever the object says), `+0` and named constants as the literal. Each is a missed finding at worst.

## Verification

- Firing: 4 `in`, 10 size comparisons, 3 bracket accesses, 4 `Object` methods. Silent: 4 fixed counterparts
  (`.includes()` twice, `.get()`, `Array.from(cache.keys())`), 11 `in` shapes, 11 comparisons, 16 Map/Set
  shapes including a subclass, a shadowing `class Map`, a shadowing `Object` and a parameter named `Object`.
- Mutation check, all 11 killed: the index-literal test, the declared-property test, the array type test,
  the decided-comparison table and its boundary, the mirror for a literal on the left, the length receiver
  types, the collections from the default library, the member-name key test, the wide-key refusal, the
  `Object` resolution, and the collection argument. A check that the size access is typed `number` alone was
  written, found unreachable by its surviving mutant (the receiver test already refuses `items?.length` on a
  list that may be `undefined`), and removed.
- The built binary reports the four planted shapes on a scratch project with ahra's compiler options.
