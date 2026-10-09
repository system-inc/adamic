# #whkxbc7 marks

The user supplied this task list and its current owners on October 8. This
supersedes the provisional list and owner guesses in REPORT.md and
historical-marks.json. Evidence is the measurement committed as 437a8e9f8;
no compiler, scanner or witness was rerun for this marking update.

“Gone” below means the existing isolated admission control builds and its
output matches Node; its one-byte output mutant fails comparison. It does
not establish complete scanner execution. Stops after the unchanged first
stop depend on private throwing placeholders, which shift some locations.

| Mark | Supplied checkpoint | Current owner | Existing evidence |
| --- | --- | --- | --- |
| Gone | corePublic.ts:9:5 index signature | records-maplike, #rhjc4x4 | Control 01-index passes native/Node comparison; mutant caught. |
| Still there | debug.ts:8:5 mutable namespace export | compiler area-stack group 1 | First stop in both split modes. |
| Still there | debug.ts:15:14 unchecked cast | compiler #b5w3ycg | Stop 2, now debug.ts:14:14. |
| Still there | diagnosticInformationMap.generated.ts:13:34 unchecked cast | compiler #b5w3ycg | Stop 3. |
| Still there | scanner.ts:3497:67 unchecked cast | compiler #b5w3ycg | Stop 5, now scanner.ts:3510:67. |
| Still there | utilities.ts:65:22 Uint16Array | runtime #4gkdjsz | Stop 4. |
| Still there in isolated control | scanner.ts:1269:28 enum slot through assignment | compiler #cvhj5fk | Control 06-enum-slot remains refused; original scanner checkpoint not established by the fifteen-stop traversal. |
| Still there | core.ts:107:20 new Array(length) with holes | compiler stricter options #k881crd | Stop 6, aliased constructor diagnostic. |
| Still there | utilities.ts:13:17 generic returns | compiler area-stack | Stop 7. |
| Gone | scanner.ts:113:5 computed field name | compiler area-stack | Control 11-computed-field passes native/Node comparison; mutant caught. |
| Still there in isolated control | scanner.ts:432:59 Object.entries | compiler #d9eemrs | Control 13-entries remains refused; original scanner checkpoint not established by this traversal. |
| Unverified, not marked gone | scanner.ts:4097:24 undefined! placeholder | compiler #9wc5q5j | Not reached in fifteen stops; undefined! placeholders remain in the adapted scanner source. No fresh diagnostic for this checkpoint. |
| Unverified, not marked gone | Error.captureStackTrace | library #jj9z3qn | Library merge was skipped after 56 conflicts; this checkpoint was not measured on a clean library integration. |

New stops relative to the supplied list follow. Owners are suggested compiler
routing only; no task assignment was supplied for these sites. Each has an
existing Node-success/compiler-failure witness in ordered-stops.json.

| Mark | Current stop and location | Feature | Suggested owner |
| --- | --- | --- | --- |
| New | 8 scanner.ts:696:32; 10 scanner.ts:1221:26 | String plus number | compiler |
| New | 9 scanner.ts:687:15 | Destructuring a string | compiler |
| New | 11 scanner.ts:1676:62 | Postfix expression used as a value | compiler |
| New | 12 debug.ts:15:74; 15 scanner.ts:2296:130 | Generic function as a value | compiler |
| New | 13 scanner.ts:1826:22 | Optional call | compiler |
| New | 14 core.ts:81:29 | Iteration over differently held union members | compiler |

The bounded traversal cannot retire an unvisited checkpoint. The scanner's
native byte comparison remains blocked at debug.ts:8:5.
