# Inputs for optional correctness checks

Run `bash cloud/setup.sh --gate-inputs`, then source the printed env.sh. Ordinary setup does no gate input downloads or archive builds and removes these exports from its env.sh. `--warm-tests` can be combined with the flag. Everything is installed under `$ADAMIC_TOOLS/gate-inputs`, with an outside-/root fallback under /tmp/adamic-gate. All downloads and build children retain the shared bounded runner's deadlines.

For auditing a clean checkout with an installer from another checkout:

```
ADAMIC_SETUP_REPOSITORY=/path/to/clean-main bash /path/to/branch/cloud/setup.sh --gate-inputs
source /path/printed/by/setup/env.sh
ADAMIC_GATE_UNCACHED=1 python3 /path/to/branch/cloud/run-gate-input-tests.py /path/to/clean-main /tmp/gate-with with
```

The installer reads npm manifests from its own cloud directory and builds the archive from the selected checkout. It does not copy source into the checkout being audited. A plain setup, source env.sh and the runner's `without` mode provide the no-flag proof. The runner refuses inherited gate exports in that mode, runs only the named 17 checks, and records exact commands, flags, loads, Go JSON events and readable logs. It caps concurrency at two for the large compiler and sanitized corpora. A missing test is reported ABSENT, never PASS or SKIP.

| Export | Installed path below gate-inputs | Exact pin and authority |
| --- | --- | --- |
| ADAMIC_TYPESCRIPT_SOURCE | typescript | TypeScript v6.0.3, 050880ce59e30b356b686bd3144efe24f875ebc8; lint/lint_test.go and parser/compilerManifest |
| ADAMIC_CSS_LIBRARY | css | postcss 8.5.16, postcss-scss 4.0.9; CSS library.mjs checks both, REPORT.md uses both |
| ADAMIC_GRAPHQL_LIBRARY | graphql | graphql 17.0.2; graphql_test.go |
| ADAMIC_MEDIA_QUERY_LIBRARY | media-query | postcss-media-query-parser 0.2.3; mediaquery_test.go |
| ADAMIC_SELECTOR_LIBRARY | selector | postcss-selector-parser 2.2.3; selector_test.go and library.mjs. Its corpus extractor also imports postcss without a version check; REPORT.md used 8.5.16, which is selected here |
| ADAMIC_VALUES_LIBRARY | values | postcss-values-parser 2.0.1; values_test.go |
| ADAMIC_JSON_PRETTIER | json-prettier | Prettier 3.9.6; JSON library.mjs checks it, GAPS.md records it |
| ADAMIC_CSS_PRINTER_LIBRARY | css-printer | Prettier 3.9.6; print_test.go and print_library.mjs |
| ADAMIC_GITIGNORE_LARGEST | gitignore/.gitignore | Deterministic exactly 100 MiB, same recipe as gitignore/testdata/cohere_side_test.go |
| ADAMIC_CLANG_TSGO_ARCHIVE | checker/tsgo.a | This checkout's bridge/tsgo/archive, go build -buildmode=c-archive with -trimpath and -buildvcs=false |

Each npm directory has its own checked-in package-lock.json. Pinned npm 11.9.0's bootstrap tarball has a checked-in SHA512 integrity value. `npm ci --ignore-scripts --bin-links=false` runs against immutable copies of both manifests and an empty download cache, checks every package's lock integrity and atomically publishes only a completed installation. No lifecycle scripts or global npm install are needed. Bin links are disabled because these oracles load packages rather than execute npm binaries. Warm keys cover lock, manifest, npm bootstrap metadata, helper implementation and Node version; warm hits validate filenames, bytes and modes. All seven seats use the same tested installer, each with its own key and stamp.

The TypeScript shallow fetch names exactly the pinned commit, then checks HEAD and runs git fsck. Warm validation includes its actual checkout bytes and modes, the URL, commit and helper implementation, and HEAD. The deterministic ignore file contains `big\n#`, x bytes to fill the size, then a final newline; its length and SHA256 are verified. Warm validation checks the file bytes and generator inputs. Important: gitignore's test reads ADAMIC_GITIGNORE_LARGEST only as a nonempty opt-in, not as a file path. It generates the same corpus itself in its own scratch tree; the installed file is provided as requested and the export enables that existing behavior. No harness changes or skips are introduced.

The checker stamp includes HEAD, Go version, Go environment, C compiler version, explicit flags, helper bytes and dependency build IDs obtained after Go validates sources, C headers, embeds, replacements and compiler options. Dirty source edits invalidate it. The archive and header bytes/modes are also validated on a hit. Inputs are checked again after compilation; a persistent concurrent change prevents publication. Deterministic build flags remove checkout and output-directory paths from the archive. `ADAMIC_GATE_UNCACHED=1` bypasses every new stamp and the tests compare the resulting artifact bytes with cached installation bytes.

Verification drivers: test_gate_inputs.py, test_markdown_setup.py with ADAMIC_MARKDOWN_SETUP_MODULE pointing at setup-gate-npm.py, and gate-input-mutants.py. Set ADAMIC_SETUP_INTEGRATION=1 for real npm and Go rebuild proofs. Logs and the first current-main verdicts live in cloud/reports/gate-inputs.
