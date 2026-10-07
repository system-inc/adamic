Re-greened the three older claimed rule ports through an owned standalone .a driver on current main e8ba3d5d81de4d3773c723914fccd4c76248b965.
Commits: prior rule landing dcfd9bbf; helper branch b28c0599 already contains current main and its passing oracle evidence; this commit adds the standalone proof.
Checks: 192 fixtures and 1,041 file/rule corpus pairs match unchanged Go on source Node, emitted JavaScript and sanitized native; nine prior rules also agree on the newly added .a source.
Mutants: module name, optional parameter, direction exemption and literal message interpolation, all compiled and caught only by complete byte comparisons on all three backends.
Not covered: default shared registration/harness integration and unsupported JSX; no new helper or rule claimed under the landing cap.

The previous compatibility patch no longer applies to the shared harness. This driver imports the unchanged .a rule classes and their shared context/parser directly, dispatches the registered node kinds, and compares findings, exact UTF-8 byte ranges, descriptions and unchanged fixed source with a Go oracle that runs the original rules. These three rules supply no fixes or suggestions. The new driver refuses an unexpected shared repair shape rather than losing it; Go output includes any repair data, so any future Go-side change also changes the comparison bytes. The generator, test harness, compiler and production rule implementations were not edited.

Commands after source /workspace/adamic-tools/env.sh:

    python3 stage1/cohere/lint/rules/structure-tailwind-no-physical-direction/validate-standalone.py --scratch /tmp/wave15-three-standalone --typescript /tmp/lint-wave1-15-typescript --mutants --throughput
    python3 stage1/cohere/lint/rules/structure-tailwind-no-physical-direction/check-new-source.py --scratch /tmp/wave15-new-source-check --prior-six /tmp/wave15-landing-current-owned --prior-three /tmp/wave15-landing-current-literals
    ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v
    bash cloud/setup.sh

The comparison, supplemental check and sentinel exit zero. The sentinel passes in 0.298s, with one native and one Node cache miss, zero hits. Setup prints Go ready 0s, clang ready 0s, Node ready 0s and submodules ready 1s. Test warm-up fails compiling shared profile_test.go:32: cannot range over portFiles, a function. No successful warm/done timing is printed. Installed Go 1.27.1, clang 20.1.8 and Node 24.19.0 run the successful comparisons; nproc=5. Output is written to logs, never piped. Full repository gate is not claimed.

Unchanged Go test families TestNoAssignModuleVariable, TestDefaultParamLast and TestNoPhysicalDirection all pass. Captured supported cases: 13 module-variable, 115 default-parameter and 53 physical-direction, 181 total. Eight additional shape, Unicode, template, message-dollar and filename probes plus three witnesses bring the supported fixture manifest to 192. All four sides emit 51,192 identical bytes. Corpus: pinned TypeScript v6.0.3 commit 050880ce59e30b356b686bd3144efe24f875ebc8, 77 compiler files and 270 stage1 .ts/.a files, 347 total, 1,041 pairs and 39,659,578 identical bytes. Four-side hashes and gzipped Go output are retained.

One upstream physical-direction JSX input is independently reproduced: Go reports useLogicalClass, while source Node, emitted JavaScript and native exit 70 with an explicit shared parser slice refusal. It remains excluded from byte parity, rather than counted as a match. Prior literal-rule exclusions and shared repair serialization gaps remain documented in their completion and landing reports.

The production sources for the other nine claimed ports and the compiler are unchanged from dcfd9bbf. Their existing current-main evidence remains valid for its 346-file corpus; the supplemental independent comparison adds the new standalone.a source to all nine on Go, Node, emitted JavaScript and sanitized native. Across both evidence sets, all twelve have supported standalone parity over the resulting 347-file corpus (4,164 pairs), with 472 supported fixtures. These observations do not certify the default CLI formatter, all-rule ordering or converging fixer.

Every mutation begins independently from the production source. Each compiled and finished normally with empty stderr on all three backends, differing only in byte output:

- module_name_ignored changes the reserved name module to moduleNever; complete finding comparisons catch the dropped report.
- optional_parameter_ignored changes QuestionToken to NeverQuestion; comparisons catch the lost optional-parameter report.
- direction_exemption_removed removes the rtl/ltr exemption; comparisons catch extra reports.
- Literal interpolation changes split/join to JavaScript replacement-string interpolation; counts remain unchanged but full message bytes differ on dollar-containing classes. Before this mutant, the direction-exemption mutation is restored.

The unchanged shared generator still requires rule.ts, and the default shared lint package still fails profile compilation. These blockers prevent certifying the complete integrated rule branch, so the WIP cap remains active and no new claim was made. The already-pushed helper branch needs no additional changes: its main revision is unchanged and its two CSS helpers passed their full external comparisons and consumer tests there.

Throughput, observations on this host: best of three interleaved complete-process runs. Each includes the 77 compiler files and a 1,000-row synthetic source (78 files); startup, file reads, parsing and visiting are included, formatting/fixes/suggestions are skipped. Build time is excluded. Native correctness and mutants use sanitizers; throughput uses the optimized unsanitized binary. Default-parameter totals include 24 real compiler findings. Node is source Node.

| Rule | Findings | Native findings/s | Node findings/s | Go findings/s |
| --- | ---: | ---: | ---: | ---: |
| @next/next/no-assign-module-variable | 1000 | 982.41 | 1311.24 | 5569.67 |
| @typescript-eslint/default-param-last | 1024 | 913.06 | 1408.69 | 5506.11 |
| structure/tailwind-no-physical-direction | 2000 | 1861.04 | 2747.00 | 10724.18 |
