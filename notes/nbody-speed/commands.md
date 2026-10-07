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

The initial two failed by exit and stdout disagreement. The final numeric probe was extended with a present-property call to exercise both outcomes of the uniform guard, and optional_write_unknown.a preserves the original no-known-offset route. Each final probe was rebuilt and executed as below. All three were temporarily registered and run uncached with:

```bash
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/notes/nbody-speed/optional_write_' -count=1 -timeout 10m
```

The guarded number probe was also run individually with that command and the selection ending optional_write_number.a$. All three failed by the expected exit/stdout disagreement. Their temporary fixture registrations were removed. They are not oracle fixtures in the commit.

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

The following commands were executed for each of the ten internal/oracle/testdata/nbody_*.a programs and each of the three notes/nbody-speed/optional_write_*.a programs. Outputs were stored under /tmp/adamic-gate/nbody-builds/<basename>.*. The two unsupported_*.a programs were also passed to build, which refused them before producing an executable.

```bash
go run ./cmd/adamic build "$file" -o "/tmp/adamic-gate/nbody-builds/$name"
"/tmp/adamic-gate/nbody-builds/$name"
node --disable-warning=ExperimentalWarning -e 'const fs=require("fs"),m=require("module");eval(m.stripTypeScriptTypes(fs.readFileSync(process.argv[1],"utf8")))' "$file"
```

Each supported fixture's native stdout was compared byte-for-byte with its source Node stdout, and native stderr checked to be empty. All agree. All three difference probes compile, Node exits 0, native exits 70. The unsupported probes' exact build diagnostics were read.

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

## Reference dependency recovery

The full uncached gate exited 1 solely on TestMarkdownUnicodeWidths: the scratch reference dependencies had not been installed. Every other package passed. Installed the versions pinned by stage1/cohere/markdownblocks/tools/generate_width/main.go, then reran the failing test:

```bash
npm install --prefix /tmp/adamic-markdown-width --ignore-scripts --no-audit --no-fund emoji-regex@10.6.0 get-east-asian-width@1.6.0 narrow-emojis@0.0.3
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/markdownblocks -run '^TestMarkdownUnicodeWidths$' -count=1 -timeout 30m
```

Installation exited 0; the isolated retry passed in 132.306s. No repository code was changed for this environment fix, and the full command was not repeated.

## Commits and pushes

```bash
git commit -m "Add oracle coverage for field writes and inherited static reads"
git push -u origin coverage/nbody-speed
git commit -m "Cover absent optional write paths and record validation results"
git push origin coverage/nbody-speed
```

The first push succeeded. The second records the completed missing-slot path audit and broad gate result.

Removed trailing spaces from blank diagnostic lines in the saved gate log (the raw log remains in /tmp/adamic-gate/nbody-gate.log), then ran git diff --check and committed the artifact formatting correction:

```bash
git commit -m "Trim blank-line whitespace in the saved gate log"
git push origin coverage/nbody-speed
```
