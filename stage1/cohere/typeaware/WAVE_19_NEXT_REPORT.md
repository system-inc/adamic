Built: no additional rule ports; the original three remain completed and pushed.
Commits: original completed tip c5a19fe0; continuation claim 3fc985765422fb29b2438cd105ec821741703de2.
Commands and outputs: fetched all origin heads successfully; inspected 325 origin refs, 197 ranked rules and claim documents; confirmed claim SHA on origin with git ls-remote.
Mutants: none for this continuation, because no implementation was written or claimed tested.
Uncovered: all three continuation rules are blocked by missing bridge facts whose registration requires a shared-file edit forbidden by Ahra's correction.

## Continuation selection

The first three available rules in VOLUME_REPORT.md's linked all-family tables,
ranked by combined compiler and repository findings descending with lexical ties,
are:

- nexus/correctness-no-process-exit-after-output
- nexus/correctness-no-uncleared-race-timeout
- nexus/correctness-require-blocking-standard-streams

All three have zero findings in the recorded selection populations. Checked ports
on origin/main and origin/codex/tsgo-c-library, and rule names in claims on every
fetched origin ref. These rules were not already ported or claimed at that
snapshot. The continuation claim was committed and pushed before implementation.
The original three ports are unchanged and remain validated as recorded in
WAVE_19_REPORT.md. No further rules were claimed.

## Exact blocker

Ahra's correction says: "Keep your changes inside your own rule directories.
Don't edit the shared registration generator or the test harness." It also says:
"If anything else blocks you, say exactly what it is and stop, rather than
editing shared files."

The bridge's existing declaration-details/node-symbol-details frames return one
parent kind, name and span. They do not expose complete declaration ancestry,
source-file external-module status or global-scope-augmentation status. These
facts are required to distinguish ambient globals from module locals and the
actual NodeJS.Process interface from another interface named Process:

- no-uncleared-race-timeout uses IsGlobalScopeAugmentation on the timer
  declaration's enclosing module and IsExternalModule on global script sources.
  Its pinned tests explicitly include a module containing declare global and a
  merged setTimeout namespace. Guessing from filename or spelling would change
  the rule's semantics.
- no-process-exit-after-output additionally checks ancestors above the Process
  interface, and follows the declaration of a resolved call signature into a
  body. Current signature frames provide return/argument types, not that
  declaration's identity and body metadata.
- require-blocking-standard-streams additionally needs every source file in the
  checker program and exact module-resolution edges, including re-exports,
  import-equals and dynamic imports. No existing question exposes that graph.

New question implementation files alone cannot make these facts callable.
Program.Inspect in bridge/tsgo/checker/facts.go selects every question through a
shared switch and returns "unsupported checker question" for an unknown name.
Registration therefore requires editing that shared file. The original unit
allowed one-line registrations; this continuation honors Ahra's later correction
restricting edits to owned rule files and requiring a stop on other blockers.
No shared-file change, temporary registration workaround, spelling-only fallback,
or silent omission was introduced.

This is a bridge registration/fact-access blocker, separate from the shared
harness's pending .a and emitted-JavaScript work. The continuation rules have not
been ported, compiled, differentially compared, mutated or timed. Their claim
remains explicit as blocked, so it cannot be mistaken for completion. Work can
resume when the bridge registration path permits these isolated fact modules.
