Attempted the pinned CommonJS checker observer natively; compilation stops before executing the checker.
Publication base 031a1259; candidate 79dd1aba; stricter tip 8f32e51e; never-pushed resolved scratch merge 67193eaa.
Both split modes stop at undeclared require; real Node declarations expose NotYet: node:module.require instead.
Two Node witnesses and computed-input mutants pass their checks; ten evidence mutants fail the verifier; Node baseline covers 301 projects.
No native checker binary or native comparison; the walk ends at the module-scope observer entry, not fifteen checker-internal stops.

## Compiler composition

The remote candidate heads were ordered by commit time. The newest was
`cloud/land-area-next-auto-547551cb`, 79dd1abae85972b8123fb813d2d98e49ef612730,
2026-10-08 12:33:05 -0600; it was not an ancestor of fetched main
031a1259bc7973934792dc6cb1bd4074fc2204b9. Other candidate heads were bfaa98b0,
f445fa4e and 0c99fc54, all older. No candidate was substituted.

A detached scratch at that candidate merged `codex/stricter-options-next`,
8f32e51e8fc41b8f1177453213ca5453ce764486. Three conflicts were resolved:
keep candidate internal/lower/lower_test.go, and keep the candidate's deletions
of internal/native/runtime/node_process.c and parallel.c. All automatically
merged compiler changes were retained. This is a **resolved composition**, not
an assertion that the conflicting branch tips combine without choices. No test
suite validated those choices. Merge diagnostics and resolution output are
retained. Scratch merge 67193eaab72750fcf32b0abe1e34988499397ffc was never pushed.
Its compiler builds successfully and embeds vcs.modified=false. Dependencies
are pinned cohere 7945d102 and TypeScript d92d9bfe; metadata is in binary.txt.

The evidence branch has no compiler/runtime/library/adaptation modifications.

## What was actually attempted

The task names scout 2de9fc1b01d3da3beb35c554e70d98cd6d9fe536. That exact commit
was used, rather than its later train report. Its checker-dump.cjs is:

```js
'use strict';
require('./component-dump.cjs').run('checker');
```

component-dump.cjs is the real untyped CommonJS Node observer. It constructs
Programs, calls getPreEmitDiagnostics, and implements the JSONL cache protocol.
The scout's comparator consumes a native executable but **does not supply one**.
Later gathered slice entries import compiler declarations; they are admission
roots and are not implemented native JSONL drivers. They were not substituted
for the requested observer.

The raw build copied checker-dump.cjs byte for byte into an external
checker-entry.a, beside the unchanged component-dump.cjs. This exposes the same
source to Adamic without committing upstream/compiler source or asking an
unsupported filename suffix to impersonate a native program. A second profile
adds only `import type {} from "node:util";` before those exact bytes to activate
the loader's real pinned Node declarations. No fake require declaration,
dynamic-module stub or diagnostic filtering was used.

Install the scratch's locked stage3/api dependencies and run the compiler from
that checkout: its loader searches upward from its cwd for that exact Node
seat. An initial declaration profile failed because this seat was absent; npm ci
and the correct cwd cleared that prerequisite. The final observations below
exclude that preliminary setup failure and early attempts made before the
scratch compiler binary existed.

## Stops in order and ranking

| Stop | Profile and site | Exact result | Minimal program | Owner (inference) | Listed in ranking? |
|---|---|---|---|---|---|
| 1 | Unchanged observer, checker-entry.a:2:1 | TS2591: Cannot find name require | probes/commonjs.a | Compiler: CommonJS/global declaration admission | No exact reason; checker errors are an aggregate other cause |
| 2 | Node declarations, checker-node-declarations.a:3:1 | NotYet: node:module.require | probes/node-require.a | Library: native Node module-loading operation | No exact reason |

Both ADAMIC_NATIVE_SPLIT=0 and 1 exit 1 in each profile; complete stdout/stderr
match byte for byte within each pair. Stop 2's full native text is
`stage 0 can't lower node:module.require yet`. The typed minimal program stops
at the same native operation at 2:14. The raw minimal program has the required
TS2591 first-line a-check header; the NotYet program has none.

The second call is at module scope, outside every function body, and is the
only executable observer entry. Replacing it with a throwing/empty placeholder
would remove the observer rather than reveal its checker implementation.
Consequently the walk terminates after two profiles of this entry boundary.
**Fifteen checker-internal stops and a native 301-project difference cannot be
measured from this delivered driver.** A native CommonJS loader or a separately
typed, semantics-held driver is required before that experiment reaches tsc.
Neither was invented as a silent adaptation in this evidence unit.

Ranking input is origin/codex/stage3-hidden-ranking at
ec0b16c04f3bbdfbaf3932ab01307f90b08156fd, RESULT.json. ranking-reasons.json
retains all ranked reason texts and their byte credits; result.json records the
full input hash. Matching converts the CLI's NotYet wrapper to its canonical
`NotYet: node:module.require` spelling, then compares exact kind/reason text.
No family/fuzzy match or new hidden-byte estimate is claimed. Absence from this
inventory does not mean the checker aggregate has no related bytes.

## Node observations and proofs

The exact pinned stock driver ran all 300 selected acceptance projects plus
tiny. It selected 301/301, with no reused projects, in 10.798145535 driver
seconds. Golden SHA256:
`f5b3ccb280ac67ed516240b52b476a9748dd8a60a516cec86f76be371265828a`, matching
its recorded acceptance baseline. The full request and golden bytes are saved
compressed. This is a Node baseline, not native comparator success.

Both .a witnesses compute path.basename('/tmp/leaf'), print leaf, exit 0 and
have empty stderr on Node. Each scratch mutant changes the runtime input to
/tmp/changed, prints changed, still exits 0 with empty stderr, and is killed
only by the expected-output check. No sanitizer or compiler failure is counted
as a semantic mutant kill. No native allocations/lifetime claims are made.

The verifier cross-checks both split pairs against raw outputs, pins and clean
binary metadata, both Node witnesses/mutants, the 301-project hash/population,
ranking matches and the required headers. Ten evidence mutants change:
candidate pin, split stderr, Node output, ranking membership, native-run claim,
project population, owner, the TS2591 header, golden bytes and request bytes. Every run fails with its
retained AssertionError. Focused verification and mutation outputs are under
evidence/; counts.md refreshes this territory's local witness registry.

## Reproduction and coverage

All execution output went to files. No package suite, full gate or upstream
suite was run. The substantive commands were:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/step31-native-setup.log 2>&1
source /workspace/adamic-tools/env.sh
# Add detached candidate, merge stricter, resolve exactly the three documented conflicts.
git -C /tmp/step31-native-compiler submodule update --init --recursive --depth 1
(cd /tmp/step31-native-compiler && go build -o /tmp/step31-native-adamic ./cmd/adamic) > /tmp/step31-native-build.log 2>&1
npm ci --prefix /tmp/step31-native-compiler/stage3/api --ignore-scripts --no-audit --no-fund > /tmp/step31-native-api.log 2>&1
# git archive 2de9fc1b stage3/scouts/step31 into /tmp/step31-driver.
# Put the unchanged checker entry beside component-dump.cjs; create the type-only-import profile.
python3 /tmp/step31-driver/stage3/scouts/step31/prepare-corpus.py /tmp/step31-native-projects.json > /tmp/step31-native-corpus.log 2>&1
STEP31_TYPESCRIPT="$HOME/.cache/adamic-stage3/api/node_modules/typescript/lib/typescript.js" node /tmp/step31-driver/stage3/scouts/step31/checker-dump.cjs /tmp/step31-native-projects.json /tmp/step31-native-node-cache > /tmp/step31-native-node.log 2>&1
python3 stage3/scouts/step31/checker-native/run.py > /tmp/step31-native-measure.log 2>&1
python3 stage3/scouts/step31/checker-native/verify.py > /tmp/step31-native-verify.log 2>&1
python3 stage3/scouts/step31/checker-native/mutants.py > /tmp/step31-native-mutants.log 2>&1
```

The archive's stage3/drivers must point to the publication drivers for corpus
materialization. Save the exact ranking RESULT.json as /tmp/step31-ranking.json.
run.py records the observed scratch paths explicitly; use fresh scratch/cache
paths for independent runs. Owner attribution is separate from observed errors.

Setup succeeded: Node 0.023s, Go 0.031s, submodules 0.059s, markdown skipped
step 0.007s, markdown ready 0.074s, clang 0.190s, Go build 47.457s, test binaries
deferred 47.569s, cache warm 47.571s, done 47.596s. nproc=5, cpu.max=400000
100000 (four CPUs), 17.6 GB. Go 1.27.1, Node 24.19.0, clang 20.1.8. Environment
path: /workspace/adamic-tools/env.sh. Full setup lines are in setup.txt.

No native driver executable, native comparator/first difference, checker-internal
stop walk, runtime guards, ownership/leaks, emitter or binder measurement,
full integration gate, or native speed claim was produced. This report preserves
the actual host-entry blocker instead of treating a Node process or compiler
slice as native checker execution.
