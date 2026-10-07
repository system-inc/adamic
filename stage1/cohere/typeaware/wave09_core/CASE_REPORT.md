Built: native ECMAScript regexp case canonicalization, using pinned Unicode primitive tables and no AST kind checks or node fetches.
Commits: follows pushed 59201a24b on codex/typeaware-wave-09, still based on fetched current main e8ba3d5d; no new claims.
Commands/output: verify_case.py PASS, 1114114 integer inputs in both modes, 2228228 answers and 15603280 identical bytes against production Go.
Mutants: removing the non-ASCII-to-ASCII boundary differs at byte 2180; selecting the largest fold-orbit member differs at byte 379; both build and exit 0 with empty stderr.
Not covered: complete pattern rewrite, case-equivalence class widening, regexp2-compatible syntax compilation, numeric AST dispatch and complete regex rules.

The case component is part of the missing no-invalid-regexp preprocessing
dependency and can be held independently of the blocked shared AST migration.
canonicalize.a contains no node-kind reads, node fetches or parser dependency.
The six numeric listener declarations remain unchanged under listeners/.
The branch's landing cap was checked first: origin/main remains e8ba3d5d and
is already an ancestor of the pushed wave-09 branch.

Native canonicalization walks Unicode SimpleFold orbits and chooses their least
member under Unicode mode. Without that mode it preserves supplementary-plane
values and the Greek full-uppercase expansion cases, then applies primitive
uppercase with Go's ECMAScript non-ASCII-to-ASCII exclusion. That exclusion is
what keeps long s from comparing as ASCII S without a Unicode flag.

unicode_case_data.a contains 1505 changed uppercase pairs and 2994 changed
SimpleFold pairs from Go's Unicode 17.0.0 primitives. It does not contain
esregexp.Canonicalize verdicts. Its generator asks only unicode.ToUpper and
unicode.SimpleFold; the mode selection, guards, table lookup and orbit walk
are native. The verifier regenerates and compares the exact primitive data
before executing the independent canonicalization comparison, so Unicode data
drift cannot become a silent change. No new bridge question is involved.

The Go oracle directly invokes the unchanged production esregexp.Canonicalize
for every integer from -1 through 1114112, in both modes. This covers every
Unicode code point, including surrogate values, and both adjacent invalid
boundaries. It is not a claim about arbitrary fractional/nonfinite doubles;
the production Go API takes a rune and pattern scanners supply integers.
Normal native, ASan/UBSan/LeakSanitizer, source Node and emitted JavaScript match
all 15603280 output bytes with empty stderr. Both successful semantic mutants
are caught only by the production Go comparison.

Single whole-process native/Go times were 0.314882 / 0.164059 seconds for the
exhaustive stream. Native remains slower for this component; no performance
improvement is claimed. Sanitized native took 1.467287 seconds. These are
canonicalization/printing runs, not complete lint-rule timings. Setup is reused
from this workspace's 88-second toolchain setup and nproc remains 5.

Reproduce after sourcing /workspace/adamic-tools/env.sh:

```
python3 stage1/cohere/typeaware/wave09_core/invalid_regexp/verify_case.py > /workspace/wave-09-case-test.log 2>&1
```

The independent Go main is built via a temporary overlay inside cohere to
satisfy its internal import boundary. Shared compiler, parser, scanner, bridge,
registration, harness and production Go files are unchanged. All new Adamic
files are .a. Complete Go/native/sanitizer/Node and mutant streams, generated
probes, measurements and source hashes are retained in validation-case.
Existing rule corpus, ownership and listener-map evidence remains unchanged;
those unrelated suites were not rerun for this new, unintegrated pure component.

The precise shared speed-API blocker remains in LISTENERS_REPORT.md: ParseNode
has only a string kind and Rules.ask refetches that node. No string-to-number
conversion was introduced per rule to hide that gap. Complete regex ports also
still lack a full native reference-tracker/listener integration and matching
pattern compilation. This component does not call a substitute regex engine,
return Go verdicts through the bridge or claim findings/fixes/suggestions
agreement for either complete regex rule. The full repository gate remains
uncovered. No additional rule batch is claimed.
