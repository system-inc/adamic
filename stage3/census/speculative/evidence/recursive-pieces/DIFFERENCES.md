Recursive createNodeBuilder differs from the prior-binary whole reference.
Both walks cover all 30,499 owned AST nodes; this is a context difference, not a dropped body.
The authoritative full identities and boundary inventories are in builder-union/UNION.json.
63 sites are missing, 51 added, and 10 shared sites have changed independently recounted depths.
57 typed boundaries are removed and 45 added. No recursive results are published in the full census.

| Change | Actual site | Syntax kind | UTF-8 span | Diagnostic location | Exact reason | Whole depth | Piece depth |
|---|---|---|---|---|---|---:|---:|
| Missing | checker.ts:10070:50 | KindIdentifier | 579146:579152 | checker.ts:10058:27 | a function value that captures the variable its own initializer declares | 0 |  |
| Missing | checker.ts:10063:69 | KindIdentifier | 578480:578481 | checker.ts:10063:35 | a function value that captures the variable its own initializer declares | 0 |  |
| Missing | checker.ts:10071:79 | KindIdentifier | 579239:579245 | checker.ts:10071:35 | a function value that captures the variable its own initializer declares | 0 |  |
| Missing | checker.ts:10100:45 | KindIdentifier | 581091:581108 | checker.ts:10100:45 | a union of differently held members variable a function value captures | 0 |  |
| Missing | checker.ts:10101:36 | KindIdentifier | 581172:581189 | checker.ts:10101:36 | a union of differently held members variable a function value captures | 0 |  |
| Missing | checker.ts:10101:96 | KindIdentifier | 581233:581249 | checker.ts:10101:96 | a union of differently held members variable a function value captures | 0 |  |
| Missing | checker.ts:10102:58 | KindIdentifier | 581311:581327 | checker.ts:10102:58 | a union of differently held members variable a function value captures | 0 |  |
| Missing | checker.ts:10126:39 | KindIdentifier | 582855:582872 | checker.ts:10126:39 | a union of differently held members variable a function value captures | 0 |  |
| Missing | checker.ts:10127:36 | KindIdentifier | 582936:582953 | checker.ts:10127:36 | a union of differently held members variable a function value captures | 0 |  |
| Missing | checker.ts:10127:96 | KindIdentifier | 582997:583013 | checker.ts:10127:96 | a union of differently held members variable a function value captures | 0 |  |
| Missing | checker.ts:10128:58 | KindIdentifier | 583075:583091 | checker.ts:10128:58 | a union of differently held members variable a function value captures | 0 |  |
| Missing | checker.ts:10159:69 | KindIdentifier | 585086:585104 | checker.ts:10159:27 | a function value that captures the variable its own initializer declares | 0 |  |
| Missing | checker.ts:10234:188 | KindIdentifier | 590771:590792 | checker.ts:10234:27 | a function value that captures the variable its own initializer declares | 0 |  |
| Missing | checker.ts:10307:66 | KindIdentifier | 595077:595102 | checker.ts:10306:23 | a function value that captures the variable its own initializer declares | 0 |  |
| Missing | checker.ts:10315:145 | KindIdentifier | 595789:595790 | checker.ts:10314:23 | a function value that captures the variable its own initializer declares | 1 |  |
| Missing | checker.ts:10339:156 | KindIdentifier | 597743:597744 | checker.ts:10338:23 | a function value that captures the variable its own initializer declares | 0 |  |
| Missing | checker.ts:6505:20 | KindIdentifier | 344394:344408 | checker.ts:6505:20 | a boolean \| undefined variable a function value captures | 0 |  |
| Missing | checker.ts:6939:70 | KindPropertyAccessExpression | 368163:368190 | checker.ts:6939:70 | checked view union arm awaits views-v3: array element kind: TypeParameter[] | 0 |  |
| Missing | checker.ts:7036:24 | KindIdentifier | 374751:374765 | checker.ts:7036:24 | a union of differently held members variable a function value captures | 0 |  |
| Missing | checker.ts:7036:57 | KindIdentifier | 374785:374798 | checker.ts:7036:57 | a union of differently held members variable a function value captures | 0 |  |
| Missing | checker.ts:7050:53 | KindPropertyAccessExpression | 375859:375887 | checker.ts:7050:53 | checked view union arm awaits views-v3: array element kind: TypeParameter[] | 0 |  |
| Missing | checker.ts:7051:21 | KindPropertyAccessExpression | 375888:375937 | checker.ts:7051:21 | checked view union arm awaits views-v3: array element kind: TypeParameter[] | 1 |  |
| Missing | checker.ts:7051:51 | KindPropertyAccessExpression | 375939:375969 | checker.ts:7051:51 | checked view union arm awaits views-v3: array element kind: TypeParameter[] | 1 |  |
| Missing | checker.ts:7053:21 | KindPropertyAccessExpression | 376097:376146 | checker.ts:7053:21 | checked view union arm awaits views-v3: array element kind: TypeParameter[] | 1 |  |
| Missing | checker.ts:7079:49 | KindPropertyAccessExpression | 378528:378556 | checker.ts:7079:49 | checked view union arm awaits views-v3: array element kind: TypeParameter[] | 0 |  |
| Missing | checker.ts:7080:17 | KindPropertyAccessExpression | 378557:378602 | checker.ts:7080:17 | checked view union arm awaits views-v3: array element kind: TypeParameter[] | 1 |  |
| Missing | checker.ts:7080:47 | KindPropertyAccessExpression | 378604:378634 | checker.ts:7080:47 | checked view union arm awaits views-v3: array element kind: TypeParameter[] | 1 |  |
| Missing | checker.ts:7082:17 | KindPropertyAccessExpression | 378725:378770 | checker.ts:7082:17 | checked view union arm awaits views-v3: array element kind: TypeParameter[] | 1 |  |
| Missing | checker.ts:7138:26 | KindIdentifier | 383224:383254 | checker.ts:7138:26 | a union of differently held members variable a function value captures | 0 |  |
| Missing | checker.ts:7172:26 | KindIdentifier | 386356:386386 | checker.ts:7172:26 | a union of differently held members variable a function value captures | 0 |  |
| Missing | checker.ts:7477:49 | KindPropertyAccessExpression | 407444:407476 | checker.ts:7477:49 | checked view union arm awaits views-v3: array element kind: TypeParameter[] | 0 |  |
| Missing | checker.ts:7505:29 | KindPropertyAccessExpression | 409470:409496 | checker.ts:7505:29 | checked view union arm awaits views-v3: array element kind: TypeParameter[] | 0 |  |
| Missing | checker.ts:7506:59 | KindPropertyAccessExpression | 409559:409585 | checker.ts:7506:59 | checked view union arm awaits views-v3: array element kind: TypeParameter[] | 0 |  |
| Missing | checker.ts:7524:63 | KindPropertyAccessExpression | 411113:411140 | checker.ts:7524:63 | checked view union arm awaits views-v3: array element kind: TypeParameter[] | 0 |  |
| Missing | checker.ts:7900:26 | KindIdentifier | 435065:435075 | checker.ts:7900:26 | a boolean \| undefined variable a function value captures | 0 |  |
| Missing | checker.ts:8388:93 | KindIdentifier | 465544:465566 | checker.ts:8388:93 | a boolean \| undefined variable a function value captures | 0 |  |
| Missing | checker.ts:8390:36 | KindIdentifier | 465894:465901 | checker.ts:8390:36 | a union of differently held members variable a function value captures | 0 |  |
| Missing | checker.ts:8393:35 | KindIdentifier | 466273:466284 | checker.ts:8393:35 | a union of differently held members variable a function value captures | 0 |  |
| Missing | checker.ts:8397:17 | KindIdentifier | 466477:466499 | checker.ts:8397:17 | a union of differently held members variable a function value captures | 0 |  |
| Missing | checker.ts:8457:70 | KindIdentifier | 470350:470368 | checker.ts:8457:70 | a boolean \| undefined variable a function value captures | 0 |  |
| Missing | checker.ts:9066:78 | KindPropertyAccessExpression | 509394:509439 | checker.ts:9066:78 | checked view union arm awaits views-v3: array element kind: TypeParameter[] | 0 |  |
| Missing | checker.ts:9117:21 | KindIdentifier | 512758:512782 | checker.ts:9117:21 | a boolean \| undefined variable a function value captures | 0 |  |
| Missing | checker.ts:9277:29 | KindIdentifier | 522686:522695 | checker.ts:9277:29 | a boolean \| undefined variable a function value captures | 0 |  |
| Missing | checker.ts:10860:74 | KindIdentifier | 635054:635058 | checker.ts:9347:19 | a function value that captures the variable its own initializer declares | 3 |  |
| Missing | checker.ts:10880:74 | KindIdentifier | 636589:636593 | checker.ts:9347:19 | a function value that captures the variable its own initializer declares | 3 |  |
| Missing | checker.ts:10899:95 | KindIdentifier | 637968:637981 | checker.ts:9347:19 | a function value that captures the variable its own initializer declares | 3 |  |
| Missing | checker.ts:10919:78 | KindIdentifier | 639529:639542 | checker.ts:9347:19 | a function value that captures the variable its own initializer declares | 3 |  |
| Missing | checker.ts:10860:74 | KindIdentifier | 635054:635058 | checker.ts:9348:19 | a function value that captures the variable its own initializer declares | 3 |  |
| Missing | checker.ts:10880:74 | KindIdentifier | 636589:636593 | checker.ts:9348:19 | a function value that captures the variable its own initializer declares | 3 |  |
| Missing | checker.ts:10899:95 | KindIdentifier | 637968:637981 | checker.ts:9348:19 | a function value that captures the variable its own initializer declares | 3 |  |
| Missing | checker.ts:10919:78 | KindIdentifier | 639529:639542 | checker.ts:9348:19 | a function value that captures the variable its own initializer declares | 3 |  |
| Missing | checker.ts:9578:22 | KindIdentifier | 542193:542218 | checker.ts:9578:22 | a boolean \| undefined variable a function value captures | 0 |  |
| Missing | checker.ts:9593:22 | KindIdentifier | 543032:543057 | checker.ts:9593:22 | a boolean \| undefined variable a function value captures | 0 |  |
| Missing | checker.ts:9655:70 | KindIdentifier | 548386:548406 | checker.ts:9655:70 | a union of differently held members variable a function value captures | 0 |  |
| Missing | checker.ts:9656:83 | KindIdentifier | 548570:548615 | checker.ts:9656:83 | a union of differently held members variable a function value captures | 0 |  |
| Missing | checker.ts:9670:25 | KindIdentifier | 549640:549684 | checker.ts:9670:25 | a union of differently held members variable a function value captures | 0 |  |
| Missing | checker.ts:9799:154 | KindIdentifier | 559907:559952 | checker.ts:9799:154 | a union of differently held members variable a function value captures | 0 |  |
| Missing | checker.ts:9799:98 | KindIdentifier | 559852:559871 | checker.ts:9799:98 | a union of differently held members variable a function value captures | 0 |  |
| Missing | checker.ts:9940:186 | KindIdentifier | 570969:570970 | checker.ts:9940:23 | a function value that captures the variable its own initializer declares | 0 |  |
| Missing | checker.ts:9967:37 | KindIdentifier | 572765:572771 | checker.ts:9967:37 | a union of differently held members variable a function value captures | 1 |  |
| Missing | checker.ts:9980:33 | KindIdentifier | 573335:573341 | checker.ts:9980:33 | a union of differently held members variable a function value captures | 1 |  |
| Missing | checker.ts:7275:127 | KindParenthesizedExpression | 394191:394218 | checker.ts:7275:128 | an unproven relation from T to TypeReference & T: the source is not assignable to the target | 0 |  |
| Missing | checker.ts:7275:76 | KindParenthesizedExpression | 394139:394167 | checker.ts:7275:77 | an unproven relation from T to TypeReference & T: the source is not assignable to the target | 0 |  |
| Added | checker.ts:10281:53 | KindIdentifier | 593440:593459 | checker.ts:10281:53 | an overloaded function as a value |  | 0 |
| Added | checker.ts:10324:25 | KindIdentifier | 596457:596526 | checker.ts:10324:25 | an overloaded function as a value |  | 0 |
| Added | checker.ts:10336:42 | KindIdentifier | 597243:597287 | checker.ts:10336:42 | an overloaded function as a value |  | 1 |
| Added | checker.ts:10338:39 | KindIdentifier | 597541:597585 | checker.ts:10338:39 | an overloaded function as a value |  | 0 |
| Added | checker.ts:10502:39 | KindIdentifier | 609148:609161 | checker.ts:10502:39 | an overloaded function as a value |  | 0 |
| Added | checker.ts:10706:37 | KindIdentifier | 623897:623947 | checker.ts:10706:37 | an overloaded function as a value |  | 0 |
| Added | checker.ts:10787:13 | KindFunctionDeclaration | 629884:630540 | checker.ts:10787:13 | a block-scoped nested function declaration |  | 0 |
| Added | checker.ts:10798:13 | KindFunctionDeclaration | 630540:631149 | checker.ts:10798:13 | a block-scoped nested function declaration |  | 0 |
| Added | checker.ts:6875:25 | KindCallExpression | 364359:364438 | checker.ts:54330:5 | a value of type (ModuleSpecifierResolutionHost & { getCommonSourceDirectory(): string; }) \| undefined |  | 1 |
| Added | checker.ts:6875:25 | KindExpressionStatement | 364359:364439 | checker.ts:54330:5 | a value of type (ModuleSpecifierResolutionHost & { getCommonSourceDirectory(): string; }) \| undefined |  | 0 |
| Added | checker.ts:7299:29 | KindCallExpression | 395765:395992 | checker.ts:54330:5 | a value of type (ModuleSpecifierResolutionHost & { getCommonSourceDirectory(): string; }) \| undefined |  | 2 |
| Added | checker.ts:7685:29 | KindCallExpression | 421025:421163 | checker.ts:54330:5 | a value of type (ModuleSpecifierResolutionHost & { getCommonSourceDirectory(): string; }) \| undefined |  | 1 |
| Added | checker.ts:7685:29 | KindExpressionStatement | 421025:421164 | checker.ts:54330:5 | a value of type (ModuleSpecifierResolutionHost & { getCommonSourceDirectory(): string; }) \| undefined |  | 0 |
| Added | checker.ts:7775:21 | KindCallExpression | 426777:426876 | checker.ts:54330:5 | a value of type (ModuleSpecifierResolutionHost & { getCommonSourceDirectory(): string; }) \| undefined |  | 1 |
| Added | checker.ts:7775:21 | KindExpressionStatement | 426777:426877 | checker.ts:54330:5 | a value of type (ModuleSpecifierResolutionHost & { getCommonSourceDirectory(): string; }) \| undefined |  | 0 |
| Added | checker.ts:8148:25 | KindCallExpression | 450615:450822 | checker.ts:54330:5 | a value of type (ModuleSpecifierResolutionHost & { getCommonSourceDirectory(): string; }) \| undefined |  | 2 |
| Added | checker.ts:8444:17 | KindCallExpression | 469442:469534 | checker.ts:54330:5 | a value of type (ModuleSpecifierResolutionHost & { getCommonSourceDirectory(): string; }) \| undefined |  | 1 |
| Added | checker.ts:8444:17 | KindExpressionStatement | 469442:469535 | checker.ts:54330:5 | a value of type (ModuleSpecifierResolutionHost & { getCommonSourceDirectory(): string; }) \| undefined |  | 0 |
| Added | checker.ts:8456:13 | KindCallExpression | 470191:470279 | checker.ts:54330:5 | a value of type (ModuleSpecifierResolutionHost & { getCommonSourceDirectory(): string; }) \| undefined |  | 1 |
| Added | checker.ts:8456:13 | KindExpressionStatement | 470191:470280 | checker.ts:54330:5 | a value of type (ModuleSpecifierResolutionHost & { getCommonSourceDirectory(): string; }) \| undefined |  | 0 |
| Added | checker.ts:8720:29 | KindCallExpression | 487472:487617 | checker.ts:54330:5 | a value of type (ModuleSpecifierResolutionHost & { getCommonSourceDirectory(): string; }) \| undefined |  | 1 |
| Added | checker.ts:8720:29 | KindExpressionStatement | 487472:487618 | checker.ts:54330:5 | a value of type (ModuleSpecifierResolutionHost & { getCommonSourceDirectory(): string; }) \| undefined |  | 0 |
| Added | checker.ts:9198:21 | KindCallExpression | 517746:517813 | checker.ts:54330:5 | a value of type (ModuleSpecifierResolutionHost & { getCommonSourceDirectory(): string; }) \| undefined |  | 1 |
| Added | checker.ts:9198:21 | KindExpressionStatement | 517746:517814 | checker.ts:54330:5 | a value of type (ModuleSpecifierResolutionHost & { getCommonSourceDirectory(): string; }) \| undefined |  | 0 |
| Added | checker.ts:9392:31 | KindNewExpression | 530096:530175 | checker.ts:54330:5 | a value of type (ModuleSpecifierResolutionHost & { getCommonSourceDirectory(): string; }) \| undefined |  | 1 |
| Added | checker.ts:9617:21 | KindCallExpression | 544892:545033 | checker.ts:54330:5 | a value of type (ModuleSpecifierResolutionHost & { getCommonSourceDirectory(): string; }) \| undefined |  | 1 |
| Added | checker.ts:9617:21 | KindExpressionStatement | 544892:545034 | checker.ts:54330:5 | a value of type (ModuleSpecifierResolutionHost & { getCommonSourceDirectory(): string; }) \| undefined |  | 0 |
| Added | checker.ts:9727:33 | KindCallExpression | 554779:554902 | checker.ts:54330:5 | a value of type (ModuleSpecifierResolutionHost & { getCommonSourceDirectory(): string; }) \| undefined |  | 1 |
| Added | checker.ts:9727:33 | KindExpressionStatement | 554779:554903 | checker.ts:54330:5 | a value of type (ModuleSpecifierResolutionHost & { getCommonSourceDirectory(): string; }) \| undefined |  | 0 |
| Added | checker.ts:6498:9 | KindFunctionDeclaration | 343821:343936 | checker.ts:6498:9 | a block-scoped nested function declaration |  | 0 |
| Added | checker.ts:6499:9 | KindFunctionDeclaration | 343936:344059 | checker.ts:6499:9 | a block-scoped nested function declaration |  | 0 |
| Added | checker.ts:7054:93 | KindIdentifier | 376267:376286 | checker.ts:7054:93 | an overloaded function as a value |  | 0 |
| Added | checker.ts:7055:94 | KindIdentifier | 376429:376448 | checker.ts:7055:94 | an overloaded function as a value |  | 0 |
| Added | checker.ts:7164:93 | KindIdentifier | 385579:385598 | checker.ts:7164:93 | an overloaded function as a value |  | 1 |
| Added | checker.ts:7202:58 | KindIdentifier | 389006:389026 | checker.ts:7202:58 | an overloaded function as a value |  | 0 |
| Added | checker.ts:8334:46 | KindIdentifier | 461604:461623 | checker.ts:8334:46 | an overloaded function as a value |  | 0 |
| Added | checker.ts:8352:73 | KindIdentifier | 462862:462882 | checker.ts:8352:73 | an overloaded function as a value |  | 0 |
| Added | checker.ts:8836:26 | KindIdentifier | 494487:494500 | checker.ts:8836:26 | an overloaded function as a value |  | 0 |
| Added | checker.ts:8872:9 | KindFunctionDeclaration | 496663:496801 | checker.ts:8872:9 | a block-scoped nested function declaration |  | 0 |
| Added | checker.ts:8873:9 | KindFunctionDeclaration | 496801:496938 | checker.ts:8873:9 | a block-scoped nested function declaration |  | 0 |
| Added | checker.ts:9286:26 | KindIdentifier | 523374:523394 | checker.ts:9286:26 | an overloaded function as a value |  | 0 |
| Added | checker.ts:9332:26 | KindIdentifier | 526124:526144 | checker.ts:9332:26 | an overloaded function as a value |  | 0 |
| Added | checker.ts:9690:92 | KindIdentifier | 551292:551310 | checker.ts:9690:92 | a union of differently held members variable a function value captures |  | 0 |
| Added | checker.ts:9953:13 | KindFunctionDeclaration | 571454:571608 | checker.ts:9953:13 | a block-scoped nested function declaration |  | 0 |
| Added | checker.ts:9954:13 | KindFunctionDeclaration | 571608:571779 | checker.ts:9954:13 | a block-scoped nested function declaration |  | 0 |
| Added | factory/utilitiesPublic.ts:10:1 | KindFunctionDeclaration | 168:368 | factory/utilitiesPublic.ts:10:17 | a function returning T |  | 0 |
| Added | utilities.ts:10644:1 | KindFunctionDeclaration | 426479:426749 | utilities.ts:10644:17 | a function returning T |  | 0 |
| Added | utilities.ts:10654:1 | KindFunctionDeclaration | 426749:427019 | utilities.ts:10654:17 | a function returning T |  | 0 |
| Added | utilities.ts:10664:1 | KindFunctionDeclaration | 427019:427324 | utilities.ts:10664:17 | a function returning T |  | 0 |
| Added | utilities.ts:749:15 | KindIdentifier | 21316:21323 | utilities.ts:749:15 | a value of type U \| undefined |  | 2 |
| Added | utilities.ts:749:24 | KindCallExpression | 21325:21346 | utilities.ts:749:24 | a call returning U \| undefined |  | 2 |
| Depth | checker.ts:7685:109 | KindPropertyAccessExpression | 421135:421161 | checker.ts:7685:109 | a value of type __String | 0 | 2 |
| Depth | checker.ts:8444:73 | KindIdentifier | 469515:469527 | checker.ts:8444:73 | reading SymbolFlags | 0 | 2 |
| Depth | checker.ts:8720:124 | KindPropertyAccessExpression | 487597:487615 | checker.ts:8720:124 | a value of type __String | 0 | 2 |
| Depth | checker.ts:9392:71 | KindPropertyAccessExpression | 530136:530174 | checker.ts:9392:71 | a value of type (ModuleSpecifierResolutionHost & { getCommonSourceDirectory(): string; }) \| undefined | 1 | 2 |
| Depth | checker.ts:9617:59 | KindIdentifier | 544952:544956 | checker.ts:9617:59 | a generic function as a value | 0 | 2 |
| Depth | checker.ts:9617:64 | KindPropertyAccessExpression | 544957:544976 | checker.ts:9617:64 | checked view union arm awaits views-v3: array element kind: Declaration[] | 0 | 2 |
| Depth | checker.ts:9617:90 | KindIdentifier | 544982:545002 | checker.ts:9617:90 | an overloaded function as a value | 0 | 2 |
| Depth | checker.ts:9727:104 | KindIdentifier | 554883:554895 | checker.ts:9727:104 | reading SymbolFlags | 0 | 2 |
| Depth | utilities.ts:749:33 | KindIdentifier | 21335:21340 | utilities.ts:749:33 | reading value | 2 | 3 |
| Depth | utilities.ts:749:40 | KindIdentifier | 21341:21345 | utilities.ts:749:40 | reading key | 2 | 3 |

| Boundary change | File | UTF-8 span | Stock syntax kind code |
|---|---|---|---:|
| Removed | checker.ts | 344394:344408 | 80 |
| Removed | checker.ts | 368163:368190 | 212 |
| Removed | checker.ts | 374751:374765 | 80 |
| Removed | checker.ts | 374785:374798 | 80 |
| Removed | checker.ts | 375859:375887 | 212 |
| Removed | checker.ts | 375888:375937 | 212 |
| Removed | checker.ts | 375939:375969 | 212 |
| Removed | checker.ts | 376097:376146 | 212 |
| Removed | checker.ts | 378528:378556 | 212 |
| Removed | checker.ts | 378557:378602 | 212 |
| Removed | checker.ts | 378604:378634 | 212 |
| Removed | checker.ts | 378725:378770 | 212 |
| Removed | checker.ts | 383224:383254 | 80 |
| Removed | checker.ts | 386356:386386 | 80 |
| Removed | checker.ts | 407444:407476 | 212 |
| Removed | checker.ts | 409470:409496 | 212 |
| Removed | checker.ts | 409559:409585 | 212 |
| Removed | checker.ts | 411113:411140 | 212 |
| Removed | checker.ts | 435065:435075 | 80 |
| Removed | checker.ts | 465544:465566 | 80 |
| Removed | checker.ts | 465894:465901 | 80 |
| Removed | checker.ts | 466273:466284 | 80 |
| Removed | checker.ts | 466477:466499 | 80 |
| Removed | checker.ts | 470350:470368 | 80 |
| Removed | checker.ts | 509394:509439 | 212 |
| Removed | checker.ts | 512758:512782 | 80 |
| Removed | checker.ts | 522686:522695 | 80 |
| Removed | checker.ts | 542193:542218 | 80 |
| Removed | checker.ts | 543032:543057 | 80 |
| Removed | checker.ts | 548386:548406 | 80 |
| Removed | checker.ts | 548570:548615 | 80 |
| Removed | checker.ts | 549640:549684 | 80 |
| Removed | checker.ts | 559852:559871 | 80 |
| Removed | checker.ts | 559907:559952 | 80 |
| Removed | checker.ts | 570969:570970 | 80 |
| Removed | checker.ts | 572765:572771 | 80 |
| Removed | checker.ts | 573335:573341 | 80 |
| Removed | checker.ts | 578480:578481 | 80 |
| Removed | checker.ts | 579146:579152 | 80 |
| Removed | checker.ts | 579239:579245 | 80 |
| Removed | checker.ts | 581091:581108 | 80 |
| Removed | checker.ts | 581172:581189 | 80 |
| Removed | checker.ts | 581233:581249 | 80 |
| Removed | checker.ts | 581311:581327 | 80 |
| Removed | checker.ts | 582855:582872 | 80 |
| Removed | checker.ts | 582936:582953 | 80 |
| Removed | checker.ts | 582997:583013 | 80 |
| Removed | checker.ts | 583075:583091 | 80 |
| Removed | checker.ts | 585086:585104 | 80 |
| Removed | checker.ts | 590771:590792 | 80 |
| Removed | checker.ts | 595077:595102 | 80 |
| Removed | checker.ts | 595789:595790 | 80 |
| Removed | checker.ts | 597743:597744 | 80 |
| Removed | checker.ts | 635054:635058 | 80 |
| Removed | checker.ts | 636589:636593 | 80 |
| Removed | checker.ts | 637968:637981 | 80 |
| Removed | checker.ts | 639529:639542 | 80 |
| Added | checker.ts | 343821:343936 | 263 |
| Added | checker.ts | 343936:344059 | 263 |
| Added | checker.ts | 364359:364438 | 214 |
| Added | checker.ts | 364359:364439 | 245 |
| Added | checker.ts | 376267:376286 | 80 |
| Added | checker.ts | 376429:376448 | 80 |
| Added | checker.ts | 385579:385598 | 80 |
| Added | checker.ts | 389006:389026 | 80 |
| Added | checker.ts | 395765:395992 | 214 |
| Added | checker.ts | 421025:421163 | 214 |
| Added | checker.ts | 421025:421164 | 245 |
| Added | checker.ts | 426777:426876 | 214 |
| Added | checker.ts | 426777:426877 | 245 |
| Added | checker.ts | 450615:450822 | 214 |
| Added | checker.ts | 461604:461623 | 80 |
| Added | checker.ts | 462862:462882 | 80 |
| Added | checker.ts | 469442:469534 | 214 |
| Added | checker.ts | 469442:469535 | 245 |
| Added | checker.ts | 470191:470279 | 214 |
| Added | checker.ts | 470191:470280 | 245 |
| Added | checker.ts | 487472:487617 | 214 |
| Added | checker.ts | 487472:487618 | 245 |
| Added | checker.ts | 494487:494500 | 80 |
| Added | checker.ts | 496801:496938 | 263 |
| Added | checker.ts | 517746:517813 | 214 |
| Added | checker.ts | 517746:517814 | 245 |
| Added | checker.ts | 523374:523394 | 80 |
| Added | checker.ts | 526124:526144 | 80 |
| Added | checker.ts | 530096:530175 | 215 |
| Added | checker.ts | 544892:545033 | 214 |
| Added | checker.ts | 544892:545034 | 245 |
| Added | checker.ts | 551292:551310 | 80 |
| Added | checker.ts | 554779:554902 | 214 |
| Added | checker.ts | 554779:554903 | 245 |
| Added | checker.ts | 571608:571779 | 263 |
| Added | checker.ts | 593440:593459 | 80 |
| Added | checker.ts | 596457:596526 | 80 |
| Added | checker.ts | 597243:597287 | 80 |
| Added | checker.ts | 597541:597585 | 80 |
| Added | checker.ts | 609148:609161 | 80 |
| Added | checker.ts | 623897:623947 | 80 |
| Added | checker.ts | 629884:630540 | 263 |
| Added | checker.ts | 630540:631149 | 263 |
| Added | utilities.ts | 21316:21323 | 80 |
| Added | utilities.ts | 21325:21346 | 214 |
