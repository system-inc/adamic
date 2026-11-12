# Live reads with checker flow type `never`

Neither program is a silent wrong answer at the tested HEAD. Adamic refuses both `.ts` and `.a` forms before backend execution. The stop is on the condition's read narrowed to `undefined`, before the true-arm read whose checker flow type is `never`.

## Revision and setup

Run date: 2026-11-12 (UTC). Ran `git fetch origin main`, checked out `origin/main` detached, and ran `git submodule update --init --recursive`.

- Tested HEAD: `219dbb6e7e20ea99d3e5d4b61c0cb14102dc9205`.
- Cohere: `5360753e19e4417675a90e92a1823bea66b13918`.
- Cohere's nested TypeScript: `d92d9bfee114c80be2c375d72edae966176e3a4f`.
- Separate upstream TypeScript checkout: `/tmp/lower-never-live/typescript`, detached at `050880ce59e30b356b686bd3144efe24f875ebc8`, clean. Exported `ADAMIC_TYPESCRIPT_SOURCE` to this path for the probe.
- Loaded `/workspace/adamic-tools/env.sh` first. It sets `TMPDIR=/tmp/adamic-gate`; created that directory outside the repository.
- Copied `stage3/api/package.json` and `package-lock.json` to `/tmp/lower-never-live/api` and ran `npm ci --prefix /tmp/lower-never-live/api`. Verified the copied lock matches tested HEAD. Prepended its `node_modules/.bin` to PATH; `tsc --version` is `Version 6.0.3`.
- Set `NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt` and `npm_config_cafile=/etc/ssl/certs/ca-certificates.crt` for installation and the agreement probe.
- Go: `go1.27.1 linux/amd64`; Node: `v24.19.0`; Clang: `20.1.8`.

All inputs, scripts, overlay, logs, and generated test artifacts were outside the repository. Only this report was added as repository content.

## Programs and checker

Each source below was written unchanged to `/tmp/lower-never-live/<name>.ts` and `<name>.a`; `cmp` confirmed identical text within each pair. No source adjustment was needed.

### n1

```ts
interface Sym { lastAssignmentPos?: number }
function mark(s: Sym): void { s.lastAssignmentPos = 5; }
function f(s: Sym): number {
    if (s.lastAssignmentPos !== undefined) { return 1; }
    mark(s);
    return s.lastAssignmentPos !== undefined ? s.lastAssignmentPos : -1;
}
console.log(String(f({})));
```

### n2

```ts
let found: [number, number] | undefined = undefined;
function visit(): void { found = [1, 2]; }
function run(): string {
    found = undefined;
    visit();
    return found ? String(found[0]) : 'none';
}
console.log(run());
```

From `/tmp/lower-never-live`, ran separately for each `.ts`:

```sh
tsc --strict --noUncheckedIndexedAccess --noEmit /tmp/lower-never-live/n1.ts
tsc --strict --noUncheckedIndexedAccess --noEmit /tmp/lower-never-live/n2.ts
```

Both exited 0 with no diagnostics. An initial invocation from the repository directory exited 1 with the following configuration diagnostic for each file; rerunning from scratch resolved it without changing source or compiler flags:

```text
error TS5112: tsconfig.json is present but will not be loaded if files are specified on commandline. Use '--ignoreConfig' to skip this error.
```

The TypeScript 6.0.3 API's `getTypeAtLocation` reports:

| Program | True-arm read | Location | Flow type |
| --- | --- | --- | --- |
| n1 | `s.lastAssignmentPos` | 6:48 | `never` |
| n2 | `found` (receiver of `[0]`) | 6:27 | `never` |

Checker inspection script, run with `node /tmp/lower-never-live/flow.cjs`:

```js
const ts = require('/tmp/lower-never-live/api/node_modules/typescript');
for (const name of ['n1','n2']) {
 const path = `/tmp/lower-never-live/${name}.ts`;
 const p = ts.createProgram([path], {strict:true,noUncheckedIndexedAccess:true,noEmit:true});
 const checker=p.getTypeChecker(), sf=p.getSourceFile(path);
 function walk(n) {
  if(ts.isConditionalExpression(n)) {
   const read = name==='n1' ? n.whenTrue : n.whenTrue.arguments[0].expression;
   const pos=sf.getLineAndCharacterOfPosition(read.getStart(sf));
   console.log(`${name}: ${read.getText(sf)} at ${pos.line+1}:${pos.character+1}: ${checker.typeToString(checker.getTypeAtLocation(read))}`);
  }
  ts.forEachChild(n,walk);
 }
 walk(sf);
}
```

## Node and Adamic results

Node was run as `node --disable-warning=ExperimentalWarning oracle/node.mjs <absolute-file>` from the repository, and also through `onNode` for all four files in the overlay probe. Its results were:

| Input | Node stdout (quoted) | Node stderr | Node exit | Adamic lowering | Release / sanitized / JavaScript |
| --- | --- | --- | --- | --- | --- |
| n1.ts | `"5\n"` | `""` | 0 | refused | not run: no lowered program |
| n1.a | `"5\n"` | `""` | 0 | refused | not run: no lowered program |
| n2.ts | `"1\n"` | `""` | 0 | refused | not run: no lowered program |
| n2.a | `"1\n"` | `""` | 0 | refused | not run: no lowered program |

Exact stop text:

```text
/tmp/lower-never-live/n1.ts:6:12: stage 0 can't lower a read narrowed to undefined of a property a call may write yet
/tmp/lower-never-live/n1.a:6:12: stage 0 can't lower a read narrowed to undefined of a property a call may write yet
/tmp/lower-never-live/n2.ts:6:12: stage 0 can't lower a read narrowed to undefined of a binding a call may write yet
/tmp/lower-never-live/n2.a:6:12: stage 0 can't lower a read narrowed to undefined of a binding a call may write yet
```

These are lowering refusals, not runtime stderr or process exits of an emitted program. There are no release, sanitized, or JavaScript stdout/stderr/exit observations because none of the four inputs lowers.

The refusals come from `internal/lower/narrowed_nullish.go`: the registered `narrowedNullishReadGap` rule checks binding reads and delegates property reads to `narrowedNullishPropertyReadGap`. It prevents consuming the stale nullish narrowing after a call may write the place. Thus these probes do not reach the later `never` arm's value lowering. The result establishes refusal for these exact programs, not general correctness for every live read typed `never`.

## Agreement helper reproduction

Read `internal/oracle/loom_views_d_test.go` and followed `loomViewsDAgreesBackend`: obtain Node's result with `onNode`, call `lowered`, and, if successful, run `releasedUncached`, `nativelyUncached`, and `onJavaScriptBackend`. The throwaway test records refusals and differences rather than treating a refusal as a test failure, so its PASS means the observations completed; it does not mean the programs lowered.

The overlay mapped the nonexistent repository path `/workspace/adamic/internal/oracle/lower_never_live_probe_test.go` to `/tmp/lower-never-live/probe_test.go`. No test file was written in the repository.

```go
package oracle
import ("testing"; "fmt")
func TestLowerNeverLiveProbe(t *testing.T) {
 for _, name := range []string{"n1", "n2"} {
  for _, ext := range []string{"ts", "a"} {
   t.Run(name+"."+ext, func(t *testing.T) {
    path := "/tmp/lower-never-live/"+name+"."+ext
    want := onNode(t,path)
    fmt.Printf("PROBE %s node stdout=%q stderr=%q exit=%d\n",path,want.stdout,want.stderr,want.exitCode)
    p,err:=lowered(t,path)
    if err!=nil { fmt.Printf("PROBE %s STOP %s\n",path,err); return }
    fmt.Printf("PROBE %s LOWERS\n",path)
    for _, backend := range []string{"release","sanitized","javascript"} {
     var got run
     switch backend {case "release":got=releasedUncached(t,p);case "sanitized":got,_=nativelyUncached(t,p);case "javascript":got=onJavaScriptBackend(t,p)}
     fmt.Printf("PROBE %s %s stdout=%q stderr=%q exit=%d diff=%q\n",path,backend,got.stdout,got.stderr,got.exitCode,disagreement(want,got))
    }
   })
  }
 }
}
```

Command (after sourcing env.sh and exporting the paths and CA variables above):

```sh
export ADAMIC_GATE_UNCACHED=1
go test -overlay /tmp/lower-never-live/overlay.json ./internal/oracle \
  -run '^TestLowerNeverLiveProbe$' -count=1 -v -timeout=20m
```

An earlier attempt while recursive submodule setup was still running could not load `cohere/TypeScript/tsc/go.mod`; the completed run below followed successful submodule initialization.

Complete successful probe output:

```text
=== RUN   TestLowerNeverLiveProbe
=== RUN   TestLowerNeverLiveProbe/n1.ts
PROBE /tmp/lower-never-live/n1.ts node stdout="5\n" stderr="" exit=0
PROBE /tmp/lower-never-live/n1.ts STOP /tmp/lower-never-live/n1.ts:6:12: stage 0 can't lower a read narrowed to undefined of a property a call may write yet
=== RUN   TestLowerNeverLiveProbe/n1.a
PROBE /tmp/lower-never-live/n1.a node stdout="5\n" stderr="" exit=0
PROBE /tmp/lower-never-live/n1.a STOP /tmp/lower-never-live/n1.a:6:12: stage 0 can't lower a read narrowed to undefined of a property a call may write yet
=== RUN   TestLowerNeverLiveProbe/n2.ts
PROBE /tmp/lower-never-live/n2.ts node stdout="1\n" stderr="" exit=0
PROBE /tmp/lower-never-live/n2.ts STOP /tmp/lower-never-live/n2.ts:6:12: stage 0 can't lower a read narrowed to undefined of a binding a call may write yet
=== RUN   TestLowerNeverLiveProbe/n2.a
PROBE /tmp/lower-never-live/n2.a node stdout="1\n" stderr="" exit=0
PROBE /tmp/lower-never-live/n2.a STOP /tmp/lower-never-live/n2.a:6:12: stage 0 can't lower a read narrowed to undefined of a binding a call may write yet
--- PASS: TestLowerNeverLiveProbe (25.48s)
    --- PASS: TestLowerNeverLiveProbe/n1.ts (25.05s)
    --- PASS: TestLowerNeverLiveProbe/n1.a (0.17s)
    --- PASS: TestLowerNeverLiveProbe/n2.ts (0.14s)
    --- PASS: TestLowerNeverLiveProbe/n2.a (0.13s)
PASS
gate cache: native hits=0 misses=0
gate cache: node hits=0 misses=4
gate cache: probe hits=0 misses=0
ok  	github.com/system-inc/adamic/internal/oracle	25.491s
```

No silent wrong answer was observed: both programs are explicitly unsupported today at this HEAD, despite passing TypeScript and executing successfully on Node.
