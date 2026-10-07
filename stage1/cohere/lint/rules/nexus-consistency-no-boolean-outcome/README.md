# Boolean outcome envelopes

A `.a` port of pinned Go cohere's `nexus/consistency-no-boolean-outcome`.
The listener checks interfaces and bare type-literal aliases. It preserves
first property-name order with the last duplicate's value, bare boolean types,
companion-field requirements, all third-party/Properties exemptions, exact
allowedTypeNames options and repeated role-suffix stripping. It reports the
whole flag signature and offers no fixes or suggestions. PolicyMessage renders
the subset of the shared helper's Go-generated catalog; StrictOptions validates
the generated Go target descriptor, and OptionsJson reads the decoded values.

Status: validated candidate, awaiting shared `.a` registration support.
`compatibility.patch` is a reviewable proposal, not applied repository code.
It extends module/mutant discovery, copying and corpus discovery to `.a`, adds
emitted JavaScript comparisons, and reconciles the profiling test's old file-list
API. The default registry and package commands fail without it. Approval is
needed to cross the user's rule-directory territory boundary.

After sourcing the setup environment, reproduce the bounded validation with:

```sh
python3 stage1/cohere/lint/rules/nexus-consistency-no-boolean-outcome/validate.py \
  --scratch /tmp/lint-wave1-12-recheck \
  --typescript /workspace/scratch/typescript-6.0.3
```

`--prepare-only` creates the scratch overlay without running tests. Its three
patched Go sources were reproduced byte for byte against the actual test sources.
The runner never edits shared repository files. `validation_test.go.txt` supplies
additional shape/options checks and benchmarks through a virtual Go test file.

The owned mutant removes the companion-field requirement. It compiles and exits
successfully on source Node, emitted JavaScript and sanitized native, then fails
only the comparison with Go because ordinary state receives an extra finding.
All baseline/corpus/options checks pass on those executions against Go.

No full repository test gate or default-build success is claimed. The pinned
cohere CLI refuses `.a` inputs, so its dry self-lint performed no checks. The
stage0 typechecker and both lowering backends accept the new modules. Arbitrary
malformed configuration diagnostic prose and general parser recovery are outside
this slice. Raw evidence and the exact commands are in the unit claim report.
