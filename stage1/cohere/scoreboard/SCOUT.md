# Step 43 (#csz98j8): all 235,377 files scored, with every divergence reduced

Branch: `scout/43-scoreboard`. Source baseline: `fc4b4bc1a08af292a7e675b1e62b1bf4f794a749`, descended from `origin/area/stage1-lint` at `9156bf5c`. Go oracle: cohere `7945d102a6c18dd36adf9114a758ce646e8b2359`; its TypeScript pin is `d92d9bfee114c80be2c375d72edae966176e3a4f`. Only `stage1/cohere/scoreboard/` changes. The full port execution is source Node, not a native Adamic lint binary.

## Headline findings

1. All **235,377 tracked files** at the **23 requested public pins** ran; **201,196** are scripts. Both inputs are explicitly **dependency-uninstalled**. This is the public subset, not Kirk's entire quiet hundred.
2. **99,626,201 cells: 17,175,641 agree, 11,476 diverge, 82,439,084 blocked.** Per-rule, per-family and per-file answers are published in [results/REPORT.md](results/REPORT.md), the CSVs and compressed file receipts.
3. Every divergence has same-path, replayed, deletion-minimal Go/Node answers and checked port/Go responsibility references: **11,476 cells, 325 distinct witnesses, 35 rules/comparison scopes, 43 inspected causes**; no emission-only attribution remains.
4. Blocks: **80,277,204 unported-rule**, **753,828 checker-fact**, **1,205,021 parser**, **196,905 formatter-refusal**; another **6,126 execution-fault** cells are explicitly separate. Equal failures never count as agreement.
5. All **399 missing rules** are ranked by measured Go findings. `one-var` leads with **921,962**, then `id-length` **581,633**, `no-inline-comments` **181,067**. Incomplete censuses are lower bounds. Package tests, the live planted divergence, vet and the receipt/reduction closure are green.

## Corpus and measured boundary

GitHub was reachable. All 23 repositories were shallow-fetched read-only at the supplied SHAs, verified independently, without dependency installs or source-repository scripts. [public-fetch.json](public-fetch.json) records identities; [requested-all.manifest.json.gz](requested-all.manifest.json.gz) is the exact full input list; [requested-all.inventory.json](requested-all.inventory.json) gives counts by repository and extension. The eight requested extensions contribute `.ts` 127,288; `.tsx` 25,372; `.js` 48,511; `.jsx` 25; `.json` 21,636; `.md` 10,348; `.yaml` 1,239; `.css` 958. All **51 tracked symlinks** use their Git blob contents at the original paths, rather than reading their targets. Every receipt has its source SHA-256.

The earlier 23-probe run and 242,883-file broader-extension inventory remain historical artifacts (`first-run.json`, `fixture-run.json`, `public-all.manifest.json.gz`). They do not supply these numbers. The exact eight-extension denominator is 235,377.

There are **492 Go registry rules**, **93 registered port rules**, and **399 unported rules**. Every script gets the full rule census and registered-rule comparison; every input gets its applicable formatter. The primary denominator is **99,425,005 cells** (492 rules plus full all-fixes on each script, plus 235,377 formatters). The separately identified **90-listener syntax-fixes** comparison adds 201,196 cells, yielding 99,626,201. It exposes **132** fix divergences otherwise hidden by the checker-dependent all-fixes block. Full all-fixes remains blocked where checker facts are missing; syntax-fixes does not certify typed or configured cross-rule fix parity.

Registered-rule findings account for **804** divergences across **32** rules; formatting accounts for **10,540** (10,536 TypeScript-family, 4 JSON); syntax-fixes adds 132. The per-family report groups registered and unported Next rules under `next`, and TypeScript rules under `typescript`. Counts include findings on blocked inputs as measured observations, not proof that those inputs were successfully compared. Missing-rule rankings use the independent Go census, with finding-bearing files counted once per rule/file.

The census uses Go's registered default external option decoders; the registered-rule adapters use their existing internal decoded options. These are distinct dialects. The comparisons hold the adapter's rules/options equally on Go and Node; the missing-rule ranking is a default census, not each project's installed lint policy or resolved directory configuration. Checker programs are not fabricated. All absent programs, required/undecodable options, invalid Go syntax and panics remain explicit census statuses. Statuses on an unported rule overlap its primary `unported rule` cell and must not be added again to the blocked total.

## Research: implementation evidence and hard cases

The relevant docs were read before designing the boundary: `docs/0.1.md`, `docs/memory.md`, `docs/lint-registration.md`, stage1 progress/READMEs, and nearby lint, parser, printer, JSON, YAML and formatfiles READMEs/GAPS. At this pin, Go's compiler tests live at `cohere/TypeScript/tsc/testdata/tests/cases`.

| Boundary or observed cause | Port | Go oracle |
| --- | --- | --- |
| Rule discovery / registration | `stage1/cohere/lint/registry/registry.go:98`, dispatch at `:257` | `cohere/internal/lint/registry/registry.go:48`; `rule/register.go:183` |
| Human findings, byte ranges, edits | `stage1/cohere/lint/main.ts:60` and serializer; existing Linter | `cohere/internal/lint/report/report.go:48`, `:62`; existing oracle `stage1/cohere/lint/testdata/oracle.go:75`, `:114`, `:136`, `:173` |
| Comment loss: shortest held input `//` (8,541 cells), `/**/` (447) | `stage1/cohere/tsprinter/expressions.ts:1814`, whole-program entry lacks comment attachment | `cohere/internal/format/javascript/comments_hooks.go:14` |
| BOM loss: shortest held input U+FEFF (614 cells) | `stage1/cohere/tsprinter/expressions.ts:1814` omits native file wrapper | `cohere/internal/format/native/native.go:218`, restoration at `:237` |
| Lost source empty line: e.g. `""\n\ni` (398 cells) | `stage1/cohere/tsprinter/expressions.ts:623`, one hardline between statements | `cohere/internal/format/javascript/print_block.go:103` checks the source empty line |
| Missing lazy attached JSDoc trees (238 cells across witnesses) | `stage1/typescript/parser/parser.ts:1980`, source-file construction | `cohere/TypeScript/tsc/internal/parser/jsdoc.go:139` |
| Empty variable list / `of` recovery (72 cells) | `stage1/typescript/parser/statements.ts:203` | `cohere/TypeScript/tsc/internal/parser/parser.go:1553` |
| Shared configured-rule option transport, owned by lint lane | **`stage1/cohere/lint/main.ts:13`** supplies one JSON bag to all listeners | Individual oracle adapters decode their own options |

[results/witnesses.json](results/witnesses.json) enumerates **every** held input, case count, cause, example path, port site and Go line; [results/divergences-minimized.jsonl.gz](results/divergences-minimized.jsonl.gz) carries every divergence cell's replayed answer, source hash, mismatch signature, tested deletions, predicate count and emission plus causal sites. These references were checked against the pinned files. The 43 causes include import phases, constructor/static-block parsing, speculative arrows, mapped/index signatures, JavaScript type arguments, hashbang comment guards, JSX interior expressions, and lint edit ranges. Paired parser-tree evidence is retained in `results/parser-trees-initial.json`, `parser-trees-second.json` and `additional-trees.json`.

Go preserves the comment and BOM; source Node successfully emits a different formatted answer. These are successful byte divergences, not refusals. Findings, human descriptions, byte positions, IDs, suggestions, edit replacements and fixed source are compared without normalizing order, whitespace, filenames or line endings. File summaries retain exact output SHA-256, exit codes and timings; original divergent stdout answers are in `divergences-*.jsonl.gz`. Invalid UTF-8 uses explicit base64 raw bytes when needed, so JSON replacement characters cannot create false equality.

## Host design, audit and reduction limits

Persistent Go/Node workers invoke the existing implementations through owned overlays/wrappers. They reuse processes and transport, not replacement parsers or lint rules. The Go overlay calls the real native formatters, rule registry, reporting and edit engine. Node imports the existing port. A per-file source override permits same-path reduced inputs and Git symlink blobs; it does not edit source snapshots. CLI/worker fixture agreement and the fused census versus isolated per-rule oracle are checked by the package test.

Receipts stream in bounded gzip chunks, with source hashes, pins, explicit incomplete status and resumable index checks. The full collector's old stderr accumulation was repaired in this package. Separate validated copies preserve stdout, exit code, timings, source and census; the copy tool may omit only stderr, and original archives are retained in scratch. `results/raw-evidence-index.jsonl` records original SHA-256 identities and original/copy byte counts. No measured answers were rewritten. All **235,377 Go records** passed an independent byte-transport audit; **zero** needed corrected answers. Explicit raw byte fields and strings without U+FFFD are lossless; ambiguous old JSON strings were replayed independently.

The initial census request for TypeScript's 3.5 MB `tests/cases/fourslash/reallyLargeFile.ts` hit the request budget. The single-file retry uses the same path, source hash and pins, an explicit larger request/memory budget, and complete Go/Node lint, formatter and 492-rule census receipts. `results/retry-receipt.json` records that exception. Original timeout evidence remains in scratch. The final exporter validates all indices exactly once, the 93-rule order, all 492 census entries per script, the source hashes and dependency labels. [results/closure.json](results/closure.json) proves that the complete score set and complete reduction set have exactly the same 11,476 divergence keys; no missing or extra reductions remain. `results/artifacts.sha256` covers the published evidence.

For each mismatch, reduction retains the original path/script kind/options, deletes only source material, requires Go parse validity for scripts and preserves the observed mismatch signature. Cached witnesses are accepted only after same-input subsequence and same-path replay checks. Every final witness was rechecked against **every single-rune deletion**. This is a deletion fixed point, **not a proof of global shortestness**; replacements, reordering and arbitrary equivalent programs are outside the search. No project/checker semantics are asserted by these uninstalled syntax witnesses.

Compact per-file state vectors follow `summary.json`'s 93-rule order: `A` agree, `D` diverge, `C` checker fact, `P` parser, `F` formatter refusal, `U` unported rule, `E` execution fault. All unported rules apply to every script and their exact per-rule census counts/statuses remain in the summary. `all_fixes` and `syntax_fixes` have separate states. Markdown lacks a whole-file port composition API and is visibly blocked. Unknown process failures remain `execution fault`; they are not misclassified as an explicit formatter refusal.

The owned formatter adapter remains valid TypeScript/Adamic with concrete existing imports and readonly option interfaces. The host tools introduce no language syntax, GC, runtime feature or workaround for a language gap. Shared lint, parser, printer, bridge, runtime and compiler source files are untouched.

## Open questions and rulings for @system_adamic

**No open language questions remain for this piece.** The two original questions are resolved:

* **Host tooling may remain Go.** It measures Adamic binaries through stdout, exit code and timing and never links into the product build. This full receipt labels source Node separately; the earlier `native-node.json` smoke remains separate native build/replay evidence.
* **Dependency-uninstalled snapshots are a different input from installed programs.** Both labels must be present and equal. The manifest validator rejects absent, unknown or differing labels before execution; installed/uninstalled results must never be compared as the same program.

A future language question must be recorded for @system_adamic instead of answered by an implementation workaround. The next native sweep must use the actual Adamic binary and report its originating runtime/compiler refusal through its owner.

## Fixtures and mutants

* All 235,377 actual sources at the pinned 23 repositories, including both public TypeScript compiler test trees, identified by path/hash/pin in the full manifest.
* Go cohere's real `debugger;\n` fixture (`cohere/internal/lint/rules/core/no_debugger_test.go:18`): identical Go/Node finding and deletion/fixed answer.
* Named **`comment-loss`** fixture: Prisma's real empty-module comment, whose held minimal witness is `//`.
* Named **`BOM preservation`** fixture: the cohere-pinned compiler's real `bom-utf8.ts`, `\ufeffvar x=10;\r\n`, whose held minimal witness is U+FEFF. Original CRLF bytes remain held.
* Both named formatter fixtures remain in `testdata/extra-fixtures.json` and are asserted as successful-but-different Go/Node answers in the live test; `rulings-fixtures.json` preserves their earlier nine-cell receipt.
* The live scratch Node mutant runs the unchanged port with exit zero, replaces the debugger description with `planted divergence`, and must become `diverge`. Synthetic mutants cover suggestion replacement, invalid-byte normalization, matching failures, literal `skipped` text inside real output and accumulated stderr. They test the scoreboard, not source-corpus lint rules.

## Reproduction and validation

```sh
source /workspace/adamic-tools/env.sh
go test ./stage1/cohere/scoreboard -count=1 -v -timeout=30m
go vet ./stage1/cohere/scoreboard
python3 stage1/cohere/scoreboard/fetch-public.py /workspace/scratch/scoreboard-public
python3 stage1/cohere/scoreboard/corpus.py /workspace/scratch/scoreboard-public/fetch.json --all --requested-extensions --out /tmp/requested.json
go run ./stage1/cohere/scoreboard --full-tree --manifest /tmp/requested.json --out /tmp/full-receipts --workers 4
go run ./stage1/cohere/scoreboard --verify-go-sweep /tmp/full-receipts --manifest /tmp/requested.json --out /tmp/byte-verification
go run ./stage1/cohere/scoreboard --reduce-sweep /tmp/full-receipts --manifest /tmp/requested.json --out /tmp/reductions
python3 stage1/cohere/scoreboard/audit_sweep.py /tmp/requested.json /tmp/full-receipts /tmp/results --corrections /tmp/byte-verification
python3 stage1/cohere/scoreboard/attribute_sweep.py /tmp/reductions /tmp/results
python3 stage1/cohere/scoreboard/report_sweep.py /tmp/results
```

The individual retry/merge is explicit in `merge_retry.py` and its published receipt; it is not an implicit skipped input. The reducer supports observing newly completed chunks and resume. The final independent export and reduction closure are reusable without rerunning engines.

All **17 top-level package tests** pass, with every required input available and no skips; every new top-level Go test calls `t.Parallel`. `validation/full-tree-tests.txt` includes actual Go/Node workers, isolated census checks, named formatter divergences and the clean-running mutant. `full-tree-vet.txt` is clean; the generated Go worker overlay's vet receipt is `full-tree-worker-vet.txt`. The collector/byte transport race checks pass in `full-tree-race.txt`. gofmt and `git diff --check` are clean. Earlier owned native formatter build/replay evidence is retained, without claiming native coverage for this sweep.

## What the next scout takes

Take the complete score/reduction closure, the 325 replayed witness catalog, the frequency-ranked 399 missing rules and the causal tree evidence. Prioritize fixes through each owning lane; rerun the same corpus and labels to measure the change. Comment attachment, BOM preservation and source blank-line handling dominate successful formatter mismatches. Parser/JSDoc/hashbang witnesses expose multiple lint rules at once.

**`stage1/cohere/lint/main.ts:13` is the lint lane's shared option transport: name it, do not edit it here.** Distinct per-rule option bags for a simultaneous configured run need that owner to change its input contract. `context.ts`/`settings.ts` currently retain the same bag for every listener. Checker facts/program recording need the bridge/lint owners; whole-file Markdown composition needs the formatter owner. This scout stops at those boundaries.

The remaining acceptance scope is Kirk's actual quiet-hundred manifest, selections/options and the other Mac-only snapshots, followed by a separately identified installed-program/native run. Do not treat 23 full public trees, an uninstalled census lower bound, or deletion-minimal witnesses as full hundred-program parity or globally shortest programs.
