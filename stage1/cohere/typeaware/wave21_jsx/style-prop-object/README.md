# Wave 21 style prop object core

`react/style-prop-object` is implemented as a native Adamic decision core in
`rule.a`. The claim was pushed in `81cf0cfa7`, after all prior owned work was
rebased onto main `39638d9e2`, re-green and pushed in `12b844c8e`. The 603-ref
claim audit found only this one available rule, at combined ranking position
192. Explicitly released older allocations have active continuation claims.

The core declares `JsxAttribute` and `CallExpression` by the Go AST names in
rule.json and its exported `syntaxKinds`. The private driver hands it the
already fetched node only for those two kinds. The core never dispatches on a
string kind and never refetches its handed node. Raw test-provider tags are
internal integers checked against the pinned Go constants. New Adamic sources
are `.a`. No shared registration generator, test harness or compiler was edited.

Both JSX and createElement arms are ported, including the exact six literal
kinds, null/template/unary/array exemptions, empty JSX containers, parentheses,
all four pragma binding shapes, require/member spellings, first-declaration
selection for identifiers, shorthand value symbols, component allow lists and
all three report anchors. The createElement arm stops after the first relevant
property exactly as Go does. The Go rule and its pragma helpers use no regex.

This is a **prepared-AST core, not a completed native source port**. Shared JSX
parsing is now present after the rebase onto area/stage1-lint `b46914832`; normal
and sanitized native parsers accept `<div style="bad" />` and produce a nonempty
source tree. The prerequisite mutant removes parsing and returns an empty tree,
which the updated assertion catches separately from the rule mutant.

The shared `stage1/cohere/lint/context.ts` `RuleContext` provides the filename through
`parser.path`, but no project configuration, live checker program handle, or symbol/declaration
API. This blocks feeding ordered identifier and shorthand bindings into the
core through the production driver. The descriptor remains a listener declaration,
not a completed shared-registry factory. Shared context/generator/harness edits
are prohibited for this unit; no such file was edited. Live checker adaptation
and production registration remain unimplemented. No new bridge question was
added: tests still use the explicitly separate raw provider below.

`oracle.go.txt` is built as an overlay inside cohere and calls the unchanged
production rule with an independent loader, program views and listener walk.
Its separate prepare mode exports the complete primary AST and raw ordered
symbol/shorthand declarations, parent/child/role links, literal/identifier text,
canonical file identity and token byte ranges. It calls no lint predicate or
analysis helper and emits no lint verdict. AST/checker preparation is outside
Adamic and outside the production bridge. A compact raw record stream allows
runtime loading of full compiler trees without compiling generated AST source.
Finding positions retain the provider's token byte ranges; this does not prove
the missing native UTF-16 source-to-byte adapter.

All **82/82** extracted production and independent controls parse with Go.
Three profiles match every finding, message, range, fix and suggestion under
normal native, ASan/UBSan/LSan, external source Node and emitted JavaScript:

| Profile | Findings | Canonical bytes |
| --- | ---: | ---: |
| Default | 45 | 7985 |
| Allow Foo, MyComponent, div | 22 | 5973 |
| No checker | 18 | 5598 |

All fixes and suggestions are empty, matching production Go. Controls include
all pragma bindings, merged declarations, aliased/shadowed calls, parentheses,
shorthand, duplicate style properties, Unicode identifiers, bigint and regex
literals, member tags and checker absence. The two frozen populations also
match normal and sanitized Go streams: 77 compiler roots / 5087 bytes and 287
repository roots / 18485 bytes, both zero findings. Unlike the earlier JSX-only
provider, this provider exports and the core visits complete non-JSX ASTs too:
887803 primary compiler nodes with 50379 call nodes, and 138034 primary repository
nodes with 8288 calls. Including external declaration stubs, the graphs contain
932047 and 147455 nodes respectively.

The real native rule mutant removes StringLiteral from the literal set. It
compiles, exits 0 and has empty stderr; only the independent complete-byte
comparison catches it at **byte 490**. Released program probes use real C calls:
normal and sanitized builds exit 70 with the exact invalid/released-handle panic.
The retained-registry mutant exits 0 and is caught by the required panic.

Final rule suite PASS **81.507 s**, source prerequisites PASS **10.460 s**, vet
exits 0. The initial supported-core pass took 172.032 s; reusing the already
decoded primary file identity reduced compiler raw-graph loading from 13.221877
to **6.085266 s**. Full Go compiler execution took **0.282518 s**. Repository
raw-graph native execution took **0.805371 s**, full Go **0.125324 s**. Three
control profiles total **0.049822 s** native, **0.139436 s** full Go, with raw Go
preparation **0.034981 s**. These include different loading pipelines and are
not end-to-end native source performance. Compilation and sanitizers are
excluded from these normal execution sums.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE21_STYLE_ARTIFACTS=/tmp/wave21-style-artifacts \
ADAMIC_WAVE21_COMPILER_CONFIG=/workspace/wave21-compiler/src/compiler/tsconfig.json \
ADAMIC_WAVE21_COMPILER_MANIFEST=/workspace/wave21-compiler.manifest \
ADAMIC_WAVE21_REPOSITORY_MANIFEST=/workspace/wave21-repository.manifest \
go test ./stage1/cohere/typeaware/wave21_jsx/style-prop-object \
  -count=1 -v -timeout=30m > /workspace/wave21-style-final.log 2>&1
```

The prerequisite check was added after that full suite and run separately with
`-run '^TestStyleSourcePrerequisite$'`, the same retained artifact variable and
its own log. Evidence, canonical streams and hashes are in validation. The
full repository gate, native JSX source pipeline, live binding adaptation,
production registration and repair application are not covered. Reservations
remain retained; this core's source integration is blocked, not released.

## Current-main c799 landing

The branch was rebased again onto `c7991b900` after its lowering proofs and record
runtime changes landed. The complete style suite, now including its source
prerequisite, rebuilt and passed in **152.127 s**. Both production corpora, all
82 controls in three modes, original/emitted Node, sanitizers, the byte-only
rule mutant, retained-registry mutant and parser-skip prerequisite mutant pass.
The source integration gaps remain. Fresh output is in
`../validation/landing-c799/style.log`; these timings include CPU contention
with the other owned suites and are observations rather than a performance
regression claim. Checker PASS 0.187 s, filtered Node PASS 6.683 s, focused
lowering PASS 1.693 s, focused record/Node/mutant checks PASS 44.223 s, vet clean.


## Final integration-area landing

Rebased onto area `d3a37422c`, including current main `b6b1538b0` and its
native typeof correction. All seven earlier wave-21 suites PASS 1047.698 s;
JSX package PASS 1187.016 s; style package PASS 166.135 s. Independent Go bytes,
all seventeen rule mutants, sanitizers, lifetimes, source/emitted Node, named
listener metadata and focused main typeof/Node checks pass. JSX syntax is now
supported; the keyword-label input and production project/checker context remain
shared gaps. HIR/SSA/capture reservations remain parked. Required pinned-input
checks report 16 PASS, one ABSENT, zero SKIP: TestSplitTSGoAgrees has not landed
on main or this area. No full repository or green 17-check gate is claimed.
No new reservations are taken. Exact logs, mutants, commands, timings and scope
limits are in [../validation/landing-area/REPORT.md](../validation/landing-area/REPORT.md).
