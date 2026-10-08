Built: declaration-backed witnesses for SourceFile.externalModuleIndicator and Diagnostic.messageText, plus correct optional absence handling.
Commits: builds on f4e5993e and integration ba59427c; implementation tip is the commit containing this report.
Commands/results: selected production gate passed, oracle 49.176s; all nine original cases passed; vet and diff check passed.
Mutants: outer removal, wrong-member/shape acceptance, transitive removal, readiness loss and rejecting allowed absence caught in both release backends.
Uncovered: 40 candidate pairs / 146 candidate reads remain; exact whole-tsc execution reachability remains unmeasured.

Two original field-contract pairs are now certified by standalone source witnesses:
SourceFile.externalModuleIndicator (20 candidate reads) and Diagnostic.messageText
(15 candidate reads). The original target and object-member declarations are
imported in full, including inheritance, qualifiers, recursive references and
unsupported unread descendants. No reduced Node or Chain declaration is used.
This certification covers the original field contracts with representative source
reads and propagation controls. It does not mean the 35 original enclosing tsc
functions were compiled or executed. The 35 sites were audited byte for byte at
their UTF-16 positions against the pinned upstream tree. Exact production
reachability still waits for a checker-clean compiler.

The TypeScript tree is fetched from microsoft/TypeScript at the repository's
stage3/source.json pin 050880ce59e30b356b686bd3144efe24f875ebc8. No code is copied
from cohere. prepare.cjs rejects tracked source modifications, verifies all
recorded sites, and emits 78 declaration files outside Adamic. Its manifest hashes
all emitted files and records the exact complete field sets for SourceFile, Node,
Diagnostic and DiagnosticMessageChain. The oracle checks those hashes and rejects
missing/reduced field sets in the lowered contracts. Declaration emission is not
whole-program source compilation; each bound witness and all its declarations
pass ordinary Adamic load diagnostics. No diagnostics are disabled in that path.

Nine source cases cover both valid union alternatives, false outside the true
literal contract, invalid Diagnostic boolean, wrong nested SyntaxKind/code,
explicit undefined, absent optional property, uninitialized present property,
and SourceFile generic helpers/callbacks/stored aliases. Both backends pin field,
expected, found and exit 70. Native positives include ASan/UBSan and leak checks.

Original SourceFile declares externalModuleIndicator optional. The prior reduced
control declared it present with an undefined alternative, so it did not expose
missing optional storage. The owned C/JS emitters now pass field absence and
receiver optionality separately. The native helper allows a missing optional slot
before inspecting payload; a present uninitialized slot still enters the common
readiness refusal. No shared file changes were necessary in this group.

Independent release mutant observations (native / JavaScript):

- SourceFile wrong-member acceptance: false / false, exit 0.
- SourceFile outer-check removal: absent / false, exit 0; helper/callback variant
  has the same independent result.
- SourceFile wrong nested kind acceptance and transitive removal: 0 / false,
  exit 0, independently failing the SyntaxKind refusal pin.
- Diagnostic wrong-member acceptance and outer removal: wrong / wrong, exit 0.
- Diagnostic wrong nested code acceptance and transitive removal: chain:0 /
  chain:false, exit 0, independently failing the number refusal pin.
- Lose readiness on the present optional slot: absent / absent, exit 0, failing
  the named uninitialized refusal pin.
- Reject allowed absence: both stop at exit 70 after true/80/absent; the positive
  four-value Node pin catches this erroneous refusal.

Validation (output retained in logs):

```
git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /tmp/object-primitive-upstream
npm ci --prefix stage3/api --ignore-scripts --no-audit --no-fund
node stage3/adapt/00-setup/adapt.cjs /tmp/object-primitive-upstream
node stage3/interface-downcasts/lane4b/original/prepare.cjs /tmp/object-primitive-upstream /tmp/object-primitive-original-declarations
ADAMIC_OBJECT_PRIMITIVE_ORIGINAL_DECLS=/tmp/object-primitive-original-declarations ADAMIC_GATE_UNCACHED=1 go test ./internal/lower ./internal/native ./internal/javascript ./internal/oracle -run 'TestCheckedViewObjectPrimitive|TestCheckedViewLane4HelperReads|TestCheckedViewMixedSelection|TestLazyView|TestCheckedViewArrays' -count=1 -v -timeout 10m
go vet ./internal/lower ./internal/native ./internal/javascript ./internal/oracle
git diff --check
```

The original suite is opt-in to its generated external declarations; it ran with
that environment set and no original cases skipped. Native and JS packages had
no matching package-local tests; their emitters ran through the oracle. Opt-in
whole-program lazy census tests skipped. This is a selected gate, not the full
repository gate. Existing reduced and array controls passed alongside originals.

The next original pair in the inventory is the overlapping object/intersection
family (11 reads), followed by LiteralType.value (10). Neither is certified here.
The working whole-family target remains October 13, 2026, 23:00 UTC. Candidate
coverage is tracked in certification.json and the ranked lazy-pair-progress.json;
completed_reads denotes candidate reads attached to witnessed pairs, never
measured whole-program runtime reads.
