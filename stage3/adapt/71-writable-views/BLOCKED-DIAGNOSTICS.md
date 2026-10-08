Retained diagnostic views at f8dea48f: 158 location and five detached-location observations.
No public declaration or API sanction changes are permitted. These are capability requirements, not observations that tsc writes undefined.

For scalar location values the wider contract would need readonly `file`, `start`, and `length`. Detached values need the same protection: `file` must remain undefined and span fields remain numbers. Array views additionally need readonly element slots and readonly fields on exposed elements. Readonly containers alone leave element fields writable.

The census reports the first conflicting field, `file`, at all scalar sites below. Array reasons report the incompatible element domain; that element differs at `file`, `start`, and `length`. Preventing wider collection aliases from writing requires a coupled storage audit; keeping a public mutable alias could require a copy, which is a runtime edit. These sites are retained without asserting that every local reader necessarily needs a public API change.

| Site | Required fields | View |
| --- | --- | --- |
| src/compiler/binder.ts:853:49 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/binder.ts:861:75 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/binder.ts:864:53 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:12272:33 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:12276:33 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:13243:45 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:13357:29 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:15353:60 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:19332:41 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:19391:49 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:21438:21 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:21487:25 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:21502:25 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:21567:41 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:21568:45 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:21591:62 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:21600:37 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:21657:41 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:21658:45 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:22445:73 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:22454:106 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:22457:33 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:22936:42 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:24413:42 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:2515:15 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:2535:15 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:2575:43 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:2578:39 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:2590:40 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:2600:17 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:2604:35 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:2815:112 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:2815:177 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:2816:33 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:28748:48 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:28845:25 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:2923:37 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:30895:42 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:31480:46 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:3284:33 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:33274:47 | file, start, length | DiagnosticWithDetachedLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:33278:37 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:3386:13 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:33986:48 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:34855:25 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:34860:25 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:35111:47 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:35176:31 | file, start, length | DiagnosticWithLocation → Diagnostic | undefined |
| src/compiler/checker.ts:35191:39 | file, start, length | DiagnosticWithLocation | undefined → Diagnostic | undefined |
| src/compiler/checker.ts:35206:135 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:36130:42 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:36133:110 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:36136:37 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:36224:69 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:3633:51 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:36413:20 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:36440:20 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:36454:24 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:36456:20 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:36464:30 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:36467:30 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:36475:51 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:36491:24 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:36493:20 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:36592:37 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:36694:51 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:36745:32 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:36760:37 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:36782:48 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:37084:46 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation | undefined |
| src/compiler/checker.ts:37371:40 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:37378:25 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:37379:53 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:37393:40 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:37416:33 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:37480:38 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:37482:29 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:37483:71 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:3749:66 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:37554:37 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:38076:33 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:38205:29 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:3911:25 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:3973:48 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:39912:41 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:39923:37 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:39941:45 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:39957:52 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:39959:37 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:39980:53 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:4112:44 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:41167:37 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:4144:48 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:43319:25 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:43349:29 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:43818:33 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:43932:38 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:44000:21 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:44735:25 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:44738:33 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:45255:33 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:46373:54 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:46441:54 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:46804:48 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:47750:41 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:48422:33 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:4849:45 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:49035:53 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:49332:78 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:49616:45 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:49736:24 | element slots; element.file, element.start, element.length | DiagnosticWithLocation[] → Diagnostic[] |
| src/compiler/checker.ts:49742:63 | element slots; element.file, element.start, element.length | DiagnosticWithLocation[] → Diagnostic[] |
| src/compiler/checker.ts:49748:62 | element slots; element.file, element.start, element.length | DiagnosticWithLocation[] → Diagnostic[] |
| src/compiler/checker.ts:49751:20 | element slots; element.file, element.start, element.length | DiagnosticWithLocation[] → Diagnostic[] |
| src/compiler/checker.ts:51574:41 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:51672:33 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:51674:21 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:51676:33 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:51678:21 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:51843:29 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:52367:29 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:52662:64 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:52763:45 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:52772:41 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:52787:37 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:52801:56 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/checker.ts:52803:41 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:5284:41 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:53347:29 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:53356:29 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:53552:49 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/checker.ts:53573:29 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/commandLineParser.ts:2281:16 | file, start, length | DiagnosticWithLocation | undefined → Diagnostic | undefined |
| src/compiler/commandLineParser.ts:3592:21 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/commandLineParser.ts:3626:33 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/commandLineParser.ts:3636:33 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/commandLineParser.ts:3719:21 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/commandLineParser.ts:3796:9 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/emitter.ts:916:40 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/parser.ts:2510:17 | file, start, length | DiagnosticWithDetachedLocation → DiagnosticRelatedInformation |
| src/compiler/parser.ts:4573:25 | file, start, length | DiagnosticWithDetachedLocation → DiagnosticRelatedInformation |
| src/compiler/parser.ts:8470:25 | file, start, length | DiagnosticWithDetachedLocation → DiagnosticRelatedInformation |
| src/compiler/parser.ts:9633:63 | file, start, length | DiagnosticWithDetachedLocation → DiagnosticRelatedInformation |
| src/compiler/program.ts:1325:10 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/program.ts:2821:20 | element slots; element.file, element.start, element.length | DiagnosticWithLocation[] → readonly Diagnostic[] |
| src/compiler/program.ts:2824:96 | element slots; element.file, element.start, element.length | DiagnosticWithLocation[] → Diagnostic[] |
| src/compiler/program.ts:2916:17 | element slots; element.file, element.start, element.length | DiagnosticWithLocation[] → readonly Diagnostic[] | undefined |
| src/compiler/program.ts:2918:17 | element slots; element.file, element.start, element.length | DiagnosticWithLocation[] | undefined → readonly Diagnostic[] | undefined |
| src/compiler/program.ts:2937:30 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/program.ts:3131:45 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
| src/compiler/program.ts:3270:13 | element slots; element.file, element.start, element.length | DiagnosticWithLocation[] → Diagnostic[] |
| src/compiler/program.ts:4221:52 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/program.ts:4231:56 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/program.ts:4624:64 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/program.ts:4673:52 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/program.ts:4700:56 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/program.ts:4703:56 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/program.ts:4742:56 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/program.ts:4745:56 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/programDiagnostics.ts:172:57 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/programDiagnostics.ts:297:13 | file, start, length | DiagnosticWithLocation → Diagnostic |
| src/compiler/utilities.ts:6129:41 | element slots; element.file, element.start, element.length | DiagnosticWithLocation[] → Diagnostic[] |
| src/compiler/utilities.ts:8662:64 | file, start, length | DiagnosticWithLocation → DiagnosticRelatedInformation |
