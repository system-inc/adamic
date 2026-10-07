# Wave 30 parked from the unified landing

All fourteen retained wave-30 rule claims are PARKED for unified-harness landing.
There are no unblocked owned registered rules, so Ahra's all-blocked exception
skips step 1's landing branch. This branch is a parked archive, not a landing
candidate. Existing standalone proofs and components are retained for future
integration; no blocked rule is included on the new no-useless-catch branch.
The last standalone validated snapshot is d8629dc2df0ccb73f045eb062c739be2f82285ae.

The eight standalone ports require RuleContext to carry the project checker
program and a parser-node/checker-node correlation. They currently use the
separate typeaware Rules.program/Bindings bridge. Reopening a checker program
per rule would lose project options, source identities and shared lifetimes.
No shared context, registration or harness workaround is included.

| Rule | Required fact | Existing runnable reproducer |
| --- | --- | --- |
| nexus/consistency-no-iso-string-date-cut | Date/toISOString symbol provenance | TestWave30AgreementAndMutants, controls in wave_30_test.go |
| nexus/correctness-no-callback-in-parse-try | global parser and callable symbol provenance | TestWave30AgreementAndMutants, controls in wave_30_test.go |
| nexus/correctness-no-collection-misuse | collection receiver type and property presence | TestWave30NextAgreementAndMutants, controls in wave_30_next_test.go |
| nexus/correctness-no-discarded-outcome | return-type declaration ancestry | TestWave30NextAgreementAndMutants, outcome controls in wave_30_next_test.go |
| nexus/correctness-no-discarded-pure-result | callee declarations and default-library identity | TestWave30NextAgreementAndMutants, pure-result controls in wave_30_next_test.go |
| nexus/correctness-no-uncleared-race-timeout | Promise/race/setTimeout provenance and symbol identity | TestWave30ThirdAgreementAndMutants, controls in wave_30_third_test.go |
| nexus/correctness-no-process-exit-after-output | resolved calls, module resolution and symbol declarations | TestWave30ProcessExitAgreement, testdata/wave_30_exit_controls.json |
| nexus/correctness-require-blocking-standard-streams | imported stream/callee declarations and program context | TestWave30BlockingStreamsAgreement, testdata/wave_30_blocking_controls.json |

Reproduce these positive Go/standalone controls with
`go test ./stage1/cohere/typeaware -run '^TestWave30(AgreementAndMutants|NextAgreementAndMutants|ThirdAgreementAndMutants|ProcessExitAgreement|BlockingStreamsAgreement)$' -count=1 -timeout 30m -v > /tmp/wave-30-parked.log 2>&1`
after sourcing the cloud toolchain environment. These are standalone checks,
not unified certification. Each implementation above takes checker-backed
Bindings rather than RuleContext, so it cannot be registered on the unified
context without that missing integration. Prior corpus runs need the documented
TypeScript source and frozen manifests; controls themselves do not.

The remaining claims have the following exact source reproducers (TSX):

| Rule | Reproducer | Blocker |
| --- | --- | --- |
| react/jsx-fragments | `const value = <React.Fragment />;` | checker alias/import symbol facts on RuleContext |
| react/jsx-no-constructed-context-values | `function Component(){return <Ctx.Provider value={{a:1}}/>;}` | checker provider origin and symbol-follow construction facts |
| react/jsx-no-undef | `const value = <Missing />;` | checker GetSymbolAtLocation and project/global-scope facts |
| react-hooks/purity | `function Component(){return <div>{Math.random()}</div>;}` | native React HIR/SSA/render/capture analysis |
| react-hooks/refs | `function Component(props){const value=props.ref.current;return <div>{value}</div>;}` | native React HIR/SSA/capture analysis |
| react-hooks/preserve-manual-memoization | `function Component(props){const data=useMemo(()=>props.items.edges.nodes??[],[props.items?.edges?.nodes]);return <Foo data={data}/>;}` | native React HIR/SSA/reactive scopes/capture analysis |

`go test ./stage1/cohere/typeaware -run '^TestWave30(JsxPrerequisites|ReactPrerequisites)$' -count=1 -v > /tmp/wave-30-parked-jsx.log 2>&1`
loads exactly those TSX controls: Go reports and native parsing succeeds, but
full rule findings cannot be produced by the registered context. See
wave_30_jsx_prerequisites_test.go and wave_30_react_prerequisites_test.go.
Parser.path already supplies filenames. JSX parsing is not the current blocker.

No subsequent status-only turns are needed. Resume these claims only when
checker/analysis integration is available or integration explicitly assigns it.
