Statement-piece census unions match the 0442c4a9 whole-file observations exactly.
watch.ts: 315 sites in 8 pieces; transformers/ts.ts: 800 sites in 5 pieces; zero differences.
Base scaling model: 25.13 core-hours plus an unknown censored completion tail.
Largest checker piece timed out at 900.070s, peak RSS 1175036KiB; five other samples completed.
Dropped, duplicated and moved-site mutants are caught; coverage remains 65/79 files.

Task #b2wbbha, on top of 0442c4a9. No production compiler files changed. The
whole-project table stays at 3,760,099 / 10,009,820 TypeScript bytes; none of the
14 remaining files was finished. Plans cover 6,249,721 remaining source bytes,
731 pieces and 6,249,597 bytes of indivisible statements. The 124-byte difference
is structural envelope/trailing trivia, not an added coverage claim.

The union audit reports no missing sites, added sites, changed reasons, changed
depths, removed boundaries or added boundaries for either comparison. It
recalibrated 63 watch and 28 transformer provisional tags before comparing.
Full ledgers: watch-union/UNION.json and ts-union/UNION.json.

| Checker piece | Assigned bytes | Wall seconds | Peak RSS KiB | Exit |
|---|---:|---:|---:|---:|
| 0: createNodeBuilder | 327,341 | 900.070 | 1,175,036 | 124 |
| 1: checkTypeRelatedTo | 171,898 | 401.960 | 1,036,352 | 0 |
| 2: getFlowTypeOfReference | 76,626 | 148.128 | 1,035,632 | 0 |
| 3: inferTypes | 52,940 | 180.023 | 1,037,052 | 0 |
| 6: 34 statements/functions | 42,099 | 206.748 | 1,035,800 | 0 |
| 9: 39 statements/functions | 42,099 | 286.584 | 1,035,676 | 0 |

Piece 0 received its full 900-second limit. Its final checkpoint at 897.513s
was checker.ts:8783:29, with 16,861 selected-source walker nodes and 19,322
lowerer nodes; 0/1 assigned statements completed. The whole-file scanner had
visited 298,509/298,510 nodes. Scanner examination is not a finished body walk,
and the source position is not a covered prefix. Its incomplete record is excluded.

Median measured load/scan/registration cost was 108.778 seconds per checker
piece. The shortest transformer piece was 34.117 seconds. The model deliberately
charges checker prepass cost to every planned piece, then extrapolates body cost
by bytes from the five completed samples. Measured samples retain actual attempt
times; the timed-out sample supplies only a 900-second floor, not a completion
estimate. These scenario extrema are not confidence bounds.

| Cores | Fast scenario minutes | Median scenario minutes | Slow scenario minutes |
|---|---:|---:|---:|
| 14 | 99.2 | 108.7 | 127.9 |
| 32 | 44.3 | 48.2 | 56.7 |
| 64 | 22.1 | 24.2 | 30.5 |

Core-hour scenarios: fast_sample: 22.99, median_sample: 25.13, slow_sample: 29.71. All have an unresolved completion tail.
There is no justified finite forecast for finishing all 14 files. A sweep with
900-second limits has a nominal 182.75 core-hour attempt budget; its 14/32/64-core
ceilings are 795/345/180 minutes, excluding KILL grace and supervisor/publication
overhead. That budget yields attempts, not guaranteed complete coverage.

The planner leaves large containers indivisible: parser.ts has a 416,598-byte
namespace, factory/nodeFactory.ts a 309,876-byte function, emitter.ts a
223,400-byte function, and transformers/es2015.ts a 215,321-byte function. More
workers cannot shorten one such walk. Scaling assumes equivalent cores and enough
RAM; the observed peak alone is about 1.12GiB per active worker before shared
caches/OS headroom, while configured RSS allowances are 3GiB per worker.

Focused logged checks pass: check_piece_union.py twice, audit_progress.py,
audit_piece_timing.py, audit_piece_resume.py, audit_piece_publication.py,
audit_stream.py, isolation.py and actual piece byte comparison. Required union
mutants are caught in both files. Additional mutants catch shifted timing, erased
CPU usage, changed checksums, dropped successful records, piece-as-complete-file
publication, shifted depths, wrong table/coverage arithmetic and altered C/JS bytes.
Resume with 32 workers launched zero binaries and preserved completed records.
No whole-package confirmation or full gate was run.

Normal C remains 6907 bytes (169abd45ca4c225362f18da3c6c8bd8171a9939f51e329b2686e7152c77479ab);
normal JS remains 9930 bytes (918453f11ba043e10346c3dd920e5fba93d72ccca6aaed92f7126c49a808f7f5).
Overlay-off and speculative-flag-on production outputs equal the pinned main
compiler byte for byte, with empty stderr. Setup took 17.336s; nproc=5, quota=4.

See ../../pieces.md for method and commands. SHA256.json inventories every archived
file; compressed observations and logs are lossless. No fixture or counts.md change
is needed.
