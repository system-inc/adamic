# Linked cohere command scout

This is step 42's explicit-file formatting composition, written in Adamic `.a`.
JSON, CSS, YAML and GraphQL share a binary with the existing partial TypeScript
printer. Use a scratch tree with explicitly resolved Prettier defaults. The
command does not yet resolve settings or reproduce cohere's diagnostic/reporting pipeline.
Unsupported flags and unsupported formatter shapes fail loudly.

```sh
source /workspace/adamic-tools/env.sh
go run ./cmd/adamic build stage1/cohere/command/main.a -o /tmp/adamic-cohere
/tmp/adamic-cohere --format-only --format-all --no-cache --directory /tmp/project file.json style.css
ADAMIC_COMMAND_KEEP=/tmp/scout42-artifacts go test -v -count=1 -timeout 30m ./stage1/cohere/command
```

The test compiles emitted C to an object, puts it in `cohere.a`, and links that
archive with stage 0's cached runtime archive. It records binary size and final
link duration separately from C compilation. Release and sanitized native plus
source Node must produce Go cohere's exact file bytes. It also exercises Go's
real command flags, shortest command/printer boundaries, and three successful
wrong-output composition mutants. No external corpus setting enables skips.
The optional artifact path retains the release binary, archive and emitted C.

`testdata/fetch_corpus.py` shallow-fetches the 23 public commits supplied by the
user, then reads only selected source blobs. It installs nothing and executes
no downloaded code. The saved files and hashes make normal tests offline.
`testdata/inventory.py` and `inventory_go.go` reproduce the complete source
name/state/helper inventories. See SCOUT.md for scope, evidence and questions.

For the larger lint-plus-formatter probe, run
`bash stage1/cohere/command/testdata/whole.sh` from the root with the cloud toolchain
on PATH. It copies stage1 into scratch, generates registration there, builds
the unchanged checker archive, measures the whole-program archive link and
compares native output with Node. The measurement wrapper targets this cloud's
Linux clang; no shared source or generation output is changed.


The settings follow-up lives in `settings/`. Run its mandatory offline checks:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_SETTINGS_KEEP=/tmp/scout42-settings-evidence go test -v -count=1 -run '^TestSettings' -timeout 15m ./stage1/cohere/command
```

Go and Node compare exact settings answers from every held pin and Go settings
controls. The strict JSON reader builds natively and sanitized, with mutants.
The full native resolver and reporting probe deliberately retain two observed
compiler refusals; the tests prove those boundaries. The integer-domain mismatch
is also held as a boundary, not counted as parity. SCOUT.md records the supplied
host rulings and exact shared files needed before command integration. The
inherited formatting CLI still uses its fixed defaults.

`testdata/settings_sources.py` refreshes selected public source blobs and eight
embedded sets; `settings_cases.py` regenerates the Go-derived control list;
`go run stage1/cohere/command/testdata/settings_tables.go` regenerates Go's
quoting table. No downloaded config or repository script is executed.
