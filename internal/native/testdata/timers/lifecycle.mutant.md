Mutant evidence for the native timer harness and its Node oracle (#j71fn59, cut 1a).

The .a file is a Node-only oracle, not an admitted source timer binding. Its
stage-0 header records the real compiler's TS2304 for unbound timer globals.
The body still runs unmodified through Node's vm as the behavioral oracle.

Controls: TestTimersAgainstNode compares every mode to Node, runs ASan/UBSan,
checks balanced value counts and the shared leak check, and compares release.
Builds are TestProduct_TimersSanitized, TestProduct_TimersRelease and
TestProduct_TimersCounted; comparisons and leak assertions run every time.

Command after products: go test ./internal/native -count=1 -run '^(TestTimersAgainstNode|TestTimerMutants)$' -v

TestTimerMutants applies the real production mutations below, using the same
harness. All must compile. A compiler failure never counts as a caught mutant.

| Mutation | Mode | Required red check |
| --- | --- | --- |
| Admit timers under custom hooks | custom-hooks | TestTimerCustomHooksRefused exact provider NotYet |
| Release callback before its final interval fire | interval | ASan heap-use-after-free |
| Ignore the provider brand during cancellation | foreign | The foreign-handle assertion |
| Ignore unref | unref | Node stdout comparison |
| Leak a cancelled registration | cancel | Shared leakcheck's allocation/leak proof |
| Free a callback while it cancels itself | cancel-self | ASan heap-use-after-free |
| Omit the post-timer promise checkpoint | order | Node stdout comparison |

The last mutation targets the new loop's actual post-timer drain, preserving
its original purpose after integration with B2b. Each mutated probe is built
once by its TestProduct_TimerMutant unit, and the ordinary test observes it.

TestTimerCustomHooksRefused installs custom hooks with adamic_host_set_loop_hooks
and pins the exact NotYet at adamic_timer_start; no lowering binding is needed.

Runtime review pins (code-only recut):

- turn-phase: a pthread queues a host completion while a one-millisecond
  interval callback stays busy for at least five milliseconds. Node's worker
  posts during the same busy window. Both produce prefix, interval, host.
  TestTimerTurnPhaseYieldsToHostCompletion has its own three-second deadline.
  Replacing entry->deadline <= turn_time with <= timer_now() keeps rearming and
  dispatching that interval, so the named test fails by its own deadline.
- throw: a callback sets adamic_thrown / throws Error("timer callback failed").
  TestTimerCallbackThrowMatchesNode pins raw Node's exit 1 and error line, then
  compares native to the unchanged oracle/adamic.mjs format (panic line, exit 70).
  Removing timer_dispatch's adamic_uncaught line makes that named test fail.

TestTimerMutants contains both compiling runtime mutations. Their products are
TestProduct_TimerMutantTurnPhase and TestProduct_TimerMutantCallbackUncaught.
