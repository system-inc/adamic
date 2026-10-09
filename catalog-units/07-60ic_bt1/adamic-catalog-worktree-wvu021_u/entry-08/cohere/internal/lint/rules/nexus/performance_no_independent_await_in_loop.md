# `nexus/performance-no-independent-await-in-loop`

| | |
|---|---|
| **Recommendation** | **Yes**, in place of `no-await-in-loop` |
| Findings in ahra | **56** (measured 2026-10-01), against 379 for upstream `no-await-in-loop` in the same run |
| Measured precision | 53 of 56 true (94.6%): 38 plain, 15 sequential on purpose and worth a suppression |
| Plugin | cohere-native, `nexus` |
| Auto-fixable | no |
| Needs type information | yes, for binding resolution and the callee scan |

## What it checks

A loop over a collection whose iterations wait on each other for no reason the code can show. It
fires only when every condition holds; the rule's doc comment carries the precise version.

1. **The loop walks a collection.** `for...of` (not `for await`), or
   `for (let i = 0; i < xs.length; i++)` stepping by one (`i++`, `++i`, `i += 1`). `while`, `do`,
   `for...in`, `for await` and any other `for` header never fire: those are retry, polling and
   pagination.
2. **The loop itself waits.** An await, an `await using`, or a nested `for await` in the body, not in
   a nested loop's re-entered parts and not in a nested function. An await whose operand is the item
   itself (`await promise` over pre-started promises) does not count.
3. **No loop-carried state.** Bindings are resolved through the checker; a binding is outer when no
   declaration of it lies inside the loop. Any write to an outer binding is silence, as is a `var` in
   the body, a write to the indexed form's counter, `+=`/`++` on an outer property, or a consuming
   call (`pop`, `shift`, `splice`, `next`, `read`, `sort`, `reverse`) on an outer receiver.
   Collecting results is allowed: `push`, `unshift`, `set`, `add`, `append`, `delete`, any
   `add*`/`append*`/`insert*` method, `Object.assign(outer, ...)`, `outer[key] = ...` and
   `outer.member = ...` are sinks, provided the body (and the loop's iterable) never reads the filled
   path back. Paths compare as property chains, so `this.cache` and `this.fetch` do not overlap.
4. **No exit on an awaited value.** A `break`, `return`, uncaught `throw` or labeled `continue` out
   of the loop silences the rule when what decides it reads an awaited value: an enclosing `if`,
   `switch`, `case` or inner-loop test; a `catch` whose `try` awaits; a `try` with a `catch` that
   awaits before the exit; an earlier sibling statement that jumps on such a value; or the exit's own
   expression. Taint is a fixpoint over the body's declarations and assignments, and a catch
   parameter is tainted when its `try` awaits. A plain `continue` and a catch that records and moves
   on are fine.
5. **No ordered side effect**, by name, anywhere the iteration runs including callbacks:
   `console.*`, `process.stdout.*`/`process.stderr.*`, `setTimeout`/`setInterval`/`setImmediate`,
   names starting `sleep`, `delay`, `wait`, `pause`, `throttle`, `backoff`, `rateLimit`, `print`;
   names or receivers containing `progress` or `spinner`; `write`, `writeSync`, `appendFile`,
   `appendFileSync`; `log`/`info`/`warn`/`error`/`debug`/`trace` on a receiver containing `log`; a
   bare `log(...)`; a `yield`. Inside a `catch` in the body, and one level into every called
   function (resolved through the checker, imports followed by alias), only progress output and
   explicit pacing count: `console.error`/`warn`, `process.stderr`, `setTimeout` and anything inside
   the callee's own `catch` do not.

The finding spans the loop header, so `// cohere-disable-next-line
nexus/performance-no-independent-await-in-loop -- <why it is sequential>` goes above the loop.

## Why this exists

Upstream `no-await-in-loop` reported 379 sites on ahra and roughly three quarters were deliberately
sequential, which teaches a reader to suppress it unread. This rule asks the same question with the
program in hand and reports only where the iterations can be shown independent.

## How the measurement was taken

Binary built from this tree into a scratchpad, run from `~/Projects/ahra` as
`cohere --no-fix --lint -lint-config <scratch copy of CohereSettings.json>` with this rule enabled.
The scratch config's non-`**` globs were re-anchored to ahra, because a config's globs resolve
against its own directory. 3,605 files linted, 0 crashed, summary line present.

Three passes:

| pass | findings | change |
|---|---|---|
| first | 65 | 13 false positives (20%) on hand classification |
| item-itself, prefix sinks, callee scan counting every ordered call | 44 | lost the 8 intended plus 13 true positives: callees print from `catch` blocks |
| callee scan limited to progress output and pacing outside `catch` | 55 | exactly the 8 intended, plus 2 test loops whose callee prints progress |
| body `catch` narrowed the same way, after a recall sample | 56 | adds GmailApi.ts:583, silenced before only by a `console.warn` in its catch |

## The 56, classified

True positive, `Promise.all` (or a map over the collection) is correct (38):
TasksNewTaskDialog.tsx:101, SensationsDashboard.tsx:143, Decryption.ts:52, :65, Encryption.ts:49,
:62, ArraySchema.ts:93, :110, BaseSchema.ts:87, ObjectSchema.ts:70, :87,
GraphQlOperationsMetadataPlugin.ts:689, :700, NetworkService.ts:538, BackupStaleness.ts:328, :366,
:377, DiscordApi.ts:435, EnvironmentSystemCommands.ts:102, DeelSplitter.ts:318,
FinanceConnections.ts:57, MercuryAdapter.ts:187, StripePayoutSplitter.ts:436,
FinanceCommandLineInterface.ts:107, EmailApi.ts:504, :553, GmailApi.ts:583, :702, OpenAiApi.ts:206,
AhraOsCommandLineInterface.ts:119, AhraOsWisdomReach.ts:98, PhiCommandLineInterface.ts:99,
ReachAnalyticsApi.ts:1136 (the `wrap` closure pushes from outside the loop, so the rewrite should
return values rather than push), ReachDirectMessageApi.ts:328, :576, ReachMentionsApi.ts:399,
ReachSearchApi.ts:349, TasksCommandLineInterface.ts:113.

True positive, plausibly sequential on purpose, a suppression with a reason is the expected answer
(15): ImageUploader.tsx:116 (upload bandwidth), AsanaApi.ts:163, :1295 (Asana rate limit),
BackupFileRestore.ts:88 (download bandwidth), FinanceSync.ts:522 (`retryOnRateLimit`), :527,
StripeApi.ts:420 (Stripe rate limit over every customer), AhraOsReshape.ts:302 (spawns positions),
AhraOsTriggers.ts:597, :1380 (one `gh`/CLI subprocess each), AhraOsWisdom.ts:4391 (up to 400
network reads), AhraOsWisdomComments.ts:230 (subprocess each), RedditSubmitApi.ts:307 (Reddit rate
limit), FrameTvApi.ts:616 (one TV), SeeFaces.ts:363 (CPU-bound inference).

False positive (3):

- GrokMediaApi.ts:141 and OpenAiMediaApi.ts:66. `persistArtifact` probes the output directory for
  the next free `<n>.<ext>` and then writes it, so concurrent calls race to the same filename. The
  shared state is the filesystem, two calls down, and no name says so.
- SystemBasePostsPublishBatch.ts:296. "Create each post in order": creation order is observable on
  the server. Nothing local carries between iterations.

All three are state the syntax cannot see. They are the honest floor of this approach.

## Silent on purpose, checked against the tree

PromiseGroup.test.ts:304, :328, :359, :412 await promises started before the loop. FacetsMonthlies.ts:302
and AhraOsCommsCommandLineInterface.ts:907 call functions that print per item. AsanaApi.ts:366
calls `printTreeNode`. AhraOsReportPdf.ts:455 builds one PDF page by page and hands the document to
the await. PensieveDailies.test.ts:710 and :736 call a function that prints progress.

## Recall, sampled

Upstream reported 379 awaits in the same run. A textual pass paired each with its nearest `for...of`
header: 77 sit in loops this rule reports, 116 loops were declined, and the rest are in `while`,
`do`, `for await` or other `for` shapes. Sixteen declined loops were drawn at random and read. Two
were the pairing's own error (the loop held no await). Thirteen were declines the conditions call for:
console output (5), an accumulator or running max (4), find-first `break`, a shared PDF, UI
progress state, pre-started promises. One was a miss, GmailApi.ts:583, silenced by a `console.warn`
in its catch, which is what led to the body-`catch` narrowing above. The accumulator declines are
parallelizable with a restructure; the rule leaves them alone by design.

## Not auto-fixable

The rewrite is `Promise.all`, a bounded pool, or a suppression, and which one depends on the
collection's size and what the far side tolerates.
