# K1T2WP9: 4ac3f04d → 8502a46e

One foreground meter invocation at 8502a46e, compared with the committed 4ac3f04d run. Main only is counted; area repeats the same pin. First-error findings are floors, and latent lowering on a checker-rejected program does not establish native output.

## Headline

| SHA | Whole program | Own file | Lowering | tsc entry diagnostics | NotYet findings | Refused findings | Panic findings | Error findings |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 4ac3f04d | 2/79 | 58/79 | 2/79 | 144 | 1122 | 3083 | 1 | 0 |
| 8502a46e | 2/79 | 58/79 | 2/79 | 144 | 929 | 3083 | 1 | 0 |

Gate columns come from census; finding columns count unique (kind, where, reason, text), as in COMPARISON.md.

## Named reasons

| Reason | Before distinct sites | After distinct sites | Difference |
| --- | --- | --- | --- |
| checked predicate membership by an enum tag another member can hold | 214 | 0 | -214 |
| checked predicate membership over an open tag domain | 45 | 45 | 0 |

## Changed lowering reasons

| Reason | Before | After | Difference |
| --- | --- | --- | --- |
| checked predicate membership by an enum tag another member can hold | 214 | 0 | -214 |
| a checked field alias requiring an optional, accessor, or representation conversion | 29 | 49 | 20 |
| writing a checked object field without its source-slot type certificate | 2 | 3 | 1 |

Counts use distinct normalized file:line:column sites, not finding events.

## Split of removed enum-tag sites

214 removed sites: 0 reach no lowering finding in their owning units; 214 move to a later finding. 0 newly appearing sites for the original reason. A no-finding unit is not proof of native output. Later findings are matched by normalized owning unit, using the lowering phase; refusal-scan findings are not substituted for unit stops. A site with multiple owning units is classified as moved if any has a later finding. Reason totals can overlap if such units differ. Multiple removed sites can reach the same later callee site, so these attribution counts differ from the distinct-site reason changes above; the getter reason has no net distinct-site change.

| Later reason | Removed sites reaching it |
| --- | --- |
| a checked field alias requiring an optional, accessor, or representation conversion | 173 |
| a getter in a checked field contract | 32 |
| writing a checked object field without its source-slot type certificate | 9 |

<details><summary>Every removed site and its owning units</summary>

| Before site | Owning unit(s) | After |
| --- | --- | --- |
| src/compiler/binder.ts:3841:17 | src/compiler/binder.ts:3815:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1026:23 |
| src/compiler/binder.ts:429:13 | src/compiler/binder.ts:422:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:11545:44 |
| src/compiler/checker.ts:54204:13 | src/compiler/checker.ts:54202:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1026:23 |
| src/compiler/checker.ts:54217:20 | src/compiler/checker.ts:54213:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1582:34 |
| src/compiler/emitter.ts:523:28 | src/compiler/emitter.ts:516:1 | NotYet: writing a checked object field without its source-slot type certificate at src/compiler/parser.ts:1622:13 |
| src/compiler/factory/emitHelpers.ts:1487:12 | src/compiler/factory/emitHelpers.ts:1486:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:2876:37 |
| src/compiler/factory/utilities.ts:1010:36 | src/compiler/factory/utilities.ts:1008:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/factory/utilities.ts:466:9 |
| src/compiler/factory/utilities.ts:1152:12 | src/compiler/factory/utilities.ts:1151:1 | NotYet: a getter in a checked field contract at src/compiler/sourcemap.ts:474:9 |
| src/compiler/factory/utilities.ts:1156:12 | src/compiler/factory/utilities.ts:1155:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1582:34 |
| src/compiler/factory/utilities.ts:1160:12 | src/compiler/factory/utilities.ts:1159:1 | NotYet: a getter in a checked field contract at src/compiler/sourcemap.ts:474:9 |
| src/compiler/factory/utilities.ts:1164:12 | src/compiler/factory/utilities.ts:1163:1 | NotYet: a getter in a checked field contract at src/compiler/sourcemap.ts:474:9 |
| src/compiler/factory/utilities.ts:1168:12 | src/compiler/factory/utilities.ts:1167:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1582:34 |
| src/compiler/factory/utilities.ts:1571:12 | src/compiler/factory/utilities.ts:1570:1 | NotYet: a getter in a checked field contract at src/compiler/sourcemap.ts:474:9 |
| src/compiler/factory/utilities.ts:1696:12 | src/compiler/factory/utilities.ts:1695:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:4662:66 |
| src/compiler/factory/utilities.ts:1702:9 | src/compiler/factory/utilities.ts:1701:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:4662:66 |
| src/compiler/factory/utilities.ts:580:12 | src/compiler/factory/utilities.ts:579:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:1994:17 |
| src/compiler/factory/utilities.ts:586:13 | src/compiler/factory/utilities.ts:584:1 | NotYet: a getter in a checked field contract at src/compiler/sourcemap.ts:474:9 |
| src/compiler/factory/utilities.ts:602:12 | src/compiler/factory/utilities.ts:599:1 | NotYet: a getter in a checked field contract at src/compiler/sourcemap.ts:474:9 |
| src/compiler/factory/utilities.ts:618:12 | src/compiler/factory/utilities.ts:617:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:4662:66 |
| src/compiler/factory/utilities.ts:634:75 | src/compiler/factory/utilities.ts:631:1 | NotYet: a getter in a checked field contract at src/compiler/sourcemap.ts:474:9 |
| src/compiler/factory/utilities.ts:670:12 | src/compiler/factory/utilities.ts:668:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:2565:14 |
| src/compiler/factory/utilities.ts:793:16 | src/compiler/factory/utilities.ts:786:1 | NotYet: a getter in a checked field contract at src/compiler/sourcemap.ts:474:9 |
| src/compiler/factory/utilities.ts:816:23 | src/compiler/factory/utilities.ts:814:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:1994:17 |
| src/compiler/parser.ts:477:12 | src/compiler/parser.ts:476:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/parser.ts:500:12 | src/compiler/parser.ts:499:1 | NotYet: writing a checked object field without its source-slot type certificate at src/compiler/binder.ts:1109:17 |
| src/compiler/program.ts:937:9 | src/compiler/program.ts:936:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/program.ts:965:29 | src/compiler/program.ts:958:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/transformers/classFields.ts:3344:13 | src/compiler/transformers/classFields.ts:3343:1 | NotYet: a getter in a checked field contract at src/compiler/sourcemap.ts:474:9 |
| src/compiler/transformers/classFields.ts:3350:12 | src/compiler/transformers/classFields.ts:3349:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1587:41 |
| src/compiler/transformers/classFields.ts:3355:12 | src/compiler/transformers/classFields.ts:3354:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1026:23 |
| src/compiler/transformers/classFields.ts:3359:12 | src/compiler/transformers/classFields.ts:3358:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:1098:65 |
| src/compiler/transformers/classThis.ts:73:10 | src/compiler/transformers/classThis.ts:72:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:1098:65 |
| src/compiler/transformers/declarations/diagnostics.ts:139:12 | src/compiler/transformers/declarations/diagnostics.ts:138:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:2565:14 |
| src/compiler/transformers/destructuring.ts:220:25 | src/compiler/transformers/destructuring.ts:218:1 | NotYet: a getter in a checked field contract at src/compiler/sourcemap.ts:474:9 |
| src/compiler/transformers/destructuring.ts:584:9 | src/compiler/transformers/destructuring.ts:583:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1582:34 |
| src/compiler/transformers/esnext.ts:775:12 | src/compiler/transformers/esnext.ts:774:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/checker.ts:6068:28 |
| src/compiler/transformers/esnext.ts:789:12 | src/compiler/transformers/esnext.ts:788:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/transformers/module/impliedNodeFormatDependent.ts:45:13 | src/compiler/transformers/module/impliedNodeFormatDependent.ts:18:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:875:19 |
| src/compiler/transformers/namedEvaluation.ts:140:10 | src/compiler/transformers/namedEvaluation.ts:139:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:1098:65 |
| src/compiler/transformers/utilities.ts:175:103 | src/compiler/transformers/utilities.ts:173:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/checker.ts:53597:17 |
| src/compiler/transformers/utilities.ts:422:13 | src/compiler/transformers/utilities.ts:389:1, src/compiler/transformers/utilities.ts:441:1 | NotYet: a getter in a checked field contract at src/compiler/sourcemap.ts:474:9 |
| src/compiler/transformers/utilities.ts:472:12 | src/compiler/transformers/utilities.ts:471:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:6145:16 |
| src/compiler/transformers/utilities.ts:486:13 | src/compiler/transformers/utilities.ts:485:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1582:34 |
| src/compiler/transformers/utilities.ts:537:10 | src/compiler/transformers/utilities.ts:536:1 | NotYet: writing a checked object field without its source-slot type certificate at src/compiler/binder.ts:1109:17 |
| src/compiler/transformers/utilities.ts:594:52 | src/compiler/transformers/utilities.ts:593:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:1098:65 |
| src/compiler/transformers/utilities.ts:613:12 | src/compiler/transformers/utilities.ts:612:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1026:23 |
| src/compiler/transformers/utilities.ts:619:12 | src/compiler/transformers/utilities.ts:618:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1026:23 |
| src/compiler/transformers/utilities.ts:643:34 | src/compiler/transformers/utilities.ts:642:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1026:23 |
| src/compiler/transformers/utilities.ts:867:33 | src/compiler/transformers/utilities.ts:866:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1582:34 |
| src/compiler/transformers/utilities.ts:881:19 | src/compiler/transformers/utilities.ts:876:1, src/compiler/transformers/utilities.ts:878:1, src/compiler/transformers/utilities.ts:880:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:1994:17 |
| src/compiler/utilities.ts:10592:12 | src/compiler/utilities.ts:10591:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1582:34 |
| src/compiler/utilities.ts:10628:12 | src/compiler/utilities.ts:10627:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1428:51 |
| src/compiler/utilities.ts:10752:13 | src/compiler/utilities.ts:10751:1 | NotYet: a getter in a checked field contract at src/compiler/sourcemap.ts:474:9 |
| src/compiler/utilities.ts:10761:12 | src/compiler/utilities.ts:10760:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/factory/nodeFactory.ts:2864:61 |
| src/compiler/utilities.ts:11052:12 | src/compiler/utilities.ts:11051:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/utilities.ts:11115:32 | src/compiler/utilities.ts:11114:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:4662:66 |
| src/compiler/utilities.ts:11185:30 | src/compiler/utilities.ts:11184:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:4662:66 |
| src/compiler/utilities.ts:1143:12 | src/compiler/utilities.ts:1142:1 | NotYet: a getter in a checked field contract at src/compiler/sourcemap.ts:474:9 |
| src/compiler/utilities.ts:11453:13 | src/compiler/utilities.ts:11452:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:2565:14 |
| src/compiler/utilities.ts:12213:59 | src/compiler/utilities.ts:12204:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/utilities.ts:12222:12 | src/compiler/utilities.ts:12221:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/utilities.ts:12435:12 | src/compiler/utilities.ts:12434:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:11545:44 |
| src/compiler/utilities.ts:12441:13 | src/compiler/utilities.ts:12439:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/utilities.ts:1250:25 | src/compiler/utilities.ts:1238:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/utilities.ts:1277:51 | src/compiler/utilities.ts:1276:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/utilities.ts:1296:15 | src/compiler/utilities.ts:1295:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/utilities.ts:2079:12 | src/compiler/utilities.ts:2078:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/utilities.ts:2084:12 | src/compiler/utilities.ts:2083:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/utilities.ts:2089:12 | src/compiler/utilities.ts:2088:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/utilities.ts:2120:9 | src/compiler/utilities.ts:2117:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/utilities.ts:2130:12 | src/compiler/utilities.ts:2129:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/checker.ts:48243:38 |
| src/compiler/utilities.ts:2142:20 | src/compiler/utilities.ts:2134:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/checker.ts:48243:38 |
| src/compiler/utilities.ts:2225:21 | src/compiler/utilities.ts:2202:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/utilities.ts:2288:12 | src/compiler/utilities.ts:2287:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/utilities.ts:2293:12 | src/compiler/utilities.ts:2292:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/utilities.ts:2316:12 | src/compiler/utilities.ts:2315:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/utilities.ts:2321:12 | src/compiler/utilities.ts:2320:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/utilities.ts:2360:62 | src/compiler/utilities.ts:2359:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:6145:16 |
| src/compiler/utilities.ts:2375:17 | src/compiler/utilities.ts:2364:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:6145:16 |
| src/compiler/utilities.ts:2718:12 | src/compiler/utilities.ts:2717:1 | NotYet: writing a checked object field without its source-slot type certificate at src/compiler/binder.ts:1109:17 |
| src/compiler/utilities.ts:2725:12 | src/compiler/utilities.ts:2724:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:11193:40 |
| src/compiler/utilities.ts:2742:12 | src/compiler/utilities.ts:2740:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/utilities.ts:2746:12 | src/compiler/utilities.ts:2745:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1582:34 |
| src/compiler/utilities.ts:2753:12 | src/compiler/utilities.ts:2751:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/utilities.ts:2885:12 | src/compiler/utilities.ts:2884:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/factory/nodeFactory.ts:5455:20 |
| src/compiler/utilities.ts:3022:13 | src/compiler/utilities.ts:3020:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:3096:20 |
| src/compiler/utilities.ts:3029:13 | src/compiler/utilities.ts:3027:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:4662:66 |
| src/compiler/utilities.ts:3034:13 | src/compiler/utilities.ts:3033:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:2565:14 |
| src/compiler/utilities.ts:3069:54 | src/compiler/utilities.ts:3068:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/utilities.ts:3144:13 | src/compiler/utilities.ts:3142:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1026:23 |
| src/compiler/utilities.ts:3158:54 | src/compiler/utilities.ts:3157:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1026:23 |
| src/compiler/utilities.ts:3285:9 | src/compiler/utilities.ts:3283:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1582:34 |
| src/compiler/utilities.ts:3421:22 | src/compiler/utilities.ts:3420:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:2565:14 |
| src/compiler/utilities.ts:3427:13 | src/compiler/utilities.ts:3425:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:1102:65 |
| src/compiler/utilities.ts:3671:21 | src/compiler/utilities.ts:3631:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/checker.ts:4747:17 |
| src/compiler/utilities.ts:3769:12 | src/compiler/utilities.ts:3768:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/checker.ts:48804:32 |
| src/compiler/utilities.ts:3785:12 | src/compiler/utilities.ts:3784:1 | NotYet: a getter in a checked field contract at src/compiler/sourcemap.ts:474:9 |
| src/compiler/utilities.ts:3815:13 | src/compiler/utilities.ts:3814:1 | NotYet: writing a checked object field without its source-slot type certificate at src/compiler/parser.ts:1622:13 |
| src/compiler/utilities.ts:3825:12 | src/compiler/utilities.ts:3824:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1428:51 |
| src/compiler/utilities.ts:3881:12 | src/compiler/utilities.ts:3880:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:2574:14 |
| src/compiler/utilities.ts:3886:9 | src/compiler/utilities.ts:3885:1 | NotYet: a getter in a checked field contract at src/compiler/sourcemap.ts:474:9 |
| src/compiler/utilities.ts:3899:12 | src/compiler/utilities.ts:3898:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:2565:14 |
| src/compiler/utilities.ts:3906:12 | src/compiler/utilities.ts:3905:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/utilities.ts:3923:12 | src/compiler/utilities.ts:3922:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:4662:66 |
| src/compiler/utilities.ts:3934:9 | src/compiler/utilities.ts:3931:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:4662:66 |
| src/compiler/utilities.ts:3969:32 | src/compiler/utilities.ts:3968:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:4662:66 |
| src/compiler/utilities.ts:3995:9 | src/compiler/utilities.ts:3994:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:2876:37 |
| src/compiler/utilities.ts:4020:15 | src/compiler/utilities.ts:4019:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:4662:66 |
| src/compiler/utilities.ts:4030:18 | src/compiler/utilities.ts:4029:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:2565:14 |
| src/compiler/utilities.ts:4042:9 | src/compiler/utilities.ts:4041:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:4662:66 |
| src/compiler/utilities.ts:4065:9 | src/compiler/utilities.ts:4064:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/parser.ts:3774:13 |
| src/compiler/utilities.ts:4087:12 | src/compiler/utilities.ts:4086:1 | NotYet: a getter in a checked field contract at src/compiler/sourcemap.ts:474:9 |
| src/compiler/utilities.ts:4095:12 | src/compiler/utilities.ts:4094:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1582:34 |
| src/compiler/utilities.ts:4100:12 | src/compiler/utilities.ts:4099:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1582:34 |
| src/compiler/utilities.ts:4105:13 | src/compiler/utilities.ts:4104:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:4662:66 |
| src/compiler/utilities.ts:4133:12 | src/compiler/utilities.ts:4132:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:4662:66 |
| src/compiler/utilities.ts:4140:12 | src/compiler/utilities.ts:4139:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:4662:66 |
| src/compiler/utilities.ts:4149:12 | src/compiler/utilities.ts:4148:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:4662:66 |
| src/compiler/utilities.ts:4167:12 | src/compiler/utilities.ts:4166:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1582:34 |
| src/compiler/utilities.ts:4203:12 | src/compiler/utilities.ts:4202:1 | NotYet: a getter in a checked field contract at src/compiler/sourcemap.ts:474:9 |
| src/compiler/utilities.ts:4281:12 | src/compiler/utilities.ts:4280:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:4662:66 |
| src/compiler/utilities.ts:4295:12 | src/compiler/utilities.ts:4294:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:4662:66 |
| src/compiler/utilities.ts:4326:60 | src/compiler/utilities.ts:4321:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:2565:14 |
| src/compiler/utilities.ts:4476:19 | src/compiler/utilities.ts:4475:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1252:20 |
| src/compiler/utilities.ts:4488:12 | src/compiler/utilities.ts:4487:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:11545:44 |
| src/compiler/utilities.ts:4492:12 | src/compiler/utilities.ts:4491:1 | NotYet: writing a checked object field without its source-slot type certificate at src/compiler/binder.ts:1109:17 |
| src/compiler/utilities.ts:4500:12 | src/compiler/utilities.ts:4499:1 | NotYet: writing a checked object field without its source-slot type certificate at src/compiler/binder.ts:1109:17 |
| src/compiler/utilities.ts:4523:12 | src/compiler/utilities.ts:4522:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/utilities.ts:4527:12 | src/compiler/utilities.ts:4526:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/utilities.ts:4661:9 | src/compiler/utilities.ts:4654:1, src/compiler/utilities.ts:4657:1, src/compiler/utilities.ts:4658:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:2574:14 |
| src/compiler/utilities.ts:4702:14 | src/compiler/utilities.ts:4701:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1230:19 |
| src/compiler/utilities.ts:4719:9 | src/compiler/utilities.ts:4710:1 | NotYet: a getter in a checked field contract at src/compiler/sourcemap.ts:474:9 |
| src/compiler/utilities.ts:4768:9 | src/compiler/utilities.ts:4767:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1264:52 |
| src/compiler/utilities.ts:4788:16 | src/compiler/utilities.ts:4785:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/utilities.ts:4937:24 | src/compiler/utilities.ts:4935:1 | NotYet: a getter in a checked field contract at src/compiler/sourcemap.ts:474:9 |
| src/compiler/utilities.ts:4998:12 | src/compiler/utilities.ts:4997:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/utilities.ts:5036:12 | src/compiler/utilities.ts:5035:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:2565:14 |
| src/compiler/utilities.ts:5074:13 | src/compiler/utilities.ts:5073:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:875:19 |
| src/compiler/utilities.ts:5085:17 | src/compiler/utilities.ts:5079:1 | NotYet: a getter in a checked field contract at src/compiler/sourcemap.ts:474:9 |
| src/compiler/utilities.ts:5113:12 | src/compiler/utilities.ts:5112:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:6145:16 |
| src/compiler/utilities.ts:5174:12 | src/compiler/utilities.ts:5173:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1582:34 |
| src/compiler/utilities.ts:5230:12 | src/compiler/utilities.ts:5229:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/utilities.ts:5363:12 | src/compiler/utilities.ts:5362:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:6145:16 |
| src/compiler/utilities.ts:5368:12 | src/compiler/utilities.ts:5367:1 | NotYet: a getter in a checked field contract at src/compiler/sourcemap.ts:474:9 |
| src/compiler/utilities.ts:5409:17 | src/compiler/utilities.ts:5397:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:6145:16 |
| src/compiler/utilities.ts:5565:10 | src/compiler/utilities.ts:5564:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/utilities.ts:6517:22 | src/compiler/utilities.ts:6509:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:6145:16 |
| src/compiler/utilities.ts:6819:12 | src/compiler/utilities.ts:6814:1 | NotYet: writing a checked object field without its source-slot type certificate at src/compiler/binder.ts:1109:17 |
| src/compiler/utilities.ts:6838:9 | src/compiler/utilities.ts:6832:1 | NotYet: a getter in a checked field contract at src/compiler/sourcemap.ts:474:9 |
| src/compiler/utilities.ts:6893:30 | src/compiler/utilities.ts:6892:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/utilities.ts:6924:12 | src/compiler/utilities.ts:6923:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1332:9 |
| src/compiler/utilities.ts:7298:27 | src/compiler/utilities.ts:7296:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1026:23 |
| src/compiler/utilities.ts:7345:17 | src/compiler/utilities.ts:7344:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/utilities.ts:7421:12 | src/compiler/utilities.ts:7420:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:4662:66 |
| src/compiler/utilities.ts:7431:12 | src/compiler/utilities.ts:7430:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:4662:66 |
| src/compiler/utilities.ts:7456:9 | src/compiler/utilities.ts:7455:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/factory/nodeFactory.ts:3713:20 |
| src/compiler/utilities.ts:7476:12 | src/compiler/utilities.ts:7471:1, src/compiler/utilities.ts:7473:1, src/compiler/utilities.ts:7475:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:4662:66 |
| src/compiler/utilities.ts:7485:9 | src/compiler/utilities.ts:7484:1 | NotYet: a getter in a checked field contract at src/compiler/sourcemap.ts:474:9 |
| src/compiler/utilities.ts:7501:51 | src/compiler/utilities.ts:7500:1 | NotYet: a getter in a checked field contract at src/compiler/sourcemap.ts:474:9 |
| src/compiler/utilities.ts:7536:12 | src/compiler/utilities.ts:7535:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:4662:66 |
| src/compiler/utilities.ts:7564:12 | src/compiler/utilities.ts:7563:1 | NotYet: a getter in a checked field contract at src/compiler/sourcemap.ts:474:9 |
| src/compiler/utilities.ts:7576:30 | src/compiler/utilities.ts:7575:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:4662:66 |
| src/compiler/utilities.ts:7582:12 | src/compiler/utilities.ts:7581:1 | NotYet: writing a checked object field without its source-slot type certificate at src/compiler/binder.ts:1109:17 |
| src/compiler/utilities.ts:7588:12 | src/compiler/utilities.ts:7587:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:4662:66 |
| src/compiler/utilities.ts:7593:12 | src/compiler/utilities.ts:7592:1 | NotYet: a getter in a checked field contract at src/compiler/sourcemap.ts:474:9 |
| src/compiler/utilities.ts:7880:27 | src/compiler/utilities.ts:7879:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/utilities.ts:7892:9 | src/compiler/utilities.ts:7891:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1026:23 |
| src/compiler/utilities.ts:8025:12 | src/compiler/utilities.ts:8024:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:2565:14 |
| src/compiler/utilities.ts:8286:77 | src/compiler/utilities.ts:8285:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:1107:64 |
| src/compiler/utilities.ts:8291:12 | src/compiler/utilities.ts:8290:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:1994:17 |
| src/compiler/utilities.ts:8326:12 | src/compiler/utilities.ts:8325:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1026:23 |
| src/compiler/utilities.ts:8375:12 | src/compiler/utilities.ts:8374:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:4662:66 |
| src/compiler/utilities.ts:8963:12 | src/compiler/utilities.ts:8961:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:4838:14 |
| src/compiler/utilitiesPublic.ts:1000:10 | src/compiler/utilitiesPublic.ts:997:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/utilitiesPublic.ts:1008:14 | src/compiler/utilitiesPublic.ts:1004:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:2574:14 |
| src/compiler/utilitiesPublic.ts:1078:95 | src/compiler/utilitiesPublic.ts:1076:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1332:9 |
| src/compiler/utilitiesPublic.ts:1247:13 | src/compiler/utilitiesPublic.ts:1239:1 | NotYet: a getter in a checked field contract at src/compiler/sourcemap.ts:474:9 |
| src/compiler/utilitiesPublic.ts:1258:10 | src/compiler/utilitiesPublic.ts:1257:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/utilitiesPublic.ts:1353:9 | src/compiler/utilitiesPublic.ts:1351:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1332:9 |
| src/compiler/utilitiesPublic.ts:1369:12 | src/compiler/utilitiesPublic.ts:1368:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:4662:66 |
| src/compiler/utilitiesPublic.ts:1373:12 | src/compiler/utilitiesPublic.ts:1372:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:4662:66 |
| src/compiler/utilitiesPublic.ts:1377:12 | src/compiler/utilitiesPublic.ts:1376:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:2876:37 |
| src/compiler/utilitiesPublic.ts:1391:12 | src/compiler/utilitiesPublic.ts:1390:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1391:69 |
| src/compiler/utilitiesPublic.ts:1400:12 | src/compiler/utilitiesPublic.ts:1399:1 | NotYet: a getter in a checked field contract at src/compiler/sourcemap.ts:474:9 |
| src/compiler/utilitiesPublic.ts:1417:13 | src/compiler/utilitiesPublic.ts:1416:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1391:69 |
| src/compiler/utilitiesPublic.ts:1427:12 | src/compiler/utilitiesPublic.ts:1426:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1428:51 |
| src/compiler/utilitiesPublic.ts:1438:12 | src/compiler/utilitiesPublic.ts:1437:1 | NotYet: a getter in a checked field contract at src/compiler/sourcemap.ts:474:9 |
| src/compiler/utilitiesPublic.ts:1533:12 | src/compiler/utilitiesPublic.ts:1532:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:5140:20 |
| src/compiler/utilitiesPublic.ts:1575:12 | src/compiler/utilitiesPublic.ts:1574:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:1994:17 |
| src/compiler/utilitiesPublic.ts:1582:12 | src/compiler/utilitiesPublic.ts:1581:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1582:34 |
| src/compiler/utilitiesPublic.ts:1587:12 | src/compiler/utilitiesPublic.ts:1586:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1587:41 |
| src/compiler/utilitiesPublic.ts:1601:13 | src/compiler/utilitiesPublic.ts:1600:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1026:23 |
| src/compiler/utilitiesPublic.ts:1606:12 | src/compiler/utilitiesPublic.ts:1605:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:4662:66 |
| src/compiler/utilitiesPublic.ts:1681:56 | src/compiler/utilitiesPublic.ts:1680:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:1098:65 |
| src/compiler/utilitiesPublic.ts:1728:12 | src/compiler/utilitiesPublic.ts:1727:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:875:19 |
| src/compiler/utilitiesPublic.ts:1753:12 | src/compiler/utilitiesPublic.ts:1752:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1026:23 |
| src/compiler/utilitiesPublic.ts:1758:29 | src/compiler/utilitiesPublic.ts:1757:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:4662:66 |
| src/compiler/utilitiesPublic.ts:1779:12 | src/compiler/utilitiesPublic.ts:1778:1 | NotYet: a getter in a checked field contract at src/compiler/sourcemap.ts:474:9 |
| src/compiler/utilitiesPublic.ts:1873:12 | src/compiler/utilitiesPublic.ts:1872:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:2565:14 |
| src/compiler/utilitiesPublic.ts:1944:12 | src/compiler/utilitiesPublic.ts:1932:1 | NotYet: a getter in a checked field contract at src/compiler/sourcemap.ts:474:9 |
| src/compiler/utilitiesPublic.ts:1965:12 | src/compiler/utilitiesPublic.ts:1964:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:2876:37 |
| src/compiler/utilitiesPublic.ts:2144:13 | src/compiler/utilitiesPublic.ts:2143:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/utilitiesPublic.ts:2150:12 | src/compiler/utilitiesPublic.ts:2148:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/utilitiesPublic.ts:2161:12 | src/compiler/utilitiesPublic.ts:2160:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:11545:44 |
| src/compiler/utilitiesPublic.ts:2167:12 | src/compiler/utilitiesPublic.ts:2166:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:11545:44 |
| src/compiler/utilitiesPublic.ts:2171:12 | src/compiler/utilitiesPublic.ts:2170:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/checker.ts:6068:28 |
| src/compiler/utilitiesPublic.ts:2423:12 | src/compiler/utilitiesPublic.ts:2419:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilities.ts:11545:44 |
| src/compiler/utilitiesPublic.ts:2526:12 | src/compiler/utilitiesPublic.ts:2522:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1303:18 |
| src/compiler/utilitiesPublic.ts:2553:10 | src/compiler/utilitiesPublic.ts:2552:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/utilitiesPublic.ts:2593:100 | src/compiler/utilitiesPublic.ts:2592:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1026:23 |
| src/compiler/utilitiesPublic.ts:606:17 | src/compiler/utilitiesPublic.ts:603:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1032:23 |
| src/compiler/utilitiesPublic.ts:616:12 | src/compiler/utilitiesPublic.ts:615:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:1026:23 |
| src/compiler/utilitiesPublic.ts:645:9 | src/compiler/utilitiesPublic.ts:644:1 | NotYet: a checked field alias requiring an optional, accessor, or representation conversion at src/compiler/utilitiesPublic.ts:2574:14 |
| src/compiler/utilitiesPublic.ts:871:36 | src/compiler/utilitiesPublic.ts:870:1 | NotYet: a getter in a checked field contract at src/compiler/sourcemap.ts:474:9 |

</details>

## Files with a changed first stop

First recorded file finding, including the file refusal scan; this is distinct from independent unit stops.

| File | Before | After |
| --- | --- | --- |
| src/compiler/transformers/esnext.ts | NotYet src/compiler/transformers/esnext.ts:775:12: checked predicate membership by an enum tag another member can hold | NotYet src/compiler/checker.ts:6068:28: a checked field alias requiring an optional, accessor, or representation conversion |

## Panics

### src/compiler/moduleSpecifiers.ts:335:1

4ac3f04d/area/latent.jsonl.gz, 4ac3f04d/main/latent.jsonl.gz, 8502a46e/area/latent.jsonl.gz, 8502a46e/main/latent.jsonl.gz

```text
runtime error: invalid memory address or nil pointer dereference
```

## Timing on this box

GNU time -v; checker.ts is the sole root, with LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1. These are end-to-end latent driver measurements, including checker work and JSON/fsync, not isolated production lowering. Samples alternate before/after, sequentially.

| SHA | Sample | Wall (s) | User+system (s) | Max RSS (KiB) |
| --- | --- | --- | --- | --- |
| 4ac3f04d | 1 | 17.33 | 23.2 | 725984 |
| 8502a46e | 1 | 16.43 | 22.24 | 719664 |
| 4ac3f04d | 2 | 16.33 | 22.04 | 725636 |
| 8502a46e | 2 | 16.77 | 22.55 | 729352 |
| 4ac3f04d | 3 | 15.75 | 21.3 | 716328 |
| 8502a46e | 3 | 15.53 | 20.94 | 723108 |

## Box

```text
2026-10-27T05:19:57Z
5
400000 100000
               total        used        free      shared  buff/cache   available
Mem:     18882674688  1843412992 12925186048     8937472  4514095104 17039261696
Swap:              0           0           0
v24.19.0
go version go1.27.1 linux/amd64
clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)
Target: x86_64-unknown-linux-gnu
Thread model: posix
InstalledDir: /workspace/adamic-tools/llvm/bin
17179869184
```

## Reproduction and checks

Run `python3 stage3/meter/runs/walk-4ac3f04d/compare.py --k1t2wp9`. The original comparison mode remains available. Build, measurement and timing commands and raw logs are retained under k1t2wp9/.

- 4ac3f04d: per-ref pin and first-error metadata verified.
- 4ac3f04d: main/area streams agree after source-path and census-seconds normalization.
- 8502a46e: per-ref pin and first-error metadata verified.
- 8502a46e: main/area streams agree after source-path and census-seconds normalization.
- All six timing streams contain checker.ts alone; exit statuses are zero. Binary hashes match the retained after meter binary and the separately rebuilt before binary. The rebuilt before latent binary also exactly matches the SHA-256 of the committed baseline run (9ed61cdc7c4d8f0a60d4805f10f0d32ca8a2f18e643b14144ee5075496e05169).

### Timing tool

```text
time (GNU Time) UNKNOWN
Copyright (C) 2018 Free Software Foundation, Inc.
License GPLv3+: GNU GPL version 3 or later <https://gnu.org/licenses/gpl.html>.
This is free software: you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.

Written by David Keppel, David MacKenzie, and Assaf Gordon.
efc0d1112e36ec76c14bf7508e287935cf1631d9d0d91fb532802d07aef51cea  /workspace/adamic-tools/gnu-time/extracted/usr/bin/time
```

### Run and build status

```text
exit=0 elapsed=4286
exit=0 elapsed=52
```

All 746 adapted src files are byte-identical between the independently generated timing trees (input-check.txt). /usr/bin/time was absent; GNU time was extracted from the official Debian time_1.9-0.2_amd64.deb package, as in the prior run. Tool and package hashes are retained in timing/. The cgroup memory limit is 17179869184 bytes (16 GiB). The 8502a46e meter artifacts were committed and pushed immediately on exit as c4ea5bdd7, before comparison or rebuilding the before binary. The run took 4286 seconds. Both completed compiler source worktrees were removed during the latent pass to free tmpfs at the memory limit, matching the prior run cleanup. Binaries, overlays, adapted trees and logs were retained. No OOM occurred. Only test fixtures and a meter report change intervene between the pins; stage3 adaptation, census drivers, overlay generators and twice-daily.sh are unchanged; report.py and its tests/evidence changed.
