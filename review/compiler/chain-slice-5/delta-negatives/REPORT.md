Ruled admission-delta proof passes with 22 Node agreements and eight listed negative witnesses; factory-use.a agrees and is not listed.
Delivery: compiler/chain-slice-5 from 2841bccd; tool dependency compiler/admission-delta is separate and unmerged; final SHAs are reported by the worker.
Commands: complete 1,071-program census plus mandatory revision diff passes, 30 new admissions, zero omissions; native parser RSS observations are below.
Mutants: the type-correct speculative repair agrees cleanly, then retained stopping artifacts are rejected as an unlisted divergence.
Not covered: full adapted TypeScript parser, Program-region lifetime, arbitrary-input admission equivalence, the full gate, or non-Linux platforms.

This follow-up serves Outcome 24 and #wj4pmt1, and the admission proof's #936mrbr ruling. It uses the tool dependency on compiler/admission-delta without merging that unlanded branch. The ledger schema is identical on both branches. Adding an entry is a ruling routed to @system_adamic; the JSON header is its `_comment` field. The eight entries are self-contained. Their SHA-256 hashes cover exact source bytes; any source edit requires Node agreement or a new ruling. The literal type-correct repair hashes are forbidden witness keys.

The ordinary admission-delta stream comparison is exactly:

```go
return a.Error == "" && b.Error == "" && c.Error == "" && a.Exit >= 0 && a.Exit == b.Exit && a.Exit == c.Exit && a.Stdout == b.Stdout && a.Stdout == c.Stdout
```

Thus it compares stdout bytes and exit code, while retaining stderr as evidence without comparing engine text. This is specifically the admission tool's comparison. The general oracle helper `disagreement` also has `case !bytes.Equal(oracle.stderr, native.stderr): return "stderr differs"`; its dedicated factory test records the source TypeError and separately pins both inserted-check diagnostics. No oracle comparison code was edited.

All nine candidate programs were independently rerun on source Node, JavaScript backend and native release, before writing the list. Every command was bounded and output captured to files. Factory-use.a agrees by the tool's convention:

| Backend | stdout | exit | stderr |
| --- | --- | ---: | --- |
| source Node | `true\n` | 70 | `adamic: panic: TypeError: Cannot read properties of undefined (reading 'toUpperCase')\n` |
| JavaScript | `true\n` | 70 | `adamic: panic: unset use failed: field 'text', factory 'createBaseNode', use 'node.text' at {root}/stage3/parser-ahead/rulings/factory-use.a:5:13; expected string\n` |
| native | `true\n` | 70 | same exact inserted-check diagnostic as JavaScript |

It is an ordinary agreement, not a listed witness. The other eight source Node runs exit 0; both backends exit 70 with the ledger's exact whole stderr and equal backend stdout. candidate-observations.json retains all three streams for all nine. There is no candidate needing an additional disagreement ruling.

| Content SHA-256 | Ruling | Violated declaration | Type | Exact exit |
| --- | --- | --- | --- | ---: |
| `a3e359099215288bcda467d841620642a6c5438f187ea34b615a9244e89edb96` | #9wc5q5j | `internal/oracle/testdata/placeholder_nonnull_alias_reset.a:1` | string | 70 |
| `aea38b33c7c7d69d480f36f2d9085d561f6f0e4b066aed3d537f52cfb366677c` | #9wc5q5j | `internal/oracle/testdata/placeholder_nonnull_assignment_result.a:1` | string | 70 |
| `2a584b906643006c31238d1f66c6512368b3ac7f01e40da1e425dbfdd0579bde` | #9wc5q5j | `internal/oracle/testdata/placeholder_nonnull_before_use.a:4` | string | 70 |
| `be446786dd8cff3c75cd47f63d7ecc081ce94188600308b6aa72ce6b999fee6b` | #9wc5q5j | `internal/oracle/testdata/placeholder_nonnull_null_before_use.a:2` | string | 70 |
| `ae34cc94244bc08b07d01e2a3b83a8b6e0e322037099c9b5e8037f6231292af9` | #9wc5q5j | `internal/oracle/testdata/placeholder_nonnull_null_saved_leak.a:2` | string | 70 |
| `8dae05151f30455411da485778ed1ee5f04fcaba9d7719b97283d5b1a926b761` | #9wc5q5j | `internal/oracle/testdata/placeholder_nonnull_return_assignment.a:1` | string | 70 |
| `c8ab5676c003b08e57ff4f9067455b837e5e2597ff5e92d47361a3822144450b` | #9wc5q5j | `internal/oracle/testdata/placeholder_nonnull_saved_leak.a:4` | string | 70 |
| `1209cb3741181e653d4c0d6099e8f03b8284c74328958a7eb2fedb512a9c8487` | #ktz9fek | `stage3/parser-next/speculation/misfit.a:3` | number | 70 |

Exact stop messages, including newline, are in the shared ledger. Seven are the #9wc5q5j placeholder checks; the speculative result check cites #ktz9fek's step-24 parser ruling, as documented in docs/step24-parser-contract.md. `{root}` expands only to the tool's absolute head checkout prefix; all other message bytes are pinned. Both backends must match exactly. A wrong message or exit fails even if it looks like another check.

The complete census uses the existing admission-paths.json and verified Git blobs, plus the tool's mandatory changed-.a diff. It passes against main 98008bbb with 30 newly admitted: 22 agree and eight are accepted as listed witnesses, with zero omissions. All classifications complete without compiler failure or timeout. The results retain `agree=false` on listed stops, separately from `negative_witness=true`. Source Node's own engine failure on factory-use.a is not disguised as successful completion.

Commands:

- `python3 review/compiler/chain-slice-5/delta-negatives/write-manifest.py > .../manifest.json`
- `/tmp/delta-negatives-tool --base origin/main --head compiler/chain-slice-5 --manifest .../manifest.json --negative-witnesses cloud/admission-corpus/negative-witnesses.json --workers 4 --timeout 45s --compile-timeout 30s --json`, launched from the tool checkout with the slice ledger's absolute path.
- The same command with the detached repaired head and `--head-binary /tmp/delta-negatives-stopping-compiler`, required to exit 1.
- `go build -o /tmp/delta-negatives-{main,slice}-adamic ./cmd/adamic`; main compiler production sources are unchanged on the tool branch, verified by Git diff.
- Each compiler builds `stage1/typescript/parser/main.ts`; `python3 .../measure-rss.py` runs the two release binaries alternately under a hard 45s timeout.
- Integration's required lane command after committing, with the setup environment sourced.

Condition-3 proof uses slice 5's own stage3/parser-next/speculation/misfit.a. Replace exactly `return 'wrong';` with `return 1;`, keeping the declaration `as number`. Its real compiler control prints `1\n`, exit 0, empty stderr on source Node, JavaScript and native. The repaired SHA-256 is the ledger's type_correct_sha256 and is not listed. A supplied backend wrapper deliberately returns the original witness's stopping JavaScript/native artifacts only for this path. That mutant still compiles and executes, then produces empty stdout and exit 70 with the speculative-check diagnostic in both backends, against Node's `1\n` and exit 0. The tool returns `verdict=fail`, `agree=false`, `negative_witness=false`, accepted_witnesses=7; the other 22 still agree. type-correct-mutant.json and its exit file prove rejection. This is a compiler-artifact mutant, not a build-warning kill. The source is saved as speculation-type-correct.a.txt; no compilable Go probe or new oracle fixture is added, so counts.md needs no new row.

Peak RSS, observed with Linux wait4 ru_maxrss (KiB), for the largest available pinned native whole-file parse: 40 compiler .ts files, 7,903,925 source bytes, largest checker.ts. Both versions return 737,937 reachable nodes, exit 0, empty stderr, in all three alternating runs. This is the existing stage-1 TypeScript parser port, not the still-blocked adapted upstream TypeScript source parser. GNU time is unavailable; measure-rss.py uses a bounded timeout process and wait4's per-child peak measurement.

| Run | main peak RSS KiB | slice 5 peak RSS KiB |
| --- | ---: | ---: |
| 1 | 135444 | 140148 |
| 2 | 135404 | 140108 |
| 3 | 135392 | 140228 |

The added write_order pointer is eight bytes per adamic_object on this 64-bit target. The median observed peak rises by 4,744 KiB; no layout change or attribution to pointer bytes alone is claimed. Allocation size classes and other slice changes can affect the peak. parser-rss.json, parser-inputs.json, the manifest and complete run outputs retain the measurements.

First-write rescan: adamic_object_publish returns immediately when write_order is NULL or the slot is already published. A first write on a construction object scans all shape fields; publishing all n fields costs O(n squared). Emitted C for all 30 newly admitted programs shows construction shapes of at most five fields, including four-field NodeArray metadata. The measured whole-file parser's largest ordinary shape has 34 fields, but its C has zero adamic_object_construct calls, zero adamic_node_array_new calls, and zero adamic_object_publish calls. Its large ordinary shapes do not enter the rescan. Only parser-node-sized construction shapes do so in this measured scope. There is no universal size cap: a future large incomplete shape could take the quadratic path. publication-shapes.json records the source-by-source evidence; no runtime algorithm is changed.

At adamic_array.metadata the new one-line comment states that it is never retained and never escapes as a value: retaining then releasing its zero-count header frees interior array storage. This is a comment only. @system_adamic_runtime's supplied clearance covers slice 5, with these comment/report follow-ups; no new runtime behavior or layout is introduced.

Tool setup used GOPROXY=https://proxy.golang.org|direct, ADAMIC_GOCACHE_OFF=1 and /workspace/adamic-tools/env.sh. Timings: go/node 0.021s; submodules 0.060s; markdown 0.066s; clang 0.146s; cache off 0.148s; go build 166.134s; test binaries deferred 166.263s; cache warm 166.264s; done 166.293s. nproc=5, quota=4 CPUs. Tool unit leaves are below 60s; their per-leaf seconds and the required TestCallTargetReaders result are on the tool branch. These report/comment additions introduce no test leaf. No cohere code was copied and no whole package/full gate was run.

Final tool dependency: compiler/admission-delta 4e0aab75a16382732b9f22fb53cbd1aa44e73d97 (code snapshot a473fdf7). Its nine new test leaves pass (largest 0.15s in the final targeted run), and changed existing leaves top out at 0.19s. The final tool also forces Node agreement on already-admitted repairs and edits at declaration paths; a path schedules coverage but never grants an exception. TestNegativeAlreadyAcceptedRepairMutant supplies a stopping backend for a corrected source accepted by both compilers and catches the divergence. Tool lane output: lane checks 4.2 s, gofmt/tools on 14 Go files, t.Parallel on one test package, a-check one .a, vet one package. Slice lane output before final evidence: lane checks 10.3 s, gofmt/tools on 68 Go files, t.Parallel on four packages, a-check 22 .a, vet four packages. The final full-census rerun and real mutant use this tool. Final delivery adds only report/evidence after that tested source snapshot.

Final rerun: proof-final.json returns pass with 22 ordinary agreements and eight witnesses; proof-final.exit is 0. The final real repaired-source mutant returns fail with seven witnesses and the unlisted divergence; type-correct-mutant.exit is 1. proof-summary.json pins both tested revisions and the tool dependency.

Landing reconciliation: current main 72ad75ef contains only unrelated lint-test changes since the measured main 98008bbb. Both deliveries merge 72ad75ef before their final lane/push. The final census already uses 72ad75ef and records 22 agreements plus eight witnesses. The RSS baseline's production compiler/runtime sources are byte-identical between those main tips; no measurement or compiler fixture changed. The final delivery after the tested d40702f5 snapshot changes only evidence/reports and inherits that test-only main update. The tool code remains the tested a473fdf7 snapshot.
