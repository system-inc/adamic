Rebased the parked owned rule branch cleanly onto current main c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06; shared harness integration remains the named parking blocker.
Rebased code tip 51b772e08bf4f94229d905c38832a7af4fa19eea; the following report commit is the final push. Helper landing tip 8395351de6c662d8538cc515dfc03d372214bd27 is green on the same main.
Fresh independent Go/source Node/emitted JavaScript/sanitized-native comparisons PASS; earlier six-rule suite PASS 283.113s; helper packages PASS 141.544s and 152.709s.
All owned semantic mutants were caught on all three runtimes, as were all 13 numeric declaration mutants (39 comparisons).
Uncovered: full repository gate, shared registration/context/Diagnostic integration (#zmh9v36), node-local dispatch, JSX/parser exclusions and boundaries resolver findings; the claimed program cache still lacks Mutex.

Both branches rebased without conflicts. No inherited integration diff was reverted. This main landing adds stage3 work and internal/oracle/stage3_hook_test.go; lint/parser/compiler implementation files are unchanged. Checked 777 compiler/parser/owned source files byte-identical against the isolated test-support snapshot. The legacy harness copies are isolated test support; no shared harness source is added to this branch. See PARKING_REPORT.md for the original parking construction, provenance, corpus boundaries and scratch overlays.

Fresh commands, after source /workspace/adamic-tools/env.sh, ran the same six independent Python validators and the filtered older six-rule overlay suite recorded in PARKING_REPORT.md. Their outputs are retained under landing-c019-evidence. All final commands exited zero. Test output went to log files. Native correctness runs use ASan/UBSan; release throughput counts match Go and Node. No production rule or regex implementation changed in this landing unit. New regex ports must use the shared translation row or JS RegExp, rather than a hand-written matcher.

Coverage and exact bytes remain: complexity 201 upstream cases and 351 compiler/stage1 files, 13,425,677 bytes; three TypeScript rules 39/22/70 upstream cases and 351 files, 39,582,002 bytes; whitespace four supported upstream cases, 21 JSX exclusions and 351 files, 13,185,533 bytes. Boundaries has 7,742 decisions from nine actual decoded configurations, 1,941,874 bytes. Google font decisions match 85,014 bytes. Numeric listener metadata matches 574 bytes across 13 rules and 25 entries. The earlier aggregate covers 371 compiler/stage1 files (15,221,259 bytes), 551 supported upstream cases excluding two JSX and one malformed-comment input, and the separate 100-case non-null/this-alias corpus without exclusions.

Mutants: modified switch counting; function-type ArrayType parentheses omission; namespace keyword replacement; first repeated triple-slash directive retained; whitespace separator deletion; boundaries policy precedence; Google display handling; and the earlier assignment-token, RTL exemption, directive reason, parenthesized-chain, postfix-assignment suggestion and compound-this-alias mutations. Each compiled and executed successfully and was caught by comparison. Thirteen numeric declaration mutations also compiled and were caught on all three runtimes. Inherited mutants passed too. These successes do not remove the explicitly excluded parser/resolver cases.

Current throughput findings/second (native / Node / Go), best of three per validator compiler-plus-stress corpus, with concurrent focused verification in this unit:
complexity Go findings 1325 seconds 0.200574 findings/s 6606.04
complexity native findings 1325 seconds 1.419825 findings/s 933.21
complexity Node findings 1325 seconds 0.887726 findings/s 1492.58
@typescript-eslint/prefer-function-type Go findings 1000 seconds 0.200508 findings/s 4987.33
@typescript-eslint/prefer-function-type native findings 1000 seconds 3.830606 findings/s 261.06
@typescript-eslint/prefer-function-type Node findings 1000 seconds 1.496512 findings/s 668.22
@typescript-eslint/prefer-namespace-keyword Go findings 1000 seconds 0.184240 findings/s 5427.69
@typescript-eslint/prefer-namespace-keyword native findings 1000 seconds 1.055473 findings/s 947.44
@typescript-eslint/prefer-namespace-keyword Node findings 1000 seconds 0.855811 findings/s 1168.48
@typescript-eslint/triple-slash-reference Go findings 1000 seconds 0.275931 findings/s 3624.10
@typescript-eslint/triple-slash-reference native findings 1000 seconds 1.683115 findings/s 594.14
@typescript-eslint/triple-slash-reference Node findings 1000 seconds 1.133037 findings/s 882.58
better-tailwindcss/no-unnecessary-whitespace Go findings 1000 seconds 0.544858 findings/s 1835.34
better-tailwindcss/no-unnecessary-whitespace native findings 1000 seconds 1.850410 findings/s 540.42
better-tailwindcss/no-unnecessary-whitespace Node findings 1000 seconds 1.626537 findings/s 614.80

Helper vet is clean; the filtered uncached TestTheOracleCatchesOneByte passes in 0.806s with the newly landed hook compiled. No new helper claim is made. Current native-build and emitted-JavaScript cache probes both exit 1 with TS2305, adamic has no exported member Mutex. That prerequisite probe is not a semantic mutant or a cache parity proof, and the cache unblocks zero rules. The three existing delivered helpers remain green and supply 18 dependency edges to the six previously named Tailwind rules, without removing their last blockers. Stop at the missing shared runtime/API rather than editing shared files.
