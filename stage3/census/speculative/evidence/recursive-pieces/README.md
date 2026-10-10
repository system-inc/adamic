Recursive planner built; watch matches, createNodeBuilder fails the union admission check.
50 KiB threshold: 882 planned pieces across 14 files; checker has 64, createNodeBuilder eight.
Largest planned piece: 49,302 bytes, 189.813s, peak RSS 1036812KiB.
Whole createNodeBuilder: 971.045s, peak RSS 4182624KiB; all 30,499 nodes visited.
Dropped, duplicated, moved and shifted-node mutants caught; full coverage remains 65/79 files.

This continuation of Task #b2wbbha starts at 9421a92c. All edits stay in stage3/census. The recursive mode is experimental and has not passed admission for the remaining compiler files. It is selected explicitly with --threshold; default version-1 planning remains available.

watch.ts retains all 315 sites with zero identity, reason, depth or boundary differences. Its 4980 assigned nodes exclude the structural SourceFile and EndOfFileToken nodes from the 4982-node whole reference; neither produces a body finding.

createNodeBuilder whole has 4773 unique sites; the recursive union has 4761. Missing=63, added=51, changed depths=10; removed boundaries=57, added boundaries=45. Stock, whole and recursive ownership recounts all equal 30,499 nodes. The 399 provisional piece depth tags that the independent boundary union recalibrates are separate from the ten remaining whole-versus-piece depth differences. See [every difference](DIFFERENCES.md) and [full ledger](builder-union/UNION.json).

Observed implementation facts: nestedDeclarations (internal/lower/nested_functions.go:12) preallocates enclosing bindings, allocates named closure targets and sibling links, lowers their bodies and unifies captured environments. Piece preparation supplies enclosing signatures but does not replay that enclosing body state. The speculative walker also retains a shared attempted-node set, so the first lowering context affects subsequent AST probes. Inference: signatures and lexical-binding seeding do not reproduce those contexts after recursive extraction. This is not proof that compiler IR cannot split; it rejects this census implementation. Faithful ancestor declaration/capture snapshots or a deliberately rebaselined canonical-context policy are needed before a big-box recursive census.

| Missing reason | Count |
|---|---:|
| a function value that captures the variable its own initializer declares | 17 |
| a union of differently held members variable a function value captures | 22 |
| a boolean \| undefined variable a function value captures | 8 |
| checked view union arm awaits views-v3: array element kind: TypeParameter[] | 14 |
| an unproven relation from T to TypeReference & T: the source is not assignable to the target | 2 |

| Added reason | Count |
|---|---:|
| an overloaded function as a value | 15 |
| a block-scoped nested function declaration | 8 |
| a value of type (ModuleSpecifierResolutionHost & { getCommonSourceDirectory(): string; }) \| undefined | 21 |
| a union of differently held members variable a function value captures | 1 |
| a function returning T | 4 |
| a value of type U \| undefined | 1 |
| a call returning U \| undefined | 1 |

The threshold is 51,200 bytes. Earlier approximately 42 KiB pieces took 207 and 287 seconds, and a 52,940-byte piece took 180 seconds. A 48 KiB planning trial rejects parser.ts's indivisible 49,243-byte variable initializer; the archived negative witness records that rejection. At 50 KiB every remaining file plans without an oversized residual. Each large function recursively extracts nested function declarations, retaining its own statements together as a residual; namespaces extract contained statements. A residual enters ordinary body lowering with unassigned statements/bodies filtered out. Unsupported atoms fail planning explicitly.

| Measurement | Pieces | Maximum bytes | Wall seconds | Peak RSS KiB | Exit |
|---|---:|---:|---:|---:|---:|
| createNodeBuilder whole reference | 1 | 327341 | 971.045 | 4182624 | 0 |
| createNodeBuilder largest bucket (piece 6) | 8 | 40982 | 185.888 | 1037788 | 0 |
| createNodeBuilder slowest bucket (piece 5) | 8 | 40865 | 200.323 | 1037512 | 0 |
| checker largest planned bucket (piece 16) | 64 | 49302 | 189.813 | 1036812 | 0 |

| createNodeBuilder piece | Bytes | Wall seconds | Peak RSS KiB | Exit |
|---|---:|---:|---:|---:|
| 0 | 40855 | 167.869 | 1035196 | 0 |
| 1 | 40978 | 120.106 | 1035364 | 0 |
| 2 | 40974 | 180.915 | 1034548 | 0 |
| 3 | 40852 | 158.549 | 1036132 | 0 |
| 4 | 40869 | 124.795 | 1037448 | 0 |
| 5 | 40865 | 200.323 | 1037512 | 0 |
| 6 | 40982 | 185.888 | 1037788 | 0 |
| 7 | 40966 | 154.091 | 1035912 | 0 |

The whole reference uses the previous binary and previous unsplit atom under a 3,600-second outer cap, with no 600-second piece deadline. It is backgrounded, inspected repeatedly under five minutes, and bounded by 6 GiB RSS and 12 GiB address space; its Go heap target is 4 GiB. It completes in 16.184 minutes, so no smaller-function fallback is used. Corrected recursive pieces run on three single-CPU workers alongside it under 600-second deadlines, 3 GiB RSS and 6 GiB address-space limits. No piece times out. SUPERVISION.jsonl.gz records checkpoints; early inspections are also in the tool transcript.

The initial body-entry-skipped trial is preserved separately under trials. It is superseded and not used for admission: ordinary residual body entry retains parameter prologues and body-level failures. Assigned-node verification now deduplicates overlapping recursive roots; independent stock ownership recounts and its shifted-count mutant check the result.

Focused logged commands: plan_pieces.cjs (watch, createNodeBuilder, threshold rejection and byte-identical plan replay), plan_remaining.py (14 plans), timed_run.py (whole reference), run_pieces.py (both unions and the largest sample), check_piece_union.py (watch exit 0; createNodeBuilder expected failed admission exit 2), audit_progress.py, audit_piece_resume.py, audit_stream.py, audit_piece_publication.py and isolation.py. Dropped, duplicated, moved and shifted owned-node mutants are caught for both union checks. Progress/source-byte, checksum/dropped-success, stream depth/coverage/recount and piece-as-complete-file mutants are caught by their focused controls. No whole-package test or full gate is run.

Overlay-off identity witnesses retain C 6907 bytes / SHA256 169abd45ca4c225362f18da3c6c8bd8171a9939f51e329b2686e7152c77479ab and JS 9930 bytes / SHA256 918453f11ba043e10346c3dd920e5fba93d72ccca6aaed92f7126c49a808f7f5. Base, ordinary output and speculative-flag production output are byte-identical; altered-output mutants fail. Production source diff is empty. No .a oracle fixture is added, so oracle counts need no refresh.

Setup: node/go .026s, markdown .081s, submodules .090s, clang .242s, shared cache 9.906s, build 25.887s, warm cache 25.996s, total 26.023s. nproc=5 with four-CPU quota; Go 1.27.1, Node 24.19.0, clang 20.1.8.

Complete-file coverage stays 65/79 files, 3,760,099/10,009,820 TypeScript bytes (37.564102%). The full depth table is unchanged. The 14 remaining files are only planned; createNodeBuilder and one checker bucket do not establish complete checker coverage. Recursive results are excluded after failed admission. The older core-hour projection applies to the previous indivisible plan and is not reused for these 882 pieces.

Reproduce with the commands in ../../pieces.md, using the original adapted source hashes and the archived old whole-reference plan. The comparison requires the prior stock.json (SHA256 in PRODUCERS.json), archived at ../continuation/final-result/stock.json.gz. SHA256.json inventories all archived artifacts; .jsonl, .a, generated .go and logs are losslessly compressed. Lane checks run after the commit and before its single push; their final log is in /workspace/b2wbbha-recursive/lane-checks.log.
