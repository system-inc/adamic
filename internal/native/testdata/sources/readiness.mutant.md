Mutant evidence for the native readiness source and its C harness.

The source has no Adamic-program caller yet, so harness.c pins it directly, as
testdata/timers pinned the timer provider. Each source is a pipe made
non-blocking, written before the loop or left empty. A raw descriptor has no
Node oracle, so sources_test.go holds each finishing mode's whole stdout.

Controls: TestReadinessSourceModesFinish runs every finishing mode under ASan and UBSan, checks
balanced value counts and the shared leak check, then compares release.
TestSourceThrowIsUncaught, TestSourceCustomHooksRefused and
TestSourceBlockingDescriptorRefused pin the stopping modes by their exact text.
Builds are TestProduct_SourcesSanitized, TestProduct_SourcesRelease,
TestProduct_SourcesCounted and one TestProduct_SourceMutant unit per row below.

Command after products: go test ./internal/native -count=1 -run 'Source' -v

TestSourceMutants applies each real production edit below to a copy of the
runtime (timerMutant), links the same harness in release, and asks the named
pin's own check, sourceFinishes, which must fail by the cause named. Every
mutant must compile; a compiler failure never counts as a caught mutant.

| Mutation | Mode | Named pin | Caught by |
| --- | --- | --- | --- |
| The wait's predicate ignores sources | self-close | TestSourceSelfCloseExits | its own 3 s deadline |
| Liveness counts unreferenced sources | unref | TestSourceUnrefExits | its own 3 s deadline |
| Readiness read live, never a snapshot | fairness | TestSourceNeverDrainedYieldsToInterval | its own 3 s deadline |
| No checkpoint after each dispatch | checkpoint | TestSourceDispatchIsAMacrotask | stdout comparison |
| The pending close counted as done at once | pending-close | TestSourcePendingCloseDispatchesUnrefSibling | stdout comparison |
| The pending close's wait takes the default timeout | pending-close-unref-timer | TestSourcePendingCloseWaitNeverBlocks | its own 3 s deadline |
| The pending close kept until a wait finds no source ready | pending-close-once | TestSourcePendingCloseKeepsOneTurn | stdout comparison |
| A refness flip pends a close | ref-flip | TestSourceRefFlipKeepsNoPendingClose | stdout comparison |
| A stale ref answers true | ref-stale | TestSourceRefAfterCloseIsStale | the harness's assertion, a failed exit |

The edits, each in one site:

- predicate-ignores-sources, async_host_impl.h: `timeout < 0 && source_total == 0);`
  becomes `timeout < 0);`, the predicate before sources. A lone source makes no
  request and no timer, so the wait returns without polling, the source is never
  recorded ready, and the loop spins on a live source it never reads.
- liveness-counts-unref, sources_impl.h: `if (entry->referenced) return true;`
  becomes `if (entry->active) return true;`. The lone unreferenced source on an
  empty pipe keeps the loop alive, and the wait polls it forever.
- readiness-read-live, sources_impl.h: the dispatch's `chosen->ready = 0;`
  becomes a zero-timeout poll that keeps the bit while the descriptor reads
  ready, so the phase dispatches every source until none is ready. A source
  whose dispatch never drains is chosen again forever, and the interval beside
  it never fires a second time.
- dispatch-checkpoint-omitted, async.c: `while (source_dispatch()) drain_jobs();`
  becomes `while (source_dispatch()) {}`. The first source's job runs only at
  the host completions' drain, after the second source's dispatch.
- close-done-at-once, sources_impl.h: adamic_source_close's
  `source_pending_closes++;` is dropped. The closed referenced source was all
  that kept the loop, so it ends before any wait, and the unreferenced
  sibling's byte is never dispatched.
- pending-wait-default-timeout, async_host_impl.h: host_wait_ms's
  `return source_closes_waiting() ? 0 : timer_wait_ms();` becomes
  `return timer_wait_ms();`, the call kept so the mutant compiles. The pending
  close's turn waits on the unreferenced 10 s timer beside it.
- pending-until-quiet, sources_impl.h: source_record moves the pending closes
  to polled only when it recorded no source ready. The sibling, readable again
  after each of its first two dispatches, is dispatched three times.
- ref-pends-close, sources_impl.h: adamic_source_ref also counts a pending
  close, as if a flip were a close and a reopen. The unref, ref, unref flip
  then keeps a turn, and the unreferenced sibling's byte is read.
- ref-stale-admitted, sources_impl.h: adamic_source_ref's stale check answers
  true. The ref after close in the dispatch is admitted and the harness's
  assertion stops it.

The pending-close modes are Runtime's ruling on #j71fn59 (task #7s8hgd5),
after libuv's uv_loop_alive and uv__run_closing_handles: a close leaves a
pending close, which keeps the loop alive through the next wait that polls,
whose timeout is zero, and the source phase that dispatches its record.
adamic_source_ref, Runtime's later ruling on #j71fn59, is unref's mirror: a
flip either way is no close, and adamic_source_close is always a real one. The
same mutants turn internal/oracle's host_watch pins red against Node
(fswatch.mutant.md).

What each printed on Cloud, from TestSourceMutants at f57208ea08:

- predicate-ignores-sources: caught by TestSourceSelfCloseExits's deadline
  check: deadline: unfinished after 3s with stdout "prefix\n"
- liveness-counts-unref: caught by TestSourceUnrefExits's deadline check:
  deadline: unfinished after 3s with stdout "prefix\n"
- readiness-read-live: caught by TestSourceNeverDrainedYieldsToInterval's
  deadline check: deadline: unfinished after 3s with stdout "prefix\n"
- dispatch-checkpoint-omitted: caught by TestSourceDispatchIsAMacrotask's
  stdout check: stdout: "prefix\nfirst\nsecond\njob\n", want
  "prefix\nfirst\njob\nsecond\n"

A killed probe's buffered stdout is lost, so a deadline catch shows only the
prefix it flushed before the loop.
