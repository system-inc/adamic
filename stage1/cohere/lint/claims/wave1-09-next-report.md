Built: pushed the next three claims and concrete registration/JSX blocker evidence; no new rule ports.
Commits: existing work 219c4b3c pushed; new pre-code claim be287557; evidence commit follows.
Commands and outputs: setup exit 0 in 15s, nproc 5; selection audit scanned 297 refs; 14 upstream tests, registry tests and uncached byte oracle passed.
Mutants: one-byte native output rejected by external comparison; eleven inherited invalid-descriptor controls rejected; no per-rule semantic mutants run.
Not covered: new rule code, complete findings/fixes parity, rule mutants, corpus parity, throughput and full gate; shared infrastructure blocks the requested contract.

## Selection

Continued on codex/lint-wave1-09. `git push origin codex/lint-wave1-09` first
reported Everything up-to-date. Fetched all origin heads without historical
submodule recursion:

```
git fetch --no-recurse-submodules origin '+refs/heads/*:refs/remotes/origin/*'
```

Main was ef3d907e. Read the helper report on origin/codex/lint-helpers, which
links ../HELPERS.md for its original 46-rule handoff, then the full JSON inventory
on origin/codex/lint-inventory at 73ac2eb0963e1a4166eaa0fbd160203f11dcdbdf.
All 297 fetched origin refs and 27 claim Markdown documents were scanned. Rule
name boundaries prevent substring claims, and main source/registration files
outside helper, inventory and testdata data determine observed main ports.
The full queue, branch SHAs, reasons for exclusions and three selections are
recorded in wave1-09-next-evidence/selection.json. audit.py reproduces the audit.
It ignores this worker's newly appended reservation when replaying selection.

The first available rules were:

1. structure/tailwind-no-physical-direction, the last available rule in the
   original 46-name handoff.
2. @eslint-community/eslint-comments/require-description, first available
   inventory syntax-only row after that handoff.
3. @next/next/google-font-display, next available inventory syntax-only row.

Syntax-only here follows the inventory's needs_type_information=false and
binding_only=false fields, retaining JSON rule order. Both syntax-ready and
syntax-waiting-on-helpers cohorts are included, 287 entries total. The later
comments handoff in HELPERS.md is separately reported under helpers/comments;
the original helpers/REPORT.md still describes its 46-rule cohort. Claims were
pushed at be287557 before implementation. The prior React claims remain reserved.
No unclaimed-exhaustion result is asserted: these three were available.

## Why the requested ports cannot be certified in the permitted territory

The installed directory generator still reads rule.ts, renders rule.ts imports,
and requires mutant files to end in .ts. Main does not contain this registration
package; fetching it does not provide an .a-compatible replacement. The earlier
rule-directory ownership contract in CLAUDE.md remains applicable. Changing the
shared generator, copied-source harness or parser is outside these owned rule
files. No shared source file was changed here.

This is an executed control, not an assumption from the source. In a temporary
copy of the five baseline directories, the actual generator exits 0 and prints
all five expected names. Renaming only no-debugger/rule.ts to rule.a, preserving
its bytes, makes it exit 1 with rule.ts: no such file or directory. Generation
fails before any candidate rule can compile or run. probe-registration.py and
registration.log reproduce and retain the result. It receives no semantic-mutant
credit. No .ts source was substituted for a new Adamic implementation.

All three selected rules also have valid JSX input surfaces. Tailwind visits every
literal, including JSX class attributes; RequireDescription's comment corpus
includes JSX comment containers; GoogleFontDisplay listens exclusively on
JsxOpeningElement and JsxSelfClosingElement. Two executable JSX inputs are rejected
by the shared parser, and one is accepted with the wrong AST shape:

| Input | Node source | Emitted JS | Sanitized native | Go JSX parser |
|---|---:|---:|---:|---:|
| Tailwind ordinary string control | 0 | 0 | 0 | 0 |
| Tailwind JSX className | 70 | 70 | 70 | 0 |
| Directive ordinary comment control | 0 | 0 | 0 | 0 |
| Directive JSX comment container | 0, wrong AST | 0, wrong AST | 0, wrong AST | 0, JSX AST |
| Google Fonts link without display | 70 | 70 | 70 | 0 |

For each input, all three Adamic parser backend logs are byte-identical.
The independent Go parser accepts all five with no diagnostics. In particular,
`const x = <div>{/* eslint-disable */}</div>;` is a JsxElement holding a
JsxExpression under Go, but a BinaryExpression holding a TypeAssertionExpression
and RegularExpressionLiteral under Adamic. Thus a successful exit is not evidence
of JSX support. No failing parse is counted as a clean rule result.
The two panic diagnostics are expected GreaterThanToken, got Identifier, at
22 for Tailwind and at 32 for Google Fonts. See the complete backend logs.

probe-go-jsx.py builds the repository's independent Go parser oracle through a
scratch overlay and changes only its source mode from TS to TSX and its synthetic
filename to /source.tsx. The Go parser itself is unmodified. It asserts every
input succeeds and the directive JSX tree differs from the Adamic tree, then
prints the observed reinterpretation. This is a parser blocker probe, not a rule
implementation or a per-rule semantic mutant.

The sanitized parser binary and emitted-JS artifact used here were built by the
previously committed build_probe.go on this same source state; intervening commits
only add claims and evidence. Rebuild with:

```
source /workspace/adamic-tools/env.sh
go run stage1/cohere/lint/claims/wave1-09-evidence/build_probe.go > /tmp/wave109-build.log 2>&1
```

For each owned raw `.ts.txt` input, run the source parser via oracle/node.mjs,
the emitted /tmp/lint-wave1-09-parser.mjs via the same runtime loader, and
/tmp/lint-wave1-09-parser with `--whole`, redirecting each output to its own log.
The native build uses ASan/UBSan. Raw inputs are not new authored .ts modules.

## Checks performed

`bash cloud/setup.sh > /tmp/lint-wave1-09-next-setup.log 2>&1` exited 0.
Sourced /workspace/adamic-tools/env.sh. nproc printed 5. Exact timing lines:

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (15s)
setup: done in 15s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

From cohere:

```
go test ./internal/lint/rules/tailwind ./internal/lint/rules/core ./internal/lint/rules/next -run '^(TestNoPhysicalDirection|TestRequireDescription|TestGoogleFontDisplay)' -count=1 -v -timeout 15m > /tmp/lint-wave1-09-next-upstream.log 2>&1
```

Exit 0, 14 top-level tests passed; package times 0.011s, 0.012s and 0.007s.
This verifies the upstream rule suites, not Adamic parity. Full output: upstream.log.

`go test ./stage1/cohere/lint/registry -count=1 -v` exited 0; registry-tests.log
contains all rejection controls and deterministic generation checks.
`ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v -timeout 10m`
exited 0 in 0.827s. The native stdout one-byte mutation is rejected by Node output
comparison. Native and Node caches each report zero hits and one miss.
All tests wrote logs directly, with no output pipes.

The eleven registry invalid-descriptor controls are duplicate public name,
unknown field, missing named hook, bad kind, missing oracle export, no listener,
missing factory, missing class, missing finish hook, unsafe public name and
duplicate oracle adapter. All fail descriptor validation as intended. Neither
those controls nor the shared oracle's byte mutant satisfy the requested one
compiling semantic mutant per new rule. None of those three exists here.

## Limits and handoff

No new rule is registered or declared ported. No complete source/finding/fix
comparison over compiler, stage1 or upstream rule cases was possible. No native,
Node or Go findings-per-second measurement was made; rates are unavailable,
not zero. Setup's cache warmup compiled packages without running their tests.
No full gate or repository-wide vet was run for claim/evidence-only changes.
The comments helper now exists on the helper branch, but it does not supply
.a registration or independent JSX parsing, and its availability is not rule parity.

The next concrete prerequisites are .a support in the shared directory generator
and test copier, followed by a JSX parser adapter that agrees with Go and refuses
unsupported shapes explicitly. Once those are available, these three reservations
can continue as full ports without stealing another worker's rules. This report
and all raw evidence are pushed on the same branch. No pull request is opened.
