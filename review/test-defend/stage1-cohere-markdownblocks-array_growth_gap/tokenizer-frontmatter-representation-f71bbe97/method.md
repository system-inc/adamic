# Code under test and oracle

Starting origin/main: f71bbe97f35b6e18baa19ad459a5ace3b13a4048.

Tokenizer: production TokenizerEvents in tokenizerEvents.ts, input preprocessing in inputChunks.ts, and TokenArena in tokenArena.ts. The row runs source Node, emitted JavaScript, sanitized native and release native within one top-level row. These are not separate twin test rows. The oracle is independent Go cohere plus the pinned Prettier tokenizer, comparing complete event streams and final state.

Front matter: production frontmatter.ts, including whitespace, trimWhitespace, blankPrefix and parseFrontMatter. Its oracle is Go cohere front-matter parsing and the pinned Prettier parser. All serialized fields and blanked content are compared.

Representation probes: Adamic lower.Lower, including arrayMethodArguments in internal/lower/object.go and the native/JavaScript output for accepted probes. Positive results are compared with source Node. Negative NotYet/Refused labels and exact text are hand-written expectations. The production compiler acceptance boundary is mutated, never those labels or the gap sources that supply inputs.

No harness, oracle or test source was mutated. Mutants use the fixed menu: E1 changes a Boolean constant in consume; F1 changes a search bound; P1 flips the rejection bound so two-argument push is accepted and only its first value lowered. One first attempt per subject is sufficient when it supplies a unique bounded catch. No cost row or separate Node/native twin is assigned here. Logged throughput has no pass threshold.

# Coverage limits and semantic leads

Four clean go test -coverprofile runs instrument internal/lower and internal/native. These profiles measure compiler Go execution, not TypeScript statement coverage. The representative tokenizer member is 003; full-family completion is unavailable within 90 seconds. Its Go profile and static port import/caller tracing supply leads; the verdict rests on observed production byte mismatches, not profile exclusivity.

Tokenizer consume final consumed=true is asserted by the full event serialization, beyond the chunks test's preprocessing-only output. Only events_probe.ts constructs TokenizerEvents; mdast imports slicing/serialization helpers but does not execute this state machine. Frontmatter_probe.ts is the only driver importing frontmatter.ts. Other Markdown AST programs import the separate parseFrontMatter.ts port. The search bound skips the newline immediately after a three-character delimiter.

The multi-push branch is covered by RepresentationProbes and not FrontMatterStage. The recorded nested-delimiter scan excludes trailing commas and finds one multiple-argument push: gaps/10_multiple_push.ts. No production Go test-generated multi-argument call was found by rg. Source transport and oracle adapters are not mutated.

# Scope

763 top-level test names at the start, same names as prior audit. Exact current scope, bounded run selectors and completed/unknown results are in JSON artifacts. The whole package baseline timed out at 90.107 seconds with no ordinary assertion failure. All four standalone profiles passed; the bounded clean control passed at 74.389 binary seconds. Full-family mutation also cooked. Results outside completed bounded sets are unknown, so no unbounded package-wide replay is claimed.

Warm env.sh was reused; setup 0 seconds; nproc 5. npm ci stage3/api ran before baseline. Source changes are restored after each run; every standalone diff is against the starting main. Per-mutant build-cache paths prevent stale native products. Logs retain native build phases and timeout details.
