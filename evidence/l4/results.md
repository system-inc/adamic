# L4 comparison

Exact revisions: main e3aa0bb7642163ebd004a71ba40a4657d27eab71 and compiler/rule-g e5531c8c7fd9fa0b5076bb3d0d0961e6954e9ce1. Both compilers were rebuilt successfully (exit 0) with `go build -o /tmp/adamic-<short> ./cmd/adamic` in the detached worktree at each exact revision. Sources and pinned submodules verified; worktree restored to e5531c8 afterward, clean diff exit 0.

Go 1.27.1, clang 20.1.8, Node 24.19.0. l4.ts and l4.a have identical bytes matching the supplied four lines. No compiler source changes.

| Compiler | File | Native --sanitize compile | Native run | JS compile | JS run |
|---|---|---|---|---|---|
| e3aa0bb | l4.ts | Exit 1: generic-method NotYet | Not produced | Exit 1: generic-method NotYet | Not produced |
| e3aa0bb | l4.a | Exit 1: generic-method NotYet | Not produced | Exit 1: generic-method NotYet | Not produced |
| e5531c8 | l4.ts | Exit 1: constraint holding any NotYet | Not produced | Exit 1: constraint holding any NotYet | Not produced |
| e5531c8 | l4.a | Exit 1: constraint holding any NotYet | Not produced | Exit 1: constraint holding any NotYet | Not produced |

All native/JS compilation stdout is empty. Base stops at the generic-method restriction. Head stops at condition 3: the type parameter falls back to a constraint holding any. No native executable or JavaScript program was produced, so neither backend could be run. No sanitizer report exists.

## Complete compilation results

### e3aa0bb-ts-native-build

```sh
/tmp/adamic-e3aa0bb build /workspace/adamic/evidence/l4/l4.ts -o /tmp/l4-e3aa0bb-ts --sanitize
```

Exit 1. Stdout empty. Complete stderr:

```text
adamic: /workspace/adamic/evidence/l4/l4.ts:3:20: stage 0 can't lower a generic method run of type <T extends any[]>(callback: (...args: NoInfer<T>) => void, ...rest: NoInfer<T>) => void seen as (callback: (...args: string[]) => void, ...rest: unknown[]) => void yet
```

### e3aa0bb-ts-js-build

```sh
/tmp/adamic-e3aa0bb js /workspace/adamic/evidence/l4/l4.ts
```

Exit 1. Stdout empty. Complete stderr:

```text
adamic: /workspace/adamic/evidence/l4/l4.ts:3:20: stage 0 can't lower a generic method run of type <T extends any[]>(callback: (...args: NoInfer<T>) => void, ...rest: NoInfer<T>) => void seen as (callback: (...args: string[]) => void, ...rest: unknown[]) => void yet
```

### e3aa0bb-a-native-build

```sh
/tmp/adamic-e3aa0bb build /workspace/adamic/evidence/l4/l4.a -o /tmp/l4-e3aa0bb-a --sanitize
```

Exit 1. Stdout empty. Complete stderr:

```text
adamic: /workspace/adamic/evidence/l4/l4.a:3:20: stage 0 can't lower a generic method run of type <T extends any[]>(callback: (...args: NoInfer<T>) => void, ...rest: NoInfer<T>) => void seen as (callback: (...args: string[]) => void, ...rest: unknown[]) => void yet
```

### e3aa0bb-a-js-build

```sh
/tmp/adamic-e3aa0bb js /workspace/adamic/evidence/l4/l4.a
```

Exit 1. Stdout empty. Complete stderr:

```text
adamic: /workspace/adamic/evidence/l4/l4.a:3:20: stage 0 can't lower a generic method run of type <T extends any[]>(callback: (...args: NoInfer<T>) => void, ...rest: NoInfer<T>) => void seen as (callback: (...args: string[]) => void, ...rest: unknown[]) => void yet
```

### e5531c8-ts-native-build

```sh
/tmp/adamic-e5531c8 build /workspace/adamic/evidence/l4/l4.ts -o /tmp/l4-e5531c8-ts --sanitize
```

Exit 1. Stdout empty. Complete stderr:

```text
adamic: /workspace/adamic/evidence/l4/l4.ts:3:20: stage 0 can't lower a generic function of type <T extends any[]>(callback: (...args: NoInfer<T>) => void, ...rest: NoInfer<T>) => void seen as (callback: (...args: string[]) => void, ...rest: unknown[]) => void, which doesn't fix every type parameter it's instantiated with, so one falls back to a constraint holding any yet
```

### e5531c8-ts-js-build

```sh
/tmp/adamic-e5531c8 js /workspace/adamic/evidence/l4/l4.ts
```

Exit 1. Stdout empty. Complete stderr:

```text
adamic: /workspace/adamic/evidence/l4/l4.ts:3:20: stage 0 can't lower a generic function of type <T extends any[]>(callback: (...args: NoInfer<T>) => void, ...rest: NoInfer<T>) => void seen as (callback: (...args: string[]) => void, ...rest: unknown[]) => void, which doesn't fix every type parameter it's instantiated with, so one falls back to a constraint holding any yet
```

### e5531c8-a-native-build

```sh
/tmp/adamic-e5531c8 build /workspace/adamic/evidence/l4/l4.a -o /tmp/l4-e5531c8-a --sanitize
```

Exit 1. Stdout empty. Complete stderr:

```text
adamic: /workspace/adamic/evidence/l4/l4.a:3:20: stage 0 can't lower a generic function of type <T extends any[]>(callback: (...args: NoInfer<T>) => void, ...rest: NoInfer<T>) => void seen as (callback: (...args: string[]) => void, ...rest: unknown[]) => void, which doesn't fix every type parameter it's instantiated with, so one falls back to a constraint holding any yet
```

### e5531c8-a-js-build

```sh
/tmp/adamic-e5531c8 js /workspace/adamic/evidence/l4/l4.a
```

Exit 1. Stdout empty. Complete stderr:

```text
adamic: /workspace/adamic/evidence/l4/l4.a:3:20: stage 0 can't lower a generic function of type <T extends any[]>(callback: (...args: NoInfer<T>) => void, ...rest: NoInfer<T>) => void seen as (callback: (...args: string[]) => void, ...rest: unknown[]) => void, which doesn't fix every type parameter it's instantiated with, so one falls back to a constraint holding any yet
```

## Source Node

`node --experimental-strip-types /workspace/adamic/evidence/l4/l4.ts`

Exit 1; stdout empty. Complete stderr:

```text
file:///workspace/adamic/evidence/l4/l4.ts:4
slot.run((...args          ) => { for (const a of args) { console.log(a.toUpperCase()); } }, 5);
                                                                        ^

TypeError: a.toUpperCase is not a function
    at file:///workspace/adamic/evidence/l4/l4.ts:4:73
    at Provider.run (file:///workspace/adamic/evidence/l4/l4.ts:2:109)
    at Object.<anonymous> (file:///workspace/adamic/evidence/l4/l4.ts:4:6)
    at Module._compile (node:internal/modules/cjs/loader:1872:14)
    at Object..js (node:internal/modules/cjs/loader:2003:10)
    at Module.load (node:internal/modules/cjs/loader:1594:32)
    at Module._load (node:internal/modules/cjs/loader:1396:12)
    at wrapModuleLoad (node:internal/modules/cjs/loader:255:19)
    at Module.executeUserEntryPoint [as runMain] (node:internal/modules/run_main:154:5)
    at node:internal/main/run_main_module:33:47

Node.js v24.19.0
```

## Stock tsc 6.0.3

Stock Microsoft TypeScript checkout at 050880ce59e30b356b686bd3144efe24f875ebc8, unmodified tracked files (clean diff exit 0). `node /workspace/typescript-source/lib/tsc.js --version` reports Version 6.0.3.

Both commands exit 0 with no stdout or stderr:

```sh
node /workspace/typescript-source/lib/tsc.js --ignoreConfig --noEmit /workspace/adamic/evidence/l4/l4.ts
node /workspace/typescript-source/lib/tsc.js --ignoreConfig --noEmit --strict /workspace/adamic/evidence/l4/l4.ts
```

The --ignoreConfig flag isolates the exact supplied program from repository tsconfig settings. All exact command lines, full stdout, full stderr and numeric exit codes are retained alongside results.json. The runnable driver is run.py.
