Starting commit: b955d3030e79e9f01d12a96ab1fb4b167fc91c0c

CODE UNDER TEST: the Go progress inventory implementation, not an Adamic stage1 port. Reached functions: files, git, physicalLines, production, loadInventory, credit, blob, measure, finish, union, firstParagraph, appendUnique, pending. Coverage proof is functions.txt and reached.cover. main and fail were not reached.

ORACLE: self. All four tests assert hand-written fixture counts/boundaries/classifications. Git is run to construct and read real committed snapshots, not to compute a second progress inventory to compare against. No outside expected-value authority is cited.

Plan fixed before any mutant outcome: twelve production mutations from the fixed menu, distributed across the reached functions, and six distinct-entry empty-answer probes. No construction, witness, helper or family rows. No skip opt-ins. Go source already imports os, so all selectors use ADAMIC_MUTANT.
