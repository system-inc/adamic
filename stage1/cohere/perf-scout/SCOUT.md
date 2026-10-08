# Step 44: cold-process profiles and a byte-preserving lint encoder

1. The reported use-isnan 5.60 s / 1.74 s = 3.22× gap cannot be reproduced: the area's 93 registered rules contain no use-isnan port.
2. All 93 rules on 77 same compiler files: native 33.947 → 33.641 s, Go 7.265 s, 28,343 findings and 24,318,733 identical output bytes.
3. The owned encoder cuts isolated syntax instructions 29.16%, wall 20.29% and allocations 60.92%; the full-rule wall gain is only 0.90%.
4. Full-rule native remains larger than Go (1,048,352 / 646,452 KiB RSS) and records 1,719,099,181 retains; shared costs need their owners.
5. Printer fragments are slower but smaller than Go (1.3159 / 1.1394 s; 12,080 / 22,372 KiB); ahra and the private quiet hundred remain unmeasured.

## Start with the use-isnan gap

The quoted figures are reported inputs, not observations on this machine. The port at area commit `9156bf5c579a44d687c9955d13e44f9ad8bbb6f8` registers 93 rules: 90 syntax rules and three typed rules. `evidence/registered.json` is the actual inventory. `stage1/cohere/lint/PIPELINE.md:582` lists use-isnan as future work; there is no matching rule directory or descriptor. Therefore neither Node nor native can execute the purported port, and there is no defensible attribution of those 5.60 seconds to RC, checker, or process startup. The measurement revision, exact manifests, flags, process count and machine are missing. Obtain those before calling any result an explanation of that specific gap.

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

The quiet hundred lives on Kirk's Mac and is not in any repository. GitHub was reachable. All 23 user-supplied public pins were shallow-fetched successfully, source-only (supabase required a retry); together their tracked trees list 127,288 .ts paths. They are recorded in `public-pins.json`. These are the public subset, not the private corpus. From each pin select deterministically the shortest tracked non-.d.ts source between 200 and 10,000 bytes; `evidence/public-fixtures.json` records exact paths, commits, lengths and hashes. This smoke sample is intentionally small and not a performance proxy for the quiet hundred.

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

Read-only walk traffic: `stage1/typescript/parser/parser.ts:243` returns an owned table element; generated C `adamic_function_456_Parser_node` retains it after `adamic_array_at`, then callers release it. `lint/context.ts:113` delegates, and `lint/lint.ts:115` ancestry plus :124 walk repeatedly look up nodes. A textual census excluding testdata and generated files finds 783 `.node(` sites across 150 lint .ts/.a files; printer has 171 sites in expressions.ts. This proves ownership traffic exists in walks, not that every counted retain is redundant. Compiler batch counters: 68,068,538 retains, 75,040,059 releases; printer: 60,624,749 retains, 53,292,903 releases. Go's fused walker directly passes node pointers, dispatches listeners by kind, and visits each child (`cohere/internal/types/program/walk.go:1331`). Removing RC requires a lifetime ruling and belongs to shared owners.

Allocation: compiler batch has 14,506,755 allocations/frees and peak 731,221 live allocations; printer 13,055,569 allocations/frees, peak 62,100. Both finish balanced, zero regions; these are allocation counts, not byte totals or RSS. Compiler self costs include string_equal 6.76%, string_locate 6.60%, release_last 6.16%, string_slice 6.14%, string_append 5.27%, allocate 3.54%, retain 3.28%. Printer has string_equal 8.24%, release_last 8.13%, retain 4.86%, allocate 4.71%, array_index_of 4.55%, release 4.40%, array_search 3.83%, object_free_children 3.51%. Go's printer uses an AST arena and doc slab (`javascript/printer.go:90`); its output buffer marks the settled length instead of copying (`doc/printer.go:219`). Port `tsprinter/doc.ts:280` collects chunks and :525 joins; `tsprinter/transport.ts:3` does four split/join escape passes. These are next candidates, not changed here.

String conversion: `parser/nodes.ts:26` writes every printable UTF-16 code unit through a one-character slice and append. `lint/main.ts:134,138,144,148,165` calls it for finding fields, edits, suggestions and full fixed source. On scanner.ts, written has 168,846,459 inclusive instructions (36.75%), including its string/runtime callees; its self cost alone is 18,308,247 (3.98%). Go's wire encoder uses a strings.Builder and UTF-16 surrogate encoding (`lint/testdata/oracle.go:23`), avoiding character-by-character concatenation. The runtime already grows uniquely held local strings geometrically (`internal/native/runtime/string_append.c:34`) and indexes long non-ASCII strings (`string_index.c:161`); this scout does not diagnose quadratic concatenation or indexing. The gain is fewer slice/append/allocation calls. This protocol overhead is part of the parity driver, not necessarily Go cohere's production CLI; do not extrapolate its percentage to ahra.

Checker bridge: all three registered typed rules run live against the same Go program as the oracle. Node uses native-recorded, hash-checked transcripts; it validates port logic, not native checker speed. `lint/checker.a:71` creates span/frame keys only for replay/record. The live bridge locks program state and converts file/kind/question via GoStringN before Inspect (`bridge/tsgo/archive/main.go:25,112,120`); native conversion timing is exposed by `internal/native/runtime/tsgo.c`.

Boolean pilot: 2 queries, program load 49.818 ms, query total 0.190 ms, C input 0.002043 ms / Go call 0.185700 ms / output 0.001793 ms, 168 fact bytes; post-load Go-process allocation delta 2,038,552 bytes / 154 objects, zero collections. Prefer-find: 10 queries, load 52.087 ms, query total 0.809 ms, input 0.002934 ms / call 0.800695 ms / output 0.002494 ms, 960 fact bytes; 2,093,552 bytes / 588 objects, zero collections. Redundant-type pilot: zero queries, load 59.139 ms, 2,038,112 bytes / 103 objects, zero collections; its witness is a load control, not evidence of checker-query costs. Native cold wall medians are 59.361, 53.246, 59.674 ms respectively. Pilot pprof has zero samples on the boolean and redundant-type runs; it cannot attribute their CPU. Counters are the usable boundary evidence. The bridge CPU profiler starts after program load and stops at final release (`bridge/tsgo/archive/profile.go:11`). Its Go allocation deltas include profiler bookkeeping; they are not pure query payload allocations or whole-cold-process totals, and zero observed collections applies only to that interval. The native encoder adds no collector; the linked existing Go checker retains its own runtime.

Process start: empty manifest native 4.355 ms, Go 8.843 ms, Node 336.840 ms. This includes imports/runtime/shutdown, not pure exec latency. Sixteen syntax fixture rows started separately cost native 81.268 ms versus 7.194 ms as one existing batch. Guard compares Go/Node/batch bytes after changing only the envelope's per-process case number. It demonstrates process-count sensitivity without attributing the missing use-isnan run to it.

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

## Open questions for @system_adamic

- Does Adamic permit borrowing an indexed owned return such as Parser.node for a read-only walk, and what proof of owner lifetime, aliasing, mutation and reentrant callback behavior is required? ParseNode's fields and parser tables are mutable during construction; a readonly binding alone cannot prove lifetime safety. This scout does not choose borrow syntax, semantics or unchecked retain deletion. A ruling and shared compiler/runtime owners are required before exploring that optimization.
- No new language question is raised by the delivered encoder: it uses already-supported valid TypeScript operations. If shared owners decide a direct output buffer or string builder is needed next, its ownership and lifetime contract must be posed as a new open question rather than invented here.

Evidence/provenance questions, distinct from language rulings: provide the exact use-isnan port revision, manifests, native/Go commands and machine behind 5.60/1.74; provide ahra and the private quiet hundred if the 1.944 s target is to be assessed.

## Next scout

Start from the all-rule and syntax profiles, not the absent use-isnan implementation. Ask the lint-main owner to review the single import handoff, with this package's byte guards and upstream regression visible. Carry the exact case sources/options forward. The full-rule primary result is only a 0.90% encoder win; prioritize vocabulary array membership and read-only ownership traffic, then generator scope setup (`registry/registry.go:359`) with its owner. Obtain the missing benchmark revision and ahra manifest, then profile their real cold process and checker workload before proposing ownership changes. Next shared costs are parser node lookup/dispatch strings and printer document/transport allocations. Do not alter parser/bridge/runtime/compiler/harness in this lane, or equate fragment formatting with complete-file coverage.
