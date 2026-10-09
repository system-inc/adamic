The comparison uses the explicitly named baseline b3f83786 and cut 3fefe5b2. The cut's immediate parent is 1fe6d1bc, a merge with parents b3f83786 and 7b13a783. Thus the named baseline is the merge's first parent, not the wrapper cut's immediate parent. The compiler delta between the named revisions is confined to emit.go, emit_objects.go and object_metadata_test.go; other changed files are the profile branch's review evidence.

| Test family | Distinct emitted programs | Inputs |
| --- | ---: | --- |
| TestSixRuleAgreementAndMutants | 7 | sharded_suite; wrong-node, last-declaration, nullable-default and union-members generated drivers; released-inspect generated driver; fact_cost |
| TestTypeAwareAgreementAndMutants | 5 | main; wrong-main generated driver; released; query_cost; generated bad-kind |
| TestVolumeAgreementCompiler | 1 | generated volume-suite-with-program-roots |
| TestVolumeConfigGuardAndMutant | 1 | volume_suite |
| TestCSSPrinterAgreesWithGo | 4 | print_main; semicolon, indentation and width generated source variants |
| TestCompositionMatchesGoUnion | 1 | compose_main |
| TestUnicodeNodeShardPlantedFailure | 0 | Go overlay and Node --eval scanner only |

Counts are distinct C source programs, not native build invocations. Sanitized/unsanitized linking and checker Go archive overlays reuse the same C program. Corpus fixtures, generated lint witnesses, TypeScript compiler roots, manifests and shards are runtime inputs to these programs, not additional Adamic programs emitted to C. Optional benchmark rounds also reuse the same programs. The unlinked-library refusal probes stop before C emission.

The source inventory follows the cut's suite_test.go, typeaware_test.go, typeaware_budget_test.go, build_products_test.go, volume_test.go, volume_config_guard_self_prepare_test.go, CSS css_test.go, print_test.go, printer_shards_test.go, css_printer_30s_test.go and composition_shards_test.go, plus Unicode shards_test.go and canonicalize_test.go. TestVolumeAgreementCompiler is absent at the cut. Its current main volume_agreement_corpus_shards_test.go and volume_agreement_shards_test.go helpers introduce a program-roots adapter over volume_suite.ts; that entry file is unchanged between cut and delivery main. The adapter is included conservatively rather than claiming that the absent leaf exists at the cut.

prepare.py reproduces generated Adamic drivers and source mutations in scratch .a files. Original repository modules keep their original names. CSS scratch modules and imports use .a consistently. Both compiler builds consume the same absolute input paths; generated scratch locations and extensions are held constant across the comparison. No source is copied from cohere. manifest.json records every entry and its complete relative/absolute source-import closure with SHA256 hashes. The cohere gitlink is identical for both compilers.

probe.go.txt is built as a standalone temporary Go main inside each exact-revision worktree. It performs load.Load, EnableTSGo where the tests do, lower.Lower, and first-call native.TSGoC or native.C as appropriate. It never links native executables. Timing mode instead calls native.C directly on fresh IR and excludes load, lowering, TSGo library-body replacement, output writes and Go builds. time.py alternates compiler order over three trials. Largest means emitted C byte count.

Scope: byte identity and emitter timings only. No full package or full gate was run. Original semantic checker-archive and CSS mutants were materialized/emitted where they change Adamic inputs, but their semantic findings were not rerun. No new oracle fixtures, compiler changes or test leaves were added; counts.md and the call-target guard do not require updates. The Unicode named planted-failure leaf was run separately.
