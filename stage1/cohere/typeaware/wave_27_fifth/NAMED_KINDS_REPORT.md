Corrected require-await, symbol-description and valid-typeof metadata to named ast.Kind listeners.
Base remains origin/main c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06; prior pushed code 2fe6b0029.
Independent Go enum validates 351 constants and three declarations; normal and sanitizer oracle recheck passes.
Nine metadata mutants rejected; prior symbol/typeof/await mutants were caught at bytes 39039/46485/33546.
Runtime additionalHooks remains blocked on nonconstant RegExp lowering; shared registry and emitted JS remain untested.

The latest user correction supersedes numeric rule.json descriptions in historical
landing reports. Metadata now uses CallExpression, BinaryExpression and the seven
function/accessor/constructor names. The owned parser compatibility adapter still
uses numeric values internally; handed-node rule implementations did not change.
No shared generator, harness or compiler file was edited.

Validation output: /workspace/wave-27-kindnames-validation.log. The independently
compiled production Go enum executable from the c019 gate checked all 351 constants.
Each declaration was separately mutated to an unknown name, a numeric value and
SourceFile; the metadata validator rejected all nine. Python validation script
syntax also passes. The owned full validator now expects named metadata.

Oracle recheck output: /workspace/wave-27-kindnames-oracle.log. Existing unchanged
c019 normal and sanitizer binaries were compared with production Go over the
186 parseable controls (64739 bytes), 287 repository files (18485 bytes) and 77
compiler files (5934 bytes). Native and sanitizer stderr are empty; Go stderr is
its mandatory timing line. Two initial recheck attempts incorrectly required
empty Go stderr and failed on that timing line before native execution; the
corrected check allows only its exact timing format. No rule source changed,
so build, per-rule mutant and released-handle results remain those documented
in landing_c019/REPORT.md. No performance retiming was needed: prior quiet fifth
batch native/Go medians are 547.758/154.696ms on repository (3.541x) and
3249.782/423.911ms on compiler (7.666x).

The previously observed native error remains: stage 0 cannot lower RegExp with
a nonconstant pattern. The isolated additional_hooks_pattern.a has the required
new RegExp(pattern, 'u') constructor but cannot be integrated natively. See
../wave_27_third/regex_resume/REPORT.md. No new claims were taken while this
nondefault-option gap remains. Three React analysis claims remain parked.
