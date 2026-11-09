# C1/C2 table-cast probes

Compilers rebuilt successfully with `go build -o /tmp/adamic-<short> ./cmd/adamic` at exact revisions main e3aa0bb7642163ebd004a71ba40a4657d27eab71 and compiler/rule-g e5531c8c7fd9fa0b5076bb3d0d0961e6954e9ce1. Each compiler build exits 0 with empty stdout/stderr. Pinned submodules and clean tracked source recorded; detached worktree restored to e5531c8, diff exit 0. No compiler edits.

Go 1.27.1, clang 20.1.8, Node 24.19.0. For each probe, .ts and .a source bytes are identical and match the supplied program. No stock tsc run was requested for this pair.

| Compiler | Probe | Extension | Native --sanitize compile | Native run | JS compile | JS run |
|---|---|---|---|---|---|---|
| e3aa0bb | c1 | .ts | Exit 1: fixed-object-to-record NotYet | Not produced | Exit 1: same refusal | Not produced |
| e3aa0bb | c1 | .a | Exit 1: fixed-object-to-record NotYet | Not produced | Exit 1: same refusal | Not produced |
| e3aa0bb | c2 | .ts | Exit 1: fixed-object-to-record NotYet | Not produced | Exit 1: same refusal | Not produced |
| e3aa0bb | c2 | .a | Exit 1: fixed-object-to-record NotYet | Not produced | Exit 1: same refusal | Not produced |
| e5531c8 | c1 | .ts | Exit 1: fixed-object-to-record NotYet | Not produced | Exit 1: same refusal | Not produced |
| e5531c8 | c1 | .a | Exit 1: fixed-object-to-record NotYet | Not produced | Exit 1: same refusal | Not produced |
| e5531c8 | c2 | .ts | Exit 1: fixed-object-to-record NotYet | Not produced | Exit 1: same refusal | Not produced |
| e5531c8 | c2 | .a | Exit 1: fixed-object-to-record NotYet | Not produced | Exit 1: same refusal | Not produced |

Every compiler/backend combination stops at the same table cast on line 4, before the function can be called. Compilation stdout is empty. No native executable or JavaScript output was produced. No native run occurred and no sanitizer report exists.

## Complete compilation results

### e3aa0bb-c1-ts-native-build

```sh
/tmp/adamic-e3aa0bb build /workspace/adamic/evidence/c1-c2/c1.ts -o /tmp/e3aa0bb-c1-ts --sanitize
```

Exit 1. Stdout empty. Complete stderr:

```text
adamic: /workspace/adamic/evidence/c1-c2/c1.ts:4:15: stage 0 can't lower a { dog: Fn<Dog>; } seen as Record<string, Fn<any>> (fixed objects and records have different storage; copy explicitly) yet
```

### e3aa0bb-c1-ts-js-build

```sh
/tmp/adamic-e3aa0bb js /workspace/adamic/evidence/c1-c2/c1.ts
```

Exit 1. Stdout empty. Complete stderr:

```text
adamic: /workspace/adamic/evidence/c1-c2/c1.ts:4:15: stage 0 can't lower a { dog: Fn<Dog>; } seen as Record<string, Fn<any>> (fixed objects and records have different storage; copy explicitly) yet
```

### e3aa0bb-c1-a-native-build

```sh
/tmp/adamic-e3aa0bb build /workspace/adamic/evidence/c1-c2/c1.a -o /tmp/e3aa0bb-c1-a --sanitize
```

Exit 1. Stdout empty. Complete stderr:

```text
adamic: /workspace/adamic/evidence/c1-c2/c1.a:4:15: stage 0 can't lower a { dog: Fn<Dog>; } seen as Record<string, Fn<any>> (fixed objects and records have different storage; copy explicitly) yet
```

### e3aa0bb-c1-a-js-build

```sh
/tmp/adamic-e3aa0bb js /workspace/adamic/evidence/c1-c2/c1.a
```

Exit 1. Stdout empty. Complete stderr:

```text
adamic: /workspace/adamic/evidence/c1-c2/c1.a:4:15: stage 0 can't lower a { dog: Fn<Dog>; } seen as Record<string, Fn<any>> (fixed objects and records have different storage; copy explicitly) yet
```

### e3aa0bb-c2-ts-native-build

```sh
/tmp/adamic-e3aa0bb build /workspace/adamic/evidence/c1-c2/c2.ts -o /tmp/e3aa0bb-c2-ts --sanitize
```

Exit 1. Stdout empty. Complete stderr:

```text
adamic: /workspace/adamic/evidence/c1-c2/c2.ts:4:15: stage 0 can't lower a { dog: Fn<Dog>; } seen as Record<string, Fn<any>> (fixed objects and records have different storage; copy explicitly) yet
```

### e3aa0bb-c2-ts-js-build

```sh
/tmp/adamic-e3aa0bb js /workspace/adamic/evidence/c1-c2/c2.ts
```

Exit 1. Stdout empty. Complete stderr:

```text
adamic: /workspace/adamic/evidence/c1-c2/c2.ts:4:15: stage 0 can't lower a { dog: Fn<Dog>; } seen as Record<string, Fn<any>> (fixed objects and records have different storage; copy explicitly) yet
```

### e3aa0bb-c2-a-native-build

```sh
/tmp/adamic-e3aa0bb build /workspace/adamic/evidence/c1-c2/c2.a -o /tmp/e3aa0bb-c2-a --sanitize
```

Exit 1. Stdout empty. Complete stderr:

```text
adamic: /workspace/adamic/evidence/c1-c2/c2.a:4:15: stage 0 can't lower a { dog: Fn<Dog>; } seen as Record<string, Fn<any>> (fixed objects and records have different storage; copy explicitly) yet
```

### e3aa0bb-c2-a-js-build

```sh
/tmp/adamic-e3aa0bb js /workspace/adamic/evidence/c1-c2/c2.a
```

Exit 1. Stdout empty. Complete stderr:

```text
adamic: /workspace/adamic/evidence/c1-c2/c2.a:4:15: stage 0 can't lower a { dog: Fn<Dog>; } seen as Record<string, Fn<any>> (fixed objects and records have different storage; copy explicitly) yet
```

### e5531c8-c1-ts-native-build

```sh
/tmp/adamic-e5531c8 build /workspace/adamic/evidence/c1-c2/c1.ts -o /tmp/e5531c8-c1-ts --sanitize
```

Exit 1. Stdout empty. Complete stderr:

```text
adamic: /workspace/adamic/evidence/c1-c2/c1.ts:4:15: stage 0 can't lower a { dog: Fn<Dog>; } seen as Record<string, Fn<any>> (fixed objects and records have different storage; copy explicitly) yet
```

### e5531c8-c1-ts-js-build

```sh
/tmp/adamic-e5531c8 js /workspace/adamic/evidence/c1-c2/c1.ts
```

Exit 1. Stdout empty. Complete stderr:

```text
adamic: /workspace/adamic/evidence/c1-c2/c1.ts:4:15: stage 0 can't lower a { dog: Fn<Dog>; } seen as Record<string, Fn<any>> (fixed objects and records have different storage; copy explicitly) yet
```

### e5531c8-c1-a-native-build

```sh
/tmp/adamic-e5531c8 build /workspace/adamic/evidence/c1-c2/c1.a -o /tmp/e5531c8-c1-a --sanitize
```

Exit 1. Stdout empty. Complete stderr:

```text
adamic: /workspace/adamic/evidence/c1-c2/c1.a:4:15: stage 0 can't lower a { dog: Fn<Dog>; } seen as Record<string, Fn<any>> (fixed objects and records have different storage; copy explicitly) yet
```

### e5531c8-c1-a-js-build

```sh
/tmp/adamic-e5531c8 js /workspace/adamic/evidence/c1-c2/c1.a
```

Exit 1. Stdout empty. Complete stderr:

```text
adamic: /workspace/adamic/evidence/c1-c2/c1.a:4:15: stage 0 can't lower a { dog: Fn<Dog>; } seen as Record<string, Fn<any>> (fixed objects and records have different storage; copy explicitly) yet
```

### e5531c8-c2-ts-native-build

```sh
/tmp/adamic-e5531c8 build /workspace/adamic/evidence/c1-c2/c2.ts -o /tmp/e5531c8-c2-ts --sanitize
```

Exit 1. Stdout empty. Complete stderr:

```text
adamic: /workspace/adamic/evidence/c1-c2/c2.ts:4:15: stage 0 can't lower a { dog: Fn<Dog>; } seen as Record<string, Fn<any>> (fixed objects and records have different storage; copy explicitly) yet
```

### e5531c8-c2-ts-js-build

```sh
/tmp/adamic-e5531c8 js /workspace/adamic/evidence/c1-c2/c2.ts
```

Exit 1. Stdout empty. Complete stderr:

```text
adamic: /workspace/adamic/evidence/c1-c2/c2.ts:4:15: stage 0 can't lower a { dog: Fn<Dog>; } seen as Record<string, Fn<any>> (fixed objects and records have different storage; copy explicitly) yet
```

### e5531c8-c2-a-native-build

```sh
/tmp/adamic-e5531c8 build /workspace/adamic/evidence/c1-c2/c2.a -o /tmp/e5531c8-c2-a --sanitize
```

Exit 1. Stdout empty. Complete stderr:

```text
adamic: /workspace/adamic/evidence/c1-c2/c2.a:4:15: stage 0 can't lower a { dog: Fn<Dog>; } seen as Record<string, Fn<any>> (fixed objects and records have different storage; copy explicitly) yet
```

### e5531c8-c2-a-js-build

```sh
/tmp/adamic-e5531c8 js /workspace/adamic/evidence/c1-c2/c2.a
```

Exit 1. Stdout empty. Complete stderr:

```text
adamic: /workspace/adamic/evidence/c1-c2/c2.a:4:15: stage 0 can't lower a { dog: Fn<Dog>; } seen as Record<string, Fn<any>> (fixed objects and records have different storage; copy explicitly) yet
```

## Source Node, types stripped

### c1

```sh
node --experimental-strip-types /workspace/adamic/evidence/c1-c2/c1.ts
```

Exit 1. Stdout empty. Complete stderr:

```text
file:///workspace/adamic/evidence/c1-c2/c1.ts:3
const table                   = { dog:    (node     , cb                              )                => { console.log(node.bark()); return cb(node); } };
                                                                                                                             ^

TypeError: node.bark is not a function
    at dog (file:///workspace/adamic/evidence/c1-c2/c1.ts:3:126)
    at Object.<anonymous> (file:///workspace/adamic/evidence/c1-c2/c1.ts:6:24)
    at Module._compile (node:internal/modules/cjs/loader:1872:14)
    at Object..js (node:internal/modules/cjs/loader:2003:10)
    at Module.load (node:internal/modules/cjs/loader:1594:32)
    at Module._load (node:internal/modules/cjs/loader:1396:12)
    at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
    at Module.executeUserEntryPoint [as runMain] (node:internal/modules/run_main:154:5)
    at node:internal/main/run_main_module:33:47

Node.js v24.19.0
```

### c2

```sh
node --experimental-strip-types /workspace/adamic/evidence/c1-c2/c2.ts
```

Exit 1. Stdout empty. Complete stderr:

```text
file:///workspace/adamic/evidence/c1-c2/c2.ts:3
const table                   = { dog: (node     )         => node.bark() };
                                                                   ^

TypeError: node.bark is not a function
    at dog (file:///workspace/adamic/evidence/c1-c2/c2.ts:3:68)
    at Object.<anonymous> (file:///workspace/adamic/evidence/c1-c2/c2.ts:6:36)
    at Module._compile (node:internal/modules/cjs/loader:1872:14)
    at Object..js (node:internal/modules/cjs/loader:2003:10)
    at Module.load (node:internal/modules/cjs/loader:1594:32)
    at Module._load (node:internal/modules/cjs/loader:1396:12)
    at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
    at Module.executeUserEntryPoint [as runMain] (node:internal/modules/run_main:154:5)
    at node:internal/main/run_main_module:33:47

Node.js v24.19.0
```

All exact commands, stdout, stderr and numeric exit codes are retained separately and in results.json. Driver: run.py.
