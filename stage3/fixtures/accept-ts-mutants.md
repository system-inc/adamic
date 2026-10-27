# Stage3 fixtures accept TypeScript sources

Base: `origin/main` at `d579ff21e10791f78ab269628849513f7db5a39c`.

`validFixturePath` accepts exactly `.a` and `.ts`, retaining the existing portable-path guards. Directory discovery recognizes either source extension. `TestFixturePaths` accepts `main.ts` and rejects `main.js` and extensionless `main`. No fixture, status, manifest, expected output, or recorded evidence is renamed or changed.

## Extension audit

Repository searches: `git grep -n '\.a' -- stage3/fixtures`, supplemented by searches for `TrimSuffix`, `HasSuffix`, `Glob`, Python `glob`/`rglob`, suffix replacement, regular expressions, and `.a.` artifact names. Searches cover the entire fixture tree, including its Go, Python, JavaScript and shell tools.

Updated discovery in `cycles/audit.py`, `entries/check-headers.py`, `entries/check.py`, `namespaces/observe.py`, `nested-functions/record.py`, `objects/audit.py`, `records/observe.py`, `records/verify.py`, `host/source-audit.cjs`, and `nested-functions/verify-source.cjs`. Each accepts either extension and retains the source's actual filename. `host/check.py` strips the source's own suffix when forming its temporary binary path. Existing `.a` output names are unchanged. No recorded name needs changing.

The Go runner uses `entry.File` directly for source paths and test identities. No `.a` trim, join, or source glob remains there. Explicit named fixtures and extraction specifications remain unchanged. Historical reports and logs remain unchanged. `records/probe-inputs.py` intentionally converts only `.a` inputs to `.ts` for stock tsc; existing `.ts` inputs already have the target extension. Artifact helpers using `Path.stem` or `with_suffix` already derive the extension from each input.

Python syntax compilation and `node --check` for both edited JavaScript tools passed. `git diff --check` passed.

## Mutant

Temporarily restored only `validFixturePath`'s `.a`-only extension check, keeping the new `main.ts` expectation. Command: `go test ./stage3/fixtures -run '^TestFixturePaths$/^main.ts$' -count=1 -v > /tmp/accept-ts-mutant.log 2>&1`. Exit 1; zero passing test events, two failing test events (parent and subtest). Witness: `TestFixturePaths/main.ts`.

```text
fixtures_test.go:496: validFixturePath("main.ts") = false, want true
```

Restored the implementation before the full suite. An earlier attempt could not compile while setup was still populating `cohere/TypeScript/tsc/go.mod`; it supplied no mutant evidence.

## Validation

Setup: `export GOPROXY='https://proxy.golang.org|direct'; timeout --kill-after=60s 1800 bash cloud/setup.sh > /tmp/setup.log 2>&1` — exit 0. Commands use `source /workspace/adamic-tools/env.sh`. Prepared the pinned TypeScript checkout `/tmp/item8-typescript` at `050880ce59e30b356b686bd3144efe24f875ebc8` and installed locked API dependencies with `npm ci --prefix stage3/api`.

Full suite, with no fixture renamed:

```sh
ADAMIC_TYPESCRIPT_SOURCE=/tmp/item8-typescript go test ./stage3/fixtures/... -count=1 -json -timeout=60m > /tmp/accept-ts-fixtures.jsonl 2> /tmp/accept-ts-fixtures.stderr
```

PASS, exit 0: one package; 34 top-level tests and 742 subtests passed (776 passing test events), one top-level test skipped, zero failures. Package elapsed: 129.492 seconds. Stderr is empty. Fixture source bytes, statuses, manifests, unions, expected outputs, and historical evidence are unchanged.
