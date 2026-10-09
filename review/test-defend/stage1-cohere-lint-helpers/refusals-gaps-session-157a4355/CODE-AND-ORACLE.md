# Code under test

Production TypeScript helper port: main.ts dispatch and OptionsJson, OptionSchema, StrictOptions and PolicyMessage. R1 mutates PolicyMessage.template's missing-message rejection bound. G1 drops StrictOptions.matches's unsupported-kind rejection statement. Original-file locations and full edits are in plan.json and standalone diffs. Production sources and tests are restored. Neither the test/oracle inputs nor their expected answers are mutated.

# Oracle

TestMessageRefusalsMatchGo executes independent Go cohere policy.Messages.Render and compares the port's exact panic stderr. Exit 70 and the Adamic prefix are self expectations. Source Node executes the implementation; it is not the independent expected-answer authority. Exact stdout is not asserted in this refusal row.

TestKnownGapsAreExplicit compares source Node and sanitized native against handwritten NotYet labels and valid recovery output. This is a self oracle, not Go-derived diagnostic text. The common failure formatter says Go even for this row's handwritten labels.

TestHelpersMatchCohere executes independent Go helper answers and compares source Node and sanitized native over 23539 corpus cases. TestHelperMutants is a witness executing its own compiled builtin mutations; its production-mutant results are recorded but do not alone prove the quality of its witnessed comparison.

# Coverage and semantic differences

Three clean per-test go -coverprofile runs use -coverpkg on internal/lower and internal/native, the compiler code building the ports. Both subjects have zero exclusive Go blocks against their subsumer because all three compile the same port. Go cannot instrument TypeScript, so NODE_V8_COVERAGE was also collected in the unchanged test runs. Raw V8 profiles, filtered port profiles and the loader's transformed sources are preserved. V8 offsets refer to transformed JavaScript. Positive parent ranges with zero-count nested branches removed are unioned across each test's processes; coverage-diffs.json lists exclusive ranges with snippets.

The refusal row reaches seven exclusive port ranges, including the absent template guard and invalid choice/value paths. The broad agreement row exercises 190 valid messages and never enters that missing-message panic. R1 permits sentinel -1 to fall through to the arena lookup, changing exact error text while retaining exit 70. Only the refusal row detects it.

The gap row reaches six exclusive port ranges, including unsupported custom descriptor and Unicode fold reports, and unsupported schema reporting. The agreement corpus has no actual descriptor whose kind is unsupported. G1 makes a custom unsupported descriptor with null input return valid. The gap's self-pinned NotYet expectation alone catches it. A separate sanitized native rebuild reproduces valid; clean source Node produces the expected NotYet label.

No separate Node/native twin rows or cost rows are assigned. All four current top-level tests are included in both matrices, with no skips or timeouts. One unique mutant per subject completes the defense; additional unsuccessful attempts are not needed.
