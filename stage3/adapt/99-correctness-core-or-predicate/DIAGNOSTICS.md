# Separate stock-program diagnostics

The initial 711-file inventory reported 25 diagnostics. Two were TS2307: missing `diff` declarations at src/harness/harnessIO.ts:1 and src/testRunner/unittests/skipJSDocParsing.ts:1. That inventory supplied external typeRoots but had no node_modules link at the tree root. Linking the pinned upstream dependencies removed those two diagnostics.

The refreshed baseline and candidate inventories each report the same remaining 23 diagnostics. Their messages and unchanged source lines are compared separately from the adapted tree's own build. Both own `npm run build` commands exit 0, including compiler and test typechecks; no own-build diagnostic is added.

| File | Baseline line | Code | Message |
| --- | ---: | ---: | --- |
| src/compiler/checker.ts | 17134 | TS7023 | 'getUnresolvedSymbolForEntityName' implicitly has return type 'any' because it does not have a return type annotation and is referenced directly or indirectly in one of its return expressions. |
| src/compiler/checker.ts | 17140 | TS7022 | 'parentSymbol' implicitly has type 'any' because it does not have a type annotation and is referenced directly or indirectly in its own initializer. |
| src/compiler/checker.ts | 17143 | TS7022 | 'path' implicitly has type 'any' because it does not have a type annotation and is referenced directly or indirectly in its own initializer. |
| src/compiler/checker.ts | 17144 | TS7022 | 'result' implicitly has type 'any' because it does not have a type annotation and is referenced directly or indirectly in its own initializer. |
| src/compiler/checker.ts | 17392 | TS2345 | Argument of type 'Symbol \| undefined' is not assignable to parameter of type 'Symbol'.   Type 'undefined' is not assignable to type 'Symbol'. |
| src/compiler/checker.ts | 17397 | TS2345 | Argument of type 'Symbol \| undefined' is not assignable to parameter of type 'Symbol'.   Type 'undefined' is not assignable to type 'Symbol'. |
| src/compiler/path.ts | 693 | TS7022 | 'lastSlash' implicitly has type 'any' because it does not have a type annotation and is referenced directly or indirectly in its own initializer. |
| src/compiler/resolutionCache.ts | 1478 | TS2349 | This expression is not callable.   Each member of the union type '{ (callbackfn: (value: ResolutionWithFailedLookupLocations, value2: ResolutionWithFailedLookupLocations, set: Set<...>) => void, thisArg?: any): void; (callbackfn: (value: ResolutionWithFailedLookupLocations, value2: ResolutionWithFailedLookupLocations, set: Set<...>) => void, thisArg?: any): void; } \| { ...; }' has signatures, but none of those signatures are compatible with each other. |
| src/compiler/resolutionCache.ts | 1478 | TS7006 | Parameter 'resolution' implicitly has an 'any' type. |
| src/compiler/utilities.ts | 8255 | TS2349 | This expression is not callable.   Each member of the union type '{ (callbackfn: (value: U, key: K, map: ReadonlyMap<K, U>) => void, thisArg?: any): void; (callbackfn: (value: U, key: K, map: ReadonlyMap<K, U>) => void, thisArg?: any): void; } \| { ...; }' has signatures, but none of those signatures are compatible with each other. |
| src/compiler/utilities.ts | 8255 | TS7006 | Parameter 'valueInNewMap' implicitly has an 'any' type. |
| src/compiler/utilities.ts | 8255 | TS7006 | Parameter 'key' implicitly has an 'any' type. |
| src/harness/projectServiceStateLogger.ts | 253 | TS2349 | This expression is not callable.   Each member of the union type '{ (callbackfn: (value: T, key: any, map: Map<any, T>) => void, thisArg?: any): void; (callbackfn: (value: T, key: any, map: Map<any, T>) => void, thisArg?: any): void; } \| { ...; } \| { ...; }' has signatures, but none of those signatures are compatible with each other. |
| src/harness/projectServiceStateLogger.ts | 363 | TS2349 | This expression is not callable.   Each member of the union type '{ (callbackfn: (value: any, key: string, map: Map<string, any>) => void, thisArg?: any): void; (callbackfn: (value: any, key: string, map: Map<string, any>) => void, thisArg?: any): void; } \| { ...; }' has signatures, but none of those signatures are compatible with each other. |
| src/harness/projectServiceStateLogger.ts | 363 | TS7006 | Parameter 'p' implicitly has an 'any' type. |
| src/harness/projectServiceStateLogger.ts | 363 | TS7006 | Parameter 'key' implicitly has an 'any' type. |
| src/harness/projectServiceStateLogger.ts | 373 | TS2349 | This expression is not callable.   Each member of the union type '{ (callbackfn: (value: any, key: string, map: Map<string, any>) => void, thisArg?: any): void; (callbackfn: (value: any, key: string, map: Map<string, any>) => void, thisArg?: any): void; } \| { ...; }' has signatures, but none of those signatures are compatible with each other. |
| src/harness/projectServiceStateLogger.ts | 373 | TS7006 | Parameter 'p' implicitly has an 'any' type. |
| src/harness/projectServiceStateLogger.ts | 373 | TS7006 | Parameter 'key' implicitly has an 'any' type. |
| src/server/editorServices.ts | 4907 | TS2349 | This expression is not callable.   Each member of the union type '{ (callbackfn: (value: ConfiguredProject, value2: ConfiguredProject, set: Set<ConfiguredProject>) => void, thisArg?: any): void; (callbackfn: (value: ConfiguredProject, value2: ConfiguredProject, set: Set<...>) => void, thisArg?: any): void; } \| { ...; }' has signatures, but none of those signatures are compatible with each other. |
| src/server/editorServices.ts | 4907 | TS7006 | Parameter '_' implicitly has an 'any' type. |
| src/server/editorServices.ts | 4907 | TS7006 | Parameter 'project' implicitly has an 'any' type. |
| src/services/codefixes/requireInTs.ts | 93 | TS2322 | Type 'Expression' is not assignable to type 'StringLiteralLike'.   Type 'Expression' is missing the following properties from type 'NoSubstitutionTemplateLiteral': _literalExpressionBrand, text, _primaryExpressionBrand, _memberExpressionBrand, and 5 more. |

The original 25 messages remain in calls-before.json and the baseline evidence archive. Detailed refreshed before/after records, including columns, are retained in the proof archive. No unrelated diagnostic was repaired by this adaptation.
