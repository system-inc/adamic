# YAML baseline audit: port incomplete

No Adamic YAML parser, printer, or stdout driver was built.
This directory records baseline evidence, not a completed stage 1 slice.
Go matches its formatter oracle on all 36 YAML files in this checkout.
Pinned JavaScript parser comparisons passed fixtures and generated cases.
Native parity, performance measurements, and three port mutants remain unrun.

## Revisions and setup

Branch `codex/stage1-yaml` starts at main
`1a4e29984b0b1727b063e76fe01c2aa11dfa56c7`.
Cohere is pinned at `715ba94f3608a6500086b1076ce5cb7e51b836db`.
The checkout's 36 YAML files contain 200,885 source bytes: 35 under
TypeScript and one under cohere. The inventory excludes `.git` only.

`bash cloud/setup.sh` exited zero. Its reported timings were Go 0s,
clang 1s, Node 1s, submodules 1s, build cache 78s, total 78s.
`nproc` reported 5; CPU quota was four CPUs (`cpu.max: 400000 100000`).
Tools: Go 1.27.1, clang 20.1.8, Node 24.19.0.
The environment file is `/workspace/adamic-tools/env.sh`.
The complete setup output is [audit/setup.log](audit/setup.log).

## External baseline

Scratch npm dependencies were installed with exact versions:
Prettier 3.9.6, yaml 2.9.0, yaml-unist-parser 3.2.0.
The formatter tests use cohere's vendored fork bundles, not the npm
Prettier package, as their printer oracle. Lexer, composer, and unist
comparisons use the pinned npm packages under Node.
No direct npm Prettier formatting comparison is claimed.

The first Go test run passed but skipped external comparisons because
no Prettier fork path existed. After npm installation, a second run
failed: setting `COHERE_PRETTIER_FORK` also overrides the formatter's
bundle path, and npm does not provide `dist/prettier/standalone.js`.
A scratch symlink to cohere's vendored bundles supplied that path.
The corrected run passed all four packages:

| Package | Seconds | Comparisons |
| --- | ---: | --- |
| yaml | 11.202 | Formatting fixtures; 36/36 YAML files, full stack and tree loader |
| compose | 5.278 | 735 fixtures, 30,000 generated texts, 36 corpus files |
| cst | 2.191 | 586 fixtures, 2,930 chunked parses, 20,000 generated texts, 36 corpus files |
| unist | 5.110 | 444 fixtures, 20,000 generated texts, 42 corpus inputs |

The extra six corpus inputs are Markdown front matter selected by
cohere's existing tests. Their Go comparisons passed; this does not
constitute an Adamic front matter implementation.

`TestFormatOracleCanFail`, the three parser-layer `TestOracleCanFail`
tests, and `TestComparisonSeesEveryField` passed. These are existing
Go oracle checks. They are not the requested three Adamic port mutants.

YAML test-suite comparisons and the broader prose-wrap oracle test
skipped because the upstream test-suite fixture directory was absent.
All test output was redirected to log files. The successful external
run is [audit/go-upstream.log](audit/go-upstream.log).

## Reproduction

Run from the repository root, substituting its absolute path for
`/workspace/adamic` if necessary:

```sh
bash cloud/setup.sh > /tmp/stage1-yaml-setup.log 2>&1
source /workspace/adamic-tools/env.sh
npm install --prefix /tmp/stage1-yaml-library --save-exact prettier@3.9.6 yaml@2.9.0 yaml-unist-parser@3.2.0 > /tmp/stage1-yaml-npm.log 2>&1
mkdir -p /tmp/stage1-yaml-library/dist
ln -s /workspace/adamic/cohere/internal/format/prettier/bundles /tmp/stage1-yaml-library/dist/prettier
cd cohere
COHERE_PRETTIER_FORK=/tmp/stage1-yaml-library COHERE_YAML_CORPUS=/workspace/adamic go test -v -count=1 -timeout 30m ./internal/format/yaml/... > /tmp/stage1-yaml-external-fixed.log 2>&1
```

## Unfinished work

The formatter depends on the CST lexer/parser, document composer,
unist transformation and comment attachment, then the YAML printer and
document layout engine. The survey covered representative source from
these layers; it did not finish reading every implementation file.
The GraphQL reference was read. The JSON reference reading remains
incomplete, including portions of its generated tables and auxiliary
performance tooling. No implementation was edited before those reads.

No language blocker is established by this audit. There are no new
Adamic gap claims or proving programs. No file-formatting driver,
Adamic corpus comparison, native sanitizers, Go/Node/native texts-per-second
measurements, three port mutants, filtered Adamic oracle or full repository
gate was run. This audit must not be treated as a merge-ready YAML port.
