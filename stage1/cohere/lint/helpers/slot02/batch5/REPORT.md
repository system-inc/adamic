Built: DesignSystem.HasUtility, collapse.findRoots and collapse.PropertySort in separate .a files.
Commits: claims 3552136 and cb39cbd pushed before code; implementation 562790b; slice-header correction 522b2c4; prior twelve helpers already tested and pushed through 9aff573.
Commands and outputs: setup 28s, nproc 5; corrected isolated batch PASS 70.725s; full helper package PASS 278.643s; filtered uncached input oracle PASS 1.482s; touched-package go vet PASS.
Mutants: four utility lookup, four eager root/predicate and six sort delegation/value/slice mutants compile and run on three backends, then fail the actual Go comparator; every witness is recorded below.
Not covered: full repository gate, integrated rule findings/fixes/suggestions, the separately owned private propertySort port, malformed projections/byte strings and eight unavailable Tailwind live-engine/corpus gates.

# Selection and ownership

Finished and pushed all twelve prior helpers before this batch. Fetched every origin codex/lint-helpers* branch and inspected every claims file. The original comment bundle remains owned in the shared HELPERS.md; all larger named helpers are reserved. The generic seven-consumer strict option decoding and schema validation item is remaining per-rule configuration work rather than an unclaimed shared Go symbol. All three retained helpers tie the highest unclaimed concrete count at six consumers each.

Initially reserved HasUtility, HasVariant and VariantKind in 3552136 at 02:30:20 UTC. A subsequent refresh exposed slot 04’s earlier 8890cfc at 02:29:54 for both variant queries. This slot withdrew both before writing any variant-query code. Only HasUtility was retained. After refreshing all twelve then-visible helper branches and checking every claim, replacement claim cb39cbd reserved findRoots and PropertySort and was pushed before code. No duplicate variant-query file is delivered or counted. A final refresh inspected 18 matching branches; only this slot’s claim names the three retained symbols. Complete snapshots are in evidence/claims.log and final-claims.log. No additional helper is claimed.

Changes stay in slot-owned .a helper/driver files, standalone Go tests, fixtures, reports and this slot’s claim. No shared registration, rule harness, protected compiler file or tracked cohere source is edited. Capture and oracle use temporary Go overlays. New Adamic files are .a. The inherited options_json.ts is consumed unchanged and renamed .a only in temporary mutant copies. No PR opened.

# Observations and contracts

The external oracle invokes pinned Go cohere 715ba94f3608a6500086b1076ce5cb7e51b836db’s actual helper bodies. Runtime capture records 157 unique rule/file/source inputs across every consumer, including dynamically assembled strings and inputs before external-engine skips. Missing consumer coverage against readiness.json fails the test. The real TypeScript parser supplies decoded string/template text; static Go candidate/design-system test string literals provide supplementary controls, without claiming that every reconstructed Go fixture input is captured.

There are 2,633 distinct texts, including source/literal/segment controls and all actual framework utility/variant roots. Actual ParseCandidate yields 2,186 candidates and contributes their utility and compound-variant roots. The oracle directly calls LoadedDesignSystem.HasUtility for all texts, five kinds and five stores: 65,825 verdicts. Repository states include absent maps, nil/empty kind maps, explicit false flags, dual-kind roots, nonstandard-kind registrations and a store built by actual stylesheet ingestion. The oracle exports the real framework static and functional maps, including their negative registrations. Two oracle-process-only controls temporarily add an empty static declaration list and a false functional flag, then restore the globals. They distinguish presence from truthiness without guessing tables or changing tracked files.

HasUtility first tests repository root presence. A present root with an empty/nil kind map suppresses every framework fallback. A repository flag is read as a boolean value, not mere kind-key presence. When the root is absent, static lookup tests framework key presence even for an empty declaration list, while functional lookup tests its boolean value. Other kinds return false except when explicitly declared in the repository. The loader’s maps and generated framework data are supplied as explicit inputs. Nil Go kind maps are projected as empty maps with the same read observations; valid system/map projections are required.

findRoots is called for every text with four predicates: a supplied registered-root set, always true, alternating by invocation count, and always false. The actual private Go function supplies both returned candidates and the full ordered callback trace. Exact matches precede dash prefixes, which are visited longest-first. A registered prefix followed by an empty value stops the dash scan. The repeated @ predicate query preserves Go short circuit behavior even for a stateful predicate. The final @ fallback still runs after that stop, because the actual function does so even where its comment suggests otherwise. Exact-root candidates have no value; the @ fallback can have a present empty value. Unicode slicing occurs only at ASCII boundaries. Malformed Go UTF-8 bytes and isolated UTF-16 surrogates are outside the common text domain.

PropertySort delegates to the separately owned private propertySort traversal through an explicit callback. The oracle computes dependency inputs with that actual private Go function over 1,410 trees: successful CSS parses of collected text, boundary stylesheet controls and every real framework static declaration body. The public Go wrapper is called independently. Only its dependency call site is instrumented, recording call count and original slice identity while invoking the unchanged real traversal. An empty/nil node list has the same empty projection; nonempty list identity remains observable.

The public wrapper returns a Sort value, so the adapter copies both the count and the order slice header. Order is {values, length, capacity}: a fresh header shares only the backing values. The projection requires 0 <= length <= capacity <= values.length, and callers access only the logical length. Actual Go supplies initial backing values, logical length and capacity. Mutation probes change the dependency’s count, alter an order element, then reslice the dependency to length/capacity zero. The returned count and header stay fixed while the element change remains visible. Nil and empty zero-capacity slices normalize to the same empty projection. This is the public adapter’s contract, not a claim that the wider private traversal is ported here.

Expected output is removed from runtime corpus input before Adamic reads it. Dependency results are supplied as inputs to the wrapper seam, as real externally owned prerequisites. Baseline and mutants must agree byte-for-byte on source Node, emitted JavaScript and ASan/UBSan native, then with actual Go. Successful runs exit zero without stderr or sanitizer findings. A compile failure, panic or sanitizer finding is not credited as a semantic mutant kill.

# Validation and corrections

The first driver compile failed because it used a nonexistent OptionValue.bool field. The driver was corrected to consume the existing boolean node text; helpers-initial.log retains that failure. The first isolated baseline with eleven mutants passed in 49.017s. After adding kind-value and single-delegation mutants, a slice-header concern was found: directly sharing a JavaScript array would also share its changing length, while Go copies a slice header. The in-progress full regression was stopped and retained as helpers-full-interrupted.log; it is not counted as passing validation. The corrected adapter and header probe pass the isolated four-way comparison with fourteen mutants in 70.725s. The final mutant implementation replaces the whole return expression directly, avoiding an unreachable-return artifact in the earlier local mutation.

Final full package: PASS 278.643s. It reruns fourteen new compiling semantic mutants and all twenty-one prior ones; all thirty-five must be caught to pass. The filtered uncached input oracle runs all six fixtures with zero cache hits and six probe misses, PASS 1.482s. go vet ./stage1/cohere/lint/helpers passes with an empty log. The full repository gate was not run. Exact commands and outputs are retained in evidence/commands.log and the named logs.

Setup reports Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s, build cache warm 28s, done 28s. nproc prints 5. Toolchain is Go 1.27.1, clang 20.1.8 and Node 24.19.0. Builds source /workspace/adamic-tools/env.sh.

# Mutants and witnesses

All fourteen compile and run successfully on Node source, emitted JavaScript and sanitized native before the independent Go comparison rejects their output. The corrected isolated run records these witnesses; the full log retains the final repeat:

| Mutant | First difference |
|---|---|
| Ignore repository precedence | Line 13,168: utility:false versus Go utility:true. |
| Test repository kind presence rather than value | Line 39,496: utility:true versus Go utility:false. |
| Test static declaration truthiness rather than presence | Line 11,656: utility:false versus Go utility:true. |
| Test functional key presence rather than boolean value | Line 11,662: utility:true versus Go utility:false. |
| Drop exact-root match and its predicate query | Line 65,827: calls:0 versus Go calls:1. |
| Continue instead of stopping at empty value | Line 75,978: roots:2 versus Go roots:0. |
| Drop the final @ fallback | Line 72,358: roots:1 versus Go roots:2. |
| Query @ before checking the root | Line 65,847: calls:3 versus Go calls:2. |
| Invoke sort computation twice | Line 120,864: trace:2:true versus Go trace:1:true. |
| Increment returned count | Line 120,863: sort:1:[] versus Go sort:0:[]. |
| Copy the backing order values | Line 120,873: alias:2:[39] versus Go alias:2:[1039]. |
| Share the slice header | Line 120,874: header:0:0:[] versus Go header:1:1:[1039]. |
| Share the entire computation Sort | Line 120,865: alias:1000:[] versus Go alias:0:[]. |
| Copy the input node list | Line 120,872: trace:1:false versus Go trace:1:true. |

Prior twenty-one mutants rerun: raw-control JSON acceptance, overlapping oneOf acceptance, missing policy interpolation, ignored strict fields; non-Identifier JSX acceptance, lost computed property names, reader cache-key collision; React namespace substring, discarded partial class fields, lost NBSP separator, aliased class slice; dropped standalone factory, dropped bare factory calls, ignored React predicate, math substring matching, wrong calc table entry; empty prefix acceptance, rejected z, misplaced namespace match, ignored underscore skip and lost escaped underscore. Earlier reports give the individual witnesses and backends. These are inherited/previous checks, not newly claimed helpers.

# Capture and remaining limits

Capture workflow is adapted from slot 03’s earlier workflow. Regeneration reproduces both sources.jsonl.gz and coverage.json byte for byte; evidence/reproducibility.log records SHA-256 hashes. Unexpected failures or missing consumers reject regeneration. Tailwind capture exits 1 for eight known unavailable external-installation/empty-corpus gates:

- TestClassOrderFixturesActuallyRan
- TestUnknownClassFixturesActuallyRan
- TestConflictFixturesActuallyRan
- TestConflictingClassesPlacementIsAccountedFor
- TestCanonicalFixturesActuallyRan
- TestCanonicalClassesPlacementIsAccountedFor
- TestUnknownClassesPlacesEveryCorpusClass
- TestClassOrderLiveMatchesTheEngineOverTheCorpus

These failures prevent a passing integrated Tailwind rule gate. They do not block real leaf/helper adapter comparisons on captured and generated controls. No full live-engine parity, rule findings, fixes, suggestion serialization, malformed receiver/panic prose or private traversal port is claimed. No shared harness gap required editing shared files.

# Readiness inference

Each helper removes one listed dependency from better-tailwindcss/enforce-canonical-classes, enforce-consistent-class-order, enforce-consistent-variant-order, enforce-shorthand-classes, no-conflicting-classes and no-unknown-classes. Eighteen entries are removed across six distinct rules; zero final blockers are removed by this batch. RULES.md maps each helper to every consumer. readiness.json lists residual dependencies after this slot’s fifteen delivered helpers only. Other workers’ implementations and integration are not assumed. Helper observations do not imply that these rules are fully ported.
