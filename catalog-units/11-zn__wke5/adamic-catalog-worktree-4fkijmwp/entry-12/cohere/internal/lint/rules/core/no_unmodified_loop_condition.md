# `no-unmodified-loop-condition`

| | |
|---|---|
| **Recommendation** | **Strong Yes** |
| Violations in ahra | **3** |
| Plugin | `eslint core` |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

Disallow unmodified loop conditions

## Where cohere is deliberately quieter than ESLint

Upstream credits a write only when it sits in the loop, or in a function declaration whose name the
loop references. A loop that awaits hands the thread to whatever else is queued, so a flag set by a
signal handler or an abort callback ends it between iterations, and upstream reports exactly that
shape. cohere counts such a write as modifying (`unmodifiedLoopWriterRunsWhileSuspended`) when both
hold:

- the loop suspends: an `await` or `yield` in its test, body or update that belongs to the loop's own
  function, not to a function nested in it (a `for` initializer runs once and does not count);
- the write is in a different function activation from the loop's own, one that can exist before the
  loop could finish. A closure the loop's own function creates must be a function or arrow written
  before the loop's end, or a hoisted declaration whose name is referenced before the loop's end. A
  hoisted declaration further out must be referenced somewhere. A write in an enclosing function's own
  body counts as it stands.

The real sites that went silent, all three of the rule's findings in ahra and all three false:
`modules/tasks/TasksWatchCommandLineInterface.ts:305` (`stopping`, set by a hoisted SIGINT handler),
`modules/os/sensation/AhraOsMonitors.ts:548` (`abortRequested`, set by an `abort` closure beside the
async IIFE that runs the loop), and `modules/os/boot-screens/RainbowMatrix.ts:689` (`stopped`, set by
`teardown`, reached through the registered `onSignal`). 3 findings before, 0 after. None of ESLint's 37
imported cases moved.

Fixtures: `TestNoUnmodifiedLoopConditionSeesAWriterThatRunsWhileTheLoopIsSuspended` in
`no_unmodified_loop_condition_test.go`. Nine silent rows (the three sites, a generator `yield`, an
enclosing function's own write, a hoisted handler declared after the loop and registered before it,
an `await` in the update clause, a closure registered by an enclosing function after starting the
loop, a hoisted handler further out registered after the loop starts) and ten reporting rows that
keep the detector alive (an awaiting loop with no writer, a handler beside a loop that never
suspends, an `await` in a nested function, an `await` only in a `for` initializer, a write in the
loop's own activation, a closure created after the loop, hoisted writers never registered or
registered too late, and a closure writing a shadowing binding). The three ahra sites were then
rewritten so each flag is an AbortController (ahra ede7b594), whose `signal.aborted` is a member read
both engines treat as dynamic, so neither reports them (#cn8sthd).

## Why this recommendation

Catches a defect rather than a preference, and the cleanup is small enough to do in one pass.

## Violations

3 in the tree. Showing the first few.

**`modules/os/boot-screens/RainbowMatrix.ts:685`**

```
while(!stopped) {
```

> 'stopped' is not modified in this loop

**`modules/os/sensation/AhraOsMonitors.ts:547`**

```
while(!abortRequested && currentRow !== null && currentRow.enabled === 1) {
```

> 'abortRequested' is not modified in this loop

**`modules/tasks/TasksWatchCommandLineInterface.ts:305`**

```
while(!stopping) {
```

> 'stopping' is not modified in this loop

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

