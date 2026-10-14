Ran the production admission proof with the full generated manifest and mandatory revision diff.
Tool source: 543925aa; base 36e7fac4; head 7acd4496; this commit stores the result.
The exact invocation in overload-proof.inputs.json exited 0: verdict pass, admitted 1, checked 1, omitted 0.
No new mutant was planted for this evidence-only run; the unchanged tool uses its previously recorded guard mutants.
Not covered: runtime equivalence of accepted-by-both inputs; no sampling, filtering, or sharding was used.

Base: 36e7fac41ba3ecd480f10393a3a9f7b339587b14
Head: 7acd44960c497aac4b96e1a04e53f7b4e31b7c19

The original generated manifest is preserved. Its corpora contain witnesses 2, fixtures 933, gaps 107, review 6 and fuzz 0. The three diff inputs are outside the generator patterns, and are nevertheless included whole. All 1,051 unique inputs were classified: 896 accepted-by-both, 154 refused-by-both and one newly accepted. There were no compiler crashes, timeouts or errors.

| Diff path | Class | Base exit | Head exit |
| --- | --- | ---: | ---: |
| internal/oracle/testdata/overload_body_proof/block.a | accepted-by-both | 0 | 0 |
| internal/oracle/testdata/overload_body_proof/evaluate-negative.a | refused-by-both | 1 | 1 |
| internal/oracle/testdata/overload_body_proof/evaluate.a | newly-accepted | 1 | 0 |

Every newly admitted input was run on source Node using oracle/node.mjs, the JavaScript backend, and the default native release build. Each exit code was zero; JavaScript and native compilation also exited zero.

Disagreements: none.

Newly admitted: internal/oracle/testdata/overload_body_proof/evaluate.a

All three stdout byte strings: "word\ntrue:true:false\nfalse\nmissing\n7\n"
agree: true

| Phase | Seconds |
| --- | ---: |
| checkout | 2.958 |
| provision | 0.584 |
| base_build | 18.832 |
| head_build | 29.627 |
| manifest_and_verification | 2.231 |
| classification | 43.448 |
| runtime | 3.915 |
| total | 101.598 |
| External wall including cleanup | 101.913 |

Both revision compiler products were cache misses. The generator source revision and blob, the manifest blob, every corpus program blob, all observations, and the explicit admitted count are recorded in overload-proof.json. The manifest generator is taken from the pinned tool revision, with explicit provenance verification rather than clearing its metadata. Neither compiler source nor its fixtures were changed.

No Go files or tests changed in this run, so no additional Go package tests or fixture count refresh were needed. Integration lane checks run after the evidence commit before pushing compiler/admission-delta.
