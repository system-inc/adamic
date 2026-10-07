# TypeScript 6.0.3 ruled discriminant writes

Stock checker 6.0.3; source `050880ce59e30b356b686bd3144efe24f875ebc8`; 77 compiler files, zero diagnostics. The comparison ledger is `origin/codex/stage3-fixtures-assertions` at `d464614`, `stage3/fixtures/assertions/discriminant-writes.json`.

**All 437 supplied candidates: 430 accepted, 7 refused. Of the 47 outside local construction: 40 accepted, 7 refused.** The plain blanket rule with the exact narrowing test refuses 32 of those 47. The other 15 have no compatible witnessed member under the exact test; a broad candidate flag alone is insufficient. The two parser events with construction not established both pass the ruled check. All 79 kind events are inside construction, as independently reported.

This adjudicates the complete independent candidate ledger by checker types at its AST locations. The construction classifications are the independent reviewed spans, not an invented escape proof. Adamic itself uses the existing IR confinement proof. The stock checker cannot prove that every compiler factory keeps an object confined. We therefore report the independently discovered extra assignment targets below separately, without pretending a whole-program construction count is established.

The cached status retag at `tsbuildPublic.ts:1921` is accepted: `UpToDate.type` declares all three up-to-date enum literals. The seven refused events write `FlowLabel.antecedent`; `FlowUnreachable` is structurally assignable to `FlowLabel` and declares that property as `undefined`. This is a conservative consequence of the ruled possible-member relation, not a Node observation.

Reproduce:

```sh
node notes/discriminant-writes/count-typescript.mjs /path/to/TypeScript-v6.0.3 /path/to/npm/typescript /path/to/independent/discriminant-writes.json > census.json
```

Prepare the source checkout using its diagnostic-generation script and installed compiler dependencies. The script exposes the stock private `isDiscriminantProperty` in memory, uses the inherited compiler tsconfig unchanged, discovers union witnesses at expressions/type nodes/declarations/constraints/call returns, and asks stock assignability for every compatible member. No textual search discovers writes.

## Event-by-event comparison

No supplied event is dropped or duplicated; source spans identify each event. The independent JSON records construction and candidate populations, rather than a final ruled acceptance verdict. Each row below records both versions and explains its adjudication.

| Event | Property | Construction | Plain | Ruled | Explanation |
| --- | --- | --- | --- | --- | --- |
| src/compiler/binder.ts:496:64 | antecedent | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/binder.ts:636:9 | flags | outside local construction | refused | accepted | exact discriminant; written type fits every possible member |
| src/compiler/binder.ts:1030:13 | flags | outside local construction | accepted | accepted | exact narrowing test has no compatible witnessed member for this receiver |
| src/compiler/binder.ts:1032:17 | flags | outside local construction | accepted | accepted | exact narrowing test has no compatible witnessed member for this receiver |
| src/compiler/binder.ts:1033:40 | flags | outside local construction | accepted | accepted | exact narrowing test has no compatible witnessed member for this receiver |
| src/compiler/binder.ts:1037:17 | flags | outside local construction | accepted | accepted | exact narrowing test has no compatible witnessed member for this receiver |
| src/compiler/binder.ts:1040:17 | flags | outside local construction | accepted | accepted | exact narrowing test has no compatible witnessed member for this receiver |
| src/compiler/binder.ts:1066:13 | flags | outside local construction | accepted | accepted | exact narrowing test has no compatible witnessed member for this receiver |
| src/compiler/binder.ts:1104:13 | flags | outside local construction | refused | accepted | exact discriminant; written type fits every possible member |
| src/compiler/binder.ts:1112:17 | flags | outside local construction | refused | accepted | exact discriminant; written type fits every possible member |
| src/compiler/binder.ts:1375:35 | antecedent | outside local construction | refused | refused | exact discriminant; written type fails a possible member; incompatible: FlowUnreachable |
| src/compiler/binder.ts:1688:13 | antecedent | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/binder.ts:1802:13 | flags | outside local construction | refused | accepted | exact discriminant; written type fits every possible member |
| src/compiler/binder.ts:2345:13 | flags | outside local construction | accepted | accepted | exact narrowing test has no compatible witnessed member for this receiver |
| src/compiler/binder.ts:2348:13 | flags | outside local construction | accepted | accepted | exact narrowing test has no compatible witnessed member for this receiver |
| src/compiler/binder.ts:3193:13 | commonJsModuleIndicator | outside local construction | refused | accepted | exact discriminant; written type fits every possible member |
| src/compiler/builder.ts:531:15 | file | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/builder.ts:606:9 | file | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/builder.ts:607:9 | file | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/builder.ts:1327:39 | signature | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/builder.ts:1939:37 | signature | outside local construction | refused | accepted | exact discriminant; written type fits every possible member |
| src/compiler/builder.ts:1943:37 | signature | outside local construction | refused | accepted | exact discriminant; written type fits every possible member |
| src/compiler/builder.ts:2231:30 | signature | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/builder.ts:2234:38 | signature | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/builder.ts:2271:75 | signature | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/builderState.ts:332:17 | signature | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/builderState.ts:395:9 | signature | outside local construction | refused | accepted | exact discriminant; written type fits every possible member |
| src/compiler/builderState.ts:458:9 | signature | outside local construction | refused | accepted | exact discriminant; written type fits every possible member |
| src/compiler/checker.ts:2564:41 | file | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/checker.ts:2729:13 | flags | outside local construction | refused | accepted | exact discriminant; written type fits every possible member |
| src/compiler/checker.ts:5005:9 | flags | outside local construction | refused | accepted | exact discriminant; written type fits every possible member |
| src/compiler/checker.ts:7877:62 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/checker.ts:9227:53 | sym | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/checker.ts:9919:48 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/checker.ts:13571:17 | flags | outside local construction | refused | accepted | exact discriminant; written type fits every possible member |
| src/compiler/checker.ts:13840:9 | flags | outside local construction | refused | accepted | exact discriminant; written type fits every possible member |
| src/compiler/checker.ts:14183:9 | flags | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/checker.ts:14264:17 | flags | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/checker.ts:16122:18 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/checker.ts:16122:24 | parameterName | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/checker.ts:16122:39 | parameterIndex | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/checker.ts:18453:17 | flags | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/checker.ts:20610:52 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/checker.ts:20614:52 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/checker.ts:20618:52 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/checker.ts:20622:52 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/checker.ts:20626:52 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/checker.ts:21705:44 | innerExpression | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/checker.ts:21711:44 | innerExpression | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/checker.ts:21711:82 | errorMessage | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/checker.ts:21716:44 | innerExpression | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/checker.ts:25934:9 | flags | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/checker.ts:25995:9 | flags | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/checker.ts:28634:31 | flags | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/checker.ts:28922:17 | antecedent | outside local construction | refused | refused | exact discriminant; written type fails a possible member; incompatible: FlowUnreachable |
| src/compiler/checker.ts:28924:17 | antecedent | outside local construction | refused | refused | exact discriminant; written type fails a possible member; incompatible: FlowUnreachable |
| src/compiler/checker.ts:28966:17 | antecedent | outside local construction | refused | refused | exact discriminant; written type fails a possible member; incompatible: FlowUnreachable |
| src/compiler/checker.ts:28968:17 | antecedent | outside local construction | refused | refused | exact discriminant; written type fails a possible member; incompatible: FlowUnreachable |
| src/compiler/checker.ts:29101:21 | antecedent | outside local construction | refused | refused | exact discriminant; written type fails a possible member; incompatible: FlowUnreachable |
| src/compiler/checker.ts:29103:21 | antecedent | outside local construction | refused | refused | exact discriminant; written type fails a possible member; incompatible: FlowUnreachable |
| src/compiler/checker.ts:33599:21 | flags | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/checker.ts:33606:25 | flags | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/checker.ts:36742:34 | file | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/checker.ts:37724:17 | flags | outside local construction | accepted | accepted | exact narrowing test has no compatible witnessed member for this receiver |
| src/compiler/checker.ts:41754:65 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/checker.ts:45283:74 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/checker.ts:49492:67 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:128:5 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:291:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:305:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:317:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:329:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:336:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:339:13 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:349:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:352:13 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:367:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:377:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:385:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:394:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:402:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:409:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:416:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:423:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:431:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:438:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:445:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:452:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:461:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:470:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:479:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:490:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:500:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:511:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:521:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:530:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:540:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:549:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:559:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:575:5 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:608:5 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:641:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:650:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:658:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:667:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:676:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:685:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:693:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:706:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:709:13 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:720:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:730:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:741:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:758:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:771:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:783:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:795:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:806:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:818:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:828:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:838:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:852:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:861:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:869:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:879:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:888:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:897:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:907:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:920:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:930:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:940:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:950:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:960:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:970:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:980:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:990:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1000:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1010:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1022:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1031:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1040:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1049:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1058:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1068:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1077:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1086:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1098:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1117:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1127:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1139:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1143:13 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1155:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1158:13 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1168:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1171:13 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1181:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1190:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1201:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1208:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1217:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1220:13 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1229:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1239:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1248:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1256:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1264:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1267:13 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1275:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1286:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1295:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1304:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1315:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1325:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1337:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1344:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1351:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1363:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1371:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1380:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1393:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1402:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1411:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1418:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1427:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1440:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1449:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1460:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1471:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1480:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1488:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1496:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1504:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1512:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1521:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1530:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1540:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1549:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1561:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1570:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1580:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1590:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1599:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1608:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1616:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1624:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1633:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1643:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1653:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1661:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1665:13 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1672:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1685:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1735:5 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1752:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1760:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1768:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1775:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1782:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1798:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1803:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1806:13 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1811:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1814:13 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:1819:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:2343:5 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:2346:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:2353:5 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:2359:5 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:2365:5 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:2374:13 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:2382:21 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:2385:25 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:2391:21 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:2394:25 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:2400:21 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:2403:25 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:2410:21 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/commandLineParser.ts:2413:25 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/expressionToTypeNode.ts:167:14 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/expressionToTypeNode.ts:167:20 | reportFallback | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/emitHelpers.ts:733:5 | scoped | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/emitHelpers.ts:747:5 | scoped | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/emitHelpers.ts:758:5 | scoped | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/emitHelpers.ts:770:5 | scoped | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/emitHelpers.ts:805:5 | scoped | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/emitHelpers.ts:822:5 | scoped | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/emitHelpers.ts:841:5 | scoped | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/emitHelpers.ts:849:5 | scoped | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/emitHelpers.ts:869:5 | scoped | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/emitHelpers.ts:882:5 | scoped | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/emitHelpers.ts:898:5 | scoped | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/emitHelpers.ts:918:5 | scoped | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/emitHelpers.ts:937:5 | scoped | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/emitHelpers.ts:961:5 | scoped | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/emitHelpers.ts:973:5 | scoped | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/emitHelpers.ts:996:5 | scoped | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/emitHelpers.ts:1012:5 | scoped | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/emitHelpers.ts:1023:5 | scoped | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/emitHelpers.ts:1036:5 | scoped | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/emitHelpers.ts:1117:5 | scoped | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/emitHelpers.ts:1154:5 | scoped | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/emitHelpers.ts:1173:5 | scoped | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/emitHelpers.ts:1187:5 | scoped | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/emitHelpers.ts:1214:5 | scoped | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/emitHelpers.ts:1224:5 | scoped | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/emitHelpers.ts:1284:5 | scoped | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/emitHelpers.ts:1347:5 | scoped | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/emitHelpers.ts:1372:5 | scoped | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/emitHelpers.ts:1383:5 | scoped | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/emitHelpers.ts:1417:5 | scoped | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/emitHelpers.ts:1454:5 | scoped | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/emitHelpers.ts:1469:5 | scoped | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/emitHelpers.ts:1477:5 | scoped | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/emitNode.ts:205:110 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/emitNode.ts:218:112 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/nodeFactory.ts:1335:39 | flags | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/nodeFactory.ts:2332:9 | modifiers | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/nodeFactory.ts:2357:13 | modifiers | outside local construction | refused | accepted | exact discriminant; written type fits every possible member |
| src/compiler/factory/nodeFactory.ts:2376:9 | modifiers | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/nodeFactory.ts:2936:9 | flags | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/nodeFactory.ts:3001:9 | flags | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/nodeFactory.ts:3071:9 | flags | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/nodeFactory.ts:3187:9 | modifiers | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/nodeFactory.ts:3255:9 | modifiers | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/nodeFactory.ts:3370:9 | operator | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/nodeFactory.ts:3396:9 | operator | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/nodeFactory.ts:3778:9 | flags | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/nodeFactory.ts:4278:9 | flags | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/nodeFactory.ts:4310:9 | modifiers | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/nodeFactory.ts:4376:17 | modifiers | outside local construction | refused | accepted | exact discriminant; written type fits every possible member |
| src/compiler/factory/nodeFactory.ts:4552:9 | flags | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/nodeFactory.ts:6049:9 | flags | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/nodeFactory.ts:6075:9 | externalModuleIndicator | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/nodeFactory.ts:6121:9 | flags | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/nodeFactory.ts:6136:9 | flags | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/nodeFactory.ts:6328:9 | flags | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/nodeFactory.ts:6337:9 | flags | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/nodeFactory.ts:6352:9 | flags | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/nodeFactory.ts:6361:9 | flags | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/nodeFactory.ts:6395:9 | flags | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/factory/nodeFactory.ts:7395:5 | flags | outside local construction | refused | accepted | exact discriminant; written type fits every possible member |
| src/compiler/moduleSpecifiers.ts:332:27 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/moduleSpecifiers.ts:400:13 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/moduleSpecifiers.ts:414:30 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/moduleSpecifiers.ts:415:37 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/moduleSpecifiers.ts:484:18 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/moduleSpecifiers.ts:507:26 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/moduleSpecifiers.ts:552:40 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/moduleSpecifiers.ts:553:45 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/moduleSpecifiers.ts:554:43 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/moduleSpecifiers.ts:555:11 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/parser.ts:1341:5 | externalModuleIndicator | outside local construction | refused | accepted | exact discriminant; written type fits every possible member |
| src/compiler/parser.ts:1403:5 | flags | outside local construction | accepted | accepted | exact narrowing test has no compatible witnessed member for this receiver |
| src/compiler/parser.ts:1857:13 | flags | outside local construction | accepted | accepted | exact narrowing test has no compatible witnessed member for this receiver |
| src/compiler/parser.ts:2603:13 | flags | outside local construction | accepted | accepted | exact narrowing test has no compatible witnessed member for this receiver |
| src/compiler/parser.ts:2611:13 | flags | outside local construction | accepted | accepted | exact narrowing test has no compatible witnessed member for this receiver |
| src/compiler/parser.ts:4447:13 | flags | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/parser.ts:6406:21 | flags | outside local construction | accepted | accepted | exact narrowing test has no compatible witnessed member for this receiver |
| src/compiler/parser.ts:6514:13 | flags | outside local construction | refused | accepted | exact discriminant; written type fits every possible member |
| src/compiler/parser.ts:7482:17 | flags | not established | refused | accepted | exact discriminant; written type fits every possible member |
| src/compiler/parser.ts:8113:21 | flags | not established | refused | accepted | exact discriminant; written type fits every possible member |
| src/compiler/parser.ts:9686:21 | flags | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/program.ts:1760:73 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/program.ts:1766:90 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/program.ts:1772:191 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/program.ts:1782:93 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/program.ts:1802:25 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/program.ts:1817:82 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/program.ts:1821:91 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/program.ts:2038:13 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/program.ts:3313:9 | flags | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/program.ts:3314:9 | flags | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/program.ts:3766:19 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/program.ts:3786:93 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/program.ts:3896:87 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/program.ts:3900:21 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/program.ts:3901:31 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/program.ts:3966:27 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/program.ts:4579:13 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/program.ts:4952:51 | directoryExists | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/scanner.ts:953:21 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/sourcemap.ts:534:69 | done | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/sourcemap.ts:559:37 | done | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/sys.ts:1619:30 | module | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/sys.ts:1619:59 | modulePath | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/sys.ts:1619:71 | error | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/sys.ts:1622:30 | module | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/sys.ts:1622:49 | modulePath | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/sys.ts:1622:72 | error | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/transformer.ts:338:146 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/transformers/classFields.ts:2110:73 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/transformers/classFields.ts:2133:77 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/transformers/classFields.ts:2795:17 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/transformers/classFields.ts:2806:17 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/transformers/classFields.ts:2838:13 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/transformers/classFields.ts:2865:17 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/transformers/classFields.ts:2897:17 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/transformers/classFields.ts:2923:13 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/transformers/es2015.ts:484:14 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/transformers/esDecorators.ts:351:17 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/transformers/esDecorators.ts:365:17 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/transformers/esDecorators.ts:382:17 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/transformers/esDecorators.ts:398:21 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/transformers/esDecorators.ts:888:19 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/transformers/esDecorators.ts:1296:34 | computed | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/transformers/esDecorators.ts:1299:34 | computed | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/transformers/esDecorators.ts:1304:38 | computed | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/transformers/esDecorators.ts:1309:38 | computed | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/transformers/esDecorators.ts:1315:17 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/transformers/esDecorators.ts:1407:69 | extraInitializersName | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/transformers/generators.ts:2189:13 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/transformers/generators.ts:2213:13 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/transformers/generators.ts:2310:13 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/transformers/generators.ts:2328:13 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/transformers/generators.ts:2355:13 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/transformers/generators.ts:2369:13 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/transformers/generators.ts:2390:13 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/transformers/generators.ts:2400:13 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/transformers/module/module.ts:2501:5 | scoped | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/tsbuildPublic.ts:905:9 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/tsbuildPublic.ts:946:9 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/tsbuildPublic.ts:1148:17 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/tsbuildPublic.ts:1154:52 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/tsbuildPublic.ts:1268:21 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/tsbuildPublic.ts:1304:13 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/tsbuildPublic.ts:1471:45 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/tsbuildPublic.ts:1477:49 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/tsbuildPublic.ts:1500:21 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/tsbuildPublic.ts:1509:25 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/tsbuildPublic.ts:1526:13 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/tsbuildPublic.ts:1535:13 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/tsbuildPublic.ts:1542:13 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/tsbuildPublic.ts:1553:13 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/tsbuildPublic.ts:1572:17 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/tsbuildPublic.ts:1587:17 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/tsbuildPublic.ts:1606:17 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/tsbuildPublic.ts:1626:17 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/tsbuildPublic.ts:1648:21 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/tsbuildPublic.ts:1680:13 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/tsbuildPublic.ts:1704:21 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/tsbuildPublic.ts:1712:21 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/tsbuildPublic.ts:1740:21 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/tsbuildPublic.ts:1757:17 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/tsbuildPublic.ts:1782:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/tsbuildPublic.ts:1800:18 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/tsbuildPublic.ts:1885:9 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/tsbuildPublic.ts:1921:29 | type | outside local construction | refused | accepted | exact discriminant; written type fits every possible member |
| src/compiler/tsbuildPublic.ts:1930:33 | type | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/types.ts:10250:9 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/types.ts:10254:9 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/types.ts:10258:9 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/types.ts:10261:9 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/types.ts:10264:9 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/types.ts:10268:9 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/types.ts:10272:9 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/types.ts:10276:9 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/types.ts:10280:9 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/utilities.ts:969:13 | flags | outside local construction | refused | accepted | exact discriminant; written type fits every possible member |
| src/compiler/utilities.ts:975:9 | flags | outside local construction | refused | accepted | exact discriminant; written type fits every possible member |
| src/compiler/utilities.ts:8473:5 | flags | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/utilities.ts:8491:5 | flags | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/utilities.ts:8499:5 | flags | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/utilities.ts:8509:5 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/utilities.ts:8511:5 | flags | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/utilities.ts:8523:5 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/utilities.ts:8525:5 | flags | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/utilities.ts:8535:5 | kind | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/utilities.ts:8537:5 | flags | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/utilities.ts:8622:9 | file | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/utilities.ts:8724:9 | file | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/utilities.ts:8739:9 | file | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/utilities.ts:8987:17 | externalModuleIndicator | outside local construction | refused | accepted | exact discriminant; written type fits every possible member |
| src/compiler/utilities.ts:8992:17 | externalModuleIndicator | outside local construction | refused | accepted | exact discriminant; written type fits every possible member |
| src/compiler/utilities.ts:9004:58 | externalModuleIndicator | outside local construction | refused | accepted | exact discriminant; written type fits every possible member |
| src/compiler/utilities.ts:10689:9 | flags | outside local construction | accepted | accepted | exact narrowing test has no compatible witnessed member for this receiver |
| src/compiler/watch.ts:846:9 | useCaseSensitiveFileNames | inside | accepted | accepted | same construction classification; accepted initialization |
| src/compiler/watchPublic.ts:808:17 | version | outside local construction | refused | accepted | exact discriminant; written type fits every possible member |

## Extra structural views absent from the independent candidate ledger

The independent discovery indexes property/root declarations. Adamic must also handle structural base interfaces: a narrower union member can inhabit a receiver even when its property declaration is different. A separate traversal of every explicit checker assignment target found **71 extra events: 45 assignability-safe; 26 requiring confinement or refusal**. This is a real scope difference, so the total site sets do not agree beyond the supplied 437. Factory construction and publication have not been independently proven for these extra events. Each location and incompatible member is retained in the JSON. Array tuple-length domains are included because the same exact checker predicate narrows on them; they are not excluded by an invented discriminant definition.

| Extra event | Property | Result before freshness | Incompatible members |
| --- | --- | --- | --- |
| src/compiler/core.ts:307:5 | length | requires freshness or refusal | [] |
| src/compiler/core.ts:312:5 | length | requires freshness or refusal | [string, string], [modifiers: readonly Modifier[] &#124; undefined, typeParameters: readonly TypeParameterDeclaration[] &#124; undefined, parameters: readonly ParameterDeclaration[], type: TypeNode], [typeParameters: readonly TypeParameterDeclaration[] &#124; undefined, parameters: readonly ParameterDeclaration[], type: TypeNode], [node: ConstructorTypeNode, modifiers: readonly Modifier[] &#124; undefined, typeParameters: NodeArray<TypeParameterDeclaration> &#124; undefined, parameters: NodeArray<...>, type: TypeNode], [node: ConstructorTypeNode, typeParameters: NodeArray<TypeParameterDeclaration> &#124; undefined, parameters: NodeArray<ParameterDeclaration>, type: TypeNode], [IndexSignatureDeclaration], [DiagnosticMessage], [DiagnosticMessage, string &#124; number], [DiagnosticMessage, string], [number], [number, Expression], [Statement], [Expression &#124; undefined], [Expression, Expression], [signature: string], [string], [fileId: IncrementalBuildInfoFileId], [fileId: IncrementalBuildInfoFileId, emitKind: BuilderFileEmit], [IncrementalBuildInfoFileId], [IncrementalBuildInfoFileId, BuilderFileEmit.None &#124; BuilderFileEmit.Js &#124; BuilderFileEmit.JsMap &#124; BuilderFileEmit.JsInlineMap &#124; ... 6 more ... &#124; BuilderFileEmit.All], [DiagnosticMessage, number], [DiagnosticMessage, number, string], [DiagnosticMessage, number, number], [DiagnosticMessage, string, string] |
| src/compiler/core.ts:1596:13 | length | requires freshness or refusal | [] |
| src/compiler/tracing.ts:71:9 | length | accepted |  |
| src/compiler/tracing.ts:163:9 | length | requires freshness or refusal | [] |
| src/compiler/tracing.ts:170:9 | length | accepted |  |
| src/compiler/sys.ts:233:17 | length | requires freshness or refusal | [] |
| src/compiler/factory/nodeFactory.ts:1623:9 | modifiers | accepted |  |
| src/compiler/factory/nodeFactory.ts:1654:9 | modifiers | accepted |  |
| src/compiler/factory/nodeFactory.ts:1729:9 | modifiers | accepted |  |
| src/compiler/factory/nodeFactory.ts:1773:9 | modifiers | accepted |  |
| src/compiler/factory/nodeFactory.ts:1822:9 | modifiers | accepted |  |
| src/compiler/factory/nodeFactory.ts:1869:9 | modifiers | accepted |  |
| src/compiler/factory/nodeFactory.ts:1953:9 | modifiers | accepted |  |
| src/compiler/factory/nodeFactory.ts:1975:13 | modifiers | accepted |  |
| src/compiler/factory/nodeFactory.ts:1987:9 | modifiers | accepted |  |
| src/compiler/factory/nodeFactory.ts:2043:9 | modifiers | accepted |  |
| src/compiler/factory/nodeFactory.ts:2106:9 | modifiers | accepted |  |
| src/compiler/factory/nodeFactory.ts:2232:9 | modifiers | accepted |  |
| src/compiler/factory/nodeFactory.ts:2698:9 | operator | accepted |  |
| src/compiler/factory/nodeFactory.ts:3189:9 | name | requires freshness or refusal | FunctionExpression & { readonly name?: undefined; }, Expression & FunctionExpression & { readonly name?: undefined; } |
| src/compiler/factory/nodeFactory.ts:3659:9 | modifiers | accepted |  |
| src/compiler/factory/nodeFactory.ts:3660:9 | name | requires freshness or refusal | ClassExpression & { readonly name?: undefined; }, Expression & ClassExpression & { readonly name?: undefined; } |
| src/compiler/factory/nodeFactory.ts:3878:9 | modifiers | accepted |  |
| src/compiler/factory/nodeFactory.ts:4391:9 | modifiers | accepted |  |
| src/compiler/factory/nodeFactory.ts:4444:9 | modifiers | accepted |  |
| src/compiler/factory/nodeFactory.ts:4481:9 | modifiers | accepted |  |
| src/compiler/factory/nodeFactory.ts:4516:9 | modifiers | accepted |  |
| src/compiler/factory/nodeFactory.ts:4551:9 | modifiers | accepted |  |
| src/compiler/factory/nodeFactory.ts:4628:9 | modifiers | accepted |  |
| src/compiler/factory/nodeFactory.ts:4643:13 | modifiers | accepted |  |
| src/compiler/factory/nodeFactory.ts:4656:9 | modifiers | accepted |  |
| src/compiler/factory/nodeFactory.ts:4658:9 | isTypeOnly | requires freshness or refusal | ImportEqualsDeclaration & { readonly isTypeOnly: true; } |
| src/compiler/factory/nodeFactory.ts:4698:9 | modifiers | accepted |  |
| src/compiler/factory/nodeFactory.ts:4732:9 | isTypeOnly | requires freshness or refusal | ImportClause & { readonly isTypeOnly: true; readonly name: Identifier; } |
| src/compiler/factory/nodeFactory.ts:4762:9 | token | accepted |  |
| src/compiler/factory/nodeFactory.ts:4813:9 | token | accepted |  |
| src/compiler/factory/nodeFactory.ts:4897:9 | isTypeOnly | requires freshness or refusal | ImportSpecifier & { readonly isTypeOnly: true; } |
| src/compiler/factory/nodeFactory.ts:4922:9 | modifiers | accepted |  |
| src/compiler/factory/nodeFactory.ts:4955:9 | modifiers | accepted |  |
| src/compiler/factory/nodeFactory.ts:4956:9 | isTypeOnly | requires freshness or refusal | ExportDeclaration & { readonly isTypeOnly: true; readonly moduleSpecifier: Expression; } |
| src/compiler/factory/nodeFactory.ts:4991:17 | modifiers | accepted |  |
| src/compiler/factory/nodeFactory.ts:5016:9 | isTypeOnly | requires freshness or refusal | ExportSpecifier & { readonly isTypeOnly: true; } |
| src/compiler/factory/nodeFactory.ts:5872:9 | token | accepted |  |
| src/compiler/factory/nodeFactory.ts:5930:9 | modifiers | accepted |  |
| src/compiler/factory/nodeFactory.ts:5949:13 | modifiers | accepted |  |
| src/compiler/factory/nodeFactory.ts:5966:9 | modifiers | accepted |  |
| src/compiler/factory/nodeFactory.ts:5984:13 | modifiers | accepted |  |
| src/compiler/factory/emitNode.ts:308:9 | length | requires freshness or refusal | [] |
| src/compiler/parser.ts:2283:17 | length | requires freshness or refusal | [] |
| src/compiler/parser.ts:6749:9 | modifiers | accepted |  |
| src/compiler/parser.ts:7547:13 | modifiers | accepted |  |
| src/compiler/parser.ts:7939:9 | modifiers | accepted |  |
| src/compiler/parser.ts:8142:9 | modifiers | accepted |  |
| src/compiler/parser.ts:8380:9 | modifiers | accepted |  |
| src/compiler/parser.ts:8867:13 | length | requires freshness or refusal | [] |
| src/compiler/checker.ts:1974:9 | length | accepted |  |
| src/compiler/checker.ts:8129:25 | length | requires freshness or refusal | [] |
| src/compiler/checker.ts:8132:25 | length | requires freshness or refusal | [] |
| src/compiler/checker.ts:16171:13 | length | requires freshness or refusal | [] |
| src/compiler/checker.ts:43553:17 | value | requires freshness or refusal | IteratorReturnResult<undefined> |
| src/compiler/checker.ts:51498:31 | modifiers | accepted |  |
| src/compiler/sourcemap.ts:316:13 | length | requires freshness or refusal | [number], [fileId: IncrementalBuildInfoFileId], [fileId: IncrementalBuildInfoFileId, emitKind: BuilderFileEmit], [IncrementalBuildInfoFileId], [IncrementalBuildInfoFileId, BuilderFileEmit.None &#124; BuilderFileEmit.Js &#124; BuilderFileEmit.JsMap &#124; BuilderFileEmit.JsInlineMap &#124; ... 6 more ... &#124; BuilderFileEmit.All] |
| src/compiler/transformers/esnext.ts:351:25 | length | accepted |  |
| src/compiler/program.ts:364:17 | length | requires freshness or refusal | [string, string], [signature: string], [string], [] |
| src/compiler/program.ts:371:13 | length | requires freshness or refusal | [string, string], [signature: string], [string], [] |
| src/compiler/builder.ts:1424:16 | length | requires freshness or refusal | [fileId: IncrementalBuildInfoFileId], [], [IncrementalBuildInfoFileId] |
| src/compiler/watch.ts:832:13 | version | requires freshness or refusal | FilePresenceUnknownOnHost, any |
| src/compiler/watchPublic.ts:773:21 | version | requires freshness or refusal | FilePresenceUnknownOnHost |
| src/compiler/expressionToTypeNode.ts:469:21 | modifiers | accepted |  |
| src/compiler/binder.ts:1016:21 | node | requires freshness or refusal | FlowUnreachable, FlowLabel, FlowAssignment, FlowCondition, FlowSwitchClause, FlowArrayMutation, FlowCall, FlowReduceLabel |

## Corrections to the earlier 30-site census

The earlier census is preserved as `baseline-census.json`, not presented as the final answer. It only discovered single-literal member domains and treated every factory as outside construction. The independent census includes creation events and wider/partial property domains. Every old site is accounted for below. Its one diagnostic was caused by forcing noEmit on the compiler project; the ruled census uses the inherited options unchanged and has zero diagnostics.

| Old event | Field | Comparison |
| --- | --- | --- |
| src/compiler/builder.ts:1424:16 | length | structural view absent from declaration-indexed ledger; requires freshness or refusal |
| src/compiler/checker.ts:1974:9 | length | structural view absent from declaration-indexed ledger; accepted |
| src/compiler/checker.ts:8129:25 | length | structural view absent from declaration-indexed ledger; requires freshness or refusal |
| src/compiler/checker.ts:8132:25 | length | structural view absent from declaration-indexed ledger; requires freshness or refusal |
| src/compiler/checker.ts:16171:13 | length | structural view absent from declaration-indexed ledger; requires freshness or refusal |
| src/compiler/core.ts:307:5 | length | structural view absent from declaration-indexed ledger; requires freshness or refusal |
| src/compiler/core.ts:312:5 | length | structural view absent from declaration-indexed ledger; requires freshness or refusal |
| src/compiler/core.ts:1596:13 | length | structural view absent from declaration-indexed ledger; requires freshness or refusal |
| src/compiler/factory/emitNode.ts:308:9 | length | structural view absent from declaration-indexed ledger; requires freshness or refusal |
| src/compiler/factory/nodeFactory.ts:3370:9 | operator | same event; inside; ruled accepted |
| src/compiler/factory/nodeFactory.ts:4658:9 | isTypeOnly | structural view absent from declaration-indexed ledger; requires freshness or refusal |
| src/compiler/factory/nodeFactory.ts:4732:9 | isTypeOnly | structural view absent from declaration-indexed ledger; requires freshness or refusal |
| src/compiler/factory/nodeFactory.ts:4897:9 | isTypeOnly | structural view absent from declaration-indexed ledger; requires freshness or refusal |
| src/compiler/factory/nodeFactory.ts:4956:9 | isTypeOnly | structural view absent from declaration-indexed ledger; requires freshness or refusal |
| src/compiler/factory/nodeFactory.ts:5016:9 | isTypeOnly | structural view absent from declaration-indexed ledger; requires freshness or refusal |
| src/compiler/parser.ts:2283:17 | length | structural view absent from declaration-indexed ledger; requires freshness or refusal |
| src/compiler/parser.ts:8867:13 | length | structural view absent from declaration-indexed ledger; requires freshness or refusal |
| src/compiler/program.ts:364:17 | length | structural view absent from declaration-indexed ledger; requires freshness or refusal |
| src/compiler/program.ts:371:13 | length | structural view absent from declaration-indexed ledger; requires freshness or refusal |
| src/compiler/sourcemap.ts:316:13 | length | structural view absent from declaration-indexed ledger; requires freshness or refusal |
| src/compiler/sys.ts:233:17 | length | structural view absent from declaration-indexed ledger; requires freshness or refusal |
| src/compiler/tracing.ts:71:9 | length | structural view absent from declaration-indexed ledger; accepted |
| src/compiler/tracing.ts:163:9 | length | structural view absent from declaration-indexed ledger; requires freshness or refusal |
| src/compiler/tracing.ts:170:9 | length | structural view absent from declaration-indexed ledger; accepted |
| src/compiler/transformers/esnext.ts:351:25 | length | structural view absent from declaration-indexed ledger; accepted |
| src/compiler/utilities.ts:8509:5 | kind | same event; inside; ruled accepted |
| src/compiler/utilities.ts:8523:5 | kind | same event; inside; ruled accepted |
| src/compiler/utilities.ts:8535:5 | kind | same event; inside; ruled accepted |
| src/compiler/watchPublic.ts:773:21 | version | structural view absent from declaration-indexed ledger; requires freshness or refusal |
| src/compiler/watchPublic.ts:808:17 | version | same event; outside local construction; ruled accepted |
