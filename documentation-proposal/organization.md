# Adamic organization proposal

For Kirk and Ahra. Proposal only, October 7, 2026. No existing file is moved, renamed, removed or edited by this unit.

Measured base: `origin/main` at `ef3d907ecdc4c771b016f7d9c52372def057a340`. Branch: `codex/organization-sweep`. The stated roughly 150 branches in flight is planning context from the task, not a branch count measured here.

## Measurement and scope

The main repository contains **1,864 tracked blobs**, **298,983 text LF lines**, and **19,856,295 bytes** at the base commit. Counts include fixtures, notices, hidden tooling, generated data and evidence, not just production code. The 1,864 blobs comprise 1,861 regular-file blobs and three symlink blobs under the walking fixture; links are measured as their stored target-path bytes, never followed. This supersedes the preliminary working-tree count, which followed one link and omitted dangling/directory links. A line is a newline byte, like `wc -l`; files containing a NUL byte contribute bytes and files but no text lines. Compressed profiles and images therefore do not inflate the text ranking. Directory sizes are recursive; parent and child rows overlap and must not be added. Git directories, build outputs, caches, untracked scratch, and this new proposal are excluded.

`cohere` is a gitlink pinned to `715ba94f3608a6500086b1076ce5cb7e51b836db`, not an ordinary directory of Adamic-owned blobs. Its recursive dependency checkout is excluded from these totals. It supplies cohere and typescript-go; reorganizing its upstream tree is outside this unit.

Read first: all of `CLAUDE.md`, then `README.md`, `docs/0.1.md`, `docs/memory.md`, setup and root configuration. The complete tracked path inventory and the internal/stage1 package source headers, gap/report descriptions, function inventories and path consumers informed the purposes below. This is a layout review, not a whole-source correctness audit.

Reproduce the measurement without depending on a dirty working tree:

```python
import subprocess
from collections import defaultdict
base = "ef3d907ecdc4c771b016f7d9c52372def057a340"
sizes = defaultdict(lambda: [0, 0, 0])
files = []
for entry in subprocess.check_output(["git", "ls-tree", "-rz", base]).split(b"\0"):
    if not entry:
        continue
    metadata, raw_path = entry.split(b"\t")
    mode, kind, oid = metadata.split()
    if kind != b"blob":
        continue
    path = raw_path.decode()
    data = subprocess.check_output(["git", "cat-file", "blob", oid])
    lines = data.count(b"\n") if b"\0" not in data else 0
    files.append((path, len(data), lines, b"\0" in data))
    parts = path.split("/")
    for depth in range(1, len(parts)):
        total = sizes["/".join(parts[:depth])]
        total[0] += 1
        total[1] += lines
        total[2] += len(data)
print("total", len(files), sum(f[2] for f in files), sum(f[1] for f in files))
for directory, total in sorted(sizes.items()):
    print(directory, *total)
for file in sorted((f for f in files if not f[3]), key=lambda f: (-f[2], f[0]))[:20]:
    print(file)
```

## Current top-level layout

| Directory | Files | Text lines | Bytes | Purpose |
|---|---:|---:|---:|---|
| `.claude/` | 2 | 753 | 24,168 | Claude tooling and removal guard hooks. |
| `bench/` | 18 | 3,572 | 380,429 | Deterministic benchmark programs, runners and historical regex measurements. |
| `bridge/` | 30 | 2,660 | 108,303 | Opt-in external typescript-go C archive, ABI, checker facts and independent Go reference. |
| `cloud/` | 15 | 574 | 29,015 | Linux toolchain setup plus one historical JSON.stringify run bundle. |
| `cmd/` | 49 | 6,938 | 206,593 | Go command entry points and command-specific test corpora. |
| `dedication/` | 4 | 51 | 1,957 | Ken Thompson dedication, first native program and build recipe. |
| `docs/` | 93 | 25,198 | 2,340,020 | Language and memory contracts, feature designs, surveys and historical evidence. |
| `internal/` | 738 | 142,752 | 6,120,483 | Stage 0 compiler, analyses, runtimes and correctness harnesses. |
| `oracle/` | 2 | 197 | 7,498 | Independent Node source loader and Adamic host runtime, shared by tests and tools. |
| `review/` | 25 | 2,336 | 93,525 | One-off review probes, refusals, logs and review reports. |
| `stage1/` | 876 | 113,255 | 10,503,236 | Adamic ports of cohere components and TypeScript scanner/parser with comparison harnesses. |
| `cohere/` | 1 gitlink | excluded | excluded | Pinned upstream cohere/checker dependency; not an Adamic-owned source subtree. |

The root also has 12 regular files: `.gitattributes`, `.gitignore`, `.gitmodules`, `CLAUDE.md`, `CohereSettings.json`, the two license files, `README.md`, `THIRD_PARTY_NOTICES.md`, `go.mod`, `go.work`, and `tsconfig.json`. These are repository-wide contracts and stay at root.

## Every nested directory and package

The `internal/*` rows are the eleven Go packages. Stage 1 has seventeen component package roots; `testdata` helpers and generator commands are shown separately rather than mistaken for additional ports. Container rows have recursive sizes too.

| Directory | Files | Text lines | Bytes | Purpose |
|---|---:|---:|---:|---|
| `.claude/hooks/` | 1 | 736 | 23,862 | Container for hooks artifacts. |
| `bench/regex/` | 9 | 2,525 | 344,852 | Regexp benchmark cases, runner, counters and measured results. |
| `bench/unwinding/` | 1 | 270 | 6,782 | Standalone C exception-unwinding mechanism benchmark. |
| `bridge/tsgo/` | 30 | 2,660 | 108,303 | External checker C ABI and linked archive integration. |
| `bridge/tsgo/archive/` | 4 | 295 | 9,770 | C archive entry point, boundary ownership and profiling hooks. |
| `bridge/tsgo/checker/` | 7 | 1,008 | 33,576 | Go checker program handles, facts, scope and metadata queries. |
| `bridge/tsgo/cost/` | 1 | 75 | 2,011 | Checker query-cost command. |
| `bridge/tsgo/oracle/` | 1 | 95 | 3,020 | Independent direct-shim checker reference command. |
| `bridge/tsgo/profile/` | 5 | 233 | 10,468 | Reusable checker/native profiling and volume measurement scripts. |
| `bridge/tsgo/spec/` | 1 | 18 | 405 | Checker fact framing and ABI specification code. |
| `bridge/tsgo/testdata/` | 5 | 89 | 3,792 | Package-local fixture inputs and independent reference helpers. |
| `cloud/reports/` | 14 | 457 | 23,308 | Historical run evidence for reports; some bundles also contain measurement scripts. |
| `cloud/reports/json-stringify/` | 14 | 457 | 23,308 | Historical run evidence for json-stringify; some bundles also contain measurement scripts. |
| `cmd/adamic/` | 2 | 183 | 5,082 | Compiler CLI for types, C, native builds and JavaScript. |
| `cmd/adamic-fuzz/` | 1 | 171 | 5,681 | Differential fuzz command. |
| `cmd/adamic-meter/` | 11 | 1,441 | 48,380 | Source feature and adaptation census command. |
| `cmd/adamic-meter/testdata/` | 5 | 35 | 956 | Package-local fixture inputs and independent reference helpers. |
| `cmd/adamic-meter/testdata/adapt/` | 2 | 7 | 152 | Nested command or package fixture corpus (adapt). |
| `cmd/adamic-meter/testdata/corpus/` | 2 | 6 | 161 | Nested command or package fixture corpus (corpus). |
| `cmd/adamic-meter/testdata/optional/` | 1 | 22 | 643 | Nested command or package fixture corpus (optional). |
| `cmd/adamic-stage1-progress/` | 3 | 783 | 29,760 | Branch/main port coverage and progress reporting command. |
| `cmd/adamic-test262/` | 32 | 4,360 | 117,690 | Test262 classification, adaptation, verdicts and execution command. |
| `cmd/adamic-test262/testdata/` | 21 | 99 | 2,243 | Package-local fixture inputs and independent reference helpers. |
| `cmd/adamic-test262/testdata/corpus/` | 12 | 42 | 597 | Nested command or package fixture corpus (corpus). |
| `cmd/adamic-test262/testdata/mini/` | 3 | 17 | 366 | Nested command or package fixture corpus (mini). |
| `cmd/adamic-test262/testdata/mini/test/` | 3 | 17 | 366 | Nested command or package fixture corpus (test). |
| `cmd/adamic-test262/testdata/mini/test/pass/` | 1 | 6 | 180 | Nested command or package fixture corpus (pass). |
| `cmd/adamic-test262/testdata/mini/test/refuse/` | 1 | 5 | 115 | Language-contract programs deliberately rejected by stage 0. |
| `cmd/adamic-test262/testdata/mini/test/skip/` | 1 | 6 | 71 | Nested command or package fixture corpus (skip). |
| `cmd/adamic-test262/testdata/verdicts/` | 6 | 40 | 1,280 | Nested command or package fixture corpus (verdicts). |
| `docs/images/` | 1 | 0 | 818,544 | README artwork. |
| `docs/library-object/` | 24 | 2,002 | 74,398 | Object library implementation and mutant/run evidence. |
| `docs/library-object/mutants/` | 11 | 88 | 8,319 | Historical run evidence for mutants; some bundles also contain measurement scripts. |
| `docs/library-string/` | 8 | 2,426 | 97,317 | String library implementation and measurement evidence. |
| `docs/stage1-progress/` | 3 | 7,932 | 407,116 | Measured progress snapshots and pending/main listings. |
| `docs/tsc-strictness/` | 11 | 8,553 | 539,955 | Pinned source census, probe scripts and adaptation measurements. |
| `docs/utf16-views/` | 18 | 401 | 89,825 | UTF-16 cache measurements, sweep/mutant scripts and profiles. |
| `docs/verification/` | 16 | 458 | 30,958 | Historical verification evidence. |
| `docs/verification/regex-cycle-proof/` | 16 | 458 | 30,958 | Historical run evidence for regex-cycle-proof; some bundles also contain measurement scripts. |
| `internal/flow/` | 16 | 5,177 | 205,823 | IR control flow, SSA, liveness, aliasing and mutation ranges. |
| `internal/flow/testdata/` | 2 | 209 | 5,555 | Package-local fixture inputs and independent reference helpers. |
| `internal/fresh/` | 7 | 2,113 | 64,995 | Abstract heap and interprocedural proofs that writes cannot close strong cycles. |
| `internal/fresh/testdata/` | 1 | 31 | 1,144 | Package-local fixture inputs and independent reference helpers. |
| `internal/fuzz/` | 7 | 2,243 | 82,578 | Seeded valid-program generation, differential runs and shrinking. |
| `internal/ir/` | 6 | 1,297 | 42,311 | Typed tree IR and library-specific nodes consumed by both backends. |
| `internal/javascript/` | 3 | 1,004 | 42,454 | IR-to-JavaScript backend with inserted checks and explicit captures. |
| `internal/load/` | 25 | 1,161 | 42,263 | In-process checker host, .a mapping, embedded prelude and checker pin tests. |
| `internal/load/testdata/` | 16 | 286 | 7,829 | Package-local fixture inputs and independent reference helpers. |
| `internal/load/testdata/0.1/` | 16 | 286 | 7,829 | Nested command or package fixture corpus (0.1). |
| `internal/load/testdata/0.1/compile/` | 11 | 242 | 6,783 | Approved language examples required to compile. |
| `internal/load/testdata/0.1/compile/07_modules/` | 2 | 30 | 814 | Approved language examples required to compile. |
| `internal/load/testdata/0.1/refuse/` | 5 | 44 | 1,046 | Language-contract programs deliberately rejected by stage 0. |
| `internal/lower/` | 64 | 16,840 | 720,140 | Checked AST to typed IR, diagnostics, subset refusals and ownership facts. |
| `internal/native/` | 186 | 35,932 | 1,741,220 | C emission, ownership/reuse/region plans, runtime embedding and clang driver. |
| `internal/native/performance/` | 56 | 1,144 | 553,367 | Historical run evidence for performance; some bundles also contain measurement scripts. |
| `internal/native/performance/field-access/` | 32 | 640 | 297,646 | Historical run evidence for field-access; some bundles also contain measurement scripts. |
| `internal/native/performance/utf16-cache/` | 24 | 504 | 255,721 | Historical run evidence for utf16-cache; some bundles also contain measurement scripts. |
| `internal/native/runtime/` | 66 | 24,561 | 800,046 | Embedded C11 runtime, private string implementation headers and generated Unicode tables. |
| `internal/native/testdata/` | 2 | 83 | 7,856 | Package-local fixture inputs and independent reference helpers. |
| `internal/oracle/` | 386 | 24,297 | 697,085 | Three-way compiler harness, sanitizer/leak checks and recorded allocation counts. |
| `internal/oracle/testdata/` | 362 | 20,104 | 514,082 | Package-local fixture inputs and independent reference helpers. |
| `internal/oracle/testdata/fresh_refused/` | 42 | 662 | 21,073 | Cycle-closing programs required to remain refused. |
| `internal/oracle/testdata/modules/` | 3 | 22 | 459 | Multi-file module order and import fixtures. |
| `internal/oracle/testdata/reading/` | 7 | 7,703 | 38,584 | File input and malformed UTF-8 fixture data. |
| `internal/oracle/testdata/sweeps/` | 1 | 56 | 2,343 | Generated semantic sweeps against external behavior. |
| `internal/oracle/testdata/walking/` | 11 | 3 | 51 | Directory traversal fixture hierarchy. |
| `internal/oracle/testdata/walking/inner/` | 1 | 1 | 10 | Directory traversal fixture hierarchy. |
| `internal/oracle/testdata/weak/` | 4 | 96 | 2,838 | Weak-handle ownership and lifetime probes. |
| `internal/regexp/` | 30 | 7,897 | 1,507,001 | ECMAScript regexp syntax, Go matcher, native serialization and provider adapters. |
| `internal/regexp/testdata/` | 12 | 948 | 1,317,080 | Package-local fixture inputs and independent reference helpers. |
| `internal/unicodeproperties/` | 8 | 44,791 | 974,613 | Generated Unicode property/alias/canonicalization data and Node agreement tests. |
| `review/class-features-coverage/` | 1 | 41 | 1,371 | One-off review task artifacts for class-features-coverage. |
| `review/fnexpr-coverage/` | 1 | 59 | 2,187 | One-off review task artifacts for fnexpr-coverage. |
| `review/generator-oct6/` | 1 | 15 | 495 | One-off review task artifacts for generator-oct6. |
| `review/library-map-set/` | 18 | 2,016 | 74,939 | One-off review task artifacts for library-map-set. |
| `review/object-coverage/` | 1 | 14 | 556 | One-off review task artifacts for object-coverage. |
| `review/oct6/` | 2 | 128 | 9,726 | One-off review task artifacts for oct6. |
| `review/regex-cycle-coverage/` | 1 | 63 | 4,251 | One-off review task artifacts for regex-cycle-coverage. |
| `stage1/cohere/` | 778 | 100,663 | 9,771,145 | Cohere port slices, each paired with upstream reference tests. |
| `stage1/cohere/css/` | 123 | 10,228 | 1,395,329 | CSS AST parsing, value composition and document printing. |
| `stage1/cohere/css/gaps/` | 9 | 57 | 2,741 | Deliberate unsupported or refused minimal probes; preserve their exact inputs. |
| `stage1/cohere/css/performance/` | 20 | 1,161 | 899,445 | Historical run evidence for performance; some bundles also contain measurement scripts. |
| `stage1/cohere/css/testdata/` | 9 | 1,775 | 68,691 | Package-local fixture inputs and independent reference helpers. |
| `stage1/cohere/css/verification/` | 52 | 1,414 | 140,688 | Historical run evidence for verification; some bundles also contain measurement scripts. |
| `stage1/cohere/cssnumbers/` | 9 | 1,040 | 46,007 | CSS numeric conversion port. |
| `stage1/cohere/cssnumbers/testdata/` | 3 | 44 | 1,734 | Package-local fixture inputs and independent reference helpers. |
| `stage1/cohere/cssstrings/` | 11 | 867 | 37,939 | CSS string conversion port. |
| `stage1/cohere/cssstrings/gaps/` | 1 | 3 | 79 | Deliberate unsupported or refused minimal probes; preserve their exact inputs. |
| `stage1/cohere/cssstrings/testdata/` | 3 | 44 | 1,723 | Package-local fixture inputs and independent reference helpers. |
| `stage1/cohere/formatfiles/` | 7 | 1,803 | 77,280 | File enumeration, nested repository boundaries and formatter discovery. |
| `stage1/cohere/formatfiles/testdata/` | 1 | 455 | 16,492 | Package-local fixture inputs and independent reference helpers. |
| `stage1/cohere/gitignore/` | 22 | 3,567 | 147,552 | Git ignore/glob matching, path semantics and case handling. |
| `stage1/cohere/gitignore/gaps/` | 10 | 66 | 2,129 | Deliberate unsupported or refused minimal probes; preserve their exact inputs. |
| `stage1/cohere/gitignore/testdata/` | 1 | 623 | 26,729 | Package-local fixture inputs and independent reference helpers. |
| `stage1/cohere/graphql/` | 14 | 3,995 | 156,258 | GraphQL lexer/parser and block string handling. |
| `stage1/cohere/graphql/gaps/` | 2 | 29 | 888 | Deliberate unsupported or refused minimal probes; preserve their exact inputs. |
| `stage1/cohere/graphql/testdata/` | 2 | 859 | 34,208 | Package-local fixture inputs and independent reference helpers. |
| `stage1/cohere/json/` | 47 | 6,688 | 399,136 | JSON parser, formatter and document width tables. |
| `stage1/cohere/json/gaps/` | 4 | 13 | 363 | Deliberate unsupported or refused minimal probes; preserve their exact inputs. |
| `stage1/cohere/json/performance/` | 22 | 617 | 188,729 | Historical run evidence for performance; some bundles also contain measurement scripts. |
| `stage1/cohere/json/testdata/` | 5 | 219 | 6,412 | Package-local fixture inputs and independent reference helpers. |
| `stage1/cohere/lint/` | 114 | 11,731 | 2,169,565 | Syntax lint runner, rules, settings, messages and volume driver. |
| `stage1/cohere/lint/gaps/` | 5 | 42 | 1,212 | Deliberate unsupported or refused minimal probes; preserve their exact inputs. |
| `stage1/cohere/lint/performance/` | 63 | 2,163 | 1,660,458 | Historical run evidence for performance; some bundles also contain measurement scripts. |
| `stage1/cohere/lint/testdata/` | 5 | 482 | 18,535 | Package-local fixture inputs and independent reference helpers. |
| `stage1/cohere/lint/validation/` | 8 | 592 | 60,780 | Historical run evidence for validation; some bundles also contain measurement scripts. |
| `stage1/cohere/lint/volume_evidence/` | 13 | 1,976 | 161,868 | Historical run evidence for volume_evidence; some bundles also contain measurement scripts. |
| `stage1/cohere/markdownblocks/` | 100 | 16,585 | 680,593 | Markdown block parser, AST representation, decoding, width and entity tables. |
| `stage1/cohere/markdownblocks/gaps/` | 9 | 77 | 2,306 | Deliberate unsupported or refused minimal probes; preserve their exact inputs. |
| `stage1/cohere/markdownblocks/testdata/` | 35 | 2,206 | 81,983 | Package-local fixture inputs and independent reference helpers. |
| `stage1/cohere/markdownblocks/tools/` | 3 | 413 | 11,888 | Generators for Markdown class, entity and width data. |
| `stage1/cohere/markdownblocks/tools/generate_classes/` | 1 | 88 | 2,314 | Generated table tool: generate classes. |
| `stage1/cohere/markdownblocks/tools/generate_entities/` | 1 | 118 | 3,290 | Generated table tool: generate entities. |
| `stage1/cohere/markdownblocks/tools/generate_width/` | 1 | 207 | 6,284 | Generated table tool: generate width. |
| `stage1/cohere/markdowninline/` | 12 | 1,287 | 60,515 | Markdown inline parser and character classes. |
| `stage1/cohere/markdowninline/gaps/` | 1 | 2 | 142 | Deliberate unsupported or refused minimal probes; preserve their exact inputs. |
| `stage1/cohere/markdowninline/testdata/` | 3 | 162 | 6,545 | Package-local fixture inputs and independent reference helpers. |
| `stage1/cohere/mediaquery/` | 17 | 1,957 | 84,676 | CSS media-query parsing. |
| `stage1/cohere/mediaquery/gaps/` | 6 | 52 | 2,058 | Deliberate unsupported or refused minimal probes; preserve their exact inputs. |
| `stage1/cohere/mediaquery/testdata/` | 2 | 302 | 11,466 | Package-local fixture inputs and independent reference helpers. |
| `stage1/cohere/selector/` | 19 | 2,049 | 79,183 | Lossless CSS selector AST parser. |
| `stage1/cohere/selector/gaps/` | 6 | 36 | 992 | Deliberate unsupported or refused minimal probes; preserve their exact inputs. |
| `stage1/cohere/selector/testdata/` | 3 | 377 | 12,939 | Package-local fixture inputs and independent reference helpers. |
| `stage1/cohere/suppression/` | 12 | 2,134 | 87,655 | Lint directive scan, parsing and suppression index. |
| `stage1/cohere/suppression/gaps/` | 2 | 21 | 991 | Deliberate unsupported or refused minimal probes; preserve their exact inputs. |
| `stage1/cohere/suppression/testdata/` | 2 | 402 | 16,209 | Package-local fixture inputs and independent reference helpers. |
| `stage1/cohere/typeaware/` | 255 | 34,005 | 4,250,989 | Native type-aware lint rules over external checker facts and shared parser walk. |
| `stage1/cohere/typeaware/testdata/` | 10 | 715 | 24,861 | Package-local fixture inputs and independent reference helpers. |
| `stage1/cohere/typeaware/validation/` | 12 | 1,261 | 95,770 | Historical run evidence for validation; some bundles also contain measurement scripts. |
| `stage1/cohere/typeaware/validation-profile/` | 57 | 15,796 | 2,240,200 | Historical run evidence for validation-profile; some bundles also contain measurement scripts. |
| `stage1/cohere/typeaware/validation-six/` | 28 | 6,028 | 798,750 | Historical run evidence for validation-six; some bundles also contain measurement scripts. |
| `stage1/cohere/typeaware/validation-volume/` | 62 | 2,241 | 548,492 | Historical run evidence for validation-volume; some bundles also contain measurement scripts. |
| `stage1/cohere/typeaware/validation-volume-profile/` | 49 | 1,290 | 229,991 | Historical run evidence for validation-volume-profile; some bundles also contain measurement scripts. |
| `stage1/cohere/values/` | 16 | 2,727 | 98,468 | CSS value tokenizer/parser, used by CSS composition. |
| `stage1/cohere/values/gaps/` | 5 | 40 | 1,272 | Deliberate unsupported or refused minimal probes; preserve their exact inputs. |
| `stage1/cohere/values/testdata/` | 2 | 353 | 12,146 | Package-local fixture inputs and independent reference helpers. |
| `stage1/typescript/` | 98 | 12,592 | 732,091 | TypeScript scanner and parser ports, not another checker implementation. |
| `stage1/typescript/parser/` | 48 | 5,722 | 253,636 | TypeScript expressions, statements, declarations and type nodes, over shared scanner. |
| `stage1/typescript/parser/gaps/` | 7 | 35 | 1,113 | Deliberate unsupported or refused minimal probes; preserve their exact inputs. |
| `stage1/typescript/parser/testdata/` | 1 | 242 | 7,241 | Package-local fixture inputs and independent reference helpers. |
| `stage1/typescript/parser/validation/` | 21 | 615 | 37,515 | Historical run evidence for validation; some bundles also contain measurement scripts. |
| `stage1/typescript/scanner/` | 50 | 6,870 | 478,455 | TypeScript token scanner and rescan modes. |
| `stage1/typescript/scanner/gaps/` | 2 | 7 | 291 | Deliberate unsupported or refused minimal probes; preserve their exact inputs. |
| `stage1/typescript/scanner/performance/` | 34 | 3,718 | 327,923 | Historical run evidence for performance; some bundles also contain measurement scripts. |
| `stage1/typescript/scanner/testdata/` | 1 | 119 | 3,251 | Package-local fixture inputs and independent reference helpers. |

## Twenty biggest text files

Ranked by LF lines, then path for ties. Byte size is also shown so a profile log is not confused with executable code.

| Rank | Path | Lines | Bytes | Kind |
|---:|---|---:|---:|---|
| 1 | `internal/unicodeproperties/tables.go` | 33,646 | 738,103 | Generated Unicode data |
| 2 | `internal/native/runtime/normalize_tables.h` | 8,032 | 264,861 | Generated Unicode data |
| 3 | `internal/unicodeproperties/canonicalize_tables.go` | 7,972 | 127,929 | Generated Unicode data |
| 4 | `internal/oracle/testdata/reading/malformed.txt` | 7,698 | 38,474 | Oracle fixture input |
| 5 | `docs/stage1-progress/snapshots.json` | 6,711 | 249,783 | Measured snapshot/census data |
| 6 | `stage1/cohere/typeaware/validation-profile/phases-before-raw.log` | 4,279 | 532,205 | Historical run output |
| 7 | `stage1/cohere/typeaware/validation-profile/baseline-raw.log` | 4,244 | 496,225 | Historical run output |
| 8 | `internal/regexp/canonicalize.go` | 3,019 | 56,756 | Generated Unicode data |
| 9 | `docs/tsc-strictness/sample.json` | 3,002 | 229,301 | Measured snapshot/census data |
| 10 | `internal/native/runtime/ieee754.c` | 2,733 | 93,682 | Attributed numerical algorithm port |
| 11 | `internal/native/runtime/regexp_fold.h` | 2,690 | 44,293 | Generated Unicode data |
| 12 | `stage1/typescript/scanner/performance/measurements.json` | 2,670 | 78,468 | Measured snapshot/census data |
| 13 | `docs/tsc-strictness/syntax-census.json` | 2,158 | 63,490 | Measured snapshot/census data |
| 14 | `cmd/adamic-test262/adapt.go` | 2,089 | 45,749 | Handwritten source |
| 15 | `stage1/cohere/typeaware/validation-profile/phases-after-raw.log` | 2,007 | 271,177 | Historical run output |
| 16 | `stage1/cohere/markdownblocks/REPORT.txt` | 1,957 | 134,502 | Historical report |
| 17 | `internal/lower/object.go` | 1,944 | 81,237 | Handwritten source |
| 18 | `internal/fresh/fresh.go` | 1,868 | 55,971 | Handwritten source |
| 19 | `stage1/cohere/typeaware/validation-profile/compiler-go.stdout.log` | 1,841 | 288,482 | Historical run output |
| 20 | `stage1/cohere/typeaware/validation-profile/compiler-native.stdout.log` | 1,841 | 288,482 | Historical run output |

## Observations and proposed decisions

1. **Run evidence has several homes.** There are 365 tracked `.log` files totaling 31,775 LF lines. They appear under `cloud/reports`, `docs`, `internal/native/performance`, stage-1 validation/performance/verification directories and `review`. `.gitignore` already says logs should never reach a commit; its `*.log` rule cannot untrack historical files. The proposal is to export raw run bundles to the owning task, retaining bytes, filenames, base/source hashes, versions, commands and mutant outcomes. Durable interpretation belongs under `docs/reports/`. No evidence is discarded before its task attachment and checksum manifest are accessible to Kirk and Ahra.

2. **Review scratch is not a permanent test suite.** `review` has 25 files: dated review narrative, refusal transcripts, a class probe and map/set run output. No live `review/` pathname consumer was found in Go, Python, shell or Adamic source; remaining mentions there are historical comments. Archive these exact files to their review tasks. If `review/oct6/class_oct6_construction.a` exposes an unrepresented regression, first land a separate fixture-promotion unit under `internal/oracle/testdata/`, with a count row, independent Node result and a caught mutant. Do not silently drop the probe. Keep the `review/**` cohere exclusion until the last reviewer branch has landed.

3. **Two oracle names have different jobs.** Root `oracle/` executes source independently on Node; `internal/oracle/` is the Go comparison harness. `bridge/tsgo/oracle/` is an independent direct-shim checker reference. These are not duplicate implementations. Many stage-1 tests, fuzzing and profiling tools refer to `oracle/node.mjs`; renaming it would create cross-area conflicts with no semantic benefit. Keep all three and document the distinction.

4. **Unicode canonicalization really has two Go homes.** `internal/regexp/canonicalize.go` carries generated `simpleCaseFold`/`legacyUppercase` data; `internal/unicodeproperties/canonicalize.go` and `canonicalize_tables.go` expose the same conceptual Unicode/legacy operations. Both advertise Unicode 17. Regex already imports the property provider. This is a candidate for one generator/data owner, but equivalence is not established by filenames or matching versions. A later semantic unit should compare every relevant code point and flag mode against Node, including Kelvin sign, long s and surrogate cases, before delegating regexp folding to the property package. There is no file-move-only deduplication proposed here. C folding tables remain a distinct runtime representation.

5. **Repeated width data is another candidate, not a proven copy.** JSON `widthTables.ts` is 1,773 lines; Markdown `widthTables.ts` is 1,751 lines, and CSS also has width data. Markdown has a pinned generator under `tools/generate_width`. Entity tables and character classes also specialize different parsers. Keep tables local while callers and generators differ. A shared `stage1/cohere/doc/` module would be a functional extraction requiring dependency and behavior checks, not a mechanical organization batch.

6. **Byte-identical evidence exists.** SHA-256 grouping of blobs over 1,000 bytes found `internal/native/performance/field-access/json-before-top.txt`, `internal/native/performance/utf16-cache/json-before-top.txt`, and `stage1/cohere/json/performance/final-top.txt` identical. Typeaware profile and six-rule compiler manifests are identical; their Go/native compiler stdout logs form a four-file identical group. Preserve their provenance when exporting rather than pretending four copies are four independent observations. Identical license files remain beside their attributed ports; legal notices are intentional duplication.

7. **Naming drift is visible.** `docs/library_function_expressions_for_in.md` uses underscores while neighboring feature docs use hyphens. Stage 1 mixes camel-case filenames (`blockString.ts`, `identifierTables.ts`, `astArena.ts`) and underscores (`parse_value.ts`, `type_fact.ts`); `REPORT.md`, `REPORT.txt`, `WHOLE_REPORT.md`, and `VOLUME_PROFILE_REPORT.md` describe task reports inconsistently. Normalize report paths mechanically in the table below. Do not rename all imported source files with 150 branches in flight. Use names spelled out in full for new files; all new Adamic source is `.a`, per Kirk’s October 6 direction, even though the older language document allows both extensions. Legacy `.ts`, fixture imports, gap paths and `.d.ts` declarations stay until their owning slice receives an explicit extension migration unit.

8. **Some giant files mix areas; others are data.** `object.go` mixes object literals, array/map/string calls, iteration and switch lowering. `fresh.go` combines the abstract heap, graph interpreter and summary replay. `adapt.go` combines parsing and rewrites. Whole-function extractions below can reduce these without changing package names. `tables.go`, normalization/folding headers and `canonicalize.go` are generator outputs, not hand-edited modules; splitting them for size alone adds churn. `ieee754.c` is an attributed algorithm port and should retain its provenance. `malformed.txt` is a semantic input, not a log. `snapshots.json` is deliberate progress data, not compiler source.

## Proposed layout

```text
root contracts and configurations        unchanged
bench/                                  reusable benchmark programs and measurement tools
bridge/tsgo/                            archive, checker, spec, reference and profile tools
cloud/setup.sh                          machine setup only
cmd/                                    existing command paths
dedication/                             first program and build recipe
docs/                                   contracts, designs and curated survey/progress inputs
  reports/<original area>/              durable historical interpretations, named by topic
internal/                               existing eleven packages, fixture paths and runtime layout
oracle/                                 independent Node source runner
stage1/cohere/<slice>/                   source, tests, gaps, licenses and reusable tools
stage1/typescript/{scanner,parser}/      source, tests, gaps, licenses and reusable tools
documentation-proposal/organization.md   this reviewable proposal
owning task attachments                 raw logs/profiles/transcripts, outside the repository
```

No generic `src/`, `tests/`, `archive/` or stage-0 wrapper is added. The existing package seams describe the compiler well. The new report home groups prose by its original area, avoiding one enormous report directory.

### Exact whole-file move ledger

Each row is one proposed move, not a glob. `task:organization-archive/<old path>` is an exact logical attachment name outside Git, not a filesystem directory or an existing service URL. Ahra maps each bundle to its owning task before implementation; task IDs and upload availability were not measured here. Preserve that attachment name and publish its actual URL plus SHA-256 in the durable report. A bundle with no available task destination stays in Git until that prerequisite is met. Reusable scripts and `.gitattributes` remain at their current paths; scripts receive explicit archive input paths, with their output bytes checked in a separate follow-up if their defaults depend on adjacency. Every executable relocation is deferred.

Reports move with byte-identical bodies first; adjust only relative links and command paths necessary to keep the report usable. Do not rewrite the historical commands as if they ran at their new location. Reports that use original task/commit-relative paths should label them as historical. Existing links in `GAPS.md`, READMEs and other reports need matching updates in the same area batch.

| Old path | New path | Batch |
|---|---|---|
| `bench/regex/RESULTS.md` | `docs/reports/bench/regex/results.md` | report: bench/regex |
| `bridge/tsgo/REPORT.md` | `docs/reports/bridge/tsgo/report.md` | report: bridge/tsgo |
| `cloud/reports/json-stringify/after.json` | `task:organization-archive/cloud/reports/json-stringify/after.json` | cloud evidence |
| `cloud/reports/json-stringify/before.json` | `task:organization-archive/cloud/reports/json-stringify/before.json` | cloud evidence |
| `cloud/reports/json-stringify/counts-refresh.log` | `task:organization-archive/cloud/reports/json-stringify/counts-refresh.log` | raw evidence: cloud |
| `cloud/reports/json-stringify/counts.log` | `task:organization-archive/cloud/reports/json-stringify/counts.log` | raw evidence: cloud |
| `cloud/reports/json-stringify/fresh-final.log` | `task:organization-archive/cloud/reports/json-stringify/fresh-final.log` | raw evidence: cloud |
| `cloud/reports/json-stringify/full-gate-incomplete.log` | `task:organization-archive/cloud/reports/json-stringify/full-gate-incomplete.log` | raw evidence: cloud |
| `cloud/reports/json-stringify/mutants.log` | `task:organization-archive/cloud/reports/json-stringify/mutants.log` | raw evidence: cloud |
| `cloud/reports/json-stringify/native-final.log` | `task:organization-archive/cloud/reports/json-stringify/native-final.log` | raw evidence: cloud |
| `cloud/reports/json-stringify/oracle-final.log` | `task:organization-archive/cloud/reports/json-stringify/oracle-final.log` | raw evidence: cloud |
| `cloud/reports/json-stringify/oracle.log` | `task:organization-archive/cloud/reports/json-stringify/oracle.log` | raw evidence: cloud |
| `cloud/reports/json-stringify/packages-final.log` | `task:organization-archive/cloud/reports/json-stringify/packages-final.log` | raw evidence: cloud |
| `cloud/reports/json-stringify/refusals-and-key-order-mutant.log` | `task:organization-archive/cloud/reports/json-stringify/refusals-and-key-order-mutant.log` | raw evidence: cloud |
| `cloud/reports/json-stringify/report.txt` | `docs/reports/cloud/json-stringify/report.txt` | cloud report |
| `cloud/reports/json-stringify/setup.log` | `task:organization-archive/cloud/reports/json-stringify/setup.log` | raw evidence: cloud |
| `docs/library-object/assign-mutant-final.log` | `task:organization-archive/docs/library-object/assign-mutant-final.log` | raw evidence: docs |
| `docs/library-object/counts.log` | `task:organization-archive/docs/library-object/counts.log` | raw evidence: docs |
| `docs/library-object/format.log` | `task:organization-archive/docs/library-object/format.log` | raw evidence: docs |
| `docs/library-object/gate.log` | `task:organization-archive/docs/library-object/gate.log` | raw evidence: docs |
| `docs/library-object/mutants.log` | `task:organization-archive/docs/library-object/mutants.log` | raw evidence: docs |
| `docs/library-object/mutants/assign-a-plus-one.log` | `task:organization-archive/docs/library-object/mutants/assign-a-plus-one.log` | raw evidence: docs |
| `docs/library-object/mutants/entries-wrong-key.log` | `task:organization-archive/docs/library-object/mutants/entries-wrong-key.log` | raw evidence: docs |
| `docs/library-object/mutants/freeze-no-flag.log` | `task:organization-archive/docs/library-object/mutants/freeze-no-flag.log` | raw evidence: docs |
| `docs/library-object/mutants/freeze-write-ignored.log` | `task:organization-archive/docs/library-object/mutants/freeze-write-ignored.log` | raw evidence: docs |
| `docs/library-object/mutants/has-own-always-false.log` | `task:organization-archive/docs/library-object/mutants/has-own-always-false.log` | raw evidence: docs |
| `docs/library-object/mutants/keys-descending.log` | `task:organization-archive/docs/library-object/mutants/keys-descending.log` | raw evidence: docs |
| `docs/library-object/mutants/keys-integer-last.log` | `task:organization-archive/docs/library-object/mutants/keys-integer-last.log` | raw evidence: docs |
| `docs/library-object/mutants/same-value-nan-false.log` | `task:organization-archive/docs/library-object/mutants/same-value-nan-false.log` | raw evidence: docs |
| `docs/library-object/mutants/same-value-zero.log` | `task:organization-archive/docs/library-object/mutants/same-value-zero.log` | raw evidence: docs |
| `docs/library-object/mutants/spread-reuses-frozen.log` | `task:organization-archive/docs/library-object/mutants/spread-reuses-frozen.log` | raw evidence: docs |
| `docs/library-object/mutants/values-first-slot.log` | `task:organization-archive/docs/library-object/mutants/values-first-slot.log` | raw evidence: docs |
| `docs/library-object/oracle.log` | `task:organization-archive/docs/library-object/oracle.log` | raw evidence: docs |
| `docs/library-object/refusals.log` | `task:organization-archive/docs/library-object/refusals.log` | raw evidence: docs |
| `docs/library-object/setup.log` | `task:organization-archive/docs/library-object/setup.log` | raw evidence: docs |
| `docs/library-object/vet.log` | `task:organization-archive/docs/library-object/vet.log` | raw evidence: docs |
| `docs/library-string/gate.log` | `task:organization-archive/docs/library-string/gate.log` | raw evidence: docs |
| `docs/library-string/mutants.log` | `task:organization-archive/docs/library-string/mutants.log` | raw evidence: docs |
| `docs/library-string/packages.log` | `task:organization-archive/docs/library-string/packages.log` | raw evidence: docs |
| `docs/library-string/setup.log` | `task:organization-archive/docs/library-string/setup.log` | raw evidence: docs |
| `docs/library_function_expressions_for_in.md` | `docs/library-function-expressions-for-in.md` | documentation naming |
| `docs/utf16-views/counts.log` | `task:organization-archive/docs/utf16-views/counts.log` | raw evidence: docs |
| `docs/utf16-views/gate.log` | `task:organization-archive/docs/utf16-views/gate.log` | raw evidence: docs |
| `docs/utf16-views/gofmt.log` | `task:organization-archive/docs/utf16-views/gofmt.log` | raw evidence: docs |
| `docs/utf16-views/mutant-accounting.log` | `task:organization-archive/docs/utf16-views/mutant-accounting.log` | raw evidence: docs |
| `docs/utf16-views/mutant-append.log` | `task:organization-archive/docs/utf16-views/mutant-append.log` | raw evidence: docs |
| `docs/utf16-views/mutant-backward.log` | `task:organization-archive/docs/utf16-views/mutant-backward.log` | raw evidence: docs |
| `docs/utf16-views/mutant-supplementary.log` | `task:organization-archive/docs/utf16-views/mutant-supplementary.log` | raw evidence: docs |
| `docs/utf16-views/native.log` | `task:organization-archive/docs/utf16-views/native.log` | raw evidence: docs |
| `docs/utf16-views/oracle.log` | `task:organization-archive/docs/utf16-views/oracle.log` | raw evidence: docs |
| `docs/utf16-views/scanner.log` | `task:organization-archive/docs/utf16-views/scanner.log` | raw evidence: docs |
| `docs/utf16-views/setup.log` | `task:organization-archive/docs/utf16-views/setup.log` | raw evidence: docs |
| `docs/utf16-views/sweep.log` | `task:organization-archive/docs/utf16-views/sweep.log` | raw evidence: docs |
| `docs/utf16-views/vet.log` | `task:organization-archive/docs/utf16-views/vet.log` | raw evidence: docs |
| `docs/verification/regex-cycle-proof/closed-gap.log` | `task:organization-archive/docs/verification/regex-cycle-proof/closed-gap.log` | raw evidence: docs |
| `docs/verification/regex-cycle-proof/css-composition.log` | `task:organization-archive/docs/verification/regex-cycle-proof/css-composition.log` | raw evidence: docs |
| `docs/verification/regex-cycle-proof/css-final.log` | `task:organization-archive/docs/verification/regex-cycle-proof/css-final.log` | raw evidence: docs |
| `docs/verification/regex-cycle-proof/final-proof-oracle.log` | `task:organization-archive/docs/verification/regex-cycle-proof/final-proof-oracle.log` | raw evidence: docs |
| `docs/verification/regex-cycle-proof/fixture-counts.log` | `task:organization-archive/docs/verification/regex-cycle-proof/fixture-counts.log` | raw evidence: docs |
| `docs/verification/regex-cycle-proof/fixture.log` | `task:organization-archive/docs/verification/regex-cycle-proof/fixture.log` | raw evidence: docs |
| `docs/verification/regex-cycle-proof/future-mutant.log` | `task:organization-archive/docs/verification/regex-cycle-proof/future-mutant.log` | raw evidence: docs |
| `docs/verification/regex-cycle-proof/gate.log` | `task:organization-archive/docs/verification/regex-cycle-proof/gate.log` | raw evidence: docs |
| `docs/verification/regex-cycle-proof/main-dependency.log` | `task:organization-archive/docs/verification/regex-cycle-proof/main-dependency.log` | raw evidence: docs |
| `docs/verification/regex-cycle-proof/operand-mutant.log` | `task:organization-archive/docs/verification/regex-cycle-proof/operand-mutant.log` | raw evidence: docs |
| `docs/verification/regex-cycle-proof/proof.log` | `task:organization-archive/docs/verification/regex-cycle-proof/proof.log` | raw evidence: docs |
| `docs/verification/regex-cycle-proof/scratch-counts-failure.log` | `task:organization-archive/docs/verification/regex-cycle-proof/scratch-counts-failure.log` | raw evidence: docs |
| `docs/verification/regex-cycle-proof/setup.log` | `task:organization-archive/docs/verification/regex-cycle-proof/setup.log` | raw evidence: docs |
| `docs/verification/regex-cycle-proof/unsafe-mutant.log` | `task:organization-archive/docs/verification/regex-cycle-proof/unsafe-mutant.log` | raw evidence: docs |
| `docs/verification/regex-cycle-proof/vet.log` | `task:organization-archive/docs/verification/regex-cycle-proof/vet.log` | raw evidence: docs |
| `internal/native/performance/field-access/REPORT.md` | `docs/reports/internal/native/performance/field-access/report.md` | report: internal/native/performance/field-access |
| `internal/native/performance/field-access/conflict-mutant.log.gz` | `task:organization-archive/internal/native/performance/field-access/conflict-mutant.log.gz` | raw evidence: internal/native |
| `internal/native/performance/field-access/diff-check.log.gz` | `task:organization-archive/internal/native/performance/field-access/diff-check.log.gz` | raw evidence: internal/native |
| `internal/native/performance/field-access/disabled-mutant.log.gz` | `task:organization-archive/internal/native/performance/field-access/disabled-mutant.log.gz` | raw evidence: internal/native |
| `internal/native/performance/field-access/fields-test.log.gz` | `task:organization-archive/internal/native/performance/field-access/fields-test.log.gz` | raw evidence: internal/native |
| `internal/native/performance/field-access/filtered-oracle.log.gz` | `task:organization-archive/internal/native/performance/field-access/filtered-oracle.log.gz` | raw evidence: internal/native |
| `internal/native/performance/field-access/full-gate.log.gz` | `task:organization-archive/internal/native/performance/field-access/full-gate.log.gz` | raw evidence: internal/native |
| `internal/native/performance/field-access/generated-c-sha256.txt` | `task:organization-archive/internal/native/performance/field-access/generated-c-sha256.txt` | raw evidence: internal/native |
| `internal/native/performance/field-access/gofmt.log.gz` | `task:organization-archive/internal/native/performance/field-access/gofmt.log.gz` | raw evidence: internal/native |
| `internal/native/performance/field-access/json-after-compare.log.gz` | `task:organization-archive/internal/native/performance/field-access/json-after-compare.log.gz` | raw evidence: internal/native |
| `internal/native/performance/field-access/json-after-top.txt` | `task:organization-archive/internal/native/performance/field-access/json-after-top.txt` | raw evidence: internal/native |
| `internal/native/performance/field-access/json-after.callgrind.gz` | `task:organization-archive/internal/native/performance/field-access/json-after.callgrind.gz` | raw evidence: internal/native |
| `internal/native/performance/field-access/json-after.stderr` | `task:organization-archive/internal/native/performance/field-access/json-after.stderr` | raw evidence: internal/native |
| `internal/native/performance/field-access/json-before-compare.log.gz` | `task:organization-archive/internal/native/performance/field-access/json-before-compare.log.gz` | raw evidence: internal/native |
| `internal/native/performance/field-access/json-before-top.txt` | `task:organization-archive/internal/native/performance/field-access/json-before-top.txt` | raw evidence: internal/native |
| `internal/native/performance/field-access/json-before.callgrind.gz` | `task:organization-archive/internal/native/performance/field-access/json-before.callgrind.gz` | raw evidence: internal/native |
| `internal/native/performance/field-access/json-before.stderr` | `task:organization-archive/internal/native/performance/field-access/json-before.stderr` | raw evidence: internal/native |
| `internal/native/performance/field-access/json-parity.log.gz` | `task:organization-archive/internal/native/performance/field-access/json-parity.log.gz` | raw evidence: internal/native |
| `internal/native/performance/field-access/measurements.txt` | `task:organization-archive/internal/native/performance/field-access/measurements.txt` | raw evidence: internal/native |
| `internal/native/performance/field-access/missing-runtime-mutant.log.gz` | `task:organization-archive/internal/native/performance/field-access/missing-runtime-mutant.log.gz` | raw evidence: internal/native |
| `internal/native/performance/field-access/scanner-after-top.txt` | `task:organization-archive/internal/native/performance/field-access/scanner-after-top.txt` | raw evidence: internal/native |
| `internal/native/performance/field-access/scanner-after.callgrind.gz` | `task:organization-archive/internal/native/performance/field-access/scanner-after.callgrind.gz` | raw evidence: internal/native |
| `internal/native/performance/field-access/scanner-after.stderr` | `task:organization-archive/internal/native/performance/field-access/scanner-after.stderr` | raw evidence: internal/native |
| `internal/native/performance/field-access/scanner-before-top.txt` | `task:organization-archive/internal/native/performance/field-access/scanner-before-top.txt` | raw evidence: internal/native |
| `internal/native/performance/field-access/scanner-before.callgrind.gz` | `task:organization-archive/internal/native/performance/field-access/scanner-before.callgrind.gz` | raw evidence: internal/native |
| `internal/native/performance/field-access/scanner-before.stderr` | `task:organization-archive/internal/native/performance/field-access/scanner-before.stderr` | raw evidence: internal/native |
| `internal/native/performance/field-access/scanner-count-compare.log.gz` | `task:organization-archive/internal/native/performance/field-access/scanner-count-compare.log.gz` | raw evidence: internal/native |
| `internal/native/performance/field-access/scanner-parity.log.gz` | `task:organization-archive/internal/native/performance/field-access/scanner-parity.log.gz` | raw evidence: internal/native |
| `internal/native/performance/field-access/setup.log.gz` | `task:organization-archive/internal/native/performance/field-access/setup.log.gz` | raw evidence: internal/native |
| `internal/native/performance/field-access/unsafe-mutant.txt` | `task:organization-archive/internal/native/performance/field-access/unsafe-mutant.txt` | raw evidence: internal/native |
| `internal/native/performance/field-access/vet.log.gz` | `task:organization-archive/internal/native/performance/field-access/vet.log.gz` | raw evidence: internal/native |
| `internal/native/performance/utf16-cache/REPORT.md` | `docs/reports/internal/native/performance/utf16-cache/report.md` | report: internal/native/performance/utf16-cache |
| `internal/native/performance/utf16-cache/cache-mutant.log` | `task:organization-archive/internal/native/performance/utf16-cache/cache-mutant.log` | raw evidence: internal |
| `internal/native/performance/utf16-cache/final-core.log.gz` | `task:organization-archive/internal/native/performance/utf16-cache/final-core.log.gz` | raw evidence: internal/native |
| `internal/native/performance/utf16-cache/final-index.log` | `task:organization-archive/internal/native/performance/utf16-cache/final-index.log` | raw evidence: internal |
| `internal/native/performance/utf16-cache/final-profile-compare.log` | `task:organization-archive/internal/native/performance/utf16-cache/final-profile-compare.log` | raw evidence: internal |
| `internal/native/performance/utf16-cache/full-gate.log.gz` | `task:organization-archive/internal/native/performance/utf16-cache/full-gate.log.gz` | raw evidence: internal/native |
| `internal/native/performance/utf16-cache/gofmt.log` | `task:organization-archive/internal/native/performance/utf16-cache/gofmt.log` | raw evidence: internal |
| `internal/native/performance/utf16-cache/json-before-top.txt` | `task:organization-archive/internal/native/performance/utf16-cache/json-before-top.txt` | raw evidence: internal/native |
| `internal/native/performance/utf16-cache/json-before.callgrind.gz` | `task:organization-archive/internal/native/performance/utf16-cache/json-before.callgrind.gz` | raw evidence: internal/native |
| `internal/native/performance/utf16-cache/json-before.stderr` | `task:organization-archive/internal/native/performance/utf16-cache/json-before.stderr` | raw evidence: internal/native |
| `internal/native/performance/utf16-cache/json-final-parity.log.gz` | `task:organization-archive/internal/native/performance/utf16-cache/json-final-parity.log.gz` | raw evidence: internal/native |
| `internal/native/performance/utf16-cache/json-final-top.txt` | `task:organization-archive/internal/native/performance/utf16-cache/json-final-top.txt` | raw evidence: internal/native |
| `internal/native/performance/utf16-cache/json-final.callgrind.gz` | `task:organization-archive/internal/native/performance/utf16-cache/json-final.callgrind.gz` | raw evidence: internal/native |
| `internal/native/performance/utf16-cache/json-final.stderr` | `task:organization-archive/internal/native/performance/utf16-cache/json-final.stderr` | raw evidence: internal/native |
| `internal/native/performance/utf16-cache/measurements.txt` | `task:organization-archive/internal/native/performance/utf16-cache/measurements.txt` | raw evidence: internal/native |
| `internal/native/performance/utf16-cache/scanner-before-top.txt` | `task:organization-archive/internal/native/performance/utf16-cache/scanner-before-top.txt` | raw evidence: internal/native |
| `internal/native/performance/utf16-cache/scanner-before.callgrind.gz` | `task:organization-archive/internal/native/performance/utf16-cache/scanner-before.callgrind.gz` | raw evidence: internal/native |
| `internal/native/performance/utf16-cache/scanner-before.stderr` | `task:organization-archive/internal/native/performance/utf16-cache/scanner-before.stderr` | raw evidence: internal/native |
| `internal/native/performance/utf16-cache/scanner-final-parity.log.gz` | `task:organization-archive/internal/native/performance/utf16-cache/scanner-final-parity.log.gz` | raw evidence: internal/native |
| `internal/native/performance/utf16-cache/scanner-final-top.txt` | `task:organization-archive/internal/native/performance/utf16-cache/scanner-final-top.txt` | raw evidence: internal/native |
| `internal/native/performance/utf16-cache/scanner-final.callgrind.gz` | `task:organization-archive/internal/native/performance/utf16-cache/scanner-final.callgrind.gz` | raw evidence: internal/native |
| `internal/native/performance/utf16-cache/scanner-final.stderr` | `task:organization-archive/internal/native/performance/utf16-cache/scanner-final.stderr` | raw evidence: internal/native |
| `internal/native/performance/utf16-cache/unsafe-mutant.txt` | `task:organization-archive/internal/native/performance/utf16-cache/unsafe-mutant.txt` | raw evidence: internal/native |
| `internal/native/performance/utf16-cache/vet.log` | `task:organization-archive/internal/native/performance/utf16-cache/vet.log` | raw evidence: internal |
| `review/class-features-coverage/brand.txt` | `task:organization-archive/review/class-features-coverage/brand.txt` | review archive |
| `review/fnexpr-coverage/refused.txt` | `task:organization-archive/review/fnexpr-coverage/refused.txt` | review archive |
| `review/generator-oct6/planted-unsigned-shift.txt` | `task:organization-archive/review/generator-oct6/planted-unsigned-shift.txt` | review archive |
| `review/library-map-set/after.json` | `task:organization-archive/review/library-map-set/after.json` | review archive |
| `review/library-map-set/after.log` | `task:organization-archive/review/library-map-set/after.log` | review archive |
| `review/library-map-set/before.json` | `task:organization-archive/review/library-map-set/before.json` | review archive |
| `review/library-map-set/before.log` | `task:organization-archive/review/library-map-set/before.log` | review archive |
| `review/library-map-set/counts.log` | `task:organization-archive/review/library-map-set/counts.log` | review archive |
| `review/library-map-set/cycle-mutant.log` | `task:organization-archive/review/library-map-set/cycle-mutant.log` | review archive |
| `review/library-map-set/final-cycles.log` | `task:organization-archive/review/library-map-set/final-cycles.log` | review archive |
| `review/library-map-set/final-lower.log` | `task:organization-archive/review/library-map-set/final-lower.log` | review archive |
| `review/library-map-set/final-mutants.log` | `task:organization-archive/review/library-map-set/final-mutants.log` | review archive |
| `review/library-map-set/final-oracle.log` | `task:organization-archive/review/library-map-set/final-oracle.log` | review archive |
| `review/library-map-set/format.log` | `task:organization-archive/review/library-map-set/format.log` | review archive |
| `review/library-map-set/gate.log` | `task:organization-archive/review/library-map-set/gate.log` | review archive |
| `review/library-map-set/mutants.log` | `task:organization-archive/review/library-map-set/mutants.log` | review archive |
| `review/library-map-set/normalize-rerun.log` | `task:organization-archive/review/library-map-set/normalize-rerun.log` | review archive |
| `review/library-map-set/oracle.log` | `task:organization-archive/review/library-map-set/oracle.log` | review archive |
| `review/library-map-set/report.md` | `task:organization-archive/review/library-map-set/report.md` | review archive |
| `review/library-map-set/setup.log` | `task:organization-archive/review/library-map-set/setup.log` | review archive |
| `review/library-map-set/vet.log` | `task:organization-archive/review/library-map-set/vet.log` | review archive |
| `review/object-coverage/from_entries.txt` | `task:organization-archive/review/object-coverage/from_entries.txt` | review archive |
| `review/oct6/REPORT.md` | `task:organization-archive/review/oct6/REPORT.md` | review archive |
| `review/oct6/class_oct6_construction.a` | `task:organization-archive/review/oct6/class_oct6_construction.a` | review archive |
| `review/regex-cycle-coverage/refused.txt` | `task:organization-archive/review/regex-cycle-coverage/refused.txt` | review archive |
| `stage1/cohere/css/NATIVE_REPORT.md` | `docs/reports/stage1/cohere/css/native-report.md` | report: stage1/cohere/css |
| `stage1/cohere/css/PERFORMANCE.md` | `docs/reports/stage1/cohere/css/performance.md` | report: stage1/cohere/css |
| `stage1/cohere/css/PRINTER_REPORT.md` | `docs/reports/stage1/cohere/css/printer-report.md` | report: stage1/cohere/css |
| `stage1/cohere/css/REPORT.md` | `docs/reports/stage1/cohere/css/report.md` | report: stage1/cohere/css |
| `stage1/cohere/css/performance/baseline-callgrind.out.gz` | `task:organization-archive/stage1/cohere/css/performance/baseline-callgrind.out.gz` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/performance/baseline-count.stderr` | `task:organization-archive/stage1/cohere/css/performance/baseline-count.stderr` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/performance/baseline-count.stdout` | `task:organization-archive/stage1/cohere/css/performance/baseline-count.stdout` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/performance/baseline-main.c.gz` | `task:organization-archive/stage1/cohere/css/performance/baseline-main.c.gz` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/performance/baseline-profile.stderr` | `task:organization-archive/stage1/cohere/css/performance/baseline-profile.stderr` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/performance/baseline-profile.stdout` | `task:organization-archive/stage1/cohere/css/performance/baseline-profile.stdout` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/performance/baseline-sha256.json` | `task:organization-archive/stage1/cohere/css/performance/baseline-sha256.json` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/performance/baseline-top.log` | `task:organization-archive/stage1/cohere/css/performance/baseline-top.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/performance/final-callgrind.out.gz` | `task:organization-archive/stage1/cohere/css/performance/final-callgrind.out.gz` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/performance/final-count.stderr` | `task:organization-archive/stage1/cohere/css/performance/final-count.stderr` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/performance/final-count.stdout` | `task:organization-archive/stage1/cohere/css/performance/final-count.stdout` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/performance/final-main.c.gz` | `task:organization-archive/stage1/cohere/css/performance/final-main.c.gz` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/performance/final-profile.stderr` | `task:organization-archive/stage1/cohere/css/performance/final-profile.stderr` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/performance/final-profile.stdout` | `task:organization-archive/stage1/cohere/css/performance/final-profile.stdout` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/performance/final-sha256.json` | `task:organization-archive/stage1/cohere/css/performance/final-sha256.json` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/performance/final-top.log` | `task:organization-archive/stage1/cohere/css/performance/final-top.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/performance/manifest.tsv` | `task:organization-archive/stage1/cohere/css/performance/manifest.tsv` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/performance/measurements.json` | `task:organization-archive/stage1/cohere/css/performance/measurements.json` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/performance/sample-expected.txt` | `task:organization-archive/stage1/cohere/css/performance/sample-expected.txt` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/performance/sample.txt` | `task:organization-archive/stage1/cohere/css/performance/sample.txt` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/css.log` | `task:organization-archive/stage1/cohere/css/verification/css.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/native-agreement.log` | `task:organization-archive/stage1/cohere/css/verification/native-agreement.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/native-build-parser.log` | `task:organization-archive/stage1/cohere/css/verification/native-build-parser.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/native-build-printer.log` | `task:organization-archive/stage1/cohere/css/verification/native-build-printer.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/native-cycle-count-stderr.log` | `task:organization-archive/stage1/cohere/css/verification/native-cycle-count-stderr.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/native-cycle-count-stdout.log` | `task:organization-archive/stage1/cohere/css/verification/native-cycle-count-stdout.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/native-cycle-oracle.log` | `task:organization-archive/stage1/cohere/css/verification/native-cycle-oracle.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/native-diff-check.log` | `task:organization-archive/stage1/cohere/css/verification/native-diff-check.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/native-fresh.log` | `task:organization-archive/stage1/cohere/css/verification/native-fresh.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/native-gofmt.log` | `task:organization-archive/stage1/cohere/css/verification/native-gofmt.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/native-memory-mutants.log` | `task:organization-archive/stage1/cohere/css/verification/native-memory-mutants.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/native-optimized-parser.log` | `task:organization-archive/stage1/cohere/css/verification/native-optimized-parser.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/native-optimized-printer.log` | `task:organization-archive/stage1/cohere/css/verification/native-optimized-printer.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/native-proofs.log` | `task:organization-archive/stage1/cohere/css/verification/native-proofs.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/native-raw-and-boundaries.log` | `task:organization-archive/stage1/cohere/css/verification/native-raw-and-boundaries.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/native-regex-regression.log` | `task:organization-archive/stage1/cohere/css/verification/native-regex-regression.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/native-setup.log` | `task:organization-archive/stage1/cohere/css/verification/native-setup.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/native-throughput.log` | `task:organization-archive/stage1/cohere/css/verification/native-throughput.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/native-vet.log` | `task:organization-archive/stage1/cohere/css/verification/native-vet.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/printer-boundaries.log` | `task:organization-archive/stage1/cohere/css/verification/printer-boundaries.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/printer-corpus.log` | `task:organization-archive/stage1/cohere/css/verification/printer-corpus.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/printer-css-regression.log` | `task:organization-archive/stage1/cohere/css/verification/printer-css-regression.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/printer-dependencies.log` | `task:organization-archive/stage1/cohere/css/verification/printer-dependencies.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/printer-diff-check.log` | `task:organization-archive/stage1/cohere/css/verification/printer-diff-check.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/printer-named-capture-before-integration-fix.log` | `task:organization-archive/stage1/cohere/css/verification/printer-named-capture-before-integration-fix.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/printer-native-gap.log` | `task:organization-archive/stage1/cohere/css/verification/printer-native-gap.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/printer-regex-oracle.log` | `task:organization-archive/stage1/cohere/css/verification/printer-regex-oracle.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/printer-regex-package.log` | `task:organization-archive/stage1/cohere/css/verification/printer-regex-package.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/printer-regression-before-integration-fix.log` | `task:organization-archive/stage1/cohere/css/verification/printer-regression-before-integration-fix.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/printer-setup-initial-failure.log` | `task:organization-archive/stage1/cohere/css/verification/printer-setup-initial-failure.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/printer-setup.log` | `task:organization-archive/stage1/cohere/css/verification/printer-setup.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/printer-throughput.log` | `task:organization-archive/stage1/cohere/css/verification/printer-throughput.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/printer-vet.log` | `task:organization-archive/stage1/cohere/css/verification/printer-vet.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/range-probes.log` | `task:organization-archive/stage1/cohere/css/verification/range-probes.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/regexp-oracle.log` | `task:organization-archive/stage1/cohere/css/verification/regexp-oracle.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/reused-slices.log` | `task:organization-archive/stage1/cohere/css/verification/reused-slices.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/setup.log` | `task:organization-archive/stage1/cohere/css/verification/setup.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/speed-accounting-mutant.log` | `task:organization-archive/stage1/cohere/css/verification/speed-accounting-mutant.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/speed-apt-failure.log` | `task:organization-archive/stage1/cohere/css/verification/speed-apt-failure.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/speed-archive-check.log` | `task:organization-archive/stage1/cohere/css/verification/speed-archive-check.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/speed-baseline-artifacts.log` | `task:organization-archive/stage1/cohere/css/verification/speed-baseline-artifacts.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/speed-diff.log` | `task:organization-archive/stage1/cohere/css/verification/speed-diff.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/speed-final-artifacts.log` | `task:organization-archive/stage1/cohere/css/verification/speed-final-artifacts.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/speed-gate.log` | `task:organization-archive/stage1/cohere/css/verification/speed-gate.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/speed-gofmt.log` | `task:organization-archive/stage1/cohere/css/verification/speed-gofmt.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/speed-regex-oracle.log` | `task:organization-archive/stage1/cohere/css/verification/speed-regex-oracle.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/speed-rejected-trim-asan.log` | `task:organization-archive/stage1/cohere/css/verification/speed-rejected-trim-asan.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/speed-setup.log` | `task:organization-archive/stage1/cohere/css/verification/speed-setup.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/speed-timings.log` | `task:organization-archive/stage1/cohere/css/verification/speed-timings.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/speed-vet.log` | `task:organization-archive/stage1/cohere/css/verification/speed-vet.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/throughput.log` | `task:organization-archive/stage1/cohere/css/verification/throughput.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/css/verification/upstream-surrogate.log` | `task:organization-archive/stage1/cohere/css/verification/upstream-surrogate.log` | raw evidence: stage1/cohere/css |
| `stage1/cohere/cssnumbers/REPORT.md` | `docs/reports/stage1/cohere/cssnumbers/report.md` | report: stage1/cohere/cssnumbers |
| `stage1/cohere/cssstrings/REPORT.md` | `docs/reports/stage1/cohere/cssstrings/report.md` | report: stage1/cohere/cssstrings |
| `stage1/cohere/json/PERFORMANCE.md` | `docs/reports/stage1/cohere/json/performance.md` | report: stage1/cohere/json |
| `stage1/cohere/json/performance/baseline-counted.stderr` | `task:organization-archive/stage1/cohere/json/performance/baseline-counted.stderr` | raw evidence: stage1/cohere/json |
| `stage1/cohere/json/performance/baseline-top.txt` | `task:organization-archive/stage1/cohere/json/performance/baseline-top.txt` | raw evidence: stage1/cohere/json |
| `stage1/cohere/json/performance/baseline.callgrind.gz` | `task:organization-archive/stage1/cohere/json/performance/baseline.callgrind.gz` | raw evidence: stage1/cohere/json |
| `stage1/cohere/json/performance/baseline.stderr` | `task:organization-archive/stage1/cohere/json/performance/baseline.stderr` | raw evidence: stage1/cohere/json |
| `stage1/cohere/json/performance/c-equivalence.log` | `task:organization-archive/stage1/cohere/json/performance/c-equivalence.log` | raw evidence: stage1/cohere/json |
| `stage1/cohere/json/performance/final-counted.stderr` | `task:organization-archive/stage1/cohere/json/performance/final-counted.stderr` | raw evidence: stage1/cohere/json |
| `stage1/cohere/json/performance/final-top.txt` | `task:organization-archive/stage1/cohere/json/performance/final-top.txt` | raw evidence: stage1/cohere/json |
| `stage1/cohere/json/performance/final.callgrind.gz` | `task:organization-archive/stage1/cohere/json/performance/final.callgrind.gz` | raw evidence: stage1/cohere/json |
| `stage1/cohere/json/performance/final.stderr` | `task:organization-archive/stage1/cohere/json/performance/final.stderr` | raw evidence: stage1/cohere/json |
| `stage1/cohere/json/performance/format-check.log` | `task:organization-archive/stage1/cohere/json/performance/format-check.log` | raw evidence: stage1/cohere/json |
| `stage1/cohere/json/performance/gofmt.log` | `task:organization-archive/stage1/cohere/json/performance/gofmt.log` | raw evidence: stage1/cohere/json |
| `stage1/cohere/json/performance/lint.log` | `task:organization-archive/stage1/cohere/json/performance/lint.log` | raw evidence: stage1/cohere/json |
| `stage1/cohere/json/performance/manifest.txt` | `task:organization-archive/stage1/cohere/json/performance/manifest.txt` | raw evidence: stage1/cohere/json |
| `stage1/cohere/json/performance/measurements.txt` | `task:organization-archive/stage1/cohere/json/performance/measurements.txt` | raw evidence: stage1/cohere/json |
| `stage1/cohere/json/performance/mutant-accounting.log` | `task:organization-archive/stage1/cohere/json/performance/mutant-accounting.log` | raw evidence: stage1/cohere/json |
| `stage1/cohere/json/performance/oracle.log` | `task:organization-archive/stage1/cohere/json/performance/oracle.log` | raw evidence: stage1/cohere/json |
| `stage1/cohere/json/performance/prepare.log` | `task:organization-archive/stage1/cohere/json/performance/prepare.log` | raw evidence: stage1/cohere/json |
| `stage1/cohere/json/performance/snapshots.log` | `task:organization-archive/stage1/cohere/json/performance/snapshots.log` | raw evidence: stage1/cohere/json |
| `stage1/cohere/json/performance/streams.log` | `task:organization-archive/stage1/cohere/json/performance/streams.log` | raw evidence: stage1/cohere/json |
| `stage1/cohere/json/performance/suite.log` | `task:organization-archive/stage1/cohere/json/performance/suite.log` | raw evidence: stage1/cohere/json |
| `stage1/cohere/json/performance/timing.log` | `task:organization-archive/stage1/cohere/json/performance/timing.log` | raw evidence: stage1/cohere/json |
| `stage1/cohere/json/performance/vet.log` | `task:organization-archive/stage1/cohere/json/performance/vet.log` | raw evidence: stage1/cohere/json |
| `stage1/cohere/lint/PERFORMANCE.md` | `docs/reports/stage1/cohere/lint/performance.md` | report: stage1/cohere/lint |
| `stage1/cohere/lint/REPORT.md` | `docs/reports/stage1/cohere/lint/report.md` | report: stage1/cohere/lint |
| `stage1/cohere/lint/performance/1-build-parity.log` | `task:organization-archive/stage1/cohere/lint/performance/1-build-parity.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/1-cohere.log` | `task:organization-archive/stage1/cohere/lint/performance/1-cohere.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/1-release-and-mutant.log` | `task:organization-archive/stage1/cohere/lint/performance/1-release-and-mutant.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/1-scalar-comments-callgrind.stderr` | `task:organization-archive/stage1/cohere/lint/performance/1-scalar-comments-callgrind.stderr` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/1-scalar-comments-measurements.json` | `task:organization-archive/stage1/cohere/lint/performance/1-scalar-comments-measurements.json` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/1-scalar-comments-profile.log` | `task:organization-archive/stage1/cohere/lint/performance/1-scalar-comments-profile.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/1-scalar-comments.callgrind.gz` | `task:organization-archive/stage1/cohere/lint/performance/1-scalar-comments.callgrind.gz` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/2-build-parity.log` | `task:organization-archive/stage1/cohere/lint/performance/2-build-parity.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/2-cohere.log` | `task:organization-archive/stage1/cohere/lint/performance/2-cohere.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/2-position-arrays-callgrind.stderr` | `task:organization-archive/stage1/cohere/lint/performance/2-position-arrays-callgrind.stderr` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/2-position-arrays-measurements.json` | `task:organization-archive/stage1/cohere/lint/performance/2-position-arrays-measurements.json` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/2-position-arrays-profile.log` | `task:organization-archive/stage1/cohere/lint/performance/2-position-arrays-profile.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/2-position-arrays-rss.log` | `task:organization-archive/stage1/cohere/lint/performance/2-position-arrays-rss.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/2-position-arrays.callgrind.gz` | `task:organization-archive/stage1/cohere/lint/performance/2-position-arrays.callgrind.gz` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/2-release-and-mutant.log` | `task:organization-archive/stage1/cohere/lint/performance/2-release-and-mutant.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/2b-build-parity.log` | `task:organization-archive/stage1/cohere/lint/performance/2b-build-parity.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/2b-cohere.log` | `task:organization-archive/stage1/cohere/lint/performance/2b-cohere.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/2b-position-arrays-callgrind.stderr` | `task:organization-archive/stage1/cohere/lint/performance/2b-position-arrays-callgrind.stderr` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/2b-position-arrays-measurements.json` | `task:organization-archive/stage1/cohere/lint/performance/2b-position-arrays-measurements.json` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/2b-position-arrays-profile.log` | `task:organization-archive/stage1/cohere/lint/performance/2b-position-arrays-profile.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/2b-position-arrays.callgrind.gz` | `task:organization-archive/stage1/cohere/lint/performance/2b-position-arrays.callgrind.gz` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/2b-release.log` | `task:organization-archive/stage1/cohere/lint/performance/2b-release.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/3-build-parity.log` | `task:organization-archive/stage1/cohere/lint/performance/3-build-parity.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/3-cohere.log` | `task:organization-archive/stage1/cohere/lint/performance/3-cohere.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/3-filled-masks-callgrind.stderr` | `task:organization-archive/stage1/cohere/lint/performance/3-filled-masks-callgrind.stderr` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/3-filled-masks-measurements.json` | `task:organization-archive/stage1/cohere/lint/performance/3-filled-masks-measurements.json` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/3-filled-masks-profile.log` | `task:organization-archive/stage1/cohere/lint/performance/3-filled-masks-profile.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/3-filled-masks.callgrind.gz` | `task:organization-archive/stage1/cohere/lint/performance/3-filled-masks.callgrind.gz` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/3-release.log` | `task:organization-archive/stage1/cohere/lint/performance/3-release.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/4-build-parity.log` | `task:organization-archive/stage1/cohere/lint/performance/4-build-parity.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/4-cohere.log` | `task:organization-archive/stage1/cohere/lint/performance/4-cohere.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/4-kind-mutant.log` | `task:organization-archive/stage1/cohere/lint/performance/4-kind-mutant.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/4-kind-switches-callgrind.stderr` | `task:organization-archive/stage1/cohere/lint/performance/4-kind-switches-callgrind.stderr` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/4-kind-switches-measurements.json` | `task:organization-archive/stage1/cohere/lint/performance/4-kind-switches-measurements.json` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/4-kind-switches-profile.log` | `task:organization-archive/stage1/cohere/lint/performance/4-kind-switches-profile.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/4-kind-switches.callgrind.gz` | `task:organization-archive/stage1/cohere/lint/performance/4-kind-switches.callgrind.gz` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/4-release-and-mutant.log` | `task:organization-archive/stage1/cohere/lint/performance/4-release-and-mutant.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/5-build-parity.log` | `task:organization-archive/stage1/cohere/lint/performance/5-build-parity.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/5-cohere.log` | `task:organization-archive/stage1/cohere/lint/performance/5-cohere.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/5-kind-dispatch-callgrind.stderr` | `task:organization-archive/stage1/cohere/lint/performance/5-kind-dispatch-callgrind.stderr` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/5-kind-dispatch-measurements.json` | `task:organization-archive/stage1/cohere/lint/performance/5-kind-dispatch-measurements.json` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/5-kind-dispatch-profile.log` | `task:organization-archive/stage1/cohere/lint/performance/5-kind-dispatch-profile.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/5-kind-dispatch-rss.log` | `task:organization-archive/stage1/cohere/lint/performance/5-kind-dispatch-rss.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/5-kind-dispatch.callgrind.gz` | `task:organization-archive/stage1/cohere/lint/performance/5-kind-dispatch.callgrind.gz` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/accounting-mutant.log` | `task:organization-archive/stage1/cohere/lint/performance/accounting-mutant.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/baseline-callgrind.stderr` | `task:organization-archive/stage1/cohere/lint/performance/baseline-callgrind.stderr` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/baseline-full-top20.log` | `task:organization-archive/stage1/cohere/lint/performance/baseline-full-top20.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/baseline-full.callgrind.gz` | `task:organization-archive/stage1/cohere/lint/performance/baseline-full.callgrind.gz` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/baseline-measurements.json` | `task:organization-archive/stage1/cohere/lint/performance/baseline-measurements.json` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/baseline-profile.log` | `task:organization-archive/stage1/cohere/lint/performance/baseline-profile.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/baseline-rss.log` | `task:organization-archive/stage1/cohere/lint/performance/baseline-rss.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/baseline-top20.log` | `task:organization-archive/stage1/cohere/lint/performance/baseline-top20.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/baseline.callgrind.gz` | `task:organization-archive/stage1/cohere/lint/performance/baseline.callgrind.gz` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/filtered-oracle.log` | `task:organization-archive/stage1/cohere/lint/performance/filtered-oracle.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/final-cohere.log` | `task:organization-archive/stage1/cohere/lint/performance/final-cohere.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/final-suite.log` | `task:organization-archive/stage1/cohere/lint/performance/final-suite.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/final-vet.log` | `task:organization-archive/stage1/cohere/lint/performance/final-vet.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/go.callgrind.gz` | `task:organization-archive/stage1/cohere/lint/performance/go.callgrind.gz` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/go.stderr.log` | `task:organization-archive/stage1/cohere/lint/performance/go.stderr.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/hash-evidence.log` | `task:organization-archive/stage1/cohere/lint/performance/hash-evidence.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/setup.log` | `task:organization-archive/stage1/cohere/lint/performance/setup.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/performance/vet.log` | `task:organization-archive/stage1/cohere/lint/performance/vet.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/validation/cohere.log` | `task:organization-archive/stage1/cohere/lint/validation/cohere.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/validation/decoration-and-count-mutants.log` | `task:organization-archive/stage1/cohere/lint/validation/decoration-and-count-mutants.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/validation/filtered-oracle.log` | `task:organization-archive/stage1/cohere/lint/validation/filtered-oracle.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/validation/gaps.log` | `task:organization-archive/stage1/cohere/lint/validation/gaps.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/validation/release.log` | `task:organization-archive/stage1/cohere/lint/validation/release.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/validation/setup.log` | `task:organization-archive/stage1/cohere/lint/validation/setup.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/validation/suite-and-mutants.log` | `task:organization-archive/stage1/cohere/lint/validation/suite-and-mutants.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/validation/vet.log` | `task:organization-archive/stage1/cohere/lint/validation/vet.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/volume_evidence/batch1-final-cohere.log` | `task:organization-archive/stage1/cohere/lint/volume_evidence/batch1-final-cohere.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/volume_evidence/batch1-final-parity.log` | `task:organization-archive/stage1/cohere/lint/volume_evidence/batch1-final-parity.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/volume_evidence/batch1-format.log` | `task:organization-archive/stage1/cohere/lint/volume_evidence/batch1-format.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/volume_evidence/batch1-mutants.log` | `task:organization-archive/stage1/cohere/lint/volume_evidence/batch1-mutants.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/volume_evidence/batch1-parity.log` | `task:organization-archive/stage1/cohere/lint/volume_evidence/batch1-parity.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/volume_evidence/batch2-filtered-oracle.log` | `task:organization-archive/stage1/cohere/lint/volume_evidence/batch2-filtered-oracle.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/volume_evidence/batch2-final-cohere.log` | `task:organization-archive/stage1/cohere/lint/volume_evidence/batch2-final-cohere.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/volume_evidence/batch2-full-package.log` | `task:organization-archive/stage1/cohere/lint/volume_evidence/batch2-full-package.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/volume_evidence/batch2-throughput.log` | `task:organization-archive/stage1/cohere/lint/volume_evidence/batch2-throughput.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/volume_evidence/batch2-vet.log` | `task:organization-archive/stage1/cohere/lint/volume_evidence/batch2-vet.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/volume_evidence/frequency-manifest.txt` | `task:organization-archive/stage1/cohere/lint/volume_evidence/frequency-manifest.txt` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/volume_evidence/frequency.log` | `task:organization-archive/stage1/cohere/lint/volume_evidence/frequency.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/lint/volume_evidence/setup.log` | `task:organization-archive/stage1/cohere/lint/volume_evidence/setup.log` | raw evidence: stage1/cohere/lint |
| `stage1/cohere/markdownblocks/REPORT.txt` | `docs/reports/stage1/cohere/markdownblocks/report.txt` | report: stage1/cohere/markdownblocks |
| `stage1/cohere/markdowninline/REPORT.txt` | `docs/reports/stage1/cohere/markdowninline/report.txt` | report: stage1/cohere/markdowninline |
| `stage1/cohere/selector/REPORT.md` | `docs/reports/stage1/cohere/selector/report.md` | report: stage1/cohere/selector |
| `stage1/cohere/typeaware/PROFILE_REPORT.md` | `docs/reports/stage1/cohere/typeaware/profile-report.md` | report: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/REPORT.md` | `docs/reports/stage1/cohere/typeaware/report.md` | report: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/SIX_RULE_REPORT.md` | `docs/reports/stage1/cohere/typeaware/six-rule-report.md` | report: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/VOLUME_PROFILE_REPORT.md` | `docs/reports/stage1/cohere/typeaware/volume-profile-report.md` | report: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/VOLUME_REPORT.md` | `docs/reports/stage1/cohere/typeaware/volume-report.md` | report: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/after-rendering.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/after-rendering.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/baseline-cumulative.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/baseline-cumulative.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/baseline-native-leaves.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/baseline-native-leaves.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/baseline-raw.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/baseline-raw.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/baseline-top.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/baseline-top.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/baseline.cpu` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/baseline.cpu` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/bench-1-after.stderr.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/bench-1-after.stderr.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/bench-1-after.stdout.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/bench-1-after.stdout.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/bench-1-before.stderr.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/bench-1-before.stderr.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/bench-1-before.stdout.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/bench-1-before.stdout.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/bench-1-oracle.stderr.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/bench-1-oracle.stderr.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/bench-1-oracle.stdout.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/bench-1-oracle.stdout.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/bench-2-after.stderr.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/bench-2-after.stderr.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/bench-2-after.stdout.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/bench-2-after.stdout.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/bench-2-before.stderr.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/bench-2-before.stderr.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/bench-2-before.stdout.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/bench-2-before.stdout.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/bench-2-oracle.stderr.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/bench-2-oracle.stderr.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/bench-2-oracle.stdout.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/bench-2-oracle.stdout.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/bench-3-after.stderr.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/bench-3-after.stderr.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/bench-3-after.stdout.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/bench-3-after.stdout.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/bench-3-before.stderr.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/bench-3-before.stderr.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/bench-3-before.stdout.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/bench-3-before.stdout.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/bench-3-oracle.stderr.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/bench-3-oracle.stderr.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/bench-3-oracle.stdout.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/bench-3-oracle.stdout.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/cohere-final.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/cohere-final.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/compiler-go.stderr.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/compiler-go.stderr.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/compiler-go.stdout.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/compiler-go.stdout.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/compiler-native.stderr.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/compiler-native.stderr.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/compiler-native.stdout.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/compiler-native.stdout.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/compiler.manifest` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/compiler.manifest` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/facts-unit.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/facts-unit.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/first-attempt.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/first-attempt.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/first-rule.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/first-rule.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/gofmt.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/gofmt.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/measurements.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/measurements.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/node-filtered.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/node-filtered.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/paired-bench.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/paired-bench.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/phases-after-cumulative.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/phases-after-cumulative.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/phases-after-native-leaves.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/phases-after-native-leaves.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/phases-after-raw.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/phases-after-raw.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/phases-after.cpu` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/phases-after.cpu` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/phases-after.stderr.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/phases-after.stderr.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/phases-after.stdout.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/phases-after.stdout.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/phases-before-cumulative.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/phases-before-cumulative.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/phases-before-native-leaves.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/phases-before-native-leaves.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/phases-before-raw.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/phases-before-raw.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/phases-before.cpu` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/phases-before.cpu` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/phases-before.stderr.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/phases-before.stderr.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/phases-before.stdout.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/phases-before.stdout.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/regression.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/regression.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/requests.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/requests.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/setup.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/setup.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/suite.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/suite.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-profile/vet.log` | `task:organization-archive/stage1/cohere/typeaware/validation-profile/vet.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-six/asan-length-mutant.log` | `task:organization-archive/stage1/cohere/typeaware/validation-six/asan-length-mutant.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-six/cohere.log` | `task:organization-archive/stage1/cohere/typeaware/validation-six/cohere.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-six/compiler.manifest` | `task:organization-archive/stage1/cohere/typeaware/validation-six/compiler.manifest` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-six/first-rule.log` | `task:organization-archive/stage1/cohere/typeaware/validation-six/first-rule.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-six/gofmt.log` | `task:organization-archive/stage1/cohere/typeaware/validation-six/gofmt.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-six/guards.log` | `task:organization-archive/stage1/cohere/typeaware/validation-six/guards.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-six/measurements.log` | `task:organization-archive/stage1/cohere/typeaware/validation-six/measurements.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-six/node-filtered.log` | `task:organization-archive/stage1/cohere/typeaware/validation-six/node-filtered.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-six/node-one-byte.log` | `task:organization-archive/stage1/cohere/typeaware/validation-six/node-one-byte.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-six/regression.log` | `task:organization-archive/stage1/cohere/typeaware/validation-six/regression.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-six/requests.log` | `task:organization-archive/stage1/cohere/typeaware/validation-six/requests.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-six/setup.log` | `task:organization-archive/stage1/cohere/typeaware/validation-six/setup.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-six/six-compiler-go.stderr.log` | `task:organization-archive/stage1/cohere/typeaware/validation-six/six-compiler-go.stderr.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-six/six-compiler-go.stdout.log` | `task:organization-archive/stage1/cohere/typeaware/validation-six/six-compiler-go.stdout.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-six/six-compiler-native.stderr.log` | `task:organization-archive/stage1/cohere/typeaware/validation-six/six-compiler-native.stderr.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-six/six-compiler-native.stdout.log` | `task:organization-archive/stage1/cohere/typeaware/validation-six/six-compiler-native.stdout.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-six/six-cross-file-go.stderr.log` | `task:organization-archive/stage1/cohere/typeaware/validation-six/six-cross-file-go.stderr.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-six/six-cross-file-go.stdout.log` | `task:organization-archive/stage1/cohere/typeaware/validation-six/six-cross-file-go.stdout.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-six/six-cross-file-native.stderr.log` | `task:organization-archive/stage1/cohere/typeaware/validation-six/six-cross-file-native.stderr.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-six/six-cross-file-native.stdout.log` | `task:organization-archive/stage1/cohere/typeaware/validation-six/six-cross-file-native.stdout.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-six/six-generated-go.stderr.log` | `task:organization-archive/stage1/cohere/typeaware/validation-six/six-generated-go.stderr.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-six/six-generated-go.stdout.log` | `task:organization-archive/stage1/cohere/typeaware/validation-six/six-generated-go.stdout.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-six/six-generated-native.stderr.log` | `task:organization-archive/stage1/cohere/typeaware/validation-six/six-generated-native.stderr.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-six/six-generated-native.stdout.log` | `task:organization-archive/stage1/cohere/typeaware/validation-six/six-generated-native.stdout.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-six/suite.log` | `task:organization-archive/stage1/cohere/typeaware/validation-six/suite.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-six/vet.log` | `task:organization-archive/stage1/cohere/typeaware/validation-six/vet.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/after-cumulative.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/after-cumulative.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/after-root-cumulative.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/after-root-cumulative.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/after-root-final.pprof` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/after-root-final.pprof` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/after-root-final.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/after-root-final.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/after-root-final.stdout` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/after-root-final.stdout` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/agreement-final.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/agreement-final.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/agreement-root-final.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/agreement-root-final.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/agreement.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/agreement.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/baseline-cumulative.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/baseline-cumulative.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/baseline-nested.pprof` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/baseline-nested.pprof` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/baseline-nested.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/baseline-nested.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/baseline-nested.stdout` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/baseline-nested.stdout` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/bench.json` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/bench.json` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/bench.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/bench.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/checker-final.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/checker-final.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/checker.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/checker.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/cohere-final.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/cohere-final.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/cohere.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/cohere.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/compiler-1-before.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/compiler-1-before.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/compiler-1-go.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/compiler-1-go.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/compiler-1-native.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/compiler-1-native.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/compiler-2-before.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/compiler-2-before.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/compiler-2-go.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/compiler-2-go.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/compiler-2-native.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/compiler-2-native.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/compiler-3-before.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/compiler-3-before.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/compiler-3-go.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/compiler-3-go.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/compiler-3-native.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/compiler-3-native.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/decoder.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/decoder.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/findings.json` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/findings.json` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/gofmt.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/gofmt.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/initial-bench.json` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/initial-bench.json` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/initial-bench.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/initial-bench.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/missing-binding.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/missing-binding.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/node.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/node.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/phases-cache-attempt.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/phases-cache-attempt.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/phases-cache-attempt.stdout` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/phases-cache-attempt.stdout` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/phases.json` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/phases.json` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/repository-1-before.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/repository-1-before.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/repository-1-go.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/repository-1-go.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/repository-1-native.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/repository-1-native.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/repository-2-before.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/repository-2-before.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/repository-2-go.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/repository-2-go.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/repository-2-native.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/repository-2-native.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/repository-3-before.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/repository-3-before.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/repository-3-go.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/repository-3-go.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/repository-3-native.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/repository-3-native.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/requests.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/requests.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/setup.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/setup.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume-profile/vet.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume-profile/vet.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/bridge-final.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/bridge-final.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/checker-final.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/checker-final.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/checker-test.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/checker-test.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/cohere-check.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/cohere-check.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/cohere-final-edge.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/cohere-final-edge.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/cohere-final.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/cohere-final.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/compiler-1-go.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/compiler-1-go.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/compiler-1-native.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/compiler-1-native.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/compiler-2-go.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/compiler-2-go.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/compiler-2-native.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/compiler-2-native.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/compiler-3-go.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/compiler-3-go.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/compiler-3-native.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/compiler-3-native.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/compiler-all.counts` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/compiler-all.counts` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/compiler-all.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/compiler-all.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/compiler-sources.sha256.json` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/compiler-sources.sha256.json` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/compiler.counts` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/compiler.counts` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/compiler.findings.gz` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/compiler.findings.gz` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/compiler.manifest` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/compiler.manifest` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/controls.findings.gz` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/controls.findings.gz` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/cost-lint.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/cost-lint.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/fact-regression.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/fact-regression.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/findings.json` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/findings.json` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/gofmt-final.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/gofmt-final.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/guard-final.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/guard-final.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/node-filtered.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/node-filtered.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/property-info-m-1-go.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/property-info-m-1-go.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/property-info-m-1-native.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/property-info-m-1-native.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/property-info-m-2-go.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/property-info-m-2-go.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/property-info-m-2-native.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/property-info-m-2-native.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/property-info-m-3-go.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/property-info-m-3-go.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/property-info-m-3-native.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/property-info-m-3-native.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/query-probe.txt` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/query-probe.txt` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/query-results.json` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/query-results.json` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/query-source.ts.txt` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/query-source.ts.txt` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/regression.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/regression.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/repository-1-go.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/repository-1-go.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/repository-1-native.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/repository-1-native.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/repository-2-go.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/repository-2-go.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/repository-2-native.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/repository-2-native.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/repository-3-go.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/repository-3-go.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/repository-3-native.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/repository-3-native.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/repository-all.counts` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/repository-all.counts` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/repository-all.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/repository-all.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/repository-sources.sha256.json` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/repository-sources.sha256.json` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/repository.counts` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/repository.counts` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/repository.findings.gz` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/repository.findings.gz` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/repository.manifest` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/repository.manifest` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/root-declarations-mutant.go.txt` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/root-declarations-mutant.go.txt` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/root-declarations.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/root-declarations.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/root-extension-mutant.go.txt` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/root-extension-mutant.go.txt` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/root-extension.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/root-extension.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/setup.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/setup.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/vet-final.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/vet-final.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/volume-final.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/volume-final.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/whole-run-results.json` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/whole-run-results.json` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/whole-run.log` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/whole-run.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/widened-shape-1-go.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/widened-shape-1-go.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/widened-shape-1-native.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/widened-shape-1-native.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/widened-shape-2-go.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/widened-shape-2-go.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/widened-shape-2-native.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/widened-shape-2-native.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/widened-shape-3-go.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/widened-shape-3-go.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation-volume/widened-shape-3-native.stderr` | `task:organization-archive/stage1/cohere/typeaware/validation-volume/widened-shape-3-native.stderr` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation/bridge.log` | `task:organization-archive/stage1/cohere/typeaware/validation/bridge.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation/cohere.log` | `task:organization-archive/stage1/cohere/typeaware/validation/cohere.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation/compiler-regression-initial.log` | `task:organization-archive/stage1/cohere/typeaware/validation/compiler-regression-initial.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation/gofmt.log` | `task:organization-archive/stage1/cohere/typeaware/validation/gofmt.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation/node-oracle.log` | `task:organization-archive/stage1/cohere/typeaware/validation/node-oracle.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation/parser-scanner-lint.log` | `task:organization-archive/stage1/cohere/typeaware/validation/parser-scanner-lint.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation/setup.log` | `task:organization-archive/stage1/cohere/typeaware/validation/setup.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation/timing.log` | `task:organization-archive/stage1/cohere/typeaware/validation/timing.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation/typeaware.log` | `task:organization-archive/stage1/cohere/typeaware/validation/typeaware.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/cohere/typeaware/validation/vet.log` | `task:organization-archive/stage1/cohere/typeaware/validation/vet.log` | raw evidence: stage1/cohere/typeaware |
| `stage1/typescript/parser/REPORT.md` | `docs/reports/stage1/typescript/parser/report.md` | report: stage1/typescript/parser |
| `stage1/typescript/parser/WHOLE_REPORT.md` | `docs/reports/stage1/typescript/parser/whole-report.md` | report: stage1/typescript/parser |
| `stage1/typescript/parser/validation/callback-gap.log` | `task:organization-archive/stage1/typescript/parser/validation/callback-gap.log` | raw evidence: stage1/typescript/parser |
| `stage1/typescript/parser/validation/cohere.log` | `task:organization-archive/stage1/typescript/parser/validation/cohere.log` | raw evidence: stage1/typescript/parser |
| `stage1/typescript/parser/validation/count-mutant.log` | `task:organization-archive/stage1/typescript/parser/validation/count-mutant.log` | raw evidence: stage1/typescript/parser |
| `stage1/typescript/parser/validation/full-gate.log` | `task:organization-archive/stage1/typescript/parser/validation/full-gate.log` | raw evidence: stage1/typescript/parser |
| `stage1/typescript/parser/validation/gaps-and-generated.log` | `task:organization-archive/stage1/typescript/parser/validation/gaps-and-generated.log` | raw evidence: stage1/typescript/parser |
| `stage1/typescript/parser/validation/generated-and-compiler.log` | `task:organization-archive/stage1/typescript/parser/validation/generated-and-compiler.log` | raw evidence: stage1/typescript/parser |
| `stage1/typescript/parser/validation/oracle-source-view.log` | `task:organization-archive/stage1/typescript/parser/validation/oracle-source-view.log` | raw evidence: stage1/typescript/parser |
| `stage1/typescript/parser/validation/performance.log` | `task:organization-archive/stage1/typescript/parser/validation/performance.log` | raw evidence: stage1/typescript/parser |
| `stage1/typescript/parser/validation/setup.log` | `task:organization-archive/stage1/typescript/parser/validation/setup.log` | raw evidence: stage1/typescript/parser |
| `stage1/typescript/parser/validation/statement-module.log` | `task:organization-archive/stage1/typescript/parser/validation/statement-module.log` | raw evidence: stage1/typescript/parser |
| `stage1/typescript/parser/validation/whole-cohere.log` | `task:organization-archive/stage1/typescript/parser/validation/whole-cohere.log` | raw evidence: stage1/typescript/parser |
| `stage1/typescript/parser/validation/whole-compiler.log` | `task:organization-archive/stage1/typescript/parser/validation/whole-compiler.log` | raw evidence: stage1/typescript/parser |
| `stage1/typescript/parser/validation/whole-filtered-oracle.log` | `task:organization-archive/stage1/typescript/parser/validation/whole-filtered-oracle.log` | raw evidence: stage1/typescript/parser |
| `stage1/typescript/parser/validation/whole-final.log` | `task:organization-archive/stage1/typescript/parser/validation/whole-final.log` | raw evidence: stage1/typescript/parser |
| `stage1/typescript/parser/validation/whole-obsolete-assert.log` | `task:organization-archive/stage1/typescript/parser/validation/whole-obsolete-assert.log` | raw evidence: stage1/typescript/parser |
| `stage1/typescript/parser/validation/whole-performance.log` | `task:organization-archive/stage1/typescript/parser/validation/whole-performance.log` | raw evidence: stage1/typescript/parser |
| `stage1/typescript/parser/validation/whole-regression-final.log` | `task:organization-archive/stage1/typescript/parser/validation/whole-regression-final.log` | raw evidence: stage1/typescript/parser |
| `stage1/typescript/parser/validation/whole-regression.log` | `task:organization-archive/stage1/typescript/parser/validation/whole-regression.log` | raw evidence: stage1/typescript/parser |
| `stage1/typescript/parser/validation/whole-setup.log` | `task:organization-archive/stage1/typescript/parser/validation/whole-setup.log` | raw evidence: stage1/typescript/parser |
| `stage1/typescript/parser/validation/whole-vet-final.log` | `task:organization-archive/stage1/typescript/parser/validation/whole-vet-final.log` | raw evidence: stage1/typescript/parser |
| `stage1/typescript/parser/validation/whole-vet.log` | `task:organization-archive/stage1/typescript/parser/validation/whole-vet.log` | raw evidence: stage1/typescript/parser |
| `stage1/typescript/scanner/PERFORMANCE.md` | `docs/reports/stage1/typescript/scanner/performance.md` | report: stage1/typescript/scanner |
| `stage1/typescript/scanner/REPORT.md` | `docs/reports/stage1/typescript/scanner/report.md` | report: stage1/typescript/scanner |
| `stage1/typescript/scanner/performance/1-comparison.log` | `task:organization-archive/stage1/typescript/scanner/performance/1-comparison.log` | raw evidence: stage1/typescript/scanner |
| `stage1/typescript/scanner/performance/1-identifier.callgrind.stderr` | `task:organization-archive/stage1/typescript/scanner/performance/1-identifier.callgrind.stderr` | raw evidence: stage1/typescript/scanner |
| `stage1/typescript/scanner/performance/1-identifier.txt` | `task:organization-archive/stage1/typescript/scanner/performance/1-identifier.txt` | raw evidence: stage1/typescript/scanner |
| `stage1/typescript/scanner/performance/2-advance.callgrind.stderr` | `task:organization-archive/stage1/typescript/scanner/performance/2-advance.callgrind.stderr` | raw evidence: stage1/typescript/scanner |
| `stage1/typescript/scanner/performance/2-advance.txt` | `task:organization-archive/stage1/typescript/scanner/performance/2-advance.txt` | raw evidence: stage1/typescript/scanner |
| `stage1/typescript/scanner/performance/2-comparison.log` | `task:organization-archive/stage1/typescript/scanner/performance/2-comparison.log` | raw evidence: stage1/typescript/scanner |
| `stage1/typescript/scanner/performance/3-comparison.log` | `task:organization-archive/stage1/typescript/scanner/performance/3-comparison.log` | raw evidence: stage1/typescript/scanner |
| `stage1/typescript/scanner/performance/3-count-driver.callgrind.stderr` | `task:organization-archive/stage1/typescript/scanner/performance/3-count-driver.callgrind.stderr` | raw evidence: stage1/typescript/scanner |
| `stage1/typescript/scanner/performance/3-count-driver.txt` | `task:organization-archive/stage1/typescript/scanner/performance/3-count-driver.txt` | raw evidence: stage1/typescript/scanner |
| `stage1/typescript/scanner/performance/4-bmp-read.callgrind.stderr` | `task:organization-archive/stage1/typescript/scanner/performance/4-bmp-read.callgrind.stderr` | raw evidence: stage1/typescript/scanner |
| `stage1/typescript/scanner/performance/4-bmp-read.txt` | `task:organization-archive/stage1/typescript/scanner/performance/4-bmp-read.txt` | raw evidence: stage1/typescript/scanner |
| `stage1/typescript/scanner/performance/4-comparison.log` | `task:organization-archive/stage1/typescript/scanner/performance/4-comparison.log` | raw evidence: stage1/typescript/scanner |
| `stage1/typescript/scanner/performance/5-comparison.log` | `task:organization-archive/stage1/typescript/scanner/performance/5-comparison.log` | raw evidence: stage1/typescript/scanner |
| `stage1/typescript/scanner/performance/5-punctuation.callgrind.stderr` | `task:organization-archive/stage1/typescript/scanner/performance/5-punctuation.callgrind.stderr` | raw evidence: stage1/typescript/scanner |
| `stage1/typescript/scanner/performance/5-punctuation.txt` | `task:organization-archive/stage1/typescript/scanner/performance/5-punctuation.txt` | raw evidence: stage1/typescript/scanner |
| `stage1/typescript/scanner/performance/6-comparison.log` | `task:organization-archive/stage1/typescript/scanner/performance/6-comparison.log` | raw evidence: stage1/typescript/scanner |
| `stage1/typescript/scanner/performance/6-raw-units.callgrind.stderr` | `task:organization-archive/stage1/typescript/scanner/performance/6-raw-units.callgrind.stderr` | raw evidence: stage1/typescript/scanner |
| `stage1/typescript/scanner/performance/6-raw-units.txt` | `task:organization-archive/stage1/typescript/scanner/performance/6-raw-units.txt` | raw evidence: stage1/typescript/scanner |
| `stage1/typescript/scanner/performance/7-single-piece.callgrind.gz` | `task:organization-archive/stage1/typescript/scanner/performance/7-single-piece.callgrind.gz` | raw evidence: stage1/typescript/scanner |
| `stage1/typescript/scanner/performance/7-single-piece.callgrind.stderr` | `task:organization-archive/stage1/typescript/scanner/performance/7-single-piece.callgrind.stderr` | raw evidence: stage1/typescript/scanner |
| `stage1/typescript/scanner/performance/7-single-piece.txt` | `task:organization-archive/stage1/typescript/scanner/performance/7-single-piece.txt` | raw evidence: stage1/typescript/scanner |
| `stage1/typescript/scanner/performance/accounting-mutant.log` | `task:organization-archive/stage1/typescript/scanner/performance/accounting-mutant.log` | raw evidence: stage1/typescript/scanner |
| `stage1/typescript/scanner/performance/all-snapshots.log` | `task:organization-archive/stage1/typescript/scanner/performance/all-snapshots.log` | raw evidence: stage1/typescript/scanner |
| `stage1/typescript/scanner/performance/baseline.callgrind.gz` | `task:organization-archive/stage1/typescript/scanner/performance/baseline.callgrind.gz` | raw evidence: stage1/typescript/scanner |
| `stage1/typescript/scanner/performance/baseline.callgrind.stderr` | `task:organization-archive/stage1/typescript/scanner/performance/baseline.callgrind.stderr` | raw evidence: stage1/typescript/scanner |
| `stage1/typescript/scanner/performance/baseline.txt` | `task:organization-archive/stage1/typescript/scanner/performance/baseline.txt` | raw evidence: stage1/typescript/scanner |
| `stage1/typescript/scanner/performance/cohere.log` | `task:organization-archive/stage1/typescript/scanner/performance/cohere.log` | raw evidence: stage1/typescript/scanner |
| `stage1/typescript/scanner/performance/filtered-oracle.log` | `task:organization-archive/stage1/typescript/scanner/performance/filtered-oracle.log` | raw evidence: stage1/typescript/scanner |
| `stage1/typescript/scanner/performance/final-comparison.log` | `task:organization-archive/stage1/typescript/scanner/performance/final-comparison.log` | raw evidence: stage1/typescript/scanner |
| `stage1/typescript/scanner/performance/go-profile.stderr` | `task:organization-archive/stage1/typescript/scanner/performance/go-profile.stderr` | raw evidence: stage1/typescript/scanner |
| `stage1/typescript/scanner/performance/go-profile.stdout` | `task:organization-archive/stage1/typescript/scanner/performance/go-profile.stdout` | raw evidence: stage1/typescript/scanner |
| `stage1/typescript/scanner/performance/go.callgrind.gz` | `task:organization-archive/stage1/typescript/scanner/performance/go.callgrind.gz` | raw evidence: stage1/typescript/scanner |
| `stage1/typescript/scanner/performance/measurements.json` | `task:organization-archive/stage1/typescript/scanner/performance/measurements.json` | raw evidence: stage1/typescript/scanner |
| `stage1/typescript/scanner/performance/release-mutant.log` | `task:organization-archive/stage1/typescript/scanner/performance/release-mutant.log` | raw evidence: stage1/typescript/scanner |

This ledger contains 603 whole-file moves: 27 within the repository and 576 to task attachments. It proposes no directory wildcard moves and no deletion without a verified archive. All unlisted whole-file paths stay as they are.

### Deferred mechanical function splits

These are partial extractions, not whole-file renames. Each source remains and keeps all unlisted declarations. Move complete declarations with their comments, preserve order within each group, and add only the imports each destination needs. Package names, APIs, function bodies, emitted output and ownership rules stay identical. These are separate units after the relevant branches drain; none touches the four protected files named in this task.

| Old path and complete functions/types to extract | New path |
|---|---|
| `internal/lower/object.go`: `forOf`, `forOfMap`, `switchStatement` | `internal/lower/control.go` |
| `internal/lower/object.go`: `arrayMethod`, `newArrayFilled`, `arrayVisit`, `arrayReduce`, `arraySort` | `internal/lower/library_array.go` |
| `internal/lower/object.go`: `stringMethod`, `stringCall`, `stringFromCodes` | `internal/lower/library_string.go` |
| `internal/lower/object.go`: `numberFormat`, `numberCall`, `refusedRandom` | `internal/lower/library_math_number.go` |
| `internal/lower/object.go`: `mapTypes`, `mapMethod` | `internal/lower/map.go` |
| `internal/fresh/fresh.go`: `object`, `objects`, `value`, `state` declarations and all their methods; `newest`, `older`, `siteOf`, `outsideValue`, `newState` | `internal/fresh/heap.go` |
| `internal/fresh/fresh.go`: `made`, `argument`, `functionName`, `origins`, `edgeTerms`, `deferredKeys`, `sortTerms`, `sortWriteKeys`, `String`, `equal`, `summarize`, `sortedKeys` | `internal/fresh/summaries.go` |
| `cmd/adamic-test262/adapt.go`: `rewriteVars`, `declSafe`, `capturedAcrossLoop`, `scopeWithin`, `annotateCallbacks`, `callbackFn`, `inferCall`, `receiverElem`, `rewriteEquals`, `rewriteThrows` | `cmd/adamic-test262/adapt_rewrites.go` |

Keep `objectLiteral`, object-field dispatch, array/tuple literal and index lowering in `object.go` for this pass; `newExpression` still dispatches more than one construction area. Keep `fresh.go` orchestration/interpreter and `adapt.go` parser/types together. Confirm destination symbol names on the integration snapshot before extraction. If a listed function is already moved, omit that extraction; never introduce another definition. Existing `CLAUDE.md` native-emitter and string-runtime ownership is binding and already addresses the earlier monoliths.

## Landing order with roughly 150 branches in flight

1. **Land this proposal alone.** Ahra records the base, owners and bundle-to-task destinations. Take a fresh tracked-path/size snapshot immediately before each later batch; the measured numbers here are frozen observations, not predictions of future main. Do not rebase this proposal into a moving inventory without relabeling its snapshot.

2. **Export evidence outside Git by area before removal.** Start with cloud JSON.stringify, then completed review bundles, then one native-performance bundle. Attach original files and a manifest with source commit, command, versions, timestamps when recorded, SHA-256 and attachment URLs. Compare archive bytes to Git blobs. Only then remove those exact old paths in a separate mechanical commit. Keep any unpromoted regression probe or inaccessible bundle in place. Raw evidence under `docs` goes to its task in one feature bundle at a time.

3. **Move durable report prose one area at a time.** Suggested order: bridge, regex benchmark, completed small cohere slices, native performance, then scanner and parser. Move a report and update its inbound/outbound links in the same commit; no prose revision or source cleanup. Keep one bundle to at most 25 file moves per commit; split larger raw bundles by run prefix (baseline, after, compiler, repository, mutants) and retain an archive manifest covering the whole original bundle.

4. **Drain the busy stage-1 areas before their evidence/report batches.** CSS, lint and typeaware have multiple report/validation generations and cross-area users. Process CSS parsing and printing separately, lint baseline/performance/volume separately, and typeaware initial/six/profile/volume/volume-profile separately. Preserve old run names; they identify measurements at different implementations. A raw-output commit and a report-link commit should each stay within one area whenever possible. Do not combine these with source imports or checker ABI changes.

5. **Make the single documentation naming change** after report links settle. Update all incoming references to `library_function_expressions_for_in.md` in that commit. Do not use it as a reason to bulk-rename source files or package directories.

6. **Split handwritten files last, with their owners.** Test262 rewrite extraction first because it avoids core compiler churn. Then one `object.go` group per commit (control, array, string, number, map). Then fresh heap and summaries in separate commits. Announce each old/new function inventory to Kirk and Ahra before landing so workers can finish against the old path or replay onto the new one. Never combine extraction with behavioral repairs, formatter sweeps, package renames or generated-table regeneration. Use `git log --follow -C1% -- <file>` for whole-function provenance.

7. **Leave semantic deduplication and extension migration to separate units.** Unicode folding consolidation, shared width data, report summarization, new fixture promotion, measurement-script default changes and `.ts` to `.a` conversion need their own independent oracle and mutant evidence. They are not hidden inside mechanical move batches. No broad directory move is necessary to improve organization.

For every batch, publish the exact path/function ledger against its starting SHA. Ahra lands one area batch at a time, then workers merge current main normally. Small rename/removal commits let Git recognize unchanged blobs and limit add/delete conflicts; do not force-push or rewrite history. No claim is made that Git will resolve every rename/edit conflict automatically. Avoid changing shared `go.mod`, `go.work`, root configs, counts or giant generated tables during archival batches.

## What stays and why

- `cmd`, all `internal` package paths and `bridge/tsgo` remain. Their responsibilities are coherent; a `src` wrapper would rewrite every Go import and build command. `internal/load/tsgo.go`, lowering/native adapters and `cmd/adamic/tsgo.go` are different layers of the bridge, not spare copies of it.
- `internal/native/runtime` stays embedded beside its emitter. Preserve runtime filenames, include order, static linkage and generators. `string.c` remains the sole translation unit for its private implementation headers. Do not move `emit.go`, `lower.go`, `native.go` or `oracle_test.go` in this sweep.
- `internal/oracle/testdata`, package-local `testdata`, `gaps`, `GAPS.md`, sample cases and known-upstream-difference inputs stay beside tests. Their paths encode purpose and some files are intentionally invalid. `internal/oracle/counts.md` is an executable regression baseline, not a historical run log; keep it there.
- `stage1/cohere/{css,cssnumbers,cssstrings,values,mediaquery,selector}` are related components with different upstream entry points and contracts. Markdown block and inline parsers likewise are separate slices. Preserve these package directories and independent Go/Node references. Stage 1 compares to upstream Go cohere/checker; a shared name is not proof that two ports can merge.
- Generated tables, attributed numerical ports, licenses and notices remain local. Generated files may dwarf algorithms; size alone does not justify hand-editing or fragmenting them. Any later generator split must regenerate reproducibly and compare bytes/semantics, retaining attribution.
- `dedication` stays top-level because the README identifies it as the first native program. `bench` stays because its programs and repeated measurement machinery are reusable. Keep `cases.json`, `originals.json`, `original-sources.json` and the benchmark methods with the regex runner; move only its historical `RESULTS.md` as listed.
- `docs/0.1.md`, `docs/memory.md`, feature designs, progress snapshots, strictness surveys, images and reproducible survey/UTF-16 scripts remain in `docs`. These are durable contracts or deliberate measured inputs. Logs are exported separately; curated JSON data is not removed merely for being JSON.
- Root `.gitignore`, module/workspace configuration, `tsconfig.json`, `CohereSettings.json`, `CLAUDE.md`, legal notices and tooling hooks stay. Never ignore `*.a`. Keep compiler options synchronized with `internal/load/load.go`; named imports/exports use explicit relative extensions and no import cycles. New Adamic source is `.a`; declarations such as the prelude remain `.d.ts`.
- Measurement scripts inside validation/performance bundles stay for now. They are runnable tools, not raw evidence. After export, an owner may give them explicit task-artifact inputs and redirect outputs to scratch in a separate unit. Do not assume adjacency-based scripts still work after evidence is removed.

## Validation and limits for this proposal

Toolchain setup: `bash cloud/setup.sh > /tmp/organization-setup.log 2>&1`, followed by `source /workspace/adamic-tools/env.sh`. Go 1.27.1, clang 20.1.8, Node v24.19.0. Timing: Go 0s, clang 0s, Node 0s, submodules 0s, build cache warm 97s, done 97s. `nproc`: 5; `cpu.max`: `400000 100000` (four-core quota). Setup compiled all packages and test binaries with `go build ./...` and `go test -count=1 -run "^$" ./...`; the latter executes no tests and is not a full gate.

The proposal integrity check recomputes the frozen Git-blob measurements, checks every listed old path, checks unique destinations, and compares the twenty largest entries. Negative controls alter a recorded directory file count, substitute a nonexistent move source, and duplicate a destination in memory; each must fail the corresponding check. Observed result: PASS for all 156 directory rows, twenty ranked files and 603 whole-file moves. The count mutant failed with `directory count: bench`; the nonexistent source failed with `missing source: bench/regex/MISSING.md`; the duplicate target failed with `duplicate destination`. Existing tracked files were unchanged. Scratch controls and logs stay outside the repository. Check command: `python3 /tmp/organization-check.py > /tmp/organization-check.log 2>&1`; the validator is unit scratch, while the complete measurement recipe and move inputs are recorded here.

Focused compiler oracle command: `go test ./internal/oracle -run "^TestTheOracleCatchesOneByte$" -count=1 -timeout 30m > /tmp/organization-oracle.log 2>&1`. It exercises the existing one-byte stdout mutant against the independent Node oracle; PASS: `ok github.com/system-inc/adamic/internal/oracle 9.274s`. The existing test changed the lowered dedication string by appending `!` and required exactly `stdout differs`, proving the source/native comparison caught the mutant. No compiler source mutant or new compiler check is introduced by this prose-only unit.

Before any future source extraction: run the touched packages plus the relevant uncached Node oracle, ASan/UBSan and leak checks; compare emitted C/JS for the same fixture inputs; run a behavioral mutant that only the claimed check catches. A link/count/manifest failure cannot stand in for a semantic mutant. Before main moves, the integration worker runs the complete uncached gate prescribed by `CLAUDE.md`, with output to a log.

Not covered here: exhaustive source correctness, tests for proposed future moves, full compiler gate execution, exhaustive Unicode/width equivalence, live branch conflict prediction, existing task attachment IDs or upload retention, and the cohere submodule’s internal organization. Proposed archive and split benefits are recommendations, not measured runtime improvements. No task was messaged and no pull request was opened.
