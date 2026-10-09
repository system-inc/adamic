Bases: main 48c05d091 and area bdef33ba5; both merged. Cohere oracle pin 7945d102a6c18dd36adf9114a758ce646e8b2359. Ledger unchanged.

Fresh initial parity PASS 194.09s; required pinned compiler + stage1 corpus PASS 415.42s, 486 files. All four owned mutants caught on Node, emitted JavaScript and native, including constructor-property-name-ignored and ordering-relations-ignored. Full witnesses failed only on inherited nexus/consistency-no-single-line-jsdoc: /**\u0085Unicode prose.\u0085*/ expected a replacement retaining U+0085 but the port removed it.

Cherry-picked b31e6965c: the inherited rule-local trim helper now matches current upstream ECMAScript whitespace (retain NEL, trim BOM). No harness, oracle, fixture or assertion changed. Final TestRulesAgree + full TestOwnedWitnesses PASS 121.594s; 13739539 parity bytes and 300148 witness bytes identical on Go, Node, emitted JavaScript and native. Initial corpus and owned mutant runs preceded this inherited trim repair; their rule code and anchors did not change. No full repository gate claimed.

Four standalone TestCompileProfiles passed with WAVE07_PROFILE and WAVE07_ARTIFACTS explicitly unset: no-underscore-dangle 15.349s, no-unsafe-negation 17.448s, no-unsafe-optional-chaining 14.884s, typescript-no-this-alias 15.806s. No skips. All compile sanitized native, release native and emitted JavaScript.

Original wave07 production sources match this branch except profile tests/validation scripts and claim reports. The shared parity/corpus/mutant evidence applies to both. consistent-return and constructor-super remain unfinished on the original branch only. Raw logs retain the initial witness failure and final pass.
