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
parsing still refuses a real source positive at byte 5: expected
GreaterThanToken, got Identifier in `<div style="bad" />`. Normal and sanitized
parser builds reproduce that refusal. Skipping parser.file makes the
prerequisite mutant exit 0 and is caught by the required refusal assertion;
this is separate from the rule mutant. The shared node/checker adaptation is
also absent on this main: the harness/context from `41eb6eab2` is not landed in
this checkout. The descriptor is a listener declaration, not a complete shared
registry factory. Source parsing, live checker queries and production
registration remain unimplemented. No bridge question or registration arm was
added in this batch; tests use the explicitly separate raw provider below.

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
