Current parked rule: nexus/consistency-no-hand-rolled-delay. Its descriptor directory is removed from the landing branch, with implementation and prior oracle evidence preserved in d6290b64c.

nexus/consistency-no-hand-rolled-delay: top-level-await.ts.txt is accepted by Go, but stage 1 source Node exits 70 with expected semicolon at 6 before rule dispatch.

Resolved historical blocker, nexus/consistency-no-return-void: jsx-fix-reparse.tsx.txt emitted the correct fix, then source Node exits 70 with expected GreaterThanToken, got Identifier at 24 in fixed source. Fixed-source reparsing loses JSX mode.

Prior full logs: d6290b64c:stage1/cohere/lint/rules/nexus-consistency-no-return-void/evidence/parking/d3a-rules.log and d3a-jsx-reparse.log. These shared parser/repair gaps are filed with integration. The shared fix 4f18a05c9 preserves JSX mode during fixed-source reparsing. Return-void is restored to this landing branch and its exact formerly failing source is now testdata/jsx-fix-reparse.tsx.txt; the unified upstream, witness and mutant comparisons pass. Only delay remains unregistered and uncertified.
