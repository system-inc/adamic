# Slot 04 wave 18

Built isBackgroundSize, Theme.ResolveWith and resolveThemeArgument in separate `.a` modules. Claim 8f04f915 was pushed before code. All fifty previous retained helpers were complete and green on main b8fb957a, pushed through 3a99edeb. All twenty origin helper branches and every claim tree were checked after wildcard fetch. These unclaimed symbols tie the highest available fan-out, four consumers each; final inspection found no competing reservation. Main advanced to 39638d9e during validation, so the completed batch is rebased and rechecked before pushing; see the landing evidence added beside this report.

## Observations

Background-size handling preserves the Go skip/count behavior: cover,nonsense succeeds, while cover,a b c fails. Segmentation and length/percentage predicates are callback dependencies. The nested theme helper resolves a key with present=true, skips absent nested entries, returns inline values for option bit 1, otherwise uses a variable reference, and preserves the base value and map presence. Repeated nested names overwrite their map entry as in Go. The argument helper preserves trailing versus embedded wildcard paths and selects the final nested key, including a present empty value. It uses this batch's ResolveWith in the comparison driver. Theme key lookup, entry lookup, references and Resolve remain callback dependencies.

The actual Go consumer suite passed in 0.489s with temporary overlays. It captured 112 asserted fixtures across all four consumers and three live calls, all isBackgroundSize. Neither nested-theme helper was reached in that selected consumer suite; the 674 explicit Go controls exercise them directly. Consumer source texts are also supplied as helper input; this is not native execution of complete lint rules. No shared harness, registration, rule or compiler file was edited, and no regex matcher or finding-position conversion was introduced.

Pre-rebase `go test -count=1 -v ./stage1/cohere/lint/helpers/slot04_wave18` passed in 40.555s. Go, source Node, ASan/UBSan native and emitted JavaScript match 674 targeted controls, 112 consumer-source records and three captured-call records. Each of these 789 records observes the background result, nested-theme success/map presence/base value/each requested extra, and argument resolution. Controls cross background grammar, prefixes, options 0 through 3, namespace/nested/missing/literal/last-key arguments, repeated nested names, empty stored values, missing keys and allocated empty maps. Callback argument guards verify candidate/presence/namespaces/options in the driver.

Twelve semantic mutants compiled and ran successfully in every Adamic mode, then differed from actual Go:

- Omit cover keyword recognition.
- Invert all-valid group counting.
- Accept zero counted groups.
- Allow three parts instead of two.
- Invert base-key success.
- Skip present nested entries.
- Use reference bit 2 for nested inline selection.
- Use reference bit 2 for base inline selection.
- Omit trailing namespace resolution.
- Select the first nested key instead of the last.
- Invert nested resolution success.
- Accept a missing final nested key.

Four consumer omissions failed the independent coverage check. An exploratory suffix-guard mutation survived because its fallback reconstructed the same result in these controls; it is not credited. The final suite replaces it with the observable omitted-resolution mutant. The retained early helpers.log records that failed attempt; helpers-final.log is the successful final suite. Control regeneration reproduces identical SHA-256 bytes.

`go vet ./stage1/cohere/lint/helpers/slot04_wave18` passed with an empty log, and whitespace checks passed. A filtered uncached inherited-static-field Node oracle passed in 0.667s, with three native misses, two Node misses and zero cache hits. Logs are in evidence/. Inherited cloud setup passed: Go 0s, clang 1s, Node 1s, submodules 2s, warm 165s, total 165s; nproc printed 5 again. Build shells source /workspace/adamic-tools/env.sh. The landing rerun is reported separately.

## Readiness inference and limits

Each helper removes its named prerequisite from each of:

- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

That is twelve prerequisite removals across four distinct rules, assuming the external callback dependencies are supplied. None becomes free of every listed blocker through this batch alone. Unintegrated work on other branches is not counted as landed. All fifty-three retained helpers are complete.

Not covered: complete lint-rule findings/fixes/suggestions or registry integration, live reachability of the two nested-theme helpers in the selected consumer fixtures, new implementations of their supplied callbacks, arbitrary callback side effects or every theme profile, malformed UTF-8, exhaustive CSS inputs, and the full repository gate. The shared finding model 41eb6eab2 is not a dependency of these helpers, and incoming shared changes are not reverted.
