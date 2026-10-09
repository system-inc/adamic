/*
 Generated from cohere's policy/messages/ by `go run ./policy/tools/generate`. Do not edit it here:
 edit the message files and run that again. While the two differ, a test in each engine fails.
 */
enum RuleMessages {
    /* A finding's message: the id the catalog files it under, and its text. */
    struct Message: Equatable, Sendable {
        let id: String
        let text: String
    }

    /* SHA-256 of the messages that reach Swift, resolved, which a test recomputes from policy/messages/ on disk. */
    static let sourceDigest = "61435a5fb4d922e644b6286e3835d882b44ed395354cb7849537c58cd06a6f4a"

    /* Every message here, as `Rule.id`, for the test that fails on one no rule renders. */
    static let all = [
        "AlwaysUseLowerCamelCase.alwaysUseLowerCamelCase",
        "AlwaysUseLowerCamelCase.incumbentFailed",
        "ConcurrencyNoCheckThenWrite.checkThenWrite",
        "ConcurrencyNoLostUpdate.lostUpdate",
        "ConcurrencyRequireEscapeHatchReason.escapeHatchWithoutReason",
        "ConsistencyNoAbbreviatedIdentifier.abbreviatedIdentifier",
        "ConsistencyNoAbbreviatedIdentifier.abbreviatedSuffix",
        "ConsistencyNoAbbreviatedIdentifier.abbreviatedWordSegment",
        "ConsistencyNoAbbreviatedIdentifier.millisecondSuffix",
        "ConsistencyNoAmbiguousIdentifier.noAmbiguousE",
        "ConsistencyNoAmbiguousIdentifier.noSingleLetter",
        "ConsistencyNoBareThrow.bareThrow",
        "ConsistencyNoBooleanOutcome.booleanOutcome",
        "ConsistencyNoHandRolledDelay.handRolledDelay",
        "ConsistencyNoHandRolledDelay.handRolledDelayDoCatch",
        "ConsistencyNoIsoStringDateCut.isoDateWithoutTimeZone",
        "ConsistencyNoIsoStringDateCut.isoStringCutToDate",
        "ConsistencyNoPrint.consistencyNoConsole",
        "ConsistencyNoPropertyAlias.noPropertyAlias",
        "ConsistencyNoStutteringName.stutteringName",
        "ConsistencyNoUtilsFolder.noUnderscoreUtils",
        "ConsistencyNoUtilsFolder.noUtils",
        "ConsistencyRequireMatchingFileName.extensionOutsideItsFile",
        "ConsistencyRequireMatchingFileName.requireMatchingFileName",
        "CorrectnessNoDefaultForOwnedEnum.noDefaultForOwnedEnum",
        "CorrectnessNoDiscardedTryOptional.discardedTryOptional",
        "CorrectnessNoDiscardedTryOptional.discardedTryOptionalInClosure",
        "CorrectnessNoIdenticalBranches.identicalBranches",
        "CorrectnessNoUnclearedRaceTimeout.unclearedRaceTimeout",
        "CorrectnessNoWriteOnlyCollection.writeOnlyCollection",
        "CorrectnessRequireResponseStatusCheck.bodyReadWithoutStatusCheck",
        "CorrectnessRequireResponseStatusCheck.responseDiscarded",
        "ForceCast.forceCast",
        "PerformanceNoIndependentAwaitInLoop.independentAwaitInLoop",
        "SecurityNoInterpolatedShellCommand.interpolatedShellCommand",
        "SecurityNoInterpolatedSqlString.interpolatedSqlString",
        "ToolchainRequireStrictMemorySafety.strictMemorySafetyMissing",
        "ToolchainRequireSwift6LanguageMode.languageModeBelowSix",
        "ToolchainRequireUpcomingFeatures.upcomingFeatureMissing",
    ]

    /* cohere-swift/always-use-lower-camel-case */
    enum AlwaysUseLowerCamelCase {
        static func alwaysUseLowerCamelCase(swiftFormatFinding: String) -> Message {
            Message(
                id: "alwaysUseLowerCamelCase",
                text:
                    #"\#(swiftFormatFinding). Swift spells values in lowerCamelCase and types in UpperCamelCase, so a reader tells which is which at a glance, and an underscore inside a name is a word boundary camel case already marks."#,
            )
        }

        static func incumbentFailed(incumbentRule: String, error: String) -> Message {
            Message(
                id: "incumbentFailed",
                text:
                    #"swift-format's \#(incumbentRule) could not run on this file, so it was not checked: \#(error)"#,
            )
        }
    }

    /* cohere-swift/concurrency-no-check-then-write */
    enum ConcurrencyNoCheckThenWrite {
        static func checkThenWrite(write: String) -> Message {
            Message(
                id: "checkThenWrite",
                text:
                    #"This loop looks for a file name that is free, and the `\#(write)` after it creates that file in a separate step, so two saves running at once (in this process or in two) can both find the same name free and the second silently overwrites the first. Claim the name in the step that creates the file: `data.write(to: url, options: .withoutOverwriting)` (or `FileManager.moveItem` from a temporary file) fails with `CocoaError.fileWriteFileExists` when the name is taken, so try the next name on that error instead of checking first."#,
            )
        }
    }

    /* cohere-swift/concurrency-no-lost-update */
    enum ConcurrencyNoLostUpdate {
        enum SuspendedRead {
            case variable(target: String)
        }

        static func lostUpdate(suspendedRead: SuspendedRead) -> Message {
            let suspendedReadText =
                switch suspendedRead {
                    case .variable(let target):
                        #"`\#(target)` as it was before an await, and the function is suspended in between, so anything else that changes `\#(target)` during the suspension (another call on this actor, another task on the main actor) is overwritten by this line. Read `\#(target)` after the suspension, or keep the whole read, compute and write on one side of it."#
                }
            return Message(
                id: "lostUpdate",
                text:
                    #"This write loses updates. Its new value is computed from \#(suspendedReadText)"#,
            )
        }
    }

    /* cohere-swift/concurrency-require-escape-hatch-reason */
    enum ConcurrencyRequireEscapeHatchReason {
        static func escapeHatchWithoutReason(spelling: String) -> Message {
            Message(
                id: "escapeHatchWithoutReason",
                text:
                    #"\#(spelling) tells the compiler to trust this code instead of checking it. Say why that is safe in a comment directly above the declaration, so the next reader can check the reasoning the compiler no longer does."#,
            )
        }
    }

    /* cohere-swift/consistency-no-abbreviated-identifier */
    enum ConsistencyNoAbbreviatedIdentifier {
        static func abbreviatedIdentifier(name: String, advice: String) -> Message {
            Message(
                id: "abbreviatedIdentifier",
                text:
                    #"Identifier "\#(name)" should not be abbreviated. \#(advice) A name is written once and read everywhere, so the letters saved at the declaration are paid back at every call site by a reader who has to expand the abbreviation themselves and hope they expanded it the way the author meant."#,
            )
        }

        static func abbreviatedSuffix(name: String, suffix: String, advice: String) -> Message {
            Message(
                id: "abbreviatedSuffix",
                text:
                    #"Identifier "\#(name)" should not end with "\#(suffix)". \#(advice) A name is written once and read everywhere, so the letters saved at the declaration are paid back at every call site by a reader who has to expand the abbreviation themselves and hope they expanded it the way the author meant."#,
            )
        }

        static func abbreviatedWordSegment(name: String, word: String, suggestion: String) -> Message {
            Message(
                id: "abbreviatedWordSegment",
                text:
                    #"Identifier "\#(name)" abbreviates "\#(word)". Use "\#(suggestion)". A name is written once and read everywhere, so the letters saved at the declaration are paid back at every call site by a reader who has to expand the abbreviation themselves and hope they expanded it the way the author meant."#,
            )
        }

        static func millisecondSuffix(name: String, suggestion: String) -> Message {
            Message(
                id: "millisecondSuffix",
                text:
                    #"Identifier "\#(name)" should not abbreviate milliseconds as "Ms". Use "\#(suggestion)", which is how the rest of the tree spells a millisecond value: "durationInMilliseconds" outnumbers "durationMs" more than two to one for the identical value, so the rename follows what the codebase already decided rather than introducing a third spelling."#,
            )
        }
    }

    /* cohere-swift/consistency-no-ambiguous-identifier */
    enum ConsistencyNoAmbiguousIdentifier {
        enum Context {
            case error
            case event
            case unclear
        }

        static func noAmbiguousE(context: Context, suggestedName: String) -> Message {
            let contextText =
                switch context {
                    case .error:
                        #" (appears to be an error)"#
                    case .event:
                        #" (appears to be an event)"#
                    case .unclear:
                        #" (context unclear)"#
                }
            return Message(
                id: "noAmbiguousE",
                text:
                    #"Variable named "e" is too ambiguous\#(contextText). It is the one name that could be an error or an event, and a reader has to find the declaration to learn which. Use "\#(suggestedName)" or a more descriptive name."#,
            )
        }

        static func noSingleLetter(name: String) -> Message {
            Message(
                id: "noSingleLetter",
                text:
                    #"Single-letter identifier "\#(name)" is not descriptive enough. The name is read everywhere it is used and declared only once, so the saving is at the declaration and the cost is at every call site."#,
            )
        }
    }

    /* cohere-swift/consistency-no-bare-throw */
    enum ConsistencyNoBareThrow {
        static func bareThrow(domain: String) -> Message {
            Message(
                id: "bareThrow",
                text:
                    #"This throws an NSError made up here, with the domain \#(domain), which names no declared failure. Declare the failure as a case of an error type of our own (an enum conforming to Error, and LocalizedError for its text) and throw that case, so a caller catches it by case. Until then a caller can match it only by repeating the domain string and the code, and nothing keeps the two in step."#,
            )
        }
    }

    /* cohere-swift/consistency-no-boolean-outcome */
    enum ConsistencyNoBooleanOutcome {
        static func booleanOutcome(flagName: String, declaration: String, suggested: String) -> Message {
            Message(
                id: "booleanOutcome",
                text:
                    #"\#(flagName): Bool on \#(declaration) collapses every outcome into one bit, at the moment the distinction is cheapest to keep. It also leaves the reader to know which other fields hold for which value of the flag. Return a named outcome instead: an enum with a case for each way the operation can turn out and the payload on the case that carries it, or Result, or a throw for the failure. Suggested name: \#(suggested)Outcome."#,
            )
        }
    }

    /* cohere-swift/consistency-no-hand-rolled-delay */
    enum ConsistencyNoHandRolledDelay {
        static func handRolledDelay() -> Message {
            Message(
                id: "handRolledDelay",
                text:
                    #"This try? await Task.sleep is a hand-rolled Task.sleepUnlessCancelled. Task.sleep throws only CancellationError, so the try? does nothing but let the pause end early when the task is cancelled, which is all the primitive does. Write await Task.sleepUnlessCancelled with the same duration (declared once per project, in Task+SleepUnlessCancelled.swift, taking for: or nanoseconds:), so the pause reads as a plain pause and the one place that says why cancellation is let go says it for every call. A do/catch with an empty catch around the sleep is the same pause spelled longer, not the repair."#,
            )
        }

        static func handRolledDelayDoCatch() -> Message {
            Message(
                id: "handRolledDelayDoCatch",
                text:
                    #"This do/catch around Task.sleep, with nothing in the catch, is the house's Task.sleepUnlessCancelled written out by hand: Task.sleep throws only CancellationError, so the catch does nothing but let the pause end early when the task is cancelled, which is all the primitive does. Write await Task.sleepUnlessCancelled with the same duration (declared once per project, in Task+SleepUnlessCancelled.swift, taking for: or nanoseconds:), so the pause reads as a plain pause and the one place that says why cancellation is let go says it for every call."#,
            )
        }
    }

    /* cohere-swift/consistency-no-iso-string-date-cut */
    enum ConsistencyNoIsoStringDateCut {
        enum Renderer {
            case dateFormatter
            case formatStyle
        }

        static func isoDateWithoutTimeZone(renderer: Renderer) -> Message {
            let rendererText =
                switch renderer {
                    case .dateFormatter:
                        #"ISO8601DateFormatter is UTC unless its timeZone is set, so this is the UTC day whether or not that was the day meant, and in Utah the UTC day turns over at 5 pm in winter and 6 pm in summer: an evening run gets tomorrow. Say the zone: set the formatter's timeZone to .current when the local day is meant, or to .gmt when the UTC day is."#
                    case .formatStyle:
                        #"Date.ISO8601FormatStyle is UTC unless it is given one, so this is the UTC day whether or not that was the day meant, and in Utah the UTC day turns over at 5 pm in winter and 6 pm in summer: an evening run gets tomorrow. Say the zone: Date.ISO8601FormatStyle(timeZone: .current) when the local day is meant, or timeZone: .gmt when the UTC day is."#
                }
            return Message(
                id: "isoDateWithoutTimeZone",
                text:
                    #"This formats a date as its ISO 8601 calendar day without naming a time zone. \#(rendererText)"#,
            )
        }

        static func isoStringCutToDate() -> Message {
            Message(
                id: "isoStringCutToDate",
                text:
                    #"This cuts an ISO 8601 timestamp down to its date part. ISO 8601 formatting is UTC unless it is given a time zone, so the cut is the UTC calendar day whether or not that was the day meant, and in Utah the UTC day turns over at 5 pm in winter and 6 pm in summer: an evening run gets tomorrow. Format the day itself and say the zone: Date.ISO8601FormatStyle(timeZone: .current).year().month().day() when the local day is meant, or timeZone: .gmt when the UTC day is."#,
            )
        }
    }

    /* cohere-swift/consistency-no-print */
    enum ConsistencyNoPrint {
        static func consistencyNoConsole(kind: String) -> Message {
            Message(
                id: "consistencyNoConsole",
                text:
                    #"Do not call 'print'. In \#(kind) it writes to a stdout nobody reads. Log through os.Logger, so the message lands in the unified log with a subsystem, a category and a level."#,
            )
        }
    }

    /* cohere-swift/consistency-no-property-alias */
    enum ConsistencyNoPropertyAlias {
        static func noPropertyAlias(name: String, reach: String) -> Message {
            Message(
                id: "noPropertyAlias",
                text:
                    #"`\#(name)` is a pure alias for `\#(reach)`. It is read once, and nothing runs between its declaration and that read. Write `\#(reach)` where `\#(name)` is read and delete the declaration, so the value keeps the object it came from and a naked local still means this scope made it."#,
            )
        }
    }

    /* cohere-swift/consistency-no-stuttering-name */
    enum ConsistencyNoStutteringName {
        static func stutteringName(name: String) -> Message {
            Message(
                id: "stutteringName",
                text:
                    #""\#(name).\#(name)" stutters, which means the name is carrying nothing: it repeats the field instead of saying which \#(name) this is. Rename the value for what it holds, the type it came back as or whatever distinguishes it from another \#(name) in this scope, so a reader forty lines down does not have to find the declaration."#,
            )
        }
    }

    /* cohere-swift/consistency-no-utils-folder */
    enum ConsistencyNoUtilsFolder {
        static func noUnderscoreUtils(folder: String, replacement: String) -> Message {
            Message(
                id: "noUnderscoreUtils",
                text:
                    #"Folder name "\#(folder)" is not allowed. Use "\#(replacement)" instead."#,
            )
        }

        static func noUtils(folder: String, replacement: String) -> Message {
            Message(
                id: "noUtils",
                text:
                    #"Folder name "\#(folder)" is not allowed. Use "\#(replacement)" instead."#,
            )
        }
    }

    /* cohere-swift/consistency-require-matching-file-name */
    enum ConsistencyRequireMatchingFileName {
        enum Placement {
            case besideType(extendedName: String, typeName: String, suggestion: String)
            case withoutType(fileName: String, extendedName: String, suggestion: String)
        }

        static func extensionOutsideItsFile(placement: Placement) -> Message {
            let placementText =
                switch placement {
                    case .besideType(let extendedName, let typeName, let suggestion):
                        #"An extension of \#(extendedName) in the file for \#(typeName). Extensions of another type belong in that type's own file, such as \#(suggestion), so everything added to \#(extendedName) is found together."#
                    case .withoutType(let fileName, let extendedName, let suggestion):
                        #"This file is \#(fileName) and extends \#(extendedName). Name it \#(suggestion) after what it adds, so everything added to \#(extendedName) is found together."#
                }
            return Message(
                id: "extensionOutsideItsFile",
                text:
                    #"\#(placementText)"#,
            )
        }

        enum Alternative {
            case none
            case renameComponent(fileBaseName: String)
        }

        static func requireMatchingFileName(typeName: String, fileName: String, alternative: Alternative) -> Message {
            let alternativeText =
                switch alternative {
                    case .none:
                        #""#
                    case .renameComponent(let fileBaseName):
                        #", or rename the component to `\#(fileBaseName)` if the file's name is the right one"#
                }
            return Message(
                id: "requireMatchingFileName",
                text:
                    #"`\#(typeName)` is the first type declared in `\#(fileName)`, and the file is not named for it. Name it `\#(typeName).swift`, so a reader looking for `\#(typeName)` opens it without a search\#(alternativeText)."#,
            )
        }
    }

    /* cohere-swift/correctness-no-default-for-owned-enum */
    enum CorrectnessNoDefaultForOwnedEnum {
        static func noDefaultForOwnedEnum(enumName: String) -> Message {
            Message(
                id: "noDefaultForOwnedEnum",
                text:
                    #"This default answers for every case of \#(enumName), including any added later, so the compiler can no longer say this switch does not handle a new one. List the remaining cases instead."#,
            )
        }
    }

    /* cohere-swift/correctness-no-discarded-try-optional */
    enum CorrectnessNoDiscardedTryOptional {
        static func discardedTryOptional() -> Message {
            Message(
                id: "discardedTryOptional",
                text:
                    #"This try? throws the error away and keeps nothing, so a failure here is invisible. Use do/catch, and say in the catch why the failure can be ignored if it can."#,
            )
        }

        static func discardedTryOptionalInClosure() -> Message {
            Message(
                id: "discardedTryOptionalInClosure",
                text:
                    #"This try? is the whole body of a closure whose result goes nowhere (it returns nothing, or it is the value of a task nobody keeps), so the error is thrown away unseen. Use do/catch inside the closure, and say in the catch why the failure can be ignored if it can."#,
            )
        }
    }

    /* cohere-swift/correctness-no-identical-branches */
    enum CorrectnessNoIdenticalBranches {
        static func identicalBranches() -> Message {
            Message(
                id: "identicalBranches",
                text:
                    #"Every branch of this conditional does the same thing, so its condition chooses nothing: the code runs the same way whether it is true or false. Usually one branch was meant to differ and a copy was never edited (highQuality ? "pro" : "pro"). Write the branch that was meant, or, if both really are the same, drop the conditional and keep one copy."#,
            )
        }
    }

    /* cohere-swift/correctness-no-uncleared-race-timeout */
    enum CorrectnessNoUnclearedRaceTimeout {
        static func unclearedRaceTimeout(group: String) -> Message {
            Message(
                id: "unclearedRaceTimeout",
                text:
                    #"This child only sleeps and then gives up, racing the group's other work for the first result, but nothing cancels the group once that result is in. A task group waits for every child before it returns, so when the work wins the caller still sits out the whole timeout, and a timeout that returns rather than throws bounds nothing, since the group then waits for the work too. Call \#(group).cancelAll() once the first result is in (a defer { \#(group).cancelAll() } at the top of the group's body covers every path), so the losing child is cancelled and its sleep ends at once."#,
            )
        }
    }

    /* cohere-swift/correctness-no-write-only-collection */
    enum CorrectnessNoWriteOnlyCollection {
        static func writeOnlyCollection() -> Message {
            Message(
                id: "writeOnlyCollection",
                text:
                    #"This collection is only ever added to: every use of it puts something in or takes something out, and nothing reads what it holds. The work that fills it is thrown away. Either the code that was meant to read it is missing (a value kept for the result and never copied in), or the collection is dead and should be deleted along with its writes."#,
            )
        }
    }

    /* cohere-swift/correctness-require-response-status-check */
    enum CorrectnessRequireResponseStatusCheck {
        static func bodyReadWithoutStatusCheck() -> Message {
            Message(
                id: "bodyReadWithoutStatusCheck",
                text:
                    #"This reads the body of the response URLSession returned on a path where its statusCode is never read. URLSession does not throw on an HTTP error, so a 401 or 500 error page arrives here as data and is decoded, returned or kept as if the request had succeeded. Cast the response to HTTPURLResponse and check statusCode before trusting the body, and throw or return the failure when it is not a success."#,
            )
        }

        static func responseDiscarded() -> Message {
            Message(
                id: "responseDiscarded",
                text:
                    #"This takes the response URLSession returned and drops it. Its status goes with it. URLSession does not throw on an HTTP error, so a request that fails with a 400 or 500 runs on as if it had succeeded. Keep the response, cast it to HTTPURLResponse and check statusCode, and throw or return the failure when it is not a success."#,
            )
        }
    }

    /* cohere-swift/force-cast */
    enum ForceCast {
        static func forceCast() -> Message {
            Message(
                id: "forceCast",
                text:
                    #"as! crashes the process when the value is not that type. Use as? and say what happens when it is not."#,
            )
        }
    }

    /* cohere-swift/performance-no-independent-await-in-loop */
    enum PerformanceNoIndependentAwaitInLoop {
        static func independentAwaitInLoop() -> Message {
            Message(
                id: "independentAwaitInLoop",
                text:
                    #"This loop awaits once per item, so the items run one after another, yet nothing carries from one iteration to the next: each await is the item's own work, no local of the function is reassigned and no shared state is written, nothing leaves the loop on an awaited result or a thrown error, and nothing prints, logs or sleeps between iterations. Start them together with `withTaskGroup` (collect each result with its index when the order matters, since a group hands results back as they finish), with `async let` when the items are a fixed few, or with a bounded group when the collection can be large or the far side rate-limits."#,
            )
        }
    }

    /* cohere-swift/security-no-interpolated-shell-command */
    enum SecurityNoInterpolatedShellCommand {
        static func interpolatedShellCommand() -> Message {
            Message(
                id: "interpolatedShellCommand",
                text:
                    #"This command string is run by a shell, and a value built into it is not a literal, so the shell parses that value as code: a `"`, `'` or space in it breaks the command, and a `$(...)` or backtick in it runs. Quoting it by hand is one escape away from the same bug. Keep the script literal and pass the value after it as its own argument, which the shell reads as `$1` and never parses (`["-c", "cd \"$1\" && make", "zsh", path]`), or run the program directly with no shell (`currentDirectoryURL` in place of a `cd`)."#,
            )
        }
    }

    /* cohere-swift/security-no-interpolated-sql-string */
    enum SecurityNoInterpolatedSqlString {
        static func interpolatedSqlString() -> Message {
            Message(
                id: "interpolatedSqlString",
                text:
                    #"This value is written between the single quotes of a SQL string, so a quote in it ends the string early: the query breaks on ordinary input like O'Brien, and a crafted value runs as SQL. Bind it as a parameter instead (WHERE name = ? with the value in the parameters beside the statement). Where the query cannot take parameters, give the value a type that cannot hold a quote (an integer, a closed enum's case), or escape it in place by doubling backslashes and quotes: .replacingOccurrences(of: "\\", with: "\\\\").replacingOccurrences(of: "'", with: "''")."#,
            )
        }
    }

    /* cohere-swift/toolchain-require-strict-memory-safety */
    enum ToolchainRequireStrictMemorySafety {
        static func strictMemorySafetyMissing(target: String) -> Message {
            Message(
                id: "strictMemorySafetyMissing",
                text:
                    #"\#(target) does not compile under strict memory safety. Add .strictMemorySafety() to its swiftSettings, so every unsafe construct is visible where it is used; then replace each flagged site with a safe API where one exists, and mark unsafe only what must be."#,
            )
        }
    }

    /* cohere-swift/toolchain-require-swift-6-language-mode */
    enum ToolchainRequireSwift6LanguageMode {
        static func languageModeBelowSix(target: String, languageMode: String) -> Message {
            Message(
                id: "languageModeBelowSix",
                text:
                    #"\#(target) compiles in Swift \#(languageMode) language mode. Swift 6 is required: it makes data-race safety a compile error instead of a warning nobody reads. Set .swiftLanguageMode(.v6) on the target, or remove the setting that lowers it."#,
            )
        }
    }

    /* cohere-swift/toolchain-require-upcoming-features */
    enum ToolchainRequireUpcomingFeatures {
        static func upcomingFeatureMissing(
            target: String,
            features: String,
            settings: String,
            reasons: String,
        ) -> Message {
            Message(
                id: "upcomingFeatureMissing",
                text:
                    #"\#(target) does not enable \#(features). Add \#(settings) to its swiftSettings. \#(reasons)."#,
            )
        }
    }
}
