# Wave 08 recovery landing

Seven winning rule copies from claims/wave1-08.md. Eleven duplicate claims remain assigned to the winning workers in DEDUP_LEDGER.md; no losing copies are included here.

Based on lint-batch/wave2-02 95968dd93ad0876245f181af63931c3134e14b3d, with area/stage1-lint 334509eea8a49b8085187e206c495cc6aa24c5c4 merged. All implementation changes are within these rule directories. Rules take the supplied node and declare named kinds and node=true in rule.json. Shared reporting preserves automatic edits and suggestions; no private transport or helper copies.

| Rule | Upstream cases | Mutant |
|---|---:|---|
| no-useless-computed-key | 122 | computed_replacement_wrong |
| react-hooks/gating | 42 | gating_invalid_suppressed |
| react/forbid-foreign-prop-types | 60 | foreign_read_suppressed |
| @typescript-eslint/no-unnecessary-type-constraint | 43 | constraint_suggestion_wrong |
| @typescript-eslint/prefer-as-const | 69 | const_append_wrong |
| @typescript-eslint/prefer-enum-initializers | 21 | enum_second_suggestion_wrong |
| default-param-last | 128 | core_default_suppressed |

Observed: all 485 upstream cases and owned witnesses match Go cohere byte for byte on source Node, emitted JavaScript and sanitized native. Each listed mutant compiled and was caught by comparison on each of those three backends. No recovery, JSX or option rows were excluded. TestWave08Complete passed in 1425.59 seconds.

Observed: TestWave08CompleteCorpus passed in 357.03 seconds, comparing 1080 compiler/stage1 files (including .a), 7560 selected rule rows, and 109390064 identical output bytes across the three backends and Go. Compiler input is TypeScript 6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8. Go cohere is pinned at 7945d102a6c18dd36adf9114a758ce646e8b2359.

Reproduce from this directory with the toolchain environment sourced:

```sh
python3 certify.py --typescript /path/to/TypeScript --log /tmp/wave08-certification.log
```

certify.py uses a Go overlay adding the owned certification_test.go.txt as a test, without replacing or editing shared harness files. The executed command is go test -overlay=<overlay> ./stage1/cohere/lint -run '^TestWave08Complete' -count=1 -v -timeout=45m -parallel=2 with ADAMIC_GATE_UNCACHED=1 and ADAMIC_TYPESCRIPT_SOURCE set. Original run log: /tmp/wave08-unpark-certification-v3.log.

Setup: GOPROXY=https://proxy.golang.org|direct; cloud/setup.sh passed in 944.247 seconds; nproc=5. All seven oracle.go files and certification_test.go.txt were gofmt checked.

No runtime-compiled regex, missing checker facts, React Compiler IR or control-flow graph is needed by these seven rules. Coverage excludes other workers' duplicate copies and future new rules. Full-package results are recorded alongside this file in package-results.json.
