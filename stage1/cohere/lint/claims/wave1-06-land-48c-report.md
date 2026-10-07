Bases: main 48c05d091 and area bdef33ba5; merged cleanly. Upstream cohere pin now 7945d102a6c18dd36adf9114a758ce646e8b2359. Dedup ledger unchanged.

Setup initially failed writing the default Go build cache because the root disk was full. Cleared only the disposable default Go build cache and retried in a task-specific cache; retry passed in 189.711s with prescribed GOPROXY. Both logs retained.

Initial fresh TestRulesAgree passed; all seven owned mutants were caught on Node, emitted JavaScript and native. Full TestOwnedWitnesses found an inherited nexus/consistency-no-single-line-jsdoc whitespace mismatch: source /**\u0085Unicode prose.\u0085*/ produces a Go replacement retaining U+0085, while the port dropped it. Current cohere comments.ContentLines uses text.TrimWhitespace (ECMAScript), replacing old Go TrimSpace semantics. The existing fixture also covers U+FEFF.

Repair: rule-local trim.a retains U+0085 and trims U+FEFF. No harness, oracle, fixture or assertion changed. Follow-up TestRulesAgree, full TestOwnedWitnesses and existing JSDoc replacement mutant PASS 183.709s; witness outputs match all four backends, 306733 bytes. The seven own mutant results precede this inherited helper repair; their rule code and anchors did not change. No full repository gate claimed.

Evidence: wave1-06-land-48c-evidence contains setup failure, setup retry, initial test failure and final passing comparison logs.
