# Completed L2 proof

The user authorized both the two filled number-array initializers and the six
checked indexed reads. The real-call count completed before implementation.
No Adamic compiler source, lane runner, upstream lib or existing adapter changes.
Both original 30 string-read assertions are unchanged.

| Proof item | Result |
| --- | --- |
| Main base | f0271c6a7b2e92900eb7046d8246cfb89f52cfb6 |
| TypeScript pin | 050880ce59e30b356b686bd3144efe24f875ebc8, 6.0.3 |
| Setup and pinned API install | Exit 0 |
| Full main apply, --build-product | Exit 0 |
| Full main plus this adapter apply, --build-product | Exit 0 |
| Real-call differential count | 798965 calls, 0 differences, 398 accepted shards |
| Unchanged full lane control | Pass; accepted shard identities, counts and baseline diff match exactly |
| Node witnesses | 14 of 14 return values equal; requested and observed classes covered |
| Owner contract | Pass; exact before/after pins, every partial checked-read refusal, LF/CRLF, foreign text and no-write replay |
| Full paired adapter replay | 84 adapters; only this adapter's outcome differs |
| Adapter 30 full-tree replay | Identical preexisting addRange decline refusal on both sides; zero candidate-caused refusal |
| Adapter 30 checkpoint replay with all eight edits | Both exit 0, identical outcomes, zero source changes |
| Final replay sources | All 746 files byte-identical after both replays converge |
| Ordinary adamic types, all 79 roots together | Both exit 0, zero diagnostics |
| Ordinary compiler builds | Both exit 0 |
| Complete core JavaScript, comments kept and removed | Only the two initializers differ; six ! assertions erase |
| All nine emitted JavaScript files | Only _tsc.js, typescript.js differ, at those same initializers |
| typescript.d.ts | Byte-identical |
| tsserverlibrary.d.ts | Byte-identical |
| Source scope | All 746 sources compared; only the complete pinned core.ts owner differs |
| Complete core.ts site census | 59 before, 44 after; F22..F36 removed; zero new sites anywhere |
| Initializer-only mutant | Original allocation stops return; all six ! assertions and every other source byte retained |
| Archive manifests | Every member SHA-256 plus outer archive SHA-256 |
| Worker cleanup | No live task workers at closeout |

F IDs retain the historical baseline ordering from getOwnValues onward; they
are not renumbered after removal. census-comparison.json retains every removed
finding and the complete initializer-only mutant owner findings. Both census
runs use LATENT_FULL=1, LATENT_ONLY_FILE=core.ts, LATENT_ASSERT_NO_OUTPUT=1 and
LATENT_TIMES, exactly as the reference census. No accepted-program findings
are filtered except Boundary markers, and every core.ts site is compared.
The no-output assertions run before the first census record.

The full-tree adapter-30 refusal labels addRange's original from[i] decline
site core.ts:992:21, already changed by later main adapters. It is identical
before and after this edit. The independent post-30 checkpoint pair succeeds
with the exact eight-site owner adapter. No existing adapter is amended.

The original stopping report and its seven initializer-only checker errors
remain historical in initial-report.tar.gz. Two discarded count/control
attempts are in excluded-attempts.tar.gz: the duplicate-module-link setup
failure recorded no calls; the memory-contended partial count recorded 2,611
calls and zero differences, and its parallel full control had three timeouts.
The accepted control and the complete accepted shard union ran in isolation.
No excluded call, seed witness or diagnostic retry inflates COUNT.json.

The existing published branch history is preserved with plain pushes. The
post-apply checkpoint preserves both successful applies and the complete count;
its plain push succeeded after using the provided credential header. Later evidence
files change product keys without changing the frozen adapter or owner pins;
provenance.json checks their hashes against the implementation checkpoint.
Every retained Go source has a .go.txt suffix. All command output is logged.

Reproduction uses the exact scratch paths retained by the scripts, after
source /workspace/adamic-tools/env.sh, bash cloud/setup.sh and npm ci --prefix
stage3/api, with ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source at the
pinned commit. instrument.cjs, measure-prepare.py and count.py build and count
both bodies; reconcile.py checks the archive union against the unchanged full
control. after-apply.py, post-apply.py, replay.py, checkpoint30-six.py and
census.py record the remaining proof steps. The initial report archive is
historical; its runner cleanup transcript contains old PIDs and must not run.
