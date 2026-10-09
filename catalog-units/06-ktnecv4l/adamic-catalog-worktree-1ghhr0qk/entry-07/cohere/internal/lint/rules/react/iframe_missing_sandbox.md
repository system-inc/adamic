# `react/iframe-missing-sandbox`

| | |
|---|---|
| **Recommendation** | **Yes** |
| Violations in ahra | none |
| Plugin | `eslint-plugin-react` |
| Auto-fixable | no |
| Needs type information | yes |

## What it checks

Enforce sandbox attribute on iframe elements

## Where cohere differs from upstream

Upstream reports `allow-scripts allow-same-origin` on every iframe. The escape it warns about needs a framed document in the page's own origin, so cohere is silent on the pair only when the frame is provably cross-origin: an `http:` or `https:` URL or a `//host` whose host is complete, a `data:` URL, or another scheme no page is served from. A cross-origin embed such as a YouTube player stays silent. A same-origin frame reports (a relative or absent `src`, an `about:`, `blob:` or `javascript:` URL, or any `srcDoc`), and so does a frame whose origin it cannot prove (a plain `string`, a bare `src`, a spread after the last `src`), since a dynamic `src` that resolves same-origin is the case the warning exists for (#3rxx2y9). A template is read from its head, a `const` from its initializer, and any other value from its type's literal constituents, all of which must be cross-origin.

## Why this recommendation

Stylistic, but the tree already satisfies it, so enabling costs nothing and prevents drift.

## Violations

None. The tree already satisfies this rule, so enabling it is a guardrail against future drift rather than a cleanup.

