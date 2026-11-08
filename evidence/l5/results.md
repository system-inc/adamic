# L5 comparison

Exact compilers: main e3aa0bb7642163ebd004a71ba40a4657d27eab71 and compiler/rule-g e5531c8c7fd9fa0b5076bb3d0d0961e6954e9ce1. Both rebuilt successfully (exit 0) using `go build -o /tmp/adamic-<short> ./cmd/adamic` at each exact revision. Pinned submodules and clean source recorded. Detached test worktree restored to e5531c8, diff exit 0. No compiler changes.

Go 1.27.1; clang 20.1.8; Node 24.19.0. l5.ts and l5.a have identical bytes matching the supplied program.

| Compiler | File | Native --sanitize compile | Native run | JS compile | JS run |
|---|---|---|---|---|---|
| e3aa0bb | l5.ts | Exit 1: adamic/no-void-read | Not produced | Exit 1: same refusal | Not produced |
| e3aa0bb | l5.a | Exit 1: adamic/no-void-read | Not produced | Exit 1: same refusal | Not produced |
| e5531c8 | l5.ts | Exit 1: adamic/no-void-read | Not produced | Exit 1: same refusal | Not produced |
| e5531c8 | l5.a | Exit 1: adamic/no-void-read | Not produced | Exit 1: same refusal | Not produced |

Both compilers stop at the same no-void-read refusal. Compilation stdout is empty in every case. No native executable or emitted JavaScript was produced, so no backend execution could occur and no sanitizer report exists.

## Complete compiler refusals

### e3aa0bb-ts-native-build

```sh
/tmp/adamic-e3aa0bb build /workspace/adamic/evidence/l5/l5.ts -o /tmp/l5-e3aa0bb-ts --sanitize
```

Exit 1. Stdout empty. Complete stderr:

```text
adamic: /workspace/adamic/evidence/l5/l5.ts:3:27: Adamic 0.1 refuses a generic instantiated at void whose body reads the value: provide's T = void (read from the signature), read at /workspace/adamic/evidence/l5/l5.ts:2:117 as a spread; give the function a return type, or don't hand a void container to a generic that reads its elements (adamic/no-void-read)
```

### e3aa0bb-ts-js-build

```sh
/tmp/adamic-e3aa0bb js /workspace/adamic/evidence/l5/l5.ts
```

Exit 1. Stdout empty. Complete stderr:

```text
adamic: /workspace/adamic/evidence/l5/l5.ts:3:27: Adamic 0.1 refuses a generic instantiated at void whose body reads the value: provide's T = void (read from the signature), read at /workspace/adamic/evidence/l5/l5.ts:2:117 as a spread; give the function a return type, or don't hand a void container to a generic that reads its elements (adamic/no-void-read)
```

### e3aa0bb-a-native-build

```sh
/tmp/adamic-e3aa0bb build /workspace/adamic/evidence/l5/l5.a -o /tmp/l5-e3aa0bb-a --sanitize
```

Exit 1. Stdout empty. Complete stderr:

```text
adamic: /workspace/adamic/evidence/l5/l5.a:3:27: Adamic 0.1 refuses a generic instantiated at void whose body reads the value: provide's T = void (read from the signature), read at /workspace/adamic/evidence/l5/l5.a:2:117 as a spread; give the function a return type, or don't hand a void container to a generic that reads its elements (adamic/no-void-read)
```

### e3aa0bb-a-js-build

```sh
/tmp/adamic-e3aa0bb js /workspace/adamic/evidence/l5/l5.a
```

Exit 1. Stdout empty. Complete stderr:

```text
adamic: /workspace/adamic/evidence/l5/l5.a:3:27: Adamic 0.1 refuses a generic instantiated at void whose body reads the value: provide's T = void (read from the signature), read at /workspace/adamic/evidence/l5/l5.a:2:117 as a spread; give the function a return type, or don't hand a void container to a generic that reads its elements (adamic/no-void-read)
```

### e5531c8-ts-native-build

```sh
/tmp/adamic-e5531c8 build /workspace/adamic/evidence/l5/l5.ts -o /tmp/l5-e5531c8-ts --sanitize
```

Exit 1. Stdout empty. Complete stderr:

```text
adamic: /workspace/adamic/evidence/l5/l5.ts:3:27: Adamic 0.1 refuses a generic instantiated at void whose body reads the value: provide's T = void (read from the signature), read at /workspace/adamic/evidence/l5/l5.ts:2:117 as a spread; give the function a return type, or don't hand a void container to a generic that reads its elements (adamic/no-void-read)
```

### e5531c8-ts-js-build

```sh
/tmp/adamic-e5531c8 js /workspace/adamic/evidence/l5/l5.ts
```

Exit 1. Stdout empty. Complete stderr:

```text
adamic: /workspace/adamic/evidence/l5/l5.ts:3:27: Adamic 0.1 refuses a generic instantiated at void whose body reads the value: provide's T = void (read from the signature), read at /workspace/adamic/evidence/l5/l5.ts:2:117 as a spread; give the function a return type, or don't hand a void container to a generic that reads its elements (adamic/no-void-read)
```

### e5531c8-a-native-build

```sh
/tmp/adamic-e5531c8 build /workspace/adamic/evidence/l5/l5.a -o /tmp/l5-e5531c8-a --sanitize
```

Exit 1. Stdout empty. Complete stderr:

```text
adamic: /workspace/adamic/evidence/l5/l5.a:3:27: Adamic 0.1 refuses a generic instantiated at void whose body reads the value: provide's T = void (read from the signature), read at /workspace/adamic/evidence/l5/l5.a:2:117 as a spread; give the function a return type, or don't hand a void container to a generic that reads its elements (adamic/no-void-read)
```

### e5531c8-a-js-build

```sh
/tmp/adamic-e5531c8 js /workspace/adamic/evidence/l5/l5.a
```

Exit 1. Stdout empty. Complete stderr:

```text
adamic: /workspace/adamic/evidence/l5/l5.a:3:27: Adamic 0.1 refuses a generic instantiated at void whose body reads the value: provide's T = void (read from the signature), read at /workspace/adamic/evidence/l5/l5.a:2:117 as a spread; give the function a return type, or don't hand a void container to a generic that reads its elements (adamic/no-void-read)
```

## Stock tsc 6.0.3

Stock Microsoft TypeScript checkout 050880ce59e30b356b686bd3144efe24f875ebc8, unmodified tracked files. Compiler reports Version 6.0.3. Both commands exit 0 with empty stdout and stderr:

```sh
node /workspace/typescript-source/lib/tsc.js --ignoreConfig --noEmit /workspace/adamic/evidence/l5/l5.ts
node /workspace/typescript-source/lib/tsc.js --ignoreConfig --noEmit --strict /workspace/adamic/evidence/l5/l5.ts
```

--ignoreConfig isolates the program from repository configuration.

## Source Node

`node --experimental-strip-types /workspace/adamic/evidence/l5/l5.ts`

Exit 1. Stdout empty. Complete stderr:

```text
file:///workspace/adamic/evidence/l5/l5.ts:4
slot.run((...args          ) => { for (const a of args) { console.log(a.toUpperCase()); } }, 5);
                                                                        ^

TypeError: a.toUpperCase is not a function
    at file:///workspace/adamic/evidence/l5/l5.ts:4:73
    at Object.provide [as run] (file:///workspace/adamic/evidence/l5/l5.ts:2:105)
    at Object.<anonymous> (file:///workspace/adamic/evidence/l5/l5.ts:4:6)
    at Module._compile (node:internal/modules/cjs/loader:1872:14)
    at Object..js (node:internal/modules/cjs/loader:2003:10)
    at Module.load (node:internal/modules/cjs/loader:1594:32)
    at Module._load (node:internal/modules/cjs/loader:1396:12)
    at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
    at Module.executeUserEntryPoint [as runMain] (node:internal/modules/run_main:154:5)
    at node:internal/main/run_main_module:33:47

Node.js v24.19.0
```

All commands, stdout, stderr and numeric exit codes are retained separately and in results.json. The executable driver is run.py.
