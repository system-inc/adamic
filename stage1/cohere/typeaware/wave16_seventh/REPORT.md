Built react/jsx-fragments and react/jsx-no-undef; context-value analysis is partial.
Commits: claim 5282ee8e5; implementation 5188a56fa rebased onto main c01907a70.
Validation: prior ports 431.670s; bridge 65.387s; new gate passes 293 controls and both corpora.
Mutants: all 18 native rule mutants, the new signature-fact mutant and retained-handle mutants caught.
Not covered: cross-file context integration, emitted JavaScript and the complete repository gate.

No further claims were taken. All source changes are inside this unit's rule
modules and claim file. Shared parser, compiler, generator and harness files
were not changed. Rebase and final fetch confirm current main c01907a70 is
an ancestor of the worker branch. Only codex/typeaware-wave-16 is pushed.

The new modules consume cached numeric nodes supplied by the driver. rule.json
listeners use parser SyntaxKind values: fragments [285,286,289], context
values and undefined JSX names [286,287]. The driver switches on numeric
kinds. A single compatibility boundary converts legacy parser kind names once
per node; rule bodies do not compare kind strings or refetch parser nodes.

Each canonical comparison includes byte ranges, message IDs and text, fixes
and suggestions. These three Go rules offer no fixes or suggestions; the
empty arrays are compared too, including generic fragment tags whose types
Go deliberately does not rewrite. Four runner profiles cover syntax/element
fragments and allowGlobals false/true. The independent oracle uses the
unmodified cohere production registry and imports no bridge implementation.

The new raw checker question is wave16-resolved-signature-declaration. It
returns declaration file/kind/span, body presence and async/generator flags.
All construction, memo, return-path and escape judgments remain in Adamic.
Its Go and Adamic modules live together in the rule directory. The owned
validator installs the Go module using a tagged build overlay and applies
the existing additionalAnswer dispatcher hook virtually. Shared registration
is left to integration, consistent with this unit's ownership restriction.

The supported context rule includes direct and indirect constructions, provider
resolution, function/class component classification, bounded recursion,
primitive checks, memo inputs, nested memos, resolved local helpers, async and
generator results, return paths and symbol-identity escape checks. Go's Unicode
capitalization and fmt %q behavior are generated from its Unicode tables.

Observed blocker: the standalone foreign arena builds and runs normally.
Replacing the full suite's Context with context_foreign_gap.a rejects the
program at stage1/typescript/parser/parser.ts:34:20:

    Adamic 0.1 refuses this escaping a constructor before every field is set
    (stored, passed, or a method called on it, which could read a field that
    holds undefined while its type says otherwise); assign every field first,
    then use this

This is a full-rule-graph integration refusal, not evidence that the parser
or standalone adapter is generally unavailable. Its cause is not established
here. The owned validator preserves both observations. A separate imported
helper control reports normally in Go and reaches an explicit native NotYet
refusal (exit 70); it is not counted as agreement. The context rule therefore
remains blocked and incomplete. No shared compiler/parser workaround was made.

Validation commands, with output written directly to logs:

    source /workspace/adamic-tools/env.sh
    python3 stage1/cohere/typeaware/wave16_seventh/validate.py /workspace/wave16-artifacts/seventh-gate
    ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./stage1/cohere/typeaware -run '^TestWave16(Agreement|FollowupAgreement|ThirdAgreement|FourthAgreement|FifthAgreement)AndMutants$' -v
    go test -count=1 -timeout 30m ./bridge/tsgo/...
    go vet ./stage1/cohere/typeaware ./bridge/tsgo/checker
    gofmt -l cmd internal

The prior-port command additionally sets its five artifact directories, both
corpus manifests and ADAMIC_TYPESCRIPT_SOURCE; the detailed commands and
outputs are preserved in the evidence and original landing reports. Its five
suites pass controls, option profiles, full canonical corpus comparisons,
sanitizers, fifteen native mutants and five released-handle counterexamples.
The new validator passes 293 controls with findings 388/368/381/361 across
its four profiles; compiler 77 files and repository 287 files each find zero,
matching the recorded volume. All normal and sanitizer comparisons agree.
The new released-handle probe panics 70; retaining the registry entry changes
it to exit 0, which the refusal check catches. Vet and formatting logs are empty.

The continuation reuses the earlier successful toolchain setup: timing lines
Go 0s, clang 1s, Node 1s, submodules 1s, cache 85s, done 85s. nproc is 5, with
four effective cores recorded by cpu.max. It did not rerun setup.

Native versus Go process times, medians of three alternating serial count-mode
runs after concurrent gates ended, including checker load and native parsing:

| Corpus | Native | Go | Native / Go |
| --- | ---: | ---: | ---: |
| TypeScript compiler | 2.467959s | 0.285674s | 8.64x |
| Repository | 0.354363s | 0.124575s | 2.84x |

Numeric listener declarations do not make this legacy-parser path faster than
Go. The shared kind-indexed parser/driver integration and performance work
remain outside this unit. Source hashes, pins, commands, canonical outputs,
sanitizer stderr, mutant bytes and serial timing samples are under evidence/.
Build archives and executables remain in scratch, not in the commit.

Mutants below all compile and exit 0 with empty native stderr. Only the
independent Go finding/fix/suggestion comparison catches them:

| Mutant | Change | First differing byte |
| --- | --- | ---: |
| fragment-name | React.Fragment recognition changed | 464 |
| undef-intrinsic | Uppercase names treated as intrinsic | 62 |
| context-object | Object construction labeled as array | 43450 |
| resolved-signature-present | Raw checker declaration body hidden | 112133 |

The raw-question counterexample is an additional manual tagged archive/native
build; its source, build logs, run output and result JSON are preserved.
All prior rule mutants are listed below and in provenance.json. Retained
registry entries are separately caught by the required panic-70 check.

| Prior mutant | First differing byte |
| --- | ---: |
| throw-range | 130 |
| backreference-range | 9283 |
| arrow-fix | 6056 |
| listener-general | 723 |
| render-number | 7165 |
| mock-range | 3576 |
| function-range | 132 |
| native-range | 4598 |
| wrapper-range | 5183 |
| import-write | 66 |
| exponent-base | 15678 |
| nan-comma | 12079 |
| shell-heredoc | 8299 |
| sql-string | 11784 |
| alert-range | 130 |

Coverage limits: no complete go test ./... gate, no emitted-JavaScript
comparison, no shared configuration/module-loader changes, and no successful
full-graph cross-file context analysis. The three HIR-dependent React-hook
claims remain parked under the separate analysis-module instruction.
