Built a fresh adapted tsc entry tree after merging main efe9f404 into the existing branch.
Merge commit: 364cd19a; both native split builds use compiler source identical to main.
First stop: builder.ts:1246:69, TS2345; the minimal Node witness exits 0.
Latent status: new by exact message text in main's stage3/census/latent/REPORT.md.
Fifteen checker stops: all new versus the requested census, 12 distinct messages; no native binary.

The exact stopping diagnostic is:

```text
src/compiler/builder.ts:1246:69: error TS2345: Argument of type 'Path | undefined' is not assignable to parameter of type 'string'.
  Type 'undefined' is not assignable to type 'string'.
```

Both `ADAMIC_NATIVE_SPLIT=0` and `1 ADAMIC_NATIVE_JOBS=5` exit 1 at checking,
with identical full diagnostic streams. The first minimal witness is
[01-indexed-path.a](probes/01-indexed-path.a); Node 24.19.0 prints `undefined`
and exits 0. Its native attempt gives the same TS2345 text and exits 1.

Every reached file outside `src/compiler`, called out separately:

- `src/tsc/tsc.ts`
- `src/tsc/_namespaces/ts.ts`

The closure is 81 files and matches stock TypeScript's independent program
loader. `executeCommandLine.ts` remains inside `src/compiler`.
The actual adapted tsc entry runs on Node and prints `Version 6.0.3`.

The requested main latent report records main b8fb957a, not current efe9f404.
Matching uses only its main-unmerged all-exact-reasons table, unescapes Markdown
pipes, and removes diagnostic/kind prefixes. All fifteen checker messages are absent (12 distinct),
so they are marked new relative to that table. No lowering reason was observed.
The report hash and original census provenance are recorded in the evidence.

Setup succeeded: Go ready 0.109s, Node ready 0.114s, markdown ready 0.255s,
submodules ready 0.270s, clang ready 0.657s, Go build ready 60.453s,
cache warm 60.587s, done 60.755s. `nproc=5`, cgroup quota 4 CPUs.
The environment file is `/workspace/adamic-tools/env.sh`.

Fresh evidence is under `evidence/main-efe9f404/`; the preceding unit's evidence
remains intact. All production compiler and adaptation sources match main.
The full repository gate and the conditional native tiny harness have not run.

The first stop was pushed in `c3d8934b` before continuation completed.
The continuation used `/tmp/tsc-entry-refresh-src`, a separate uncommitted copy,
and replaced fourteen bodies with throwing placeholders. Both split modes
exit 1 at each of fifteen stops and have byte-identical diagnostic streams.
The sequence is unchanged from the preceding unit. All fifteen sites are
inside `src/compiler`; neither outside file becomes a stopping site.

The table uses original adapted-tree coordinates. Scratch coordinates, exact
full messages, replacement bodies, exit codes and Node stdout are recorded in
[evidence/main-efe9f404/stops.json](evidence/main-efe9f404/stops.json).
Each linked `.a` is a minimal Node-running witness; all fifteen exit 0.
Thirteen reproduce the complete first diagnostic message. Stops 3 and 12 use
smaller structural object types and reproduce TS2375/TS2379 and the same
present-undefined optional-field failure, with different displayed object types.
Stops 9 and 14 are absent from the pristine diagnostic stream and arise after
placeholders change inference. They are experiment artifacts, not established
failures of the unmodified tree.

| Stop | File within src/compiler | Original line:column | Message | Latent census | Minimal program |
| --- | --- | --- | --- | --- | --- |
| 1 | builder.ts | 1246:69 | error TS2345: Argument of type 'Path \| undefined' is not assignable to parameter of type 'string'. | New | [01-indexed-path.a](probes/01-indexed-path.a) |
| 2 | builder.ts | 1258:65 | error TS2488: Type '[Path, FileInfo] \| undefined' must have a '[Symbol.iterator]()' method that returns an iterator. | New | [02-tuple-parameter.a](probes/02-tuple-parameter.a) |
| 3 | builder.ts | 2273:9 | error TS2375: Type '{ fileInfos: Map<Path, BuilderState.FileInfo>; compilerOptions: CompilerOptions; semanticDiagnosticsPerFile: Map<Path, readonly ReusableDiagnostic[]>; ... 7 more ...; checkPending: true \| undefined; }' is not assignable to type 'ReusableBuilderProgramState' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties. | New | [03-present-undefined-field.a](probes/03-present-undefined-field.a) |
| 4 | builder.ts | 395:118 | error TS2345: Argument of type 'Path \| undefined' is not assignable to parameter of type 'Path'. | New | [04-optional-path-argument.a](probes/04-optional-path-argument.a) |
| 5 | builder.ts | 991:17 | error TS2345: Argument of type 'Path \| undefined' is not assignable to parameter of type 'Path'. | New | [05-optional-path-argument.a](probes/05-optional-path-argument.a) |
| 6 | checker.ts | 10022:63 | error TS2345: Argument of type 'Symbol \| undefined' is not assignable to parameter of type 'Symbol'. | New | [06-optional-symbol-argument.a](probes/06-optional-symbol-argument.a) |
| 7 | checker.ts | 10026:97 | error TS2345: Argument of type 'Symbol \| undefined' is not assignable to parameter of type 'Symbol'. | New | [07-optional-symbol-argument.a](probes/07-optional-symbol-argument.a) |
| 8 | checker.ts | 10033:67 | error TS18048: 'm' is possibly 'undefined'. | New | [08-optional-member-read.a](probes/08-optional-member-read.a) |
| 9 | checker.ts | 10034:53 | error TS2345: Argument of type '"real"' is not assignable to parameter of type 'never'. | New | [09-never-argument.a](probes/09-never-argument.a) |
| 10 | checker.ts | 10319:83 | error TS2345: Argument of type 'Symbol \| undefined' is not assignable to parameter of type 'Symbol'. | New | [10-optional-symbol-argument.a](probes/10-optional-symbol-argument.a) |
| 11 | checker.ts | 10320:51 | error TS2345: Argument of type '(Symbol \| undefined)[]' is not assignable to parameter of type 'readonly Symbol[]'. | New | [11-optional-symbol-array.a](probes/11-optional-symbol-array.a) |
| 12 | checker.ts | 10941:33 | error TS2379: Argument of type '{ name: ComputedPropertyName \| Identifier \| NumericLiteral \| PrivateIdentifier \| StringLiteral; questionToken: PunctuationToken<...> \| undefined; modifiers: ... \| undefined; }' is not assignable to parameter of type 'SignatureToSignatureDeclarationOptions' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties. | New | [12-present-undefined-argument.a](probes/12-present-undefined-argument.a) |
| 13 | checker.ts | 12034:30 | error TS18048: 'file' is possibly 'undefined'. | New | [13-optional-file-read.a](probes/13-optional-file-read.a) |
| 14 | checker.ts | 12859:13 | error TS2322: Type 'void \| Type' is not assignable to type 'Type \| undefined'. | New | [14-void-result.a](probes/14-void-result.a) |
| 15 | checker.ts | 13986:47 | error TS2345: Argument of type 'Declaration \| undefined' is not assignable to parameter of type 'Node'. | New | [15-optional-declaration-argument.a](probes/15-optional-declaration-argument.a) |

First minimal program, running on Node with stdout `undefined` and exit 0:

```typescript
type Path = string & { readonly pathBrand: unknown };
function accept(path: string): void { console.log(String(path)); }
const paths: readonly Path[] = [];
accept(paths[0]);
```

Commands run from the checkout, with logs retained in the refreshed evidence:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/tsc-entry-refresh-setup.log 2>&1
source /workspace/adamic-tools/env.sh
bash stage3/drivers/tsc-entry/run.sh /tmp/tsc-entry-refresh > /tmp/tsc-entry-refresh-run.log 2>&1
export SCANNER_TYPESCRIPT=/home/agent/.cache/adamic-stage3/api/node_modules/typescript/lib/typescript.js
cp -a /tmp/tsc-entry-refresh/adapted/src /tmp/tsc-entry-refresh-src
python3 stage3/drivers/tsc-entry/continue.py /tmp/tsc-entry-refresh-src /tmp/tsc-entry-refresh/adamic /tmp/tsc-entry-refresh-progress > /tmp/tsc-entry-refresh-progress.log 2>&1
python3 stage3/drivers/tsc-entry/collect.py /tmp/tsc-entry-refresh-progress --evidence stage3/drivers/tsc-entry/evidence/main-efe9f404 --census-report stage3/census/latent/REPORT.md > /tmp/tsc-entry-refresh-collect.log 2>&1
python3 stage3/drivers/tsc-entry/verify.py stage3/drivers/tsc-entry/evidence/main-efe9f404 > /tmp/tsc-entry-refresh-verify.log 2>&1
python3 stage3/drivers/tsc-entry/mutants.py --evidence stage3/drivers/tsc-entry/evidence/main-efe9f404 > /tmp/tsc-entry-refresh-mutants.log 2>&1
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/undefined_references.a$' -count=1 -timeout 30m > /tmp/tsc-entry-refresh-oracle.log 2>&1
```

Verification reports `15 stops verified; 15 Node witnesses exit 0; both split
modes agree`. The filtered oracle passes in 0.740s. Setup, the graph check and
Node CLI version check pass. The runner records native failures rather than
turning them into successful binary comparisons.

Five mutants were run and caught:

| Mutant | What caught it | Exit |
| --- | --- | --- |
| Remove traversal of the entry's imports | Independent stock TypeScript program closure comparison | 1 |
| Replace probe 1's diagnostic with its defined-value control's real lowering diagnostic | Exact probe diagnostic comparison | 1 |
| Change probe 1's Node stdout | Expected Node output comparison | 1 |
| Append a byte-changing line to split 1 stderr | Full split stream comparison | 1 |
| Omit stop 15 from the stop list | Required population check | 1 |

The defined-value control changes `paths[0]` to `paths[0] ?? ""`.
`adamic types` exits 0, proving the initial checker failure can disappear;
the native control instead reaches `stage 0 can't lower an array of Path yet`.
This is not a successful native control. Logs for each mutant and the control
are retained under the refreshed evidence directory.

Coverage stops at fifteen checker diagnostics. There is no binary to run the
conditional `--tiny` harness, no native/Node diagnostics comparison, and no
observed lowering frontier. The full repository gate was not run; the filtered
oracle and territory's evidence/graph checks are the recorded validation.
Compiler, adaptation and library sources were not edited by this unit.
