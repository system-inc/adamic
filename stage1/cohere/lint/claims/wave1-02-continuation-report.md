Built: two `.a` rule candidates and a three-runtime JSX blocker probe; the requested three complete ports are not delivered.\
Commits: claim `782e031e`, Tailwind `61ebcccb`, descriptions `240626b9`, font blocker `35d5b75c`; branch `codex/lint-wave1-02`.\
Checks: isolated Go overlay compares 200 compiler/stage1 files (12,542,232 identical bytes), 160/162 upstream cases and owned witnesses on Node, emitted JavaScript and ASan/UBSan native.\
Mutants: omitted rtl/ltr exemption and accepted empty description compile, run cleanly and fail only output comparison on all three runtimes.\
Uncovered: shared `.a` registration, two upstream findings-plus-fixes cases, the JSX font rule and its semantic mutant, malformed option rejection, exhaustive fuzzing and the full repository gate.

Existing work was pushed before fetching all origin heads. Main was `ef3d907e`. The first 45 helper-ready entries appeared in origin claims; the last helper entry was available. The next two candidates came from the syntax-only inventory on `origin/codex/lint-inventory`. The three new reservations were pushed in `782e031e` before implementation: `structure/tailwind-no-physical-direction`, `@eslint-community/eslint-comments/require-description`, and `@next/next/google-font-display`. The earlier no-restricted-types reservation remains explicitly blocked.

The Tailwind candidate follows the Go rule's file gate, class-shape heuristic, Unicode field splitting, ordered physical-to-logical mapping, negative and rtl/ltr exemptions, cooked literal and template-chunk traversal, exact message and no-fix behavior. The comment candidate follows the Go rule's directive recognition, decoded ignore/additional-directive settings, separator/reason logic, full-comment range and exact no-fix message. Its collector protects parser-provided literal ranges and terminates line comments at ECMAScript boundaries. The real compiler comparison caught the initial CRLF range error; a lexical witness confirms bare CR and regex protection. Raw failed comparisons are kept alongside the passing correction.

Shared repository files were not edited. The current registry requires `.ts` modules and rejects `.a` mutants. An isolated, reproducible scratch overlay enables `.a` discovery/copying, emitted-JavaScript comparison, original upstream filename extensions, and TSX oracle parsing. Its original compatibility changes came from the owned overlay on `origin/codex/lint-wave1-12`. The owned preparation script copies and patches only temporary Go files. It does not apply the patch to the repository. Therefore ordinary lint tests and setup still fail without this compatibility overlay. These candidates need shared foundation integration before merging.

The source corpus is TypeScript v6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`, plus every `.ts`/`.a` file under this branch's stage1 tree: 200 total, zero parser exclusions. The output comparison includes reports, messages, byte ranges, repairs/suggestions and final source. Neither new Go rule offers a fix, and their unchanged final-source bytes are checked. Sanitized native is used for parity and mutants; optimized native is used for throughput.

Upstream capture ran the unmodified Go rule tests and collected 379 unique cases across registered rules. Of the 162 cases for the two candidates, 160 were compared successfully (49,808 bytes after the collector correction). One Tailwind JSX literal is refused by Adamic's parser. One comment fixture is intentionally unterminated: Go cohere's findings test accepts it, but Go's fix engine refuses it before any fix. A direct attempt to include that recovery input confirms the fix refusal in `evidence/upstream-final.log`; it is not silently counted as passing. The final coverage helper explicitly records both exclusions. The Node coverage classifier uses a temporary throwing panic only to enumerate refusals; all output comparisons use the unmodified Adamic runtime.

The font rule is not implemented or registered. Its first real failing JSX fixture is stored in `.a`. `TestWave02JSXBlocker` proves source Node, emitted JavaScript and sanitized native all exit 70 with the exact parser refusal at offset 32. All 16 Go fixture rows pass. No zero-findings placeholder, semantic mutant, or throughput number is claimed for this rule. Shared JSX parsing and attribute decoding are outside this unit's ownership under `docs/parallel-work.md`, read on `origin/codex/no-shared-lists`.

Measured throughput, one end-to-end run per cell including startup, reading and parsing:

| Corpus and rule | Findings | Native findings/s | Node findings/s | Go findings/s |
| --- | ---: | ---: | ---: | ---: |
| 200 compiler/stage1 files, Tailwind | 0 | 0.00 | 0.00 | 0.00 |
| 200 compiler/stage1 files, descriptions | 125 | 53.70 | 108.43 | 397.16 |
| 1,000 repeated owned witnesses, Tailwind | 2,000 | 101,809.71 | 14,332.57 | 93,732.91 |
| 1,000 repeated owned lexical witnesses, descriptions | 2,000 | 71,455.32 | 13,303.47 | 67,608.26 |

The real Tailwind pass took native 1.368s, Node 0.858s and Go 0.238s. Its zero findings do not establish positive findings throughput. Witness measurements are separate synthetic evidence, not representative compiler performance. No extrapolation to the blocked font rule is made.

Toolchain setup initially failed at cache warming because registration tried to open `rule.ts`. With the scratch overlay, setup passed: Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s, build cache warm 14s, done 14s. `nproc` is 5; cgroup quota is 4 CPUs. Tools are Go 1.27.1, clang 20.1.8 and Node 24.19.0, sourced from `/workspace/adamic-tools/env.sh`. A full-history compiler clone was interrupted and replaced by a successful shallow fetch of the exact corpus commit. An initial standalone native probe lacked the runtime header; the final native probe is built through `native.Build` with sanitizers and passes its expected-refusal check.

Commands and final outcomes (all test output went directly to log files):

```sh
python3 stage1/cohere/lint/rules/structure-tailwind-no-physical-direction/prepare-validation.py /tmp/wave02-validation
source /workspace/adamic-tools/env.sh
export GOFLAGS=-overlay=/tmp/wave02-validation/overlay.json
export ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-02-typescript
# PASS: witnesses, including corrected lexical case, 11,066 identical bytes, 14.41s.
go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1 -v -timeout 10m > /tmp/witness.log 2>&1
# PASS: 160 accepted cases, two explicit exclusions, 49,808 identical bytes, 17.19s.
go test ./stage1/cohere/lint -run '^TestWave02UpstreamSubset$' -count=1 -v -timeout 10m > /tmp/upstream.log 2>&1
# PASS: 200 files, 12,542,232 identical bytes, 38.466s.
go test ./stage1/cohere/lint -run '^TestWave02CompilerSubset$' -count=1 -v -timeout 15m > /tmp/compiler.log 2>&1
# PASS: compiling semantic mutants caught by output comparison on all three runtimes, 26.622s.
go test ./stage1/cohere/lint -run '^TestMutants/(empty_directive_reason_accepted|physical_direction_exemption_omitted)$' -count=1 -v -timeout 10m > /tmp/mutants.log 2>&1
# PASS: matched refusal on all three runtimes, 15.871s.
go test ./stage1/cohere/lint -run '^TestWave02JSXBlocker$' -count=1 -v -timeout 10m > /tmp/jsx.log 2>&1
# PASS: source benchmark 11.536s; final witness benchmark 5.967s.
go test ./stage1/cohere/lint -run '^TestWave02SourceThroughput$' -count=1 -v -timeout 15m > /tmp/source-throughput.log 2>&1
go test ./stage1/cohere/lint -run '^TestWave02Throughput$' -count=1 -v -timeout 10m > /tmp/witness-throughput.log 2>&1
# PASS: deterministic generation and 11 descriptor rejection mutants, 0.012s.
go test ./stage1/cohere/lint/registry -count=1 -v > /tmp/registry.log 2>&1
# PASS: filtered input oracle, six program cases plus input cache test, 1.224s.
go test ./internal/oracle -run Input -count=1 -v -timeout 10m > /tmp/oracle.log 2>&1
# PASS: warm compilation only, not a full test gate.
ADAMIC_TOOLS=/workspace/adamic-tools bash cloud/setup.sh > /tmp/setup.log 2>&1
# In cohere: PASS, 16 font-rule fixtures, 0.005s.
go test ./internal/lint/rules/next -run '^TestGoogleFontDisplay' -count=1 -v -timeout 10m > /tmp/font-go.log 2>&1
```

Actual invocation logs, including setup failures, cold dependency notices, broad test failures, a missing corpus environment variable, corrected comparisons and final mutants, are committed under `stage1/cohere/lint/rules/structure-tailwind-no-physical-direction/evidence/`. `fixed.log` contains the successful 160-case upstream and witness comparisons plus the separate missing-environment failure; `compiler-fixed.log`, `mutants-final.log`, `font-blocker.log`, `final-mutants-bench.log` (source throughput only), and `throughput-final.log` contain final passing evidence. `reproduce.log` confirms the checked-in preparation script builds a working overlay. Raw logs, fixture bytes and patch context preserve intentional whitespace.

The full repository runtime gate was not run. Warm compilation, touched registry tests, bounded lint comparisons and the filtered input oracle were run. No shared compiler/parser/native code was changed. No new Adamic `.ts` files, pull request, or claim of three complete ports is provided.
