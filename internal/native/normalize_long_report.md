# Long-string normalization, hebkgxh

Streaming normalization, conservative quick check, and verified repetition copying replace the whole-output code-point vector. No heap, append-path, compiler, or Unicode table changes.

On this worker, the largest valid FDFA NFKD case improves from 11.056 s and 3.001 GiB peak RSS to 0.483 s and 1.001 GiB; Node takes 1.957 s and 2.136 GiB in the final matrix. These are best-of-five normalization times, with process-wide peak memory.

Base: `d090af531216ddd3c25a0dede6b82d7c0a6edf76`, fetched from origin/main before branch `codex/normalize-long`. The requested M4 Max observation (main 5d4c801, 29 s and 6.5 GB) is the user's observation, not reproduced here.

## Protocol and limits

Cloud Linux AMD EPYC 9V74, approximately 2.596 GHz, `nproc` = 5, CPU quota 4 cores, cgroup memory limit 16 GiB. Release clang 20.1.8, Node 24.19.0, Go 1.27.1. Setup output: Go ready 0 s; clang ready 0 s; Node ready 0 s; submodules 0 s; gate cache warm 73 s; done 73 s. Environment script: `/workspace/adamic-tools/env.sh`.

Each cell gives best normalization seconds of five and that process's peak RSS MiB, measured with wait4. Normalization timing excludes input construction; RSS includes construction, input, output, and process overhead. Node input repeat ropes are flattened before timing. Native and Node execute serially. Five motifs are repeated to exactly n input code points, with residual ASCII a for the five-point Latin and three-point jamo motifs. ASCII is also the quick-check-passing compatibility case. Native release compilation uses the same default O2 runtime flags before and after. No LTO. The unchanged baseline matrix finished before any repository edit.

Before one-minute system load range: 1.000 to 6.529. After one-minute system load range: 1.001 to 1.367. The initial small baseline trials overlapped the pre-change correctness sweep; large baseline trials mostly ran at load 1.0. Final after measurements ran with no other deliberate workload.

The 100M FDFA NFKC/NFKD inputs would produce 1.8 billion UTF-16 units, exceeding V8/Adamic's 536,870,888-unit limit. These are invalid results, not valid performance comparisons. The first uncapped baseline native NFKC attempt threw RangeError after 73.279 s wall and 10,547,876 KiB peak RSS. Its Node attempt was terminated after 265.152 s and 6,665,824 KiB. Remaining invalid trials were capped at 10 s. `capped` cells report the fastest capped wall time and observed RSS, not a completed best-of-five normalization. `error` cells report the fastest error's total wall time (including input creation) and peak RSS. Uncapped best-of-five error completion is explicitly not covered. Native has a 12 GiB address-space cap; all 590 valid trials in each matrix completed successfully. Native and Node output byte lengths match for every valid trial.

## Measurements

Cell format: **seconds / MiB peak RSS**. Small times use scientific notation to retain precision.

### Before

| Input code points | Motif | Form | Native | Node |
| ---: | --- | --- | ---: | ---: |
| 100,000 | ASCII a | NFC | 2.724e-05 / 5.4 | 0.000164834 / 26.2 |
| 100,000 | ASCII a | NFD | 2.725e-05 / 5.4 | 0.000160348 / 26.2 |
| 100,000 | ASCII a | NFKC | 2.725e-05 / 5.4 | 0.00016907 / 26.2 |
| 100,000 | ASCII a | NFKD | 2.725e-05 / 5.4 | 0.000163833 / 26.2 |
| 100,000 | Latin e + acute + a + dot below + acute | NFC | 0.00377844 / 5.4 | 0.00176657 / 26.5 |
| 100,000 | Latin e + acute + a + dot below + acute | NFD | 0.0022373 / 5.4 | 0.000241858 / 26.4 |
| 100,000 | Latin e + acute + a + dot below + acute | NFKC | 0.00369905 / 5.4 | 0.00173525 / 26.5 |
| 100,000 | Latin e + acute + a + dot below + acute | NFKD | 0.00227723 / 5.4 | 0.000259655 / 26.4 |
| 100,000 | Hangul syllable 각 | NFC | 0.00701749 / 5.4 | 0.000157794 / 26.5 |
| 100,000 | Hangul syllable 각 | NFD | 0.00356708 / 5.4 | 0.00151255 / 27.2 |
| 100,000 | Hangul syllable 각 | NFKC | 0.00643028 / 5.4 | 0.000170323 / 26.5 |
| 100,000 | Hangul syllable 각 | NFKD | 0.00369871 / 5.4 | 0.00146966 / 27.2 |
| 100,000 | Hangul jamo 각 | NFC | 0.00331994 / 5.4 | 0.000297381 / 26.5 |
| 100,000 | Hangul jamo 각 | NFD | 0.00236 / 5.4 | 0.000145656 / 26.6 |
| 100,000 | Hangul jamo 각 | NFKC | 0.0032179 / 5.4 | 0.00030387 / 26.5 |
| 100,000 | Hangul jamo 각 | NFKD | 0.0023674 / 5.4 | 0.000151435 / 26.6 |
| 100,000 | FDFA ﷺ | NFC | 0.00395606 / 5.4 | 0.000152286 / 26.4 |
| 100,000 | FDFA ﷺ | NFD | 0.00233603 / 5.4 | 0.000125426 / 26.5 |
| 100,000 | FDFA ﷺ | NFKC | 0.064419 / 11.3 | 0.00433017 / 32.2 |
| 100,000 | FDFA ﷺ | NFKD | 0.0372134 / 11.3 | 0.00542771 / 32.2 |
| 10,000,000 | ASCII a | NFC | 0.00283184 / 10.4 | 0.0176622 / 72.2 |
| 10,000,000 | ASCII a | NFD | 0.00274237 / 10.4 | 0.0172739 / 72.2 |
| 10,000,000 | ASCII a | NFKC | 0.00277641 / 10.4 | 0.0181618 / 72.2 |
| 10,000,000 | ASCII a | NFKD | 0.0027542 / 10.4 | 0.0175937 / 72.2 |
| 10,000,000 | Latin e + acute + a + dot below + acute | NFC | 0.371395 / 67.8 | 0.178333 / 86.0 |
| 10,000,000 | Latin e + acute + a + dot below + acute | NFD | 0.232667 / 69.7 | 0.0283556 / 62.7 |
| 10,000,000 | Latin e + acute + a + dot below + acute | NFKC | 0.369434 / 67.8 | 0.178529 / 86.0 |
| 10,000,000 | Latin e + acute + a + dot below + acute | NFKD | 0.228587 / 69.6 | 0.028082 / 62.7 |
| 10,000,000 | Hangul syllable 각 | NFC | 0.6479 / 172.7 | 0.0178197 / 62.7 |
| 10,000,000 | Hangul syllable 각 | NFD | 0.351556 / 229.6 | 0.170071 / 177.5 |
| 10,000,000 | Hangul syllable 각 | NFKC | 0.642517 / 172.5 | 0.0168402 / 62.9 |
| 10,000,000 | Hangul syllable 각 | NFKD | 0.355101 / 229.8 | 0.166497 / 177.5 |
| 10,000,000 | Hangul jamo 각 | NFC | 0.329379 / 77.3 | 0.031526 / 75.7 |
| 10,000,000 | Hangul jamo 각 | NFD | 0.226815 / 96.3 | 0.0164056 / 62.7 |
| 10,000,000 | Hangul jamo 각 | NFKC | 0.369475 / 77.3 | 0.0339913 / 75.7 |
| 10,000,000 | Hangul jamo 각 | NFKD | 0.255158 / 96.3 | 0.0171818 / 62.9 |
| 10,000,000 | FDFA ﷺ | NFC | 0.442025 / 96.3 | 0.0174966 / 62.9 |
| 10,000,000 | FDFA ﷺ | NFD | 0.232197 / 96.3 | 0.016329 / 62.7 |
| 10,000,000 | FDFA ﷺ | NFKC | 6.61987 / 1030.9 | 0.537422 / 749.7 |
| 10,000,000 | FDFA ﷺ | NFKD | 3.61763 / 1030.9 | 0.634857 / 749.7 |
| 100,000,000 | ASCII a | NFC | 0.028282 / 96.2 | 0.201078 / 501.4 |
| 100,000,000 | ASCII a | NFD | 0.0276721 / 96.1 | 0.206899 / 501.1 |
| 100,000,000 | ASCII a | NFKC | 0.0325442 / 96.1 | 0.211375 / 501.4 |
| 100,000,000 | ASCII a | NFKD | 0.0281679 / 96.0 | 0.200414 / 501.4 |
| 100,000,000 | Latin e + acute + a + dot below + acute | NFC | 3.80652 / 668.5 | 1.79424 / 635.2 |
| 100,000,000 | Latin e + acute + a + dot below + acute | NFD | 2.27149 / 687.6 | 0.279808 / 406.0 |
| 100,000,000 | Latin e + acute + a + dot below + acute | NFKC | 3.71406 / 668.4 | 1.77589 / 635.2 |
| 100,000,000 | Latin e + acute + a + dot below + acute | NFKD | 2.3141 / 687.7 | 0.280054 / 406.1 |
| 100,000,000 | Hangul syllable 각 | NFC | 6.34455 / 1717.5 | 0.177791 / 406.1 |
| 100,000,000 | Hangul syllable 각 | NFD | 4.35912 / 2289.7 | 1.70647 / 1550.8 |
| 100,000,000 | Hangul syllable 각 | NFKC | 6.39422 / 1717.5 | 0.17714 / 406.0 |
| 100,000,000 | Hangul syllable 각 | NFKD | 4.31395 / 2289.6 | 1.6901 / 1550.9 |
| 100,000,000 | Hangul jamo 각 | NFC | 3.26366 / 763.9 | 0.328923 / 533.5 |
| 100,000,000 | Hangul jamo 각 | NFD | 2.30172 / 954.6 | 0.170791 / 406.0 |
| 100,000,000 | Hangul jamo 각 | NFKC | 3.22439 / 763.9 | 0.320338 / 533.5 |
| 100,000,000 | Hangul jamo 각 | NFKD | 2.32372 / 954.6 | 0.16947 / 406.0 |
| 100,000,000 | FDFA ﷺ | NFC | 3.96118 / 954.7 | 0.187163 / 406.0 |
| 100,000,000 | FDFA ﷺ | NFD | 2.29532 / 954.7 | 0.163579 / 406.0 |
| 100,000,000 | FDFA ﷺ | NFKC | capped 10.169 / 3290.0 | capped 10.224 / 6509.6 |
| 100,000,000 | FDFA ﷺ | NFKD | capped 10.160 / 3257.9 | capped 10.223 / 6509.6 |
| 29,826,160 | FDFA ﷺ | NFKD | 11.056 / 3072.8 | 1.86658 / 2186.7 |

### After

| Input code points | Motif | Form | Native | Node |
| ---: | --- | --- | ---: | ---: |
| 100,000 | ASCII a | NFC | 5.43e-05 / 5.4 | 0.000166257 / 26.9 |
| 100,000 | ASCII a | NFD | 5.43e-05 / 5.4 | 0.00016237 / 27.1 |
| 100,000 | ASCII a | NFKC | 5.43e-05 / 5.4 | 0.000176872 / 26.9 |
| 100,000 | ASCII a | NFKD | 5.43e-05 / 5.4 | 0.000175791 / 26.9 |
| 100,000 | Latin e + acute + a + dot below + acute | NFC | 0.000162762 / 5.4 | 0.00174704 / 27.4 |
| 100,000 | Latin e + acute + a + dot below + acute | NFD | 0.000251353 / 5.4 | 0.000250491 / 27.1 |
| 100,000 | Latin e + acute + a + dot below + acute | NFKC | 0.000167669 / 5.4 | 0.00174555 / 27.2 |
| 100,000 | Latin e + acute + a + dot below + acute | NFKD | 0.000252285 / 5.4 | 0.000261367 / 27.1 |
| 100,000 | Hangul syllable 각 | NFC | 0.000224964 / 5.4 | 0.000158054 / 27.2 |
| 100,000 | Hangul syllable 각 | NFD | 0.000813648 / 5.4 | 0.00155346 / 28.0 |
| 100,000 | Hangul syllable 각 | NFKC | 0.000220397 / 5.4 | 0.000169571 / 27.1 |
| 100,000 | Hangul syllable 각 | NFKD | 0.000820447 / 5.4 | 0.00154607 / 28.0 |
| 100,000 | Hangul jamo 각 | NFC | 0.000231463 / 5.4 | 0.000325332 / 27.2 |
| 100,000 | Hangul jamo 각 | NFD | 0.000359002 / 5.4 | 0.000146888 / 27.2 |
| 100,000 | Hangul jamo 각 | NFKC | 0.000233897 / 5.4 | 0.000306995 / 27.1 |
| 100,000 | Hangul jamo 각 | NFKD | 0.000359603 / 5.4 | 0.000157503 / 27.1 |
| 100,000 | FDFA ﷺ | NFC | 0.000224884 / 5.4 | 0.000154819 / 27.1 |
| 100,000 | FDFA ﷺ | NFD | 0.00019527 / 5.4 | 0.000131586 / 27.1 |
| 100,000 | FDFA ﷺ | NFKC | 0.00149325 / 5.4 | 0.00452042 / 33.0 |
| 100,000 | FDFA ﷺ | NFKD | 0.00150897 / 5.5 | 0.00585068 / 33.1 |
| 10,000,000 | ASCII a | NFC | 0.00552064 / 10.4 | 0.0185746 / 73.0 |
| 10,000,000 | ASCII a | NFD | 0.00553389 / 10.4 | 0.0179185 / 73.1 |
| 10,000,000 | ASCII a | NFKC | 0.00560847 / 10.4 | 0.0186023 / 73.0 |
| 10,000,000 | ASCII a | NFKD | 0.00556205 / 10.4 | 0.017888 / 73.1 |
| 10,000,000 | Latin e + acute + a + dot below + acute | NFC | 0.0155669 / 29.5 | 0.177438 / 86.9 |
| 10,000,000 | Latin e + acute + a + dot below + acute | NFD | 0.0254173 / 16.2 | 0.0274266 / 63.5 |
| 10,000,000 | Latin e + acute + a + dot below + acute | NFKC | 0.0155861 / 29.4 | 0.177406 / 86.7 |
| 10,000,000 | Latin e + acute + a + dot below + acute | NFKD | 0.0255745 / 16.2 | 0.028206 / 63.6 |
| 10,000,000 | Hangul syllable 각 | NFC | 0.0222936 / 29.5 | 0.0185162 / 63.6 |
| 10,000,000 | Hangul syllable 각 | NFD | 0.081392 / 115.4 | 0.171192 / 178.3 |
| 10,000,000 | Hangul syllable 각 | NFKC | 0.0222444 / 29.4 | 0.0174987 / 63.5 |
| 10,000,000 | Hangul syllable 각 | NFKD | 0.0781991 / 115.4 | 0.172066 / 178.3 |
| 10,000,000 | Hangul jamo 각 | NFC | 0.0233062 / 39.2 | 0.0319238 / 76.6 |
| 10,000,000 | Hangul jamo 각 | NFD | 0.0363122 / 29.4 | 0.0174233 / 63.5 |
| 10,000,000 | Hangul jamo 각 | NFKC | 0.0208992 / 39.2 | 0.0323633 / 76.5 |
| 10,000,000 | Hangul jamo 각 | NFKD | 0.0369467 / 29.4 | 0.0177744 / 63.5 |
| 10,000,000 | FDFA ﷺ | NFC | 0.023432 / 29.5 | 0.0185803 / 63.5 |
| 10,000,000 | FDFA ﷺ | NFD | 0.0194458 / 29.6 | 0.0168921 / 63.5 |
| 10,000,000 | FDFA ﷺ | NFKC | 0.158814 / 344.3 | 0.558284 / 750.4 |
| 10,000,000 | FDFA ﷺ | NFKD | 0.159222 / 344.3 | 0.658513 / 750.4 |
| 100,000,000 | ASCII a | NFC | 0.0574856 / 96.2 | 0.210368 / 502.1 |
| 100,000,000 | ASCII a | NFD | 0.0578024 / 96.2 | 0.222301 / 502.1 |
| 100,000,000 | ASCII a | NFKC | 0.0571081 / 96.2 | 0.217424 / 502.1 |
| 100,000,000 | ASCII a | NFKD | 0.0562978 / 96.1 | 0.21123 / 502.1 |
| 100,000,000 | Latin e + acute + a + dot below + acute | NFC | 0.157308 / 287.1 | 1.82089 / 636.0 |
| 100,000,000 | Latin e + acute + a + dot below + acute | NFD | 0.256865 / 153.6 | 0.282024 / 406.9 |
| 100,000,000 | Latin e + acute + a + dot below + acute | NFKC | 0.151489 / 287.1 | 1.81957 / 636.1 |
| 100,000,000 | Latin e + acute + a + dot below + acute | NFKD | 0.252992 / 153.5 | 0.285713 / 406.7 |
| 100,000,000 | Hangul syllable 각 | NFC | 0.224761 / 286.9 | 0.183739 / 406.9 |
| 100,000,000 | Hangul syllable 각 | NFD | 0.794065 / 1145.4 | 1.72329 / 1551.6 |
| 100,000,000 | Hangul syllable 각 | NFKC | 0.228231 / 287.0 | 0.180374 / 406.9 |
| 100,000,000 | Hangul syllable 각 | NFKD | 0.784761 / 1145.4 | 1.71864 / 1551.6 |
| 100,000,000 | Hangul jamo 각 | NFC | 0.213397 / 382.4 | 0.332206 / 534.4 |
| 100,000,000 | Hangul jamo 각 | NFD | 0.372422 / 287.0 | 0.175441 / 406.9 |
| 100,000,000 | Hangul jamo 각 | NFKC | 0.214006 / 382.4 | 0.332557 / 534.2 |
| 100,000,000 | Hangul jamo 각 | NFKD | 0.372326 / 286.9 | 0.180967 / 406.7 |
| 100,000,000 | FDFA ﷺ | NFC | 0.225203 / 287.0 | 0.183662 / 406.7 |
| 100,000,000 | FDFA ﷺ | NFD | 0.194576 / 287.0 | 0.164828 / 406.9 |
| 100,000,000 | FDFA ﷺ | NFKC | error 0.578 / 286.9 | capped 10.106 / 6510.3 |
| 100,000,000 | FDFA ﷺ | NFKD | error 0.603 / 286.9 | capped 10.108 / 6510.5 |
| 29,826,160 | FDFA ﷺ | NFKD | 0.482752 / 1024.9 | 1.95704 / 2187.5 |

## Where time and memory go

Counted from the original code, rather than a sampled profile: it has **one**, not several, uint32_t vector. It decodes and recursively decomposes into this growing vector, reorders the whole vector, optionally composes in place, encodes once to count bytes, then encodes again into the allocated string. At 29,826,160 FDFA scalars the decomposition contains 536,870,880 points. Its capacity is 536,870,912 points, or 2 GiB. Input is 89,478,480 bytes and output is 984,263,280 bytes; together these are 1 GiB minus 64 bytes. The expected simultaneous footprint is therefore about 3 GiB plus overhead, matching measured cloud RSS. Realloc can transiently overlap old and new allocations; this explains a possible higher high-water mark but does not establish why the M4 observation differs. Every decomposed point triggers ordering/encoding work, with repeated binary-search lookups. Composition adds another pass. Original insertion ordering also has quadratic worst-case cost on a long scrambled combining run.

The replacement first checks whether the input can be retained. ASCII retains after one byte scan. Decomposed forms check mappings and canonical order; composed forms conservatively accept only individually normalized starters with no composition across their decomposed boundary. Marks take the streaming path. This is not a generated NFC_QC/NFKC_QC table, and can conservatively miss already normalized marked strings.

The general path makes two bounded streaming passes: size/check UTF-16 limits, then emit UTF-8 to one exact result allocation. A batch target is 64 points, but it flushes only at a starter. All nonzero-class runs remain intact. Composition keeps a trailing starter pending when the next starter could join, including Hangul and Bengali. Unsorted runs of at most 64 marks use insertion ordering; longer ones use stable class counting sort. Memory is input + output + the longest unfinished canonical run and its temporary sort buffer, not constant memory for an arbitrarily long all-mark string. Three small uint32_t vectors can coexist in the general path (batch, raw scalar decomposition, normalized scalar block); the prefix shortcut uses batch plus raw/normalized motif vectors. At the ligature motif they have 64-point capacities (256 bytes each), with a 72-byte encoded-motif allocation. The three verified-key thread-local lookup caches total 1,856 bytes per thread on this 64-bit target.

Repeated starter-safe motifs up to eight input points are normalized once and copied by doubling. Every boundary is checked; the final copy remains raw decomposition so following marks reorder correctly. Repeated single scalars have a similar guarded bulk path. Thus these intentionally repeated benchmark inputs mostly cost input scanning and output copies rather than hundreds of millions of scalar encodings. General nonperiodic throughput was not benchmarked.

## Correctness and failing mutants

The default randomized test independently generates 100,000 strings in C and Node and compares all four forms as WTF-8. It includes Hangul, canonical singletons, compatibility forms, scrambled marks, leading/trailing marks, surrogate halves and pairs. Most strings have 1 to 80 points; periodic samples have 511/512/513-point scrambled runs, and every seventeenth sample repeats a 1-to-8-point motif followed by marks. Sanitizers are enabled. Existing exhaustive tests cover all 1,114,112 scalar values and 346,200 short contexts.

The extended oracle fixture covers run lengths 63/64/65 and 127/128/129, multiple runs in a batch, blocked equal-class composition, precomposed starters plus lower-class marks, Hangul/Bengali joining, and repeated motifs followed by scrambled marks. The huge integration probe is opt-in, not registered as an oracle fixture.

Each mutant below compiled and then failed a semantic comparison; no compiler error counted as a catch. Source was restored after each run.

| Mutant | Catcher | Observed failure |
| --- | --- | --- |
| Reorder only the first canonical run in a batch | normalize_long_marks.a, Node oracle | stdout differs |
| Flush at 64 points even inside a nonzero-class run | normalize_long_marks.a, Node oracle | stdout differs |
| Return every nonempty input from quick check | normalize_long_marks.a, Node oracle | stdout differs |
| Drop a blocked combining point | normalize_long_marks.a, Node oracle | stdout differs |
| Allow equal-class blocked composition | normalize_long_marks.a, Node oracle | stdout differs |
| Accept a combining-class cache entry without checking its key | TestNormalizeRandomMatchesNode | randomized output mismatches |
| Carry the final repeated copy precomposed | normalize_long_marks.a, Node oracle | stdout differs |
| Remove output-length guards | independent one-ligature-over-limit native/Node probe | mutant accepts 536,870,898 units; correct native and Node throw RangeError |

Large-output bytes were also checked against Node's single-ligature normalization at every 33-byte block, covering all 984,263,280 bytes. Node independently confirmed the full result equals its normalized single-ligature block repeated 29,826,160 times. An actual compiled Adamic integration program printed 536870880. Its best of five wall times was 0.774 s, with 1,049,596 KiB peak RSS (1.001 GiB); construction and startup are included. The program is `const normalized = 'ﷺ'.repeat(29826160).normalize('NFKD'); console.log(`${normalized.length}`);`, built with `go run ./cmd/adamic build /tmp/normalize-long/integration.a -o /tmp/normalize-long/integration`. All five executions exited 0. Logs: integration-build.log and integration-measure.log under the scratch directory.

## Reproduction and verification scope

All test output was redirected to files, never piped. Source the tool environment first. The following commands were run (full logs in `/tmp/normalize-long`, setup in `/tmp/normalize-setup.log`):

```sh
bash cloud/setup.sh > /tmp/normalize-setup.log 2>&1
source /workspace/adamic-tools/env.sh
go test ./internal/native -run '^TestNormalizeMatchesNode$' -count=1 -v > /tmp/normalize-long/test-before.log 2>&1
go test ./internal/native -count=1 -parallel=1 -timeout 30m > /tmp/normalize-long/native-final-retry.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/(normalize|normalize_form|normalize_long_marks|library_string_existing|library_string_prototype|optional_strings).a$' -count=1 -v -timeout 30m > /tmp/normalize-long/oracle-final.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m > /tmp/normalize-long/counts-final.log 2>&1
go vet ./... > /tmp/normalize-long/vet.log 2>&1
ADAMIC_NORMALIZE_BENCH_LOG=/tmp/normalize-long/after.jsonl go test ./internal/native -run '^TestNormalizeLongMeasurements$' -count=1 -v -timeout 30m > /tmp/normalize-long/measure-after.log 2>&1
```

The baseline used the same C/Node probe and Python wait4 driver as the committed opt-in test, compiled directly against unchanged runtime sources before edits. Raw baseline and after records are `/tmp/normalize-long/before.jsonl` and `after.jsonl`. The ordinary test gate skips the 610-process performance probe.

Final native package: `ok  	github.com/system-inc/adamic/internal/native	132.658s`.

Filtered oracle: `ok  	github.com/system-inc/adamic/internal/oracle	1.129s`.

Recorded counts: `ok  	github.com/system-inc/adamic/internal/oracle	13.301s`.

Performance probe: `ok  	github.com/system-inc/adamic/internal/native	265.632s`.

The touched fixture also passed cohere format-only (`go run ./command/cohere --directory /workspace/adamic --format-only --no-cache internal/oracle/testdata/normalize_long_marks.a`, run from cohere, output in `/tmp/normalize-long/cohere-format.log`), followed by an uncached oracle rerun. This command checks formatting only; types and lint are explicitly skipped. Vet, gofmt and git diff --check passed. One earlier concurrent native-package attempt failed in unrelated TestRuntimeCacheConcurrentBuilders with a temporary compiler-wrapper `text file busy` error. That test passed three retries, and the entire native package subsequently passed with parallel=1. All seven repository mutants were rerun against the final implementation.

Not covered: M4 Max timings; the complete repository `go test ./...` gate (native package, six relevant oracle fixtures, all recorded allocation counts, and repository-wide vet were run); uncapped oversized FDFA error trials; performance of huge nonperiodic strings or huge all-mark runs. Unicode tables and their generator are unchanged.
