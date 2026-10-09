Built the tsc entry with a disposable area/compiler plus stricter-options composition; native emission remains blocked.
Publication main: efe9f404; compiler parents b68b2fe1, c09c22b8, b150f83c, aee98c83, ded85b2d; exact pins and source choices are in evidence.
Both split builds exit 1 at src/tsc/_namespaces/ts.ts:3:15, TS6305; owning-project snapshots admit 13/15 saved sites.
Loader tests, 32 lane tests, evidence verification and ten rejection mutants pass; adapted source and oracle/lane inputs match main.
No adaptation edits; two surviving diagnostics are placeholder artifacts; no tsc binary or --tiny comparison was produced.

## Compiler composition and limits

The experiment is reproducible with `compose-mode.py NEW_SCRATCH` after fetching
the pins in `evidence/mode-c09c22b8/provenance.json`. Merge area/compiler with
options using `-X theirs`, then records/indexed/catch using `-X ours`.
Catch is already in options. Merging these branches directly was incompatible:
class helper changes and typed-array IR versions did not compile together.
The scratch resolves 24 paths to the exact c09 versions (four paths are absent
there). Those choices, individual hashes, merge logs, binary hash and Go build
metadata are preserved. They are **uncommitted scratch changes**, not proposed
compiler fixes. This is an experimental reconciled composition, not evidence
that the full combined compiler tips are compatible. Backend outcomes must be
read with that limitation. No publication compiler file is changed.

After initializing the pinned cohere/TypeScript dependencies, source
`/workspace/adamic-tools/env.sh` and run from the scratch compiler:

```sh
go build -o /tmp/tsc-entry-mode-adamic ./cmd/adamic > /tmp/mode-build.log 2>&1
```

Use Node 24.19.0. Setup completed in 37.978 seconds (Go build 37.942,
Go cache 37.944); `nproc` is 5, quota 4. Full setup output is retained.
The exact binary reports scratch merge 29933a3e and `vcs.modified=true`.

## First native entry stop

The adapted tree was rebuilt using main's unchanged `bash stage3/apply.sh
/tmp/tsc-entry-mode-final-adapted`. Both commands, run from the scratch compiler,
exit 1 with identical output:

```sh
ADAMIC_NATIVE_SPLIT=0 /tmp/tsc-entry-mode-adamic build /tmp/tsc-entry-mode-final-adapted/src/tsc/tsc.ts -o /tmp/tsc-entry-mode-0 > /tmp/entry-0.stdout 2> /tmp/entry-0.stderr
ADAMIC_NATIVE_SPLIT=1 /tmp/tsc-entry-mode-adamic build /tmp/tsc-entry-mode-final-adapted/src/tsc/tsc.ts -o /tmp/tsc-entry-mode-1 > /tmp/entry-1.stdout 2> /tmp/entry-1.stderr
```

First file **src/tsc/_namespaces/ts.ts**, line **3**, column **15**, outside
`src/compiler`:

> error TS6305: Output file '/tmp/tsc-entry-mode-final-adapted/built/local/compiler/_namespaces/ts.d.ts' has not been built from source file '/tmp/tsc-entry-mode-final-adapted/src/compiler/_namespaces/ts.ts'.

This is a project-reference output prerequisite. The new loader uses the entry's
actual `tsconfig`; it does not flatten away its compiler project reference.
Missing referenced declarations cause subsequent property diagnostics. This
failure occurs before the entry build can demonstrate admission or lowering of
its compiler dependency. No missing declaration was fabricated and no project
configuration was weakened to bypass the stop.

The minimal Node witness is `probes/missing-project-output/app/main.a`, importing
`dependency/value.a` and printing 7. Copy these to `main.ts` and `value.ts`, change
the import to `../dependency/value.js`, and use the saved two composite
`tsconfig.json` files under `evidence/mode-c09c22b8/witnesses/`. Leave `out/`
absent. Node's documented stock loader prints `7` and exits 0; the native build
stops at app/main.ts:1:23 with the same TS6305 missing-referenced-output mechanism.
Its message necessarily names the smaller dependency. Authored sources are `.a`.
`mode-witnesses.py BINARY PROBE NEW_OUTPUT` reconstructs and runs these witnesses;
set `SCANNER_TYPESCRIPT` to the stock TypeScript API path used by the Node loader.

The two reached files outside compiler remain `src/tsc/tsc.ts` and
`src/tsc/_namespaces/ts.ts`; the latter owns this first stop.

## All fifteen saved sites

`mode-snapshots.py SCRATCH_COMPILER ADAPTED_TREE NEW_OUTPUT` builds a Go probe
against the scratch loader, recreates each prior snapshot in order, preserving
CRLF and saved UTF-16 offsets, and loads that site's **owning compiler project**
through production `load.Load`. It does not substitute an admission allowlist.
The baseline stop number, original coordinate, probe and complete message are
retained in `admission.json`; raw full-loader outputs are `NN-load.json`.
This is a per-site checker comparison, not fifteen successful native builds.

| Stop | Original site | Mode result |
| --- | --- | --- |
| 1 | builder.ts:1246:69 | Not diagnosed |
| 2 | builder.ts:1258:65 | Not diagnosed |
| 3 | builder.ts:2273:9 | Scheduled optional-presence check |
| 4 | builder.ts:395:118 | Not diagnosed |
| 5 | builder.ts:991:17 | Not diagnosed |
| 6 | checker.ts:10022:63 | Not diagnosed |
| 7 | checker.ts:10026:97 | Not diagnosed |
| 8 | checker.ts:10033:67 | Not diagnosed |
| 9 | checker.ts:10034:53 | Ordinary TS2345 remains; placeholder artifact |
| 10 | checker.ts:10319:83 | Not diagnosed |
| 11 | checker.ts:10320:51 | Not diagnosed |
| 12 | checker.ts:10941:33 | Scheduled optional-presence check |
| 13 | checker.ts:12034:30 | Scheduled indexed-presence check |
| 14 | checker.ts:12859:13 | Ordinary TS2322 remains; placeholder artifact |
| 15 | checker.ts:13986:47 | Not diagnosed |

Count: **13 admitted at the saved sites**, ten absent under actual project
options and three scheduled checks. A scheduled check is a loader obligation,
not proof that its guard was emitted and executed. Stops 9 and 14 were already
absent from pristine diagnostics in the baseline; throwing replacements changed
inferred callback/function result types. Their saved scratch coordinates are
10034:53 and 12579:13 respectively. The minimal programs remain
`probes/09-never-argument.a` and `probes/14-void-result.a`; Node prints `real` and
`ok`, while the project loader reports TS2345 and TS2322 respectively.

## Mode scope

The mode is not restricted to the 79 ledger roots. `nearestProject` and
`projectOptionsForRoots` in `internal/load/project_loader.go` select the nearest
`tsconfig.json` for explicit `.ts` roots. That project's source graph uses its
options plus mandatory strict null/function checks; supported stricter-option
sites go through the production option audit. Explicit `.a` roots and standalone
`.ts` roots retain Adamic's fallback; mixed checker ownership is refused.
Thus it is project-driven, rather than unconditional for every possible `.ts`.
`LoadOptionLedger` calls ordinary `Load`; ledger identities provide reporting
and never override admission. Contrary to the question's premise, **builder.ts
is entry 2 in the current 79-root manifest** (JSON line 3).

The positive scope witness `probes/project-indexed-read.a`, copied into a `.ts`
project with the saved config, lives outside every ledger root. Its `values[0]`
argument schedules an indexed-presence check and the loader succeeds. Node
prints 2. Native emission then refuses `console.log` with "stage 0 can't lower
the library method log yet" in this reconciled compiler. That backend refusal
is retained; this witness establishes loader scope, not guard execution.

## Adaptation audit of what remains

Following the added instruction, the adaptation decision is limited to stops
**9 and 14**, which remain after the mode check. The rule sections of the
20/30/31/32/33/75 READMEs define the following boundaries:

| Rule | Stop 9, literal argument to never | Stop 14, void or Type result |
| --- | --- | --- |
| 20 optional declarations | No optional property declaration or present-undefined object assignment at this stopping site | Function result includes void; no optional declaration to widen |
| 30 indexed core | Not an indexed read and not a core target | Not an indexed read and not a core target |
| 31 indexed checker | File is covered, but call's never parameter is not an indexed read | File is covered, but inferred void return is not an indexed read |
| 32 indexed program | checker.ts is outside its program/builder target files; no reviewed erased-iterator call shape | Same; neither indexed read nor reviewed iterator generic closure |
| 33 indexed emit | Not an emit target or indexed read | Not an emit target or indexed read |
| 75 optional widening | No fixed reviewed lazy-cache/visibility/JSON optional owner | Function result union is outside those optional owner families |

Neither is a missed file-list application, an excluded indexed pattern that
satisfies the rule, or a later-added optional field. They are ordinary compiler
checker refusals caused by the exploratory placeholders. The proper action is
to preserve them as such, not widen adaptations to make scratch-induced types
pass. No existing adaptation should cover either as written. No adaptation
extension was made. For the thirteen admitted sites, the added instruction's
"adaptation check only for what remains" supersedes the earlier request to
change applicable rules before testing this mode.

## Identity, tests and mutants

All **746** files under adapted `src/` have identical before/after SHA-256
maps; generated `patch-set.md` is identical. `git diff origin/main --
stage3/adapt stage3/apply.py stage3/apply.sh stage3/oracle stage3/lane` is empty.
These prove source and oracle/lane **input** identity. Full upstream oracle/lane
outputs were not rerun: the requirement to prove their outputs after adaptation
changes is conditional, and no adaptation changed. No full gate or full Go
suite is claimed.

Commands, with outputs retained in this evidence directory:

```sh
go test ./internal/load -run 'TestOptionLedger|TestUnsupportedOptionContract|TestPerFileProject|TestProjectOptions' -count=1 -timeout 20m > loader-tests.log 2>&1
python3 -m unittest discover -s stage3/lane -p 'test_*.py' > lane-tests.log 2>&1
python3 stage3/drivers/tsc-entry/verify-mode.py > verification.log 2>&1
python3 stage3/drivers/tsc-entry/mode-mutants.py > mutants.log 2>&1
```

Loader tests pass in 3.290s; lane passes 32 tests. Verification cross-checks the
fifteen classifications against raw loader sites, split outcomes, source/control
identity, project scope and four Node witnesses. Mutants change a classification,
remove a stop, change a split result, corrupt the first diagnostic, change a
source hash, claim changed control inputs, remove builder from the manifest,
change the Node oracle, erase the outside-ledger scheduled check, and erase a
remaining ordinary diagnostic. Every mutant is caught by `verify-mode.py`; the
mutant runner first requires the original evidence to pass.

No adaptation-cleared stops: **0**. Project mode admits **13** saved checker
stops; the native entry is still blocked earlier by TS6305. No successful tsc
binary, inserted-guard execution, --tiny comparison, or compatibility proof for
the unreconciled merged compiler branches is claimed.
