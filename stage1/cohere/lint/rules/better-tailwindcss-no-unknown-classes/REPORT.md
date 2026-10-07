Status: .a decision/reporting core implemented and registered; default listener explicitly refuses the missing live design-system/program adapter.

The next paragraph records the earlier dependency-only investigation. Current results supersede it and are linked below.

Observed dependencies, beyond the shared harness: Go no_unknown_classes.go declares ReadsCompilerOptions | ReadsDesignSystem and calls DesignSystemForProgram before building listeners. class_existence.go asks ParseCandidate followed by ClassValueResolvesIn against LoadedDesignSystem. No implementations of those engine contracts, ClassLiteralReader or classTokens were found under stage1/cohere. The corresponding Go candidate.go, design_system.go and value_parser.go alone contain 2,167 lines; generated class-name tables are not a substitute for project-specific @utility and variant resolution. RuleContext supplies source/parser/scanner/settings, but no program, compiler options, project root or recorded stylesheet filesystem.

Inference: a faithful registered rule cannot resolve two projects' distinct stylesheet utilities through the current context. Omitting that input or checking a shipped list would silently report the wrong findings. Adding a program/design-system contract to shared files is outside this worker's authorized territory. The absent engine implementation is a dependency gap beyond the automatic-fix serializer gap. Following Ahra's instruction, "If anything else blocks you, say exactly what it is and stop, rather than editing shared files", this worker stopped at this investigation and left all three claims reserved. No new rule code, registration, parity certificate, mutant result or findings/s is claimed.

External baseline command, with output redirected to baseline.log:

    cd cohere
    go test ./internal/lint/rules/tailwind -run 'Test(NoDeprecatedClasses|NoDuplicateClasses|NoUnknownClasses|UnknownIgnore|ExistenceComesFromTheRepository|KnownRootWithUnknownValue|UnknownClassFixturesActuallyRan)' -count=1 -timeout 10m -v

The deprecated and duplicate tests pass. The aggregate exits 1: unknown-class fixtures skip because the test's hardcoded /Users/kirkouimet/Projects/ahra/app/_theme/styles cannot locate installed tailwindcss. TestUnknownClassFixturesActuallyRan deliberately fails to reject a false passing result. This is external baseline coverage only, not a comparison of Adamic implementations.

The dependency-only investigation above is historical. Current implementation, successful bounded comparisons, semantic mutants, exact limitations and retained logs are in ../better-tailwindcss-no-deprecated-classes/testdata/evidence/REPORT.md. It supersedes the earlier claim that no implementation or registration exists.
