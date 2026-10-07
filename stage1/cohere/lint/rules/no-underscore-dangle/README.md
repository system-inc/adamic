# no-underscore-dangle

Own-directory .a port of the pinned Go cohere rule, registered by rule.json with no manual dispatch or harness changes. The original claim was pushed before implementation.

Validation uses independent parsing and the unmodified upstream Go rule. Source Node, emitted JavaScript and ASan/UBSan native match findings, messages, ranges, suggestion IDs/messages/edit ranges/replacements, and unchanged automatic-fix output byte for byte. There are 93 unique upstream source and repaired-expression literals, with option variants. See evidence/validation.txt and the compressed raw outputs. Options are valid decoded upstream settings; arbitrary invalid configuration diagnostic parity is outside this profile contract.

The corpus includes TypeScript src/compiler at 050880ce59e30b356b686bd3144efe24f875ebc8 and all stage1 .ts/.a sources. Final corpus evidence includes the Google Font modules added during this completion. The semantic mutant constructor-property-name-ignored builds and executes cleanly, then is rejected only by byte comparison on all three Adamic runtimes.

Reproduce with `source /workspace/adamic-tools/env.sh` and `python3 stage1/cohere/lint/rules/no-underscore-dangle/validate.py --scratch /tmp/wave07-underscore --compiler /tmp/wave07-typescript`, redirecting test output to a log. The compiler checkout must have the pinned revision. Go overlays expose the actual rule and do not alter cohere sources.

Observed 1,000-finding rates, best of five with startup included:

- THROUGHPUT Go 159964.42 findings/s best of 5 including startup
- THROUGHPUT native 130404.82 findings/s best of 5 including startup
- THROUGHPUT Node 5449.64 findings/s best of 5 including startup

The registry, touched packages, upstream core tests and filtered uncached input oracle pass. Full repository gate not run. Setup initially failed while a descriptor preceded its witness; after the owned witnesses were added, setup succeeded in 20s, nproc 5. See evidence/setup.txt. `docs/parallel-work.md` is absent on this checkout; CLAUDE.md and docs/lint-registration.md supply the existing ownership/registration contract.

All nine option effects, default-true flags, private names, exact this.constructor exemption, nested destructuring and repeated declarator spans are covered. A whole compiler corpus run with enforceInMethodNames=true triggers an upstream Go panic (`Unhandled case in Node.Text: *ast.ComputedPropertyName`), since Go reads Text before its name-kind guard. Corpus parity therefore uses defaults; option variants remain covered in upstream literals. Go's panic is not silently repaired by this port and is not counted as a pass.
