This is a `.a` candidate for `structure/tailwind-no-physical-direction`, using the unmodified Go cohere rule as the oracle. The registry descriptor, Go adapter, semantic mutant, message and witness are owned by this directory. There are no automatic fixes in this Go rule.

The complete compiler and stage1 corpus agrees on source Node, emitted JavaScript and sanitized native. One upstream JSX case cannot run through the current shared parser. This is not a complete port. The shared registration generator still requires `rule.ts`; the checked-in compatibility patch is applied only to temporary Go overlays for validation. It is not applied to shared repository files.

The patch started from the `.a` validation overlay on `origin/codex/lint-wave1-12`, then added filename preservation for upstream captures and a coverage classifier. Go parsing chooses TSX for `.tsx`/`.jsx`. The coverage classifier uses a throwing panic solely to enumerate unsupported inputs. Actual comparisons use the unmodified Adamic runtime and unmodified Go rule implementations. Invalid files are excluded explicitly because Go's fix engine refuses them; exclusions appear in the logs.

Reproduce from the repository root after sourcing the toolchain environment:

```sh
python3 stage1/cohere/lint/rules/structure-tailwind-no-physical-direction/prepare-validation.py /tmp/wave02-validation
export GOFLAGS=-overlay=/tmp/wave02-validation/overlay.json
export ADAMIC_TYPESCRIPT_SOURCE=/path/to/TypeScript
# The checkout must be 050880ce59e30b356b686bd3144efe24f875ebc8 (v6.0.3).
go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestWave02UpstreamSubset|TestWave02CompilerSubset|TestWave02JSXBlocker|TestWave02SourceThroughput|TestWave02Throughput)$' -count=1 -v -timeout 20m > /tmp/wave02-validation.log 2>&1
go test ./stage1/cohere/lint -run '^TestMutants/(empty_directive_reason_accepted|physical_direction_exemption_omitted)$' -count=1 -v -timeout 10m > /tmp/wave02-mutants.log 2>&1
```

The two throughput tests distinguish real compiler/stage1 files from 1,000 repeated owned witnesses. Each number is one end-to-end run including process startup, file reads and parsing. Native throughput uses an optimized build; byte comparisons and mutants use ASan/UBSan builds. Raw successes and failures are preserved in `evidence/`.
