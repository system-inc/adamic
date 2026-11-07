# assign ordinary compiler census

Unmodified Adamic compiler sources on the frozen main plus getownvalues stack.
The only Go overlay changes the census main enumeration into bounded root groups.
Its ordinary Load/check/Lower inspect function is unchanged. Each row measures one
root independently, with all checking enabled. No source fallback or refusal
suppression is used. All 79 roots are measured for baseline and each spelling.

Baseline groups are 20, 20, 20 and 19 sorted compiler roots.

| Baseline roots | Kind | First stop | Exact reason |
| ---: | --- | --- | --- |
| 73 | NotYet | `core.ts:1364:29` | a non-object representation passed through an opaque object parameter |
| 1 | NotYet | `debug.ts:243:59` | a generic instantiation whose type argument for T, constrained to Node, can't be read |
| 2 | accepted | `-` | - |
| 1 | NotYet | `semver.ts:71:9` | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |
| 1 | NotYet | `commandLineParser.ts:293:35` | reading an enum before its runtime initialization; move the call after the enum declaration |
| 1 | NotYet | `diagnosticInformationMap.generated.ts:13:45` | reading an enum before its runtime initialization; move the call after the enum declaration |

## maplike-unknown

Spelling: `assign<T extends MapLike<unknown>>`.

Types over all 79 roots: exit 1, 1 diagnostics. Core JavaScript byte-identical with comments kept: True; removed: True. All caller source bytes unchanged.

| Former 73 roots | Kind | First stop | Exact reason |
| ---: | --- | --- | --- |
| 73 | checker | `commandLineParser.ts:2712:36` | error TS2345: Argument of type 'object' is not assignable to parameter of type 'MapLike&lt;unknown&gt;'.<br>  Index signature for type 'string' is missing in type '{}'. |

Group 1 through 20 (78.894 seconds):

| Roots | Kind | First stop | Exact reason |
| ---: | --- | --- | --- |
| 19 | checker | `commandLineParser.ts:2712:36` | error TS2345: Argument of type 'object' is not assignable to parameter of type 'MapLike&lt;unknown&gt;'.<br>  Index signature for type 'string' is missing in type '{}'. |
| 1 | accepted | `-` | - |

Group 21 through 40 (86.772 seconds):

| Roots | Kind | First stop | Exact reason |
| ---: | --- | --- | --- |
| 19 | checker | `commandLineParser.ts:2712:36` | error TS2345: Argument of type 'object' is not assignable to parameter of type 'MapLike&lt;unknown&gt;'.<br>  Index signature for type 'string' is missing in type '{}'. |
| 1 | accepted | `-` | - |

Group 41 through 60 (102.264 seconds):

| Roots | Kind | First stop | Exact reason |
| ---: | --- | --- | --- |
| 20 | checker | `commandLineParser.ts:2712:36` | error TS2345: Argument of type 'object' is not assignable to parameter of type 'MapLike&lt;unknown&gt;'.<br>  Index signature for type 'string' is missing in type '{}'. |

Group 61 through 79 (108.455 seconds):

| Roots | Kind | First stop | Exact reason |
| ---: | --- | --- | --- |
| 19 | checker | `commandLineParser.ts:2712:36` | error TS2345: Argument of type 'object' is not assignable to parameter of type 'MapLike&lt;unknown&gt;'.<br>  Index signature for type 'string' is missing in type '{}'. |

## self-mapped

Spelling: `assign<T extends object & { [K in keyof T]: unknown }>`.

Types over all 79 roots: exit 0, 0 diagnostics. Core JavaScript byte-identical with comments kept: True; removed: True. All caller source bytes unchanged.

| Former 73 roots | Kind | First stop | Exact reason |
| ---: | --- | --- | --- |
| 73 | NotYet | `core.ts:1364:29` | a non-object representation passed through an opaque object parameter |

Group 1 through 20 (339.418 seconds):

| Roots | Kind | First stop | Exact reason |
| ---: | --- | --- | --- |
| 16 | NotYet | `core.ts:1364:29` | a non-object representation passed through an opaque object parameter |
| 1 | NotYet | `debug.ts:243:59` | a generic instantiation whose type argument for T, constrained to Node, can't be read |
| 1 | accepted | `-` | - |
| 1 | NotYet | `semver.ts:71:9` | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |
| 1 | NotYet | `commandLineParser.ts:293:35` | reading an enum before its runtime initialization; move the call after the enum declaration |

Group 21 through 40 (370.589 seconds):

| Roots | Kind | First stop | Exact reason |
| ---: | --- | --- | --- |
| 19 | NotYet | `core.ts:1364:29` | a non-object representation passed through an opaque object parameter |
| 1 | accepted | `-` | - |

Group 41 through 60 (394.889 seconds):

| Roots | Kind | First stop | Exact reason |
| ---: | --- | --- | --- |
| 20 | NotYet | `core.ts:1364:29` | a non-object representation passed through an opaque object parameter |

Group 61 through 79 (366.530 seconds):

| Roots | Kind | First stop | Exact reason |
| ---: | --- | --- | --- |
| 18 | NotYet | `core.ts:1364:29` | a non-object representation passed through an opaque object parameter |
| 1 | NotYet | `diagnosticInformationMap.generated.ts:13:45` | reading an enum before its runtime initialization; move the call after the enum declaration |

## checked-record-local

Spelling: `const record: MapLike<unknown> = arg; hasProperty(record, p)`.

Types over all 79 roots: exit 1, 1 diagnostics. Core JavaScript byte-identical with comments kept: False; removed: False. All caller source bytes unchanged.

| Former 73 roots | Kind | First stop | Exact reason |
| ---: | --- | --- | --- |
| 73 | checker | `core.ts:1363:15` | error TS2322: Type 'T' is not assignable to type 'MapLike&lt;unknown&gt;'.<br>  Type 'object' is not assignable to type 'MapLike&lt;unknown&gt;'.<br>    Index signature for type 'string' is missing in type '{}'. |

Group 1 through 20 (84.613 seconds):

| Roots | Kind | First stop | Exact reason |
| ---: | --- | --- | --- |
| 19 | checker | `core.ts:1363:15` | error TS2322: Type 'T' is not assignable to type 'MapLike&lt;unknown&gt;'.<br>  Type 'object' is not assignable to type 'MapLike&lt;unknown&gt;'.<br>    Index signature for type 'string' is missing in type '{}'. |
| 1 | accepted | `-` | - |

Group 21 through 40 (84.208 seconds):

| Roots | Kind | First stop | Exact reason |
| ---: | --- | --- | --- |
| 19 | checker | `core.ts:1363:15` | error TS2322: Type 'T' is not assignable to type 'MapLike&lt;unknown&gt;'.<br>  Type 'object' is not assignable to type 'MapLike&lt;unknown&gt;'.<br>    Index signature for type 'string' is missing in type '{}'. |
| 1 | accepted | `-` | - |

Group 41 through 60 (95.589 seconds):

| Roots | Kind | First stop | Exact reason |
| ---: | --- | --- | --- |
| 20 | checker | `core.ts:1363:15` | error TS2322: Type 'T' is not assignable to type 'MapLike&lt;unknown&gt;'.<br>  Type 'object' is not assignable to type 'MapLike&lt;unknown&gt;'.<br>    Index signature for type 'string' is missing in type '{}'. |

Group 61 through 79 (90.894 seconds):

| Roots | Kind | First stop | Exact reason |
| ---: | --- | --- | --- |
| 19 | checker | `core.ts:1363:15` | error TS2322: Type 'T' is not assignable to type 'MapLike&lt;unknown&gt;'.<br>  Type 'object' is not assignable to type 'MapLike&lt;unknown&gt;'.<br>    Index signature for type 'string' is missing in type '{}'. |

## generic-own-helper

Spelling: `hasProperty<T extends object>(map: T, key: string)`.

Types over all 79 roots: exit 0, 0 diagnostics. Core JavaScript byte-identical with comments kept: True; removed: True. All caller source bytes unchanged.

| Former 73 roots | Kind | First stop | Exact reason |
| ---: | --- | --- | --- |
| 73 | Refused | `core.ts:1484:12` | a cast the runtime can't check |

Group 1 through 20 (403.908 seconds):

| Roots | Kind | First stop | Exact reason |
| ---: | --- | --- | --- |
| 16 | Refused | `core.ts:1484:12` | a cast the runtime can't check |
| 1 | NotYet | `debug.ts:243:59` | a generic instantiation whose type argument for T, constrained to Node, can't be read |
| 1 | accepted | `-` | - |
| 1 | NotYet | `semver.ts:71:9` | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration |
| 1 | NotYet | `commandLineParser.ts:293:35` | reading an enum before its runtime initialization; move the call after the enum declaration |

Group 21 through 40 (411.292 seconds):

| Roots | Kind | First stop | Exact reason |
| ---: | --- | --- | --- |
| 19 | Refused | `core.ts:1484:12` | a cast the runtime can't check |
| 1 | accepted | `-` | - |

Group 41 through 60 (460.801 seconds):

| Roots | Kind | First stop | Exact reason |
| ---: | --- | --- | --- |
| 20 | Refused | `core.ts:1484:12` | a cast the runtime can't check |

Group 61 through 79 (424.627 seconds):

| Roots | Kind | First stop | Exact reason |
| ---: | --- | --- | --- |
| 18 | Refused | `core.ts:1484:12` | a cast the runtime can't check |
| 1 | NotYet | `diagnosticInformationMap.generated.ts:13:45` | reading an enum before its runtime initialization; move the call after the enum declaration |

CENSUS.json and TRIALS.json retain every root and exact diagnostics.
Raw grouped output and type-command logs are in the closeout archive.
A checker result stops before lowering; no later lowering stop is claimed for
a checker-rejected spelling.
