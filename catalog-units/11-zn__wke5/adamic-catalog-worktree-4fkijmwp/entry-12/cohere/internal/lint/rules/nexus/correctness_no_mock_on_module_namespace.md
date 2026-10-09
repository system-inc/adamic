# `nexus/correctness-no-mock-on-module-namespace`

| | |
|---|---|
| **Recommendation** | **Yes, at `error`.** Zero findings today, and every finding is a test that throws before its first assertion |
| Findings | **ahra 0** (measured 2026-10-03; research count 0, and 2 before `d75fe890`) |
| Measured precision | no findings to read; the shape always throws, so a finding cannot be false in a file that runs as an ES module |
| Plugin | cohere-native, `nexus` |
| Auto-fixable | no |
| Needs type information | yes, to resolve the mock call to node:test's `MockTracker` and the argument to a namespace import; without a checker the rule registers nothing |

## What it checks

node:test's `mock.method`, `mock.getter`, `mock.setter` or `mock.property` (through `mock`, `test.mock` or a
context's `t.mock`) called with an `import * as` namespace as the object to patch, in a file that runs as an
ES module. A module namespace's properties are non-configurable, so node's `ObjectDefineProperty` throws
`Cannot redefine property`, or, for `getter` and `setter`, the missing accessor throws `must be a method`
first. The doc comment has the details: which call and which argument count, and when a file is an ES
module.

## Where it came from

`modules/pensieve/PensieveTransportSecurity.test.ts` in ahra, before `d75fe890`. Both transport security tests
called `test.mock.method(NodeChildProcess, 'execFileSync', ...)` on `import * as NodeChildProcess from
'node:child_process'`, and both threw. The fix mocks the CommonJS exports object reached through
`NodeModule.createRequire(import.meta.url)('node:child_process')`, typed `typeof NodeChildProcess`, and calls
`syncBuiltinESMExports()`. Found by the own-history pass of the new-rules sweep (`#tevhg3f`, item C), built in
`#j03vwm6`. `nexus/import-require-node-namespace` makes the namespace import the house spelling for every
builtin, which is why this shape is the first one a test author writes.

## Existing rules checked

- `no-import-assign` (on in ahra) already reports `Object.assign(namespace, ...)`, `Object.defineProperty`
  and the `Reflect` mutators on a namespace, and assignment to a namespace member. It does not know
  node:test's `MockTracker`, whose signature accepts any object.
- No typescript-eslint, unicorn or sonarjs rule covers a mocking API on a namespace. jest's `spyOn` is left
  out on purpose: under jest's CommonJS transform the namespace is a plain copied object, and
  `libraries/structure/libraries/nexus/source/protocols/oauth/OAuth1a.test.ts:26` spies on one and works.

## What it declines

- A namespace copied into a local first, or reached through `await import(...)`: following values is not
  what the rule claims. Missed, never false.
- `.cts`/`.cjs` files, programs with `module: commonjs` (and AMD, UMD, System), and `.ts` files under
  `node16`/`nodenext`, whose format comes from a `package.json` the rule does not read. ahra is
  `"module": "esnext"` in a `"type": "module"` package, so every test file is read.

## Verification

- Fixtures from the real site both ways: `PensieveTransportSecurity.test.ts` before `d75fe890` (two
  findings) and after (silent: the mocked object is the CommonJS exports typed `typeof NodeChildProcess`, and
  the namespace import is `import type`). 7 more firing shapes (a project module's namespace, `getter`,
  `setter`, `property`, `t.mock`, parentheses, `as`, `!`); 10 silent calls (a copy typed `typeof` the
  namespace, a spread copy, a default import, an `import type` namespace, a shadowing parameter, `jest.spyOn`,
  another package's `MockTracker`, a local `MockTracker`, `mock.fn`, `mock.restoreAll`); the extension cases
  (`.mts` under CommonJS fires, `.cts` under ESNext is silent) and the `module` cases (`esnext`, `es2020`,
  `preserve` fire; `commonjs`, `nodenext` are silent).
- Mutation check, each mutant a copy through `go test -overlay -count=1`, all 8 killed: the namespace-kind
  test, the type-only test, the MockTracker declaration test, the `node:test` module name, the ES module gate,
  the `.cts` extension, `preserve`, and looking through `as`.
- The built binary, on a scratch project resolving ahra's own `@types/node` 26.2.0, reports the planted
  `test.mock.method(NodeChildProcess, ...)` and leaves its fixed form alone.
