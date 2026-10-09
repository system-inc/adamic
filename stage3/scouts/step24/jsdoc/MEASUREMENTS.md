# Node JSDoc measurements

Timing is instrumented attached-comment time divided by total parsing wall time, excluding IO, imports and traversal. Values are five-round medians; ranges show timing noise. All case node totals agree with the checked-in driver reference.

| Corpus | Files | Nodes | JSDoc roots | Descendants in JSDoc | Node share | Parse ms | Time share (range) |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| compiler | 82 | 968181 | 4157 | 13093 | 1.352% | 568.664 | 3.160% (3.007-4.025%) |
| parser | 800 | 63183 | 68 | 143 | 0.226% | 41.240 | 0.801% (0.621-0.993%) |
| jsx | 175 | 18848 | 0 | 0 | 0.000% | 11.426 | 0.000% (0.000-0.000%) |
| salsa | 108 | 7706 | 85 | 669 | 8.682% | 4.700 | 12.673% (6.929-14.104%) |
| remaining | 9323 | 1233374 | 1605 | 11722 | 0.950% | 665.925 | 1.897% (1.775-2.070%) |
| directed | 4 | 62 | 4 | 37 | 59.677% | 0.056 | 57.221% (38.300-71.788%) |
| all_cases | 10406 | 1323111 | 1758 | 12534 | 0.947% | 724.739 | 1.877% (1.788-2.027%) |
