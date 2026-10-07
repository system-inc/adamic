# Commands and observations

Working directory: /workspace/adamic. Test output was redirected into /tmp/adamic-gate and read after completion. Setup used the printed tool location /workspace/adamic-tools/env.sh.

## Branch inspection

```bash
cat CLAUDE.md
cat README.md docs/0.1.md docs/memory.md
git fetch origin codex/nbody-speed:refs/remotes/origin/codex/nbody-speed
git fetch origin main:refs/remotes/origin/main
git switch -c coverage/nbody-speed origin/codex/nbody-speed
git log --format=fuller origin/main..origin/codex/nbody-speed
git diff origin/main...origin/codex/nbody-speed
```

The first diff was made before refreshing main and was too large to inspect in one tool result; the refreshed comparison and changed compiler/runtime files were read separately. Additional rg searches and reads checked existing oracle programs and the helpers called by the branch.

## Setup

```bash
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
nproc
```

## Differential runs

```bash
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/nbody_' -count=1 -timeout 10m
```

Initial seven fixtures passed. With the two notes programs temporarily registered in fixtures:

```bash
go test ./internal/oracle -run 'TestNativeAgreesWithNode/notes/nbody-speed/optional_write_' -count=1 -timeout 10m
```

Both failed by exit and stdout disagreement. Their temporary fixture registrations were removed. They are not oracle fixtures in the commit.

```bash
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/nbody_static_collision.a$' -count=1 -timeout 10m
```

Ran with the one-line mutation described in coverage.md: failed under UBSan after successful compilation. Restored the original line, then reran the nbody_ selection: passed.

```bash
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/nbody_slot_kinds.a$' -count=1 -timeout 10m
```

Three versions: optional boolean field refused, heterogeneous scalar union field refused, then the final supported boolean/string/function program passed. The refused versions were saved under notes/nbody-speed/unsupported_*.a.

```bash
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/nbody_runtime_fields.a$' -count=1 -timeout 10m
```

Runtime Error stores passed.

```bash
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/nbody_narrowed_receiver.a$' -count=1 -timeout 10m
```

A final emitter review found that heterogeneous union narrowing can produce a cast receiver. This additional fixture passed and its generated C confirms the non-C-name write fallback. The spread fixture was also refined from TypeScript private syntax to JavaScript #private storage to exercise public-layout compaction, then rerun with the same single-fixture command using nbody_spread_fields.a. This moved that new row by one retain and release; the counts check caught it, and counts were regenerated before the successful final run. The full gate had already started before these final fixture refinements; the final complete new-fixture run and final counts regeneration below include all ten fixtures.

Final complete new-fixture run:

```bash
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/nbody_|TestCountsAreRecorded' -count=1 -timeout 10m
```

## Standalone programs

The following commands were executed for each of the ten internal/oracle/testdata/nbody_*.a programs and each of the two notes/nbody-speed/optional_write_*.a programs. Outputs were stored under /tmp/adamic-gate/nbody-builds/<basename>.*. The two unsupported_*.a programs were also passed to build, which refused them before producing an executable.

```bash
go run ./cmd/adamic build "$file" -o "/tmp/adamic-gate/nbody-builds/$name"
"/tmp/adamic-gate/nbody-builds/$name"
node --disable-warning=ExperimentalWarning -e 'const fs=require("fs"),m=require("module");eval(m.stripTypeScriptTypes(fs.readFileSync(process.argv[1],"utf8")))' "$file"
```

Each supported fixture's native stdout was compared byte-for-byte with its source Node stdout, and native stderr checked to be empty. All agree. Both difference probes compile, Node exits 0, native exits 70. The unsupported probes' exact build diagnostics were read.

Generated C inspected for each new oracle program:

```bash
go run ./cmd/adamic c "$file"
```

## Counts and gate

```bash
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts
gofmt -l cmd internal
go vet ./...
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./...
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m
git diff --check
```

Counts were regenerated as the seven, eight, nine and ten-program fixture sets were completed. The final table must contain ten added rows. No pre-existing row is intentionally changed. gofmt output was empty; go vet exited 0. See the saved logs for the final counts, oracle and gate results.
