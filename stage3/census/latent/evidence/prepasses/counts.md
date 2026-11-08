
## first-error

| View | Cleared | New | Changed | Newly stopped units | Changed existing first stops | Panics before | Panics after |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| raw_findings | 5 | 18 | 0 | 8 | 13 | 0 | 0 |
| distinct_reads | 5 | 18 | 0 | 8 | 13 | 0 | 0 |

New findings by exact reason and source file:

| Reason | File | Raw | Distinct |
| --- | --- | ---: | ---: |
| writing a possibly absent optional own field | src/compiler/builder.ts | 2 | 2 |
| writing a possibly absent optional own field | src/compiler/checker.ts | 1 | 1 |
| writing a possibly absent optional own field | src/compiler/executeCommandLine.ts | 3 | 3 |
| writing a possibly absent optional own field | src/compiler/factory/emitNode.ts | 4 | 4 |
| writing a possibly absent optional own field | src/compiler/factory/nodeFactory.ts | 1 | 1 |
| writing a possibly absent optional own field | src/compiler/moduleNameResolver.ts | 2 | 2 |
| writing a possibly absent optional own field | src/compiler/parser.ts | 1 | 1 |
| writing a possibly absent optional own field | src/compiler/utilities.ts | 2 | 2 |
| writing a possibly absent optional own field | src/compiler/watchPublic.ts | 1 | 1 |
| writing a possibly absent optional own field | src/compiler/watchUtilities.ts | 1 | 1 |

Cleared is an old finding with no remaining reason at that identity. Changed retains another reason at the identity. New includes replacement reasons. A cleared first-error observation can be hidden by an earlier new stop. Only lowering-phase findings define unit first stops. Raw and distinct are unique identities, not event multiplicity.

## core-full

| View | Cleared | New | Changed | Newly stopped units | Changed existing first stops | Panics before | Panics after |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| raw_findings | 0 | 0 | 0 | 0 | 0 | 0 | 0 |
| distinct_reads | 0 | 0 | 0 | 0 | 0 | 0 | 0 |

New findings by exact reason and source file:

| Reason | File | Raw | Distinct |
| --- | --- | ---: | ---: |
| None | | 0 | 0 |

Cleared is an old finding with no remaining reason at that identity. Changed retains another reason at the identity. New includes replacement reasons. A cleared first-error observation can be hidden by an earlier new stop. Only lowering-phase findings define unit first stops. Raw and distinct are unique identities, not event multiplicity.
