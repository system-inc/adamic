Built: native constructed-context source candidate; complete parity is blocked by an unchanged Go panic.
Commits: candidate follows d2e8b4e753f9ee36d96c1ae2701ba7533dd6b92c on codex/typeaware-wave-04.
Checks: 109/110 individual Go matches, 110 sanitizer executions, both frozen corpora equal, 4,287 Unicode/quote cases equal.
Mutants: span, depth, escape, primitive classification, uppercase and DEL quoting were caught by byte comparisons; released handle exits 70.
Uncovered: full source oracle fails on satisfies; shared checker/factory integration and full repository gate remain unfinished. No new claims.

The source visitor implements provider/value extraction, component recognition, construction tracking, render scope, memo dependencies, bounded callee return analysis and escape tracking. Native Adamic files use .a. Listener kinds are JsxOpeningElement and JsxSelfClosingElement; the visitor consumes the handed node. Existing bridge questions suffice. No shared harness, generator, protected compiler file or upstream rule was edited.

The full validator failed, and remains failed, on this valid TypeScript input:

```tsx
declare const Ctx:any;
function Component(){return <Ctx.Provider value={({} satisfies object)}/>;}
```

Production cohere/internal/lint/rules/react/jsx_no_constructed_context_values.go:484 groups AsExpression with SatisfiesExpression and calls AsAsExpression. Go exits 2 with `interface conversion: ast.nodeData is *ast.SatisfiesExpression, not *ast.AsExpression`. Native exits zero under sanitizers. The reproducer is retained in jsx_no_constructed_context_values/testdata/satisfies.tsx.txt. The aggregate comparison and both triage validators deliberately retain nonzero overall status. No input was removed to obtain green. This is not an HIR parking claim.

Commands, with output redirected to scratch log files:

- validate_source.py --normal-only: aggregate Go controls fail with exit 2.
- validate_unaffected.py: all 110 inputs attempted using the same full program roots, 109 complete wire matches; control-035.tsx is the sole Go panic. Overall exit 1.
- validate_remaining.py: every input ran with ASan/UBSan/leak checks; 109 matched Go, the panic witness received memory-only validation. Both frozen corpora matched complete output (compiler 6088 bytes, repository 18485 bytes, zero findings). All six message IDs occur among matching controls, including zero fixes and suggestions as in Go. Four independently compiling, sanitizer-clean source mutants were caught solely by comparison: span on control 0, depth on 83, escape on 88 and primitive classification on 86. Released program is refused with exit 70. Overall exit 1 retains the upstream blocker.
- validate_unicode.py: Go Unicode 15.0.0 IsUpper/IsPrint tables translated into real RegExp literals and strconv.Quote behavior matched across Go, native, sanitized native, source Node and emitted JavaScript: 4287 inputs, 77340 bytes. Uppercase and DEL-quote mutants caught.
- go test ./stage1/cohere/typeaware -run '^(TestPinnedTypeFlags|TestSixPinnedFlags)$' -count=1: PASS, 0.004s. An earlier incorrect filter ran no tests; it is not counted as validation.
- go list ./...: PASS. Standalone owned Go oracle artifacts now carry build-ignore tags; explicit overlay builds for fragments and undefined-name still pass. No correctness test is excluded by these tags.

Observed corpus wall times, isolated invocations rather than a controlled benchmark: compiler native 1.730993s versus Go 0.293625s; repository native 0.277340s versus Go 0.131411s. Sanitized native: 5.992570s and 0.990313s respectively. Tool setup from the restored landing: readiness 0s, warm 124s, total 124s; nproc 5.

The last pushed landing includes main c7991b900362796aefd111474e65eb5398e91953 and lint area b84a9d9314b65d3d0261ee017e233287b4f071da, with eight previous source-rule oracles green (LANDING_C799.md). Initial credential verification failed, but retry succeeded and the candidate was pushed at da9f8729102b2e714ce0ff64543bc46df467efa9. The area then advanced to b46914832d70e00847d82d5d221ab7bb24040c53; the owned branch was rebased with merge topology preserved. Its integration diff is confined to stage1/cohere/lint, leaving every type-aware/compiler/oracle input unchanged. Freshly rebuilt candidate comparisons still return overall exit 1 with 109/110 matches; pinned flag tests pass in 0.003s. The eight earlier rule suites retain LANDING_C799 evidence; they were not all rerun after this latest unrelated-area rebase. The branch is not claimed fully green or landing-ready while the Go panic persists. No push to main or an area branch is attempted.

Full native source execution has not been checked against emitted JavaScript; five-backend comparison applies to the Unicode/quote helpers. Standard shared RuleContext still lacks the program/checker access needed for factory integration. Virtual default-library foreign declaration paths are not covered. The full gate and its 17 required-input checks were not run in this filtered worker; none were weakened, deleted or bypassed. All three JSX claims remain reserved; existing React HIR claims retain their earlier parking status.

Raw outputs, sanitizer stderr, full failure stacks, individual comparison results, timings and mutant evidence are in evidence/constructed-context/validation.tar.gz.
