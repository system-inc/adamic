# watchFile listener and lifetime evidence

The host binding is stat polling on the accepted timer provider, not a new loop.
The fixtures in internal/oracle/testdata/host_watch_file run against oracle/node.mjs,
the generated JavaScript, and ASan/UBSan native code with an explicit environment.

| Mutation | Named comparison that turns red |
| --- | --- |
| Make watch_changed return only its size comparison | TestWatchFileSameSizeMtimeMatchesNode (own bounded child deadline) |
| Suppress the initial missing-file notification | TestWatchFileInitialMissingMatchesNode |
| Keep a persistent:false interval referenced | TestWatchFileUnrefInitialStatMatchesNode |
| Cancel the wrong listener identity | TestWatchFileCancelOneMatchesNode |
| Drop timer_dispatch's adamic_uncaught call | TestWatchFileThrowingListenerMatchesNode |
| Drain microtasks inside the listener batch | TestWatchFileListenerBatchBeforeJobsMatchesNode |
| Replace the first interval on another listener registration | TestWatchFileFirstIntervalWinsMatchesNode |
| Admit timer_start under custom host loop hooks | TestWatchFileCustomHooksRefused |

The runtime mutations compile private archives; compilation failure never counts
as a catch. TestWatchFileMutant* runs the corresponding wrong program. The
custom-hooks fixture starts watchFile through its public native binding, after
adamic_host_set_loop_hooks, and pins exactly:

    adamic: panic: NotYet: timers with custom host loop hooks

With that refusal removed, the mutant admits watchFile under custom hooks, and its
unwatchFile reaches adamic_source_pending_close, which aborts with
"adamic: a pending close under custom host loop hooks: nothing would ever poll it"
rather than pend a close no custom wait would record (TestWatchFileMutantAdmitsCustomHooks).

Normal callback dispatch is one macrotask for the whole listener batch, followed
by microtasks. Initial stat remains live even for an unreferenced watcher. All
registrations and retained callbacks are released on cancellation or async teardown.
No recorded stage3 outcome is changed.

The live-registration throw probe also runs raw Node: exit 1 with the Error line.
Native pins adamic_uncaught's existing panic line and exit 70. The ordinary
source fixture unregisters before throwing so the unchanged house oracle's
uncaughtException handler has no live handle left and exits 70 as usual.

Polling intervals are not minimum delays measured from registration: libuv uses
cached loop time. The shared-interval probe counts notifications over a bounded
500 ms window instead of asserting a fragile first-callback wall time.
The probe and source-fixture bytes, mutation recipe and compiler/tool reports are
explicit product inputs; TestWatchFileProductKeyIncludesItsHarness pins that a
harness edit changes the key. HOME and temporary paths are not inputs.
