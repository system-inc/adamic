# Regex capture checker dependency

The checker hook is pending with cohere's shim owner. This branch does not modify the cohere submodule. Its first commit, a4453390, cherry-picks library commit c5d21b44 and supplies internal/regexp.CaptureFacts. The function proves participation for numeric captures and named groups, including duplicate names across alternatives. Its fixed table covers all 88 compiler literals.

## Requested exported API

Expose these declarations through the existing checker shim:

```go
type RegExpCaptureFactsProvider func(literal *ast.Node) (
    indices []bool,
    names map[string]bool,
    known bool,
)

func (c *Checker) SetRegExpCaptureFactsProvider(
    provider RegExpCaptureFactsProvider,
) error
```

The setter is per checker, not process-global. A nil provider keeps current behavior. It returns an error if installed after expression types or semantic diagnostics have been computed. The provider receives only an original RegularExpressionLiteral node; it must not receive a runtime pattern disguised as a literal. Returned maps and slices are copied or treated as immutable. Index zero describes the whole match. A true entry proves participation on every successful match; false retains string | undefined. known=false grants no literal-specific proof. Adamic's provider splits the literal's pattern and flags, calls CaptureFacts, and returns known=false on a parser error.

## Installation and checker call sites

internal/load installs the provider on each checker after compiler.NewProgram and before loaded.diagnostics, GetSemanticDiagnostics or any type query. The same installation is needed for the Node-module program rebuilt by load. No diagnostic is filtered after the fact.

Inside cohere's checker, call the provider from Checker.checkRegularExpressionLiteral after ordinary grammar validation, before its result type is returned and cached. Preserve the literal's RegExp identity and ordinary library members. Propagated facts require proof that the pattern is unchanged; an unknown pattern, mutation or unproved alias must lose the literal-specific proof.

Checker.checkCallExpression applies those facts to the result of the actual library RegExp.exec, non-global String.match, and String.matchAll intrinsics before their result types are returned and cached. Global match has whole-match elements, not positional capture elements. matchAll specializes each yielded result, not the iterator container.

Checker.getContextualTypeForArgumentAtIndex applies the literal's capture facts to the replace/replaceAll callback argument before callback parameters are contextually checked. The whole match is string; the known capture positions use string or string | undefined; the offset is number and input is string. The groups argument has known named fields with their proven participation types. When there are no named groups that argument is absent. Explicit callback annotations must still be checked against the possible runtime arguments.

A runtime pattern has no participation proof. Result captures and named lookups remain string | undefined. Callback arity also needs care: without a proven capture count a positional argument might instead be the numeric offset or input, so the hook must not assert that every callback position is a capture. Preserve a sound argument-kind union or leave that callback surface unproved for the existing checked adapter.

## Permitted overrides

The hook may refine only the capture-bearing result and contextual callback types of genuine library intrinsics. It must not alter unrelated user methods, overload selection, argument validation, non-capture metadata or pattern grammar diagnostics. Known absent numeric captures and named lookups are undefined. For a finite literal groups shape, a name not present in the pattern is undefined; unknown runtime groups remain conservative.

Specialization narrows the loader's corrected regex library surface. Stock lib.d.ts incorrectly omits undefined from some capture element types, so narrowing the unmodified stock surface cannot itself express the conservative capture contract. Preserve a separate unmodified-library diagnostic gate: no program rejected by stock TypeScript under Adamic's options becomes accepted. Record any changed corrected-library diagnostic with its witness in the implementation commit. No diagnostics change in this dependency-only delivery.

Array mutation can invalidate capture-index guarantees even when participation is proven. Removing, replacing or shifting an element, escaping the result to an unknown mutator, or changing the RegExp pattern must retain or restore conservative types. The hook cannot permanently brand a mutable match array's indices as string merely because they were present at creation.

## Prepared checks and observations

TestRegExpStockDiagnosticBoundary uses an unmodified bundled library with Adamic's compiler options. Direct capture assignments for exec, match and matchAll each report TS2322, even for /(a)/, due to noUncheckedIndexedAccess. The loader must preserve those rejections after specialization.

TestRuntimeRegExpCaptureBoundary pins string | undefined for exec, match, matchAll and named lookups with runtime patterns. TestRegExpCaptureTypes keeps the existing unsafe capture-read refusals. These are checker-side boundary tests, not a completed hook test suite.

Focused commands, with output redirected to logs:

```
go test ./internal/load -run 'TestRegExpStockDiagnosticBoundary|TestRuntimeRegExpCaptureBoundary|TestRegExpCaptureTypes' -count=1 -v
go test ./internal/regexp -run 'TestCaptureFactsCompilerTable|TestCaptureFactsNodeBranches' -count=1 -v
go test ./internal/regexp -run 'TestCaptureFactsTSCCallbacksNode|TestCaptureFactsOptionalMutant' -count=1 -v
```

The fixed table checked 88 literal sites. Node branch checks covered 144 inputs and 117 successful matches. Callback-pattern checks covered 87 inputs and 62 successful matches. The library's optional-as-always mutant fails on Node's undefined capture observation for /(a)?(b)/ with input b. This proves the library fact catcher; it does not prove the absent checker integration. The runtime-pattern checker mutant, native and JavaScript fixtures, sanitizers and new counts remain pending. No oracle fixture is added here, so no count row changes.

Setup passed: Node 0.025s, Go 0.028s, submodules 0.075s, markdown 0.082s, clang 0.195s, build 69.172s, deferred test binaries 69.454s, cache warm 69.456s, done 69.492s. nproc is 5. This delivery records the dependency toward roadmap step 23, rulings 1 and 2; it does not claim those rulings are implemented.
