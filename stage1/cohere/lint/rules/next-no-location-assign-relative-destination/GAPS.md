# Next.js integration gaps

Seven selected rules need JSX nodes the stage1 parser cannot produce:
`no-head-element`, `no-html-link-for-pages`, `no-img-element`,
`no-page-custom-font`, `no-styled-jsx-in-document`, `no-sync-scripts`,
and `no-unwanted-polyfillio`.

[The ledger](evidence/parser-gaps.json) retains the shortest captured Go-positive
source for each, its original filename, findings, and exact Node/native refusal.
These seven are not registered or counted as source parity. Both backends exit
70 with matching parser panic text. The parser probe compiles successfully;
a non-JSX control exits zero on both backends with an empty stderr. Native
uses ASan, UBSan and leak checking. This is a parser boundary measurement,
not an implementation of those listeners.

The Go result for `no-html-link-for-pages` comes from its original successful
fixture, which supplies a program and route files. Its route model additionally
requires `Context.Program`, compiler options and disk traversal. RuleContext
has no program context. A handwritten context with Program nil registers no
listener in Go too; that zero cannot establish route-model parity.

Two selected non-JSX listeners also cannot meet the current shared registration
witness contract: `no-head-import-in-document` and `no-typos`.
[boundary_test.go](testdata/probes/boundary_test.go) runs each unchanged Go rule
on a valid positive source at its original path, then at `case-000.ts`.
Both produce one finding at the original path and zero at the renamed path.
The shared capture's record excludes `File`; its replay writes every input as
`case-NNN.ts`. The owned-witness runner copies each input as `<slug>-N.ts`.
Both lose the document basename or Pages Router directory. The shared oracle
also hardcodes ScriptKindTS, which must change before replaying JSX.
The parser exposes its real path, so this second boundary concerns verification
infrastructure, not missing filename data in RuleContext.

CLAUDE.md's registration contract says: "Never edit a dispatch, oracle, corpus
or copied-file list." This unit stays in its owned rule directory and claim;
it does not change shared capture or parser code to work around these limits.

To reproduce from the repository root:

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/lint/rules/next-no-location-assign-relative-destination/testdata/probes/capture.py /workspace/scratch/nextjs-lint-probes > /tmp/nextjs-capture.log 2>&1
python3 stage1/cohere/lint/rules/next-no-location-assign-relative-destination/testdata/probes/parser_gaps.py /workspace/scratch/nextjs-lint-probes > /tmp/nextjs-parser-gaps.log 2>&1
```

The capture runs the original ten-rule tests and the filename probe through a
Go overlay without modifying the cohere submodule. It asserts a nonempty corpus
and a positive finding in every captured family. Custom multi-file assertions
are not all captured; the 315 records are not a complete replay of every route
fixture. The parser probe bounds each child to ten seconds, requires the clean
controls, and pins both exit code and matching panic bytes. It fails if JSX
parsing starts succeeding, so this ledger cannot silently become stale.

To unblock the nine ports: provide the common JSX AST adapter and parser modes;
preserve original filenames/extensions in discovered witnesses and capture;
then provide the program/route context for `no-html-link-for-pages`. The helper
branch at `29990b4` provides options/schema/messages, not JSX parsing, and marks
all selected Next.js rules helper_ready false. No Node-only or externally
supplied AST listener parity is claimed here.
