# Constructor family

Four explicit any tokens are removed: the Symbol links write and the Symbol,
Signature and SourceMapSource allocator bridges. Receivers are partial while
empty. Symbol writes every required field before it returns. SourceMapSource's
lineMap cache is optional, matching its allocation and getLineStarts' lazy fill.
Signature allocation now returns AllocatedSignature: flags plus optional checker
and partial remaining fields. createSignature fills every required field before
its final Signature completion assertion. These are staged shapes, not an empty
object asserted to already have its eventual fields.

Actual constructor bodies run on Node. Required fields are read independently
from the interfaces. Observations are identical before/after; dropping Symbol.id,
SourceMapSource.text or Signature.parameters makes the completion proof fail.
All three touched source files emit byte-identical JavaScript, and all ten built
JavaScript artifacts plus the public API are byte-identical to the timer family.

The census has 35 direct any sites before and after; its first allocator callback
still hides later callbacks. Refused falls 5157 -> 5156. Restoring only the actual
Symbol links any cast in a source copy returns Refused to 5157, with exactly one
extra unchecked-cast refusal. This catches the mutant even though the direct-any
headline cannot see the later body cast. Counts and exact site are in delta.json.

The lane passes: 106366 passing, one sanctioned API failure, zero pending,
222 existing declaration sanctions, identical to main. All 79 measured compiler
source files match the lane's fresh apply output. Stock tsc reports no diagnostics.

Inserted declarations exposed two composition guards: multiline rules needed the
file's newline convention, and the void adapter's fixed line moved. Exact anchors
now permit their text within the completed insertion while rejecting outside
duplicates. Void owners are matched by complete line, occurrence and parent/expression
AST guards. Idempotence and LF reconstruction pass. Real-source duplicate interface,
comment-drift interface and duplicate shifted void mutants all fail their guards.
The earlier failed lane was a discarded guard probe; the composed lane is the result.

Remaining allocator casts are Node, Token, Identifier, PrivateIdentifier,
SourceFile and Type, plus NodeLinks. Their advertised shapes still have absent
required fields or fields containing undefined. They are enumerated in the final
residue ledger; no fully initialized shape is substituted for those headers here.

Commands: stock tsc.js -p TREE/src/compiler --noEmit; family-proof.cjs;
constructor-proof.cjs; guard-proof.cjs; latent census and census-report.py on
control and actual links-cast mutant; bash stage3/lane/run.sh
/workspace/adaptation41-constructors-composed-lane. Outputs are logged and retained
under evidence/constructors/. No full Go-package gate or new fixture was added.
