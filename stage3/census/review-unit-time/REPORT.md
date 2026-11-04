# Review unit timing census

Measured 2026-11-04 UTC, Linux amd64, at origin/main `ec0209711ba51bdb6836f272e06dc671561c0528`. Report only, not for landing. No committed compiler or test source changed.

## Finding

Both named programs spend the cold time inside the **source Node leg, before Node starts**, in oracle cache identity initialization. In the concurrent shard run, `n_paren` initializes the shared `gateOnce`; `defined_lent_field` waits for the same `sync.Once`. Each loses about 23.003 s there. Their actual Node executions take 0.090 s and 0.085 s, lowering about 0.052 s each, and their own sanitized and release builds less than 0.224 s each.

These are the first active programs in their shards. `_006` starts with `n_paren`; `_007` skips three pending programs before `defined_lent_field`. Later siblings inherit initialized identity and runtime products. The controls take 0.45 s and 0.50 s in that same process, but about 25 s when each is put first in its own cold process. The timing follows first use rather than these two inputs.

The exact historical 39.78 s and 38.56 s did not reproduce on this box at current main. All measured active units passed the unchanged 30 s check. This establishes the dominant leg and shared wait in the measured cold case; it does not establish the historical runtime cache states or historical machine contention. With warm build products, first-use identity still costs 3.580 to 4.036 s in the four isolated runs.

## Programs and harness

All four inputs are under `internal/oracle/testdata/review/agree/`.

| Short name | Full subtest name | Role |
| --- | --- | --- |
| paren | `TestReviewProgramsAgreeWithNode_006/fxspptb_047cb0d_n_paren.a` | target |
| field | `TestReviewProgramsAgreeWithNode_007/fxspptb_aa17d3c_defined_lent_field.a` | target |
| template | `TestReviewProgramsAgreeWithNode_006/fxspptb_047cb0d_n_template.a` | fast sibling control |
| super | `TestReviewProgramsAgreeWithNode_007/fxspptb_abfe962_super_closure.a` | fast sibling control |

Read the whole [review_test.go](../../../internal/oracle/review_test.go), its execution helpers in [oracle_test.go](../../../internal/oracle/oracle_test.go), and the cache and JSON helpers before timing. The causal path is:

1. `reviewEach` starts its 30 s clock before `reviewAgreement`.
2. `reviewAgreement` lowers, calls source `onNode`, emitted JavaScript, `natively`, `released`, and leak checking, in that order.
3. `onNode` enters `cachedNode`, which calls `identity` before the result-cache bypass is consulted.
4. [identity](../../../internal/oracle/cache_test.go#L172), under `gateOnce.Do`, obtains three runtime libraries in sequence: sanitized, release, counted. These calls form runtime requests and obtain their buildcache products. A concurrent caller blocks on that same once.
5. [runtimeRequest and headersRead](../../../internal/native/library.go#L353) scan runtime header dependencies before obtaining the product. The scans are memoized in process, so a fresh process pays even for product hits. Cold products also compile the archives there.

`natively` is already the native **sanitized** comparison build and run. There is no separate third plain native comparison in `reviewAgreement`. `released` builds the unsanitized native executable with `native.Options{}`. The table measures those two builds. On Linux the leak check runs the sanitized binary again with leak detection enabled; no additional leak build occurs. The harness's release helper does not set `Options.Release` (the separate shipped ThinLTO variant).

## Method and cache boundary

Cloned `https://github.com/system-inc/adamic.git`, created `typescript/review-unit-time` from `origin/main`, and ran exactly:

```sh
export GOPROXY='https://proxy.golang.org|direct'
timeout --kill-after=60s 1800 bash cloud/setup.sh > /workspace/scratch/setup.log 2>&1
```

Setup passed in 319.622 s. Tools: Go 1.27.1, Node v24.19.0, clang 20.1.8. `nproc` is 5; cgroup `cpu.max` is `400000 100000` (four CPUs of quota). Setup and test compilation finished before measurement. No synthetic background workload was added.

Instrumentation exists only in `/workspace/scratch/timing-copy`, a copy of the checkout with the pinned cohere tree linked in. Timers surround `lowered`, `onNode`, `onJavaScriptBackend`, `natively`, `released`, `leaks`, each `native.Build`, actual child executions, `identity`, and each of its three runtime library calls. `t.Logf` records elapsed wall seconds. The original 30 s check, programs, and comparison logic were retained. Logs were written to files throughout, never piped to head or tail.

The compiled scratch binary is `/workspace/scratch/oracle-timing.test`. `ADAMIC_GATE_UNCACHED=1` bypasses saved oracle results so every execution actually happens, while retaining build-product caching. This also means the Linux leak run occurs at `reviewAgreement`'s final leak call, rather than within `cachedNative`'s miss callback. Its measured duration is reported separately. `ADAMIC_BUILD_STORE=off` prevents remote product fetches; `ADAMIC_BUILD_CACHE=on`. Native object trust and split mode remain unset, as inherited, so the program's own main.c is compiled and linked on both warm and cold runs.

For each isolated program, set **both** `ADAMIC_BUILD_CACHE_DIR` and `ADAMIC_BUILD_OUTPUT_DIR` to a new, initially absent directory under `/workspace/scratch/timings/cache-{paren,field,template,super}`. Using both prevents fallback to another warm directory. No existing cache was cleared. Run `TestProduct_JSONTypes` first in that directory, outside timing; every measured json-types lookup was a hit. Then run the selected exact subtest once cold, and repeat in a fresh process using the same directory warm. Thus cold includes missing shared runtime products, not cold Go compilation or a json-types build. Warm means warm disk build products, not warm in-process `sync.Once` state.

Each execution used the binary with `-test.count=1 -test.timeout=30m -test.parallel=5 -test.v=test2json`, wrapped by `go tool test2json -t -p oracle`. The isolated selector was `^TestReviewProgramsAgreeWithNode_006$/^fxspptb_047cb0d_n_paren\.a$`, with corresponding exact names for the other cases.

The loaded cold case used another fresh directory, `cache-loaded`, prepared only with json-types, then selected **both full shards at once** with `-test.run='^TestReviewProgramsAgreeWithNode_00[67]$' -test.parallel=5`. These wrappers call `t.Parallel`, but their children run sequentially, so this requested load has two concurrent shard bodies, not five parallel children. Shared initialization and build products are intentionally shared in this case.

## Leg timings

All entries are wall seconds. Node and JavaScript totals include their nested rows; do not add totals to their components. Unit totals come from test events, rounded by the test runner to 0.01 s. Other values are timer measurements rounded here to 0.001 s. `<0.001` is a measured positive duration below that precision. Cache states below are warm / cold / loaded respectively.

| Program | Leg | Warm s | Cold s | Loaded s | Hit or built / execution |
| --- | --- | ---: | ---: | ---: | --- |
| paren | Lowering | 0.030 | 0.032 | 0.052 | no build |
| paren | Source Node total | 3.955 | 23.225 | 23.093 | includes identity below |
| paren | Identity within source Node | 3.863 | 23.137 | 23.003 | runtime hits / runtime builds / shared initialization or wait |
| paren | Source Node child run | 0.092 | 0.087 | 0.090 | executed / executed / executed |
| paren | JavaScript backend total | 1.578 | 1.497 | 1.565 | json-types hit / hit / hit |
| paren | JavaScript child run | 0.044 | 0.040 | 0.038 | executed / executed / executed |
| paren | Native sanitized build | 0.188 | 0.180 | 0.174 | main.c built each time; base runtime already obtained by identity |
| paren | Native sanitized comparison run | 0.005 | 0.004 | 0.003 | executed / executed / executed |
| paren | Release build | 0.086 | 0.086 | 0.078 | main.c built each time; base runtime already obtained by identity |
| paren | Release run | 0.001 | 0.001 | 0.001 | executed / executed / executed |
| paren | Sanitized leak run | 0.008 | 0.008 | 0.008 | same sanitized binary, executed each time |
| paren | **Unit total** | **5.85** | **25.03** | **24.98** | pass / pass / pass |
| field | Lowering | 0.038 | 0.039 | 0.052 | no build |
| field | Source Node total | 3.661 | 25.006 | 23.088 | includes identity below |
| field | Identity within source Node | 3.593 | 24.916 | 23.002 | runtime hits / runtime builds / shared initialization or wait |
| field | Source Node child run | 0.069 | 0.090 | 0.085 | executed / executed / executed |
| field | JavaScript backend total | 1.501 | 1.488 | 1.638 | json-types hit / hit / hit |
| field | JavaScript child run | 0.037 | 0.039 | 0.040 | executed / executed / executed |
| field | Native sanitized build | 0.188 | 0.217 | 0.223 | main.c built each time; base runtime already obtained by identity |
| field | Native sanitized comparison run | 0.004 | 0.004 | 0.004 | executed / executed / executed |
| field | Release build | 0.098 | 0.109 | 0.101 | main.c built each time; base runtime already obtained by identity |
| field | Release run | 0.001 | 0.001 | 0.001 | executed / executed / executed |
| field | Sanitized leak run | 0.009 | 0.009 | 0.009 | same sanitized binary, executed each time |
| field | **Unit total** | **5.50** | **26.88** | **25.12** | pass / pass / pass |
| template | Lowering | 0.031 | 0.032 | 0.033 | no build |
| template | Source Node total | 3.663 | 23.063 | 0.079 | includes identity below |
| template | Identity within source Node | 3.580 | 22.974 | <0.001 | runtime hits / runtime builds / already initialized |
| template | Source Node child run | 0.081 | 0.089 | 0.079 | executed / executed / executed |
| template | JavaScript backend total | 1.497 | 1.556 | 0.067 | json-types hit / hit / hit |
| template | JavaScript child run | 0.037 | 0.038 | 0.047 | executed / executed / executed |
| template | Native sanitized build | 0.184 | 0.168 | 0.172 | main.c built each time; base runtime already obtained by identity |
| template | Native sanitized comparison run | 0.004 | 0.004 | 0.003 | executed / executed / executed |
| template | Release build | 0.090 | 0.084 | 0.085 | main.c built each time; base runtime already obtained by identity |
| template | Release run | 0.001 | 0.001 | 0.001 | executed / executed / executed |
| template | Sanitized leak run | 0.008 | 0.008 | 0.009 | same sanitized binary, executed each time |
| template | **Unit total** | **5.48** | **24.92** | **0.45** | pass / pass / pass |
| super | Lowering | 0.045 | 0.034 | 0.036 | no build |
| super | Source Node total | 4.127 | 22.811 | 0.077 | includes identity below |
| super | Identity within source Node | 4.036 | 22.739 | <0.001 | runtime hits / runtime builds / already initialized |
| super | Source Node child run | 0.091 | 0.072 | 0.077 | executed / executed / executed |
| super | JavaScript backend total | 1.558 | 1.439 | 0.060 | json-types hit / hit / hit |
| super | JavaScript child run | 0.037 | 0.039 | 0.040 | executed / executed / executed |
| super | Native sanitized build | 0.201 | 0.194 | 0.213 | main.c built each time; base runtime already obtained by identity |
| super | Native sanitized comparison run | 0.004 | 0.004 | 0.008 | executed / executed / executed |
| super | Release build | 0.116 | 0.104 | 0.096 | main.c built each time; base runtime already obtained by identity |
| super | Release run | 0.001 | 0.001 | 0.001 | executed / executed / executed |
| super | Sanitized leak run | 0.017 | 0.011 | 0.009 | same sanitized binary, executed each time |
| super | **Unit total** | **6.07** | **24.60** | **0.50** | pass / pass / pass |

The cold identity subcomponents distinguish the work owner from a blocked caller:

| Program | Runtime lookup | Warm s | Cold s | Loaded s | Product state warm / cold / loaded |
| --- | --- | ---: | ---: | ---: | --- |
| paren | sanitized | 1.176 | 10.448 | 10.282 | hit / built / built once by paren |
| paren | release | 1.293 | 6.463 | 6.296 | hit / built / built once by paren |
| paren | counted | 1.382 | 6.215 | 6.410 | hit / built / built once by paren |
| field | sanitized | 1.221 | 10.996 | shared wait | hit / built / built once by paren |
| field | release | 1.185 | 6.816 | shared wait | hit / built / built once by paren |
| field | counted | 1.174 | 7.092 | shared wait | hit / built / built once by paren |
| template | sanitized | 1.219 | 10.160 | already initialized | hit / built / already supplied |
| template | release | 1.135 | 6.211 | already initialized | hit / built / already supplied |
| template | counted | 1.214 | 6.591 | already initialized | hit / built / already supplied |
| super | sanitized | 1.347 | 10.176 | already initialized | hit / built / already supplied |
| super | release | 1.380 | 6.277 | already initialized | hit / built / already supplied |
| super | counted | 1.296 | 6.274 | already initialized | hit / built / already supplied |

For example, paren's cold identity's sanitized / release / counted calls are 10.448 / 6.463 / 6.215 s. The inner product census reports 9.25 / 5.24 / 5.02 s for those builds; request formation and header scanning sit outside those product timers. Warm inner product hits report about 0.02 s each, but their surrounding runtime calls still take about 1.2 to 1.4 s. A hit therefore does not mean the whole call is free.

## Products and actual compiler work

The three generic products are `native-runtime` keys `c58ea05bc708` (sanitized), `5b0dc7a6403c` (release), and `2f2ef973406c` (counted). Each isolated cold case built all three: 222 runtime C compiles and three archives total. Each isolated warm case hit all three, with no runtime C compiles. The program's two main.c compiles and two links are logged `off` on every isolated run: they ran without object caching, not hits. The counted runtime is obtained for identity even though the Linux comparison does not run a counted binary.

json-types key `e5c733960966` was built in each untimed prerequisite run and hit during measurement. On the isolated measured paths, its hit lookup cost about 1.3 to 1.4 s inside the JavaScript backend total. It did not build there.

The loaded run built the same three generic products once, plus Node-host sanitized/release runtime products `91f42d6ce7d4` and `a7c0b2d824a5` when the later host programs needed them. Its complete census has 372 runtime C compiles, five archives, 28 program C compiles, and 28 program links. All 14 measured json-types lookups were hits. This explains the additional roughly 17 s host units below, separately from the two target units' first-use identity wait.

## Every loaded subtest

The full invocation passed in 44.921 s. `_006` took 44.91 s and `_007` 44.45 s, concurrently. There were 14 passing active programs and eight pending skips. Skips occur before `reviewEach` starts its 30 s clock.

| Shard | Program | Wall s | Result |
| --- | --- | ---: | --- |
| 006 | `fxspptb_047cb0d_n_paren.a` | 24.98 | pass |
| 006 | `fxspptb_047cb0d_n_template.a` | 0.45 | pass |
| 006 | `fxspptb_4ddd17f_opt_3.a` | 0.00 | skip |
| 006 | `fxspptb_a34388d_cyc_iface_control.a` | 0.46 | pass |
| 006 | `fxspptb_a34388d_cyc_iterable.a` | 0.00 | skip |
| 006 | `fxspptb_e4eec87_v02_typeof_null_literal.a` | 0.44 | pass |
| 006 | `fxspptb_fa49e43_spreadfresh.a` | 0.51 | pass |
| 006 | `fxspptb_oct8_nodehost_fs_numeric_error.a` | 0.00 | skip |
| 006 | `fxspptb_oct8_nodehost_host_directory.a` | 17.08 | pass |
| 006 | `fxspptb_oct9_native_p08_null_array_write.ts` | 0.53 | pass |
| 006 | `fxspptb_oct9_views_p09_tuple_destructure_wrong.ts` | 0.45 | pass |
| 006 | `fxspptb_oct9_views_p38_same_name_string_undefined.ts` | 0.00 | skip |
| 007 | `fxspptb_047cb0d_n_try.a` | 0.00 | skip |
| 007 | `fxspptb_4e1649c_generic_template.a` | 0.00 | skip |
| 007 | `fxspptb_903f25b_groups_proto2.a` | 0.00 | skip |
| 007 | `fxspptb_aa17d3c_defined_lent_field.a` | 25.12 | pass |
| 007 | `fxspptb_abfe962_super_closure.a` | 0.50 | pass |
| 007 | `fxspptb_bdb950b_gap_nul.a` | 0.00 | skip |
| 007 | `fxspptb_d96d304_normalize_long.a` | 0.46 | pass |
| 007 | `fxspptb_e4eec87_w01_uncovered_paths.a` | 0.53 | pass |
| 007 | `fxspptb_oct8_nodehost_fs_large_roundtrip.a` | 17.36 | pass |
| 007 | `fxspptb_oct9_views_p38_same_name_closure_value.ts` | 0.49 | pass |

## Evidence retained locally

Only this report is committed. Setup output is `/workspace/scratch/setup.log`; structured measurement logs and build census logs are `/workspace/scratch/timings/{paren,field,template,super}-{cold,warm}.{log,build.log}` and `loaded.{log,build.log}`. `status.json` records commands and exit statuses; `/workspace/scratch/run-measure.py` and `run-control.py` are the drivers. The initial `n_try` control probe skipped because of its pending sidecar; the active `super_closure` control replaced it. All active measurements and the full loaded invocation exited successfully.

SHA-256 checksums of the main structured timing and build logs:

| File under timings/ | SHA-256 |
| --- | --- |
| `paren-cold.log` | `be4bb6d7acbf3c26499966b07d9dfb19c9347426f1a59d55e6a812b50d87e657` |
| `paren-cold.build.log` | `04153d2b41675bb80dd77d6644563d9bad32da3b6b8ba896af07191e27de3e88` |
| `paren-warm.log` | `437559a8721ebf2b037ca6e048a4ecc6b3f5057c4b512ddec8951d62b7bf39dc` |
| `paren-warm.build.log` | `b16fa71a5dd0a300ac378ded2ab4b22f4d6a7cc77f2ea8e8323f943baad7df5b` |
| `field-cold.log` | `0fcf7cd3b869dbebcde858f499b040f1317fde76d9113656bf0a196b565b52d7` |
| `field-cold.build.log` | `70d6c93cbaececd0565861f99a819f5c769045877aaa5902bdad764b3faa8613` |
| `field-warm.log` | `04f648f76dd1795e65b4649b2b1dcb7a3fc0dec7532ef8d839ebd56fddfc12f7` |
| `field-warm.build.log` | `05499612e4e6bed86548279a9a07c3e915ebe8bbecf9b9346fd3a50b41c42c71` |
| `template-cold.log` | `398e632e3c6e90c730097bf79a2809b0ebd8358f5d3a7ababf373b110a08056e` |
| `template-cold.build.log` | `0e60e25f3e052fdcedadce25795401a27434d9ce0d6f253a64e43d841cecfb3e` |
| `template-warm.log` | `2d8ae2da31b8f45f6bcb41fdc83e7c83a4fc9a9c894f948849b4600a705e0421` |
| `template-warm.build.log` | `f2ee5327c29a229ba61804953f98bbfdc43393a2952b28b556096a9ee1db3700` |
| `super-cold.log` | `2c3e025b170c928e2ac5ddee21a7529470e7c4f82dabfbca8079c2919be187a9` |
| `super-cold.build.log` | `4d0d144936f77f19bbbec6a14a52db43d163742cb6d4e37c3f90e52f1ae59660` |
| `super-warm.log` | `0d21220ff5b28bb72d226a1dce31fa5258c9d3923b8d4c9aa9cf75bc704b529c` |
| `super-warm.build.log` | `660fe3af9433b307be39162033cbdf1c904a7a1574c72855880c637be5efd018` |
| `loaded.log` | `8e47fc32ec42cc7762b97a4364dbfdbf724a16fe7d0fb2ebd5585ee9ec11f456` |
| `loaded.build.log` | `7e333971b33d262a87f33f194b5267dbb01eca74c0f19c61e6ee6b64ccffe8dc` |
