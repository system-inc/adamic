Ported the indexed ESTree model, visitor keys, locations and TypeScript postprocessing.
The source driver currently converts expressions, basic statements and selected types.
47 generated files match 46,768 bytes on Go, source Node, sanitized native and emitted JS.
Three wrong-output port mutants finish normally and are caught on Node and native.
This is a first checkpoint; declaration conversion and repository coverage are pending.

# Non-JSON ESTree slice

The claim and origin survey are in the repository's `CLAIM.md`. This unit uses
cohere at `715ba94f3608a6500086b1076ce5cb7e51b836db`; `parse_json.go` is excluded.
The existing TypeScript parser and scanner are imported without edits. The
node table owns its nodes; all child links are numeric indexes, so there are
no owning parent/child cycles. Own fields retain insertion order and absence
is distinct from an explicitly stored null. The visitor table preserves all
337 type entries from Go, including entries not yet produced by this driver.

`main.ts <file.ts>` prints the canonical converted tree to stdout.
`main.ts --manifest <file>` reads one input path per line. This boundary is
an ESTree tree, not formatted JavaScript source. Output includes every own
field and field order, node ranges in UTF-8 bytes, content ends, parenthesis
flags, printer-visible locations, ignored-node semicolon behavior, merged
comments and byte-preserving stripped source. Strings use escaped UTF-16
units. The independent Go driver calls the unmodified public ESTree API
through an overlay, with no oracle implementation copied into the port.

The initial generated cases cover directives, ASI and semicolons, scalar,
regexp and template literals, omitted elements, object properties, spread,
assignment patterns, all expression operator categories, optional chains,
parenthesis boundaries, non-null and type assertions, basic declarations,
if/while/do/with/labels, break/continue/debugger, return/throw, ordinary and
nestled JSDoc comments, hashbangs, Unicode and CRLF, keyword/reference/array/
indexed/tuple/conditional/union/intersection types and type operators.

## Checkpoint validation

`go test -count=1 -v ./stage1/cohere/estree` runs generated agreement and
three successful mutants. Full output is in `validation/step1.log`.
The member mutant reverses computed access; the postprocess mutant removes
logical rebalancing; the JSDoc mutant changes the merged comment value.
Compilation failures, sanitizer crashes and nonzero exits do not qualify.

The focused cohere pass checks all source files and reports 100 percent
Adamic-ready. Original libraries are installed at pinned versions in scratch:
`@typescript-eslint/typescript-estree@8.65.0`, `typescript@6.0.3`, and
`prettier@3.9.6`. Independent library comparisons are pending.

Setup command: `bash cloud/setup.sh > /tmp/stage1-estree-setup.log 2>&1`, then
`source /workspace/adamic-tools/env.sh`. `nproc`: 5.

```text
setup: go ready (1s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (72s)
setup: done in 72s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

## Limits

Source extensions other than TypeScript are refused at this checkpoint.
Unsupported conversion kinds and generic argument wrappers fail explicitly.
The native parser does not promise diagnostic parity, JSX or full JSDoc
parsing. Parser list ranges are not part of its current node interface.
No compiler or runtime files were edited. No complete repository gate or
throughput measurement has been run yet.
