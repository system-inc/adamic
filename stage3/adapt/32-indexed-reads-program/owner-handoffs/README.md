# Owning-declaration handoffs

On merged integration 86cc9b7, parser.ts is **1 -> 0**,
transformers/utilities.ts is **3 -> 0** and factory/emitNode.ts is **2 -> 0**, all diagnostic codes checked by the
unchanged latent run 0 feature compiler. All 79 roots are checked (hostErrors.ts
is the extra root); **50 -> 53 of the original 78 files** are zero. Fair whole-tree
count is **329 -> 321**, with installed dependencies identical and no new
findings. The initial pre-install 330 count includes a missing source-map-support
package and is not credited to this adaptation. Only types.ts changes.

The six requested owners and upstream proposals are individually recorded in
../handoff-sites.json. AmdDependency.name admits undefined because the parser
always constructs the own name property from the optional AMD pragma name.
AllDecorators.parameters admits outer undefined because the class, accessor and
method result literals explicitly store it when parameter decorators are absent
or legacy decorators are disabled, while member decorators retain the result.
CommentRange.hasTrailingNewLine admits the undefined optional argument that
both synthetic-comment producers store in an own property. AutoGenerateInfo
prefix and suffix admit the optional inputs both generated-name producers
store in own properties. These five applied declarations change no objects,
reads, stores or Node computation. The two AutoGenerateInfo contracts reduce
factory/nodeFactory.ts **4 -> 2**, leaving the generic-write findings at
localSymbol and typeExpression; no zero is claimed for that file.

RawSourceMap.sourcesContent's proposed element union is truthful but deferred.
Applying it alone makes the upstream stock build fail at sourcemap.ts:216, where
setSourceContent still accepts string | null. The unchanged local storage also
excludes undefined. Partition 33 owns those two declarations and has not yet
landed their matching unions. This unit does not edit its source or adaptation.
Sourcemap.ts therefore remains **one TS2345**, with the exact remainder in
after.json and the failed stock-build log retained. Once the companion owners
land, the proposed RawSourceMap union can be promoted from its declined ledger
entry under the same proofs. No assertion is justified: caller-owned getter
reads can return different content, and the final undefined payload is stored.

Ten parsed producer bodies are protected by stock-emitted runtime hashes:
processPragmasIntoFields; appendSourceMap and setSourceContent; and the three
decorator result producers; the two synthetic-comment producers; and the two
generated-identifier producers. These other partition files are read only. The
adapter parses stock 6.0.3 current text, checks each exact optional owner and its
public/internal visibility, and plans all 26 files before any write. Complete-
tree reruns recognize only the seven pre-existing adaptation 47 Buffer
assertions recorded in sites.json; this unit inserts none of those assertions.

handoff-api.cjs composes adaptation 70's exact parsed-owner proof with 20's
189 optional API owners, 40's exactly 28 sanctioned owners, 70's proven public
readonly view, and precisely two new API owners: AmdDependency.name and
CommentRange.hasTrailingNewLine. The API diff is exactly 220 changed lines.
AllDecorators, AutoGenerateInfo and RawSourceMap are internal and absent from
the public snapshot. The complete emitted API must equal the reconstruction byte for byte;
all 60,930 other reference baselines remain pristine.

Default oracle: **106,367 passing, zero failing, empty baseline diff** after that
exact mechanical API acceptance. All 26 partition files emit byte-identical
JavaScript against untouched input; CRLF and a zero-edit second run pass.
Owner-removal mutants restore parser's one finding and decorator utilities'
three findings and fail their owner guards. Removing the comment owner restores
both emitNode overload findings. Removing either generated-name owner restores
both corresponding nodeFactory contracts (four total including the two
unrelated generic-write findings). Replacing an existing watch required
! with ?? 0 fails both emitted bytes and the indexed ledger. A wrong producer
invariant hash is rejected before any write. An unlisted API declaration fails
full snapshot reconstruction. Every mutant was run, with results retained.

Watch's two public optional host findings remain queued separately. Tracing's
four remaining Node host bindings are unchanged; adaptation 47 already supplies
its catch helper. PerformanceCore's require bindings remain skipped as directed.
