# `require-await`

| | |
|---|---|
| **Recommendation** | **Yes, enabled** |
| Violations in ahra | **3** (cohere, 2026-10-01); ESLint reports 89 |
| Plugin | `eslint core` |
| Auto-fixable | no (a suggestion, never a fix) |
| Needs type information | no upstream; cohere asks the checker for the contract exemption |

## What it checks

Disallow async functions which have no `await` expression

## Why this recommendation

Catches a defect rather than a preference. The classic catch is `items.map(async (item) => transform(item))`,
which hands back an array of promises where values were wanted.

The 637 this document used to cite were the `generateMetadata` world: Next.js route files writing
`export async function generateMetadata(): Promise<Metadata>` without awaiting. That world is gone. On
2026-10-01, 43 route files write `generateMetadata` without `async`, and none of the 22 that keep it
reports here.

## Violations

Re-measured 2026-10-01 against the ahra tree, `cohere --no-fix --lint` beside ESLint:

| | count | |
|---|---|---|
| both report | 1 | `modules/replicate/ReplicateApi.ts:193`, `listVersions` awaits nothing |
| ESLint only, cohere right | 86 | the position's declared type demands a promise: 54 pensieve test doubles, 12 `AdsRegistry` collectors, 10 `AhraOsReactions`/`AhraOsTriggers` conditions, 5 finance adapters, and 5 more implementing or overriding a promise-returning member |
| ESLint only, ESLint right | 0 | was 2 before the generic self-inference fix; see below |

cohere's 3 are the `ReplicateApi` site and `useMetricsExportMenu.tsx:74` and `:92`.

**The contract exemption is cohere's, not upstream's.** A function written into a position whose
declared type is `=> Promise<T>` has to be `async` (or return a promise some other way that changes
throw semantics), so reporting it tells the author to break the assignment. ESLint judges one function
at a time and cannot see the position, which is the whole of the 86.

**The exemption once graded its own homework.** Inside a generic call the contextual type is read after
inference, and inference reads this very function: `[1, 2].map(async (n) => n)` resolves `U` to
`Promise<number>` and the callback appeared to be demanded. That hid `useMetricsExportMenu.tsx:74` and
`:92` (`onSelected: async function () {...}` inside `React.useMemo(() => [...])`), which ESLint reports
and is right about. The rule now replays a generic call's DECLARED parameter type instead; see
`requireAwaitDeclaredGenericDemand` in `require_await.go` and
`TestRequireAwaitDoesNotLetAGenericInferItsOwnContract`.

### What fixing looks like

Not auto-fixable. Each site needs a human decision, which is what makes the count above the real cost.

