Preserved diagnostic inputs in five internal generic consumers and narrowed two JSON error sinks to their push operation.
Base family commit: 3bc49948; fresh apply matches all 746 source files used by the successful oracle.
Census 1112 -> 1037: 75 removed, none added, identical checker diagnostics and eligibility.
Undoing addRelatedInfo restores 57 rows; a diagnostic.file writer is caught before adaptation edits.
Remaining diagnostic object and array views retain mutable contracts; no row is claimed to need a runtime change solely because it remains.

Internal generic inputs capture the diagnostic type they actually receive. addRelatedInfo now captures the related-information element type too, and its rest array is readonly: the function reads that array and pushes those references into the diagnostic's separate relatedInformation array. The other four generic inputs either read the diagnostic or forward the captured type to that helper. They never overwrite file, start or length. Existing category-copy and related-information construction writes remain exactly as before.

convertConfigFileToObject and convertToJson only use errors.push, and every value they append is created by createDiagnosticForNodeInSourceFile. Their parameter is now Pick<DiagnosticWithLocation[], "push">. They accept both a location-diagnostic array and a broader diagnostic sink without promising a view that can insert a non-location diagnostic into the former. The public convertToObject signature remains unchanged. diagnostics.json names all edits and seven reviewed bodies; runtime fingerprints reject changed algorithms before source edits.

Exact reason deltas: DiagnosticWithLocation as Diagnostic 93 -> 77; as DiagnosticRelatedInformation 58 -> 6; DiagnosticWithDetachedLocation as DiagnosticRelatedInformation 5 -> 0; DiagnosticWithLocation[] as Diagnostic[] 9 -> 7. Other optional and readonly-array views remain. The table's exact writable reasons receive all 75 removals. Cumulative writable-view count against main is 1174 -> 1037, or 137 removed including the first subset.

Evidence: complete normalized before/after/undo census, all source hashes, full oracle output, emitted JavaScript hashes, fresh-apply source identity, and lane verdict are under evidence/diagnostics/. All ten emitted JavaScript files and the API declaration are byte-identical to main. The baseline oracle ran all suites with four workers, no filter: 106366 passing, one failing, zero pending, only api/typescript.d.ts, exactly main. Install and build exit 0; tests exit 1 as on main. NODE_OPTIONS=--max-old-space-size=1536 and four-CPU affinity were retained.

Commands: bash stage3/apply.sh /tmp/unit71-diagnostics-replay; bash stage3/oracle/run.sh /tmp/unit71-diagnostics-lane2/adapted-tree /tmp/unit71-diagnostics-lane2/oracle3; python3 stage3/lane/check.py /tmp/unit71-diagnostics-check. The oracle source was repaired after the discarded collection experiment, then all 746 source files were proved byte-identical to the separate fresh apply replay. The lane evidence records both tree paths explicitly. The unmodified lane checker returns PASS; no expected counts, references or API sanctions changed. Tests took 877.979 seconds while the independent census runs shared the CPU quota.

The initial full lane run failed during clone because scratch trees filled the filesystem; it and the interrupted census were discarded. A second run caught the discarded collection experiment's isolatedDeclarations errors. That experiment removed 109 sites but was not buildable: a ReturnType derived from an inferred exported factory return violates tsc's explicit declaration requirement. Its build log is retained. Neither its census nor its signature changes are included in this commit.

Mutants: family-mutant.cjs diagnostics removes the generic readonly rest edit in addRelatedInfo. The complete census returns 1037 -> 1094, adding exactly 52 location and five detached related-information rows; diagnostics and eligibility stay identical. Separately, adding diagnostic.file = undefined to addErrorOrSuggestion makes adapt.cjs exit 1 with reviewed runtime body changed: addErrorOrSuggestion. The guard executes before the write loop. The original range and cache guards remain in force.

Unadapted sites and runtime limits

Every retained diagnostic site is listed below and in evidence/diagnostics/remaining.json. Most remaining scalar sites store diagnostics in arrays or DiagnosticCollection, rather than just read the parameter. Making only the incoming view readonly would expose it again through the collection's mutable Diagnostic entries or the public DiagnosticRelatedInformation fields. Mutable array returns additionally require proof of whether helpers return their input by identity: concatenate does so when either input is empty. Those cases are not completed no-write proofs.

A type-only fix could propagate readonly element fields through the storage and public API, but the landing lane's current API sanction does not permit that change. To retain the current writable API while protecting a narrower holder, a copy at the boundary would be a runtime edit. Such copies were not made. This is a conditional limitation of preserving the API, not evidence that tsc actually writes undefined into file at these sites. No supplied table row is reclassified as a language question by this family. The earlier literal-holder witnesses establish what a wider writer could do, not reachability of that writer inside tsc.

- `src/compiler/binder.ts:853:49`: a value of type DiagnosticWithLocation seen as DiagnosticRelatedInformation, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/binder.ts:864:53`: a value of type DiagnosticWithLocation seen as DiagnosticRelatedInformation, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/builder.ts:2116:17`: a value of type readonly DiagnosticWithLocation[] seen as readonly Diagnostic[], which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:13357:29`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:19332:41`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:19391:49`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:21567:41`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:21568:45`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:21657:41`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:21658:45`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:22445:73`: a value of type DiagnosticWithLocation seen as DiagnosticRelatedInformation, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:22454:90`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:22457:33`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:22936:42`: a value of type DiagnosticWithLocation seen as DiagnosticRelatedInformation, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:24413:42`: a value of type DiagnosticWithLocation seen as DiagnosticRelatedInformation, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:2515:15`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:2535:15`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:2604:35`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:2815:112`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:2815:177`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:28845:25`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:2923:37`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:33278:37`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:35176:31`: a value of type DiagnosticWithLocation seen as Diagnostic | undefined, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:35191:39`: a value of type DiagnosticWithLocation | undefined seen as Diagnostic | undefined, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:36133:94`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:36136:37`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:36413:20`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:36440:20`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:36454:24`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:36456:20`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:36464:30`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:36467:30`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:36491:24`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:36493:20`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:36592:37`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:36745:32`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:36760:37`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:37084:46`: a value of type DiagnosticWithLocation seen as DiagnosticRelatedInformation | undefined, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:37378:25`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:37416:33`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:37482:29`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:37554:37`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:38076:33`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:38205:29`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:39912:41`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:39923:37`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:39941:45`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:39959:37`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:43818:33`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:44738:33`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:46373:54`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:46441:54`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:47750:41`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:4849:45`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:49035:53`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:49616:45`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:49736:24`: a value of type DiagnosticWithLocation[] seen as Diagnostic[], which can write Diagnostic where DiagnosticWithLocation is read
- `src/compiler/checker.ts:49742:63`: a value of type DiagnosticWithLocation[] seen as Diagnostic[], which can write Diagnostic where DiagnosticWithLocation is read
- `src/compiler/checker.ts:49748:62`: a value of type DiagnosticWithLocation[] seen as Diagnostic[], which can write Diagnostic where DiagnosticWithLocation is read
- `src/compiler/checker.ts:49751:20`: a value of type DiagnosticWithLocation[] seen as Diagnostic[], which can write Diagnostic where DiagnosticWithLocation is read
- `src/compiler/checker.ts:51574:41`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:51672:33`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:51676:33`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:52763:45`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:52772:41`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:52787:37`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:52803:41`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:5284:41`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:53347:29`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:53356:29`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/checker.ts:53573:29`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/commandLineParser.ts:2281:16`: a value of type DiagnosticWithLocation | undefined seen as Diagnostic | undefined, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/commandLineParser.ts:3592:21`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/commandLineParser.ts:3626:33`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/commandLineParser.ts:3636:33`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/commandLineParser.ts:3719:21`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/commandLineParser.ts:3796:9`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/emitter.ts:916:40`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/program.ts:1325:10`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/program.ts:2821:20`: a value of type DiagnosticWithLocation[] seen as readonly Diagnostic[], which can write SourceFile | undefined where SourceFile is read
- `src/compiler/program.ts:2824:96`: a value of type DiagnosticWithLocation[] seen as Diagnostic[], which can write Diagnostic where DiagnosticWithLocation is read
- `src/compiler/program.ts:2916:17`: a value of type DiagnosticWithLocation[] seen as readonly Diagnostic[] | undefined, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/program.ts:2918:17`: a value of type DiagnosticWithLocation[] | undefined seen as readonly Diagnostic[] | undefined, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/program.ts:2937:30`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/program.ts:3270:13`: a value of type DiagnosticWithLocation[] seen as Diagnostic[], which can write Diagnostic where DiagnosticWithLocation is read
- `src/compiler/program.ts:4221:52`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/program.ts:4231:56`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/program.ts:4624:64`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/program.ts:4673:52`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/program.ts:4700:56`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/program.ts:4703:56`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/program.ts:4742:56`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/program.ts:4745:56`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/program.ts:5084:23`: a value of type readonly DiagnosticWithLocation[] seen as readonly Diagnostic[], which can write SourceFile | undefined where SourceFile is read
- `src/compiler/program.ts:645:45`: a value of type readonly DiagnosticWithLocation[] seen as readonly Diagnostic[] | undefined, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/programDiagnostics.ts:172:57`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/programDiagnostics.ts:297:115`: a value of type DiagnosticWithLocation[] | undefined seen as DiagnosticRelatedInformation[] | undefined, which can write DiagnosticRelatedInformation where DiagnosticWithLocation is read
- `src/compiler/programDiagnostics.ts:297:13`: a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/programDiagnostics.ts:298:61`: a value of type DiagnosticWithLocation[] | undefined seen as DiagnosticRelatedInformation[] | undefined, which can write DiagnosticRelatedInformation where DiagnosticWithLocation is read
- `src/compiler/utilities.ts:6129:41`: a value of type DiagnosticWithLocation[] seen as Diagnostic[], which can write Diagnostic where DiagnosticWithLocation is read
- `src/compiler/utilities.ts:8662:64`: a value of type DiagnosticWithLocation seen as DiagnosticRelatedInformation, which can write SourceFile | undefined where SourceFile is read
- `src/compiler/watch.ts:601:42`: a value of type readonly DiagnosticWithLocation[] seen as readonly Diagnostic[] | undefined, which can write SourceFile | undefined where SourceFile is read
