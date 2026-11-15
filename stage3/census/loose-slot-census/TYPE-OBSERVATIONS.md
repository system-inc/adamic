# Type-level observations

22 events; 22 distinct exact source positions. The exhaustive production count remains UNKNOWN. Each source line is quoted in [SITES.md](SITES.md).

| Form | Events | Sites |
| --- | ---: | ---: |
| captured | 7 | 7 |
| literal init | 9 | 9 |
| write | 6 | 6 |

| File | Events | Sites |
| --- | ---: | ---: |
| checker.ts | 2 | 2 |
| commandLineParser.ts | 10 | 10 |
| executeCommandLine.ts | 2 | 2 |
| moduleNameResolver.ts | 1 | 1 |
| moduleSpecifiers.ts | 1 | 1 |
| sys.ts | 3 | 3 |
| tracing.ts | 1 | 1 |
| tsbuildPublic.ts | 1 | 1 |
| watchPublic.ts | 1 | 1 |

Whole-program run: 29.502959 s wall; 1125988 KiB peak RSS. Input: 79 roots; 799,863 non-type AST nodes; 33 candidates; H.every=true; 190 explicit H types; 4,457 shapes.

The private instrument cannot reconstruct all lowering-generated stores, carried slots, concrete generic instantiations or retained capture metadata. Its results are source/type judgments under the archived reconstruction, with neither an upper nor a lower bound on an exhaustive production count.
