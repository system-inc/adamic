Built: .a candidates for structure/tailwind-no-physical-direction, @next/next/no-assign-module-variable and @typescript-eslint/default-param-last in owned rule directories.
Commits: claim a2febc6e pushed before code; module d4dbd1b6; parameters 25e98fdb; Tailwind and evidence 2b486ce0.
Commands and outputs: overlay corpus parity PASS, 214 files / 12,630,367 bytes; supported upstream PASS, 183 cases / 57,931 bytes; setup 28s, nproc 5.
Mutants: direction_escape_removed, module_declaration_suppressed and optional_parameter_ignored all ran successfully and were caught only by output comparison on source Node, emitted JavaScript and sanitized native.
Not covered: complete upstream parity fails on one JSX case; default .a registration and the shared profile test remain blocked; full repository test gate was not run.

All prior commits were pushed first. Every origin head was fetched with
`git fetch --no-recurse-submodules origin '+refs/heads/*:refs/remotes/origin/*'`.
Selection checked ported rules on main ef3d907e and claim files on all fetched
origin branches. Helper-ready position 46 was the last available helper rule.
The next two available entries came from the inventory branch's syntax-ready
cohort, in its order. Claim a2febc6e was pushed before source edits.

Each directory owns its descriptor, .a rule and messages, Go adapter, mutation,
and raw witness. No shared dispatch, registry, oracle, copied-file list, parser,
native compiler or lowering source was changed. No authored Adamic .ts files
were added. These three Go rules offer neither fixes nor suggestions: parity
checks cover descriptions, spans, empty fix/suggestion fields and unchanged
fixed source. Tailwind preserves Go's filename restriction to .ts and .tsx,
including silence for .a sources, Unicode White_Space splitting and all twenty
physical/logical mapping entries. Parameter initialization examines scanner gaps
between direct AST children to distinguish actual defaults from nested binding
initializers, type syntax and comments.

The shared foundation on this branch still requires rule.ts and its profile test
ranges over the old portFiles slice, which is now a function. Observed default
commands failed as follows:

```
go run ./cmd/lint-registry
open stage1/cohere/lint/rules/next-no-assign-module-variable/rule.ts: no such file or directory
exit status 1

go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1
stage1/cohere/lint/profile_test.go:32:23: cannot range over portFiles
FAIL [build failed]
```

The ownership rule prevents applying the necessary shared changes in this unit.
The reviewable compatibility.patch under the Tailwind directory is applied only
to scratch copies through Go's overlay. It permits .a directory registration,
source copying/import rewriting and mutants, repairs the profile function call,
and compares emitted JavaScript too. Capture retains upstream filename suffixes
and includes the filename in its deduplication key. comparison_oracle.go.txt is
the existing Go comparison driver with TSX/JSX ScriptKind selected by extension;
upstream Go rule implementations are unchanged. Previously forcing every case
to .ts would erase a material part of Tailwind's file semantics. These commits
are bounded candidates, not a claim that the default integration gate is green.

Reproduce after sourcing /workspace/adamic-tools/env.sh, with a checkout of
TypeScript v6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8:

```
python3 stage1/cohere/lint/rules/structure-tailwind-no-physical-direction/validate.py \
  --scratch /tmp/lint-wave1-08-reproduce \
  --typescript /tmp/lint-wave1-08-typescript
```

The runner writes each command's output to its own log, runs all gates even when
one fails, and exits nonzero for the complete upstream JSX blocker. Its scratch
preparation was reproduced successfully; its three proposed shared files match
the copies used in the observed runs. The following actual commands used
`-overlay=/tmp/lint-wave1-08-next/overlay.json`; lint commands ran with
ADAMIC_GATE_UNCACHED=1 where specified and the pinned ADAMIC_TYPESCRIPT_SOURCE.

| Command | Observed result |
| --- | --- |
| GOFLAGS=-overlay=... bash cloud/setup.sh | PASS; Go, clang, Node, submodules ready 0s each; cache warm 28s; total 28s; 5 processors |
| go run -overlay=... ./cmd/lint-registry | PASS; eight descriptors including all three assigned rules |
| go test -overlay=... ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1 -v -timeout=10m | PASS 49.824s; 8,438 equal bytes on Go, source Node, emitted JS, ASan/UBSan native |
| go test -overlay=... ./stage1/cohere/lint -run '^(TestRulesAgree|TestCompilerAndStage1Agree|TestWave08Shapes|TestWave08Throughput)$' -count=1 -v -timeout=20m | FAIL 197.775s overall; corpus and throughput PASS; full upstream and JSX shape FAIL |
| go test -overlay=... ./stage1/cohere/lint -run '^TestWave08UpstreamSupported$' -count=1 -v -timeout=10m | PASS 37.966s; module 13/13, parameter 117/117, Tailwind 53/54; 57,931 equal bytes |
| go test -overlay=... ./stage1/cohere/lint -run '^TestWave08Shapes$/^NonTSX$' -count=1 -v -timeout=10m | PASS 21.364s; twelve extra edge cases, 10,775 equal bytes |
| go test -overlay=... ./stage1/cohere/lint -run '^TestMutants$/(direction_escape_removed|module_declaration_suppressed|optional_parameter_ignored)$' -count=1 -v -timeout=10m | PASS 69.152s; three mutations caught on all three Adamic executions |
| go test -overlay=... ./stage1/cohere/lint/registry -count=1 -v | PASS 0.120s; deterministic generation and eleven invalid-descriptor mutants rejected |
| go vet -overlay=... ./... | PASS, exit 0, empty output |
| ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v -timeout=10m | PASS 15.047s; native and Node cache hits 0, misses 1 |

The compiler/stage1 corpus contains 77 compiler files and 137 stage1 .ts/.a
files, including gaps. All eight registered rules ran, yielding 12,630,367
identical output bytes. Upstream capture found 402 unique cases across those
eight rules. The complete gate fails when parsing the actual Tailwind JSX input
`export const c = <div className="flex ml-4" />;`: Node exits 70 with
`parser slice expected GreaterThanToken, got Identifier at 22`. A separately
run minimal JSX shape fails the same way at offset 15. The supported upstream
gate explicitly excludes and logs exactly that one source while preserving .tsx
extensions for the other 52 Tailwind TSX cases. No JSX parity is inferred from
the green subset. The earlier broad non-TSX-only subset log is retained as an
intermediate observation, superseded by this narrower exclusion.

Mutations change the real implementation, not the oracle. Each was compiled and
executed normally with exit zero and no stderr, so compiler diagnostics or a
sanitizer crash did not kill it. module_declaration_suppressed changes the
identifier match to modules, omitting the finding for `let a = 0, module = {};`.
optional_parameter_ignored removes the optional-parameter branch, omitting the
finding for `function f(a?: number, b: number) {}`. direction_escape_removed
removes the rtl/ltr escape, adding an rtl:ml-4 finding to
`const c = 'flex rtl:ml-4 mr-2';`. Source Node, emitted JS and sanitized native
each disagree with independent Go for each mutation. The core one-byte oracle
mutant also failed its intended comparison in the separate oracle test.

Throughput is whole-process findings per second, best of five rotating rounds.
Each rule gets the 77-file compiler corpus plus one file of 1,000 positive
examples, 78 files total. Count output matches Go in every round. Native here
is the release build; correctness executions above use ASan and UBSan. Times
include parser work, process startup and count formatting. These are not speeds
for the unmodified compiler corpus alone and show Go faster on this workload.

| Rule | Findings | Native findings/s | Node findings/s | Go findings/s |
| --- | ---: | ---: | ---: | ---: |
| structure/tailwind-no-physical-direction | 2,000 | 1,317.56 | 1,706.45 | 7,974.57 |
| @next/next/no-assign-module-variable | 1,000 | 651.46 | 891.02 | 4,093.28 |
| @typescript-eslint/default-param-last | 1,024 | 690.09 | 1,024.44 | 4,216.86 |

All raw logs are under
../rules/structure-tailwind-no-physical-direction/evidence/.
The exploratory CLI command `adamic build --sanitize` failed with usage exit 2
because that flag is unsupported; sanitizer verification succeeded through the
native.Build test harness instead. Raw patch context and diagnostic logs retain
original whitespace, so git diff --check flags those artifact lines; authored
rule sources and descriptors pass the whitespace check. Original claimed rules
22-24 remain unimplemented as reported in wave1-08-report.md. No PR was opened.
