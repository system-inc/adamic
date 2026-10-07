# no-unsafe-negation

Own-directory .a port of the pinned Go cohere rule, registered by rule.json with no manual dispatch or harness changes. The original claim was pushed before implementation.

Validation uses independent parsing and the unmodified upstream Go rule. Source Node, emitted JavaScript and ASan/UBSan native match findings, messages, ranges, suggestion IDs/messages/edit ranges/replacements, and unchanged automatic-fix output byte for byte. There are 62 unique upstream source and repaired-expression literals, with option variants. See evidence/validation.txt and the compressed raw outputs. Options are valid decoded upstream settings; arbitrary invalid configuration diagnostic parity is outside this profile contract.

The corpus includes TypeScript src/compiler at 050880ce59e30b356b686bd3144efe24f875ebc8 and all stage1 .ts/.a sources. Final corpus evidence includes the Google Font modules added during this completion. The semantic mutant ordering-relations-ignored builds and executes cleanly, then is rejected only by byte comparison on all three Adamic runtimes.

Reproduce with `source /workspace/adamic-tools/env.sh` and `python3 stage1/cohere/lint/rules/no-unsafe-negation/validate.py --scratch /tmp/wave07-negation --compiler /tmp/wave07-typescript`, redirecting test output to a log. The compiler checkout must have the pinned revision. Go overlays expose the actual rule and do not alter cohere sources.

Observed 1,000-finding rates, best of five with startup included:

- THROUGHPUT Go 198183.80 findings/s best of 5 including startup
- THROUGHPUT native 92265.76 findings/s best of 5 including startup
- THROUGHPUT Node 8990.31 findings/s best of 5 including startup

The registry, touched packages, upstream core tests and filtered uncached input oracle pass. Full repository gate not run. Setup initially failed while a descriptor preceded its witness; after the owned witnesses were added, setup succeeded in 20s, nproc 5. See evidence/setup.txt. `docs/parallel-work.md` is absent on this checkout; CLAUDE.md and docs/lint-registration.md supply the existing ownership/registration contract.

Both suggestions attach directly to the finding, with token-trimmed source slices. The first changes the comparison and the second preserves the existing negation; neither auto-applies. All four ordering operators remain opt-in. The source-literal extractor initially picked up prose/comment strings and refused them; it was corrected to extract bounded code literals before the successful runs.
