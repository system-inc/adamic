Built: two .a helpers, addWhitespaceAroundMathOperators and unescapeCSSIdentifier; twelve dependency entries removed across six Tailwind rules, zero final blockers alone.
Commits: claims 89afa076 and a70d3b5b pushed before code; math delivery 9a5ef4c2; final delivery SHA is in the response.
Checks: frozen final owned-package gate PASS 31.029s, 1,165,857 inputs, 6,482,634 identical bytes per Node/emitted JavaScript/ASan+UBSan native; vet clean; setup 76s, nproc 5.
Mutants: comma_spacing_disabled and scalar_one_replaced compile and exit successfully with empty stderr on all three runtimes, then fail only comparison against real Go.
Limits: full upstream consumer package FAIL 3.650s on absent hard-coded Ahra corpus/Tailwind installation; valid Unicode inputs and source-derived helper controls, no end-to-end rule findings or full repository gate.

The frozen readiness list assigns this helper to:

- better-tailwindcss/enforce-canonical-classes
- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-consistent-variant-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

This removes six dependency entries. Each consumer retains other blockers, so zero rules become fully helper-ready from this helper alone. The shared readiness.json is left untouched.

Selection inspected all 397 fetched origin refs and twelve distinct helper claims, including short names in claim prose. The already delivered comments bundle is reserved separately in HELPERS.md. All higher concrete-symbol fan-outs were reserved; this helper ties the highest unclaimed count at six.

The corpus generator uses Go's parser and strconv.Unquote to extract 959 actual string literals from all six consuming test files. Original literals and calc-wrapped variants are compared, plus Cartesian controls across every math function, unknown/uppercase function names, operators, units, signs, exponents, nested functions, non-ASCII spaces and supplementary Unicode. These are source-derived helper inputs, not a measurement of runtime helper calls or a replay of final rule findings. The Go oracle imports the unchanged private implementation through an added export in a scratch overlay; no Go implementation is copied or rewritten.

The first differential run failed: Go's string(byte) expands non-ASCII bytes rather than copying their UTF-8 encoding. The fixed Adamic helper reads utf8At and reconstructs the same byte-valued rune stream only after Go's early math-name bail. For example a scanner-path emoji expands to four Latin-1 code points. This behavior is observed in Go and is intentionally preserved. The failed run and final successful run are retained separately; mutant results before a passing baseline are not credited.

Commands, each with output sent directly to the linked evidence logs:

```
ADAMIC_TOOLS=/workspace/adamic-tools bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/lint/helpers/from-wave1-10/testdata/generate.py
go test -count=1 -v -timeout 15m ./stage1/cohere/lint/helpers/from-wave1-10
go vet ./stage1/cohere/lint/helpers/from-wave1-10
```

See [final comparison](evidence/math-tests-final.log.txt), [first mismatch](evidence/math-tests-first.log.txt), [setup](evidence/setup.log.txt), [generation](evidence/generate.log.txt) and [vet](evidence/vet.log.txt).

Second helper: css_identifier_unescape.a ports unescapeCSSIdentifier, claimed and pushed at a70d3b5b before code. It serves the same six rules, removing six further dependency entries and still zero final blockers. Every scalar through 0x110000, including NUL, surrogates and out-of-range replacement, is compared against unchanged Go, plus six-digit bounds, five exact CSS terminators, literal supplementary-rune escapes and final backslashes. There are 1,116,695 cases and 5,608,876 identical output bytes per runtime. The scalar_one_replaced mutant maps valid U+0001 to U+FFFD; every runtime compiles and exits successfully, then fails only comparison.

The complete two-helper gate passed in 31.933s. Go vet passes for the owned package. The exhaustive scalar corpus is compressed and the test expands it only in scratch. Regenerate with testdata/generate_unescape.py, then run the same owned-package command above. See [two-helper comparison](evidence/both-tests.log.txt), [unescape generation](evidence/unescape-generate.log.txt) and [both-helper vet](evidence/both-vet.log.txt).

Consumer-rule validation is blocked: the full original Go Tailwind package exits FAIL in 3.650s. TestUnknownClassFixturesActuallyRan, TestCanonicalFixturesActuallyRan, TestConflictFixturesActuallyRan and TestClassOrderFixturesActuallyRan explicitly fail because no installed tailwindcss is found from /Users/kirkouimet/Projects/ahra/app/_theme/styles. Other tests skip the absent /Users/kirkouimet/Projects/ahra/app/_theme/styles/theme.css repository. TestClassOrderLiveMatchesTheEngineOverTheCorpus also fails. The bounded helper corpus uses all six real test files but cannot replace those missing repositories and installations or prove their final findings. See [full consumer package](evidence/upstream-all.log.txt). No shared file was changed to bypass this failure. Further claims stop at this validation blocker.

Observed dependency removals: twelve entries across six distinct rules. Inference: rule ports can reuse these two transformations once their other shared dependencies arrive; this is not a claim that six complete rules are unblocked.

Final frozen sources rerun: PASS 31.029s. See [final two-helper gate](evidence/final.log.txt). Final all-origin refresh inspected seventeen distinct claims; neither helper has a competing reservation. Prior rule branch remains pushed through e9066bff; its shared JSX, multi-edit and general runtime-regex boundaries remain documented on that branch.

Landing readiness, October 7

The helper branch rebased cleanly onto origin/main e8ba3d5d81de4d3773c723914fccd4c76248b965. The first rebase onto e011f8f6 passed its oracle in 94.345s; main advanced during that run, so the final rebase was verified again. Final command: ADAMIC_GATE_UNCACHED=1 go test -count=1 -v -timeout 15m ./stage1/cohere/lint/helpers/from-wave1-10. PASS 29.258s: all 1,165,857 inputs, both compiling output-only mutants, and 6,482,634 identical output bytes per source Node, emitted JavaScript and ASan/UBSan native. Final focused go vet exits 0. Setup completed in 128s on 5 processors; Go/clang/Node/submodules each ready in 0s, build-cache warm 128s. Full repository tests and the externally blocked consuming-rule gate were not rerun. See evidence/landing-current-oracle.log.txt, evidence/landing-current-vet.log.txt and evidence/landing-setup.log.txt.

The other pushed branch, codex/lint-wave1-10, is not landing-ready. Rebase onto both fetched main versions stops while applying the inherited registration commit 29175443 in four shared files: stage1/cohere/lint/README.md, lint.ts, lint_test.go and testdata/oracle.go. The requested ownership restriction reserves shared registration/harness work for its worker. Both attempts were aborted without changing those files; the rule branch remains at its previously pushed e9066bff. See evidence/landing-rule-rebase.log.txt and evidence/landing-rule-conflicts.log.txt. Resolving the registration foundation on main is required before this unit can re-green that branch. No new helper or rule claims were made under the work-in-progress cap.

Landing refresh: helper branch rebased onto main f8013f0baac41ddc340d76f83bddde38536a8f07. ADAMIC_GATE_UNCACHED=1 go test -count=1 -v -timeout 15m ./stage1/cohere/lint/helpers/from-wave1-10 passes in 39.234s, with all 1,165,857 cases and both compiling output-only mutants on source Node, emitted JavaScript and ASan/UBSan native. Focused go vet exits 0. See evidence/landing-f8013f0b-oracle.log.txt and evidence/landing-f8013f0b-vet.log.txt. Rule branch now remains at pushed 641b887a, including the independently checked twelve numeric listener declarations. Its rebase onto f8013f0b still conflicts at inherited registration commit 29175443 in the same four shared lint files and was aborted without edits. No new claims or main/area pushes were made; runtime listener node delivery and shared Diagnostic integration remain pending.

Parking refresh: rebased onto main c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06. Owned helper oracle PASS 33.362s, all 1,165,857 inputs and both compiling output-only mutants across source Node, emitted JavaScript and sanitized native; focused vet passes. The rule branch is being parked under the explicitly authorized #zmh9v36 exception, with only owned commits retained and the frozen compatibility oracle named on it. The historical Tailwind external-corpus limitation remains; it does not block a separately validated regexp helper. No new claim has been made before both branches satisfy landing readiness. See evidence/parking-c01907a7-oracle.log.txt and evidence/parking-c01907a7-vet.log.txt.

Third helper: regexp.modifierGroup is delivered in regexp_modifier_group.a. Four frozen consumer edges, zero final blockers. The claim efb00d444 preceded code. Focused differential gate PASS 13.416s over 1,066,656 queries and 6,799,954 identical bytes per runtime; duplicate_guard_disabled compiles and is caught only by Go output comparison on all three runtimes. All four original Go consumer families pass, and focused vet/diff checks pass. See MODIFIER.md and evidence/modifier-*.log.txt for exact domains, fixtures, commands and first preparation failure.

Fourth helper: regexp.simpleFold is delivered in regexp_simple_fold.a, including its pinned Go standard-library transition data. Claim 005c910cb preceded code. Four frozen consumer edges, zero final blockers. Focused differential gate PASS 10.656s over 1,195,698 queries and 8,047,532 identical bytes per runtime; maximum_fold_selected compiles and is caught only by private-Go byte comparison on all three runtimes. Vet passes. See FOLD.md and evidence/fold-*.log.txt for table provenance, exact domains and first driver compilation failure. These two new regexp helpers add eight prerequisite edges to the prior twelve Tailwind edges, across ten distinct rules; no complete rule-readiness claim is inferred.

Final complete owned-package gate: go test -count=1 -v -timeout 15m ./stage1/cohere/lint/helpers/from-wave1-10 PASS 55.021s, all four helpers and four compiling semantic mutants, 3,428,211 queries and 21,330,120 identical bytes per source Node/emitted JavaScript/ASan+UBSan native. See evidence/four-helpers-final.log.txt. Final all-origin refresh still places main at c01907a7, an ancestor of both owned pushed branches; modifierGroup and simpleFold have no competing helper claim file. Rule branch remains explicitly parked at 551c529c8 for #zmh9v36. All claimed work is committed and pushed only to the two owned branch names.

Landing refresh on current main b8fb957aa839a9e8cb0b54279dd9864fa317bd30: clean rebase, complete owned-package go test -count=1 -v -timeout 15m ./stage1/cohere/lint/helpers/from-wave1-10 PASS 55.364s. All four helpers, 3,428,211 queries, 21,330,120 identical bytes per source Node/emitted JavaScript/sanitized native, and comma_spacing_disabled, scalar_one_replaced, duplicate_guard_disabled and maximum_fold_selected compile and are caught only by byte comparison on every runtime. Focused vet passes. Setup ready steps each 0s, build-cache warm and total 83s, nproc 5. See evidence/landing-b8fb957a-*.log.txt. No new claim: the separately rebased rule branch still cannot pass full native/emitted validation because runtime RegExp lowering and Go/JS option dialect remain blocked. ab70f38d4 is still outside main. No shared harness, compiler or leak-check implementation was edited.
