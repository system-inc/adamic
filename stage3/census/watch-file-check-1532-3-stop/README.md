# Stopped: exact-owner pin composition

No adaptation commit was created or pushed. Parent: 13549c787e027a72afc4a9a6b0e10f07857abcfd. Source: 63f46490764862fef5af3ce6793557d4000994bc.

93-builtin-imports owns all of getNodeSystem. The parent constructs patchedWriteFile there; the source adds a worker interval widening and exact ownerBeforeText/ownerAfterText assertions. The combined adapter retains both edits. Its --pin branch regenerates the hashes but spreads the old manifest without updating the exact owner text. Pin generation succeeds, first apply succeeds, replay fails with whole getNodeSystem text drift. The assertion diff identifies patchedWriteFile and its property assignment as the difference. No hash was hand-edited.

This is a bounded tooling reproduction, not a claim that the requested full composed apply or census ran. Per the requested stop rule, work stopped before changing the pin generator's contract. All requested full verification remains unrun. The shared checkout's working files were untouched; fetched objects and worktree registrations were added.

## Reproduction

Use a fresh detached scratch worktree at the parent, with npm ci in stage3/api. Apply combined-adapter.patch there. Obtain proofs/targets.json from the source commit. Create a scratch fixture with src/compiler/tsconfig.json containing {"files":["sys.ts"]}. Write sys.ts as the exact concatenation of `import { errorCode } from "./hostErrors.js";` plus LF, the source manifest's ownerBeforeText, then LF. Preserve CRLF inside the owner text.

With NODE_PATH pointing to that worktree's stage3/api/node_modules, run its 93-builtin-imports/adapt.cjs --pin FIXTURE, then adapt.cjs FIXTURE twice. Observed exit codes: 0, 0, 1. These are 93-pin.log, 93-apply.log and 93-replay.log. generated-pins.json records the tool's output manifest. No compiler build is needed for this fixture.

Setup also completed git submodule update --init --recursive --depth 1 and cloned TypeScript v6.0.3 at the requested commit. No Go process ran. This reproduction used only scratch writes; no adaptation commit or target branch was created.
