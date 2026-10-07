# Wave1-07 split for integration

Merge codex/lint-wave1-07-land at 33b2ed4dbab6dfbb8c83608455b6950e15563765 for the finished rules. It is based on current area e667e3e1 and current main 71d7e491 and contains only the four finished ledger-winning rule directories plus the finished-only claim and fresh validation evidence. No consistent-return or constructor-super directory, descriptor or ownership reservation is imported into the landing branch.

This original branch retains the unfinished consistent-return and constructor-super reservations. Their completion is independent of landing the four finished rules. The existing claim history remains here.

Fresh landing TestRulesAgree, four owned TestMutants and TestOwnedWitnesses passed 208.194s. Both formerly surviving mutants, constructor-property-name-ignored and ordering-relations-ignored, are caught on Node, emitted JavaScript and sanitized native by comparison with independent upstream Go. Dedicated option-bearing witnesses reach the previously unexercised branches; production predicates and mutation anchors are unchanged. The landing and original executable trees match exactly; this commit changes only this handoff note. Fresh raw evidence is on the landing branch, in stage1/cohere/lint/claims/wave1-07-land-evidence/tests.log.

Only own branches are pushed. No main or area push, no new claim or helper work, and no full-gate result claimed.
