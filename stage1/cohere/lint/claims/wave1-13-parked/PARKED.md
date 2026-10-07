Parked rules from codex/lint-wave1-13, preserved in prior commit d6290b64c. Their descriptor directories are removed from the landing branch; their implementations and oracle evidence remain available in that commit.

nexus/consistency-no-hand-rolled-delay: top-level-await.ts.txt is accepted by Go, but stage 1 source Node exits 70 with expected semicolon at 6 before rule dispatch.

nexus/consistency-no-return-void: jsx-fix-reparse.tsx.txt emits the correct fix, then source Node exits 70 with expected GreaterThanToken, got Identifier at 24 in fixed source. Fixed-source reparsing loses JSX mode.

Prior full logs: d6290b64c:stage1/cohere/lint/rules/nexus-consistency-no-return-void/evidence/parking/d3a-rules.log and d3a-jsx-reparse.log. These shared parser/repair gaps are filed with integration. Neither rule is certified or registered on this landing branch.
