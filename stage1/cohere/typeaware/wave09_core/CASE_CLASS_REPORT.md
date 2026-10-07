Built: native literal case-class widening, class-rune escaping and trailing-dash escaping for the existing regex claim.
Commits: follows pushed bc98abebe on wave-09, based on current main f8013f0b; no new claims or React parking.
Commands/output: verify_class.py PASS for 2,228,228 widening queries, 4,021 rune escapes and 273 dash cases, Go/native/sanitizers/both Node modes.
Mutants: omit the opening-bracket escape, caught at byte 141223; reverse backslash parity, caught at byte 0; both native builds and runs succeed with empty stderr.
Not covered: class-body decoding/closure, regex rewriting and engine, full rule parity, numeric handed-node integration or another corpus sweep.

CaseClasses widens a literal through the existing native equivalence tables,
spelling each group member with escapeClassRune. It preserves the production
empty pattern/false result for isolated runes. escapeClassRune normalizes
invalid Go runes, including surrogates, to RuneError before conversion and
escapes the five punctuation characters with class meaning. escapeTrailingDash
leaves an already escaped terminal dash intact and escapes a literal terminal
dash according to the parity of immediately preceding backslashes.

The independent oracle directly calls unchanged production CaseClass and
EscapeClassRune. An owned Go overlay exposes the unchanged private
escapeTrailingDash helper. The probe prints strings as sequences of scalar
values, so all character values, including NUL, are compared without differing
console encodings. This serialization preserves complete returned strings.
Every integer -1 through 1114112 is queried in both modes. Both positive and
negative results are checked: an unexpected nonempty pattern or true result
adds a line and fails comparison. Rune escape controls cover invalid boundaries,
all five meaningful punctuation characters, controls, supplementary characters
and 4,000 deterministic sampled integers. Dash controls cover thirteen counts
of backslashes with seven prefixes and three suffixes, including astral and NUL.

Reproduce after sourcing the existing successful toolchain:

```
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/typeaware/wave09_core/invalid_regexp/verify_class.py > /workspace/wave-09-class-test.log 2>&1
```

The final source passes normal native, ASan/UBSan/LeakSanitizer, source Node
and emitted JavaScript. Both semantic mutants compile and run with exit zero
and empty stderr; only comparison catches their changed outputs. Full streams,
generated probes, hashes and final measurements live in validation-case-class.
Class output is 225437 bytes; dash output is 7362 bytes. Native versus Go
whole-process widening/escape time is 0.216356 versus 0.064018 seconds.
Sanitized native takes 1.666990 seconds. These are single component runs,
not full lint benchmarks. Setup is reused: 88 seconds, nproc 5.

No AST kinds are read, no node is fetched and no checker ABI is changed.
Shared parser, generator, harness and compiler files were not edited. The
existing numeric rule.json declarations are unchanged. The shared parser
still provides string kinds rather than a numeric handed-node contract,
blocking migration of legacy execution. No Diagnostic migration SHA was named.
Exact regex class-body parsing/rewrite and engine work remain unfinished;
these helpers do not certify the complete no-invalid-regexp rule. Full
misleading-class constructor tracking also remains unfinished. None is a React
analysis blocker, so the new React parking exception was not applied. Prior
released-handle and corpus evidence remains in LANDING_F801_REPORT.md.
