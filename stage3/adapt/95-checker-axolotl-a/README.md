# 95: required index-signature parameter reads

Head A removes exactly 13 checker diagnostics: checker.ts own-file 30 to 17,
tsc entry whole-program 144 to 131, with no added diagnostic anywhere.
Stock TypeScript 6.0.3 emits byte-identical checker JavaScript, with comments
retained and removed. Only 13 reads in one private nested function change.
The one-read reversion restores exactly one diagnostic and fails the verifier.

lane: run as part of a Correctness wave

## Scope and invariant

Base: fetched main `a23ba096388ceb3774b94d2db70009316d90b0cf`, branch
`typescript/axolotl-checker-a`. Upstream TypeScript 6.0.3 is pinned at
`050880ce59e30b356b686bd3144efe24f875ebc8`.
[sites.md](sites.md) inventories all 30 diagnostics and proposes dispositions
for the other 17; B, C and D have no implementation here. The brief does not
assign those heads' membership, so the ledger explicitly defines its survey
labels. The public TypeChecker contract mismatch at checker.ts:1683 is reported
only. No public API line is edited.

`checkGrammarIndexSignatureParameters` deliberately saves
`const parameter = node.parameters[0]` before checking arity. An empty list
must still report the invalid-arity diagnostic on `node`. Both arms of the
invalid-arity branch return. Past it, the parser-built NodeArray is dense and
has exactly one populated ParameterDeclaration. The immutable saved local
therefore holds that declaration at every later use, including after calls.

The adapter adds a required-value `!` at each of the 13 later reads, leaving
the initializer and all invalid-arity handling alone. Each read remains at
its original evaluation point. In Adamic these are loud required-value checks;
on stock Node the syntax erases. An additional runtime guard or restructured
narrowing would change emitted JavaScript. No such JavaScript-changing rewrite
is included. No optional chain, default, hoist or initializer assertion is added.

Adaptation 90 pins checker.ts's full source hash. 95 runs after 90, so 90's
pin is checked before any 95 edit. Neither 90 nor its pin is changed. Rerun 95
on its own output for idempotence; do not rerun 90 on a tree already changed
by a later adaptation. 95's before hash is exactly 90's checker after hash.

## Adapter contract

Following 31's parsed occurrence inventory and 91's typed guards, stock npm
TypeScript 6.0.3 parses current text and resolves the saved local symbol.
The adapter requires one private function under `createTypeChecker`, the exact
absence-handling initializer, the exactly-one guard and its two returns,
13 symbol-identical reads, their ordered parsed parent contexts, and the exact
`ParameterDeclaration | undefined` type at each operand. Full before/after
source hashes and a normalized owner hash refuse any uninventoried shape.
Lines and columns in sites.json are documentation, never edit addresses.

All edits are planned and parsed, both stock JavaScript comparisons pass,
and the unchanged input is rechecked before any write. Source slices preserve
CRLF without a source printer. Only checker.ts is written. A second run reports
`files: 0, assertions: 0`; `--check` performs no writes.

`verify.cjs` independently compares all 711 source files, requires exactly the
13 assertion spellings in the private function and otherwise identical source,
checks CRLF counts, and compares complete stock transpileModule JavaScript.
It does not call the adapter's edit planner. Before/after checker hashes and
all 13 parsed addresses are in [sites.json](sites.json).

## Diagnostic proof and mutant

The meter's own executable, built with
`go build -o /tmp/axolotl-census ./stage3/census/tool`, loads only
`src/tsc/tsc.ts`, using production `internal/load` options unchanged.
The census emits a root observation and an identical aggregate observation;
`compare-census.py` verifies equality and counts one observation, rather than
adding the same diagnostics twice. Own-file findings use the primary file
location, not indented diagnostic elaborations.

[proofs/reproduction.json](proofs/reproduction.json) records all 144 normalized
entry diagnostics and their exact match against the brief's committed meter
run. [proofs/census.json](proofs/census.json) checks complete diagnostic set
differences, including all 13 exact locations and no added diagnostic:

| Tree | Whole program | checker.ts own file | Head A |
| --- | ---: | ---: | ---: |
| Before | 144 | 30 | 13 |
| After | 131 | 17 | 0 |
| One-read mutant | 132 | 18 | 1 |

The isolated mutant undoes only the first `parameter!.dotDotDotToken` read,
restoring TS18048 at checker.ts:52426:13. The independent verifier exits 1:
`all 13 required read checks must be present`, `12 !== 13`. The adapter's
own check also exits 1 for `uninventoried checker source`. The real census
restores exactly that original diagnostic, with every other finding unchanged.

A separate in-memory mutant changes the private grammar function's final
`return false` to `return true`. The independent stock JavaScript comparator
rejects it in both comments modes. [prove-mutants.cjs](prove-mutants.cjs)
replays the rejection controls; [proofs/mutants.json](proofs/mutants.json)
records their catchers. No mutant enters the oracle tree.

## JavaScript identity

[proofs/source.json](proofs/source.json) records one changed source file out of
711 and exact stock checker JavaScript equality:

| Comments | Bytes | SHA-256, before and after |
| --- | ---: | --- |
| Retained | 2,892,419 | `786ee138ff00765ab7c7ea7651de48df32c5981ec2c0810c68e302570d566f6e` |
| Removed | 2,354,716 | `3664a2f5fb2aba0e839a3153e6ff4d093b530db890b5b2433be6cd38e6e91aac` |

There is no JavaScript diff and no JavaScript-changing head requiring Node
before/after behavioral witnesses. Public source declarations are unchanged;
the only changed function is private and its explicit boolean result is intact.

## Commands and environment

All test output goes to log files. CRLF counts are 54,434 before and after.
Toolchain setup completed successfully:
Node ready 0.071s; Go ready 0.074s; clang ready 0.486s; tsc ready 0.744s;
markdown ready 1.766s; submodules ready 229.932s; Go build ready 498.729s;
build cache warm 498.847s; done 498.883s. `nproc=5`, cgroup CPU quota 4.
Node v24.19.0, Go 1.27.1, clang 20.1.8. Every build/test shell sources
`/workspace/adamic-tools/env.sh`.

The initial census build failed with
`cannot load module cohere/TypeScript/tsc ... go.mod: no such file or directory`
while setup was still fetching the nested submodule. Retrying after that
checkout succeeded. No compiler workaround or option change was used.

The commands below record the historical run; its first control apply ran
before directory 95 existed. To reproduce a control now, run full apply into
a fresh directory, copy that tree, and restore only checker.ts from apply's
`refs/stage3/95-checker-axolotl-a-before` Git tree into the control. The copy
then has exactly the control source hash in sites.json. For example:

```sh
git -C CONTROL_COPY show refs/stage3/95-checker-axolotl-a-before:src/compiler/checker.ts > CONTROL_COPY/src/compiler/checker.ts
```

```sh
source /workspace/adamic-tools/env.sh
export NODE_PATH="$PWD/stage3/api/node_modules"
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/axolotl-upstream
bash stage3/apply.sh /workspace/scratch/axolotl-before > /tmp/axolotl-apply-before.log 2>&1
go build -o /tmp/axolotl-census ./stage3/census/tool > /tmp/axolotl-census-build.log 2>&1
/tmp/axolotl-census /workspace/scratch/axolotl-before/src/tsc/tsc.ts /workspace/scratch/axolotl-before-census.jsonl > /tmp/axolotl-before-census.log 2>&1
# Copy the before tree to a fresh after path, then:
node stage3/adapt/95-checker-axolotl-a/adapt.cjs /workspace/scratch/axolotl-after > /tmp/axolotl-adapt.log 2>&1
node stage3/adapt/95-checker-axolotl-a/verify.cjs /workspace/scratch/axolotl-before /workspace/scratch/axolotl-after > /tmp/axolotl-verify.log 2>&1
node stage3/adapt/95-checker-axolotl-a/adapt.cjs /workspace/scratch/axolotl-after > /tmp/axolotl-idempotence.log 2>&1
/tmp/axolotl-census /workspace/scratch/axolotl-after/src/tsc/tsc.ts /workspace/scratch/axolotl-after-census.jsonl > /tmp/axolotl-after-census.log 2>&1
# Copy after to an isolated mutant and undo only occurrence 1, then:
/tmp/axolotl-census /workspace/scratch/axolotl-mutant/src/tsc/tsc.ts /workspace/scratch/axolotl-mutant-census.jsonl > /tmp/axolotl-mutant-census.log 2>&1
python3 stage3/adapt/95-checker-axolotl-a/compare-census.py /workspace/scratch/axolotl-before-census.jsonl /workspace/scratch/axolotl-after-census.jsonl /workspace/scratch/axolotl-mutant-census.jsonl > /tmp/axolotl-census-comparison.json 2>&1
node stage3/adapt/95-checker-axolotl-a/prove-mutants.cjs /workspace/scratch/axolotl-before /workspace/scratch/axolotl-after /workspace/scratch/axolotl-mutant > /tmp/axolotl-mutants.json 2>&1
timeout --kill-after=20s 880 bash stage3/oracle/run.sh /workspace/scratch/axolotl-after /workspace/scratch/axolotl-oracle --limit-seconds 720 > /tmp/axolotl-oracle.log 2>&1
```

Ordinary apply discovers the new numbered adapter and writes the new patch table
inside its output. `--write-table` is an explicit optional publication action,
not a requirement for a new adaptation. stage3/patch-set.md is left unchanged.
No stage3/lane files or sharded commands are changed or run, following
Correctness's final instruction. Native tsc compilation, lowering, browser and
ESLint integrations, and the full repository Go gate are outside this head.

## Completed end-to-end apply and standalone oracle

Fresh numbered apply, including 95, exits 0 in 211 seconds. Its incremental
row is `95-checker-axolotl-a | 1 | 13 | 13`. The fresh tree's complete 711-source
map equals the oracle tree's map. Its separate real census again has exactly
131 whole-program and 17 checker own-file diagnostics, identical to the first
after observation. Rerunning 95 on this fresh output makes zero edits.
[proofs/apply.json](proofs/apply.json), [proofs/final-source.json](proofs/final-source.json),
[proofs/final-idempotence.json](proofs/final-idempotence.json) and the compressed
apply log retain these results; [proofs/patch-set.md](proofs/patch-set.md) is
the generated output table. The checked-in table remains untouched.

The standalone default-runner oracle completed within 15 minutes. It used all
runners, no filter, eight upstream workers, `--light=false`, and lint disabled.
Its raw status is **fail**, exit 1, solely for the existing public API
acknowledgement failure. This is not a claim of zero upstream failing tests.

| Phase | Exit | Wall seconds |
| --- | ---: | ---: |
| Install | 0 | 2.553 |
| Build | 0 | 2.187 |
| Tests | 1 | 356.202 |
| Complete oracle | 1 | 361.032 |

Counts: **106,366 passing, 1 failing, 0 pending**. The only failure is exactly
`unittests:: Public APIs for typescript.d.ts should be acknowledged when they change`.
The sole baseline difference is `api/typescript.d.ts`; **new baseline
differences are 0**. There are no timeouts or exceptions. Cgroup OOM and
OOM-kill counts remain zero; memory-limit event count is 20,809, recorded
rather than described as no memory pressure.

Both actual declaration artifacts match main's recorded bytes in
`40-explicit-any/evidence/brands-undefined-3/api-runtime.json`:

- Built/local API: 590,996 bytes, SHA-256
  `b901adcc6ec29e299ec21401b2a155b2b2c14c046ec743fed6ee45679b726d5d`.
- Accepted reference API: 588,626 bytes, SHA-256
  `f87d6c30da287293aae741711fa94594b75d49e0977cea6bfd5982f90e2bafcf`.

The local API baseline also equals the built declaration byte for byte. Thus
both sides of the one existing baseline diff match main; 95 introduces no
public API delta. The retained complete baseline.diff is 51,123 bytes,
SHA-256 `6d52cd09811c6846c977ea492071c14325816a631706cc79128c4299b9faa031`.
The older pre-brand-retype 51,093-byte diff is not used as today's expected hash.

[proofs/oracle-summary.json](proofs/oracle-summary.json) and
[proofs/oracle-report.json](proofs/oracle-report.json) retain the raw verdict,
counts, exits and byte comparisons. Compressed phase logs and the complete
baseline diff are beside them. No lane acceptance predicate is run or changed;
Loom's Correctness wave supplies that proof.
