# `nexus/concurrency-no-check-then-write`

| | |
|---|---|
| **Recommendation** | **Yes**, at `error` once ahra's 8 findings are fixed (`#9erqsas`, one library helper for all eight). Registered `off` in ahra until then |
| Findings | **ahra 8, www-phi-health 0, www-connected-app 0, api-phi-health 0** (measured 2026-10-02) |
| Measured precision | 8 of 8 true: every finding is a free-name search whose name a later non-exclusive write claims |
| Research count | 8, probe `P1c_freeNameProbeLoopThenNonExclusiveWrite`, the same 8 sites |
| Plugin | cohere-native, `nexus` |
| Auto-fixable | no |
| Needs type information | yes, every callee is resolved through the checker; without a checker the rule registers nothing |

## What it checks

A loop that looks for a free file name by checking whether each candidate exists, then a write after
the loop that creates the chosen name without refusing a file already there. Two writers can both see
the name free, and the second write replaces the first's file. The rule's doc comment carries the
precise version: the two loop shapes that exit exactly on a free name, Node's `fs` resolved through the
checker, the counter, the writes and how their flags are read, the structural path proof, and the
stability check on every variable the path reads.

## Where it came from

`modules/intelligence/IntelligenceMediaGenerationApi.ts`, `persistArtifact`: an `access()` loop chose
the next `<index>.<ext>` and a plain `writeFile` wrote it, so two generations saving at once lost an
image. Fixed in 0f33c3f0 with `writeFile(..., { flag: 'wx' })` retried on `EEXIST` (now `:121`). The
new-rules sweep (`#tevhg3f`) found eight live copies, filed as `#9erqsas`. Built as `#2xqtnzj`.

## Nine before, eight found

The brief counted nine sites: the eight live ones and the incident's own before-file. The rule fires
on the eight and, by decision, not on the ninth. The before-file read its path as
``NodePath.join(input.outputDirectory, `${fileIndex}.${input.extension}`)``, and a property can be
changed by anything else holding `input` during the `await access(...)` between the check and the
write, so the two paths cannot be proven equal. A fixture holds the real before-file and asserts it is
silent; the same function with the two fields read into locals fires. After the fix the count is 0 on
all nine, the fixed `IntelligenceMediaGenerationApi.ts` included.

## The findings, read one by one

**ahra (8)**. Cohere reports on the check call. Two line numbers differ from the research report's
(`MagnificApi.ts:406` and `OpenAiApi.ts:309`): commits 4e385b61 and 1382cbe9 moved those files after
the probe ran, and the reported line is the same `access` call.

| Site | Write | Reading |
|---|---|---|
| `modules/kling/KlingApi.ts:253` | `writeFile` of the mp4, after a whole network download | **Bug.** In-process and cross-process race, with the widest window of the eight |
| `modules/kling/KlingApi.ts:383` | `writeFile` of a png, inside `urls.map(async ...)` | **Bug.** Siblings of one call race each other: two downloads can settle on the same index |
| `modules/magnific/MagnificApi.ts:416` | `writeFile` after a download | **Bug.** Same loop, same race |
| `modules/openai/OpenAiApi.ts:310` | `writeFile` of decoded or downloaded bytes | **Bug.** Same loop, same race |
| `modules/openai/OpenAiCommandLineInterface.ts:144` | `copyFile` from the generator's temporary file | **Bug.** Same loop; the copy replaces an existing file |
| `modules/art/ArtApi.ts:231` | `writeFileSync` of the video | **Bug**, narrower: synchronous, so it races only across two CLI runs |
| `modules/art/ArtLibrary.ts:65` | `copyFileSync` into the library | **Bug**, narrower, as above |
| `modules/phi/social/PhiSocialLibrary.ts:210` | sharp's `toFile`, through two `const`s | **Bug**, narrower, as above. The write is matched because `toFile` resolves to the declaration file this file's `sharp` import resolved to |

**www-phi-health, www-connected-app, api-phi-health (0)**. None of the three has a free-name loop. The
only other existence-check loop on the four trees is
`api-phi-health/libraries/base/command-line/graphql/GraphQlSchemaDestination.ts:223`, which removes
empty folders and writes nothing.

## Declined, by decision

All in the doc comment, each a missed finding rather than a false one: a check with no loop (the
research probe `P1b` counted 27 on ahra, nearly all create-if-missing and cache files), a write inside
the loop or in a callback, a check that is not the loop's exit, a path read through a property or a
call other than `path.join`, sharp's `toFile` in a file that does not import `sharp`, and `open`, `cp`
and every other third-party writer.

## Verification

- Fixtures both ways from the real sites: `persistArtifact` (fields in locals), `KlingApi` video and
  `ArtApi`, `ArtLibrary` and `PhiSocialLibrary` each before (one finding on the check, the message
  naming the write) and after (silent); `KlingApi` images and `OpenAiCommandLineInterface` before. Nine
  more firing shapes (every loop kind, named imports, `fs.promises`, `rename`, `createWriteStream`, a
  branch on either side), `@types/node` 20's declaration shape, twenty-two silent shapes, the two
  lookalikes and the real before-file of the incident (silent, the property-read miss).
- Mutation check, each mutant through `go test -overlay` and each killed: Node module resolution,
  the counter, the exclusive flag, the path proof, the stability check, the `catch` being only
  `break`, no `finally`, the promise check awaited, one argument to the check, only counters after it,
  stopping at an enclosing loop, no write from a nested function, sharp resolved by the resolver,
  reading through only later or loop-body `const`s, only `path.join` counted pure, the loop-body
  `const` window, no counter writes from a nested function, the condition exiting on a free name, and
  the write running after the loop.
