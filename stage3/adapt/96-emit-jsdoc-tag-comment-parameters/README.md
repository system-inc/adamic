# JSDoc tag comment parameters, adaptation 96-emit

Emit, task #v879py5. This is type-only: the emitted JavaScript does not change.

## The edit

`adapt.cjs` widens one parameter in each of four memoized arrows in
`src/compiler/factory/nodeFactory.ts`, from `comment?: NodeArray<JSDocComment>`
to `comment?: string | NodeArray<JSDocComment>`:

| Arrow | Upstream 6.0.3 line | Worker it forwards to |
|---|---|---|
| `getJSDocSimpleTagCreateFunction` | 508 | `createJSDocSimpleTagWorker` (:5468) |
| `getJSDocSimpleTagUpdateFunction` | 509 | `updateJSDocSimpleTagWorker` (:5481) |
| `getJSDocTypeLikeTagCreateFunction` | 510 | `createJSDocTypeLikeTagWorker` (:5494) |
| `getJSDocTypeLikeTagUpdateFunction` | 511 | `updateJSDocTypeLikeTagWorker` (:5506) |

It parses nodeFactory.ts with stock TypeScript 6.0.3 and refuses before any
write unless it finds exactly one declaration of each of the four, each
`memoizeOne` of an arrow returning an arrow whose parameter list is upstream's
own, text for text. A parameter already widened keeps its bytes, so a second run
changes nothing. No source coordinate is a rewrite address.

The directory sorts after 91-patcher-discard, whose pin covers
createNodeFactory's exact text before its own edit. A first cut numbered 73 ran
before 91 and 91's pin refused the changed function (J-spvkw8k), so this one
runs after it.

## The lie, and why widening is the truth

Each arrow is handed out by a lazy getter in createNodeFactory's factory literal
(nodeFactory.ts:877 to 952, every third line: 13 tags, a create and an update
each) as a `NodeFactory` member. types.ts declares every one of those members as
taking `comment?: string | NodeArray<JSDocComment>` (for example :9177
`createJSDocTypeTag`, :9195 `createJSDocAuthorTag`). tsc relates a method's
parameters both ways, so it lets a public caller's string reach an arrow typed
to refuse it. Adamic relates them one way and refused all 26 getters: "a
function taking NodeArray<JSDocComment> | undefined seen as one taking string |
NodeArray<JSDocComment> | undefined".

The string does no harm at run time: each arrow's body does nothing with
`comment` but pass it to its worker, and all four workers already take
`string | NodeArray<JSDocComment>`. The narrow annotation says less than the
code does. Widening it makes the arrow's type the type of what it does.

## Certificate

Certified by the compiled program's gate, as adaptation 72 is. The premise:
each arrow only forwards `comment` to a worker typed with `string`. If that were
wrong, Adamic would judge the forwarding call at its own relation site inside
the compiled program and refuse it there, not trust it.

## Public consequence

None. The four arrows are file-local consts inside createNodeFactory, never
exported, so `typescript.d.ts` is byte-identical. No sanction is needed.

## Measured, from J-tr6jzr8

Sealed, on azure-proof-7, at scratch fb51adb2 (this directory on main cf468df0),
before against after, upstream 6.0.3, 1246 s:

- **Sources:** of 748 src files only `compiler/factory/nodeFactory.ts` differs,
  on lines 508 to 511 alone.
- **TypeScript diagnostics:** 7057 over 605 roots before and after, 0 new and 0
  gone. Widening the parameter moves no inference: `T` at the getters infers as
  before, or a diagnostic or a refusal elsewhere would have moved.
- **Census:** latent total 3941 to 3915. The family goes 26 to 0. One other
  reason reads as +1 and -1, commandLineParser.ts:3253:24's `getPropFromRaw`
  `T = void`; it is the same site, whose reason text carries the before or after
  tree's own path. No other reason moved.
- **Emit:** of 1365 built files, typescript.js, typescript.d.ts, tsc.js and
  _tsc.js are byte-identical. Four differ: three source maps (_tsc.js.map,
  run.js.map, typescript.js.map) and compiler/tsconfig.tsbuildinfo, which carry
  the source text and its hashes.
- **Lane:** 398 units on the after tree, 0 failing; pass 106366, fail 1, as the
  plan holds. The lane guards (test_check's count, sanction freshness, the
  plan's runtime) and TestProduct_Stage3AdaptedTree pass.
- **Mutant:** the adaptation dropped is the before tree, and it brings the 26
  back.
