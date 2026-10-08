# Step 44: vocabulary wall profile and compiler borrow handoff

1. Same 1,083,820 identifier calls: Go's candidate gate rejects 98.37%; the port walks vocabulary first.
2. Native visits 209,273,669 entries versus Go's 287,641; sampled walk wall is 6.468 s versus 0.050 s.
3. Native wall samples: retain/release 19.62%, equality 20.73%, other strings/regex 16.79%, allocation/free 0.43%.
4. All 77 files / 93 rules remain byte-identical; cold medians are native 34.114 s and Go 7.438 s.
5. Container/field borrowing is ruled; COMPILER-HANDOFF.md contains both per-site ranks and exact 68,639,633-pair compiler77 targets.

## Start with the use-isnan gap

The provenance is now recovered: `lint-rules/facts-wave16` at `850d851d4`. Its checked-in REPORT distinguishes 316 unique upstream cases from the timed typed corpus: 554 files, 1,108 default/indexOf-option rows, 29,171,974 bytes. Native 5.599130126 s versus Go 1.741986491 s is 3.21×. The report records four effective cores, release O2, load before 0.0835 and after 0.4458, and no timing skips. The landing corpus script and file manifest were scratch artifacts, not checked in; the aggregate hash alone cannot recreate those inputs. Preserved originals are in `evidence/boundary/use-isnan-*`. This area's 93-rule executable does not include that facts-branch rule. Therefore the measurements below explain area's costs, but do not attribute the specific facts-branch gap to RC, checker or process startup. The branch distinction is resolved; same-graph attribution remains unmeasured.

Go cohere at `7945d102a6c18dd36adf9114a758ce646e8b2359` implements use-isnan in `cohere/internal/lint/rules/core/use_isnan.go:22`: eight equality/relational operators, three listeners at :114, :142, :161, and five listener-side NaN-reference calls. It requires a checker (:97), but gates symbol queries by spelling. The reference recognizer (:197) handles bare NaN, Number.NaN, Number['NaN'], parentheses, and the last element of a comma expression. Shadow checks consult actual symbols (`core/no_undef_init.go:159`); they do not guess from text. No fixes are offered (:90). Options default switch checking on and indexOf checking off.

Shortest distinguishing shapes: `x===NaN` (comparison), `NaN!==x` (left operand), `switch(NaN){}` (switch), `switch(x){case NaN:}` (case), `a.indexOf(NaN)` (indexOf option on), `x===Number['NaN']` (static property), `x===(0,NaN)` (sequence), `function f(NaN){return x===NaN}` (shadow, silent). These are explanatory reductions; the evidence corpus retains real test sources. Go's own four direct tests cover 8 firing comparisons, 3 firing switches, 3 switch-option cases and 12 silent cases (`use_isnan_test.go:11,45,80,97`). Its corpus test (:24) contains 234 rows, including 73 indexOf-option rows checked again with the option off, and 18 shadowing rows. These Go-only cases are not counted as three-way port coverage. `go test -count=1 -v -run '^TestUseIsNaN' ./internal/lint/rules/core` passed at the oracle pin; `evidence/use-isnan-go.log.gz` preserves every case and option-off check.

## Scope and reproduction

Only this package is changed. The implementation is `encode.ts`; measurement integration changes the import in a scratch copy of lint/main.ts. The production handoff requires the shared owner to change `stage1/cohere/lint/main.ts:2`. Stop at that ownership boundary: no parser, lint harness, bridge, runtime or compiler tracked file is modified.

Read before designing: docs/0.1.md, docs/memory.md (borrowed parameters), docs/devirtualize.md, docs/utf16-views.md, docs/lint-registration.md; lint README, GAPS, PERFORMANCE and CHECKER_REPORT; parser GAPS; tsprinter README, GAPS and TSC_CORPUS; typeaware README/GAPS; existing runtime performance REPORT. Counts in prose may lag the generated registry, so the run preserves its actual descriptors.

Run from the repository root, with Go, clang, Node and Valgrind available. All public checkout operations below fetch source only: no dependencies, installs or project scripts. The compiler build is our repository's authorized build, not a public corpus build.

```sh
export GOMAXPROCS=4
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/public-pins/microsoft__TypeScript-6.0.3
export ADAMIC_SCOUT_RESULTS=/workspace/scratch/perf-scout/run
export ADAMIC_SCOUT_COMPILER=/workspace/scratch/perf-scout/adamic
export VALGRIND_LIB=/workspace/scratch/perf-scout/valgrind/usr/libexec/valgrind
VALGRIND=/workspace/scratch/perf-scout/valgrind/usr/bin/valgrind
go build -o "$ADAMIC_SCOUT_COMPILER" ./cmd/adamic
python3 -B stage1/cohere/perf-scout/scout.py fetch /workspace/scratch/public-pins
python3 -B stage1/cohere/perf-scout/scout.py prepare "$ADAMIC_SCOUT_RESULTS" --corpus "$ADAMIC_TYPESCRIPT_SOURCE"
python3 -B stage1/cohere/perf-scout/scout.py prepare-encoding "$ADAMIC_SCOUT_RESULTS"
python3 -B stage1/cohere/perf-scout/scout.py measure "$ADAMIC_SCOUT_RESULTS" --rounds 5 --valgrind "$VALGRIND"
python3 -B stage1/cohere/perf-scout/scout.py encoding "$ADAMIC_SCOUT_RESULTS" --rounds 5 --valgrind "$VALGRIND"
python3 -B stage1/cohere/perf-scout/scout.py fused "$ADAMIC_SCOUT_RESULTS" --rounds 5
python3 -B stage1/cohere/perf-scout/scout.py fused-profile "$ADAMIC_SCOUT_RESULTS" --valgrind "$VALGRIND"
go test -count=1 -v ./stage1/cohere/perf-scout
go vet ./stage1/cohere/perf-scout
gofmt -d stage1/cohere/perf-scout/scout_test.go
```

Use a fresh run directory when sources change. Preparation rejects a mismatched TypeScript pin and changed source fingerprint, and takes a preparation lock. Set every environment variable: tests fail on absent evidence; none skip. The default public checkout location is `run-directory.parent.parent/public-pins`; adjust run placement accordingly. Reproduction can consume the preserved real Go fixtures in `evidence/go-fixtures.tar.gz`, but preparing runs captures them independently from Go's tests through an owned overlay. No shared harness is edited.

Machine: Linux 6.18.44, AMD EPYC 9V74, five affinity CPUs (0–4), four-core cgroup quota (`400000 100000`), GOMAXPROCS=4, no explicit CPU pinning. Go 1.27.1, clang 20.1.8, Node 24.19.0, Valgrind/Callgrind 3.24.0. Native release: C11, `-Wall -Wextra -Werror -pedantic`, the existing unused-variable/function/parameter and self-assign suppressions, `-O2 -ffp-contract=off -fno-optimize-sibling-calls`; profiling adds `-g` (`internal/native/native.go:78`). Count builds add existing allocation counters and are separate from timing. Printer profile uses the same optimization/FP/tail-call flags, without diagnostic flags. Native lint links the existing Go checker archive. Go builds use normal release defaults, with opt-in CPU pprof overlays. Callgrind Go comparison sets GOMAXPROCS=1 and GODEBUG=asyncpreemptoff=1; the all-rule profile applies these to both sides.

Cold means a new process for every timed batch, including startup, read, parse, rule dispatch, repair, protocol encoding and shutdown; filesystem page caches are warm after byte guards. No OS cache flush or private ahra checkout is claimed. Five rounds rotate executable order; results are medians, RSS is maximum over the rounds, wall/file is amortized batch wall. The tiny `testdata/usage.c` launcher measures each child's wait4 RSS rather than inherited Python RSS; wall includes its uniform fork/exec/wait overhead. CPU sampling and counters run separately from timed rounds. One-minute load: baseline 0.44→1.13; paired encoder 0.58→1.35. Full samples, per-row/per-file walls, output lengths and hashes are preserved in JSON.

## Corpus and full-byte guards

The quiet hundred lives on Kirk's Mac and is not in any repository. GitHub was reachable. All 23 user-supplied public pins were shallow-fetched successfully, source-only (supabase required a retry); together their tracked trees list 127,288 .ts paths. They are recorded in `public-pins.json`. These 23 pins are the accepted corpus for this brief; the private quiet hundred is not an outstanding input request. From each pin select deterministically the shortest tracked non-.d.ts source between 200 and 10,000 bytes; `evidence/public-fixtures.json` records exact paths, commits, lengths and hashes. This smoke sample is intentionally small and not a performance proxy for the quiet hundred.

`evidence/fixture-list.json.gz` retains every lint manifest row, source hash/length and options/recovery fields. `evidence/go-fixtures.tar.gz` preserves 5,205 captured Go source fixtures and all owned syntax fixtures. The actual measured compiler checkout was `/workspace/scratch/typescript-6.0.3`; the command above reuses the identical commit from the fetched public pins. The external TypeScript source pin is 6.0.3 `050880ce59e30b356b686bd3144efe24f875ebc8`; cohere's nested TypeScript compiler/cases is pinned at `d92d9bfee114c80be2c375d72edae966176e3a4f`. The pinned stage3 tsc corpus and TypeScript's compiler sources provide printer fragments via the area's existing independent selector.

| Set | Real files / cases | Rules | Findings | Byte guard |
|---|---:|---|---:|---:|
| owned | 148 / 148 | 90 syntax-rule fixture selectors | 322 | 177,735 |
| upstream | 4942 / 4942 | 90 syntax-rule fixture selectors | 2671 | 13,610,492 |
| compiler | 77 / 77 | no-var (isolate shared driver) | 450 | 11,618,333 |
| public | 23 / 23 | no-var (isolate shared driver) | 4 | 8,702 |
| repository | 480 / 480 | no-var (isolate shared driver) | 0 | 3,257,764 |
| printer | 262 / 38790 | supported statement printer, width 80 | — | 2,323,679 |

Baseline guards cover 30,996,705 output bytes across six sets, including finding IDs, spans, descriptions, primary and extra edits, suggestions and final fixed source. There are 100 syntax-recovery rows in the upstream set; those have the existing findings-only contract, not a claim of formatting invalid syntax. The actual pin has zero unsupported-recovery exclusions in the retained syntax set. The 263 typed captured rows are outside the syntax baseline and are not silently counted as run; all three registered typed rules have explicit live-program pilots, and the separate fused set runs all 93 rules on the same 77 compiler files. Its strict program uses the source-only checkout, with no generated diagnostics or installed host declarations; unresolved imports/diagnostics are possible on both oracle and bridge. It is a lint-parity workload, not a successful upstream compiler self-check.

Printer: 38,790 real supported statement fragments from 262 distinct source files, selected from 300 pinned tsc sources and 77 compiler sources. 290 generated selector probes are excluded from these timings. These are fragments, not 262 fully formatted files; preserved coverage reports zero complete files. The selector records 54,118 unsupported candidates and one file-level parse refusal. Original source labels, code, expected selector records and transport inputs are in `printer-fixtures.json.gz`, `printer-cases.txt.gz`, `printer-coverage.json.gz`. Unported syntax remains a gap, never rewritten to fit.

## Wall and memory, same inputs and rules

Baseline medians; each row is one cold batch. Printer wall/file means fragments grouped by original source, not complete-file formatting.

| Set | Go s | Native s | Node s | Go ms/file | Native ms/file | Go / native RSS KiB |
| owned | 0.021970 | 0.031830 | 0.397897 | 0.1484 | 0.2151 | 34,868 / 14,768 |
| upstream | 0.507219 | 1.056493 | 1.204104 | 0.1026 | 0.2138 | 135,124 / 125,404 |
| compiler | 0.445242 | 1.806365 | 1.578382 | 5.7824 | 23.4593 | 126,472 / 142,424 |
| public | 0.010593 | 0.008142 | 0.321472 | 0.4605 | 0.3540 | 30,260 / 14,524 |
| repository | 0.271167 | 0.756491 | 0.810627 | 0.5649 | 1.5760 | 38,200 / 24,616 |
| printer | 1.139401 | 1.315851 | 1.042532 | 4.3489 | 5.0223 | 22,372 / 12,080 |

## All registered rules: the primary driver result

The fused manifest loads one strict program and runs `all` on each of 77 compiler files, covering the actual 93 descriptors and all three typed rules. Every side produces 28,343 findings and 24,318,733 identical bytes, with no skipped/refused lines. Node replays the same recorded, source-hash-checked queries; native and Go use live programs. All five rounds completed after the environment restart; the killed partial attempt is discarded.

| Side | Cold batch median s | Amortized ms/file | Maximum RSS KiB |
|---|---:|---:|---:|
| Go | 7.265425 | 94.3562 | 646,452 |
| before | 33.947422 | 440.8756 | 1,062,184 |
| after | 33.640997 | 436.8961 | 1,048,352 |
| Node_before | 18.181929 | 236.1290 | 2,432,096 |
| Node_after | 17.733023 | 230.2990 | 2,453,372 |

Allocation counts fall 147,658,681 → 137,941,751 (9,716,930 fewer; 6.58%); frees match on both sides. Retains fall 1,729,382,367 → 1,719,099,181; releases 1,684,385,637 → 1,664,385,047. Peak live allocations stay 6,304,948, regions zero. Full-rule wall improves only 0.90%, with a 1.30% lower peak RSS. The encoder is a real but small whole-driver improvement. It does not remove the largest whole-driver costs; doing that crosses the explicitly forbidden shared ownership boundary. No such change is made here.

Live bridge counters over the full run: program load 308.312 ms, 16,723 queries / 707.248 ms total, C input conversion 2.947 ms / calls 696.107 ms / output conversion 6.022 ms, 2,807,768 fact bytes, native run 34.175 s. Thus the measured bridge/load boundary is roughly a second, not the explanation for the whole 33-second native batch. Post-load Go allocation delta (including profiler machinery) is 252,210,784 bytes / 1,889,396 objects / two collections. CPU pprof samples 31.75 s: 67.21% unresolved native `[scanner]`, 24.31% lost external code, 5.45% libc; it cannot resolve individual native functions. Callgrind supplies that attribution.

All-rule Callgrind uses the same scanner.ts row on both sides, retaining the same 77-root checker program and cold process. These are instructions for one measured file **plus shared program load**, not a per-root parse budget:

| Side | Ir for scanner.ts + program |
|---|---:|
| Go | 4,920,049,900 |
| before | 10,463,526,515 |
| after | 10,220,882,446 |

Top native before self costs: string_equal 6.09%, array_index_of 4.31%, abbreviation 3.30%, retain 2.79%, release 2.73%, release_last 2.67%, array_search_from 2.29%; Go-runtime scanObject 5.27%, findObject 4.10%, tryDeferToSpanScan 3.79%. After has the same leading port costs. Go-runtime GC scheduling differs between independent instruction runs, so the 2.32% total Ir reduction is not attributed entirely to the encoder. Recursive/callback inclusive sums exceed the global summary in this trace; use the checked **self** accounting for rankings, never those inclusive percentages. Raw traces and full summaries are preserved; summary self sums equal each trace total.

One-minute load: fused batch 0.81→1.48; instruction profiles 1.48→1.04. All timed rounds finished before Go oracle-test verification, counters and profiling. Package verification briefly overlaps instruction profiling, which is not used for wall figures.

The generated switch dispatch already groups listeners by kind; do not propose a kind table as if none exists. `stage1/cohere/lint/registry/registry.go:359` nevertheless emits `checker.enter` and a reads array for every listener when a program exists, including syntax listeners; `checker.a:59` replaces rule/reads fields. Read-only word membership in `stage1/cohere/lint/rules/nexus-consistency-no-abbreviated-identifier/abbreviation_vocabulary.a:34` and generic array searches are measurable shared costs too. Changing generator dispatch/scope setup requires its shared owner; changing node ownership requires a language ruling and parser/compiler/runtime owners. Both stop here. The next scout should rank these **self** costs and obtain owner scope rather than generalize the isolated encoder percentage.

## Native attribution with profiles

Raw Callgrind instruction traces and Go CPU profiles accompany the summary (`evidence/compiler`, `evidence/printer`). Self sums must equal the Callgrind summary; inclusive costs overlap and must not be added. Release includes the runtime's final destruction helpers, not just the common release entry. These are observed instructions, not elapsed-time proportions.

| Representative same source | Rows | Go Ir | Native Ir | Native Ir per row |
|---|---:|---:|---:|---:|
| compiler: scanner.ts | 1 | 193,089,571 | 459,433,440 | 459,433,440 |
| printer: core.ts | 523 | 112,306,598 | 108,384,700 | 207,237 |

| Native self category | Compiler Ir (%) | Printer Ir (%) |
|---|---:|---:|
| retain | 15,075,993 (3.28%) | 5,269,940 (4.86%) |
| release | 50,484,803 (10.99%) | 18,667,259 (17.22%) |
| allocation | 22,262,173 (4.85%) | 10,477,996 (9.67%) |
| strings | 182,942,009 (39.82%) | 15,912,512 (14.68%) |
| other | 188,668,462 (41.07%) | 58,056,993 (53.57%) |

Read-only walk traffic: `stage1/typescript/parser/parser.ts:243` returns an owned table element; generated C `adamic_function_456_Parser_node` retains it after `adamic_array_at`, then callers release it. `lint/context.ts:113` delegates, and `lint/lint.ts:115` ancestry plus :124 walk repeatedly look up nodes. A textual census excluding testdata and generated files finds 783 `.node(` sites across 150 lint .ts/.a files; printer has 171 sites in expressions.ts. This proves ownership traffic exists in walks, not that every counted retain is redundant. Compiler batch counters: 68,068,538 retains, 75,040,059 releases; printer: 60,624,749 retains, 53,292,903 releases. Go's fused walker directly passes node pointers, dispatches listeners by kind, and visits each child (`cohere/internal/types/program/walk.go:1331`). The step-41 ruling supplied by @system_adamic permits proven non-escaping read-only node borrows; implementing inference belongs to shared owners.

Allocation: compiler batch has 14,506,755 allocations/frees and peak 731,221 live allocations; printer 13,055,569 allocations/frees, peak 62,100. Both finish balanced, zero regions; these are allocation counts, not byte totals or RSS. Compiler self costs include string_equal 6.76%, string_locate 6.60%, release_last 6.16%, string_slice 6.14%, string_append 5.27%, allocate 3.54%, retain 3.28%. Printer has string_equal 8.24%, release_last 8.13%, retain 4.86%, allocate 4.71%, array_index_of 4.55%, release 4.40%, array_search 3.83%, object_free_children 3.51%. Go's printer uses an AST arena and doc slab (`javascript/printer.go:90`); its output buffer marks the settled length instead of copying (`doc/printer.go:219`). Port `tsprinter/doc.ts:280` collects chunks and :525 joins; `tsprinter/transport.ts:3` does four split/join escape passes. These are next candidates, not changed here.

String conversion: `parser/nodes.ts:26` writes every printable UTF-16 code unit through a one-character slice and append. `lint/main.ts:134,138,144,148,165` calls it for finding fields, edits, suggestions and full fixed source. On scanner.ts, written has 168,846,459 inclusive instructions (36.75%), including its string/runtime callees; its self cost alone is 18,308,247 (3.98%). Go's wire encoder uses a strings.Builder and UTF-16 surrogate encoding (`lint/testdata/oracle.go:23`), avoiding character-by-character concatenation. The runtime already grows uniquely held local strings geometrically (`internal/native/runtime/string_append.c:34`) and indexes long non-ASCII strings (`string_index.c:161`); this scout does not diagnose quadratic concatenation or indexing. The gain is fewer slice/append/allocation calls. This protocol overhead is part of the parity driver, not necessarily Go cohere's production CLI; do not extrapolate its percentage to ahra.

Checker bridge: all three registered typed rules run live against the same Go program as the oracle. Node uses native-recorded, hash-checked transcripts; it validates port logic, not native checker speed. `lint/checker.a:71` creates span/frame keys only for replay/record. The live bridge locks program state and converts file/kind/question via GoStringN before Inspect (`bridge/tsgo/archive/main.go:25,112,120`); native conversion timing is exposed by `internal/native/runtime/tsgo.c`.

Boolean pilot: 2 queries, program load 49.818 ms, query total 0.190 ms, C input 0.002043 ms / Go call 0.185700 ms / output 0.001793 ms, 168 fact bytes; post-load Go-process allocation delta 2,038,552 bytes / 154 objects, zero collections. Prefer-find: 10 queries, load 52.087 ms, query total 0.809 ms, input 0.002934 ms / call 0.800695 ms / output 0.002494 ms, 960 fact bytes; 2,093,552 bytes / 588 objects, zero collections. Redundant-type pilot: zero queries, load 59.139 ms, 2,038,112 bytes / 103 objects, zero collections; its witness is a load control, not evidence of checker-query costs. Native cold wall medians are 59.361, 53.246, 59.674 ms respectively. Pilot pprof has zero samples on the boolean and redundant-type runs; it cannot attribute their CPU. Counters are the usable boundary evidence. The bridge CPU profiler starts after program load and stops at final release (`bridge/tsgo/archive/profile.go:11`). Its Go allocation deltas include profiler bookkeeping; they are not pure query payload allocations or whole-cold-process totals, and zero observed collections applies only to that interval. The native encoder adds no collector; the linked existing Go checker retains its own runtime.

Process start: empty manifest native 4.355 ms, Go 8.843 ms, Node 336.840 ms. This includes imports/runtime/shutdown, not pure exec latency. Sixteen syntax fixture rows started separately cost native 81.268 ms versus 7.194 ms as one existing batch. Guard compares Go/Node/batch bytes after changing only the envelope's per-process case number. It demonstrates process-count sensitivity without attributing the separate facts-branch use-isnan run to it.

## First piece and design

The largest isolated cost removable in this owned package with no language ruling is the lint wire encoder. This is a scope-specific result; full-rule profiles below must be used to rank shared dispatch and checker work. Scan UTF-16 code units as before; emit the unchanged escape for controls, non-ASCII units and backslashes; copy each intervening ordinary span once; return the existing string directly for an all-printable input. Only local counters/result bindings mutate. The input remains unchanged. There is no new syntax, buffer intrinsic, unchecked ownership operation, compiler exception or garbage collector. Existing string slice, concatenation and conversion semantics suffice. Supplementary characters remain two surrogate escapes, lone surrogates retain the port's current behavior. A rejected array-of-spans trial raised peak live allocations from 731,221 to 890,537; the delivered immediate-append variant restores the original peak.

Paired medians, rerunning Go and both native/Node variants rather than comparing separate-run medians:

| Set | Go s | Native before s | Native after s | Node before / after s | Before / after RSS KiB |
|---|---:|---:|---:|---:|---:|
| owned | 0.019986 | 0.027021 | 0.026900 | 0.381703 / 0.387909 | 14,916 / 14,660 |
| upstream | 0.479553 | 1.054303 | 1.096185 | 1.183548 / 1.177379 | 125,488 / 125,428 |
| compiler | 0.457240 | 1.850594 | 1.475120 | 1.630992 / 1.463112 | 142,564 / 140,456 |
| public | 0.010703 | 0.007419 | 0.006970 | 0.314389 / 0.319210 | 14,524 / 14,636 |
| repository | 0.237729 | 0.772154 | 0.657898 | 0.737748 / 0.724466 | 24,568 / 24,248 |

Representative compiler Ir: 459,433,440 → 325,451,611 (29.16% lower). Compiler batch allocations: 14,506,755 → 5,669,095 (60.92% lower), retains 68,068,538 → 58,668,536, releases 75,040,059 → 56,802,396, peak 731,221 → 731,221, balanced frees and zero regions. Wall improves 20.29% on compiler and 14.80% on repository sources. Compiler peak RSS falls 142,564 → 140,456 KiB, still above Go's 128,404 KiB. Upstream wall regresses 3.97%; fixture-heavy short strings are not a universal win. Own fixture wall is effectively flat. Do not claim the step's faster/leaner goal is met.

## Hard cases and mutants

`testdata/encodingMain.ts` is the shortest executable encoder matrix: empty string, printable `hello`, LF, CR+tab, NUL, backslash, DEL, é, 😀, lone high surrogate, lone low surrogate, mixed `a\b\nc`, and printable space/tilde. Node and sanitized native must match thirteen explicit canonical lines. The UTF-8 Go corpus covers ordinary Unicode; lone surrogate strings are port-contract checks, not a claim that Go UTF-8 strings represent lone UTF-16 units. Remove the backslash condition: both successfully executed Node and native mutants must differ, exposing a wire delimiter break.

The shortest real selected input for each supported statement kind, with original file/offset and Go formatting, is preserved in `evidence/printer-minimal-real.json` (for example `;`, `break;`, `continue;`, `return;`, `debugger;`). These were part of all three measured executions.

The saved-answer guard also checks every real Go/native/Node set. Successful-answer mutants change a byte, drop a row, reorder rows, and truncate each saved real answer; each must be rejected. Empty input is a failure. An instruction profile with self sum seven and summary eight must be rejected. A successful counted run with allocations seven and frees six must also be rejected; every saved workload requires balanced counters. A skip/refusal line must fail, while those words inside serialized source must pass. The all-rule test requires 77 files, 93 rules, five complete rounds on every side and identical output lengths/hashes. No timeout-induced skip or missing-environment skip is allowed. All six new top-level package tests and both overlay tests call t.Parallel. The final package run passed all six tests and all nine Python guard checks; vet and gofmt produced empty logs in `evidence/`. Sanitizers check the encoder and its mutation; release profiles and counters are measured separately.

## Follow-up: per-site shared boundaries

The user supplied the step-41 ruling: “a read-only walk that only reads node references and never stores or returns them needs no retain or release, proven by borrow inference.” This closes the earlier Parser.node language question. This experiment edits generated C in a scratch copy of this package; shared parser, compiler, runtime, bridge and lint files remain unchanged. It proposes no syntax or collector. Original owned-return functions remain intact; three narrow internal borrowed forwarders serve only selected non-escaping callers. All child-field ownership operations remain intact. The lexical filter is a hand-measurement aid, not a production borrow inference proof: owner lifetime and the frozen post-parse phase must still be proved by the compiler.

Census: 643 static boundary sites, comprising 634 node returns, eight encoder calls, one membership primitive. Of these, 562 node sites pass the conservative hand filter. Calls that return, store, pass the node to a helper, or obtain mutable slot pointers remain owned. Every chosen node result and direct alias only feeds field value reads and matching releases; parsing/construction functions are excluded. `selected.json.gz` records exclusions and selected aliases; `hand-borrow.patch.gz` contains the complete scratch-only change. The smallest relevant read is `const kind = context.node(index).kind`; `return context.node(index)` and helper transfer `visit(context.node(index))` are hard cases excluded by the guard. Keeping the release after borrowing is an executed ASan heap-use-after-free mutant.

| Workload (all 93 rules) | Before cold median | Borrow cold median | Pairs removed | Allocation/peak change |
|---|---:|---:|---:|---|
| scanner.ts, same 77-root program | 1.107153 s | 1.075181 s | 1,786,701 | none |
| selected public23 | 0.234949 s | 0.238124 s | 15,953 | none |
| compiler77 | 33.218778 s | 33.540668 s | 68,639,633 | none |

Three rotating release rounds per variant; process startup and live checker are included. Counts and sampled clocks are separate instrumented runs. Full compiler before retains/releases are 1,729,382,367 / 1,684,385,637; after 1,660,742,734 / 1,615,746,004. Allocations/frees stay 147,658,681, peak 6,304,948, regions zero. The equal 68,639,633 reduction is the acceptance target (3.97% of original retains), not all read-only ownership traffic. Scanner executes 252 eligible sites, public23 188. The encoder-only variant is also measured: full median 33.550628 s; its prior small full-rule win did not repeat. No whole-program speed or memory win is established here.

The new native machine record is AMD EPYC 9V74, Linux 6.18.44, affinity cores 0–4, quota 400000/100000 (four effective cores), GOMAXPROCS=4; one-minute load before 0.01, after 1.00. Clang 20.1.8, Go 1.27.1, Node 24.19.0. Release uses C11/O2, debug symbols and disabled sibling-call optimization; the complete command flags and source SHA are in build.json. Instrumented variants add ADAMIC_COUNT; separate sanitizer variant uses O1 and ASan/UBSan. The preceding Callgrind self-instruction profiles remain the instruction evidence; these scoped clock/count probes supplement them and are not new instruction counts.

Time probes sample every 1,024 boundary calls, subtract the separately measured empty-clock median, and estimate totals. Vocabulary full-walk samples every 64 calls. Per-site callee retains/releases and paired caller releases are distinct columns. Sampled nanoseconds describe counted builds, exclude caller-release time, and are noisy near clock resolution; they cannot be substituted for release wall time. Nested scopes overlap and must not be added. `ranked-sites.csv` ranks every executed site by estimated time; raw JSON preserves calibration, sample counts, counts and all before/after rows.

Full compiler vocabulary at `rules/nexus-consistency-no-abbreviated-identifier/rule.a:291` calls `abbreviation` 1,083,820 times: 978,781,510 retains, 975,463,417 releases, sampled estimate 6.769 s. The internal `allowedNames.includes(text)` at `abbreviation_vocabulary.a:35` has zero retains/releases and only a 0.029 s sampled estimate. The roughly 56.6% retain share belongs to the whole vocabulary walk, including field ownership and escaping result construction, not the membership primitive. The step-41 node rule does not justify deleting all of that traffic.

Top full-run eligible node lookup time estimates (each loses one retain plus paired caller release per call):

| Call site | Calls/pairs | Counted estimate before |
|---|---:|---:|
| `lint/lint.ts:117`, ancestry | 2,647,053 | 112.79 ms |
| `rules/typescript-no-wrapper-object-types/rule.ts:14`, prepare | 2,647,053 | 107.51 ms |
| `RuleContext_has#0:node:1`, generated C line in ranked CSV | 2,310,588 | 90.31 ms |
| `rules/no-labels/rule.ts:18`, walk | 2,647,053 | 83.41 ms |
| `nexus-consistency-no-single-line-jsdoc/comment_anchors.a:62` | 2,646,893 | 78.41 ms |

The highest node retain counts are context kind (3,300,944), `all` and `collectListInteriors` (2,984,309 each); their generated C identities and lines are recorded separately rather than guessing source mappings. `Parser.node` is `stage1/typescript/parser/parser.ts:243`, delegated by `lint/context.ts:113`. Go's direct pointer walker is `cohere/internal/types/program/walk.go:1331`; it has no analogous Adamic retain/release calls. Encoder boundary `lint/main.ts:2` feeds eight calls at :51, :134 twice, :138, :144 twice, :148 and :165. Full fixed-source encoding alone executes 77 times, retaining 9,413,396 and releasing 20,355,585 inside the original encoder. The owned span encoder variant and original byte guard remain available; changing the shared import requires its owner's handoff.

The same real fixture inventory is used: Go-owned upstream tests, the pinned TypeScript compiler77, adamic sources and the 23 pinned public repositories. Public23 retains all original recovery flags for the two deliberately invalid TypeScript corpus files; no file is repaired, dropped or silently skipped. Node runs the same port against recorded, hash-checked checker transcripts and is a parity oracle here, not a native checker speed measurement. Borrowed native matches every byte on scanner/public23/compiler77 and on owned148/upstream4942/repository480. Sanitized scanner matches as well. The complete counts, output hashes, three-round samples, mutant failure and original facts-branch report are preserved in evidence/boundary.

Reproduce after the earlier scout.py baseline preparation:

```sh
export GOMAXPROCS=4
python3 -B boundary.py prepare /workspace/scratch/perf-boundary/run --baseline /workspace/scratch/perf-scout/run --archive /workspace/scratch/perf-boundary/checker.a
python3 -B boundary.py measure /workspace/scratch/perf-boundary/run --baseline /workspace/scratch/perf-scout/run --rounds 3
python3 -B vocabulary.py prepare /workspace/scratch/perf-boundary/run --baseline /workspace/scratch/perf-scout/run --archive /workspace/scratch/perf-boundary/checker.a
python3 -B vocabulary.py measure /workspace/scratch/perf-boundary/run --baseline /workspace/scratch/perf-scout/run
ADAMIC_SCOUT_BOUNDARY=/workspace/scratch/perf-boundary/run ADAMIC_SCOUT_RESULTS=/workspace/scratch/perf-scout/run ADAMIC_SCOUT_COMPILER=/workspace/scratch/perf-scout/adamic go test -count=1 -v ./stage1/cohere/perf-scout
```

Read the scripts' --help for the vocabulary command (its interface differs). Rebuild the existing bridge archive with `go build -buildmode=c-archive -o ... ./bridge/tsgo/archive`; no shared files need editing. All package inputs are mandatory and all new top-level Go tests call t.Parallel. Escape/return/slot mutants prove the hand selection can reject unsafe uses; the unmatched-release mutant proves native lifetime failure. Existing answer/drop/reorder/truncation, instruction-count and encoder mutants continue to run.

## Ruling from @system_adamic, step 44

The vocabulary/container/field question is closed. Ruling on scout/44-boundary-sites f9a54c16: a value read from an array element, map entry or field (including a string field) is borrowed with no retain/release only when the compiler proves all three conditions: (1) the owner outlives the read, (2) nothing in the borrow span overwrites/removes the slot, including calls that might do so unless proven not to, and (3) the value is not stored, returned or captured. If any condition cannot be proved, the read remains counted. Compiler borrow inference supplies the proof; there are no hand annotations and no borrowing on faith. The earlier 68,639,633-pair hand experiment is acceptance evidence, not the production pass. This follow-up profiles the existing vocabulary walk before any further hand-borrowing.

No language question remains open for this measurement. The single compiler handoff is [COMPILER-HANDOFF.md](COMPILER-HANDOFF.md), containing every static boundary site's identity, rankings and exact three-workload count targets.

## Vocabulary wall profile, same 77 files and 93 rules

This follow-up measures the unchanged counted vocabulary, with no new hand-borrowing. `vocabulary_wall.py` builds private generated-C probes and private Go overlays in scratch. It never edits a shared rule, compiler, parser, runtime, bridge or harness. Native and Go run the exact original strict 77-root manifest, all 93 registered rules and the same repair/output contract. Every release and probe answer matches all 24,318,733 bytes (SHA 2467d7084b692cecde6b2de7492088871c219f92d479351bca58d95b405d2fe9). Source fingerprints, program config and registered inventory are in the new fixtures.json. No installs or source scripts are used for the 23 accepted public pins; this profile uses the existing compiler77 subset.

Release cold process medians over three rotating rounds: native 34.114374 s, Go 7.438175 s; amortized wall/file 443.044 ms and 96.600 ms. The profiles are separate from those rounds, and the two Callgrind jobs are SIGSTOP-paused during release/probe measurement. Native vocabulary is sampled every 64 calls with CLOCK_MONOTONIC; Go uses time.Now/Since at the same stride. These estimates include clock/wrapper overhead and are not exact elapsed sums. Native additionally samples the active walk PC every millisecond with a Linux SIGEV_THREAD_ID timer directed at the native main thread: 6,564 samples, no overflow. No signal, logging or entry counter runs in the untouched release binaries.

| Actual workload | Native | Go |
|---|---:|---:|
| Identifier calls seen by rule / gate | 1,083,820 | 1,083,820 |
| Calls reaching vocabulary | 1,083,820 | 15,846 |
| Measured vocabulary entry visits | 209,273,669 | 287,641 |
| Sampled total walk wall | 6.468032 s | 0.049672 s |
| Amortized ns per visited entry | 30.907 | 172.688 |
| Amortized ns per vocabulary call | 5,967.810 | 3,134.681 |

The per-entry numbers divide whole walk time by visits, including membership, maps, message construction and return costs; they are not individual entry latency measurements. Native visits are whole/early/suffix/late/segment: 42,067,126 / 42,011,003 / 41,883,819 / 41,688,974 / 41,622,747. Go visits in four filtered lists are 66,672 / 155,703 / 61,366 / 3,900; whole-word lookup is a map query, not an entry visit. This difference is part of the algorithm, not a claim of equivalent per-entry work.

The Go gate at `cohere/internal/lint/rules/nexus/consistency_no_abbreviated_identifier.go:138` runs before foreign/import and framework skips. Its derived lookups are `abbreviation_gate.go:83`. Of 1,083,820 calls it admits 17,651 (1.63%); subsequent skips leave 15,846 vocabulary calls. The separate gate timer estimates 0.262 s; CPU pprof sees 0.12 s in gate stacks. This sampling difference is retained, not averaged away. Go `abbreviation_vocabulary.go:289` starts with allowed-name/whole-word maps, then ordered filtered lists at :298, :307, :326 and :337. Native `rule.a:291` calls abbreviation before skips. `abbreviation_vocabulary.a:40` scans every word, and :59–63 repeatedly scans the entire table for all four phases. Gate placement plus phase-specific lists is the principal observed difference in work volume. An equivalent gate must admit every vocabulary finding: false positives cost work, false negatives silently lose findings. Do not equate these observations with a measured native gate speedup.

Native sampled wall shares are exclusive PCs: retain 12.74%, release (including release_last) 6.89%, string comparison 20.73%, other string/regex 16.79%, allocation/free 0.43%, other 42.34%. The walk body alone has 2,746 of 6,564 samples (41.83%), covering loop control, repeated indexed loads, phase tests and call setup. The libc samples were resolved from the loaded IFUNC entries and their exact ELF .eh_frame ranges: 644 in memcmp, four in memcpy. Six samples (0.09%) remain unresolved in the native ELF. Function-size bounds and exact unwind ranges prevent substituting a nearby exported libc symbol for an optimized implementation. The resolved ranges are preserved in wall.json. Native has no observed hash/map samples in this walk. The ~978.8 million retains are counted calls, including immortal-literal reads that runtime skips, not proof of that many reference-count writes (`internal/native/runtime/heap.c:218,334`).

Go's full CPU profile contains only one 10 ms sample in vocabulary-find stacks; it is too sparse for wall percentages by subcategory. It is preserved with its focus output. Scoped timers and exact call/entry counts are the usable time evidence for that small region; instruction/category shares are reported separately below. No CPU profile percentage is presented as a wall fraction.

### Instruction and cache evidence

The first native attempt was interrupted by the environment restart at case 70 (71 files begun), before its final dump; it supplies no usable native counters and is excluded. The rerun uses a private checkpoint wrapper, dumping after each 65,536 completed vocabulary calls, outside the collection scope. The vocabulary function body is byte-for-byte identical; the normal checkpoint-copy run matches all output bytes. No retain/release changes occur. Checkpoints reset event counters, not simulated caches, and their exclusive sums plus the final dump cover the complete workload. The extra scalar counter and dump client request are outside the walk's counters; they perturb surrounding cache context slightly and are disclosed rather than treated as production code.

Full-manifest Callgrind collection uses `--cache-sim=yes --collect-atstart=no --toggle-collect=FUNCTION`. Collection is on only inside the actual vocabulary function and its descendants, while the full program remains instrumented so simulated cache context includes the surrounding rules. No record/replay or altered-rule execution is used for these profiles. GOMAXPROCS=1 and GODEBUG=asyncpreemptoff=1 are set on both profile processes; release/probes use GOMAXPROCS=4. Exclusive self event sums must equal the profile summary for every event; recursive inclusive costs are never summed. `vocabulary_accounting.py` enforces this and a doubled-edge mutant must fail. All 1,083,820 native and 15,846 Go vocabulary calls are accounted for. The prior counted-runtime probe measures 978,781,510 scoped retains and 975,463,417 scoped releases. Callgrind call edges sum to 980,189,039 retains and 975,750,750 releases, exceeding that probe by 1,407,529 and 287,333 respectively. Event collection toggles do not establish equivalently scoped call-edge counters in runtime helpers; preserve this discrepancy rather than using graph calls as semantic borrow-removal targets. A new counted-runtime copy of the exact current source reproduces 978,781,510 / 975,463,417 and every output byte, confirming the discrepancy is in graph accounting rather than a changed input or ownership path. The helper-edge discrepancy remains an instrumentation limitation, not a language question. All event sums, direct vocabulary call totals and bytes are verified. Optimized debug line zero accounts for 753,833,425 release edges and 10,255 retain edges; the handoff labels them unresolved and does not invent a source-site pairing. Exact semantic vocabulary site-pair targets require a separate counter probe before compiler acceptance. Returned ownership also means retains and releases are not interchangeable pair counts.


| Scoped event | Native | Go |
|---|---:|---:|
| Instructions (Ir) | 59,448,034,858 | 479,102,554 |
| L1 instruction misses (I1mr) | 32,222,311 | 4,131,646 |
| L1 data misses (D1mr + D1mw) | 109,368,875 | 3,806,674 |
| LL instruction misses (ILmr) | 5,593 | 2,540 |
| LL data misses (DLmr + DLmw) | 290,857 | 225,460 |
| Instructions per visited entry (amortized) | 284.068 | 1,665.627 |

| Exclusive instruction share | Native | Go |
|---|---:|---:|
| retain | 12.875% | 0.000% |
| release | 12.131% | 0.000% |
| hash/map | 0.000% | 8.828% |
| string comparison | 16.462% | 3.144% |
| other strings/regex | 19.382% | 55.580% |
| allocation/free | 0.393% | 9.824% |
| Go GC/barrier | 0.000% | 9.922% |
| other | 38.757% | 12.703% |

Instruction fractions and wall fractions answer different questions. In particular immortal-string retain fast paths execute instructions without updating a reference count. Go hashing is a subset of its map work, not an extra category. Raw native checkpoint parts, exclusive accounting and per-line instruction costs are preserved; every part and the summed full run pass all-event self/summary checks. The full profiled answer is also byte-identical.

Hardware counters are unavailable: perf_event_open for CPU cycles returns ENOENT. These are simulated misses, not measured hardware misses. The Callgrind 3.24.0 model is I1/D1 32 KiB, 64-byte lines, eight-way; LL 256 MiB, 64-byte lines, direct-mapped. The automatically detected aggregate L3 was rounded by the simulator, with warnings preserved. This is a reproducible model, not a claim about EPYC's per-core physical cache topology or a proof that a hardware miss limits wall time.

Machine: AMD EPYC 9V74, Linux 6.18.44, affinity CPUs 0–4, quota 400000/100000 (four effective cores), no explicit core pinning. One-minute load release/probes 2.00 → 1.09 (lagging after profiler pause); during Callgrind observation 1.98. Initial Callgrind load was not captured; native checkpoint observation load was 1.00 and completed-profile one-minute load was 0.77. Tool versions: clang 20.1.8, Go 1.27.1, Node 24.19.0. Native uses existing C11/O2/-g/-ffp-contract=off/-fno-optimize-sibling-calls and warnings-as-errors flags, linked to the unchanged Go checker archive. No LTO, -march tuning or ownership rewrite is introduced. Go uses normal release defaults. Exact build and profile environments are preserved.

### Fixtures, guards and reproduction

All 77 files and all 93 rules are required. Entry counters run separately from PC samples to avoid including 209 million counter increments in the wall distribution; the native entry-count probe estimates 6.735 s versus the wall-only 6.468 s, a visible instrumentation difference. The Go entry probe reports 0.050142 s versus wall-only 0.049672 s. All probe outputs remain identical. Go's own gate tests also pass on the checked-in 82,368 distinct real ahra identifier spellings: 472 vocabulary reports and 481 gate admissions. This is a spelling fixture, not access to ahra's source tree or a measurement against its 1.944 s cold target. The lost-arm tests preserve elapsedMs, Cwd-style segments, prefixes, suffixes and whole-word names. Existing full-answer mutations and encoder/native lifetime mutants remain in the package. New guards reject missing profiles, changed entry-loop fixtures, edge double-counting, mismatched answers, and wrong compiler-handoff totals. Every new top-level Go test calls t.Parallel; all measurement environments are mandatory.

After the earlier baseline and boundary preparation:

```sh
export GOMAXPROCS=4
python3 -B vocabulary_wall.py prepare /workspace/scratch/perf-vocabulary/run --baseline /workspace/scratch/perf-scout/run --original /workspace/scratch/perf-boundary/run --archive /workspace/scratch/perf-boundary/checker.a
python3 -B vocabulary_wall.py measure /workspace/scratch/perf-vocabulary/run --baseline /workspace/scratch/perf-scout/run --original /workspace/scratch/perf-boundary/run --rounds 3
python3 -B vocabulary_wall.py gate /workspace/scratch/perf-vocabulary/run --baseline /workspace/scratch/perf-scout/run --original /workspace/scratch/perf-boundary/run
python3 -B vocabulary_wall.py checkpoint-prepare /workspace/scratch/perf-vocabulary/run --baseline /workspace/scratch/perf-scout/run --original /workspace/scratch/perf-boundary/run --archive /workspace/scratch/perf-boundary/checker.a --valgrind /workspace/scratch/perf-scout/valgrind/usr/bin/valgrind
VALGRIND_LIB=/workspace/scratch/perf-scout/valgrind/usr/libexec/valgrind python3 -B vocabulary_wall.py profile /workspace/scratch/perf-vocabulary/run --baseline /workspace/scratch/perf-scout/run --original /workspace/scratch/perf-boundary/run --valgrind /workspace/scratch/perf-scout/valgrind/usr/bin/valgrind
```

Use the exact Valgrind commands in profile-environment.json, the baseline lint/oracle and boundary before executable, and the baseline fused/manifest.txt. Native toggle is adamic_function_164_abbreviation; Go toggle is github.com/system-inc/cohere/internal/lint/rules/nexus.(*abbreviationVocabulary).find. Set VALGRIND_LIB to the extracted distribution's usr/libexec/valgrind. Profile output is checked against the same saved Go bytes. All package checks set ADAMIC_SCOUT_RESULTS, ADAMIC_SCOUT_COMPILER, ADAMIC_SCOUT_BOUNDARY and ADAMIC_SCOUT_VOCABULARY; no test skips.

## Next scout

Give [COMPILER-HANDOFF.md](COMPILER-HANDOFF.md) directly to the compiler owner: both ranks, all 643 static identities, 562 selected sites and exact three-workload borrow-count targets fit in that single file. Prove all three ruled conditions before removing container/field ownership; preserve ordinary counted reads whenever proof fails. The node acceptance target remains 68,639,633 pairs, with no claimed cold-wall improvement.

Give the vocabulary owner the gate and phase-volume evidence before any further hand-borrowing. The shared boundary needing a change is rules/nexus-consistency-no-abbreviated-identifier/rule.a:291 (candidate gate before lookup), with its derived vocabulary helpers, ordered forms, superset tests and existing full-byte oracle guards. This branch measures and stops at that boundary. Retains/releases are about one fifth of sampled walk time; allocation is small, and the majority is loop/string work repeated for names Go rejects before walking. Carry the same 77-source manifest, all 93 rules, and the lost-arm real fixtures forward. Shared parser, bridge, runtime, compiler and lint harness remain untouched.

Validation for this follow-up: all eight package tests pass with all four input environments set and no skips; go vet is clean, gofmt reports no files. Full release, sampled, entry, gate, checkpoint and counted cross-check outputs match the Go oracle. Validation output is preserved in evidence/vocabulary/package-tests.txt.
