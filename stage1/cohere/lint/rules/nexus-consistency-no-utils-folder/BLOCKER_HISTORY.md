# Filename transport blocker history

Resolved by shared commit 59451e23e. Current green validation is in LANDING.md.
The observations below describe the earlier harness.

The port is complete for the path predicate, ordered messages and zero-width
SourceFile range. It listens only to SourceFile and receives the handed node.
The upstream rule has no options struct; its adapter decodes supplied JSON
into upstream's any argument, which the rule itself ignores. Both actual Go
test names are captured by TestConsistencyNoUtilsFolder.

The shared harness rewrites every owned witness filename to a temporary
<slug>-<index>.ts, and rewrites captured upstream paths to case-NNN/<basename>.
No source text or options can make this path-only rule fire there: upstream
ignores all options. A witness-options sidecar alone cannot preserve a path.

Minimal reproducer: put testdata/witness.ts.txt's source at
/project/source/utils/Thing.ts. The real upstream rule reports noUtils at
range 0..0. The same source at /tmp/case-000/Thing.ts is clean.
Also test /project/source/_utils/Thing.ts and
/project/_utils/utils/Thing.ts for the ordered pair of findings.

Required shared change: transport a witness's logical filename, including
directory segments, to both Go and native/Node; preserve captured upstream
directory paths too. Do not simulate filenames through invented options or
modify the independent upstream rule. No shared harness changes are made.

Observed unified failures are recorded in REPORT.md and evidence/. Real-path
Go/Node/emitted/sanitizer agreement and the caught mutant are also preserved.
