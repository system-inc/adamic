# Weak-interface primitive collection agreement probe

All three programs pass TypeScript 6.0.3 and run successfully on Node. Adamic refuses all three during lowering with `adamic/no-optional-widening`, before native, sanitized native, or JavaScript execution. **None of the three backends agrees with Node for any of these programs: each is stopped, rather than producing a differing runtime answer.**

## Revision and setup

- Adamic base: `16429a30452d027809f0ea85bc395479423e493e` (requested main HEAD). Ran `git fetch origin main`, checked out that commit detached, and ran `git submodule update --init --recursive` successfully.
- Recursive submodules: cohere `5360753e19e4417675a90e92a1823bea66b13918`; cohere/TypeScript `d92d9bfee114c80be2c375d72edae966176e3a4f`.
- External upstream checkout: `/tmp/lower-weak-probe/TypeScript`, verified HEAD `050880ce59e30b356b686bd3144efe24f875ebc8`; exported `ADAMIC_TYPESCRIPT_SOURCE` to that directory.
- Copied stage3/api's package.json and package-lock.json outside the repository and ran `npm ci --prefix /tmp/lower-weak-probe/api`. Pinned packages include TypeScript `6.0.3` and `@types/node` `25.3.3`. Exported `NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt` and `npm_config_cafile=/etc/ssl/certs/ca-certificates.crt` for npm.
- Put `/tmp/lower-weak-probe/api/node_modules/.bin` on PATH. Linked `/tmp/lower-weak-probe/node_modules` to that external installation so the pinned Node types are discoverable by tsc for the external inputs.
- Sourced `/workspace/adamic-tools/env.sh`: Go `go1.27.1 linux/amd64`, Clang `20.1.8`; Node `v24.19.0`. Created `/tmp/adamic-gate` outside the repository for TMPDIR.
- Inputs and scratch driver reside under `/tmp/lower-weak-probe`. No repository source or test file was written or edited. The only repository file added for this work is this report.

## TypeScript and Node

From `/tmp/lower-weak-probe`, ran the exact check for each input:

```sh
tsc --strict --noEmit w1.ts
tsc --strict --noEmit w2.ts
tsc --strict --noEmit w3.ts
```

All three exit **0**, with empty stdout and stderr and no diagnostics. `tsc --version` reports `Version 6.0.3`. These checks use the pinned Node declaration packages described above. A separate check with `--lib es2020,dom` also passed each input.

From the repository root, ran:

```sh
node --disable-warning=ExperimentalWarning oracle/node.mjs /tmp/lower-weak-probe/w1.ts
node --disable-warning=ExperimentalWarning oracle/node.mjs /tmp/lower-weak-probe/w2.ts
node --disable-warning=ExperimentalWarning oracle/node.mjs /tmp/lower-weak-probe/w3.ts
```

| Program | Node exit | Node stdout (escaped, including final newline) | Node stderr |
| --- | --- | --- | --- |
| w1.ts | 0 | `"x\n1\n"` | empty |
| w2.ts | 0 | `"1\n"` | empty |
| w3.ts | 0 | `"x\n"` | empty |

## Adamic backend results

Read `internal/oracle/static_self_construction_test.go` and `internal/oracle/loom_views_d_test.go`. The former enters through `loomViewsDAgrees`; that delegates to `loomViewsDAgreesBackend`. This helper runs Node, calls `lowered`, and only after successful lowering selects `releasedUncached`, `nativelyUncached`, or `onJavaScriptBackend`.

Used a throwaway `_test.go` under `/tmp` with Go's `-overlay`, exposing it as a virtual test file in package oracle without writing it into the repository. The observation driver called `onNode` and `lowered`, recorded exact stops, and was prepared to inspect IR and invoke those same backend helpers for successful lowerings. Its test process exited 0 because it records refusals as observations; that is not a backend success.

Then directly drove the named agreement helper for every program/backend combination:

```go
func TestLowerWeakAgreement(t *testing.T) {
    for _, name := range []string{"w1", "w2", "w3"} {
        t.Run(name, func(t *testing.T) {
            for _, backend := range []string{"native", "sanitized", "javascript"} {
                t.Run(backend, func(t *testing.T) {
                    loomViewsDAgreesBackend(t,
                        "../../tmp/lower-weak-probe/"+name+".ts", backend)
                })
            }
        })
    }
}
```

Ran from the repository root with `ADAMIC_GATE_UNCACHED=1`, the exported source and tool environment above, and an overlay mapping `/workspace/adamic/internal/oracle/lower_weak_probe_test.go` to `/tmp/lower-weak-probe/probe_test.go`:

```sh
go test -overlay /tmp/lower-weak-probe/overlay.json ./internal/oracle \
  -run '^TestLowerWeakAgreement$' -count=1 -v
```

The agreement test process exited **1**: all nine subtests failed at the common lowering gate. The cache summary recorded nine Node misses and zero native hits/misses. The selected sanitized path uses the standard `native.Options{Sanitize: true}` helper (ASan and UBSan, `-fsanitize=address,undefined`, with `ASAN_OPTIONS=detect_leaks=0` on Linux), but no input reached its build or execution.

| Program | Native | ASan + UBSan native | JavaScript backend |
| --- | --- | --- | --- |
| w1.ts | lowering stop at 4:9; does not agree with Node | same stop; does not agree with Node | same stop; does not agree with Node |
| w2.ts | lowering stop at 5:7; does not agree with Node | same stop; does not agree with Node | same stop; does not agree with Node |
| w3.ts | lowering stop at 5:7; does not agree with Node | same stop; does not agree with Node | same stop; does not agree with Node |

For every cell above, **backend process exit: not applicable; backend stdout/stderr: not produced**, because no backend program was generated or run. The Go test exit 1 is the harness exit, not an Adamic program exit. No sanitizer runtime result can be claimed.

### Exact stop text

w1.ts, at `return wide`:

```text
/tmp/lower-weak-probe/w1.ts:4:9: Adamic 0.1 refuses optional property a in Weak absent from structural source {}, which can hide fields; declare a on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

w2.ts, at the first `s.add(one)`:

```text
/tmp/lower-weak-probe/w2.ts:5:7: Adamic 0.1 refuses optional property a in Weak absent from structural source {}, which can hide fields; declare a on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

w3.ts, at `m.set(text, 'x')`:

```text
/tmp/lower-weak-probe/w3.ts:5:7: Adamic 0.1 refuses optional property a in Weak absent from structural source {}, which can hide fields; declare a on the source type, or build a fresh object with known fields (adamic/no-optional-widening)
```

Each exact diagnostic was repeated unchanged for that program's native, sanitized, and JavaScript helper invocations.

## IR types

No program lowers successfully, so there is no completed `ir.Program` from which to report a Map key or Set element `ir.Type`, or the `ir.Type` of `wide`, `one`, `again`, `text`, or `built`. Assigning representation types based on their source annotations would be inference, not an observed lowered IR result. In particular, the source interface name `Weak` does not establish the unrelated runtime representation `ir.Weak`.

## Exact inputs

### w1.ts

```ts
interface Weak { a?: number }
function key(n: number): Weak {
	const wide: {} = n;
	return wide;
}
const m = new Map<Weak, string>();
m.set(key(1), 'x');
console.log(String(m.get(key(1))));
console.log(String(m.size));
```

### w2.ts

```ts
interface Weak { a?: number }
const s = new Set<Weak>();
const one: {} = 1;
const again: {} = 2 - 1;
s.add(one);
s.add(again);
console.log(String(s.size));
```

### w3.ts

```ts
interface Weak { a?: number }
const m = new Map<Weak, string>();
const text: {} = 'k';
const built: {} = ['k'].join('');
m.set(text, 'x');
console.log(String(m.get(built)));
```

This report is the sole file committed on `proof/lower-weak-probe-report`, created from the requested detached HEAD. Only that branch is the publication target.
