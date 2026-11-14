# Environment presence 1965: PASS

Implementation: [typescript/env-presence-1965-1](https://github.com/system-inc/adamic/commit/61d3479cd8a0e2d708fa53abde10a74442a914fb), `61d3479cd8a0e2d708fa53abde10a74442a914fb`. Exactly one commit on `8e3a784aa5a601f57342c07b68f145fe56252c21`, six files: one new adapter directory and one ledger line. Task: #baobab item 11. No Loom submission. All pushes used new branch names and ordinary non-force pushes.

| Proof | Result |
| --- | --- |
| Count before construction | Complete unfiltered oracle, eight workers: 106,366 passing, one sanctioned API failure, zero pending. Nine visits: nine truthy with getEnvironmentVariable as a function; zero truthy without; zero falsy. Actual-condition controls fire for all categories. |
| Node witness | Executes the actual adapted polling helper and pinned block. Present System: identical nine polling reads, NODE_ENV read and Normal assertion level, receiver preserved. Undefined: both skip. Labeled missing-member hypothetical: original skips, new enters, helper returns early and NODE_ENV call throws TypeError. |
| Full sys.ts census | CommonJSBundle, LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1 and exact LATENT_ONLY_FILE: 185 -> 182 findings; zero new, zero changed. |
| All gone findings | refusal_scan Refused at 1965:12; lowering NotYet at 1965:12; lowering rollback Boundary at 1965:1. delta-full.json retains all complete original records, including phase, diagnostic and byte spans. |
| Restoring mutant | Original condition restored byte for byte; full census returns all 185 findings and exactly matches baseline, including both target findings. |
| Stock tsc 6.0.3 | All 79 roots: zero diagnostics before and after, zero new. |
| Adamic types | All 79 sorted compiler roots, exit zero and zero diagnostics. |
| JavaScript | 604 source files / 1,208 emits, comments retained and removed. Only the one condition text differs in sys.ts; every other emitted file is byte-identical. |
| Upstream declarations | Independent npm ci then node_modules/.bin/hereby dts on main and candidate pass; typescript.d.ts and tsserverlibrary.d.ts byte-identical. declarations.diff is empty; declarations.json records hashes. |
| Contract | LF and CRLF adaptation, exact mutant, zero-write replay; drift, half-adapted text, duplicates, mixed before/after and mixed line endings refuse with zero writes. |
| Integrated apply | PASS; all 747 src files byte-identical to proved candidate. Adapter replay writes zero files; bytes and mtime unchanged. |
| Final full lane | PASS, unfiltered, eight workers, 106,366 passing, only sanctioned Public APIs failure, zero pending; zero timeouts, exceptions or extra failures. API guard: 273 composed declarations equal all 273 sanctions. |
| Commit identity | Every committed blob matches the pre-lane six-file manifest. Post-commit apply input key equals the tested key d54755c39231380c03d2870065bc2831ef9caae30ed99e650a9da744d1183f7b. |

The block passes sys to setCustomPollingValues, which reads the LOW/MEDIUM/HIGH settings for TSC_WATCH_POLLINGINTERVAL, TSC_WATCH_POLLINGCHUNKSIZE and TSC_WATCH_UNCHANGEDPOLLTHRESHOLDS. It then calls sys.getEnvironmentVariable("NODE_ENV") with its receiver intact, selecting AssertionLevel.Normal for case-insensitive development and None otherwise. These are its only other direct sys reads: the argument and the method-call receiver. The following separate block reads sys.debugMode.

The dropped member-presence test is JavaScript-changing and value-identical on the measured domain. The ledger names that test and #baobab item 11. No existing whole-owner pin includes this top-level if statement; ownership.json records the required search. owners.json pins the whole statement before/after with exact text and SHA-256.

Lane ran before the implementation commit, so execution.commit names the base. implementation-files.json and implementation-final.json prove the committed bytes and post-commit input key are exactly those measured; integrated-source.json ties the fully composed tree to the census/emit/declaration candidate.

Setup used a fresh detached private worktree, recursive depth-one submodules, upstream v6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8, Node v24.19.0 and Go go1.27.1. All writes and caches were in private scratch. The first counting attempt was invalidated by one OOM kill in RAM-backed /tmp and is preserved explicitly as invalid. Scratch was relocated to disk-backed /var/tmp; the complete retry and final lane had no additional OOM kill. No incomplete count was accepted.

See COMMANDS.md, raw compressed logs/censuses, delta.json, delta-full.json, span-audit.json, declarations.diff and final-lane-verdict.txt. The span audit independently verifies all 184 retained span fields move by exactly the 30-byte deletion, and no other amount.
