| Phase / input | Pass | Fail | Native s | Node s | Go single s | Go default s | 1.78 s bar |
|---|---:|---:|---:|---:|---:|---:|---|
| correctness: acceptance | 301 | 0 | | | | | |
| correctness: tiny | 1 | 0 | | | | | |
| correctness: baselines | 13693 | 0 | | | | | |
| timing: 001_varianceCantBeStrictWhileStructureIsnt | | | 1.0344 ± 0.0471 | 1.0134 ± 0.0457 | 0.2850 ± 0.0186 | 0.2864 ± 0.0329 |  |
| timing: 024_genericTypeParameterEquivalence2 | | | 0.9986 ± 0.0355 | 1.0158 ± 0.0342 | 0.2780 ± 0.0128 | 0.2779 ± 0.0141 |  |
| timing: 056_genericCallInferenceInConditionalTypes1 | | | 1.0690 ± 0.1161 | 1.0117 ± 0.0582 | 0.2905 ± 0.0333 | 0.2742 ± 0.0278 |  |
| timing: mitt | | | 0.3827 ± 0.0185 | 0.3770 ± 0.0312 | 0.0545 ± 0.0059 | 0.0572 ± 0.0046 |  |
| timing: zod | | | 4.3152 ± 0.1434 | 4.2964 ± 0.1689 | 1.2729 ± 0.0579 | 0.9882 ± 0.0694 |  |
| timing: typescript-compiler | | | 9.1687 ± 0.4466 | 8.9089 ± 0.2447 | 2.9664 ± 0.1193 | 1.9559 ± 0.1169 | missed |
