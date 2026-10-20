# Daily meter

The stage 3 meter's daily recipe (stage3/meter/README.md) as a systemd timer on a Linux build
box, so it needs no one's session. Never on a Mac: both scripts refuse to run there.

- `run.sh` runs the meter once, as adamicmeter (network, no keys, no push rights, no sudo), on
  main's tip. It clones adamic over HTTPS into its own checkout, detaches it at the sha origin/main
  resolves to, and runs twice-daily.sh with `per-ref`, `first-error` and a 2 hour limit, pinned to
  that sha. Its state directory, /var/lib/adamic-meter (adamicmeter's, with an ACL so ahra can read
  it), holds `work/` (the checkout), `runs/<stamp>/` (the output, plus `toolchain.txt` naming the
  node, go, clang and npm it ran with), `logs/<stamp>.log` (every step and the meter's whole output)
  and `status`.
- `status` is key=value: `outcome` (Running, Succeeded or Failed), the `step` it reached (toolchain,
  clone, fetch, checkout, submodules, npm, meter), the pinned `sha`, the `run` directory, the `log`
  and the failing step's last line. A failure keeps the log and pushes nothing.
- `submit.sh`, as ahra, reports the last status, then commits the finished run alone on a fresh
  branch `meter/daily-<stamp>` cut from origin/main's tip, its logs gzipped to `.log.gz` and
  nothing force-added past git's ignore rules, pushes it with the deploy key and posts it to Loom's Queue for
  system_adamic_compiler_correctness: `POST https://loom.system.inc/changes` with the submit token
  as a bearer `Authorization` header and the body `{sha, base, owner, paths}`, the call
  `ahra loom submit` makes with `--base` off. It exits 3, "not configured", pushing nothing, until /etc/adamic-meter/submit.env names
  `ADAMIC_METER_DEPLOY_KEY` and `ADAMIC_METER_SUBMIT_TOKEN`, each a readable file. `--dry-run`
  builds the commit and the body and prints them without either secret.
- `meter-daily.service` (User=adamicmeter, TimeoutStartSec=2h30m, its own /tmp) starts
  `meter-daily-submit.service` (User=ahra) only on success. `meter-daily.timer` starts it daily at
  06:17 America/Denver, Persistent=true.

What adamicmeter needs: HTTPS out to github.com, the npm registry and the Go module proxy, and
read on the box's toolchain env.sh (`ADAMIC_METER_TOOLS_ENV`, default
/home/ahra/adamic-tools/env.sh) with Go, Node 24, clang, gofmt, python3, git and curl. npm comes
from the pin in cloud/markdown-width/npm-bootstrap.json when the toolchain has none.

Install, then prove one run by hand before the timer:

```sh
sudo useradd --system --create-home --shell /usr/sbin/nologin adamicmeter
sudo setfacl -m u:adamicmeter:--x /home/ahra
sudo setfacl -R -P -m u:adamicmeter:rX -m d:u:adamicmeter:rX /home/ahra/adamic-tools
sudo install -d -o adamicmeter -g adamicmeter -m 750 /var/lib/adamic-meter
sudo setfacl -m u:ahra:rx -m d:u:ahra:rx -m d:u:adamicmeter:rwx /var/lib/adamic-meter
sudo install -d -m 755 /opt/adamic-meter
sudo install -m 755 cloud/meter-daily/run.sh cloud/meter-daily/submit.sh /opt/adamic-meter/
sudo install -m 644 cloud/meter-daily/meter-daily.service cloud/meter-daily/meter-daily-submit.service cloud/meter-daily/meter-daily.timer /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl start --no-block meter-daily.service
cat /var/lib/adamic-meter/status
sudo systemctl enable --now meter-daily.timer
```

Follow a run with `journalctl -u meter-daily.service -u meter-daily-submit.service` and the log
`status` names. There is no route yet for a box to comment on a task, so a failure's report is
the status file, the log and the journal.
