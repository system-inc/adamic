# Probe runner

`run.sh` runs a directory of `.a` and `.ts` probes on Node and on both of Adamic's backends, and prints one table: for each probe and each build, whether the JavaScript backend and the native binary (under ASan and UBSan) agree with Node, differ from it, or stopped at compile time and why. It's the runner Language's attacks use, checked in so no window rebuilds it (#y1p4qzt).

```sh
cloud/probe/run.sh probe/mine                     # the head this checkout is at
cloud/probe/run.sh probe/mine --against origin/main  # the head and main, side by side
cloud/probe/run.sh --check-control                # the runner's own proof
```

It needs Go, clang and Node where it runs, a box or a Codex task, never the Mac. A backend agrees only when its stdout and exit code are Node's byte for byte, and Node runs the source through `oracle/node.mjs`, never anything Adamic lowered. Full output for every probe goes to the logs directory the table names.

On Codex, push a branch holding the probes and ask for `cloud/probe/run.sh <directory> --against origin/main`, written into a file the task leaves as its diff.

`--check-control` fails unless `control/agrees.ts` agrees everywhere and `control/differs.a`, Kirk's Oct 5 difference (a write past an array's end panics where Node grows it), differs everywhere. A runner that read everything as agreement would pass any probe, so that's what its proof rules out.
