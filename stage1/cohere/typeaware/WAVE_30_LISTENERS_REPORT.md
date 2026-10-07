Built: added numeric listener declarations to the eight completed wave-30 rule files; no new claims, shared parser changes or shared harness edits.
Commits: listener implementation dc40511f, based on previously pushed 04f2b72c; origin/main remains e8ba3d5d and is an ancestor of this branch.
Commands and outputs: listener gate PASS 49.914s; all nine existing wave-30 gates PASS 356.358s; vet PASS; setup 84s, nproc 5.
Mutants: all eight new listener mutants and thirteen existing rule/component/state/graph mutants exit successfully and fail independent Go byte comparison.
Not covered: numeric dispatch and handed-node entry points, complete React rules, full repository gate or full oracle fixture matrix; shared prerequisites block further implementation.

Only codex/typeaware-wave-30 has been pushed by this unit. Fetching all origin
heads and checking the remote directly confirms main remains e8ba3d5d. The branch
was already rebased and green on that base in 04f2b72c; this change receives a fresh
native oracle run before pushing to the same branch. No main or area/ branch push
is attempted. No additional rules are claimed while the retained React claims
remain incomplete.

Each implemented rule now exports a readonly numeric SyntaxKind list and exposes
it as readonly syntaxKinds on its rule instance. Numeric values come from the
pinned Go cohere AST constants, verified independently against its actual
production rule.Listeners source maps. The private Go oracle parses those maps
and resolves their keys through the pinned ast package; it does not use native
rule declarations or bridge questions. Native and sanitized native declaration
output matches all 348 oracle bytes.

| Rule | Numeric listener kinds |
| --- | --- |
| ISO date cut | 214 CallExpression, 213 ElementAccessExpression, 261 VariableDeclaration |
| callback in parse try | 259 TryStatement |
| collection misuse | 227 BinaryExpression, 213 ElementAccessExpression, 214 CallExpression |
| discarded outcome | 245 ExpressionStatement |
| discarded pure result | 245 ExpressionStatement |
| uncleared race timeout | 214 CallExpression |
| process exit after output | 307 SourceFile |
| require blocking standard streams | 307 SourceFile |

The speed requirement is only partially implemented. ParseNode in the shared
stage1/typescript/parser/nodes.ts still exposes readonly kind: string and no
numeric kind field. Today's shared rule interfaces invoke run(), not a callback
receiving a node and numeric kind. Consequently, existing implementations still
read kind strings and retrieve nodes; the declarations do not change that driver
behavior and are not a performance fix. Converting the parser and driver is
outside this unit's explicitly restricted territory. No string-to-number adapter,
per-rule checker kind queries, or new shared-file edits hide that missing API.

The complete react-hooks/purity, react-hooks/refs and
react-hooks/preserve-manual-memoization analyses remain blocked by native JSX
parsing and the native React HIR/SSA/reactive memoization pipeline. The renewed
prerequisite test shows production Go reports on all three valid TSX controls:
purity fails native parsing at DotToken 38, preserve-manual-memoization fails at
Identifier 118, and refs parses JSX as TypeAssertionExpression with no JSX node.
This probe is a passing reproduction of blockers, not full-rule parity. The
existing ref lattice, purity property propagation and manual-memo scope verdict
components still match Go, sanitized native and emitted JavaScript on Node.

Commands, after source /workspace/adamic-tools/env.sh, output exclusively to logs:

- ADAMIC_TOOLS=/workspace/adamic-tools bash cloud/setup.sh. Timing lines: Go 0s,
  clang 1s, Node 1s, submodules 1s, cache warm 84s, total 84s; nproc 5, quota 4 cores.
- go test ./stage1/cohere/typeaware -run '^TestWave30NumericListeners$' -count=1
  -timeout 15m -v.
- go test ./stage1/cohere/typeaware -run '^TestWave30(AgreementAndMutants|NextAgreementAndMutants|ThirdAgreementAndMutants|ProcessOutputStateAgreement|ProcessGraphAgreement|ProcessExitAgreement|BlockingStreamsAgreement|ReactComponents|ReactPrerequisites)$'
  -count=1 -timeout 30m -v. ADAMIC_TYPESCRIPT_SOURCE points to the pinned
  /workspace/wave-30-typescript checkout. Each WAVE_30, NEXT, THIRD and PROCESS
  compiler/repository manifest variable points to /workspace/wave-30-compiler.manifest
  and /workspace/wave-30-repository.manifest respectively. No corpus was skipped.
- go vet ./stage1/cohere/typeaware; git diff --check.

The eight existing completed behavior ports agree on findings, fixes and
suggestions for controls and the frozen 77-root compiler and 287-root repository
corpora, with identical sanitizer runs and released-handle failures. Their corpus
finding counts are zero; the positive controls prove nonzero decisions. The
existing full stream evidence remains in validation-wave-30-landing. This pass
uses automatically cleaned scratch outputs and preserves complete test logs,
including byte counts, mutant catches, corpus timings and safety checks, in
validation-wave-30-listeners.

All new mutants alter the first numeric listener, compile, and exit 0 with empty
stderr. Only independent Go comparison catches them:

| Listener mutant | Change | First differing byte |
| --- | --- | ---: |
| ISO date cut | 214 to 213 | 37 |
| callback parse try | 259 to 258 | 86 |
| collection misuse | 227 to 226 | 123 |
| discarded outcome | 245 to 244 | 168 |
| discarded pure result | 245 to 244 | 209 |
| race timeout | 214 to 213 | 251 |
| process exit | 307 to 306 | 296 |
| blocking streams | 307 to 306 | 346 |

Existing mutants again caught only by Go comparison: collection boundary byte
430, outcome range 27672, pure range 1119, throwable graph fork 988, process exit
write state 78, blocking stream write state 107, refs convergence identity 86865,
purity container guard 1368, memo pruning condition 435, ISO range 80, callback
range 4166, timer range 89 and catch-sensitive output state 5. Path lengths change
serialized offsets between scratch runs; each mutant still changes the same
judgment or finding as the prior recorded proof.

Whole-process native versus Go timings from this run, including program loading:

| Runner | Corpus | Native seconds | Go seconds |
| --- | --- | ---: | ---: |
| ISO and callback | compiler | 1.757062 | 0.296041 |
| ISO and callback | repository | 0.292377 | 0.139698 |
| collection and discarded rules | compiler | 2.773507 | 0.805979 |
| collection and discarded rules | repository | 0.386011 | 0.160146 |
| race timeout | compiler | 1.576854 | 0.279774 |
| race timeout | repository | 0.253496 | 0.114903 |

These are single observations, not benchmark medians or a speedup claim. The
numeric declarations are ready for shared driver integration; native execution
remains slower in these observations. Bridge code is unchanged in this commit;
its current-base gates were already green in the landing evidence. All new
Adamic source files use .a; Go comparison fixtures remain .go.
