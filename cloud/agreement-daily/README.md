# Daily agreement

The admission delta's agreement mode (cmd/adamic-admission-delta/README.md, "Agreement mode") as a
daily sealed Loom job in four shards, and `diff.py`, which holds each day against the one before
(#wfnmhe9). The agreement run judges every program head accepts on Node, the JavaScript backend and a
sanitized native build, so a wrong answer already on main shows here when no head's gate can see it.

- `job.sh <k/n>` runs one shard. The day is four jobs:

  ```sh
  source /opt/adamic-block/environment.sh && source cloud/sandboxed-job.sh && AGREEMENT_OPEN_TASKS="$open" bash cloud/agreement-daily/job.sh 1/4
  source /opt/adamic-block/environment.sh && source cloud/sandboxed-job.sh && AGREEMENT_OPEN_TASKS="$open" bash cloud/agreement-daily/job.sh 2/4
  source /opt/adamic-block/environment.sh && source cloud/sandboxed-job.sh && AGREEMENT_OPEN_TASKS="$open" bash cloud/agreement-daily/job.sh 3/4
  source /opt/adamic-block/environment.sh && source cloud/sandboxed-job.sh && AGREEMENT_OPEN_TASKS="$open" bash cloud/agreement-daily/job.sh 4/4
  ```

  `$open` is every task a pending sidecar at the asked commit names that is open when the day is
  asked, ids without their `#`, separated by spaces. The sealed job can't ask the task store, so the
  asker reads it, and the tool's `--open-tasks` holds it: a sidecar naming any other task excuses
  nothing, and its program is judged as if it had none (#92cnfpf). Unset, the job runs nothing and
  exits 64; set and empty, no task is open.

  It lays the submodules as repositories (`sandboxed_repositories`), installs stage3/api offline
  (`sandboxed_npm`), builds the tool, writes the manifest at the job's commit and runs the shard on
  half the box's threads, since a run on all of them timed out under its own load (Nov 21). Then it
  prints the compact report, gzipped and base64'd on one line between `AGREEMENT_REPORT_BEGIN` and
  `AGREEMENT_REPORT_END`; the tool's stderr without its progress lines (the timeout line, every
  `disagreeing`, `pending`, `closed sidecar`, `compiler failure` and `stale run-time stop` line, and the summary);
  and last one line `AGREEMENT_DAILY shard=<k/n> head=<sha> exit=<n>`, so the job's ending, which
  quotes its log's last 40 lines, reads as the run and never as base64. The compact report is
  `diff.py compact`'s: the head, shard, verdict, void, void_reason, counts, and each program's
  path, judgment, outcome_hash and task. It exits with the tool's exit, 0 for a pass or a partial
  shard, 1 red and 3 void; 64 is a shard that isn't k/n or AGREEMENT_OPEN_TASKS unset, 65 a preamble step, 66 the build or the
  manifest, and 67 a report.json that didn't parse.
- `diff.py extract <job log>...` reads the four logs back into one day, as JSON: the head, the
  shards, the counts summed and every program by path. It refuses, exiting 2 and naming the log, a
  log with no report, two logs of one shard, shards of different counts, a path in two shards, and
  logs of different heads.
- `diff.py compare <yesterday.json> <today.json>` prints today's head and counts, then its
  sections: `moved` (the judgment changed), `output changed` (the judgment held and
  `outcome_hash` changed), `new`, `gone`, `void` and `pending`, each pending program with the
  task it awaits, for the reader to check the task is still open, since the sealed run can't ask.
  A program void either day is listed under `void` alone, never as moved or changed, since a void
  proves nothing. It exits 0 when moved, output changed, new and gone are empty, and 1 otherwise.

```sh
python3 -I cloud/agreement-daily/diff.py extract shard-1.log shard-2.log shard-3.log shard-4.log > today.json
python3 -I cloud/agreement-daily/diff.py compare yesterday.json today.json
```

A move or an output change opens a task: a human reads that program's three runs and says whether
the change is right. Nothing here decides it. `diff_test.py` holds `diff.py`, and
cloud/agreement_daily_test.go runs it in the gate.
