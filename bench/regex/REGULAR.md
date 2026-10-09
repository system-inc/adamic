# Regular engine performance and correctness

Baseline: `8ac567550ce006a535a703979e16f0061aa4c9bb`. Optimized: the commit containing this report. Frozen benchmark inputs and Node scripts are identical across builds. Hardware: Intel Xeon Platinum 8573C, five visible CPUs, four-core quota. Go 1.27.1, clang 20.1.8 (`-O2`), Node 24.19.0, Valgrind Callgrind.

Wall times are the best of five, rotating native-before, native-after and Node order. Compilation, process startup and warmup are outside the timed region; native warms 100 calls and Node 10,000. Callgrind measures one input cycle after warmup, excluding printing. Instructions and wall times below are per call; these short timings are microbenchmark observations, not whole-application throughput.

The four hard cases improve 22.2x (keywords), 44.0x (email), 1.63x (case-insensitive search) and 21.2x (lookbehind). Native now leads Node by 3.1x on email and 2.3x on lookbehind. Node remains 2.5x faster on keywords and 6.3x faster on case-insensitive search.

Across all 102 rows, native beats Node on 50. The median before/after speedup is 0.919x, and 66 rows are slower than baseline. The hard-case gains come with remaining short-pattern overhead; this change does not improve every lint pattern. The full measurements, including regressions, follow and are also available in [regular-results.csv](regular-results.csv).

## Per-pattern measurements

Times are nanoseconds per call. Speedup is before / after; values below one are regressions. Native / Node is after / Node; values below one favor native.

| Pattern | Before ns | After ns | Node ns | Speedup | Native / Node | Before instructions | After instructions |
|---|---:|---:|---:|---:|---:|---:|---:|
| cohere/internal/lint/configuration/configuration.go:839 | 184.4 | 191.9 | 84.0 | 0.961x | 2.285x | 2535.889 | 2603.778 |
| cohere/internal/lint/ecmascript/dotnotation/dotnotation.go:121 | 319.1 | 340.1 | 96.3 | 0.938x | 3.533x | 4205.556 | 4273.222 |
| cohere/internal/lint/ecmascript/dotnotation/dotnotation.go:141 | 200.4 | 188.3 | 60.3 | 1.064x | 3.121x | 2884.333 | 2952.222 |
| cohere/internal/lint/housesets/housesets.go:112 | 432.4 | 65.3 | 158.8 | 6.624x | 0.411x | 6912.222 | 864.444 |
| cohere/internal/lint/rules/base/consistency_require_pagination_argument_name.go:27 | 8.4 | 10.0 | 70.3 | 0.845x | 0.142x | 150.000 | 169.222 |
| cohere/internal/lint/rules/base/consistency_require_pagination_argument_name.go:30 | 8.2 | 9.5 | 71.3 | 0.862x | 0.133x | 147.111 | 166.889 |
| cohere/internal/lint/rules/core/default_case.go:29 | 6.5 | 7.6 | 103.6 | 0.846x | 0.074x | 113.000 | 122.000 |
| cohere/internal/lint/rules/core/no_inline_comments.go:38 | 205.5 | 230.1 | 74.5 | 0.893x | 3.088x | 2969.444 | 3037.222 |
| cohere/internal/lint/rules/core/no_warning_comments.go:280 | 280.6 | 225.3 | 171.8 | 1.246x | 1.311x | 3347.444 | 3034.556 |
| cohere/internal/lint/rules/core/require_description.go:214 | 1197.9 | 1345.4 | 104.4 | 0.890x | 12.891x | 15742.111 | 16492.444 |
| cohere/internal/lint/rules/core/require_description.go:217 | 242.4 | 279.5 | 64.3 | 0.867x | 4.347x | 3525.556 | 3593.333 |
| cohere/internal/lint/rules/core/require_description.go:222 | 193.5 | 213.5 | 72.3 | 0.906x | 2.951x | 2927.000 | 2994.889 |
| cohere/internal/lint/rules/core/require_description.go:233 | 190.5 | 188.6 | 60.4 | 1.010x | 3.124x | 2839.667 | 2907.556 |
| cohere/internal/lint/rules/next/no_html_link_for_pages_routes.go:139 | 157.7 | 17.4 | 63.1 | 9.055x | 0.276x | 2533.556 | 274.556 |
| cohere/internal/lint/rules/next/no_html_link_for_pages_routes.go:140 | 159.4 | 165.8 | 60.0 | 0.961x | 2.766x | 2537.556 | 2605.444 |
| cohere/internal/lint/rules/next/no_html_link_for_pages_routes.go:141 | 164.3 | 171.9 | 94.7 | 0.956x | 1.814x | 2517.778 | 2585.667 |
| cohere/internal/lint/rules/next/no_html_link_for_pages_routes.go:142 | 182.4 | 185.5 | 68.1 | 0.983x | 2.725x | 2538.889 | 2606.778 |
| cohere/internal/lint/rules/next/no_html_link_for_pages_routes.go:251 | 161.5 | 20.3 | 180.5 | 7.951x | 0.112x | 2561.000 | 302.556 |
| cohere/internal/lint/rules/next/no_location_assign_relative_destination.go:26 | 517.8 | 502.5 | 83.4 | 1.031x | 6.024x | 6196.889 | 6264.778 |
| cohere/internal/lint/rules/nexus/abbreviation_vocabulary.go:162 | 6.1 | 7.5 | 68.8 | 0.821x | 0.109x | 105.889 | 114.889 |
| cohere/internal/lint/rules/nexus/abbreviation_vocabulary.go:176 | 1934.2 | 2114.6 | 149.6 | 0.915x | 14.138x | 22843.000 | 23572.556 |
| cohere/internal/lint/rules/nexus/abbreviation_vocabulary.go:177 | 4557.0 | 4811.9 | 113.7 | 0.947x | 42.318x | 48894.333 | 52921.778 |
| cohere/internal/lint/rules/nexus/abbreviation_vocabulary.go:178 | 207.6 | 223.6 | 101.4 | 0.929x | 2.205x | 3004.444 | 2925.778 |
| cohere/internal/lint/rules/nexus/consistency_no_abbreviated_identifier.go:50 | 5203.0 | 6425.3 | 129.7 | 0.810x | 49.558x | 55369.111 | 59385.000 |
| cohere/internal/lint/rules/nexus/consistency_no_ambiguous_identifier.go:17 | 2617.2 | 2836.2 | 292.8 | 0.923x | 9.686x | 30828.556 | 31590.222 |
| cohere/internal/lint/rules/nexus/consistency_no_ambiguous_identifier.go:18 | 5.8 | 7.2 | 58.1 | 0.800x | 0.125x | 105.889 | 114.889 |
| cohere/internal/lint/rules/nexus/consistency_no_ambiguous_identifier.go:19 | 6.4 | 7.8 | 55.8 | 0.828x | 0.139x | 109.333 | 118.333 |
| cohere/internal/lint/rules/nexus/consistency_no_boolean_outcome.go:48 | 616.1 | 182.7 | 170.8 | 3.373x | 1.069x | 8275.556 | 2168.000 |
| cohere/internal/lint/rules/nexus/consistency_no_return_void.go:100 | 220.2 | 271.6 | 68.7 | 0.811x | 3.953x | 3222.556 | 3290.444 |
| cohere/internal/lint/rules/nexus/consistency_no_return_void.go:116 | 187.3 | 206.8 | 73.9 | 0.905x | 2.799x | 2864.889 | 2932.778 |
| cohere/internal/lint/rules/nexus/consistency_require_constant_casing.go:350 | 427.9 | 55.0 | 82.0 | 7.779x | 0.671x | 6705.778 | 685.667 |
| cohere/internal/lint/rules/nexus/consistency_require_matching_return_type.go:216 | 204.6 | 191.0 | 68.2 | 1.072x | 2.800x | 2847.667 | 2915.444 |
| cohere/internal/lint/rules/nexus/shouting.go:121 | 108.3 | 84.8 | 108.0 | 1.278x | 0.785x | 1424.778 | 1433.778 |
| cohere/internal/lint/rules/nexus/shouting.go:122 | 106.8 | 107.1 | 98.3 | 0.997x | 1.090x | 1400.444 | 1409.444 |
| cohere/internal/lint/rules/nexus/shouting.go:123 | 473.1 | 546.0 | 166.7 | 0.867x | 3.275x | 6742.000 | 6536.222 |
| cohere/internal/lint/rules/nexus/shouting.go:147 | 200.4 | 46.1 | 99.2 | 4.347x | 0.465x | 2814.444 | 680.667 |
| cohere/internal/lint/rules/nexus/shouting.go:153 | 168.8 | 29.4 | 172.8 | 5.748x | 0.170x | 2578.556 | 444.000 |
| cohere/internal/lint/rules/nexus/shouting.go:160 | 155.4 | 17.3 | 170.3 | 8.974x | 0.102x | 2522.000 | 263.556 |
| cohere/internal/lint/rules/nexus/shouting.go:166 | 220.3 | 90.4 | 174.3 | 2.436x | 0.519x | 3216.444 | 1229.889 |
| cohere/internal/lint/rules/nexus/shouting.go:175 | 237.1 | 99.6 | 164.1 | 2.381x | 0.607x | 3228.556 | 1240.111 |
| cohere/internal/lint/rules/nexus/shouting.go:177 | 6.6 | 7.5 | 66.4 | 0.877x | 0.113x | 121.556 | 130.556 |
| cohere/internal/lint/rules/nexus/shouting.go:178 | 226.2 | 274.7 | 75.5 | 0.824x | 3.636x | 3389.556 | 3456.333 |
| cohere/internal/lint/rules/nexus/shouting.go:181 | 189.6 | 207.9 | 86.4 | 0.912x | 2.405x | 2827.444 | 2895.222 |
| cohere/internal/lint/rules/nexus/shouting.go:182 | 160.0 | 165.1 | 68.1 | 0.969x | 2.425x | 2503.889 | 2571.778 |
| cohere/internal/lint/rules/nexus/shouting.go:183 | 168.9 | 158.0 | 57.1 | 1.069x | 2.766x | 2469.444 | 2537.333 |
| cohere/internal/lint/rules/nexus/shouting.go:184 | 357.4 | 343.6 | 61.9 | 1.040x | 5.548x | 4592.222 | 4660.111 |
| cohere/internal/lint/rules/nexus/shouting.go:464 | 404.8 | 473.4 | 133.5 | 0.855x | 3.547x | 5035.889 | 4970.111 |
| cohere/internal/lint/rules/react/boolean_prop_naming.go:621 | 174.6 | 25.7 | 118.9 | 6.793x | 0.216x | 2639.778 | 377.667 |
| cohere/internal/lint/rules/react/conformance/classify.go:95 | 3467.8 | 3889.7 | 240.3 | 0.892x | 16.186x | 42844.556 | 44324.000 |
| cohere/internal/lint/rules/react/conformance/classify.go:98 | 290.9 | 260.6 | 134.4 | 1.116x | 1.940x | 3886.333 | 3514.333 |
| cohere/internal/lint/rules/react/exhaustive_deps.go:451 | 182.4 | 42.5 | 107.6 | 4.294x | 0.395x | 2667.667 | 612.778 |
| cohere/internal/lint/rules/react/no_unstable_nested_components.go:450 | 187.5 | 183.0 | 72.7 | 1.024x | 2.518x | 2731.333 | 2799.111 |
| cohere/internal/lint/rules/react/sort_comp.go:119 | 182.6 | 24.1 | 168.5 | 7.566x | 0.143x | 2622.222 | 350.667 |
| cohere/internal/lint/rules/tailwind/class_existence.go:63 | 196.9 | 208.8 | 66.8 | 0.943x | 3.127x | 2787.889 | 2855.778 |
| cohere/internal/lint/rules/tailwind/class_existence.go:70 | 161.1 | 204.1 | 73.2 | 0.789x | 2.789x | 2580.667 | 2648.556 |
| cohere/internal/lint/rules/tailwind/class_order_strict.go:89 | 8416.2 | 9201.5 | 413.0 | 0.915x | 22.280x | 106308.111 | 106376.000 |
| cohere/internal/lint/rules/tailwind/collapse/utility_nodes.go:131 | 7.5 | 8.4 | 84.4 | 0.887x | 0.099x | 134.333 | 146.889 |
| cohere/internal/lint/rules/tailwind/collapse/utility_nodes.go:138 | 238.8 | 102.5 | 149.3 | 2.330x | 0.686x | 3402.333 | 1379.444 |
| cohere/internal/lint/rules/tailwind/collapse/utility_nodes.go:140 | 225.9 | 215.9 | 70.6 | 1.046x | 3.059x | 2952.222 | 3003.222 |
| cohere/internal/lint/rules/tailwind/collapse/utility_nodes.go:142 | 444.3 | 98.8 | 164.9 | 4.497x | 0.599x | 6983.556 | 1263.667 |
| cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:65 | 6.1 | 7.1 | 79.0 | 0.855x | 0.090x | 109.889 | 118.889 |
| cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:66 | 6.4 | 7.9 | 59.8 | 0.811x | 0.132x | 110.889 | 119.889 |
| cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:67 | 6.3 | 7.5 | 70.4 | 0.846x | 0.106x | 109.778 | 118.778 |
| cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:68 | 6.2 | 7.3 | 77.1 | 0.849x | 0.094x | 106.444 | 115.444 |
| cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:69 | 6.7 | 8.0 | 71.2 | 0.835x | 0.113x | 111.444 | 120.444 |
| cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:70 | 6.6 | 7.8 | 60.7 | 0.846x | 0.128x | 108.111 | 117.111 |
| cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:74 | 177.7 | 184.6 | 76.5 | 0.962x | 2.413x | 2777.444 | 2845.222 |
| cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:75 | 157.7 | 191.7 | 59.8 | 0.823x | 3.206x | 2589.444 | 2657.333 |
| cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:76 | 180.6 | 192.9 | 65.4 | 0.936x | 2.950x | 2619.000 | 2686.889 |
| cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:77 | 180.2 | 171.7 | 92.8 | 1.050x | 1.849x | 2609.667 | 2677.556 |
| cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:78 | 161.0 | 174.9 | 72.3 | 0.921x | 2.419x | 2598.556 | 2666.444 |
| cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:79 | 168.2 | 196.8 | 85.2 | 0.855x | 2.310x | 2663.667 | 2731.556 |
| cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:83 | 6.1 | 7.6 | 94.6 | 0.799x | 0.081x | 109.778 | 118.778 |
| cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:84 | 182.2 | 169.4 | 72.3 | 1.076x | 2.344x | 2577.889 | 2645.778 |
| cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:85 | 6.1 | 7.2 | 59.9 | 0.841x | 0.120x | 108.667 | 117.667 |
| cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:86 | 175.4 | 195.4 | 85.5 | 0.898x | 2.285x | 2556.556 | 2624.444 |
| cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:88 | 6.4 | 7.8 | 71.3 | 0.820x | 0.109x | 113.667 | 122.667 |
| cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:89 | 7.0 | 7.7 | 109.2 | 0.905x | 0.071x | 112.556 | 121.556 |
| cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:90 | 6.1 | 7.7 | 71.6 | 0.798x | 0.107x | 112.556 | 121.556 |
| cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:93 | 6.2 | 8.1 | 84.5 | 0.773x | 0.095x | 111.444 | 120.444 |
| cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:94 | 6.3 | 8.4 | 83.8 | 0.758x | 0.100x | 113.111 | 122.111 |
| cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:95 | 6.5 | 7.5 | 62.6 | 0.873x | 0.119x | 112.000 | 121.000 |
| cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:96 | 6.7 | 7.6 | 58.6 | 0.885x | 0.129x | 113.667 | 122.667 |
| cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:97 | 6.8 | 7.8 | 81.9 | 0.865x | 0.095x | 112.556 | 121.556 |
| cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:98 | 7.2 | 8.1 | 85.7 | 0.885x | 0.094x | 114.222 | 123.222 |
| cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:99 | 6.4 | 8.8 | 71.6 | 0.727x | 0.123x | 113.111 | 122.111 |
| cohere/internal/lint/rules/tailwind/no_deprecated_classes.go:100 | 6.5 | 7.9 | 75.6 | 0.823x | 0.104x | 114.778 | 123.778 |
| cohere/internal/lint/rules/tailwind/no_physical_direction.go:284 | 939.2 | 978.0 | 118.4 | 0.960x | 8.257x | 11196.667 | 11264.333 |
| cohere/internal/lint/rules/tailwind/tools/generate_variant/main.go:221 | 675.2 | 1001.7 | 122.6 | 0.674x | 8.172x | 9594.889 | 9302.889 |
| cohere/internal/lint/rules/typescript/ban_tslint_comment.go:24 | 200.3 | 218.4 | 73.0 | 0.917x | 2.990x | 3068.222 | 3136.111 |
| cohere/internal/lint/rules/typescript/no_unnecessary_template_expression.go:574 | 7.4 | 8.8 | 83.5 | 0.849x | 0.105x | 134.222 | 147.000 |
| cohere/internal/lint/rules/typescript/switch_exhaustiveness_check.go:381 | 6.4 | 8.2 | 82.1 | 0.780x | 0.100x | 113.000 | 122.000 |
| hard/keywords | 1801056.0 | 81270.1 | 32756.2 | 22.161x | 2.481x | 19211865.000 | 328039.000 |
| hard/email | 2796201.8 | 63611.6 | 199295.4 | 43.957x | 0.319x | 25185223.000 | 266375.500 |
| hard/ignorecase | 39231.7 | 24021.5 | 3829.4 | 1.633x | 6.273x | 392643.500 | 332911.000 |
| hard/lookbehind | 589146.5 | 27741.7 | 64918.6 | 21.237x | 0.427x | 4971206.500 | 267739.000 |
| original/eslint-default-case | 6.4 | 8.2 | 62.3 | 0.773x | 0.132x | 113.000 | 122.000 |
| original/typescript-switch-exhaustiveness | 6.0 | 7.6 | 75.8 | 0.796x | 0.100x | 113.000 | 122.000 |
| original/typescript-ban-tslint-comment | 370.4 | 312.3 | 69.6 | 1.186x | 4.488x | 3074.222 | 3142.111 |
| original/react-nested-components | 214.0 | 213.5 | 69.1 | 1.002x | 3.091x | 2915.667 | 2983.444 |
| original/react-sort-comp | 205.8 | 46.1 | 164.6 | 4.462x | 0.280x | 2906.000 | 635.889 |
| original/typescript-template-dollar | 207.8 | 17.7 | 104.0 | 11.713x | 0.171x | 2747.444 | 259.556 |

## Engine and semantics

Stage 0 derives a capture-free Thompson automaton from the existing bytecode. Eligibility rejects every backreference and lookaround, reversed consumption and string-valued sets. Quantifier bounds above 32, expansions above 768 states, or construction above 6,144 recursive segment calls retain the existing VM. The work limit matters for nested empty counted bodies that consume no states. The emitted C contains the automaton, ASCII alphabet classes and consuming masks; no syntax tree reaches the runtime.

For non-stateful ASCII `test`, the second engine lazily interns DFA states and transitions. A state includes the NFA subset and previous-character context for anchors and word boundaries. Each call owns a bounded 128-state stack cache. Cache exhaustion returns “use fallback”, never “no match”. No shared mutable cache or heap allocation is introduced.

For captures and stateful calls, a bitset NFA search tracks the earliest start at each state. Threads from an earlier start finish before selecting a candidate, even if a later start has already accepted. The backtracking VM then runs once at that start against the full input to recover ECMAScript's preferred end, captures and resets. Recovery deliberately keeps the original input context rather than slicing it, so `$` and word boundaries remain correct. The existence-only DFA need not choose a greedy or lazy end for a non-stateful boolean result.

Eligibility is decided at compile time. Production retains the existing runner for inputs shorter than 2,048 units and for patterns with an existing straight-line ASCII runner, avoiding cache setup on short calls. Tests force the regular search for eligible captures independently of this threshold. Non-ASCII capture search uses the code-point-aware NFA; the lazy DFA specialization handles ASCII. A nonzero test instruction budget uses the VM, preserving loud exhaustion rather than reporting a failed match.

Two supporting optimizations accompany the engine: tiny necessary unanchored ASCII first-character sets use emitted `memchr` scanners before UTF-16 conversion, also jumping directly to a candidate in straight-line runners; a leading positive literal ASCII lookbehind supplies a necessary context filter. All lookbehind evaluation and captures still use the VM. Candidate advancement reuses the existing UTF-16/code-point advance helper. Static VM helpers remain local, with equivalent exported wrappers for the second engine.

## Validation

Final commands, all with output redirected to logs:

```sh
source /workspace/adamic-tools/env.sh
go test ./internal/native ./internal/regexp ./bench/regex -run 'TestRegExp|TestMatcher|TestRegular' -count=1 -timeout 15m -v > /tmp/regex-regular-final-gate.log 2>&1
# PASS: native 176.515s, regexp 5.933s; benchmark has no test files.
go test ./internal/oracle -run 'TestNativeAgreesWithNode/.*/.*/.*/(regexp|sweeps)' -count=1 -timeout 15m -v > /tmp/regex-regular-final-oracle.log 2>&1
# PASS 43.351s: regex fixtures and sweeps, including 63,960 regex method probes.
python3 internal/native/testdata/run-regexp-regular-mutants.py > /tmp/regex-regular-final-mutants.log 2>&1
# All ten mutants caught; each source edit restored.
go test ./internal/native ./internal/regexp -run '^(TestRegExp(Search|RegularEngine)Node|TestRegular.*)$' -count=1 -timeout 5m -v > /tmp/regex-regular-final-restored.log 2>&1
# PASS after restoration.
gofmt -l cmd internal bench/regex > /tmp/regex-regular-release-gofmt.log
# Empty output.
go vet ./... > /tmp/regex-regular-release-vet.log 2>&1
# Exit 0.
git diff --check
# Exit 0.
```

Go and native each checked all 127,369 test262 execution cases and 10,000 fixed-seed randomized cases in three configurations: budgeted VM, unlimited regular search with VM recovery, and unlimited VM. Every configuration had zero disagreements and zero unavailable properties. The regular engine qualified for 127,024 test262 executions (2,547 unique native patterns) and 7,369 randomized executions (5,783 patterns); the remaining cases retained the VM. The existing straight-line path can still serve eligible boolean tests. Native checks cover both `test` and `exec`, capture strings and indices, named groups and lastIndex, under ASan, UBSan and leak detection.

Every frozen benchmark input was also checked separately against Node: 102 rows, 890 inputs, zero disagreements in each configuration. Another 44 search controls cover Unicode traversal and first/context filters; ten priority/cache controls cover capture recovery and cache exhaustion. Go adds 36 Node priority controls and a counted-empty construction control. The native corpus alone checks 764,214 API operations across the three configurations.

The complete touched-package gate also passed (`go test ./internal/native ./internal/regexp ./bench/regex -count=1 -timeout 15m -v`, native 488.990s, regexp 5.440s), before the final scanner refinements; all regex and matcher tests and the method oracle were rerun afterwards as listed above. The full allocation-count gate passed (`go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 15m`, 429.937s) after the regular engine was added. The final scanner changes introduce no counted allocations, and the checked-in table is unchanged. The full repository test gate was not run. Unchanged Unicode tables were not exhaustively rechecked over every code point in this performance step.

The final qualifier skips the sparse scanner for anchored patterns. After that dispatch-only refinement, `go test ./internal/native -run 'TestRegExp(Search|LintPatterns|RegularEngine)Node$' -count=1 -timeout 5m -v` passed in 23.160s, including every benchmark input in all three configurations.

## Mutants

These change real compiler/runtime behavior. Nine were caught by Node result disagreements; the construction-work mutant was caught by the test timeout, not a compiler warning.

| Mutant | Check that caught it |
|---|---|
| Sparse first filter omits a valid candidate | native search controls |
| Search advances by code points in legacy mode | surrogate-half search control |
| Go merges the latest start | Go Node priority controls |
| Go removes the empty-expansion work limit | `TestRegularCompileBudget`, loud three-second timeout |
| DFA selected for a backreference | `^(a)\1$` on `a` and `aa`, native Node disagreement |
| Native merges the latest start | native priority/cache controls |
| First acceptance treated as leftmost | earlier-start/later-end priority control |
| DFA cache exhaustion becomes no match | exponential-suffix language control |
| DFA forgets previous word context | native word-boundary controls |
| Literal lookbehind context changed | native lookbehind controls |

## Reproduction and scope

Build an `8ac5675` checkout into `/tmp/regex-regular-before` and the optimized checkout into `/tmp/regex-regular-release`, using the same command in each checkout:

```sh
source /workspace/adamic-tools/env.sh
export REGEXP_CALLGRIND_HEADER=/usr/include/valgrind/callgrind.h
go run ./bench/regex -out /tmp/regex-regular-release -build-only > /tmp/regex-build.log 2>&1
python3 bench/regex/measure.py /tmp/regex-regular-before /tmp/regex-regular-release > /tmp/regex-times.csv 2> /tmp/regex-times.log
python3 bench/regex/count.py /tmp/regex-regular-before > /tmp/regex-before-ir.csv 2> /tmp/regex-before-ir.log
python3 bench/regex/count.py /tmp/regex-regular-release > /tmp/regex-after-ir.csv 2> /tmp/regex-after-ir.log
```

Here Valgrind was already extracted locally from the previous step. `CALLGRIND=/workspace/scratch/regex-speed-tools/usr/bin/valgrind` and `VALGRIND_LIB=/workspace/scratch/regex-speed-tools/usr/libexec/valgrind` selected it; the header was `/workspace/scratch/regex-speed-tools/usr/include/valgrind/callgrind.h`. Toolchain setup printed Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s, build cache warm 75s, total 75s; `nproc` is 5 and CPU quota four cores.

This is a warmed `test` microbenchmark, not whole-project lint throughput or capture-method throughput. The input inventory contains 86 static Go lint patterns, six representative dynamic instantiations, six audited JavaScript originals, and four hard cases. Other Go rows are syntax translations; dynamic vocabularies and all possible inputs are not exhausted. Existing sources and hashes are recorded in `original-sources.json`. Stateful/capture searches use tagged NFA search plus VM recovery rather than a prioritized capture DFA, and string-valued sets and larger expansions retain the VM. The compile-time qualifier and runtime size/cache fallbacks are deliberate resource limits, not changes to match results.
