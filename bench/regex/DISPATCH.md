# Regex dispatch repair against 8ac5675

The regular-engine rollout regressed 66 of 102 benchmark rows. The repaired dispatcher keeps the large-input wins and removes unnecessary work from short boolean searches. Against `8ac5675`, median speedup is **1.93x**, with **zero instruction increases** and **zero wall regressions beyond the declared 5% noise allowance**. One row, `shouting.go:122`, is 2.1% slower; it remains visible below. Native is faster than Node on **48/102** rows.

The four hard cases are 24.35x (keywords), 37.68x (email), 1.24x (ignorecase) and 42.95x (lookbehind) faster than the baseline. Node remains 2.12x faster on keywords and 7.64x faster on ignorecase; native is 2.94x faster on email and 4.63x faster on lookbehind.

## Complete per-pattern results

Wall times are nanoseconds per call, best of five, with engine order rotated each round. Short rows use an untimed baseline pilot to target at least 100ms per baseline sample; hard rows retain 200 calls. Both native binaries and Node use identical calls and checksums. Native warms 100 calls; Node warms 10,000. The machine was not running concurrent tests or instruction profiling during this timing run. The 5% wall allowance and 1% instruction guard are benchmark bounds, not statistical confidence intervals. Actual instruction counts have no increases even below 1%.

Instruction counts exclude startup and warmup, using Callgrind over one warmed nine-input lint cycle or two hard-case inputs. Counts below are normalized per call. [CSV with raw seconds, calls, checksums and counts](dispatch-results.csv).

| Pattern | 8ac ns | Current ns | Node ns | vs 8ac | vs Node | 8ac instructions/call | Current instructions/call |
|---|---:|---:|---:|---:|---:|---:|---:|
| `cohere/internal/lint/configuration/configuration.go:839` | 177.2 | 78.7 | 20.0 | 2.25x | 0.25x | 2535.9 | 935.7 |
| `cohere/internal/lint/ecmascript/dotnotation/dotnotation.go:121` | 331.1 | 263.8 | 37.7 | 1.25x | 0.14x | 4205.6 | 2608.2 |
| `cohere/internal/lint/ecmascript/dotnotation/dotnotation.go:141` | 217.5 | 99.8 | 23.2 | 2.18x | 0.23x | 2884.3 | 1287.0 |
| `cohere/internal/lint/housesets/housesets.go:112` | 475.9 | 63.8 | 126.2 | 7.45x | 1.98x | 6912.2 | 801.4 |
| `cohere/internal/lint/rules/base/consistency_require_pagination_argument_name.go:27` | 10.7 | 5.0 | 21.9 | 2.16x | 4.41x | 150.0 | 75.7 |
| `cohere/internal/lint/rules/base/consistency_require_pagination_argument_name.go:30` | 10.5 | 5.0 | 21.7 | 2.09x | 4.32x | 147.1 | 74.8 |
| `cohere/internal/lint/rules/core/default_case.go:29` | 9.1 | 5.1 | 19.1 | 1.77x | 3.71x | 113.0 | 70.8 |
| `cohere/internal/lint/rules/core/no_inline_comments.go:38` | 240.8 | 115.3 | 25.7 | 2.09x | 0.22x | 2969.4 | 1366.9 |
| `cohere/internal/lint/rules/core/no_warning_comments.go:280` | 245.6 | 127.1 | 125.7 | 1.93x | 0.99x | 3347.4 | 1459.3 |
| `cohere/internal/lint/rules/core/require_description.go:214` | 1226.9 | 1184.1 | 78.3 | 1.04x | 0.07x | 15742.1 | 14368.8 |
| `cohere/internal/lint/rules/core/require_description.go:217` | 259.8 | 168.9 | 23.0 | 1.54x | 0.14x | 3525.6 | 1923.0 |
| `cohere/internal/lint/rules/core/require_description.go:222` | 212.5 | 115.3 | 23.8 | 1.84x | 0.21x | 2927.0 | 1308.8 |
| `cohere/internal/lint/rules/core/require_description.go:233` | 213.7 | 100.9 | 19.8 | 2.12x | 0.20x | 2839.7 | 1230.1 |
| `cohere/internal/lint/rules/next/no_html_link_for_pages_routes.go:139` | 188.6 | 23.1 | 22.3 | 8.18x | 0.97x | 2533.6 | 279.9 |
| `cohere/internal/lint/rules/next/no_html_link_for_pages_routes.go:140` | 177.6 | 80.8 | 18.9 | 2.20x | 0.23x | 2537.6 | 932.7 |
| `cohere/internal/lint/rules/next/no_html_link_for_pages_routes.go:141` | 191.5 | 82.2 | 22.4 | 2.33x | 0.27x | 2517.8 | 915.2 |
| `cohere/internal/lint/rules/next/no_html_link_for_pages_routes.go:142` | 187.6 | 75.6 | 19.3 | 2.48x | 0.26x | 2538.9 | 933.2 |
| `cohere/internal/lint/rules/next/no_html_link_for_pages_routes.go:251` | 181.2 | 24.9 | 148.7 | 7.27x | 5.97x | 2561.0 | 308.7 |
| `cohere/internal/lint/rules/next/no_location_assign_relative_destination.go:26` | 500.3 | 400.5 | 44.1 | 1.25x | 0.11x | 6196.9 | 4598.8 |
| `cohere/internal/lint/rules/nexus/abbreviation_vocabulary.go:162` | 8.3 | 5.1 | 18.8 | 1.63x | 3.69x | 105.9 | 72.1 |
| `cohere/internal/lint/rules/nexus/abbreviation_vocabulary.go:176` | 2054.1 | 1998.3 | 107.3 | 1.03x | 0.05x | 22843.0 | 21234.6 |
| `cohere/internal/lint/rules/nexus/abbreviation_vocabulary.go:177` | 4674.2 | 47.3 | 111.0 | 98.72x | 2.34x | 48894.3 | 480.9 |
| `cohere/internal/lint/rules/nexus/abbreviation_vocabulary.go:178` | 214.7 | 114.8 | 69.6 | 1.87x | 0.61x | 3004.4 | 1317.8 |
| `cohere/internal/lint/rules/nexus/consistency_no_abbreviated_identifier.go:50` | 5273.7 | 18.6 | 131.4 | 284.17x | 7.08x | 55369.1 | 291.7 |
| `cohere/internal/lint/rules/nexus/consistency_no_ambiguous_identifier.go:17` | 3038.9 | 2532.6 | 262.1 | 1.20x | 0.10x | 30828.6 | 29223.4 |
| `cohere/internal/lint/rules/nexus/consistency_no_ambiguous_identifier.go:18` | 8.2 | 5.9 | 19.3 | 1.39x | 3.29x | 105.9 | 72.1 |
| `cohere/internal/lint/rules/nexus/consistency_no_ambiguous_identifier.go:19` | 8.5 | 4.8 | 17.1 | 1.78x | 3.57x | 109.3 | 67.0 |
| `cohere/internal/lint/rules/nexus/consistency_no_boolean_outcome.go:48` | 663.5 | 214.6 | 138.8 | 3.09x | 0.65x | 8275.6 | 2048.7 |
| `cohere/internal/lint/rules/nexus/consistency_no_return_void.go:100` | 257.0 | 138.5 | 23.3 | 1.85x | 0.17x | 3222.6 | 1617.7 |
| `cohere/internal/lint/rules/nexus/consistency_no_return_void.go:116` | 213.2 | 113.4 | 22.0 | 1.88x | 0.19x | 2864.9 | 1246.7 |
| `cohere/internal/lint/rules/nexus/consistency_require_constant_casing.go:350` | 480.9 | 55.5 | 24.3 | 8.66x | 0.44x | 6705.8 | 640.7 |
| `cohere/internal/lint/rules/nexus/consistency_require_matching_return_type.go:216` | 217.5 | 106.8 | 28.2 | 2.04x | 0.26x | 2847.7 | 1231.0 |
| `cohere/internal/lint/rules/nexus/shouting.go:121` | 123.2 | 100.2 | 42.4 | 1.23x | 0.42x | 1424.8 | 1381.4 |
| `cohere/internal/lint/rules/nexus/shouting.go:122` | 129.4 | 132.1 | 44.3 | 0.98x | 0.34x | 1400.4 | 1356.4 |
| `cohere/internal/lint/rules/nexus/shouting.go:123` | 473.6 | 382.1 | 159.9 | 1.24x | 0.42x | 6742.0 | 5003.4 |
| `cohere/internal/lint/rules/nexus/shouting.go:147` | 192.7 | 53.0 | 70.4 | 3.64x | 1.33x | 2814.4 | 636.1 |
| `cohere/internal/lint/rules/nexus/shouting.go:153` | 206.3 | 37.8 | 159.4 | 5.45x | 4.21x | 2578.6 | 403.1 |
| `cohere/internal/lint/rules/nexus/shouting.go:160` | 186.1 | 23.8 | 153.1 | 7.81x | 6.42x | 2522.0 | 269.7 |
| `cohere/internal/lint/rules/nexus/shouting.go:166` | 249.5 | 115.1 | 152.4 | 2.17x | 1.32x | 3216.4 | 1119.6 |
| `cohere/internal/lint/rules/nexus/shouting.go:175` | 262.7 | 111.5 | 134.6 | 2.36x | 1.21x | 3228.6 | 1128.4 |
| `cohere/internal/lint/rules/nexus/shouting.go:177` | 8.1 | 5.8 | 30.5 | 1.39x | 5.26x | 121.6 | 80.1 |
| `cohere/internal/lint/rules/nexus/shouting.go:178` | 251.5 | 145.2 | 33.9 | 1.73x | 0.23x | 3389.6 | 1793.1 |
| `cohere/internal/lint/rules/nexus/shouting.go:181` | 225.0 | 108.2 | 25.4 | 2.08x | 0.24x | 2827.4 | 1229.3 |
| `cohere/internal/lint/rules/nexus/shouting.go:182` | 170.6 | 70.2 | 18.6 | 2.43x | 0.27x | 2503.9 | 899.0 |
| `cohere/internal/lint/rules/nexus/shouting.go:183` | 163.9 | 73.9 | 28.6 | 2.22x | 0.39x | 2469.4 | 871.3 |
| `cohere/internal/lint/rules/nexus/shouting.go:184` | 401.8 | 266.6 | 31.3 | 1.51x | 0.12x | 4592.2 | 2994.9 |
| `cohere/internal/lint/rules/nexus/shouting.go:464` | 407.1 | 268.0 | 79.9 | 1.52x | 0.30x | 5035.9 | 3283.2 |
| `cohere/internal/lint/rules/react/boolean_prop_naming.go:621` | 180.6 | 26.8 | 95.6 | 6.75x | 3.57x | 2639.8 | 380.9 |
| `cohere/internal/lint/rules/react/conformance/classify.go:95` | 3698.8 | 3490.3 | 205.8 | 1.06x | 0.06x | 42844.6 | 41335.7 |
| `cohere/internal/lint/rules/react/conformance/classify.go:98` | 316.1 | 167.7 | 80.5 | 1.89x | 0.48x | 3886.3 | 2012.3 |
| `cohere/internal/lint/rules/react/exhaustive_deps.go:451` | 179.8 | 46.5 | 53.9 | 3.86x | 1.16x | 2667.7 | 557.6 |
| `cohere/internal/lint/rules/react/no_unstable_nested_components.go:450` | 195.8 | 108.4 | 21.6 | 1.81x | 0.20x | 2731.3 | 1131.1 |
| `cohere/internal/lint/rules/react/sort_comp.go:119` | 185.1 | 27.7 | 145.0 | 6.69x | 5.24x | 2622.2 | 356.8 |
| `cohere/internal/lint/rules/tailwind/class_existence.go:63` | 207.5 | 99.2 | 20.7 | 2.09x | 0.21x | 2787.9 | 1186.9 |
| `cohere/internal/lint/rules/tailwind/class_existence.go:70` | 191.9 | 92.2 | 20.2 | 2.08x | 0.22x | 2580.7 | 980.4 |
| `cohere/internal/lint/rules/tailwind/class_order_strict.go:89` | 8742.5 | 160.5 | 358.9 | 54.47x | 2.24x | 106308.1 | 2224.1 |
| `cohere/internal/lint/rules/tailwind/collapse/utility_nodes.go:131` | 9.6 | 9.2 | 29.4 | 1.04x | 3.21x | 134.3 | 119.2 |
| `cohere/internal/lint/rules/tailwind/collapse/utility_nodes.go:138` | 253.6 | 105.2 | 85.8 | 2.41x | 0.82x | 3402.3 | 1294.9 |
| `cohere/internal/lint/rules/tailwind/collapse/utility_nodes.go:140` | 228.9 | 105.3 | 31.0 | 2.17x | 0.29x | 2952.2 | 1346.9 |
| `cohere/internal/lint/rules/tailwind/collapse/utility_nodes.go:142` | 464.5 | 92.8 | 125.9 | 5.00x | 1.36x | 6983.6 | 1156.0 |
| `cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:65` | 7.1 | 6.3 | 19.1 | 1.13x | 3.03x | 109.9 | 69.7 |
| `cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:66` | 7.4 | 4.7 | 18.9 | 1.58x | 4.00x | 110.9 | 69.2 |
| `cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:67` | 10.8 | 5.6 | 23.7 | 1.93x | 4.25x | 109.8 | 69.0 |
| `cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:68` | 6.9 | 3.9 | 37.1 | 1.79x | 9.58x | 106.4 | 67.4 |
| `cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:69` | 12.3 | 8.0 | 22.4 | 1.54x | 2.81x | 111.4 | 70.2 |
| `cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:70` | 9.0 | 5.5 | 19.7 | 1.63x | 3.58x | 108.1 | 68.1 |
| `cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:74` | 219.6 | 105.6 | 21.5 | 2.08x | 0.20x | 2777.4 | 1170.2 |
| `cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:75` | 194.0 | 80.0 | 21.8 | 2.42x | 0.27x | 2589.4 | 979.1 |
| `cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:76` | 184.3 | 83.4 | 22.7 | 2.21x | 0.27x | 2619.0 | 1007.1 |
| `cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:77` | 183.7 | 82.4 | 22.3 | 2.23x | 0.27x | 2609.7 | 997.8 |
| `cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:78` | 194.0 | 78.9 | 20.5 | 2.46x | 0.26x | 2598.6 | 988.2 |
| `cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:79` | 186.3 | 88.8 | 19.4 | 2.10x | 0.22x | 2663.7 | 1043.9 |
| `cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:83` | 8.1 | 4.5 | 17.9 | 1.78x | 3.96x | 109.8 | 69.0 |
| `cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:84` | 194.1 | 80.5 | 29.4 | 2.41x | 0.37x | 2577.9 | 968.3 |
| `cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:85` | 8.8 | 5.8 | 19.4 | 1.53x | 3.36x | 108.7 | 69.1 |
| `cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:86` | 186.8 | 80.6 | 20.4 | 2.32x | 0.25x | 2556.6 | 950.1 |
| `cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:88` | 8.1 | 5.0 | 19.9 | 1.61x | 3.96x | 113.7 | 70.9 |
| `cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:89` | 8.8 | 4.8 | 18.6 | 1.86x | 3.91x | 112.6 | 70.1 |
| `cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:90` | 7.6 | 5.0 | 18.9 | 1.53x | 3.80x | 112.6 | 70.1 |
| `cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:93` | 8.2 | 4.5 | 18.2 | 1.84x | 4.08x | 111.4 | 69.0 |
| `cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:94` | 8.5 | 5.3 | 19.2 | 1.62x | 3.65x | 113.1 | 69.7 |
| `cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:95` | 9.8 | 5.0 | 21.7 | 1.95x | 4.32x | 112.0 | 69.2 |
| `cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:96` | 8.6 | 4.9 | 19.5 | 1.76x | 3.98x | 113.7 | 69.9 |
| `cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:97` | 8.5 | 4.6 | 18.7 | 1.87x | 4.09x | 112.6 | 69.9 |
| `cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:98` | 7.8 | 4.9 | 18.6 | 1.61x | 3.82x | 114.2 | 70.6 |
| `cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:99` | 9.3 | 4.8 | 18.8 | 1.92x | 3.89x | 113.1 | 70.1 |
| `cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:100` | 7.9 | 4.9 | 17.8 | 1.62x | 3.66x | 114.8 | 70.8 |
| `cohere/internal/lint/rules/tailwind/no_physical_direction.go:284` | 1006.5 | 853.1 | 71.6 | 1.18x | 0.08x | 11196.7 | 9599.3 |
| `cohere/internal/lint/rules/tailwind/tools/generate_variant/main.go:221` | 722.9 | 717.1 | 51.5 | 1.01x | 0.07x | 9594.9 | 7677.1 |
| `cohere/internal/lint/rules/typescript/ban_tslint_comment.go:24` | 231.1 | 120.7 | 25.3 | 1.92x | 0.21x | 3068.2 | 1457.1 |
| `cohere/internal/lint/rules/typescript/no_unnecessary_template_expression.go:574` | 9.4 | 4.7 | 19.5 | 1.99x | 4.12x | 134.2 | 68.2 |
| `cohere/internal/lint/rules/typescript/switch_exhaustiveness_check.go:381` | 8.9 | 5.1 | 19.0 | 1.74x | 3.70x | 113.0 | 70.8 |
| `hard/keywords` | 1773451.2 | 72839.2 | 34372.4 | 24.35x | 0.47x | 19211865.0 | 328043.0 |
| `hard/email` | 2352015.3 | 62427.1 | 183507.2 | 37.68x | 2.94x | 25185223.0 | 266361.5 |
| `hard/ignorecase` | 28947.5 | 23378.9 | 3061.8 | 1.24x | 0.13x | 392643.5 | 332863.5 |
| `hard/lookbehind` | 514422.4 | 11976.1 | 55502.8 | 42.95x | 4.63x | 4971206.5 | 92014.0 |
| `original/eslint-default-case` | 8.2 | 5.2 | 18.7 | 1.59x | 3.62x | 113.0 | 70.8 |
| `original/typescript-switch-exhaustiveness` | 7.4 | 4.7 | 19.8 | 1.57x | 4.20x | 113.0 | 70.8 |
| `original/typescript-ban-tslint-comment` | 209.2 | 122.5 | 26.1 | 1.71x | 0.21x | 3074.2 | 1463.1 |
| `original/react-nested-components` | 217.4 | 109.0 | 21.0 | 1.99x | 0.19x | 2915.7 | 1310.0 |
| `original/react-sort-comp` | 203.5 | 47.7 | 133.7 | 4.27x | 2.81x | 2906.0 | 637.6 |
| `original/typescript-template-dollar` | 193.9 | 20.9 | 19.3 | 9.30x | 0.92x | 2747.4 | 262.9 |

## Binary footprint

| Binary | text | data | bss | total bytes |
|---|---:|---:|---:|---:|
| 8ac5675 | 738550 | 324320 | 65984 | 1128854 |
| Previous rollout d38492b | 832790 | 334056 | 65984 | 1232830 |
| Current repair | 841262 | 339192 | 65984 | 1246438 |

## What the regressed calls paid for

Callgrind separates the extra checks from matching work. For the configuration pattern's nine-input cycle, `8ac5675` spent 405 instructions in `adamic_regex_test` and 1,123 in its execution/search kernel. `d38492b` spent 765 and 1,374 respectively, even though this row never selected the regular engine. Input decoding remained 16,086 instructions in both. The increased dispatch and per-candidate checks account for 611 extra instructions, about 68 per call. Many tiny anchored rows paid another nine instructions per call. These are observed instruction counts, rather than an inference from elapsed time alone.

The repair emits the whole boolean search for forward straight-line patterns, preserving compile-time flags and canonicalized ASCII predicates. Boolean calls need no capture registers or instruction accounting in unlimited mode. Fixed widths reject insufficient input early. A trailing non-multiline end assertion determines the only possible start; sticky calls must already be at that start. Full capture calls still use the existing VM or regular search with VM recovery.

Short general searches use a separately compiled VM kernel whose candidate loop has no regular-engine or lookbehind-filter tests. The same source kernel is specialized with an `enhanced` compile-time constant, rather than maintained as two independent interpreters. A nonzero harness budget always uses the VM. Eligible long calls retain the lazy DFA and NFA search; forced-regular tests bypass the production length threshold. The hot boolean dispatcher reads cached integer string lengths directly and sends general cases to a separate function. The callback and its prefix-selection bit sit with the hot descriptor fields, and emitted descriptors are cache-line aligned. Ordinary literal searches reject missing prefixes in the caller, avoiding the extra indirect call on negative inputs; compile-time metadata excludes suffix and sticky searches from this prefilter. A first scalar predicate already proved by the emitted first-set scan is not checked twice.

ASCII input widening now uses a vectorizable byte-to-UTF-16 loop. The general WTF-8 decoder remains necessary for uncached and non-ASCII strings. Necessary literal windows after a fixed number of scalar characters use `memchr` and `memcmp` before VM setup, stopping analysis before branches, repetition, backreferences or variable-width sets. This removes repeated failed VM attempts in patterns such as `[a-z0-9]cfg($|[A-Z0-9])` and `([a-z])Ms($|[A-Z])`. These filters operate only on known ASCII inputs; the matcher retains full input context and evaluates all assertions and captures normally.

A narrow one-pass specialization also handles stateless, non-multiline anchored scalar stars followed by one scalar. It tests the tail before the repeated body, allowing zero repetitions, and stops when the body cannot consume. Boolean existence does not depend on greedy versus lazy end selection; global, sticky and capture operations retain their appropriate VM paths. This removes choice-point frames from patterns such as `^(.*):`.

The repair reduces work rather than claiming a smaller binary. Binary size and the complete instruction counts are reported below; a larger cold footprint remains a limitation.

## Validation and commands

All test output went to log files. The final runtime and emitter passed:

```sh
source /workspace/adamic-tools/env.sh
go test ./internal/native ./internal/regexp ./bench/regex -run 'TestRegExp|TestMatcher|TestRegular' -count=1 -timeout 15m -v > /tmp/regex-dispatch-selected-gate.log 2>&1
# PASS: native 184.451s, regexp 5.448s; benchmark has no test files.
go test ./internal/oracle -run 'TestNativeAgreesWithNode/.*/.*/.*/(regexp|sweeps)' -count=1 -timeout 15m -v > /tmp/regex-dispatch-selected-oracle.log 2>&1
# PASS 40.238s: regex fixtures and 63,960 regex method probes.
python3 internal/native/testdata/run-regexp-regular-mutants.py > /tmp/regex-dispatch-selected-mutants.log 2>&1
# All 22 semantic mutants caught, with every edit restored.
python3 bench/regex/run-dispatch-mutant.py /tmp/regex-regular-ir-before.csv /tmp/regex-dispatch-selected-performance-mutant > /tmp/regex-dispatch-selected-performance-mutant.log 2>&1
# Compiled successfully, then caught by instruction regressions on 35 rows.
go test ./internal/native ./internal/regexp -run '^(TestRegExp(Search|RegularEngine)Node|TestRegular.*)$' -count=1 -timeout 5m -v > /tmp/regex-dispatch-selected-restored.log 2>&1
# PASS after restoration, including 68 search controls in all three configurations.
gofmt -l cmd internal bench/regex > /tmp/regex-dispatch-selected-gofmt.log
# Empty output.
go vet ./... > /tmp/regex-dispatch-selected-vet.log 2>&1
# Exit 0.
git diff --check
# Exit 0.
```

Go and native each compared every one of 127,369 test262 executions and 10,000 fixed-seed randomized cases in three configurations: budgeted VM, unlimited regular search with VM recovery, and unlimited VM. Every configuration had zero disagreements and zero unavailable properties. Native compares both `test` and `exec`, captures as strings or undefined, capture indices, named groups and lastIndex, under ASan, UBSan and leak detection. The native test262 corpus alone covers 764,214 API operations across the configurations.

The regular engine qualifies for 127,024 test262 executions and 7,369 randomized executions; unsupported patterns keep the VM. Existing one-pass paths can serve eligible boolean tests. All 102 frozen benchmark rows' 890 inputs also agree with Node in every configuration. The 68 search controls include the new suffix, fixed-window, sticky, empty-pattern, lowercase-folding and non-ASCII widening witnesses. Ten regular priority/cache controls and 36 Go priority controls also pass. The final corpus run includes all 68 search controls.

The full repository test gate and allocation-count gate were not rerun in this repair. Regex package tests, the complete matching corpus, the filtered string-method oracle and repository-wide vet were run. Unchanged Unicode tables were not exhaustively rechecked over every code point. No protected emitter or lowering files were changed.

## Mutants

| Mutant | Witness |
|---|---|
| Caller prefix scan incorrectly enabled for sticky | Node sticky search control |
| First predicate omitted without a scan proof | Node `^a` on `b` |
| One-pass repeat skips body check | Node colon after newline control |
| One-pass repeat tail changed | Node anchored-repeat controls |
| One-pass repeat selected for stateful | Node greedy global `^(.*):` on `a:a:` |
| End-anchored sticky start ignored | Node `a$` with `y` on `xa` |
| Boolean minimum width doubled | Node search/capture controls |
| Literal-window offset changed | Node fixed-window controls |
| Boolean folding omitted | Node lowercase `z` and mixed-case `aZ` controls |
| Boolean end assertion removed | Node `a$b` on `ab` |
| Boolean sticky search advances | Node `a` with `y` on `xa` |
| ASCII widening used for non-ASCII | Node `\\u00C5` on `Ł` |
| Sparse first filter omits a candidate | Node search controls |
| Legacy search advances by code points | Node surrogate-half control |
| Go merges the latest start | Go Node priority controls |
| Go removes construction work limit | Loud three-second timeout |
| DFA selected for a backreference | Node `^(a)\\1$` on `a` and `aa` |
| Native merges the latest start | Node regular priority controls |
| First acceptance treated as leftmost | Earlier-start/later-end Node control |
| DFA overflow becomes no match | Node exponential-suffix language control |
| DFA forgets word context | Node word-boundary controls |
| Literal lookbehind context changed | Node lookbehind controls |
| DFA threshold reversed to select short inputs | Callgrind guard: 35 instruction regressions against `8ac5675` |

Twenty-two semantic mutants were caught by result disagreements, except the construction-work mutant, which was caught by timeout. The performance mutant compiled and completed measurement before failing the instruction guard: the configuration row used 1.5046x baseline instructions and keywords 5.5335x. During proof development, an earlier wrong-window-offset variant caused nonprogress and hit the harness timeout; the final variant advances and is caught by Node results. Folding proof initially lacked a lowercase witness; adding `z`, `aZ` and `az` made the real omitted-folding mutant fail.

## Reproduction and limits

Build `8ac5675` and this commit with the same cohere revision and toolchain, into separate output directories. The baseline binary here was retained from the previous validated phase. The final benchmark's C was regenerated after mutation restoration and compared byte for byte with the measured artifact.

```sh
source /workspace/adamic-tools/env.sh
export REGEXP_CALLGRIND_HEADER=/workspace/scratch/regex-speed-tools/usr/include/valgrind/callgrind.h
go run ./bench/regex -out /tmp/regex-dispatch-selected -build-only > /tmp/regex-dispatch-selected-build.log 2>&1
export CALLGRIND=/workspace/scratch/regex-speed-tools/usr/bin/valgrind
export VALGRIND_LIB=/workspace/scratch/regex-speed-tools/usr/libexec/valgrind
python3 bench/regex/count.py /tmp/regex-regular-before > /tmp/regex-before-ir.csv 2> /tmp/regex-before-ir.log
python3 bench/regex/count.py /tmp/regex-dispatch-selected > /tmp/regex-after-ir.csv 2> /tmp/regex-after-ir.log
python3 bench/regex/measure.py /tmp/regex-regular-before /tmp/regex-dispatch-selected --minimum-native-seconds 0.1 > /tmp/regex-times.csv 2> /tmp/regex-times.log
python3 bench/regex/check-dispatch.py /tmp/regex-before-ir.csv /tmp/regex-after-ir.csv --times /tmp/regex-times.csv > /tmp/regex-dispatch-check.log 2>&1
```

Setup printed Go ready 0s, clang ready 1s, Node ready 1s, submodules ready 1s, build cache warm 71s, total 71s. `nproc` is 5; the CPU quota is four cores. Node is 24.19.0, Go 1.27.1 and clang 20.1.8, on Intel Xeon Platinum 8573C. Valgrind 3.24.0 was already extracted locally. Native compilation uses the existing `-O2` benchmark settings.

These are warmed `RegExp.test` microbenchmarks on the frozen lint and hard-case inventory: 86 static Go patterns, six representative dynamic instantiations, six audited JavaScript originals and four hard cases. They do not establish whole-project lint throughput, capture-method throughput, cold compile performance or the same timing bound for arbitrary patterns, inputs and machines. All measured regressions, including ties within the declared noise allowance, remain in the table. Larger expansions, lookaround, backreferences and string-valued sets retain their existing VM paths.
