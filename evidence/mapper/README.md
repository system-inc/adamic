# Mapper callback probe

Exact source copied to `/tmp/mapper.ts` and `/tmp/mapper.a` (also retained here). No source edits or diagnostic suppressions.

Stock TypeScript is the repository-pinned npm TypeScript 6.0.3 from stage3/api/package-lock.json, installed with npm ci --ignore-scripts and the system CA bundle. The oracle's assignment-proofs/check-types.cjs uses that same stock package and asserts version 6.0.3. The cohere/TypeScript package's tsc is the native-port CLI and was deliberately not used as stock tsc.

Commands were run from /tmp to avoid the repository's tsconfig. Initial attempts from the repository hit TS5112 and produced no JavaScript; those setup logs are retained as *-initial.* and are not the probe result.

```sh
node /tmp/adamic-b3/stage3/api/node_modules/typescript/bin/tsc --strict --noEmit /tmp/mapper.ts
node /tmp/adamic-b3/stage3/api/node_modules/typescript/bin/tsc --strict --target es2024 --outDir /tmp/mapper-emitted /tmp/mapper.ts
node /tmp/mapper-emitted/mapper.js
```

Emission exits 2 with the same diagnostic, but default noEmitOnError=false still emits mapper.js. Node runs that exact emitted JavaScript.

Both Adamic binaries were freshly built from their exact worktrees with go build, exits 0. Each probe build uses `adamic build /tmp/mapper.<ext> -o /tmp/mapper-<revision>-<ext> --sanitize`. Exact command arrays and outputs are retained separately, with structured results.json.

| Tool/revision | Extension | Compile/check exit and result | Run output and exit |
|---|---|---|---|
| Stock tsc 6.0.3, --strict --noEmit | .ts | 2: TS2322 at 6:7; full diagnostic below | Not run from noEmit |
| Node 24.19.0, tsc-emitted JS | .js | Emit exit 2; emitted despite TS2322 | stdout empty; TypeError: d.bark is not a function; exit 1 |
| main ff41896c | .a | 1: TS2322 at 6:7 and TS2345 at 7:13 | Not run; no binary |
| main ff41896c | .ts | 1: TS2322 at 6:7 and TS2345 at 7:13 | Not run; no binary |
| c9fa1ea7 | .a | 1: TS2322 at 6:7 and TS2345 at 7:13 | Not run; no binary |
| c9fa1ea7 | .ts | 1: TS2322 at 6:7 and TS2345 at 7:13 | Not run; no binary |

## Exact stock tsc diagnostics

```text
mapper.ts(6,7): error TS2322: Type 'Animal[]' is not assignable to type 'Mapper'.
  Types of property 'map' are incompatible.
    Type '<U>(callbackfn: (value: Animal, index: number, array: Animal[]) => U, thisArg?: any) => U[]' is not assignable to type '<U>(f: (x: Dog) => U) => U[]'.
      Types of parameters 'callbackfn' and 'f' are incompatible.
        Types of parameters 'x' and 'value' are incompatible.
          Property 'bark' is missing in type 'Animal' but required in type 'Dog'.
```

Adamic's additional TS2345 is `Argument of type 'string[]' is not assignable to parameter of type 'string'.` at the console.log call. The full TS2322 chains and paths are retained in each *.compile.stderr. This source is rejected by stock strict TypeScript and both Adamic revisions before native lowering.

Node stdout is empty. Full Node stderr and stack are retained in node.stderr. No native probe runs were attempted because all four sanitized compilations were refused.
