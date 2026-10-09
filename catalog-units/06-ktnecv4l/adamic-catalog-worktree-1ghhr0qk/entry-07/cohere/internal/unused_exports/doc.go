// Package unused_exports reports code that was written and never used.
//
// This is a report someone runs on purpose, not a gate that fails a build. The distinction decides
// every judgment call in here: a gate must not report live code, so it would have to stay silent
// wherever it is unsure, and a report that stays silent wherever it is unsure finds nothing worth
// reading. So the bias here is different — say what was found, say honestly how confident it is,
// and give the reader a way to record a decision so the second reading is shorter than the first.
//
// Two analyses live here and they answer different questions:
//
//   - Never referenced: nothing anywhere imports or calls this. Needs a program-wide reverse index.
//   - Never reachable: this cannot run, whatever calls it. Needs only the control-flow graph.
//
// They are separable on purpose. Reachability is decidable per function and is as trustworthy as
// the CFG. Reference analysis is a whole-program claim that is only as good as its root set, and a
// wrong root set reports a homepage as dead code.
//
// # What counts as a reference
//
// This list is normative and it is written down because the next reader cannot recover it from the
// code. `typescript/no-unused-vars` is being ported, and oxc spends roughly 990 lines in `usage.rs`
// deciding this same question — self-references, a variable only reassigned to itself, aliases,
// re-exports, type-position uses — with every answer measured against a large corpus. The answers
// below are NOT measured against a corpus. They are measured against this substrate and this tree,
// which establishes that they behave as described but not that they are the answers a corpus would
// have chosen. So each one is marked MEASURED or JUDGMENT, and the JUDGMENT entries are exactly
// where this analysis and a future no-unused-vars could disagree forever without either noticing.
//
// A reference is recorded when an identifier resolves through the checker to a symbol, and the
// identifier is not that symbol's own declaring name. Specifically:
//
//	MEASURED  Any identifier position at all. The walk is over identifiers rather than over import
//	          statements, so a call, a type annotation, a JSX tag name, a decorator expression, a
//	          property initializer and a default-export expression are all references without any of
//	          them being enumerated. Probed on the fixture: 18 of 18 reference sites resolved.
//	MEASURED  An import binding AND what it aliases. An import specifier is a symbol in its own
//	          right whose `SymbolFlagsAlias` points at the real declaration in the other file, so
//	          both are recorded. Recording only the alias leaves every real declaration looking
//	          unused; the original probe measured 42 of 42 imported names resolving through
//	          `GetAliasedSymbol` with zero nils.
//	MEASURED  A reference from inside a function to something in an enclosing scope. The enclosing
//	          declaration is the innermost containing function, class, method, accessor, constructor,
//	          property, variable, module, or export assignment, so the edge is attributed to the
//	          nearest thing that could itself die.
//	MEASURED  A reference at module top level. This has NO enclosing declaration and is recorded as
//	          an edge from the file's module root, which is a ROOT of the closure rather than an
//	          absent edge. Getting this backwards makes either everything look dead or everything
//	          look alive, and both readings are silent.
//
// And these are the calls that are mine rather than measured against anything:
//
//	JUDGMENT  A symbol referenced ONLY BY ITSELF is not referenced. A recursive function nothing else
//	          calls produces an edge from itself to itself, and the closure never marks it because
//	          nothing marks it first. This is deliberate and it is the whole reason the closure is
//	          worth more than the flat set: the flat set counts the recursive call as a use and calls
//	          a dead function alive. oxc's `usage.rs` decides the analogous case for a variable only
//	          reassigned to itself, and reaches the same answer for the same reason, but it decides it
//	          on a corpus and this decides it on an argument.
//	JUDGMENT  A re-export is a PASS-THROUGH, not a use. `export { AmexIcon } from './AmexIcon'` does
//	          NOT mark `AmexIcon` referenced, so an export reachable only through a barrel nobody
//	          imports is reported dead. Measured on a two-file fixture rather than reasoned about,
//	          because the opposite was assumed first and was wrong: the mechanism is that
//	          `isDeclarationName` classifies an ExportSpecifier's identifier as a declaring name, so
//	          the walk never records it.
//
//	          This is the answer this analysis wants and it is worth stating why, because it is the
//	          one most likely to look like a bug. A barrel that re-exports forty symbols and is
//	          imported by nothing is not forty live symbols; it is a dead barrel. Counting the
//	          re-export as a use would make any barrel an unconditional root and would silently
//	          protect everything behind it — the failure mode where the report goes quiet and looks
//	          clean. Measured on this tree, this rule is what correctly reports four genuinely
//	          abandoned payment icons whose ONLY mention anywhere is a barrel nothing imports.
//
//	          The cost is real and paid knowingly: a barrel that IS imported re-exports through
//	          `import { X } from './barrel'`, and that import IS a reference which resolves through
//	          the alias chain to the original declaration, so live barrels behave correctly. Only an
//	          entirely unimported barrel produces this, and in that case reporting its contents dead
//	          is the right answer rather than a false positive.
//
//	          Re-checked against upstream and the answer stands, because the two tools are not asked
//	          the same question. `no-unused-vars` treats `export { X } from './x'` as clean, and it is
//	          right to: it asks whether a binding is referenced within its scope, and a re-export
//	          references it. This asks whether a symbol is REACHABLE FROM A ROOT, and an unimported
//	          barrel does not make it reachable. Both answers are correct for their own question, so
//	          a disagreement here is not evidence that either is wrong.
//
//	          Both halves are pinned: `TestReExportIsAPassThrough` asserts the unimported barrel's
//	          contents are reported, and probing the imported-barrel case on a live binary confirms
//	          the alias chain spares them. The header's warning that this analysis and a future
//	          no-unused-vars "could disagree forever without either noticing" is the thing to hold
//	          onto — this is that disagreement, and it is the benign kind.
//
//	          A barrel's own `export ... from` statements are additionally not examined AS exports:
//	          `exportedDeclarations` handles declaration statements and an ExportDeclaration is not
//	          one, so a pure barrel file contributes zero exports and never appears in the report
//	          itself. That is a genuine gap — a dead barrel is invisible while its contents are
//	          reported — and it is written down rather than fixed, because fixing it means deciding
//	          what a barrel's "own" liveness means and that decision belongs with the no-unused-vars
//	          port rather than ahead of it.
//	JUDGMENT  A type-only reference counts as a use. `import type { T }` and a `T` in a type
//	          annotation mark T referenced exactly like a value would. A type nothing uses at runtime
//	          is still a type someone reads, and treating type positions as non-uses would report
//	          every interface in a types file. oxc separates these behind an option; this does not.
//	MEASURED  A reference from inside code that is ITSELF unreachable DOES count, in the flat set and
//	          in the closure alike. This entry previously claimed the opposite — that the closure
//	          discounts such a reference, and that discounting it was what the closure bought — and
//	          that was never true of the code.
//
//	          `closure.go` has no reachability input; the word does not appear in it. `Unreachable`
//	          is a report field, collected by a separate analysis and printed, and never read back
//	          into the reference walk. The header above says the two questions are "separable on
//	          purpose", which is right about the design and is precisely why this entry could not
//	          have been true: nothing joins them.
//
//	          Pinned by `TestDeadCodeReferenceStillCountsAsAUse`, on a fixture where reachability is
//	          the only difference between two constants — same importer, one reference each, both
//	          reached from a framework-spared page. Both are spared. The unreachable statement is
//	          still reported, by the analysis that owns that question.
//
//	          What the closure actually buys is transitivity: it finds a cluster that is dead
//	          TOGETHER, which the flat set cannot see because each member references the others. That
//	          is a strictly stronger claim than the flat set makes and a wrong root cascades through
//	          it rather than mis-reporting one file, which is why it stays opt-in behind
//	          `--unused-deep`.
//
//	          Whether it SHOULD discount unreachable references is open and is not a documentation
//	          question. Doing so would mean the closure's answer depends on the CFG's, so a CFG bug
//	          would silently delete live code from the report — the direction that costs trust. The
//	          entry is recorded as measured behaviour rather than resolved policy.
//	JUDGMENT  A default export is not examined as a finding, though references FROM inside one are
//	          recorded normally. The name at the declaration site is not the name at the use site, so
//	          the symbol identity this index is keyed on does not carry across reliably, and a false
//	          positive on the most load-bearing kind of export there is would be the one that gets
//	          this report closed. Stated as a known gap rather than a solved case.
//	JUDGMENT  A property access (`object.method`) counts as a reference to the property symbol. This
//	          falls out of walking every identifier rather than being decided, and it means a
//	          string-keyed lookup (`table['method']`) does NOT count. A dynamic dispatch table is
//	          therefore invisible to this analysis, which is a real gap and the reason the report is
//	          a report rather than a gate.
package unused_exports
