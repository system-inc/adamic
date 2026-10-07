Built: general configured id-match execution using an Adamic regex VM and raw Go syntax-program facts.
Commits: the original three rules are now implemented; this report supersedes the earlier missing-regexp blocker.
Commands: 272 configured inputs, 1,050 independent regex decisions and both frozen corpora agree; default gate PASS 90.110s.
Mutants: configured rules, VM context, raw rune facts, default rules, provenance and retained released handles all caught.
Not covered: emitted-JavaScript checker adapter, every upstream fixture/options matrix, or full repository gate.

`regexp_program.a` executes the immutable program returned by the new
`regexp-program` question. `regexp_program.go` uses only Go's public regexp
syntax parser/compiler; it never sees the string being matched and never
computes a match, lint predicate, message, finding or edit. Unicode character
ranges and SimpleFold cycles are compiler data. Every match runs through the
native Thompson VM. Captures and ordered alternatives cannot change a boolean
RE2 MatchString result; their edges are followed without keeping capture data.
The VM tracks epsilon closures per code-point position, handles ASCII word
boundaries and line/text anchors, and consumes rune ranges natively. Go rejects
lookaround and backreferences, exactly as the production option decoder does.
Invalid patterns explicitly panic rather than silently matching everything.

`id_match.a` now calls this matcher from `run(pattern, flags)`. Its existing
native AST checks, all four configured flags, spans and messages are unchanged.
The old callback entry remains available, but the configured driver now calls
the actual public pattern API. No test-only underscore matcher remains in that
driver. All extractable configured production test patterns are included:
272 inputs, 209 findings and 70,068 identical bytes across Go, native and
ASan/UBSan/LeakSanitizer native, SHA256
`d3b68f4b44a262046ab95bfdf190410d657fd3befb79dc6eb892df18a20c793c`.
All fixes and suggestions are zero, exactly as production.

An independent Go regexp oracle also compares 35 patterns by 30 input strings,
1,050 decisions. Coverage includes empty programs and matches, Unicode scripts
and categories, Kelvin/long-s/sigma case folding, astral characters, Go-only
anchors and inline flags, greedy/lazy and bounded repetition, captures,
alternation, dot modes, word boundaries, multiline anchors and negated classes.
Native and fully instrumented native agree byte for byte, with empty stderr.
The VM-context mutant ignores empty-width constraints, compiles and exits 0;
only the independent match bytes catch it. A raw-fact mutant increments rune 95
in the serialized program, also compiles and exits 0 with empty stderr, and is
caught only by independent match bytes. These are not compilation or sanitizer
kills. The configured denylist and match mutants still compile, exit 0 and
change findings. Raw syntax facts tests pass on valid and invalid patterns.

The new question rejects released programs with exit 70 and
`invalid or released checker handle`. The retained-registry mutant makes that
same request exit 0, caught by the required panic check. The unchanged wave gate
also kills all three default rule mutants and the provenance mutant. Its final
41 controls have 15 findings, 9,881 equal bytes; frozen compiler 77 roots have
zero findings, 5,241 equal bytes; frozen repository 287 roots have zero findings,
18,485 equal bytes. Each also agrees under sanitizers.

The Go question and Adamic interpreter are separate new files; only one physical
registration line was added to the shared bridge dispatcher and one import to
this worker's id-match rule. No shared harness, registration generator, compiler
or parser was edited. Go regexp compilation is an explicit additional raw
compiler-fact use of the existing C bridge. Native matching and all lint
judgments remain Adamic. No submodule pin changed.

One whole-process default compiler observation: native 1.690936s versus Go
0.298427s, 5.67x; repository native 0.229002s versus Go 0.120792s, 1.90x.
These are observations, not benchmark medians. Setup passed in 26s, tools and
submodules ready 0s, cache warm 26s; nproc 5, cpu.max 400000 100000.
The bridge checker tests passed in 0.153s and vet produced empty output.

```sh
source /workspace/adamic-tools/env.sh
python3 -u stage1/cohere/typeaware/wave-29-configured/check.py \
  /workspace/wave29-regex-controls > /tmp/wave29-regex-controls.log 2>&1
python3 -u stage1/cohere/typeaware/wave-29-configured/regex_check.py \
  /workspace/wave29-regex-vm > /tmp/wave29-regex-vm.log 2>&1
ADAMIC_WAVE29_ARTIFACTS=/workspace/wave29-regex-default \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave29-typescript \
go test ./stage1/cohere/typeaware -run '^TestWave29AgreementAndMutants$' \
  -count=1 -timeout=30m -v > /tmp/wave29-regex-default.log 2>&1
python3 -u stage1/cohere/typeaware/wave-29-configured/regex_mutants.py \
  /workspace/wave29-regex-facts > /tmp/wave29-regex-facts.log 2>&1
go test ./bridge/tsgo/checker -count=1 -timeout=10m -v \
  > /tmp/wave29-regex-checker.log 2>&1
go vet ./bridge/tsgo/... ./stage1/cohere/typeaware \
  > /tmp/wave29-regex-vet.log 2>&1
```

The standalone regex fixture initially had an empty files list and was refused
by the config loader; naming its anchor fixed the input before comparison.
The RegExp constructor gap probe still refuses that separate API, but it no
longer blocks this instruction interpreter. Emitted JavaScript remains a shared
checker-adapter gap, as previously recorded. The native default corpora and
configured findings are complete for the measured populations. No full gate,
JSX/JavaScript corpus population or exhaustive options matrix is claimed.
Complete logs, fixture inputs and compressed streams are in regex-validation/.
