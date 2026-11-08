# Scope and stream corrections

The entry-reach expansion was independent of RelationOverflow. It was added to
catalogue every source reachable from a file argument during the earlier broad
tsc-entry attempt. Overflow detection does not require it: the checker already
checks imports and keeps their relation state. The expansion, its loader method
and its test have been removed. The README describes the separate coverage
issue and requires a core.ts before-and-after census for a future change.

The original measurement header is now encoded and synced before lower.Latent.
Completed file records follow. Only after normal completion is a final record
with `status: "relation_report"` written, containing both endpoint answers and
program findings. A partial stream cannot manufacture an empty end answer.
JSONL readers keep that trailer out of file/unit counts; completed relation
metadata is merged into a copied header for legacy reader compatibility.

Validation:

- Complete latent overlay suite: exit 0.
- Complete stage3/census/... Go suite: exit 0; all 12 package results are in
  package-results.json.
- Pure JSONL reader tests: 3 cases, exit 0.
- Basic census audit: exit 0.
- Removed-RelationOverflow-call mutant: exit 1;
  TestLatentCensusOverflowMarkReachesProgramAndEveryAttempt reports
  "planted mark disappeared".
- Real full-mode checker-only cutoff smoke: the child was deliberately sent
  SIGTERM after its header became visible (child exit -15). Exactly one header
  remained, with no end answer or completion trailer. Assertions exited 0.

New Go tests call t.Parallel() first and start no compiler or go subprocess:

- TestLatentStreamCutoffRetainsHeaderWithoutEndAnswer
- TestLatentStreamCompletesWithRelationTrailerAfterFiles

The requested fact run will use LATENT_FULL=1, LATENT_ONLY_FILE equal to the
absolute adapted checker.ts, LATENT_RELATION_PER_UNIT=0, and the compiler source
directory as the load input. Its wall bound is 6000 seconds (100 minutes). The
file filter keeps the original full compiler checker/binding context while
lowering only checker.ts. The fixes are pushed before that run starts; the
answers will be a separate commit.
