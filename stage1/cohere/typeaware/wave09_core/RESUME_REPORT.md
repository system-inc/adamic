Built: native character-class parsing, whole-pattern judgments, literal diagnostics/suggestions, Go rune quoting and WTF-8 flag normalization.
Commits: prior pushed implementation c5647eb1; this report and evidence accompany the resumed implementation commit on codex/typeaware-wave-09.
Commands/output: verify_patterns.py, verify_quotes.py, verify_format.py, verify_frontend.py, verify_literals.py --source-only and verify_literal_tokens.py pass in their stated scopes.
Mutants: surrogate folding, malformed-pattern rejection, rune escaping, error-prefix trimming and suggestion-edit range mutants all differ from production Go bytes.
Not covered: complete regex rules; the combined native literal driver times out in shared freshness analysis, and pattern compilation, constructor tracking and cooked/raw mapping remain unfinished.

## What changed

The original wave-09 rules and no-label-var remain pushed and unchanged. No new
rules were claimed. All new source is in the two owned regex directories, and
all Adamic files are .a. No shared harness, generator, compiler, parser, bridge
registration or production Go rule was edited. There are no new checker questions
in this resume. Existing released-handle checks remain unchanged and were not
rerun; their normal/sanitizer and retained-handle-mutant evidence is in REPORT.md.

character_class.a now ports cohere's regexsyntax class scanner and parser. It
keeps UTF-8 byte offsets, Unicode/surrogate escape folding, nested v classes,
set operators, range endpoints, malformed input and uint32 hex accumulation.
pattern_findings.a connects parsed members to the existing six detectors,
including escape provenance, range splitting, astral expansion and the whole
pattern's malformed-input/Unicode-brace gates.

literal_rule.a adds actual diagnostic and suggestion construction. Its parser
adapter accepts the native node table; its separately compiled token judgment
accepts a literal's raw text and byte bounds. literal_main.a is the combined
source-parser driver. It is a literal slice, not the complete constructor-aware
rule: constructor flag overrides/checkedByACall, alias/global-object tracking,
constant arguments and cooked-to-raw mapping are still missing.

rune_quote.a implements Go %q rune formatting, backed by Go Unicode 17.0.0
printability data in printable_ranges.a. Its generator and oracle are outside
Adamic and call unicode.IsPrint and fmt directly. Invalid rune values map to the
replacement rune, as Go does. The flags frontend now uses this formatter.
A new full-finding control exposed a real mismatch for lone surrogate flags:
Go's AST stores WTF-8, and its rune walk reads three invalid UTF-8 bytes as three
replacement runes. The native flag walk and message prefix now match that
observed behavior. These controls were retained and pass; they were not dropped.

## Commands and exact results

Source /workspace/adamic-tools/env.sh before running the verifiers. Every command
writes output to a log. Reproduction order:

```
python3 stage1/cohere/typeaware/wave09_core/misleading_character_class/verify_patterns.py > /workspace/wave-09-patterns-test.log 2>&1
python3 stage1/cohere/typeaware/wave09_core/invalid_regexp/verify_quotes.py > /workspace/wave-09-quotes-test.log 2>&1
python3 stage1/cohere/typeaware/wave09_core/invalid_regexp/verify_format.py > /workspace/wave-09-format-test.log 2>&1
python3 stage1/cohere/typeaware/wave09_core/invalid_regexp/verify_frontend.py > /workspace/wave-09-core-frontend-resume-test.log 2>&1
python3 stage1/cohere/typeaware/wave09_core/misleading_character_class/verify_literals.py --source-only > /workspace/wave-09-literals-final-source-test.log 2>&1
python3 stage1/cohere/typeaware/wave09_core/misleading_character_class/verify_literal_tokens.py > /workspace/wave-09-literal-tokens-test.log 2>&1
```

| Comparison | Inputs | Findings / bytes | Execution covered |
| --- | ---: | --- | --- |
| Class scanner/parser vs production regexsyntax | 2616 | 219914 bytes | Native, ASan/UBSan/LSan, source Node, emitted JS |
| Whole-pattern helper vs production checkRegexPatternWithReporter | 2616 | 242484 bytes | Native, ASan/UBSan/LSan, source Node, emitted JS |
| Rune quoting vs Go fmt | 4000 | 31705 bytes | Native, ASan/UBSan/LSan, source Node, emitted JS |
| Canonical flags/error formatting | 56 | 3460 bytes | Native, ASan/UBSan/LSan, source Node, emitted JS |
| Actual source flag findings vs production NoInvalidRegexp | 40 files | 35 findings, 6020 bytes | Native, ASan/UBSan/LSan |
| Literal controls vs production NoMisleadingCharacterClass, default | 27 files | 22 findings, 10971 bytes | Actual source on Node |
| Literal controls, allowEscape | 27 files | 17 findings, 8806 bytes | Actual source on Node |
| Literal judgments on explicit fixture token metadata | 27 files, both options | 19791 bytes including section labels | Native, ASan/UBSan/LSan, source Node, emitted JS |
| Compiler corpus, literal slice | 77 files | 0 findings, 7859 bytes | Actual source on Node vs production Go |
| Frozen repository corpus, literal slice | 287 files | 0 findings, 18485 bytes | Actual source on Node vs production Go |

Every comparison uses full serialized payloads. Actual diagnostics include all
rule/message fields, byte ranges, fixes and complete suggestions/edits. Parser
and pattern helper frames include their own exact fields and eligibility marks;
they are not advertised as complete rule diagnostics. The literal token test
uses explicit metadata extracted from the fixed control sources and expected
payloads from the unchanged production Go rule. It does not test native source
parsing. The source-Node test does parse the actual source with the native parser's
Adamic implementation running on Node.

The error-format test supplies raw compiler errors from Go as explicit component
test inputs, and independently asks production invalidPatternMessage for its
expected output. No native pattern compiler was implemented or tested by that
check. The full flag controls intentionally panic if pattern compilation is
reached. Flags now include controls, apostrophe/backslash, Unicode, private-use,
noncharacters and lone surrogate escapes.

Go oracles were built through temporary overlays inside cohere, so their shims
can call private production helpers. They do not change those helpers. Node runs
through the repository's existing oracle/node.mjs loader. Shared .a support is
available on origin/codex/lint-harness-dot-a at f4d98cab; it was inspected, not
merged or edited. No full shared-profile registration or repository-wide gate
was run. gofmt on the owned Go testdata and Python syntax checks pass.

## Native compilation blocker

The command without --source-only attempts the combined native source-parser
and literal listener. It has not completed:

```
python3 stage1/cohere/typeaware/wave09_core/misleading_character_class/verify_literals.py > /workspace/wave-09-literals-final-native-test.log 2>&1
```

Two builds were stopped explicitly with SIGQUIT after 177.731 and 147.257
seconds to capture stacks. A node-table-only listener build then timed out at
120 seconds. After separating token judgments into a function that independently
builds in 3.527 seconds and passes all comparisons, the combined driver still
timed out at 120 seconds. These are measured failures to finish within the stated
intervals, not proof that the compiler can never finish.

The captured running stack is internal/fresh.(*analysis).made/value/argument,
called by fresh.ProveWrites, lower.(*cycleFinder).unproven/slotsOf/findCycles,
then lower.Lower. This is a shared compiler integration/performance blocker.
No protected compiler source was modified. Dumps and all attempt logs are saved
under validation-resume. Neither native combined-driver parity nor its emitted
JavaScript is certified. Both use the lowering pass that did not finish.

## Mutants and time

Each semantic mutant builds successfully and exits 0 with empty stderr; production
Go byte comparison is what catches it:

| Mutant | First differing byte |
| --- | ---: |
| Disable Unicode surrogate-escape folding in class parser | 2888 |
| Drop the whole-pattern malformed scan guard | 238 |
| Misclassify the bell rune's escape | 50 |
| Trim one extra byte from a compiler error prefix | 372 |
| Extend the suggested u insertion range by one byte | 1434 in native token frames; 1426 on actual source Node |

Single whole-process helper timings, native / Go seconds: class parsing
0.008460 / 0.023476; whole-pattern judgments 0.021799 / 0.012542; rune quoting
0.003926 / 0.007733. These are component checks with native fixtures embedded
and Go fixtures loaded from JSON, not full-rule performance benchmarks.
The unchanged label rule's previously measured compiler/repository medians remain
1.448450 / 0.330366 and 0.211157 / 0.154313 native / Go seconds respectively.
No native full-regex-rule time is available.

Toolchain setup is reused from this same workspace: cloud/setup.sh previously
reported 88 seconds; nproc is still 5. The final source snapshots, test payloads,
measurements, mutant hashes and stack evidence are in validation-resume.
The native esregexp compiler and constructor/reference/source-map adapters are
still work remaining, independently of the shared compiler timeout. This batch
stops with both full regex rules explicitly incomplete and no further claims.
