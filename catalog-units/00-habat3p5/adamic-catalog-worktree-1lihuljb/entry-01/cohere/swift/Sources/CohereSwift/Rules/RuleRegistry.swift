/*
 Every rule this binary implements. Compiled in rather than loaded, the same choice the Go engine makes:
 a rule costs nothing to run because nothing crosses a boundary to reach it.
 */
public enum RuleRegistry {
    /*
     Every file rule, built from what a run loads. Today that is only the abbreviation vocabulary, which the
     pipeline reads before anything is checked, so a vocabulary that is missing refuses the run instead of
     leaving the naming rule judging nothing.
     */
    public static func fileRules(vocabulary: AbbreviationVocabulary) -> [any FileRule] {
        [
            ForceUnwrapping(),
            ForceTry(),
            PerformanceNoIndependentAwaitInLoop(),
            ForceCast(),
            ImplicitlyUnwrappedOptional(),
            ConsistencyRequireMatchingFileName(),
            FileLength(),
            ConsistencyNoPrint(),
            ConcurrencyRequireEscapeHatchReason(),
            FatalErrorMessage(),
            Todo(),
            ConsistencyNoAbbreviatedIdentifier(vocabulary: vocabulary),
            UnownedVariableCapture(),
            PrivateOverFileprivate(),
            DuplicateImports(),
            DuplicatedKeyInDictionaryLiteral(),
            IdenticalOperands(),
            DuplicateConditions(),
            CorrectnessNoIdenticalBranches(),
            UnusedOptionalBinding(),
            UnusedClosureParameter(),
            LegacyConstructors(),
            ConsistencyNoAmbiguousIdentifier(),
            ConsistencyNoStutteringName(),
            ConsistencyNoUtilsFolder(),
            ConsistencyNoBareThrow(),
            ConsistencyNoPropertyAlias(),
            ConsistencyNoHandRolledDelay(),
            ConcurrencyNoLostUpdate(),
            ConsistencyNoBooleanOutcome(),
            CorrectnessNoUnclearedRaceTimeout(),
            SwiftFormatRule.requireLowerCamelCase,
            SwiftFormatRule.noLeadingUnderscores,
        ]
    }

    /* The rules that read what names resolve to; see `TypedFileRule` for what a run fetches for them. */
    public static let typedRules: [any TypedFileRule] = [
        EmptyCount(),
        ContainsOverFilter(),
        ContainsOverFirstNotNil(),
        SortedFirstLast(),
        FirstWhere(),
        LastWhere(),
        IsDisjoint(),
        ReduceInto(),
        CorrectnessNoDefaultForOwnedEnum(),
        UnhandledThrowingTask(),
        RedundantTypeAnnotation(),
        CorrectnessNoDiscardedTryOptional(),
        SecurityNoInterpolatedShellCommand(),
        SecurityNoInterpolatedSqlString(),
        ConcurrencyNoCheckThenWrite(),
        ConsistencyNoIsoStringDateCut(),
        CorrectnessRequireResponseStatusCheck(),
        CorrectnessNoWriteOnlyCollection(),
    ]

    public static let packageRules: [any PackageRule] = [
        ToolchainRequireSwiftSixLanguageMode(),
        ToolchainRequireUpcomingFeatures(),
        ToolchainRequireStrictMemorySafety(),
    ]

    /* Sorted, so two binaries' lists can be compared with `diff`. */
    public static var allNames: [String] {
        /* Names only, so no vocabulary is read: `--rules` answers without one. */
        (fileRules(vocabulary: AbbreviationVocabulary()).map(\.name) + typedRules.map(\.name) + packageRules.map(\.name))
            .sorted()
    }
}
