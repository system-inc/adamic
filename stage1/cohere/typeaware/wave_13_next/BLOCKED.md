Built: original three wave 13 ports remain complete and pushed; no continuation rule implementation was written.
Commits: original evidence tip 7ca1f20e; continuation claim d37a6984 was pushed before implementation.
Commands and outputs: fetched all origin heads successfully; scanned 325 refs, 197 ranked checker rules, 25 existing checker ports and 96 claimed ranked rules.
Mutants: original validation is unchanged; no continuation mutants or agreement/sanitizer claims are made.
Not covered: all three continuation rule ports, pending callable bridge metadata within the authorized scope.

Selection

Used VOLUME_REPORT.md's linked compiler-all.counts and repository-all.counts,
summing counts, descending volume and lexical ties. Excluded production rule
names found in native sources on origin/main and origin/codex/tsgo-c-library,
including short @typescript-eslint names used by Rules.add. Scanned Markdown
claim files beneath stage1/cohere/typeaware/claims/ on every origin ref, matching
whole rule identifiers. The first three remaining rules are:

- nexus/correctness-no-process-exit-after-output, combined volume 0
- nexus/correctness-no-uncleared-race-timeout, combined volume 0
- nexus/correctness-require-blocking-standard-streams, combined volume 0

Their claim update was pushed as d37a6984 before implementation. No additional
rules were claimed, and these three are explicitly incomplete.

Observed bridge gaps

Read the production Go implementations and the current bridge fact dispatcher.
The dispatcher is a closed switch in bridge/tsgo/checker/facts.go. Adding a Go
helper in a separate file alone does not expose a callable tsgoInspect question.
Existing node-symbol-details and declaration-details return declaration path,
kind, span, declaration/default-library booleans, immediate parent kind/name/span,
JSDoc and parameter names. They do not return the facts below. Existing signature
and signature-shape return parameter/type graphs, not the selected signature's
declaration. No current question enumerates the compiler program's source files
and their resolved import edges.

Required raw fact interfaces, with all lint judgments staying native:

1. Source/declaration ancestry metadata: external-module status, global-scope
   augmentation status and named declaration ancestor chain. The race-timeout
   rule accepts ambient global setTimeout declarations, including declare global
   and default-library WindowOrWorkerGlobalScope, and rejects imported or
   module-scoped lookalikes. The output rules similarly require every declaration
   of NodeJS.Process members and global console to have the correct ancestry.
   File-name heuristics or spelling-only checks would weaken Go's predicate.
2. Resolved call declaration metadata: the selected signature declaration's
   file, kind, span, body span, generator/async flags and return type. The
   process-exit-after-output rule follows exactly one callee level, declining
   overload signatures without bodies, generators, unawaited async bodies,
   external global scripts and never-returning callees. Type-only call facts
   cannot decide those cases.
3. Program source inventory and resolved module edges: all program files,
   declaration status and exact resolved targets of static imports, reexports,
   import-equals, literal dynamic imports and require calls. The blocking-streams
   rule computes inbound imports and transitive reachability to Nexus's
   source/system/StandardStreams.ts or .d.ts. The explicit lint root manifest is
   not the whole compiler program or its resolved import closure, so substituting
   it would change the rule's entry and blocking decisions.

These facts would belong in dedicated named Go and Adamic files, with one
minimal dispatcher registration for each. No Go lint verdict or finding should
cross the bridge.

Scope blocker and action taken

Ahra's correction says: "Keep your changes inside your own rule directories.
Don't edit the shared registration generator or the test harness" and "If
anything else blocks you, say exactly what it is and stop, rather than editing
shared files." The later continuation request permits dedicated new bridge
question files but does not explicitly lift the shared-file restriction. The
missing callable facts require shared facts.go dispatcher registration, outside
that restriction. This is a bridge integration blocker, not a claim that the
three production algorithms are impossible to port.

Stopped without modifying the dispatcher, registration generator, existing test
harness, protected compiler files or submodule pins. No stub that returns zero
findings was written: these zero-volume corpora would otherwise make an empty
implementation appear to agree. No continuation byte-parity, mutant, sanitizer,
released-handle or timing result is claimed. The existing three completed ports
and their prior evidence remain unchanged. The shared harness branch
origin/codex/lint-harness-dot-a was not present among the fetched refs in this
snapshot, so its integration state was not assumed.
