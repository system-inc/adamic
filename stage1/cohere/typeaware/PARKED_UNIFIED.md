# Wave 27: parked on the unified harness

All twelve implemented rules below are parked; none is offered as green on the unified harness. Their checker-backed modules cannot link in the unified harness, whose native build does not supply a typescript-go checker archive:

- base/correctness-require-orm-column-nullable-parity
- base/correctness-require-serializable-nullable-parity
- base/correctness-require-verify-array-parity
- nexus/correctness-no-process-exit-after-output
- nexus/correctness-no-uncleared-race-timeout
- nexus/correctness-require-blocking-standard-streams
- prefer-regex-literals
- prefer-rest-params
- react-hooks/exhaustive-deps
- require-await
- symbol-description
- valid-typeof

Reproducer (from the repository root, with an Adamic compiler built at origin/area/stage1-lint):

```sh
adamic build stage1/cohere/typeaware/wave_27_fifth/suite.a -o /tmp/wave-27-unified-gap
```

Observed exit 1: `stage1/cohere/typeaware/rules.ts:104:16: Adamic 0.1 refuses an unlinked typescript-go library call; build with --tsgo <checker archive>`.

The same imported checker API is required by the other suites. Even valid-typeof and prefer-regex-literals need symbol identity to distinguish globals from shadowed bindings. RuleContext currently provides no checker facts and buildPort does not link that archive. Prior custom-bridge results remain historical evidence, not unified-harness certification.

react-hooks/exhaustive-deps also remains blocked on dynamic RegExp lowering: building wave_27_third/additional_hooks_pattern.a, whose additionalHooks option calls new RegExp(pattern, 'u'), refuses a nonconstant pattern. Runtime regex compilation is due October 9.

The earlier React claims react-hooks/set-state-in-effect, react-hooks/set-state-in-render and react-hooks/static-components remain parked on native high-level IR/single-assignment/capture analysis, as recorded in the claims.

Step 1 rule landing is skipped because every implemented owned rule is blocked. The separate JSDoc port starts from origin/area/stage1-lint and contains none of these blocked modules.
