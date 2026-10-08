Built: original-declaration certification of emitNode, JSDoc heritage and JSDoc owner reads on the bounded walk; no compiler change since the Bindable batch.
Commit: this batch follows the Bindable commit; its sha is recorded by the branch tip. Delivered as a format-patch (push 403).
Commands: all original pair tests PASS (Pairs, Bindable, Nodes); vet and gofmt clean; compiler unchanged, so the batch 1 checked-view and package runs stand.
Mutants: 16 more (emit skip, shape, nested, presence; class skip, shape, nested twice; jsdoc and root skip, shape, nested), each caught by three pins.
Not covered: 7 pairs / 10 reads, each with its reason below; a JavaScript optional-chain divergence found in shared code.

Working date: October 11, 2026. Lane 7 static candidates: 17 object pairs / 66 reads.
Certified with complete pinned declarations: 10 pairs / 56 reads.
- Earlier: 10236 SymbolTracker.moduleResolverHost (1).
- Bindable batch: 9476 left (9), 9474 expression (11), 9485 left (9), 9475 expression (4).
- This batch: 7612 GeneratedIdentifier.emitNode (4), 8883 JSDocAugmentsTag.class (4),
  8882 JSDocImplementsTag.class (3), 7642 JSDoc.parent (10), 36241 (JSDoc | undefined).parent (1).

Remaining, 7 pairs / 10 reads, each observed rather than assumed:
- 68704 SymbolTrackerImpl.moduleResolverHost (4): SymbolTrackerImpl is a
  non-exported class in checker.ts; the emitted declarations contain it zero times,
  so there is no original declaration to certify against.
- 9454 argumentExpression (1) and 9657 class.expression (1): the contract is held
  by a shared descriptor, but no admitted program reaches either read without an
  earlier read checking the same value at the same depth (the direct intersection
  cast and the unproven type predicate are refused). A skip mutant there could not
  be caught, so they are not counted.
- 92175 (JSDocImplementsTag | JSDocAugmentsTag).class (1): the read's declared
  type, a union of two identical intersections seen through a union receiver,
  has no interned contract, so the read checks only that the value is an object.
  Demand does not refuse it. Output still matches Node because deeper reads are
  checked, but the read is weaker than its declared intersection. The fix belongs
  in shared lowering (readObjectField interns no contract here); not certified.
- 7644 (JSDoc | JSDocTypeLiteral).parent (1): Node | HasJSDoc has an open-kind
  Node arm no tag separates; refused as union intersection at the read. Checking
  Node alone would be sound, since every HasJSDoc member is a Node, but needs a
  supertype-absorption rule decided with the checker. Next candidate.
- 97923 and 98493 fileInfos.forEach (1 each): stage 0 says NotYet, "can't lower
  an array of IncrementalMultiFileEmitBuildInfoFileInfo yet" (string | Omit<...> & {...}),
  before any view check.

Fixtures (original/emit-original-*, class-*, jsdoc-parent-*, jsdoc-root-*) import
the complete emitted declarations. Each wrong fixture is read only by the read its
mutant weakens: the emitNode fixtures return flags, so a wrong autoGenerate.id or
source map range is seen only by the emitNode read; the class fixtures read only
node.class; the JSDoc owners are an empty statement and a complete
FunctionDeclaration. Exact exit 70 messages: autoGenerate not initialized (the
intersection forbids EmitNode's undefined), a string id, a string range pos, a
heritage kind of 999, this as a heritage name, a heritage name without a Symbol,
string type arguments, an owner kind outside the 64 HasJSDoc kinds, an owner
without parameters, an owner name without a Symbol.

Mutants, run by original/run-pair-mutants.py (now covering both tests). Each
fixture then runs to valid output and exit 0 in all three modes:
- emit-skip, emit-shape (id accepts string), emit-nested (SourceMapRange forgets
  pos), emit-presence (autoGenerate optional).
- class-skip, class-shape (kind accepts 999), class-nested (Identifier forgets
  symbol), each on the augments and on the implements fixture.
- jsdoc-skip, jsdoc-shape (EmptyStatement accepts 999), jsdoc-nested
  (FunctionDeclaration forgets parameters); root-skip, root-shape, root-nested for
  the optional-chain read.

Found while proving root-skip: a plain tagged union read through an optional chain
whose receiver is undefined panics in the JavaScript backend only. probes/
plain-union-optional-chain.a has no intersections: Node and native release print
"false true", the JavaScript backend exits 70 with "get(found)?.item.kind is not
initialized". Observed on 9e4e13fb and on views-integration 6a7f1bf3. The shared
union discriminant wrapper reads kind without an undefined guard; native guards
NULL. Not in lane 7's files, so not fixed here. The 36241 wrong fixtures call
only the present case so their counterfactual is valid in all three modes.

Commands, output in nodes-evidence:
- node original/prepare.cjs: 78 declarations, 77 spans, pairs 1+4+9+11+9+4+1+4+3+10+1.
- go test ./internal/oracle -run '^TestCheckedViewIntersectionOriginal' -count=1 -v: see final-original.log.
- python3 original/run-pair-mutants.py <decls> <logs>: 28 runs (12 Bindable, 16 here);
  see mutants-all.log.
- Integration probe: a worktree at 6a7f1bf3 with the same scratch probe test.
