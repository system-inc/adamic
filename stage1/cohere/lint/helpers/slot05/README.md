# Slot 05 helpers

Three public helpers, each in its own `.a` file, isolate shared Go cohere judgments. [CONSUMERS.md](CONSUMERS.md) lists every dependent rule and [readiness.json](readiness.json) retains their remaining dependencies.

- `isIdentifierNamed(nodes, expression, name)` skips only `ParenthesizedExpression` nodes, then compares the decoded identifier text exactly. Assertions, non-null expressions, member accesses, literals and other kinds decline.
- `bindingsOf(nodes, declaration)` returns `{default, namespace, named}` indexes. The default is the local identifier. The namespace is the original `NamespaceImport` node, not its identifier. Named entries are original `ImportSpecifier` nodes, preserving source order. Nil, wrong kinds and side-effect imports return empty bindings. Empty Go nil slices are represented by empty arrays.
- `isComponentBase(nodes, expression, matchesBaseName)` accepts a bare component base or a property access on the fixed `React` namespace. Supply the separately owned Go-compatible `isComponentBaseName` leaf. The receiver uses this slot's `isIdentifierNamed`; the outer expression does not skip parentheses. Computed access declines. The callback keeps the leaf helper in its own owner's file rather than copying it here.

The arena contract is `HelperNode` in `nodes.a`. Index `-1` represents a nil field. Every other index must address an immutable node. `kind` is Go's kind name without its `Kind` prefix; `text` is decoded Identifier text. `expression`, `name`, `importClause`, `namedBindings` and `elements` are exact named AST fields, not guesses from child order. The import bindings retain node identities and never retain node-to-node ownership links. All arena edges are numeric indexes, so parenthesis and import links cannot form ownership cycles.

Go's private identifier helper calls `ast.SkipParentheses` before its nil test, and that routine faults on nil. This port panics explicitly on a missing expression. A malformed arena or cyclic parenthesis chain is also an explicit panic, never a clean lint result. Exact Go runtime crash text is not a supported contract. The other two helpers accept nil as Go does.

The test oracle uses an overlay to expose the real private Go helpers without editing cohere. It reads every nonempty Go string literal from every inventory-listed test file of every consumer, deduplicates the strings, adds discriminating controls, parses with pinned Go cohere's TSX parser, and queries every AST node. This deliberately includes description, option and malformed-source strings as well as fixture source literals. It does not reconstruct dynamically concatenated source, execute full rule findings, or replay external fixture files that the test imports at runtime. Missing consumer coverage refuses the test. The adapter snapshots exact named fields from Go; integration with Adamic's parser is still the shared AST adapter's work.

Identifier queries use fixed names plus every identifier's own decoded name. Import output compares exact default, namespace and named-specifier indexes. Component-base queries use the real Go leaf's answers for every identifier text in each arena; the production helper takes that leaf as an explicit function dependency. Baselines run on Node source through `oracle/node.mjs` and sanitized native. Each mutant must compile, exit 0, emit no stderr and differ from Go output; compilation failures and sanitizer failures do not count as a semantic kill.

Run from the repository root after sourcing the setup environment:

```sh
go test ./stage1/cohere/lint/helpers -run '^TestSlot05' -count=1 -v -timeout=10m > /tmp/lint05-tests.log 2>&1
```

Set `ADAMIC_SLOT05_EVIDENCE` to an existing directory to save per-consumer literal counts and deterministic generated-corpus SHA-256 hashes. Tests write no repository artifacts by default. Every run regenerates the corpus against cohere `715ba94f3608a6500086b1076ce5cb7e51b836db`; pin drift fails the tests. No existing helper, linter entry point, compiler implementation or Go cohere worktree is edited.

## Continuation

[batch2/REPORT.md](batch2/REPORT.md) records two retained Tailwind readers, their oracle checks and consumer dependencies. Factory and dispatcher claim collisions were yielded to earlier reservations. Ahra instructed workers to finish existing claims and stop claiming; no replacement for the third continuation slot was reserved. All continuation code and its independent oracle runner stay inside batch2/.

## Third batch

[batch3/REPORT.md](batch3/REPORT.md) records the next three completed helpers: exact JSX matching, filename normalization and Unicode hook-name recognition. They add 21 dependency removals; conservative accounting across all eight retained slot 05 helpers covers 64 rules and removes four final helper blockers in total.

## Fourth batch

[batch4/REPORT.md](batch4/REPORT.md) records ordered theme iteration, entries and namespace keys, validated against actual Go, Node source, emitted JavaScript and sanitized native. These remove 18 dependency occurrences across six Tailwind rules; none loses its last blocker from this batch alone. Eleven retained slot 05 helpers now cover 64 rules and remove 115 dependency occurrences, including four final helper blockers.

## Fifth batch

[batch5/REPORT.md](batch5/REPORT.md) records theme clearing, order compaction and deletion, with actual Go/source Node/emitted JavaScript/sanitized native agreement and thirteen compiling mutants caught. These remove eighteen dependencies across six Tailwind rules, with no new final blocker removed. Fourteen retained slot 05 helpers cover 64 rules and remove 133 dependency occurrences, including four final helper blockers.

## Sixth batch

[batch6/REPORT.md](batch6/REPORT.md) records fresh theme construction, ordered key resolution and literal value resolution. Actual Go/source Node/emitted JavaScript/sanitized native agree across 21,300 cases; twelve compiling semantic mutants are caught. This removes eighteen prerequisites across six Tailwind rules, with no new final blocker removed. Seventeen retained slot 05 helpers cover 64 rules and remove 151 dependency occurrences, including four final helper blockers.

## Landing cap

[LANDING_REPORT.md](LANDING_REPORT.md) records the clean rebase onto origin/main e8ba3d5 and complete helper-subtree oracle rerun. All seventeen retained helpers remain unchanged; all six packages and forty-one compiling semantic mutants pass. No new helper was reserved during this landing unit.

## Seventh batch

[batch7/REPORT.md](batch7/REPORT.md) records descriptor functional-root registration, theme namespace construction and utility evaluator construction. Three retained helpers add eighteen dependency removals across six rules, with eleven compiling semantic mutants caught in four-way Go comparisons. Current origin/main remains e8ba3d5; no additional helper is reserved.
