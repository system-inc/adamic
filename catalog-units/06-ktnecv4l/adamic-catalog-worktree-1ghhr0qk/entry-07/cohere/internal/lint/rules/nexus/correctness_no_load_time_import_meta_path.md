# `nexus/correctness-no-load-time-import-meta-path`

| | |
|---|---|
| **Recommendation** | **Yes, at `error`, in `cohere:next`**, off for `**/*.test.*` and `**/eslint.config.*`, which Node or the test runner runs directly and never bundles |
| Findings | **0** in ahra, www-phi-health-cohere-zero, www-ahra-ai-cohere-zero, www-phi-health and www-connected-app, in both engines, after ahra's fix wave (ahra f43214af, five sites) (measured 2026-10-05) |
| Measured precision | every finding before the fix wave was a module-scope path in a module a bundle could import, or a test file or ESLint config the override now leaves alone |
| Plugin | `nexus`, in both engines: cohere's rule and Nexus's ESLint twin hold the same rows |
| Auto-fixable | no |
| Needs type information | no |

## What it checks

`import.meta.dirname` or `import.meta.filename`, read while the module loads: in a top-level statement, a
module-scope initializer, a top-level call's arguments, a static field or block, a computed member name, a
decorator, a namespace body, or a function invoked on the spot. A read inside a function's body or
parameter, or an instance field's initializer, runs later and is quiet.

One shape is exempt: the main-module guard, `import.meta.filename` compared for equality with
`process.argv[N]` or with a call taking it. In a bundle it compares against undefined and is false, which is
right there.

## Why it exists

A Next server bundle defines neither name. Turbopack gives the module a synthesized `import.meta` with only
`url` and `turbopackHot`. ahra's RunSuites.ts joined `import.meta.dirname` at module scope, Next's server
bundle reached it through the commit command line interface, and the join threw at startup and took the dev
server down (ahra 36b7fe07, #ay7a8vx). A bare `const directory = import.meta.dirname` is reported too: it
doesn't throw at load, it throws later, wherever it's joined.

## Where it runs

`cohere:next` only, by @system_cohere_lint_sets's ruling: in plain Node ESM a module-scope
`import.meta.dirname` is the correct replacement for `__dirname`, so the hazard exists only where a bundler
moves the module.

## Not covered

A function declared at module scope and called there runs its body at load. The rule doesn't follow calls.

## Options

None.
