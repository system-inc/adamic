import json, subprocess
from pathlib import Path
out=Path(__file__).parent
checks=[json.loads(line) for line in (out/"checks.jsonl").read_text().splitlines()]
mutants=[json.loads(line) for line in (out/"mutants.jsonl").read_text().splitlines()] if (out/"mutants.jsonl").exists() else []
for row in json.loads((out/"grain-mutants.json").read_text()):
 row["markers"]=[row["marker"]];mutants.append(row)
base=subprocess.check_output(["git","rev-parse","origin/main"],text=True).strip()
source=subprocess.check_output(["git","rev-parse","origin/compiler/loom-slice-2-2"],text=True).strip()
text=f"""Built: slice 2 re-cut as one commit on current main, preserving callback lowering, Weak checks, Error semantics and self-comparison.
Commits: base {base}; source {source}; delivery SHA is in the final response and git log.
Commands and outputs: focused fixtures, guards, reader checks and changed-package vet pass; counts add 46 owned rows and change no existing row.
Mutants: all 17 original semantic mutants, four extended probes and one structural-only inline check were caught; every catcher is below.
Not covered: full package suites, the full gate, WASI and macOS. No PR or Loom submission was made.

No textual conflicts occurred. units.go auto-merged main's ae1d1c70 clang jobserver support with the source's inline-helper placement. Main's counts work first added digests (39b99b53, 9810993a), then 1f07ce2d moved measurements into keyed products and made counts.md counts-only again; ac8ffaac and 9f8ce955 improved shard preparation. This fetched tip therefore has no digest column. Conservative choice: preserve current main's table and product API, rather than restore an obsolete schema. The two Weak registrations now return main's preparation tuple. No production lowering or runtime code needed manual adaptation.

Main's table was taken before exactly one Linux -update-counts run, clang 20.1.8. counts-attribution.json lists all 46 additions: 35 callback-widening, three miscompile-2a, five Weak and three self-compare rows. No existing row changed or disappeared. Runtime +/- lines exactly match the source (runtime-comparison.json): adamic.h +2, exceptions.c +15/-2, object.c +2; no differences.

Guards: SortedMap, exact runtime witnesses, literal native agreement registration and parallel-test checks pass. No new compiler-consuming Go package or runtime feature set was added, so neither dependency nor runtime-set declarations needed additions. Three serial reader key-discovery leaves gained their own Not parallel explanation, retaining the source's serial preparation. Native field-layout and placeless split-build guards pass. Reader products retain relative source/export names and build through main's buildcache. TestCallTargetReaders ran with -timeout=90s; vet covered all seven changed Go packages.

Every command used the sourced setup environment, GOPROXY=https://proxy.golang.org|direct, GOMAXPROCS=4, and four-CPU affinity. nproc=5, cgroup cpu.max=400000 100000. Setup completed in 496.716 seconds; all timing lines are in setup.log. A first guard attempt before the nested submodule checkout completed could not load tsc/go.mod and ran no test. The later actual guard run passed. A transient exec transport disconnect occurred during counts; the original bounded process completed successfully and was not repeated.

The source's test splits are retained. Every completed affected leaf remains below 60 seconds; individual timings are in *-seconds.json. Runtime archives were prepared one product group at a time. Test output was always written to logs. Worker runs used focused selectors, not entire package suites. Setup compiled all repository packages. No code was copied from cohere.

## Commands

| Check | Exit | Wall seconds | Command |
|---|---:|---:|---|
"""
for row in checks:
 text+="| "+row["name"]+" | "+str(row["exit"])+" | "+str(row["seconds"])+" | `"+" ".join(row["command"]).replace("|","\\|")+"` |\n"
text+="\n## Mutants\n\n| Mutant | Caught | Seconds | Expected catcher |\n|---|---|---:|---|\n"
for row in mutants:
 text+="| "+row["mutant"]+" | "+str(row["caught"])+" | "+str(row["seconds"])+" | "+", ".join(row["markers"])+" |\n"
text+="\nMutant commands and logs are retained in this review directory. The inline-definition mutant is structural only, caught by clang and not credited as a semantic kill. Mutant Go sources use .go.txt. TestSelfCompareConstantMutant built and ran its NaN mutant successfully and observed the stdout disagreement with Node; the test passes only when that semantic mutant is killed.\n"
(out/"REPORT.md").write_text(text)
