# Daily meter

The stage 3 meter's daily recipe (stage3/meter/README.md) as a systemd timer on a Linux build
box, so it needs no one's session. Never on a Mac: both scripts refuse to run there.

- `run.sh` runs the meter once, as adamicmeter (network, no keys, no push rights, no sudo), on
  main's tip. It clones adamic over HTTPS into its own checkout, detaches it at the sha origin/main
  resolves to, and runs twice-daily.sh with `per-ref`, `first-error` and a 3 hour limit (2 hours until Nov 5; back to 2 when #20pb346 is fixed), pinned to
  that sha. Its state directory, /var/lib/adamic-meter (adamicmeter's, with an ACL so ahra can read
  it), holds `work/` (the checkout), `runs/<stamp>/` (the output, plus `toolchain.txt` naming the
  node, go, clang and npm it ran with), `logs/<stamp>.log` (every step and the meter's whole output)
  and `status`.
- `status` is key=value: `outcome` (Running, Succeeded, OverBudget, BudgetUnknown, Skipped or Failed), the
  `step` it reached (toolchain, clone, fetch, checkout, submodules, npm, meter), the pinned `sha`, the
  `run` directory, the `log` and the failing step's last line. A failure keeps the log and pushes nothing.
  The meter exits 3 when its record is whole but the tsc entry pass's relation budget isn't ok
  (#ny0c3gb), and `outcome.py` reads the verdict from report.json: red is OverBudget, unknown or
  unreadable is BudgetUnknown, and such a run is still committed, its verdict in the commit's first line.
- On a line box the timer runs outside any of Loom's gaps, so `run.sh` never starts beside a hand job
  (`hand-jobs.sh`, Loom's condition on Pupa, Nov 9). A hand job is a child of the box's keeper,
  `loom control keep`, in the process table anyone can read. While the keeper has a child, the run waits in
  60 s steps for up to 30 minutes, then skips the day: status `Skipped`, exit 75, nothing measured and nothing
  submitted. A box with no keeper hands out no jobs, so the run starts at once. cloud/meter_daily_test.go
  holds it, with a fake process table.
- `submit.sh`, as ahra, reports the last status, then commits the finished run's slim record alone on a fresh
  branch `meter/daily-<stamp>` cut from origin/main's tip, written by `slim.py` and
  nothing force-added past git's ignore rules, pushes it with the deploy key and posts it to Loom's Queue for
  system_adamic_compiler_correctness: `POST https://loom.system.inc/changes` with the submit token
  as a bearer `Authorization` header and the body `{sha, base, owner, paths}`, the call
  `ahra loom submit` makes with `--base` off. It exits 3, "not configured", pushing nothing, until /etc/adamic-meter/submit.env names
  `ADAMIC_METER_DEPLOY_KEY` and `ADAMIC_METER_SUBMIT_TOKEN`, each a readable file. `--dry-run`
  builds the commit and the body and prints them without either secret.
- `meter-daily.service` (User=adamicmeter, TimeoutStartSec=4h, its own /tmp) starts
  `meter-daily-submit.service` (User=ahra) only on success, which counts 3 (SuccessExitStatus=3). `meter-daily.timer` starts it daily at
  06:17 America/Denver, Persistent=true.

A committed record under `stage3/meter/runs/<stamp>/` holds the run's summary and the rows readers
read, never its raw files (#jnx6n47, Kirk's ruling of Nov 24): a whole run came to about 13 MB a
day, and every clone pays for what the history holds. `slim.py` writes it: `README.md`,
`report.md`, `main/report.md` and `area/report.md`, a slim `report.json` keeping only the keys a
reader of a committed record reads (its `KEPT` table names each key and its reader, among them
`trees.main.latent_lowering.reason_rows` and the counts), and `kept.json`, which names every file
the run held that the commit leaves out, with its bytes and sha256 and the job that kept it
(`--job`, from `ADAMIC_METER_JOB`; null without one). The logs, raw streams, unit times, readiness
probes and full reports stay with the job, its stdout kept by hash and the run's tarball on the
box, so every number quoted has a commit behind it. `slim.py` refuses a record holding a `.log` or
`.log.gz` file or any file over 1 MB, and `records_test.py` holds every record newer than
`20261122T131040Z.tNkys1` to the same; the 43 records up to it keep their raw files, since
rewriting public history isn't worth it.

What adamicmeter needs: HTTPS out to github.com, the npm registry and the Go module proxy, and
read on the box's toolchain env.sh (`ADAMIC_METER_TOOLS_ENV`, default
/home/ahra/adamic-tools/env.sh) with Go, Node 24, clang, gofmt, python3, git and curl. npm comes
from the pin in cloud/markdown-width/npm-bootstrap.json when the toolchain has none.

Install, then prove one run by hand before the timer. The scripts live in /srv, never /opt: a line box's /opt
stays an empty, stamped listing, since clang's driver lists /opt in every unit that runs it.

```sh
sudo useradd --system --create-home --shell /usr/sbin/nologin adamicmeter
sudo setfacl -m u:adamicmeter:--x /home/ahra
sudo setfacl -R -P -m u:adamicmeter:rX -m d:u:adamicmeter:rX /home/ahra/adamic-tools
sudo install -d -o adamicmeter -g adamicmeter -m 750 /var/lib/adamic-meter
sudo setfacl -m u:ahra:rx -m d:u:ahra:rx -m d:u:adamicmeter:rwx /var/lib/adamic-meter
sudo install -d -m 755 /srv/adamic-meter
sudo install -m 755 cloud/meter-daily/run.sh cloud/meter-daily/hand-jobs.sh cloud/meter-daily/submit.sh cloud/meter-daily/outcome.py cloud/meter-daily/slim.py /srv/adamic-meter/
sudo install -m 644 cloud/meter-daily/meter-daily.service cloud/meter-daily/meter-daily-submit.service cloud/meter-daily/meter-daily.timer /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl start --no-block meter-daily.service
cat /var/lib/adamic-meter/status
sudo systemctl enable --now meter-daily.timer
```

Follow a run with `journalctl -u meter-daily.service -u meter-daily-submit.service` and the log
`status` names. There is no route yet for a box to comment on a task, so a failure's report is
the status file, the log and the journal.
