# Octal-escape oracle investigation

The failure belongs to stage 1's shared fixture adapter, not Go cohere's rule.
Area base: `c8f6d74f5387df926ae12d8babd2780564b60097`. Cohere pin:
`7945d102a6c18dd36adf9114a758ce646e8b2359`.

## Reproduction and ownership

[evidence/reported.ts.txt](evidence/reported.ts.txt) contains the exact six bytes of the reported
fixture. The original shipping oracle, built with `goOracleIn` and the registered upstream Go
adapters, exits 2 on an unclassified row selecting `no-octal-escape`. `GOTRACEBACK=all` output is
[evidence/reported.stack.txt](evidence/reported.stack.txt).

The panic is `invalid corpus`, caused by TypeScript's octal-escape parse diagnostic. It is raised
at `stage1/cohere/lint/testdata/oracle.go:186`, in `collect`, before selecting or running a rule.
The stack's virtual filename `cohere/adamic_lint_oracle.go` is mapped to that stage-1 file by the
Go overlay. The stack continues through `run` at line 68 and `main` at line 292; it contains no
Go rule frame. The rule adapter in `rules/no-octal-escape/oracle.go` merely returns the unmodified
upstream `NoOctalEscape` and takes no options.

Cohere's independent `go test -json -count=1 -run '^TestNoOctalEscape' ./internal/lint/rules/core`
passes **117 test results, 0 fail, 0 skip** (seven top-level tests). Its
`no_octal_escape_test.go` explicitly exercises this exact fixture in `TestNoOctalEscapeFires`.
Cohere's `rule_testing.RunWithOptions` evaluates the recovered AST without rejecting input parse
diagnostics; that is how a rule about legacy escapes can be tested on those escapes. No Go rule
change is warranted.

## Smallest inputs

[evidence/smallest-octal.ts.txt](evidence/smallest-octal.ts.txt) is the four-byte complete literal
with a backslash-one escape, the smallest complete octal-escape literal. It produces the same
adapter panic; the full stack is [evidence/smallest-octal.stack.txt](evidence/smallest-octal.stack.txt).
A bare backslash-zero escape is legal and is not this failure.

The guard is more general than octal escapes. [evidence/smallest-guard.ts.txt](evidence/smallest-guard.ts.txt)
is one byte, an unterminated single quote, and also panics. Empty input exits 0. Thus one byte is
the minimum source length for this guard, while four bytes is the meaningful minimum for the
reported octal family. Its stack is [evidence/smallest-guard.stack.txt](evidence/smallest-guard.stack.txt).

## Adapter repair

The main parity test already called `recoveryRows` before replaying captured inputs, which masked
the capture defect. A shipping-oracle consumer receiving `captureUpstream`'s manifest directly
was missing that caller-specific step.

Capture now probes the same Go oracle's `--diagnostics` mode and attaches `recovery` to rows with
parse diagnostics when no existing mode owns the row. The former caller helper and capture share
one `classifyRecoveryRows` function. This retains every source, selected rule, filename and option.
The Go oracle also understands the existing `unsupported-recovery` marker as findings-only.
That marker stays in the manifest and remains an explicit port refusal in the existing parity
suite. Neither recovery mode runs the fix engine or claims a fixed file.

The unmarked-input guard remains intact. A raw invalid input without recovery metadata still
fails by name; this is deliberately checked. No port TypeScript/Adamic implementation, upstream
rule, dispatch, registry or compiler changed.

## Every upstream fixture, once per adapter version

The sweep isolates each captured case in a separate oracle process so the first panic cannot
hide later outcomes. Typed rows receive actual one-file strict programs, as in `TestRulesAgree`.
All **3,879 registered upstream source/rule/options combinations** were run with the TypeScript
source input set. This covers the stage-1 registry, not unported cohere rules.

| Rule | Panicking rows before |
|---|---:|
| `no-octal-escape` | 79 |
| `@typescript-eslint/method-signature-style` | 8 |
| `prefer-template` | 5 |
| `react/jsx-no-comment-textnodes` | 1 |
| Total | **93** |

The reported fixture is one of those 93, so **92 other rows** also panicked. All failures were the
same adapter parse guard, not rule panics. Eight already carried `unsupported-recovery`; 85 had
no recovery marker. [evidence/panics-before.json](evidence/panics-before.json) records every source,
rule, synthetic fixture filename and full stack. After repair: **3,879 exits 0, zero panics,
zero skipped rules**. [evidence/census.json](evidence/census.json) records the totals.

## Validation and controls

`ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/tsprinter-tsc/typescript` was set to pinned
`050880ce59e30b356b686bd3144efe24f875ebc8`; cohere's capture input was supplied through
`COHERE_DOCS_CAPTURE` by the existing helper. No selected test skipped an input.

- Three regressions in `recovery_inputs_test.go`: **PASS, 24.036s**. They replay the exact captured
  fixture directly, keep unsupported recovery explicit, retain the unmarked-input failure, and
  reject a missing diagnostic flag count. Log: `/tmp/octal-oracle-regressions.log`.
- `go test -v -count=1 -timeout 30m -run '^TestRulesAgree$' ./stage1/cohere/lint`:
  **PASS, 274.833s**, **13,773,852 identical bytes** across Go, source Node, emitted JavaScript and
  sanitized native. Existing parser-recovery refusals remain explicit. Log:
  `/tmp/octal-oracle-rules-agree.log`.
- Independent cohere rule suite: `/tmp/octal-cohere-own-suite.jsonl`.
- Isolated sweeps: `/tmp/octal-oracle-all-before.log` and `/tmp/octal-oracle-all-after.log`;
  individual process outputs under `/workspace/scratch/octal-oracle/{before,after}/`.
- Scratch-only controls in [evidence/run_oracle_mutants.py](evidence/run_oracle_mutants.py):
  dropping capture metadata fails naming `OctalEscape.ts`; removing the oracle's explicit
  unsupported-mode support fails naming `unsupported.ts`; overwriting an existing port boundary
  fails the classifier assertion; accepting missing flags fails its count assertion. All four
  compile and are caught by test assertions. Logs: `/tmp/octal-oracle-mutants.log` and
  `/tmp/octal-oracle-mutants/`. An initial metadata mutant had an unused-variable build error;
  it was corrected and that compiler failure was not counted.
- `go vet ./stage1/cohere/lint`: exit 0, `/tmp/octal-oracle-vet.log`; changed Go files are gofmt-clean,
  `/tmp/octal-oracle-gofmt.log`.

Setup succeeded: Go 0.020s, Node 0.024s, submodules 0.077s, markdown 0.078s, clang 0.220s,
build 38.597s, cache 38.739s, done 38.773s. `nproc`: 5; quota: 4 CPUs.
`/tmp/octal-oracle-setup.log` retains the timing lines.

The whole lint package, compiler corpus, profiling, throughput and full repository gate were not
run. The requested single-fixture reproduction, own-rule suite and all upstream oracle fixtures
were run, followed by the complete registered lint agreement. No general parser-recovery support
or converged fixes on invalid source are claimed.
