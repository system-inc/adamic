# symbol-description

The shared checker context is present. This rule stops at an unanswered
question or missing shared helper, per the unpark instruction.

Exact upstream call: `ctx.TypeChecker.GetSymbolAtLocation; symbol.Declarations`,
[cohere/internal/lint/rules/core/no_new_native_nonconstructor.go:104](/workspace/adamic/cohere/internal/lint/rules/core/no_new_native_nonconstructor.go:104).

declarations is restricted to class/interface nodes. symbol-origin does not expose declaration presence or IsDeclarationFile; see shared resolvesToAGlobal at lines 104-109.

Minimal input: [repro.ts.txt](./repro.ts.txt).
No descriptor or native rule is installed; zero upstream cases certified and
no new mutant claimed. No private checker, bridge question or shared helper
copy was added. Earlier standalone captured-fact proofs are not unified
harness certification.
