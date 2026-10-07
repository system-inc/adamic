# Wave 11 seventh batch

Native production-default ports of `react/jsx-fragments`, `react/jsx-no-undef`,
and `react/no-array-index-key`. The reservations were pushed before implementation.
All Adamic sources are `.a`.

`main.a` opens one checker program and requests one numeric parser/binding snapshot
per source file. `driver.a` indexes callbacks by the parser's numeric SyntaxKind
and hands each listener its decoded node. A rule never requests its handed node
again or compares kinds as strings. `rule.json` and exported `kinds` declare the
production listeners. The index-key rule owns an enter/exit iterator stack through
its SourceFile listener, exactly as Go cohere does; it is dispatched once per file.

`bridge/tsgo/checker/numeric_syntax_bindings.go` supplies parser nodes, byte ranges,
token starts, structural child/role identifiers, and raw symbol declarations.
It supplies no React judgments, scope classifications or lint diagnostics.
`numeric_syntax_bindings.a` owns the decoder, checked graph indices and findings.
The shared checker dispatcher has a single registration case. No shared parser,
harness, registration generator, or compiler implementation was changed.

Run `TestWave11SeventhAgreementAndMutants` in the parent Go package with
`ADAMIC_TYPESCRIPT_SOURCE` pointing to the pinned TypeScript corpus. This isolated
worker test uses unmodified production Go rules through legal cohere build
overlays, including full finding/fix/suggestion serialization. It extracts source
inputs from the pinned Go fixture tables without copying their expectations.
The test additionally checks live production listener kinds, three comparison-only
rule mutants, a raw binding mutant, released handles and normal/sanitized corpora.

The test covers production defaults. Fragment element mode and jsx-no-undef's
allowGlobals option are not exposed by this standalone entrypoint. All three
production rules emit no fixes or suggestions in their default modes. The two
frozen corpora contain no JSX findings, so positive evidence comes from JSX
controls. Shared lint-harness integration and emitted-JavaScript lint-rule
comparison remain outside this worker's authorized shared-file territory.

Four other wave-11 React claims remain parked for missing native HIR/SSA/capture
or return/escape analysis. See the claim and earlier analysis reports; these
implementations do not claim those validators are ported.
