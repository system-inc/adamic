# Trimmed table reader mutant checks

Task h2c34ad, 2026-11-07. Fetched `cohere/zero-0.2.0` into
`FETCH_HEAD` at `d45ce11df2c7871fb4d3382a57b8765f115ecbf1`.
`git merge-base --is-ancestor 27d8b9cdba FETCH_HEAD`: exit 0.

The surveys require their recorded checker inputs. Exports were reproduced
from Adamic `0806299dc`, cohere `715ba94f`, checker `8d550c837c`, and upstream
TypeScript v6.0.3 (`050880ce59e30b356b686bd3144efe24f875ebc8`). The layer
export uses optional adaptations only, before return adaptations, via a scratch
Go overlay. Both export tests exited 0. The zero tip's newer checker initially
failed the Python population assertion (359 TS18048 findings versus 354);
including return adaptations also shifted layer locations. Neither was used
for the checks below.

Commands from the work branch, with stdout/stderr saved under
`/tmp/h2c34ad-logs/` and each child's exit code reported by the runner:

```sh
python3 docs/tsc-strictness/audit.py /tmp/h2c34ad-logs/diagnostics.json /tmp/h2c34ad-typescript
TSC_SURVEY_TYPESCRIPT=/tmp/h2c34ad-npm/node_modules/typescript/lib/typescript.js node docs/tsc-strictness/layer-audit.cjs /tmp/h2c34ad-logs/diagnostics-optional.json /tmp/h2c34ad-typescript/src/compiler
/tmp/h2c34ad-cohere-bin --format-only docs/tsc-strictness.md docs/tsc-strictness-next-layer.md
```

The formatter was built with `go build -o /tmp/h2c34ad-cohere-bin ./command/cohere`
from the zero branch's clean cohere pin `06af2e399c6c1fb93fd570b774027942747400b4`;
build and `--version` exited 0. npm has no `@system-inc/cohere@0.2.0` package.
Formatting checked exactly two files and exited 0.
Both full audit commands exited 0 before formatting, after formatting, and
after restoring the mutants (logs `before-{0,1}`, `after-{0,1}`, `restored-{0,1}`).

Each mutant changed just one trimmed table cell in the formatted document,
then ran its full audit command above. All failed at the table assertion:

- `tsc-strictness.md`: E-init count 30 to 31, exit 1; E-init name to MUTANT-E-init, exit 1.
- `tsc-strictness-next-layer.md`: A-loop count 35 to 36, exit 1; A-loop name to MUTANT-A-loop, exit 1.

Logs are `mutant-{0,1}-{count,pattern}.log`. Each document was restored byte for
byte after every mutant. No mutant remains in the committed documents.

Final `cohere --no-fix --format-only` on the same two paths and
`git diff --check` both exited 0. All task subprocesses completed.
