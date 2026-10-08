Built type-only adaptations for 51 range inputs, five cache declarations, and seven diagnostic consumer bodies.
Base: ef3141e9b1152ab51b51497f8ce3a2799449c8a3; refusal-table input: d35a81d36fdafccf827bad0f572d311b2a0d4deb.
Apply builds; landing lane PASS; oracle results match main: 106366 passing, one sanctioned API-baseline failure.
Undo mutant restores two census rows; two adaptation guards and two Node witness mutants catch planted changes.
Diagnostics add 75 proven removals; tuple and empty-array follow-ups are in progress. See DIAGNOSTICS.md.

The edit changes only type annotations, preserves source bytes elsewhere, and introduces neither any nor unknown. The range inputs become Readonly<TextRange>, including their existing undefined unions. The sibling 70 checker-symbol audit proves no writes or escapes, following local callees. parameters.json is the explicit selection; evidence/range-survey.json records the wider candidate survey. These are internal functions or local closures.

Five lineMap declarations become readonly number[] | undefined. SourceFile is initialized to undefined in nodeFactory; SourceMapSourceObject starts uninitialized. The cache is subsequently populated. SourceFileLike must also admit the actual empty slot: omitting that declaration caused 38 extra checker diagnostics and changed census eligibility. That experiment is preserved under evidence/rejected-cache-contract and is excluded from the final claim. Cache members are internal; the guards reject a changed initialization contract before any edits.

Counts and provenance

The supplied table actually lives at stage3/refusal-table/TABLE.md. evidence/table-adaptations.json reproduces all 519 reasons marked adaptation, comprising 1708 historical observations. Rows with other dispositions were excluded. The 519 reasons include 252 writable-value-view reasons and 267 other refusal reasons; those 267 have not received a type-only feasibility audit in this unit.

The fresh ordinary census initially measured 637 writable-view sites. Range-only edits removed 21. That preliminary comparison differed by one dependency diagnostic and is not the final controlled measurement. To cover the table's eight large bodies, probe.py adds exactly those owner bodies to a disposable census overlay. This retains the no-output scratch loader and makes no production compiler edits. Run it with UNIT71_TABLE_PROBE=1 and UNIT71_TABLE_TREE pointing to the measured tree. These are latent observations on a checker-rejected program, not claims that every site is reachable.

The final controlled census covers 79 files and 4489 eligible units on both sides: 1174 writable-view sites before, 1112 after, 62 removed and zero added. Both checker diagnostic sets contain 323 entries; one pre-existing error changes only the displayed TextRange spelling to Readonly<TextRange>. No units become eligible or cease being eligible. Normalized before/after JSONL, source hashes, exact removed sites and reason counts are in evidence/final/. The comparison is pinned to current main and identical installed dependencies.

Matching the supplied table's exact reasons across all measured phases gives 5073 observations before and 5020 after, a reduction of 53. Its writable-view subset goes from 755 to 702. Seven exact table reasons disappear entirely and two are partially adapted. Eight table reasons are absent on current main. One table adaptation reason has verified flag writes, but its reported readonly-pos exposure still needs an audit of a narrower view. The other 234 writable-view reasons remain unproved. Full row-by-row reconciliation and current file:line sites are in evidence/reconciliation.json; REMAINING.md lists all 1112 retained writable-view sites.

DiagnosticWithLocation as Diagnostic remains 93 -> 93, as DiagnosticRelatedInformation 58 -> 58, and TupleTypeReference as TypeReference 26 -> 26. These require a coupled proof of aliases, derived types and receiving writes; a blanket readonly change has not been justified. Shared never[] views remain for the same reason. EvaluatorResult is already absent after sibling 70 on this main and receives no credit here. A retained refusal is not classified as a language question merely because this unit lacks a proof. Full requested coverage remains unfinished.

Actual writers left alone

The table's Node -> Mutable<Node> row is marked adaptation. binder.ts:1104:14 clears flags and binder.ts:1112:18 exposes the setter that writes flags at 1113:17. The same reason also occurs at utilities.ts:969:14, utilities.ts:975:10 and factory/nodeFactory.ts:7395:6, where flags are written. All five sites remain. These flag writes do not prove that pos needs to become writable: a view retaining readonly pos and mutable flags may be an adaptation. This row is therefore unclassified, not a demonstrated table error. language-questions/node-flags.a is a strict-TypeScript-valid minimal literal-holder counterexample; witness.cjs checks it with stock TypeScript and runs it in Node. It establishes a genuine flag write, not a counterexample about the reported readonly pos or reachability of a narrow holder inside tsc.

utilities.ts:10645:5 and 10655:5 also genuinely write range positions. language-questions/text-range.a shows a holder whose declared literal position is zero observing one after the generic setter. This ordinary-census writer is outside the table's selected large-body observations and was left untouched. These .a programs are standalone external-oracle witnesses, not new internal/oracle fixtures; counts.md needs no refresh.

Verification and mutants

cloud/setup.sh ran after exporting GOPROXY=https://proxy.golang.org|direct. Timing seconds: Node .066, Go .090, clang .599, Markdown deps 1.613 (install 1.425), submodules 21.354, cache warm 288.734, total 288.872. nproc reported 5; cgroup CPU quota is four and memory limit 16 GiB. Tool environment: /workspace/adamic-tools/env.sh. Go 1.27.1, Node 24.19.0, clang 20.1.8, stock TypeScript 6.0.3.

Commands run, with all test output redirected to logs:

- bash stage3/apply.sh through bash stage3/lane/run.sh /tmp/unit71-complete-lane: apply exit 0; incremental patch 15 files, 50 added, 50 removed; total 79 files, 5178 added, 5152 removed.
- bash stage3/oracle/run.sh on /tmp/unit71-main, and unchanged oracle through the lane on the final tree: both 106366 passing, 1 failing, 0 pending, only api/typescript.d.ts. The raw baseline oracle exits 1 on main too; the lane recognizes that existing sanctioned difference and returns PASS. No references were accepted or rewritten.
- Full suites used four CPUs via sched_setaffinity and NODE_OPTIONS=--max-old-space-size=1536, no suite filter. Default/five-worker attempts exhausted the cgroup and yielded incomplete results, discarded. A first control attempt was interrupted after accidental adaptation of its scratch source; that source was restored and the entire run discarded. The successful control and final runs are clean and complete.
- make_overlay.py, gofmt, go build -overlay, and stage3/census/latent/audit.py: audit passed its continuation, body range, no-IR, production-loader and attribution checks, including planted refusal and attribution mutants. The table probe was then built as a separate scratch executable and run before/after/undo with LATENT_ASSERT_NO_OUTPUT=1.
- node adapt.cjs on the final lane tree: zero files changed. All ten built/local/*.js files and built/local/typescript.d.ts are byte-identical to main; evidence/emitted-identity.json records hashes.
- mutant.cjs removes Readonly from rangeStartPositionsAreOnSameLine.range1 in a disposable tree. Census 1112 -> 1114; emitter.ts:4014:50 and 5058:57 return, with zero other removals. An earlier choice of an emitter helper restored zero measured rows and was discarded, then replaced with this observable mutant.
- guard-mutants.cjs adds range1.pos = 0, then independently changes node.lineMap = undefined! to an array initializer. Both adaptation invocations exit 1 before any source edits; source hashes verify this.
- node witness.cjs and node witness.cjs --node-flags: zero TypeScript diagnostics; Node prints 1 despite the literal-zero holder type. Each --mutant replaces the write with a zero write, prints 0, and fails the output assertion with exit 1.

Logs, verdicts, guard results, idempotence audit, witness outputs and failed witness assertions are retained under evidence/. No compiler edits, full package confirmation runs, public API expansion or runtime source changes were made. This branch is a reviewable partial result, not completion of all Disposition=adaptation rows.

The diagnostic follow-up removes another 75 sites (1112 -> 1037). See [DIAGNOSTICS.md](DIAGNOSTICS.md) for its source audit, mutations, remaining sites and fresh-apply/oracle/lane evidence. The earlier counts above describe the initial 3bc49948 subset.
