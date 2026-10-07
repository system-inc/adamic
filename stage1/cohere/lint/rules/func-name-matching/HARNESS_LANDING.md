Built: rule branch rebased onto the landed area and deduplicated; five unique ports remain, all retained Adamic entry modules are .a and descriptors use AST kind names.
Commits: area d65a8f931c98655936ae04c6899f38f14862b73e includes main 39638d9e278d38bb5aeae887f46d55a70e47aaad; green helper branch d51df48c36a0d6a98103dd7cab594c382b2dc4e7; rule delivery follows this report.
Commands and outputs: registry PASS twenty descriptors, retained core fixture/mutant/options gate PASS 97.366s; Tailwind source Node/Go PASS 4.579s and 27,103 bytes across 52 rows; stable core corpus PASS 52.934s, 26,622,644 bytes across 734 pairs; seven-helper gate PASS 191.671s; setup PASS 92s, nproc 5.
Mutants: both retained core semantic mutants compiled and were caught on Node, emitted JavaScript and ASan/UBSan native; all twenty helper mutants re-greened. Tailwind backend mutants cannot compile after the required runtime regex migration.
Not covered: the rule branch is not green or landing-ready; runtime option RegExp lowering blocks all three Tailwind ports, and variant order also lacks shared design-system inputs. No new rule/helper claimed; stop at these unowned blockers.

The DEDUP_LEDGER.md on the pinned area was read before either rebase. Six losing
owned copies were removed entirely:

| Removed rule | Winning copy |
| --- | --- |
| @typescript-eslint/no-non-null-asserted-optional-chain | codex/lint-wave1-05 |
| @typescript-eslint/no-non-null-assertion | codex/lint-wave1-05 |
| @typescript-eslint/no-this-alias | codex/lint-wave1-07 |
| @eslint-community/eslint-comments/require-description | codex/lint-wave1-05 |
| @next/next/google-font-display | codex/lint-wave1-01 |
| structure/tailwind-no-physical-direction | codex/lint-wave1-05 |

The retained unique ports are consistent-this, func-name-matching,
better-tailwindcss/enforce-consistent-important-position,
better-tailwindcss/enforce-consistent-variable-syntax and
better-tailwindcss/enforce-consistent-variant-order. The ledger specifies source
batch branches for batch-only ports, but assigns none to wave1-04. No new batch
rule was reserved while the landing-first cap is blocked.

Small diagnostic-recording and Go-whitespace utilities formerly imported from
losing copies now live in the retained func-name-matching and important-position
directories. The private historical shared-contract fixture was retired with the
optional-chain copy. Both retained validators now invoke the actual landed
registry and copy the real area substrate; no shared test or driver is patched.
The core temporary graph registers only the two remaining core descriptors;
this is a subset certificate, not the shared production driver's full gate.
The Tailwind projected-AST test driver uses the landed RuleContext, RuleSet,
Linter and VolumeRules initialization. It remains explicitly a Go projection.

All five retained descriptors already declare typescript-go AST kind names.
Obsolete numeric listener probes and their numeric-contract report were removed.
Retained rule/modules were renamed to .a, imports and mutant file metadata were
updated, and ambiguous .ts facades were removed. Generated registries stay ignored.
No finding.ts, context.ts, main.ts, registry generator, shared oracle or lint_test
comparison file is changed relative to the area. Inherited compiler and shared
leak-check integrations were accepted in the rebase.

Regex migration follows the actual codex/lint-regex table row
 tailwind/class_literals.go:167, contract new RegExp(pattern, 'u').
The old default-pattern endsWith matcher and custom-pattern refusal are gone.
Surface compiles configured/default variable patterns once in its constructor
and calls RegExp.test. No hand-written fallback or runtime pattern translation
is retained. The pinned table's README and gaps.md were also inspected.

Observed native/emitted blocker, from the actual production main.ts module:

surface.a:41:132: stage 0 can't lower RegExp with a nonconstant pattern yet

The scratch program using programArguments()[0] as a pattern reproduces the same
lowering refusal. The shared lowerer is outside this unit's ownership. The
source-only comparison passes actual Go for all 45 upstream configurations plus
six witnesses, including explicit Go JSX projection and resolved variant facts.
That result is not native or emitted-JavaScript certification, nor independent
JSX parsing support. Arbitrary Go/JS option-pattern dialects are not certified.

The real shared TestOwnedWitnesses stops before backend compilation with:
better-tailwindcss/enforce-consistent-variant-order witness reports no findings.
That is the observation. The owned variant oracle and source test supply program
and resolved design-system facts explicitly; the shared default path does not
provide the same rule inputs. The old missing-provider refusal remains, so the
runtime pattern fix alone would not certify this rule through the shared driver.
This is not parking on a missing harness: that harness has landed. The branch
remains blocked and no further helpers are claimed under the work-in-progress cap.

Current-area commands, with source /workspace/adamic-tools/env.sh:

- go run ./cmd/lint-registry: twenty validated descriptors, including five owned.
- go test ./stage1/cohere/lint/rules/func-name-matching -run
  '^(TestRulesAndSuggestions|TestOwnedMutants|TestDecodedOptionCorners)$'
  -count=1 -v -timeout=10m: PASS 97.366s, two compiling mutants caught on all targets.
- ADAMIC_TAILWIND_PACKAGE=/tmp/wave104-tailwind/node_modules/tailwindcss
  go test ./stage1/cohere/lint/rules/tailwind-important-position
  -run '^TestSourceNodeParity$' -count=1 -v -timeout=10m: PASS 4.857s,
  26,626 identical bytes, fifty-one rows, source Node and Go only.
- go test ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1 -v
  -timeout=10m: FAIL 74.557s at the Go-side variant witness zero-count guard.
- go run ./cmd/adamic js stage1/cohere/lint/main.ts: FAIL at runtime regex lowering.

Initial owned validator runs exposed newly strict complete-directory discovery
and missing JSX witness transport; only the owned copier was fixed. A regex edit
left a stray loop fragment, which source Node and the checker refused; it was
fixed. The projected driver initially omitted VolumeRules initialization, which
Node refused; the owned initialization was updated. Every failed log is retained.

The first full core run passed fixture comparisons and both mutants, but its
734-pair source corpus comparison overlapped the owned Tailwind driver edit.
The serialized fixed source differs by the just-added VolumeRules lines; this
is input drift, not a backend miscompile claim. The complete tree was then frozen
and the corpus is rerun separately below. The full repository gate was not rerun. Current core corpus and positive witness throughput follow. Historical reports describe older branch snapshots and retired
copies; this report supersedes their current landing status.

Stable-tree corpus rerun: ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-04-typescript
go test ./stage1/cohere/lint/rules/func-name-matching -run '^TestCompilerAndStage1$'
-count=1 -v -timeout=10m PASS 52.934s. All 367 compiler/stage1 files, 734 pairs and
26,622,644 output bytes match Go on source Node, emitted JavaScript and sanitized
native. The core fixture/mutant/options comparisons from the final run also pass:
197 upstream configurations plus witnesses, 36,681 bytes; both compiling mutants
caught on all targets; decoded option corner 477 bytes. Owned vet passes.
The old core JSX probe had zero applicable rows and adds no coverage claim.

The added custom variable pattern ^classes$ on const classes='!flex' now also
matches actual Go on Node. Final source-only Tailwind gate PASS 4.579s, 52 rows,
27,103 bytes. Standalone runtime RegExp source Node prints true while the
production module is refused by the lowerer. No source matcher fallback exists.

Current-area throughput, best of three wall times including process startup:

| Rule | Positive findings / 1,000 witness rows | Native findings/s | Node findings/s | Go findings/s |
| --- | ---: | ---: | ---: | ---: |
| func-name-matching | 1,000 | 44,776.80 | 7,145.72 | 73,342.59 |
| consistent-this | 2,000 | 63,326.59 | 13,487.35 | 152,948.08 |

ADAMIC_LINT_BENCH=1 go test on the owned core package -run '^TestPositiveThroughput$'
-count=1 -v -timeout=10m PASS 9.139s. The positive fixtures are the same inputs
whose compiled semantic mutants were caught. Count mode agrees with actual Go
on every timing round. No emitted-JavaScript timing is requested or claimed.
The whole compiler/stage1 corpus has zero findings for both rules: all three
reported rates are therefore 0.00 findings/s. Native/Node/Go best times are
1.733329/0.985030/0.232961s for func-name-matching and
1.816452/1.159710/0.251686s for consistent-this. Those are separate workload
observations, not a whole-project performance extrapolation from witness rates.
The first benchmark build exposed omitted os/strings imports; those owned Go
imports were added before the successful run. Failed build logs are preserved.

Evidence: landed-core-corpus-stable.log, landed-core-final.log,
landed-tailwind-custom-node.log, landed-main-js3.log,
landed-shared-witnesses.log, landed-core-throughput.log,
landed-core-positive-throughput-final.log and landed-rules-vet-final.log.
Both branches include the pinned area's current-main merge. The helper branch
is green; this rule branch is explicitly blocked and cannot count as landing-ready
under the special shared-harness parking rule, because the harness has landed.
No push to main or any area branch is made. No additional claims are made.

Recheck after required-input instruction, 2026-10-07:
All origin heads were fetched explicitly. Main 39638d9e2 and area d65a8f931
are unchanged and already ancestors of both pushed worker branches. No rebase
is necessary. A stale ignored registry initially excluded the owned Tailwind
ports and let production compilation succeed; that result adds no evidence.
After go run ./cmd/lint-registry, production JavaScript compilation again
refuses surface.a:41:132, RegExp with a nonconstant pattern.

With ADAMIC_TAILWIND_PACKAGE=/tmp/wave104-tailwind/node_modules/tailwindcss and
ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-04-typescript, the owned filtered run
TestSourceNodeParity|TestRulesAndSuggestions finishes FAIL 7.298s: source Node
passes all 52 rows and 27,103 bytes in 4.49s, backend compilation refuses the
runtime pattern in 2.80s. An initial full owned run omitted those input variables
and fails its required-input guards; that is not green and is retained. No
correctness check is skipped, relaxed or deleted to obtain a green result.
The seventeen repository-wide required-input checks were not run; this unit
stops at its reproduced unowned lowerer blocker without claiming another helper.

Landing refresh onto area b84a9d9314b65d3d0261ee017e233287b4f071da,
which includes main c7991b900362796aefd111474e65eb5398e91953:
The ledger did not change. The owned rule rebase was conflict-free. With the
pinned TypeScript corpus supplied, TestRulesAndSuggestions, TestOwnedMutants,
TestCompilerAndStage1 and TestDecodedOptionCorners all pass. Core corpus remains
734 pairs and 26,622,644 identical bytes on Go, Node, emitted JavaScript and
ASan/UBSan native; both semantic mutants compile and are caught on all targets.
Owned rule vet passes. With the actual Tailwind package supplied, source Node
passes 52 rows and 27,103 bytes in 6.12s, but TestRulesAndSuggestions again
refuses surface.a:41:132, RegExp with a nonconstant pattern (FAIL 9.564s total).
The compiler lowerer is outside this unit. The rule branch is not green; no new
helper claimed, no shared-file changes, no required check relaxed or removed.
The complete repository gate and its seventeen correctness checks were not run.
See newarea-core.log, newarea-tailwind.log and newarea-vet.log.

Registry-only harness refresh onto area b46914832d70e00847d82d5d221ab7bb24040c53:
Main remains c7991b900. The ledger is unchanged and the rebase was clean.
The owned Tailwind projected driver drops VolumeRules and linter.rules, passes
the projected root to RuleContext, and passes RuleSet to the three-argument
Linter.walk. These are owned-driver adaptations; shared files are not edited.
The retained core fixture/mutant/options/corpus gate passes in 97.190s. Both
compiling core mutants are caught on Node, emitted JavaScript and sanitized
native. After the driver edit, the frozen corpus rerun passes in 55.075s:
414 files, 828 rule/file pairs, 26,599,098 identical bytes across all targets.
Both owned rule packages pass vet. The required-input Tailwind gate matches
52 source Node/actual-Go rows and 27,103 bytes, then fails runtime RegExp
lowering at surface.a:41:132 (31.216s total). No new claim or full gate.
Commands retain the pinned TypeScript and actual Tailwind input variables:
core -run '^(TestRulesAndSuggestions|TestOwnedMutants|TestDecodedOptionCorners|TestCompilerAndStage1)$';
final core -run '^TestCompilerAndStage1$'; Tailwind -run
'^(TestSourceNodeParity|TestRulesAndSuggestions)$'; each uses -count=1 -v -timeout=10m.
Evidence: b469-core.log, b469-core-final-corpus.log, b469-tailwind.log,
b469-rule-vet.log. No required check is removed, relaxed or skipped to green.

Current-main refresh onto area d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898,
including main b6b1538b0cebc4ba6741ac34f1aedb60293c1d06 (typeof null fix):
Ledger unchanged; both rebases clean; no owned production source changes.
The required-input core fixture/mutant/corpus/options gate passes in 103.566s,
with 828 pairs and 26,599,098 identical bytes on Go, source Node, emitted
JavaScript and ASan/UBSan native. Both compiling semantic mutants are caught
on all targets. Owned rule vet passes. Actual Tailwind source parity passes
52 rows and 27,103 bytes in 6.56s, then backend comparison refuses dynamic
RegExp lowering at surface.a:41:132 (FAIL 10.096s overall). No new claims.
Commands unchanged from the preceding refresh: core four filtered correctness
tests and Tailwind two filtered correctness tests, -count=1 -v -timeout=10m,
with ADAMIC_TYPESCRIPT_SOURCE and ADAMIC_TAILWIND_PACKAGE supplied. Evidence
d3-core.log, d3-tailwind.log, d3-rule-vet.log. Full gate and seventeen
repository-wide correctness checks not run; no check relaxed, removed or skipped.

Landing split requested by Ahra: the three blocked Tailwind descriptor directories
are removed; their exact implementation and reproducer remain named in
claims/wave1-04-parked.md at the preserved pre-split commit. The current landing
branch retains consistent-this and func-name-matching and is GREEN on the real
unified harness at area d3a37422c. This supersedes the earlier blocked status.
go run ./cmd/lint-registry passes; owned gofmt -l output empty; go vet
./stage1/cohere/lint passes. ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-04-typescript
go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestMutants|TestRulesAgree)$'
-count=1 -v -timeout=30m PASS 875.988s. TestRulesAgree PASS 72.79s; complete
TestMutants PASS (all descriptors); TestOwnedWitnesses PASS 21.83s with 93,748
identical bytes on Go, Node, emitted JavaScript and sanitized native. Existing
shared recovery limitations remain explicit in the log and are not changed.
No shared harness edits. Full repository gate was not run. Evidence:
unified-landing.log, unified-landing-vet.log, landing-gofmt.log, landing-registry.log.
