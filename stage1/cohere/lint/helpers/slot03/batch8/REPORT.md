Built decodeArbitraryValue, recursivelyDecodeArbitraryValues and registerThemeBreakpointVariants in separate .a files: eighteen dependency edges across six consumers.
Commits: landing-ready branch 8e4f388 before selection; claim 48303f8 pushed before source; implementation/evidence follow the final gate.
Commands: twenty-four-helper gate PASS 378.657s, 2,986,262 comparison lines; focused final-adapter gate PASS 86.956s; vet/format clean; six uncached input fixtures PASS 1.333s; capture identical.
Mutants: ten new compiling semantic variants caught; forty-eight in the full regression plus the additional parser-argument witness in the focused run, forty-nine total unique variants and the coverage mutant.
Not covered: full repository gate, whole-rule integration, independently ported parser/math/theme/registry dependencies, arbitrary malformed adapter graphs, full int64 and external Tailwind/corpora.

## Landing and claim

The only branch this worker pushes is codex/lint-helpers-03. Before selection, it was verified based on current origin/main e8ba3d5 with oracle-green landing evidence pushed at 8e4f388. No main or area branch is pushed by this worker. The complete eighteen-branch helper fetch and claims scan found these symbols unclaimed; all tie the highest available concrete-symbol count, six. Shared HELPERS.md reserves and delivers the comment bundle. Claim 48303f8 was pushed before implementation. A subsequent complete claim scan confirms unique ownership. No shared harness, registration generator, compiler or other worker directory changed.

## Behavior and representation

The recursive walk preserves Go's in-place value mutations and traversal order through a flat arena of finite ParseValue tree nodes. url and names ending _url decode their name while leaving their whole argument tree alone. var/theme and underscore-suffix names preserve underscores only in the first word argument, while still unescaping escaped underscores. First separators/functions and later arguments recurse normally. All other functions decode names and recurse; words and separators decode ordinarily; unknown kinds remain untouched. Node kinds and array geometry remain unchanged.

The decoder's no-parenthesis fast path calls only converter. The function path passes the original input to parser, performs the owned recursive walk, writes that mutated arena, then passes the CSS to math. Parser, converter, writer and math are explicit dependencies. Theme breakpoint registration returns on nil theme or absent group, requests only --breakpoint keys, skips empty/already-registered names, then calls RegisterFrameworkVariants at the exact group order with static kind. Group lookup, theme keys and registry mutation remain separately owned dependencies.

Tests use actual Go ParseValue tree geometry, Go intermediate CSS/math observations and actual private helper outputs via an owned temporary export overlay. The test-only converter and serializer are adapters, not additional published ports. The math adapter answers only for the exact expected intermediate CSS; a different input yields a distinct output. The parser adapter validates the exact original input; a modified argument produces a distinct arena. This proves orchestration and transformation semantics under the common AST/dependency adapter assumption. It does not claim independently ported parser/math implementations or consumer-rule execution on native.

All 157 captured runtime sources from all six consumers, derived tokens and controls produce 319 decoder/AST inputs, 96 custom trees and 5,104 theme/group states. Controls include escaped underscores, nested functions, url exemptions, uppercase non-exempt names, var/theme first words/separators/functions, unknown kinds, malformed value strings, Unicode/NUL, nil theme, absent group, empty/duplicate/existing names, unrelated container keys, and group orders -17/0/64/999. Actual Go theme key results feed the dependency adapter; the real registry produces expected registrations. Existing functional/compound registration kinds remain unchanged. Every new baseline matches 5,838 lines on real Go, source Node, emitted JavaScript and sanitized native.

Valid indices, finite tree geometry, valid UTF-8-derived strings, initialized sequential dependency state and safe-integer orders are adapter prerequisites. Arbitrary cyclic/aliased Go slice graphs and full Go int64 behavior are not covered. Callback implementations must honor their published contracts; caller wiring is integration work and missing dependencies remain in readiness.json.

## Validation and compiler finding

The initial native build refused reading named function references from dependency-object fields at main.a line 37. Closure wrappers within the owned adapter compile and preserve behavior; no compiler change was needed. evidence/focused.log is the initial failed attempt, not a pass or mutant catch. The closure-focused run passes in 30.558s with nine semantic variants. Exact parser-input validation and a corresponding tenth decoder mutant were then added; evidence/input-guard.log passes in 86.956s with the final adapter and all ten variants.

The complete regression command had already compiled its nine-variant table before the additional argument witness was added. It passes the final helper/adapter baselines and all forty-eight variants from that table. The final focused run supplies the extra witness, yielding forty-nine distinct proven semantic variants across all twenty-four owned helpers. A compiler failure, crash, stderr or sanitizer error never counts as a semantic catch. Every credited variant compiles and exits 0 without stderr before its wrong output is compared with Go. Variants are temporary copies and production source is never mutated.

The complete owned package emits 2,986,262 matched Go/Node/sanitized-native lines across eight batches; batches three through eight additionally compare emitted JavaScript. The missing-consumer coverage mutant is caught again. go vet ./... and the recorded formatting scan have empty logs and exit 0. Six uncached input fixtures pass in 1.333s, zero cache hits and six probe misses. Repeated capture and coverage metadata reproduce byte for byte.

The separate upstream capture package exits 1 on installed-Tailwind and corpus guards. /Users/kirkouimet/Projects/ahra/app/_theme/styles is unavailable and the external class corpora are absent. Capture occurs before those skips and still observes every selected consumer; this is not a passing full Go rule-package or complete findings/fixes/suggestions gate.

## Consumers and readiness

Each helper removes one dependency from each of:

- better-tailwindcss/enforce-canonical-classes
- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-consistent-variant-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

Eighteen edges across six unique rules, zero final blockers removed alone. readiness.json subtracts only this slot's twenty-four helpers, without assuming other worker branches integrated. Inventory readiness is a dependency calculation under the common AST adapter assumption, not rule completion.

## Commands and environment

Source /workspace/adamic-tools/env.sh. All test output is redirected directly to logs, never piped:

- python3 stage1/cohere/lint/helpers/slot03/batch8/testdata/regenerate.py: regenerate.log and capture.log; repeated capture in reproducibility.log.
- go test ./stage1/cohere/lint/helpers/slot03 -run '^TestBatch8' -count=1 -v -timeout=20m: closures.log and final input-guard.log.
- go test ./stage1/cohere/lint/helpers/slot03 -count=1 -v -timeout=20m: final.log.
- go vet ./...: vet.log.
- gofmt -l cmd internal stage1/cohere/lint/helpers/slot03: gofmt.log.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m: oracle.log.
- bash cloud/setup.sh: setup.log.

go version go1.27.1 linux/amd64
setup: go ready (0s)
clang version 20.1.8 (https://github.com/llvm/llvm-project 87f0227cb60147a26a1eeb4fb06e3b505e9c7261)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
v24.19.0
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (135s)
setup: done in 135s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB

Cohere pin 715ba94f3608a6500086b1076ce5cb7e51b836db. nproc=5, cgroup quota 400000 100000; Go 1.27.1, clang 20.1.8, Node 24.19.0. Setup cache warm compiles tests without running them and is not a full test gate.

## Every final new mutant witness

- TestBatch8Mutants/recursively_decode_arbitrary_values.a: at line 413: got "function:102,110,|function:118,97,114,|word:45,45,97,95,98,|separator:44,|word:32,99,32,100,|separator:44,|function:117,114,108,|word:97,32,98,|" Go "function:102,110,|function:118,97,114,|word:45,45,97,95,98,|separator:44,|word:32,99,32,100,|separator:44,|function:117,114,108,|word:97,95,98,|"
- TestBatch8Mutants/recursively_decode_arbitrary_values.a#01: at line 71: got "function:99,97,108,99,|function:118,97,114,|word:45,45,97,32,98,|word:43,49,112,120,|" Go "function:99,97,108,99,|function:118,97,114,|word:45,45,97,95,98,|word:43,49,112,120,|"
- TestBatch8Mutants/recursively_decode_arbitrary_values.a#02: at line 714: got "unknown:97,32,117,114,108,|word:45,45,97,95,98,|word:99,92,95,100,95,101,|word:95,102,95,103,|function:102,111,111,95,98,97,114,|word:104,95,105,|" Go "unknown:97,95,117,114,108,|word:45,45,97,95,98,|word:99,92,95,100,95,101,|word:95,102,95,103,|function:102,111,111,95,98,97,114,|word:104,95,105,|"
- TestBatch8Mutants/decode_arbitrary_value.a: at line 70: got "99,97,108,99,40,49,112,120,43,50,112,120,41," Go "99,97,108,99,40,49,112,120,32,43,32,50,112,120,41,"
- TestBatch8Mutants/decode_arbitrary_value.a#01: at line 70: got "99,97,108,99,40,49,112,120,43,50,112,120,41," Go "99,97,108,99,40,49,112,120,32,43,32,50,112,120,41,"
- TestBatch8Mutants/decode_arbitrary_value.a#02: at line 30: got "117,110,101,120,112,101,99,116,101,100,32,109,97,116,104,32,105,110,112,117,116,58,117,110,101,120,112,101,99,116,101,100,32,112,97,114,115,101,32,105,110,112,117,116,58,85,82,76,40,97,32,98,41,32," Go "85,82,76,40,97,32,98,41,"
- TestBatch8Mutants/register_theme_breakpoint_variants.a: at line 851: got "49,48,48,:-17:static|101,120,105,115,116,105,110,103,:1:functional|115,109,:2:compound|" Go "101,120,105,115,116,105,110,103,:1:functional|115,109,:2:compound|"
- TestBatch8Mutants/register_theme_breakpoint_variants.a#01: at line 855: got "49,48,48,:-17:static|101,120,105,115,116,105,110,103,:1:functional|115,109,:2:compound|" Go "101,120,105,115,116,105,110,103,:1:functional|115,109,:2:compound|"
- TestBatch8Mutants/register_theme_breakpoint_variants.a#02: at line 747: got "101,120,105,115,116,105,110,103,:1:static|115,109,:2:compound|" Go "101,120,105,115,116,105,110,103,:1:functional|115,109,:2:compound|"
- TestBatch8Mutants/register_theme_breakpoint_variants.a#03: at line 859: got "49,48,48,:-16:static|101,120,105,115,116,105,110,103,:1:functional|115,109,:2:compound|" Go "49,48,48,:-17:static|101,120,105,115,116,105,110,103,:1:functional|115,109,:2:compound|"

The prior thirty-nine exact witnesses are rerun in final.log and documented in the preceding batch reports. Full repository testing, rule integration and independently ported dependency wiring remain outside this passing gate.
