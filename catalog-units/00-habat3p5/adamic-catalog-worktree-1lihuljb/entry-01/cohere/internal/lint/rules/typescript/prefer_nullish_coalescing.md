# `@typescript-eslint/prefer-nullish-coalescing`

| | |
|---|---|
| **Recommendation** | **Yes** |
| Violations in ahra | **64** with every check on; none under the house sets' interim options |
| Plugin | `@typescript-eslint` |
| Auto-fixable | no (a suggestion on every finding) |
| Needs type information | yes |

## What it checks

Enforce using the nullish coalescing operator instead of logical assignments or chaining. Three
shapes, each reported where upstream anchors it:

- **`||` and `||=` on a nullable left side** (`preferNullishOverOr`), at the operator. `a || b`
  falls through on every falsy value, so an empty string, a zero and a false take the right side
  along with null and undefined. `ignoreConditionalTests` defaults to true, so a `||` inside an
  `if`, loop or ternary test is not reported.
- **A ternary that tests its subject for null or undefined and returns it** (`preferNullishOverTernary`),
  at the whole ternary: `a !== null && a !== undefined ? a : b`, `a == null ? b : a`, `a ? a : b`. A
  test that rules out only one of null and undefined strictly is reported when the type cannot hold
  the other. Off with `ignoreTernaryTests`.
- **An `if` with no `else` whose body is one assignment to the subject it tested**
  (`preferNullishOverAssignment`), at the whole `if`: `if (a == null) a = b;` suggests `a ??= b;`,
  carrying any comments in the body out with it. Off with `ignoreIfStatements`.

The whole-file `noStrictNullCheck` complaint is not implemented. It needs a program compiled without
`strictNullChecks`, which no tree we lint has.

## Where cohere is deliberately quieter than ESLint

The `if` check is silent inside a function React Compiler compiles, decided by the same
`IsInsideComponentOrHook` that `logical-assignment-operators` and the react rules use. Its repair is
`??=`, and babel-plugin-react-compiler 1.0.0 refuses every logical assignment shorthand (`Handle ??=
operators in AssignmentExpression`), so the suggestion costs the whole component or hook its
compilation. The waves that took it showed it: SecretRow (ahra), WebSocketViaSharedWorkerProviderInternal
(Structure) and ChatReasoningAndTools (www) each went from CompileSuccess to CompileError on one `??=`,
and compiled again in the long form (#cn8sthd). `logical-assignment-operators` already stays silent
there, so without this gate the two rules pulled one site in opposite directions.

The option is ours, `reactCompiler`, on by default because every tree we lint runs the compiler; a
project without it sets `{"reactCompiler": false}`. The ternary and `||` checks are not gated: `??` is
an expression the compiler lowers, and a `||=` was refused before any rewrite. ESLint runs the same
gate through `NexusTypeScriptEsLintPlugin`'s wrapper of this rule.

## Why this recommendation

`??` says what the code means where the left side can be missing, and leaves the valid falsy values
alone. The ternary and `if` forms say the same thing at greater length, and name the subject twice,
so the two mentions can drift apart.

## The interim options

The ternary and `if` checks shipped late. Until 2026-10-04 the decoder parsed `ignoreTernaryTests` and
`ignoreIfStatements` while neither check existed, so the options did nothing and this page described
checks that never ran. ESLint, which does run them, found 64 sites in ahra that cohere never
reported (#10kqgs6). The house sets turned the rule on with both options set to true in both engines,
so neither engine reports those shapes while the sites are fixed. The options come off once ahra, www
and api read 0 in both engines.

Measured with both options off, against ESLint with typescript-eslint 8.67.0 on the same trees, file
by file, line and column:

| Tree | Ternary | `if` | Agreement with ESLint |
|---|---|---|---|
| ahra | 55 | 9 | 64 of 64 |
| www | 33 | 5 | 38 of 38 |
| api | 21 | 1 | 22 of 22 |

## Violations

64 in ahra with every check on. Two of them:

**`libraries/structure/source/modules/account/Account.ts:59`**

```
accountQueryData.emailAddress !== undefined ? accountQueryData.emailAddress : this.emailAddress,
```

> Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read

**`libraries/structure/source/services/cookie/CookieService.ts:46`**

```
if(cookies[name] === undefined) {
    cookies[name] = value;
}
```

> Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read

### What fixing looks like

Not auto-fixable. Every finding carries a suggestion, `accountQueryData.emailAddress ?? this.emailAddress`
and `cookies[name] ??= value;` above. It is a suggestion because a truthiness ternary (`a ? a : b`)
changes meaning for a falsy non-nullish `a`, and the `if` form accepts any assignment operator in its
body, as upstream's does, so `if (!a) a += b` suggests `a ??= b`. Each site needs a reader.
