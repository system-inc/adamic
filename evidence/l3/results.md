# L3 comparison

Exact compiler revisions:

- Base main: e3aa0bb7642163ebd004a71ba40a4657d27eab71, built with `go build -o /tmp/adamic-e3aa0bb ./cmd/adamic`, exit 0.
- Head compiler/rule-g: e5531c8c7fd9fa0b5076bb3d0d0961e6954e9ce1, built with `go build -o /tmp/adamic-e5531c8 ./cmd/adamic`, exit 0.

Go 1.27.1, clang 20.1.8, Node 24.19.0. Source l3.ts and l3.a have identical bytes and match the supplied four lines. Detached test worktree /workspace/adamic-rule-g-2594029 was switched to each exact revision, with pinned submodules and clean source. It is restored to e5531c8 afterward (exit 0).

Both CLIs support native `build <source> -o <binary> --sanitize` and `js <source>`. JavaScript output would run via Node and the revision's oracle/node.mjs resolver for the emitted `adamic` runtime import. Native runtime sanitizer options are recorded in run.py but were never applied because neither compiler produced an executable.

| Compiler | Extension | Native --sanitize compile | Native run | JS compile | JS run |
|---|---|---|---|---|---|
| e3aa0bb | .ts | Exit 1: TS2532 | Not run; no executable | Exit 1: TS2532 | Not run; no JS output |
| e3aa0bb | .a | Exit 1: TS2532 | Not run; no executable | Exit 1: TS2532 | Not run; no JS output |
| e5531c8 | .ts | Exit 1: TS2532 | Not run; no executable | Exit 1: TS2532 | Not run; no JS output |
| e5531c8 | .a | Exit 1: TS2532 | Not run; no executable | Exit 1: TS2532 | Not run; no JS output |

All compilation stdout is empty. Every native and JS compile fails during loading at the possibly undefined indexed element, before lowering/code generation. No sanitizer-instrumented program was produced, no sanitizer runtime was run, and no sanitizer report exists.

## Complete compilation results

### e3aa0bb-ts-native-build

```sh
/tmp/adamic-e3aa0bb build /workspace/adamic/evidence/l3/l3.ts -o /tmp/l3-e3aa0bb-ts --sanitize
```

Exit 1. Stdout: empty. Complete stderr:

```text
/workspace/adamic/evidence/l3/l3.ts:4:45: error TS2532: Object is possibly 'undefined'.
```

### e3aa0bb-ts-js-build

```sh
/tmp/adamic-e3aa0bb js /workspace/adamic/evidence/l3/l3.ts
```

Exit 1. Stdout: empty. Complete stderr:

```text
/workspace/adamic/evidence/l3/l3.ts:4:45: error TS2532: Object is possibly 'undefined'.
```

### e3aa0bb-a-native-build

```sh
/tmp/adamic-e3aa0bb build /workspace/adamic/evidence/l3/l3.a -o /tmp/l3-e3aa0bb-a --sanitize
```

Exit 1. Stdout: empty. Complete stderr:

```text
/workspace/adamic/evidence/l3/l3.a:4:45: error TS2532: Object is possibly 'undefined'.
```

### e3aa0bb-a-js-build

```sh
/tmp/adamic-e3aa0bb js /workspace/adamic/evidence/l3/l3.a
```

Exit 1. Stdout: empty. Complete stderr:

```text
/workspace/adamic/evidence/l3/l3.a:4:45: error TS2532: Object is possibly 'undefined'.
```

### e5531c8-ts-native-build

```sh
/tmp/adamic-e5531c8 build /workspace/adamic/evidence/l3/l3.ts -o /tmp/l3-e5531c8-ts --sanitize
```

Exit 1. Stdout: empty. Complete stderr:

```text
/workspace/adamic/evidence/l3/l3.ts:4:45: error TS2532: Object is possibly 'undefined'.
```

### e5531c8-ts-js-build

```sh
/tmp/adamic-e5531c8 js /workspace/adamic/evidence/l3/l3.ts
```

Exit 1. Stdout: empty. Complete stderr:

```text
/workspace/adamic/evidence/l3/l3.ts:4:45: error TS2532: Object is possibly 'undefined'.
```

### e5531c8-a-native-build

```sh
/tmp/adamic-e5531c8 build /workspace/adamic/evidence/l3/l3.a -o /tmp/l3-e5531c8-a --sanitize
```

Exit 1. Stdout: empty. Complete stderr:

```text
/workspace/adamic/evidence/l3/l3.a:4:45: error TS2532: Object is possibly 'undefined'.
```

### e5531c8-a-js-build

```sh
/tmp/adamic-e5531c8 js /workspace/adamic/evidence/l3/l3.a
```

Exit 1. Stdout: empty. Complete stderr:

```text
/workspace/adamic/evidence/l3/l3.a:4:45: error TS2532: Object is possibly 'undefined'.
```

## Source Node

`node --experimental-strip-types /workspace/adamic/evidence/l3/l3.ts`

Exit 1. Stdout: empty. Complete stderr:

```text
file:///workspace/adamic/evidence/l3/l3.ts:4
slot.run((...args          ) => console.log(args[0].toUpperCase()), 5);
                                                    ^

TypeError: args[0].toUpperCase is not a function
    at file:///workspace/adamic/evidence/l3/l3.ts:4:53
    at Provider.run (file:///workspace/adamic/evidence/l3/l3.ts:2:109)
    at Object.<anonymous> (file:///workspace/adamic/evidence/l3/l3.ts:4:6)
    at Module._compile (node:internal/modules/cjs/loader:1872:14)
    at Object..js (node:internal/modules/cjs/loader:2003:10)
    at Module.load (node:internal/modules/cjs/loader:1594:32)
    at Module._load (node:internal/modules/cjs/loader:1396:12)
    at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
    at Module.executeUserEntryPoint [as runMain] (node:internal/modules/run_main:154:5)
    at node:internal/main/run_main_module:33:47

Node.js v24.19.0
```

Every command's full stdout, stderr, exact arguments and numeric exit code is retained in separate files; results.json duplicates them in structured form. No compiler source changes were made.
