Built: wave 08 rebased onto lint area b46914832; ten completed profiles and two parked kernels pass again.
Commits: tested rebased tip 37dfa0e3a7820136343790c4fcbf1ca16dffa3fb; this report accompanies the evidence commit.
Checks: all owned comparisons and sanitizers, shared registry and suggestion serialization, and filtered Node oracle pass.
Mutants: all ten completed-rule mutants caught by Go bytes; kernel mutants and released-handle probes pass separately.
Uncovered: shared typed registration, two HIR rules, new quiet timings and the full required-input gate.

Rebased cleanly from published 6fd96e1fa21d8de8464b4f784738f832eaa7dc59 onto
origin/area/stage1-lint b46914832d70e00847d82d5d221ab7bb24040c53, containing
unchanged main c7991b900362796aefd111474e65eb5398e91953. The area changes migrate
legacy syntax rules into the registry and update their shared context and harness.
There are no changes under cmd, internal, bridge or stage1/cohere/typeaware in
that base delta. The owned profiles use their own contexts. Accordingly the
unchanged b84 stage-0, core native executables and normal/sanitized checker
archives were reused; the original, process, race, Globals and mutant validators
build their own executables. Shared changes were inherited without edits.

The first core comparison invocation omitted sourcing the tool environment and
failed before comparisons with `Go: Unknown option: build`. It was corrected by
sourcing /workspace/adamic-tools/env.sh and rerunning normally and sanitized.
All final validation sessions exited zero. Output is retained in validation-b469,
including complete base64 log/stdout/stderr/manifest streams in outputs.json.gz.

Commands run, with all output redirected to /workspace/wave08-b469:

- go test ./stage1/cohere/typeaware -run '^TestWave08' -count=1 -v -timeout 30m,
  with original artifact variables, TypeScript corpus root and both frozen
  manifests: PASS, 343.730 seconds. Controls, compiler 77 and repository 287
  agree normally and sanitized. The loop, require-import and cycle mutants
  compile, exit zero with empty stderr, and differ only under byte comparison
  at bytes 51, 7383 and 12196. Checker declaration, type-syntax and resolved
  module mutants are caught by direct checker comparisons.
- compare_core.py, using the captured upstream fixtures and compiler root,
  normal and ASan/UBSan: 264/264 complete streams match in each run.
  validate_mutants.py compiles atomic outdated-set, await semicolon and Symbol
  name-filter mutants: 72, 99 and 20 cases differ, first bytes 119, 715 and 147.
  All exit zero with empty stderr. Four raw-question released-handle probes
  refuse before output with exit 70.
- validate_process.py: 100 upstream programs and both frozen corpora match
  normally and sanitized. Callee and blocking-order mutants compile, exit zero
  with empty stderr, and differ in six and nineteen programs. Released callee,
  program-load and symbol-ancestry handles refuse with exit 70.
- wave08-next/validate.py: 22 controls / 13 findings / 8174 bytes and both corpora
  match normally and sanitized. Timeout mutant compiles, exits zero with empty
  stderr, and differs at byte 46. All exercised released-question modes refuse
  with exit 70 before output.
- wave08-react/validate.py: 111 typed upstream programs / 58 findings, 35 controls
  / 34 findings / 22895 bytes, JSX and both corpora match normally and sanitized.
  Ancestry mutant compiles, exits zero with empty stderr, and differs at byte
  9180 in controls and in seven upstream programs. A separately run sanitized
  node-symbol-details released-handle probe refuses with exit 70 before output.
- wave08-react/validate_cores.py: parked native kernels match Go under sanitizers;
  join and dependency-count mutants compile and differ under byte comparison.
  Both production entries refuse with exit 70. These are kernel checks, not
  complete production-rule mutants.
- go test ./stage1/cohere/lint/registry -count=1 -v: PASS, 0.474 seconds, including
  descriptor rejection mutants and .a module support.
- go test ./stage1/cohere/lint -run
  '^(TestDotARename|TestCompleteSuggestionSerialization|TestSuggestionAlongsideAutomaticFix)$'
  -count=1 -v -timeout 30m: PASS, 229.213 seconds. Go, source Node, emitted
  JavaScript and native match completely; rename preserves 646 bytes and the
  automatic-fix plus three unapplied-suggestions case preserves 721 bytes.
- go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v:
  PASS, 0.222 seconds. Native and Node observation cache hits are reported.

No selected Go check skipped. The full gate, including the 17 externally
provisioned correctness checks, was not run or claimed green. No test was
relaxed or deleted. Shared typed integration still needs a checker-program
handle and configured root list in RuleContext plus production factory/driver
registration; owned scratch overlays do not certify that integration. React
immutability and no-deriving-state-in-effects remain parked for native HIR/SSA/
capture analysis on #dnv6f2c.

No new quiet measurements were run because the compiler, bridge and profile
implementations are unchanged. The preceding LANDING_B84_REPORT.md records
quiet native/Go medians of 2.202x to 21.074x, native slower. Current validators
also record timing observations in their JSON; overlapping jobs mean those
observations are not quiet performance comparisons. Toolchain setup remains
the earlier 132-second setup; nproc is 5, quota four cores.

The all-origin audit covers 639 refs and 33 distinct claim Markdown blobs:
25 verified ports and 172 claims cover all 197 ranked rules, eligible list empty.
No new rules were reserved. Only owned claim/evidence files change, and only
codex/typeaware-wave-08 is pushed with an explicit lease on its prior published
SHA. Main and area branches are not push targets.
