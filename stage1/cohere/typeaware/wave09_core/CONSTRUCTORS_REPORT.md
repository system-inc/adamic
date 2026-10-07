Built: native constructor-argument diagnostics for strings/templates, indirect patterns and regex arguments with overriding flags.
Commits: this component follows pushed landing tip f4f8967f on codex/typeaware-wave-09, based on main e8ba3d5d.
Commands/output: verify_constructors.py PASS, 462 actual Go source cases, 289 default findings and 241 allowEscape findings with complete serialized equality.
Mutant: extend a mapped string finding's end by one byte; compiles and exits 0 with empty stderr, caught by Go bytes at offset 53.
Not covered: native reference tracking/constant-expression resolution, complete constructor listener and the no-invalid-regexp engine; no new claims.

The adapter consumes explicit resolved argument inputs. Native string/template
judgments now connect the independently held cooked/raw mapper to pattern
findings. The pattern walker accepts raw member spelling and cooked/raw offsets
so allowEscape observes string escapes, rather than only regex escapes. Reporting
uses mapped byte spans, including the production mapper's deliberately permissive
behavior around surrogate escapes. It retains Go's equal-byte-length shortcut
and declines unreconcilable mappings.

Indirect arguments have no raw member source. Every message ID is reported once
at the whole argument, even when several classes produce that ID. Regex-literal
arguments with flags are checked under those flags and offer no Unicode edits;
unknown overriding flags decline that argument. Without constructor flags they
remain the literal listener's responsibility, including its suggestions.

verify_constructors.py writes 462 real source files for the unchanged production
Go rule. Its oracle uses the actual checker, reference tracker, constant-string
evaluator and rule listeners. The native component receives explicit fixture
metadata for the resolved arguments, which is why this certifies judgments and
spans rather than native reference or constant resolution. It does not ask Go
for a verdict or source map. Go independently decides which findings are right.
Aliases and concatenations in these fixtures exercise the oracle, while their
resolved native pattern inputs are intentionally explicit. This distinction is
not counted as a full native constructor-rule port.

Inputs cover plain/JSON-escaped/backtick strings, aliases and concatenations,
Unicode and CRLF byte offsets, repeated IDs, nested classes, malformed patterns,
Unicode/non-Unicode class behavior, constructor flag overrides, unknown flags,
and regex arguments passed without overrides. Default options match 128783
bytes / 289 findings; allowEscape matches 111014 bytes / 241 findings. Every
diagnostic field, fix and suggestion is serialized. Normal native, sanitized
native (ASan/UBSan/LeakSanitizer), source Node and emitted JavaScript match Go
in both option modes with empty stderr. The native span mutant is caught only
by the byte comparison.

Single component-process native/Go times were default 0.015695 / 0.165871
seconds and allowEscape 0.015863 / 0.164831 seconds. Native has embedded resolved
fixtures while Go loads source and runs the actual tracker, so these are not
full-rule performance comparisons. No speed claim is made. Setup is reused
from this workspace, previously 88 seconds with nproc 5.

Reproduce after sourcing /workspace/adamic-tools/env.sh:

```
python3 stage1/cohere/typeaware/wave09_core/misleading_character_class/verify_constructors.py > /workspace/wave-09-constructors-test.log 2>&1
python3 stage1/cohere/typeaware/wave09_core/misleading_character_class/verify_patterns.py > /workspace/wave-09-constructor-patterns-regression.log 2>&1
python3 stage1/cohere/typeaware/wave09_core/misleading_character_class/verify_literals.py > /workspace/wave-09-constructor-literals-regression.log 2>&1
```

The existing 2616-pattern oracle and literal-source controls/corpora are rerun
because the escape-source walker changed. Evidence is retained in
validation-constructors. No shared harness, compiler, bridge registration or
Go production file changed, and no new checker question was introduced.
The prior landing and released-handle evidence remains in validation-landing;
the checker ABI did not change here. All new Adamic source is .a.

The current main compiler successfully builds the full literal slice, as the
landing report records. Its former timeout is not a current excuse for missing
constructor behavior. Remaining port work is native global/alias tracking and
constant-expression resolution, wiring checkedByACall into source traversal,
and the matching native regex engine for no-invalid-regexp. Complete upstream
options, every fixture and the full repository gate remain uncovered.
