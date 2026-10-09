# @system_adamic_tests continuity, Oct 9 2026 05:45 MDT

## Where things stand
#xphstyt (full test audit) is in its main fanout (#1b5sppe): 151 real units (13 of the original 164 were testdata side-file dirs, dropped), about 80 done, 15 Codex clones running brief v8, about 57 queued. Ledger about 764 rows. About 27 findings tasks filed to owners tonight (children of 1b5sppe, 7v0fyx6, 4b4ya6g).

## The machinery (all in my scratchpad, which dies with the session; copy before relying on it)
Scratchpad: /private/tmp/claude-501/-Users-kirkouimet-Projects-ahra/a04a0500-8f18-4a73-ba8b-f1ea5f8cf4f1/scratchpad
- manifest/units.json: the unit list (built by manifest/build.py, attached on #1p99664, now skips testdata).
- brief-for.py: brief + manifest + unit id -> the clone's prompt (adds "Your rows" for big-package slices).
- test-audit-brief-v8.md: the current brief (also attached on #system_adamic_tests; v4..v8 history there).
- fanout/state.json + fanout/state.py: the queue, running map (clone -> unit), done, dropped. EVERY edit goes through state.py (flock). A lost write dropped two units once.
- fanout/dispatch3.sh: run detached (Bash run_in_background) as `bash dispatch3.sh $S $S/test-audit-brief-v8.md >> $S/fanout/dispatch.log 2>&1`. Saves each reply to fanout/replies/<unit>.md, requeues environment failures, hands the next unit out. A Monitor only tails dispatch.log (re-arm every 30 min). Never wrap the dispatcher itself in a Monitor: expiry kills it.
- fanout/clones.txt: the 15 clone session ids (5 warm originals + 10 from the 03:53 ramp step).
- fanout/ramp2.sh: STOPPED at 04:56 on purpose. Loom is measuring the Codex account cap itself (one account, kirk@kirkouimet.com, shared with Loom's pools). Don't restart until Loom reports; @system_adamic's rule: steps of 10 only while the codex pool is idle, back out if asking workers fall.
- fanout/merge.py: folds replies into test-audit-ledger.json; falls back to rows.json or REPORT.md on the unit's evidence branch.
- fanout/findings-batch.md: every survivor seen, [FILED] marks, owners learned.
- replay/home.sh: THE key tool. `home.sh <worktree> <base> <package dir> name=diff ...` applies each mutant at the unit's base in my scratch worktree (wt-fresh, with cohere and TypeScript attached as worktrees, stage3/api npm ci'd) and runs the mutated file's own package. About a third of slice survivors die here (buildcache, internal/load options, namespace init). Run it before filing any slice survivor. internal/lower runs in about 12 s; internal/native and internal/oracle are too slow whole (use a targeted -run).
- replay/replay.sh: the central fast-path replay; parked. Loom's fast path voids at the 20-min ceiling for compiler mutants. Waits on the coverage line map (#mp71kkr: patch + cover.sh reviewed by Loom, to run on an idle star pool before a new base).

## Owners (learned the hard way)
compiler: internal/lower, internal/ir, internal/javascript, bridge/tsgo, internal/fuzz, internal/native emission. typescript: cmd/adamic-meter (stage 3's meter). library: cmd/adamic-test262, internal/regexp. runtime: internal/native runtime (C, heap, keys). developer tools: childguard, adamic-gate, cloud, buildcache. system_cohere_adamic: stage1/cohere ports.

## Decisions in force
- Brief stays v8 for the run; a brief bug means stop dispatch, fix, resume (did it once, v7->v8, family rule). v9 notes in findings-batch (inline same-check rows are a family).
- @system_adamic: one design task to compiler for a behavior helper (JS backend on Node vs Node on source, plus empty-answer refusal) for internal/lower's acceptance-only rows. #36f7nd7 holds the replay list; close per-survivor tasks only when their mutant fails against the helper.
- File findings in bundles by owner; slice survivors only after a home replay; say "repo-wide pending" when oracle might catch downstream.

## Next
Keep merging returns, home-replay survivors, file by owner. Pending: u001 M2 and u004 M06/M07 (bridge/tsgo, slow package). When the queue empties: defender wave (#az6b13b) over subsumed/untrue/vacuous rows, then the central replay (#fm89tdd) once #mp71kkr's map exists. Commit replay.sh, home.sh, state.py and the dispatcher into a real home (loom's pilots/adamic-gate or a test-audit tools dir) before this session ends.
