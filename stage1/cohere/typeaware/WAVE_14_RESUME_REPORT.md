Rebased the existing wave-14 branch onto current main; no new rules claimed.
Tested source head 3605a6fe on main c01907a7, replacing published 3da4b417b.
Owned Go byte oracles PASS 426.726 s; bridge, uncached filtered Node and vet pass.
Twenty semantic byte mutants, nine metadata mutants and released-handle mutants caught.
Regex runtime/error compatibility, JSX, undefined labels and numeric dispatch remain incomplete.

The landing-first cap makes the rebase and refreshed gate this unit. Rebase was
clean. Integration's changes were retained; no shared files were edited or
reverted. This worker has pushed only codex/typeaware-wave-14. Before publishing,
remote main was still c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06 and the remote
unit branch still 3da4b417bc6cd2c9df02e045bba4b3340974e5d8. The exact lease
against that unit SHA protects others' updates while publishing the explicitly
requested rebase. No push targets main or any area branch.

Setup succeeded: Go, clang, Node and submodules ready at 0 s, cache warm and
done at 40 s. nproc 5, cgroup cpu.max 400000 100000, memory 17.6 GB.
Go 1.27.1, clang 20.1.8, Node 24.19.0. Commands source
/workspace/adamic-tools/env.sh; every test writes directly to retained logs.

    bash cloud/setup.sh
    go test ./stage1/cohere/typeaware -run '^TestWave14' -count=1 -v -timeout=30m
    ADAMIC_TSGO_CORPUS=/workspace/wave-14-typescript go test ./bridge/tsgo/... -count=1 -v -timeout=15m
    ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' -count=1 -v -timeout=10m
    go vet ./...

The owned invocation uses ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-14-typescript,
ADAMIC_WAVE14_COMPILER_MANIFEST and ADAMIC_WAVE14_NEXT_COMPILER_MANIFEST pointing
to /workspace/wave-14-artifacts/compiler.manifest, and both REPOSITORY_MANIFEST
variables pointing to /workspace/wave-14-artifacts/repository.manifest.
ADAMIC_WAVE14_ARTIFACTS, NEXT_ARTIFACTS, RENDER_ARTIFACTS and THIRD_ARTIFACTS
point to /workspace/wave-14-resume-{original,next,render,third} respectively.
The frozen populations remain 77 compiler roots and 287 repository roots.
No inputs were filtered to evade findings. Complete suite streams are retained
compressed under validation-wave-14-resume, alongside logs and timing rounds.

| Suite | Result |
| --- | --- |
| Numeric declarations | PASS 0.01 s |
| Continuation | PASS 72.93 s |
| Render judgments | PASS 50.60 s |
| Original three rules | PASS 70.08 s |
| Third batch | PASS 233.11 s |
| Complete owned invocation | PASS 426.726 s |
| Bridge | PASS 76.247 s |
| Checker package | PASS 0.408 s |
| Filtered uncached Node | PASS 1.797 s |

Normal and ASan/UBSan/LeakSanitizer builds compare complete findings, fixes and
suggestions against unchanged production Go. Supported native stderr is empty.
Controls retain 33 original findings, 15 continuation findings, 18 synthetic
render findings, 60 third-batch findings, 58 constructor findings and 36 new
pattern findings. Both frozen corpora agree in every suite. The independent
741-constructor ASCII/range matrix also agrees under all three builds: 168
findings, 24,997 bytes, SHA256
0ca4e3e715a2f617412d56fe6650e2dc63861b7df3f02112dec8e1c13a7474c7.
Its input remains in validation-wave-14-patterns; fresh streams are retained here.

The following exact observations identify every owned mutant and its catcher.
Semantic mutants compile, exit 0 with empty stderr and differ only at full byte
comparison. Declaration mutants are metadata assertions, and retained released
handles are caught separately by the required panic assertion.

```text
    wave_14_listeners_test.go:101: Go listener kinds [221] agree; wrong-kind declaration mutant caught
    wave_14_listeners_test.go:101: Go listener kinds [227, 214, 229] agree; wrong-kind declaration mutant caught
    wave_14_listeners_test.go:101: Go listener kinds [264, 232] agree; wrong-kind declaration mutant caught
    wave_14_listeners_test.go:101: Go listener kinds [214] agree; wrong-kind declaration mutant caught
    wave_14_listeners_test.go:101: Go listener kinds [214] agree; wrong-kind declaration mutant caught
    wave_14_listeners_test.go:101: Go listener kinds [295] agree; wrong-kind declaration mutant caught
    wave_14_listeners_test.go:101: Go listener kinds [257] agree; wrong-kind declaration mutant caught
    wave_14_listeners_test.go:101: Go listener kinds [214, 215] agree; wrong-kind declaration mutant caught
    wave_14_listeners_test.go:101: Go listener kinds [307, 13] agree; wrong-kind declaration mutant caught
    wave_14_next_test.go:121: listener mutant: exit 0, empty stderr, byte oracle catches byte 93
    wave_14_next_test.go:121: namespace mutant: exit 0, empty stderr, byte oracle catches byte 4629
    wave_14_next_test.go:167: released-registry mutant exits 0, required panic catches it
    wave_14_render_test.go:90: render mutant exits 0 with empty stderr; byte comparison catches byte 52
    wave_14_test.go:116: delete mutant: exit 0, empty stderr, byte oracle catches byte 54
    wave_14_test.go:116: stringify mutant: exit 0, empty stderr, byte oracle catches byte 1531
    wave_14_test.go:116: class mutant: exit 0, empty stderr, byte oracle catches byte 4334
    wave_14_test.go:145: released-registry mutant exits 0, required panic catches it
    wave_14_third_test.go:201: cooked-mapping mutant exits 0 with empty stderr; byte oracle catches byte 56
    wave_14_third_test.go:201: constant-write mutant exits 0 with empty stderr; byte oracle catches byte 8665
    wave_14_third_test.go:201: call-flags mutant exits 0 with empty stderr; byte oracle catches byte 457
    wave_14_third_test.go:201: call-literal mutant exits 0 with empty stderr; byte oracle catches byte 10211
    wave_14_third_test.go:201: constant-dedup mutant exits 0 with empty stderr; byte oracle catches byte 6660
    wave_14_third_test.go:201: tracker-alias mutant exits 0 with empty stderr; byte oracle catches byte 12706
    wave_14_third_test.go:201: tracker-write mutant exits 0 with empty stderr; byte oracle catches byte 21832
    wave_14_third_test.go:218: label mutant exits 0 with empty stderr; full byte oracle catches byte 51
    wave_14_third_test.go:218: flags mutant exits 0 with empty stderr; full byte oracle catches byte 5098
    wave_14_third_test.go:218: unicode-quote mutant exits 0 with empty stderr; full byte oracle catches byte 5692
    wave_14_third_test.go:218: surrogate-decoding mutant exits 0 with empty stderr; full byte oracle catches byte 17957
    wave_14_third_test.go:218: class mutant exits 0 with empty stderr; full byte oracle catches byte 6839
    wave_14_third_test.go:228: scope meaning mutant exits 0 with empty stderr; full byte oracle catches byte 1344
    wave_14_third_test.go:263: pattern range mutant exits 0 with empty stderr; byte oracle catches byte 4331
```

The bridge gate repeats its ABI/type/link/ownership mutants; its retained log
reports length failures caught by ASan, wrong-type bytes by comparison, stale
handles by the released assertion, link opt-in by refusal, and missing frees/
unowned regions by LeakSanitizer. The filtered Node gate repeats the one-byte
oracle mutant and has no cache hits. Go vet produces no diagnostics.

Quiet alternating three-round whole-process count medians after all tests:

| Population | Native seconds | Go seconds | Native / Go |
| --- | ---: | ---: | ---: |
| compiler | 1.735851 | 0.329713 | 5.26x |
| repository | 0.239765 | 0.127825 | 1.88x |
| controls | 0.021808 | 0.029908 | 0.73x |
| constructors | 0.023935 | 0.032865 | 0.73x |
| upstream | 0.023395 | 0.033734 | 0.69x |
| patterns | 0.027405 | 0.037309 | 0.73x |

Compilation and sanitizers are excluded. Constructor and upstream timing rows
use --class-only in both programs. The upstream row checks its count of 89;
its previous full byte evidence remains historical, not a newly repeated full
comparison. Numeric declarations are still ignored by shared dispatch, so no
kind-indexed speed gain is claimed.

The shared regex table is now available at origin/codex/lint-regex,
071fb012848ce0408428c61aba0857cca472236f. All 107 rows were inspected;
none applies to these nine owned rules. Its README/gaps/options.a also record
runtime new RegExp(pattern, 'u') as a native gap. The same owned isolated probe
was repeated after rebase and exits 1 with:

    stage 0 can't lower RegExp with a nonconstant pattern yet

No custom matcher was added or extended. Existing partial pattern validation is
not claimed complete or compliant with the effective regex contract. Full
Go-compatible errors remain unresolved, as documented in
WAVE_14_REGEX_CONTRACT_REPORT.md. Real JSX and undefined labels remain shared
parser gaps; the tests retain independent Go-positive panic-70 witnesses.
ParseNode still exposes kind: string, so handed-node numeric dispatch remains
blocked on the shared parser/driver. No newly claimed rule or incompatible
rule.json was installed. Leaked-render does not use high-level IR, single
assignment or capture analysis, so it was not falsely parked for those analyses.

No batch-8 Diagnostic SHA was supplied. No leak-helper migration appeared in
this main advance. The full repository gate, nondefault options, new-rule emitted
JavaScript comparison, complete pattern grammar and end-to-end JSX parity are
not claimed covered. This is a landing-ready partial unit with named remaining
boundaries; no further batch was claimed during its landing-first turn.
