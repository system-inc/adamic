Self-comparison member 18: compiler/self-compare-main 387c2826.

No merge conflicts. The incoming scalar equality helper coexists with the chain emitter. Incoming main test splits are preserved. Imported historical evidence is under source-member-18 and is not new validation.

Focused command: GOMAXPROCS=4 timeout 120s go test -p 1 -parallel 4 -timeout 90s ./internal/oracle -run '^TestSelfCompare' -count=1 -json. All four tests pass: constant mutant 0.38s, closure 0.74s, array-fill 0.74s, sweep 0.77s. Controls compare Node, JavaScript, release native, ASan/UBSan, leaks and counted allocations. The actual emitted numeric self-equality mutant runs cleanly and its wrong NaN stdout is caught.

Counts are regenerated once with timeout 240s and Go timeout 210s. Individual row changes and measured output are in member-18-count-changes.json and member-18-counts.log.gz. No stage 3 Node observations were changed. WASI self-comparison and the full gate were not run for this member.
