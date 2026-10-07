Merged main 48c05d091 via area bdef33ba5. Current oracle pin 7945d102a6c18dd36adf9114a758ce646e8b2359. Ledger unchanged.

Production sources match codex/lint-wave1-07-profile-tests, excluding its four standalone Go tests/validation scripts and claim history. Shared fresh evidence: initial TestRulesAgree PASS; required compiler/stage1 corpus 486 files PASS; all four owned mutants caught on Node, emitted JavaScript and native. Initial full witness failure was inherited nexus/consistency-no-single-line-jsdoc trimming U+0085 contrary to the new upstream ECMAScript whitespace semantics.

Cherry-picked b31e6965c to retain NEL and trim BOM in that rule-local helper. Final full TestRulesAgree + TestOwnedWitnesses PASS 121.594s: 13739539 and 300148 identical bytes across Go/Node/emitted JavaScript/native. No harness, oracle, fixture or assertion changed. Initial corpus and owned mutants precede this inherited trim repair; own rule code/anchors unchanged. Raw initial failure and final passing logs retained here. No full repository gate claimed.

Finished four rules are already integrated. consistent-return awaits Judge/EndReachable integration; constructor-super remains unfinished on this original branch. Profile-test follow-up pushed separately at f51d28dc8.
