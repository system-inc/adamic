# `boundaries/dependencies`

| | |
|---|---|
| **Recommendation** | **Yes**, for Base projects: it is the last rule Base's ESLint config enables that cohere lacked |
| Findings in api-phi-health | **0** under Base's own two blocks, matching ESLint's 0 over the same 2,241 files |
| Differential | **5,449 of 5,449** identical (location and from/to layer pair) with every policy removed and `default: disallow` |
| Plugin | eslint-plugin-boundaries 7.2.0, `@boundaries/elements` 3.1.1 (the installed build is the authority) |
| Auto-fixable | no, and upstream has no fixer |
| Needs type information | no; reads the program's module resolution |

## What it checks

Each file belongs to at most one element (a layer), decided by `elements`. An import, re-export or
`import()` from a file in one element into a file in a different element is decided by `policies`,
the last matching policy winning, and then by `default`. Dependencies on packages, on unresolvable
specifiers, on local files in no element, and between files of the same element are never checked,
and a file in no element is never checked at all.

## How Base configures it

Two blocks of `BaseLintConfiguration.ts` turn it on, each with its own elements:

- `libraries/base/source/**/*.{ts,mts,tsx}`: `api`, `foundation` and `cli`, default disallow. The
  command line may reach api and foundation, foundation may reach api, api reaches nothing.
- `source/**` and `workers/**`: `project-module`, `project-source`, `project-worker`, default allow.
  A module may not import a worker.

## The one schema difference

Upstream reads elements from `settings["boundaries/elements"]`. cohere has no plugin settings and its
overrides carry rules alone, so the elements travel inside the rule's options as `elements`. Nothing
else moves.

## What is refused, by name, when the config is read

Everything upstream accepts that this port does not decide: legacy string selectors, `file`,
`module`, `dependency` and `captured` selectors, `allOf` / `noneOf`, micromatch patterns in a type,
handlebars or `${}` message templates, the legacy `rules` key, `checkAllOrigins`,
`checkUnknownLocals` or `checkInternals` set to true, element modes other than `folder`, and
`basePattern`. A refusal costs one config edit; a half-read policy enforces a narrower boundary than
the one written.

## Measured, and how

- Upstream has no corpus in its package, so the fixtures are verdicts from the installed build,
  driven through ESLint with Base's own config on planted imports at real api-phi-health paths.
- Element membership matches the shortest right-hand suffix, so `source/*` also claims
  `libraries/base/source/foundation`. Measured on the installed build with a custom config.
- micromatch's default keeps `*` and `**` out of dot-segments (`workers/*/**/*` does not match
  `workers/api/.wrangler/x.ts`). Measured on the installed micromatch.
- An entry's `from` / `to` fields replace the policy's, shallowly. Measured with a control.
- The strict differential above exercised nine layer pairs on the real tree, including the
  suffix-claimed ones, and agreed on every finding.

## Divergences

- `require('…')`: upstream reads it, this port does not. In a TypeScript file the program collects
  no `require`, and the typed harness cannot build a JavaScript program to prove a `require` branch.
  Base's blocks lint only `.ts`, `.mts` and `.tsx`, and those hold no bare `require` call, so this
  reaches nothing in api-phi-health today.
- A dependency that resolves outside the project root is skipped. Upstream's handling of one is not
  measured; the strict differential found no verdict it changes on api-phi-health.
