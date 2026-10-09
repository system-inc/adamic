# `nexus/correctness-require-response-status-check`

| | |
|---|---|
| **Recommendation** | **Yes, but registered `off` until the sites are fixed.** 162 findings across the four trees, 0 false. Turning it on at `error` today would put 144 findings on ahra |
| Findings | **ahra 144** (125 body read, 19 discarded), **www-phi-health 2, www-connected-app 3, api-phi-health 13** (measured 2026-10-02) |
| Measured precision | 162 of 162 true: in every finding the Response's status is never read on the reported path. 133 are real bugs, 29 are true but harmless (an error envelope in the body is tested instead) |
| Plugin | cohere-native, `nexus` |
| Auto-fixable | no |
| Needs type information | yes, to resolve `fetch` and `networkService.request` to their declarations; without a checker the rule registers nothing |

## What it checks

A fetch Response whose body is read (`.json()`, `.text()`, `.arrayBuffer()`, `.blob()`, `.formData()`,
`.bytes()`, `.body`) on a path that never reads `.ok`, `.status` or `.statusText` and goes on to a normal
exit, and a fetch Response awaited and dropped. Tracked producers are the platform's global `fetch` and
Structure's `networkService.request` / `baseApiRequest`, both read and confirmed not to throw on an HTTP
error. The doc comment carries the precise version: what counts as a check, an escape and a body read,
and why checking after the read is accepted.

## Where it came from

`app/(os-layout)/_hooks/useTaskInboxRequest.ts:19` in ahra (a 401 or 500 parses into `data.tasks ===
undefined`, cached as the task list) and `app/(os-layout)/finance/_components/FinanceReviewView.tsx:44`
(the categorize POST drops its Response, so a 400 or 500 reports success). Both verified on the tree
before being used as fixtures. Found by the cross-language pass of the new-rules sweep (`#tevhg3f`), built
as `#3twjf7k`.

## Reconciling with the research count

Research counted **139** on ahra (120 body read from `fetch` or `networkService.request`, 19 discarded).
cohere counts **144** (125 + 19). The 19 discards match exactly. The body reads differ by 5, all
explained, none a false positive:

- **+2, `baseApiRequest`**: the research producer set named `networkService.request` only.
  `libraries/structure/source/modules/account/profile/hooks/useAccountProfileImageUploadRequest.ts:32` and
  `.../asset/hooks/useAssetUploadRequest.ts:72` go through `baseApiRequest`, which delegates to the same
  executor.
- **+3, path-sensitive**: the research probe silenced a binding when `.status` appeared anywhere in the
  function. The rule asks per path, and these read it only on the error branch:
  `app/(os-layout)/os/wisdom/_components/WisdomActionCluster.tsx:117` (status only inside the
  `'Unreadable'` throw; a 500 shows "A draft is already waiting"), `modules/doordash/DoorDashClient.ts:107`
  (status only inside the not-GraphQl throw), `modules/system/base/BaseGraphQlClient.ts:160` (the Response
  escapes to a cookie helper only when `captureSetCookie` is true).
- Research's other 13 body reads came from wrappers (`R2Api.signedS3Request` x4, `KlingApi.this.get` /
  `this.post` x6, a `Response` parameter in `ReachHttpClient.ts` x2, one `GeminiApi.ts` site since edited),
  out of scope by design: `signedS3Request` already throws on `!ok`.
- The research list's line numbers have drifted since (`DiscordApi.ts:671` is now `:675`, and so on);
  the per-file counts match apart from the five above.

## The findings, read one by one

All 162 were read in context (`context.txt` in the build scratch directory, the declaration through ten
lines past the read).

**ahra, real (101 body reads + 19 discards)**

| Shape | Sites | Reading |
|---|---|---|
| `useReadRequest` hook: `return (await response.json()) as T` or `data.tasks` (72) | `app/(os-layout)/_hooks/` x12 (`useTaskInboxRequest.ts:19` and eleven siblings), `finance/_hooks/` x35, `finance/_components` dashboard and net worth x2, `contacts` x2, `art` x2, `SensationsDashboard.tsx:101`, `os/wisdom` reads and boards x6, `phi/social` x3, `secrets/_hooks/useSecrets.ts:53`, `see/_hooks` x4 plus `SeeFolderPicker.tsx:23` and `SeeIngestBanner.tsx:35`, `things` x2 | **Bug**: an error body is cached as the read's data, and the surface crashes or shows empty far from the cause |
| Write request or handler parsing the result: `const result = await response.json(); setX(result.y)` (17) | `_components/detail/TaskDetailActions.tsx:95`, `TaskDetailCommandRulingCard.tsx:81`, `_hooks/useTaskRunRequest.ts:28`, `os/[username]/activity/OsPositionActivityLive.tsx:59`, `os/sensation/_components/SensationNerveCard.tsx:250`, `os/wisdom` x5 (`WisdomActionCluster.tsx:117`, `WisdomBundleStepRow.tsx:162`, `WisdomCommandPreview.tsx:74`, `WisdomHome.tsx:245`, `:394`), `see/_components` x7 (`SeeCandidateSubCluster.tsx:67`, `:92`, `SeeClusterCard.tsx:124`, `:151`, `:174`, `:198`, `SeeDropZone.tsx:41`) | **Bug**: a rejected write shows `undefined` counts or a wrong status line |
| Structure upload hooks via `baseApiRequest` (2) | `useAccountProfileImageUploadRequest.ts:32`, `useAssetUploadRequest.ts:72` | **Bug**: a 401 crashes on `.map` of an error object, or returns it as the upload result |
| Device and module clients (10) | `Control4Api.ts:29`, `:43`, `:52`, `:178` (`parseInt` of an error page reads as level 0), `ElgatoRingLightApi.ts:16`, `HueApi.ts:16`, `MidjourneyApi.ts:118`, `:594` (an error page written into the image or video file), `PhiAnalyticsApi.ts:199`, `FrameTvApi.ts:308` | **Bug** |
| Discarded `await networkService.request(...)` / `await fetch(...)` (19) | `FinanceFloorView.tsx:40`, `FinanceReviewView.tsx:44`, `:52`, `FinanceTransactionDetail.tsx:63`, `:71`, `:82`, `:92`, `SensationsDashboard.tsx:129`, `:166`, `WisdomHome.tsx:354`, `SeeEnroll.tsx:82`, `SeeFaceInspector.tsx:42`, `SeePeoplePanel.tsx:28`, `Control4Api.ts:183`, `:188`, `:193`, `ElgatoRingLightApi.ts:28`, `HueApi.ts:51`, `GeminiApi.ts:283` | **Bug**: a failed write reports success (the Gemini delete is housekeeping, and still silent on failure) |

**ahra, true but harmless (24 body reads)**: the status is never read, but the body's own error envelope
is tested and the API sets it on every failure. `email/_hooks/useEmailInboxRequest.ts:30`,
`useEmailThreadRequest.ts:23`, `os/[username]/images/useSetCurrentPortraitRequest.ts:26`,
`posts/useWallPostRequest.ts:30` (`data.error`, set by our own routes, though a middleware 401 without the
field still slips through); `app/api/spotify/callback/route.ts:76` (OAuth `error`); `AsanaApi.ts:53`
(`errors`); `R2Api.ts:211` (Cloudflare `success`); `DiscordApi.ts:675`, `BaseGraphQlClient.ts:160`,
`SystemBaseGraphQl.ts:102`, `DoorDashClient.ts:107` (GraphQl `errors`); `StripeApi.ts:118` (`error`), `:433`
(`result`); `DocsApi.ts:342` (`success`); `KlingApi.ts:432`, `:444`, `:455` (`result !== 1`);
`LmStudioCommandLineInterface.ts:72` (`error`); `PresenceApi.ts:174` (the label prefix);
`TikTokAdsApi.ts:69`, `TikTokAdsAudiencesApi.ts:147`, `TikTokAdsCreativesApi.ts:136`, `:181` (`code`);
`LinkedInAuthApi.ts:422` (a debug command printing whatever came back).

**www-phi-health (2) and www-connected-app (3)**: the two Structure upload hooks in each, **bugs** as above;
connected's `libraries/library/api/BaseApi.ts:109`, harmless (GraphQl `errors` and `!data` both throw).

**api-phi-health (13)**: `FlexPaymentProcessor.ts:184`, `:280` **bugs** (an error body crashes on
`checkout_session.customer` instead of reporting Flex's error); `FlexPrivilegedGraphQlService.ts:44`, `:73`,
`:110`, `:151`, `:192`, `:234`, `:277` **bugs** (each admin write logs the body at debug and returns
success whatever the status); `VerifyWorkersCloudflareCommand.ts:69` (Cloudflare `success`),
`GoogleAdsAudienceService.ts:309`, `GoogleAdsEnhancedConversionsService.ts:465` (a missing `access_token`
throws), `workers/apple/library/AppleSignedDataVerifier.ts:415` (the OCSP parser rejects a non-OCSP body,
so verification fails closed): harmless.

**Recall spot check**: every ahra file that calls a tracked producer, never spells `.ok` / `.status`, and
has no finding was read. There is one, `useSupportTicketCreateRequest.ts`, and it returns the Response to
its caller: correctly an escape.

## Verification

- Fixtures both ways from the real sites: `useTaskInboxRequest` before (one finding on `response.json()`)
  and after (`if(!response.ok) throw`, silent); `FinanceReviewView`'s two actions before (two discards) and
  after (kept and checked, silent). 13 firing shapes, 21 silent shapes, the `@types/node`
  `declare global` declaration, and an ambient namespaced `fetch` that must not count.
- Mutation check, each mutant a copy through `go test -overlay`, and each killed but one:
  global fetch accepting any symbol (local `fetch`); the global-scope container test (namespaced fetch);
  NetworkService accepting any method, and accepting any file (another `NetworkService` class);
  status members read as neutral (checked before the parse, and nine more); a non-access use read as
  neutral (stored in an object); an unknown member read as neutral (a clone handed on); an uncalled
  `.json` read as a body read (the parse method handed on uncalled); closures not dropping the binding;
  no zero-reference discard; no discard statement (FinanceReviewView); no inline read; no `.then`
  binding; the first walk passing checks (checked before the parse); no second walk, and the second walk
  passing checks (parsed, then checked); thrown exits counted as exits (read to be thrown); the next
  loop turn not ending the Response (a loop that never leaves). **Survived**: removing the unvisited
  reference guard. A probe reporting whenever the guard fires counted 0 on all four trees, and no shape
  was found that reaches it; the doc comment says why it stays.
- `gofmt -l` and `go vet` clean through the overlay. The whole suite through the overlay is green apart
  from `TestEveryRegisteredRuleIsReachableFromTheLiveConfig` (this rule is not in the live config yet),
  `TestRulesFlagListsEveryRegisteredRule` (it builds the binary without the overlay, 460 against 461),
  and `command/cohere` tests failing to build on another author's in-progress edits to `main.go` and
  `format_test.go`, untouched here.
