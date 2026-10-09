# Native regex performance

Baseline: `59d82c5de172e4e53d1224553b319d076f22424a`. Optimized: the commit containing this report. Measured on the same Intel Xeon Platinum 8573C cloud worker, `nproc=5`, CPU quota four cores. Go 1.27.1, clang 20.1.8 at the normal native `-O2` settings, Node 24.19.0, Valgrind 3.24.0.

102 rows: 86 statically resolved production `regexp.MustCompile` calls, six representative instantiations of dynamic calls, six independently checked JavaScript originals, and four hard cases. Test-only Go files are excluded. Input strings, patterns and flags are frozen in [cases.json](cases.json).

Native improved on all 102 rows. Median wall-time speedup: 36.6x. Median instruction reduction: 96.1%. Native beat warmed Node on 34 rows. Node remains faster on every hard case. These are observations on this synthetic input mix, not a claim about whole-project lint throughput.

## Measurement

Each wall-time cell is the best of five warmed matching loops, with baseline, optimized native, and Node interleaved in rotating order. Compilation, process startup and warmup are outside the timers. Both native and Node include loop and input-selection overhead. The native loop warms 100 calls; Node warms 10,000. Node `--trace-regexp-tier-up` confirmed generated native code for the keyword case.

The lint rows cycle nine strings: a generated positive witness plus representative identifiers, source, prose and comments; 9,000 calls per timed round. The four hard rows cycle a hit and miss over approximately 15 KiB of text; 200 calls per timed round. The email pattern is deliberately a modest email-like recognizer. The keyword alternation contains 64 keywords.

Callgrind counts native instructions over one complete input cycle, with client requests starting instrumentation after warmup and stopping it before printing and teardown. Values below divide the totals by the number of inputs. Instruction counts include the matching loop, UTF-16 conversion, VM work and temporary-buffer cleanup. The collector reads `totals:`, because explicit client dumps here reported zero in `summary:`. Zero counts fail the collector.

The baseline binary uses the original runtime sources from `59d82c5`, with the expanded program descriptor header so the same emitted bytecode and benchmark driver can be used. The old interpreter ignores the new search and callback fields. Native build optimization flags are the same.

## Per pattern

Times are nanoseconds per call. `Ir` is Callgrind instruction count per call. The common `cohere/internal/lint/` prefix is omitted from row labels. Exact seconds and checksums are in [results.csv](results.csv).

| Pattern | Native before ns | Native after ns | Node ns | Speedup | Ir before | Ir after |
|---|---:|---:|---:|---:|---:|---:|
| `configuration/configuration.go:839` | 6,010.5 | 161.6 | 61.4 | 37.19x | 65,459.7 | 2,535.9 |
| `ecmascript/dotnotation/dotnotation.go:121` | 6,339.3 | 331.3 | 95.3 | 19.14x | 69,606.1 | 4,205.6 |
| `ecmascript/dotnotation/dotnotation.go:141` | 6,004.1 | 224.3 | 84.1 | 26.77x | 65,034.2 | 2,884.3 |
| `housesets/housesets.go:112` | 15,953.1 | 422.5 | 154.2 | 37.76x | 177,475.3 | 6,912.2 |
| `rules/base/consistency_require_pagination_argument_name.go:27` | 7,370.3 | 8.4 | 106.6 | 882.11x | 86,130.3 | 150.0 |
| `rules/base/consistency_require_pagination_argument_name.go:30` | 7,718.4 | 8.2 | 59.6 | 943.97x | 86,004.2 | 147.1 |
| `rules/core/default_case.go:29` | 5,745.5 | 6.3 | 62.9 | 913.32x | 64,488.2 | 113.0 |
| `rules/core/no_inline_comments.go:38` | 5,725.5 | 211.8 | 72.7 | 27.03x | 62,486.7 | 2,969.4 |
| `rules/core/no_warning_comments.go:280` | 6,909.6 | 236.7 | 166.4 | 29.19x | 78,185.7 | 3,347.4 |
| `rules/core/require_description.go:214` | 9,854.2 | 1,288.2 | 105.3 | 7.65x | 106,728.9 | 15,742.1 |
| `rules/core/require_description.go:217` | 6,069.0 | 252.2 | 91.6 | 24.06x | 66,319.9 | 3,525.6 |
| `rules/core/require_description.go:222` | 5,780.6 | 206.0 | 78.4 | 28.06x | 65,277.8 | 2,927.0 |
| `rules/core/require_description.go:233` | 5,905.9 | 197.1 | 78.0 | 29.96x | 65,279.0 | 2,839.7 |
| `rules/next/no_html_link_for_pages_routes.go:139` | 8,004.9 | 161.1 | 73.5 | 49.68x | 85,644.9 | 2,533.6 |
| `rules/next/no_html_link_for_pages_routes.go:140` | 5,769.9 | 174.3 | 78.4 | 33.10x | 64,077.9 | 2,537.6 |
| `rules/next/no_html_link_for_pages_routes.go:141` | 5,918.3 | 155.2 | 72.3 | 38.12x | 64,031.7 | 2,517.8 |
| `rules/next/no_html_link_for_pages_routes.go:142` | 5,853.1 | 174.9 | 67.9 | 33.46x | 64,081.8 | 2,538.9 |
| `rules/next/no_html_link_for_pages_routes.go:251` | 7,840.2 | 175.4 | 174.3 | 44.70x | 85,787.0 | 2,561.0 |
| `rules/next/no_location_assign_relative_destination.go:26` | 7,509.9 | 468.2 | 92.9 | 16.04x | 78,980.2 | 6,196.9 |
| `rules/nexus/abbreviation_vocabulary.go:162` | 5,847.5 | 6.4 | 66.8 | 908.75x | 64,106.3 | 105.9 |
| `rules/nexus/abbreviation_vocabulary.go:176` | 15,777.8 | 1,894.6 | 141.9 | 8.33x | 165,118.2 | 22,843.0 |
| `rules/nexus/abbreviation_vocabulary.go:177` | 9,976.9 | 4,283.8 | 99.2 | 2.33x | 111,768.0 | 48,894.3 |
| `rules/nexus/abbreviation_vocabulary.go:178` | 7,922.5 | 199.1 | 134.4 | 39.79x | 87,015.2 | 3,004.4 |
| `rules/nexus/consistency_no_abbreviated_identifier.go:50` | 10,427.4 | 4,792.3 | 123.5 | 2.18x | 115,892.2 | 55,369.1 |
| `rules/nexus/consistency_no_ambiguous_identifier.go:17` | 30,730.1 | 2,660.5 | 325.8 | 11.55x | 319,120.0 | 30,828.6 |
| `rules/nexus/consistency_no_ambiguous_identifier.go:18` | 5,877.9 | 5.9 | 93.8 | 1003.55x | 64,103.1 | 105.9 |
| `rules/nexus/consistency_no_ambiguous_identifier.go:19` | 5,681.4 | 6.3 | 59.4 | 895.88x | 64,065.3 | 109.3 |
| `rules/nexus/consistency_no_boolean_outcome.go:48` | 49,973.7 | 579.4 | 179.6 | 86.25x | 545,424.6 | 8,275.6 |
| `rules/nexus/consistency_no_return_void.go:100` | 6,086.3 | 224.7 | 78.1 | 27.08x | 66,230.2 | 3,222.6 |
| `rules/nexus/consistency_no_return_void.go:116` | 6,144.5 | 182.7 | 92.6 | 33.63x | 65,141.3 | 2,864.9 |
| `rules/nexus/consistency_require_constant_casing.go:350` | 16,520.8 | 492.8 | 61.1 | 33.52x | 180,116.1 | 6,705.8 |
| `rules/nexus/consistency_require_matching_return_type.go:216` | 5,441.4 | 190.8 | 78.3 | 28.53x | 64,106.4 | 2,847.7 |
| `rules/nexus/shouting.go:121` | 7,535.2 | 111.7 | 134.6 | 67.44x | 85,677.8 | 1,424.8 |
| `rules/nexus/shouting.go:122` | 7,458.0 | 105.8 | 79.2 | 70.49x | 85,547.4 | 1,400.4 |
| `rules/nexus/shouting.go:123` | 7,069.1 | 425.8 | 175.1 | 16.60x | 78,899.4 | 6,742.0 |
| `rules/nexus/shouting.go:147` | 8,037.6 | 184.2 | 97.8 | 43.63x | 85,806.6 | 2,814.4 |
| `rules/nexus/shouting.go:153` | 7,913.2 | 164.0 | 174.6 | 48.25x | 85,156.4 | 2,578.6 |
| `rules/nexus/shouting.go:160` | 7,574.9 | 160.3 | 175.0 | 47.26x | 85,614.1 | 2,522.0 |
| `rules/nexus/shouting.go:166` | 7,457.2 | 251.5 | 196.3 | 29.65x | 86,013.4 | 3,216.4 |
| `rules/nexus/shouting.go:175` | 7,535.3 | 221.3 | 185.6 | 34.06x | 86,713.2 | 3,228.6 |
| `rules/nexus/shouting.go:177` | 522.2 | 6.6 | 71.5 | 79.13x | 6,070.1 | 121.6 |
| `rules/nexus/shouting.go:178` | 806.6 | 234.8 | 70.5 | 3.43x | 9,020.0 | 3,389.6 |
| `rules/nexus/shouting.go:181` | 5,603.1 | 182.7 | 76.0 | 30.66x | 63,382.7 | 2,827.4 |
| `rules/nexus/shouting.go:182` | 5,670.5 | 164.2 | 68.3 | 34.53x | 64,214.7 | 2,503.9 |
| `rules/nexus/shouting.go:183` | 5,820.5 | 152.6 | 114.2 | 38.15x | 63,971.1 | 2,469.4 |
| `rules/nexus/shouting.go:184` | 6,122.9 | 335.9 | 71.5 | 18.23x | 69,940.2 | 4,592.2 |
| `rules/nexus/shouting.go:464` | 8,934.2 | 393.4 | 128.4 | 22.71x | 89,778.4 | 5,035.9 |
| `rules/react/boolean_prop_naming.go:621` | 8,378.6 | 160.7 | 113.7 | 52.15x | 86,695.4 | 2,639.8 |
| `rules/react/conformance/classify.go:95` | 22,777.4 | 3,608.2 | 209.7 | 6.31x | 244,641.8 | 42,844.6 |
| `rules/react/conformance/classify.go:98` | 7,576.2 | 273.8 | 128.2 | 27.67x | 81,217.7 | 3,886.3 |
| `rules/react/exhaustive_deps.go:451` | 7,793.4 | 176.4 | 110.4 | 44.19x | 86,055.8 | 2,667.7 |
| `rules/react/no_unstable_nested_components.go:450` | 5,730.3 | 204.6 | 67.5 | 28.00x | 64,599.0 | 2,731.3 |
| `rules/react/sort_comp.go:119` | 8,435.6 | 177.7 | 161.6 | 47.48x | 86,262.7 | 2,622.2 |
| `rules/tailwind/class_existence.go:63` | 5,612.8 | 181.7 | 76.4 | 30.89x | 65,016.7 | 2,787.9 |
| `rules/tailwind/class_existence.go:70` | 5,364.4 | 194.4 | 69.6 | 27.59x | 64,375.0 | 2,580.7 |
| `rules/tailwind/class_order_strict.go:89` | 30,633.7 | 8,794.7 | 360.8 | 3.48x | 332,676.1 | 106,308.1 |
| `rules/tailwind/collapse/utility_nodes.go:131` | 7,859.1 | 7.5 | 91.6 | 1046.90x | 85,779.0 | 134.3 |
| `rules/tailwind/collapse/utility_nodes.go:138` | 8,153.3 | 228.6 | 135.1 | 35.66x | 88,120.3 | 3,402.3 |
| `rules/tailwind/collapse/utility_nodes.go:140` | 1,631.2 | 187.1 | 115.2 | 8.72x | 16,900.2 | 2,952.2 |
| `rules/tailwind/collapse/utility_nodes.go:142` | 17,614.9 | 513.2 | 163.6 | 34.33x | 183,143.9 | 6,983.6 |
| `rules/tailwind/no_deprecated_classes.go:65` | 5,996.1 | 6.1 | 75.8 | 986.57x | 64,010.4 | 109.9 |
| `rules/tailwind/no_deprecated_classes.go:66` | 6,021.6 | 6.2 | 68.1 | 972.89x | 64,333.6 | 110.9 |
| `rules/tailwind/no_deprecated_classes.go:67` | 7,197.8 | 6.4 | 67.5 | 1116.62x | 64,286.3 | 109.8 |
| `rules/tailwind/no_deprecated_classes.go:68` | 6,026.5 | 5.9 | 58.1 | 1013.44x | 64,130.8 | 106.4 |
| `rules/tailwind/no_deprecated_classes.go:69` | 5,865.6 | 6.7 | 70.2 | 873.75x | 64,357.8 | 111.4 |
| `rules/tailwind/no_deprecated_classes.go:70` | 6,070.3 | 6.6 | 60.5 | 920.22x | 64,208.1 | 108.1 |
| `rules/tailwind/no_deprecated_classes.go:74` | 6,010.1 | 176.7 | 97.8 | 34.02x | 64,413.4 | 2,777.4 |
| `rules/tailwind/no_deprecated_classes.go:75` | 6,240.3 | 173.0 | 75.0 | 36.08x | 64,446.2 | 2,589.4 |
| `rules/tailwind/no_deprecated_classes.go:76` | 5,756.6 | 171.8 | 64.5 | 33.50x | 64,516.3 | 2,619.0 |
| `rules/tailwind/no_deprecated_classes.go:77` | 6,243.4 | 165.2 | 89.9 | 37.80x | 64,495.0 | 2,609.7 |
| `rules/tailwind/no_deprecated_classes.go:78` | 6,051.8 | 227.5 | 64.3 | 26.61x | 64,467.2 | 2,598.6 |
| `rules/tailwind/no_deprecated_classes.go:79` | 6,105.4 | 169.7 | 68.6 | 35.97x | 64,622.1 | 2,663.7 |
| `rules/tailwind/no_deprecated_classes.go:83` | 6,104.5 | 6.1 | 62.8 | 996.10x | 64,286.7 | 109.8 |
| `rules/tailwind/no_deprecated_classes.go:84` | 5,704.2 | 163.3 | 63.1 | 34.94x | 64,419.7 | 2,577.9 |
| `rules/tailwind/no_deprecated_classes.go:85` | 5,872.3 | 6.3 | 67.9 | 927.80x | 64,236.2 | 108.7 |
| `rules/tailwind/no_deprecated_classes.go:86` | 5,591.5 | 157.1 | 71.4 | 35.58x | 64,369.2 | 2,556.6 |
| `rules/tailwind/no_deprecated_classes.go:88` | 6,246.7 | 6.4 | 64.7 | 977.37x | 64,460.0 | 113.7 |
| `rules/tailwind/no_deprecated_classes.go:89` | 6,138.0 | 7.0 | 60.9 | 873.24x | 64,412.4 | 112.6 |
| `rules/tailwind/no_deprecated_classes.go:90` | 5,716.3 | 6.6 | 78.5 | 866.34x | 64,412.4 | 112.6 |
| `rules/tailwind/no_deprecated_classes.go:93` | 5,446.9 | 6.5 | 75.9 | 840.86x | 64,350.0 | 111.4 |
| `rules/tailwind/no_deprecated_classes.go:94` | 5,535.4 | 6.3 | 64.0 | 878.46x | 64,425.7 | 113.1 |
| `rules/tailwind/no_deprecated_classes.go:95` | 5,796.9 | 6.6 | 68.4 | 883.64x | 64,375.6 | 112.0 |
| `rules/tailwind/no_deprecated_classes.go:96` | 5,662.7 | 6.6 | 125.0 | 853.34x | 64,451.2 | 113.7 |
| `rules/tailwind/no_deprecated_classes.go:97` | 6,109.8 | 6.8 | 64.8 | 901.96x | 64,409.2 | 112.6 |
| `rules/tailwind/no_deprecated_classes.go:98` | 5,578.2 | 6.4 | 62.5 | 870.45x | 64,484.9 | 114.2 |
| `rules/tailwind/no_deprecated_classes.go:99` | 5,651.7 | 6.3 | 70.1 | 891.98x | 64,434.4 | 113.1 |
| `rules/tailwind/no_deprecated_classes.go:100` | 5,884.6 | 6.4 | 94.0 | 912.71x | 64,510.1 | 114.8 |
| `rules/tailwind/no_physical_direction.go:284` | 8,376.3 | 922.9 | 121.5 | 9.08x | 94,419.7 | 11,196.7 |
| `rules/tailwind/tools/generate_variant/main.go:221` | 16,455.0 | 686.2 | 127.6 | 23.98x | 178,893.7 | 9,594.9 |
| `rules/typescript/ban_tslint_comment.go:24` | 6,583.2 | 218.0 | 73.0 | 30.19x | 65,898.4 | 3,068.2 |
| `rules/typescript/no_unnecessary_template_expression.go:574` | 7,787.6 | 7.4 | 63.4 | 1046.49x | 85,743.9 | 134.2 |
| `rules/typescript/switch_exhaustiveness_check.go:381` | 5,875.8 | 6.3 | 71.3 | 935.63x | 64,488.2 | 113.0 |
| `hard/keywords` | 15,085,759.0 | 1,651,644.0 | 30,776.0 | 9.13x | 168,692,179.0 | 19,211,865.0 |
| `hard/email` | 8,305,274.4 | 2,377,553.1 | 185,724.5 | 3.49x | 76,500,341.5 | 25,185,223.0 |
| `hard/ignorecase` | 1,179,690.6 | 28,009.2 | 3,347.5 | 42.12x | 11,992,999.5 | 392,643.5 |
| `hard/lookbehind` | 1,729,949.1 | 512,021.6 | 54,514.9 | 3.38x | 16,730,201.5 | 4,971,206.5 |
| `original/eslint-default-case` | 5,971.8 | 6.6 | 64.8 | 906.50x | 66,375.4 | 113.0 |
| `original/typescript-switch-exhaustiveness` | 5,846.7 | 6.1 | 71.5 | 956.43x | 66,375.4 | 113.0 |
| `original/typescript-ban-tslint-comment` | 5,354.1 | 207.4 | 68.3 | 25.81x | 65,927.2 | 3,074.2 |
| `original/react-nested-components` | 5,875.5 | 189.1 | 109.0 | 31.06x | 65,133.0 | 2,915.7 |
| `original/react-sort-comp` | 8,038.7 | 183.5 | 172.9 | 43.80x | 87,303.1 | 2,906.0 |
| `original/typescript-template-dollar` | 16,745.3 | 186.0 | 66.8 | 90.03x | 184,665.6 | 2,747.4 |
