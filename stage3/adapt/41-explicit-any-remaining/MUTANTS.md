# Checks proved able to fail

Every entry below was run. Restored-any controls use isolated actual source trees;
counts are full census observations, measured on checker-rejected programs.
No dedicated call-return adaptation mutant is claimed for the deferred timer host.

| Mutation | Caught by and result | Evidence |
| --- | --- | --- |
| Restore actual scanner state any | Census36 ->37; exactly scanner.ts:952:103 | evidence/any-mutant-result.json |
| Change scanner state to string | Stock TS2345 at its two real callers | evidence/any-proof.json |
| Restore actual Node timer return any | Census35 ->36; exactly the ambient return site | evidence/timers/delta.json |
| Restore actual Symbol links any cast | Refused5156 ->5157; exactly one unchecked cast | evidence/constructors/delta.json |
| Omit actual Symbol.id, SourceMapSource.text, Signature.parameters | Required-field completion proof fails for each actual constructor/completer | evidence/constructors/runtime.json |
| Restore actual clone source any cast | Refused5151 ->5152; exactly one unchecked cast | evidence/copies/delta.json |
| Omit path from actual clone | Own enumerable required-field/shallow-identity proof fails | evidence/copies/runtime.json |
| Restore actual convertToJson return any | Census26 ->27; replaces a JsonConfigValue-return lesson | evidence/json-config/delta.json |
| Change actual source-map version validation to2 | Valid/invalid map assertions fail on Node | evidence/json-config/runtime.json |
| Restore any on actual emitter local | Census21 ->22; exactly its value-any site | evidence/locals/delta.json |
| Change emitter directory domain to number | Eight stock diagnostics, including actual string assignments TS2322 | evidence/locals/wrong-domain-mutant.log.gz |
| Restore comparer’s original two any parameters | Census20 ->21; exactly comparer value-any site | evidence/data/delta.json |
| Reverse actual comparer scalar inequality | Existing Node settings comparison assertions fail | evidence/data/runtime.json |
| Add throw to actual whole-file emitted JS, per touched family file | JavaScript equality fails; thirteen continuation file checks | family emission.json/proof.json; evidence/finished/json-owner-emission.json |
| Duplicate retained interface anchor, drift its comment, duplicate shifted void owner | Each fails its exact adapter guard; idempotence/LF reconstruction controls pass | evidence/finished/guards.json |
| Leave shared JSON declarations private | Actual full oracle106365 pass/2 fail; missing JsonConfigObject and unsanctioned public bundle. Exported @internal owner control106366 pass/one existing API sanction | evidence/finished/private-owner-oracle.json and lane.json |
| Restore actual source void arrow | Exactly one void refusal, after zero | evidence/void-mutant-result.json |
| Actual diagnostics callback returns push result | Node proof finds number rather than undefined | evidence/void-proof.log.txt |
| Delete actual Debug.fail debugger | Node inspector pause count1 ->0, message unchanged; debugger retained | evidence/debugger-proof.log.txt |
| Any return in otherwise compiled number-return control | Current compiler refuses a function returning any | evidence/return-any-mutant.log.txt |
| Change emitted artifact bytes, remove artifact, append unrelated public declaration | Exact byte/file-set/declaration comparisons fail | evidence/any-proof.json |
| Append number=string to actual forInStatement1 input | Control6 pass; mutant3 pass/3 fail/four changed baselines; restored6 pass/zero differences | evidence/oracle-mutant-proof.json |
| Census audit signature/body range, extra finding and attribution mutants | Existing audit rejects each | evidence/census-audit.log.txt |
| Actual list predicate returns false, strict JSON parser returns{}, header initializes parent, Type header initializes symbol | Actual-body counterexample assertions fail; unmutated observations establish the residue | evidence/finished/residue.json |
| Stock any type replaces concrete timer type in no-any assertion | Assertion fails; actual stock type is Timeout|undefined | evidence/finished/timer-type.json |

Discarded probes are not successful mutants: dst:any/src:T in the comparer
produced two TS7053 errors, so the census skipped its body. The original two-any
signature is the successful mutant. Initial broad host and Array.at proposals
failed stock checking; constructor composition guards initially failed before
their newline/occurrence repair. Those logs remain with their family evidence.
An API-fixture command with eight workers ignored its test filter; that redundant
full run was cancelled. The fresh final lane is the reported result.
