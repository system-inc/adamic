Added minimal numeric rule.json kinds manifests for twelve existing wave 20 claims; nine rules are implemented.
Base: current origin/main e8ba3d5d; prior validated and pushed tip 02d39bae.
Command: manifest_verify.py against independent production-registration oracle; PASS 12 manifests, 136 bytes.
Mutants: twelve metadata kind increments rejected; these are not new semantic or compiling native mutants.
Not covered: numeric handed-node conversion, three React source analyses, shared Diagnostic integration, new claims.

All twelve manifests contain only the requested numeric `kinds` property.
No type-aware rule.json example was found across fetched origin branches, so
no additional shared schema or module-loading fields are assumed. Three new
manifest directories describe the first batch's existing sibling root modules;
they do not move or duplicate those implementations. React manifests are
metadata for blocked claims, not completed analyses.

Validation reads the unchanged independent Go listener oracle used by
LISTENER_REPORT.md, compares all serialized kind numbers byte for byte, and
rejects each first-kind-plus-one metadata mutant. Oracle stdout/stderr, results
and command output are in validation/rule-json/. Command:

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/typeaware/prefer-promise-reject-errors/manifest_verify.py \
 --oracle /workspace/wave20-validation/listener-guards-twelve/oracle \
 --artifacts /workspace/wave20-validation/rule-json \
 > /workspace/wave20-validation/rule-json-check.log 2>&1
```

No implementation, compiler, checker, harness or shared generator changed.
The nine full byte-comparison gates, semantic/fact/registry mutants and native
sanitizers already passed at 02d39bae as documented in LISTENER_REPORT.md;
they were not repeated for JSON-only metadata. Previous native/Go timings
remain observations, with native slower by 1.69 to 5.08 times on the recorded
corpora. This metadata change establishes no speed improvement.

The current parser still has string-only ParseNode.kind. Shared numeric-node
handoff and kind-indexed dispatch are absent here, so existing handlers do
not yet satisfy the per-node speed contract. Changing shared parser or driver
is forbidden by the unit scope. JSX support exists on a separate branch but
is not integrated here; native React HIR/SSA/control analysis is also missing.
The three React claims overlap other workers and remain unfinished, so no
additional claims were taken. No batch 8 Diagnostic SHA has been named by
the user. Current main is already an ancestor; no further rebase was needed.
Only codex/typeaware-wave-20 is published, never main or any area branch.
