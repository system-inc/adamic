# visitorPublic.ts at zero

Integration input: 634ef061fc72c061e2de1606d5c8faebec4411f6, merged by fast-forward
into codex/stage3-indexed-reads-emit. TypeScript source pin and stock API: 6.0.3.
Setup: 97 seconds, nproc 5; individual readiness timings are in setup.txt.

The existing U-loop ledger entry changes from decline to assert. The loop bounds
prove parameters[i] is a member of the populated readonly NodeArray returned by
nodesVisitor. The assertion stays on that single read before its callback.

Unfiltered load.Load census: one TS2345 before, zero diagnostics of any code
after. verify.txt records byte-identical stock JavaScript and adapter contract
idempotence for all 30 partition files. The required-read mutant replaces this
! with ?? 0; mutant.json records both the emitted-JavaScript and site-contract
checks failing specifically on that change. The API validator mutant appends an
unlisted interface to the emitted declaration file; exact projection rejects it.
The emitted declaration file was restored byte for byte.

Default oracle: **106367 passing, 0 failing, 0 pending**, **zero
baseline differences**, 211.991 seconds. The user provisionally approved
adaptation 40's 28 inherited API lines alongside adaptation 20's 189 lines.
check-integrated-api.cjs reconstructs exactly these edits from parsed owners and
compares every other reference byte plus the complete reference file set. The
API and extra-reference mutants both fail; source and reference bytes are restored.
No adaptation 70 is in integration 634ef06, and no additional exception is used.

The first default oracle's sole API failure is retained in initial-oracle files.
An intervening worker exit occurred during a concurrent Go build (cgroup reported
an OOM kill); the successful oracle ran without that build competing for memory.

The source-only latent loader export matches all 382 published input diagnostics
exactly. After only the visitorPublic edit, it records 381, with checker-clean
files **36 -> 37**. latent-provenance.json lists the pinned native feature inputs.
The complete native scratch merge encountered an unrelated lowering conflict;
the census uses only its mechanically composed loader. No lowering result is
claimed. The direct integration loader census is also retained, without filtering
any codes. Source-only census avoids npm's generated-source/dependency side effects.
