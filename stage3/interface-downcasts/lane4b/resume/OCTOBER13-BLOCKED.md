Built: 12 additional original-declaration probes pin the remaining compiler dependencies; no additional runtime certificates.
Commits: d732e939 defect witnesses and c1e29126 original certification batch pushed; this frontier group follows on the same lane branch.
Checks: source Node and exact lowering refusal checks pass for all 12 new probes; existing original frontier tests also passed in the certification run.
Mutants: no production change or new runtime certificate in this group; prior three defect reversions and 54 pair mutants remain caught.
Uncovered: 16 pairs / 68 candidate reads remain blocked; certified total is 26 pairs / 113 reads; whole-tsc reachability is unmeasured.

Observed failures are in logs/oct13-remaining-frontiers.log. The full inventory
mapping is in remaining-blockers.json. Grouping other rows under a representative
frontier is an inference; it is not a claim that every production read executed.
These probes import complete original declarations, including builder.d.ts and
BuildOptions from its actual module, tsbuildPublic.d.ts. Source Node runs before
ordinary strict loading and lowering. Each final probe pins the exact refusal.

| Dependency | Pairs | Candidate reads |
| --- | ---: | ---: |
| Union intersection descendant checks | 7 | 45 |
| Dictionary cast admission and lookup | 2 | 11 |
| Tuple admission and positional contract | 1 | 4 |
| Callable member | 1 | 3 |
| Branded tuple array elements | 2 | 2 |
| Tuple union member | 1 | 1 |
| Object plus primitive array consumers | 2 | 2 |

CompilerOptions and BuildOptions probes refuse their casts, before lookup.
The tuple-element probe also refuses its cast, so it proves no read boundary.
EmitHelper.text is blocked as a callable union field. Incremental root/signature
arrays refuse their element representation. Bundle outSignature reaches the new
named tuple union refusal. The fileInfo forEach probes call the method with a
receiver; they refuse array element representation rather than using a detached
method. JSDoc parent unions refuse the demanded descendant kind intersection.
The existing bindable and CompilerOptions probes retain their same named failures.

No failed frontier can emit a native or JavaScript program, so these probes have
no runtime count rows, leak runs, sanitizer runs, or acceptance mutants. Their
source Node outputs and compile refusals are evidence of blockers only. No shared
compiler implementation was changed to absorb another lane's obligation.

Command, with tool environment sourced and output redirected:

```
ADAMIC_OBJECT_PRIMITIVE_ORIGINAL_DECLS=/tmp/lane4b-declarations ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewObjectPrimitiveRemainingFrontiers$' -count=1 -v -timeout 5m
```

Development corrected BuildOptions' module after a source import error and
replaced detached method probes with receiver-bearing calls. Neither earlier
failure is used as certification evidence. Final evidence is the rerun above.
There is no time cutoff; this handoff is due to explicit compiler dependencies.
