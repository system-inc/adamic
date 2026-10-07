Built: JSX AttributeName, property.Name and Tailwind ClassLiteralReaderFor, each in a separate .a file.
Commits: 369c879 (JSX), a202730 (property), c85c62f (reader and final consumer corpus); claims b25d3cd, 0806aac and 4a377bb were pushed before code.
Commands and outputs: setup 113s on nproc=5; uncached full helper package PASS 105.059s; filtered uncached oracle PASS 27.453s; go vet ./... PASS.
Mutants: accept namespaced JSX names, accept computed identifiers and remove the reader cache namespace; all compile and execute, then the Go comparisons catch them.
Not covered: full repository gate, integrated rule findings/fixes, live Tailwind design systems, arbitrary malformed AST projections or prerequisite implementations owned by other workers.

# Observations

Read CLAUDE.md, helper REPORT.md, README.md and readiness.json before changes. The specific unit's helper branch instruction overrides the generic main instruction. Base was origin/codex/lint-helpers at 29990b47913aa77d542045194b512484af995d29, which includes the four initial helpers. Explicit wildcard fetch was necessary because the checkout fetched only main by default. All six origin codex/lint-helpers* branches and their claims were inspected repeatedly and at final selection; final-claims.log.gz retains the final snapshot byte for byte.

Identifier ownership raced with slot 05; our claim was withdrawn. AttributeName claim b25d3cd precedes slot 05's duplicate by one second, and slot 05 withdrew it. Whitespace ownership raced with slot 03; its claim preceded ours by twelve seconds. Although the local exhaustive comparison passed, all whitespace deliverables were removed and that helper is not counted. ClassLiteralReaderFor replaced it after all-branch refresh and its claim was pushed before implementation. Highest unclaimed remaining-consumer counts at the three selections were 21, 14 and 12, respectively; reserved comments are excluded. Other workers' helpers are neither copied nor claimed delivered.

## Go behavior and evidence

AttributeName: 21 consumer rules, 21 fixture files, 1,056 captured inputs, 1,068 distinct parsed sources including controls, 19,610 AST nodes and 19,611 node queries. Go, Node source and ASan/UBSan native agree. Only JsxAttribute with Identifier name answers; missing, spread and namespaced names decline, while present-empty is distinguished from absent.

Property Name: 14 consumer rules, 15 fixture files, 3,308 captured inputs, 3,327 distinct parsed sources including controls, 44,212 AST nodes and 44,213 queries, each checked against all 64 flag combinations (2,829,632 verdicts). Go, Node source and ASan/UBSan native agree. Computed identifiers/private identifiers decline even with their flags set; only computed names unwrap parentheses. The projection carries real parser Text() including decoded identifiers and normalized numbers.

The combined capture has 35 consumers, 4,893 records and 4,364 unique rule/file/source inputs. The actual Go consumer tests in six packages pass (capture.log). The localization custom multi-file harness is captured with an overlay rather than modifying cohere. Regeneration is byte-for-byte reproducible, and an empty-capture control is refused. Corpus SHA256: e7c1bbac5d7a9e69dbcdb56636e16ecf0fba2a830d440a7ac328ffd83a952fc3. Coverage SHA256: 6d2546098c4d6778d1fef7a53f6e5f051d9842601072c7996781cfb15016cc5e. The coverage artifact enumerates every AST consumer and its actual captured inputs.

ClassLiteralReaderFor: 12 consumer rules share this contract. Eight real Go settings exercise default/empty settings, duplicates, invalid regex, embedded NUL and settings-key collisions. Eight per-setting binding/fill observations plus a 64-entry identity matrix give 80 observations matching Go, Node and sanitized native. Same-file cache identity preserves values; different files and nil/zero caches receive fresh values maps; compiled maps/patterns stay shared; heterogeneous cache type collisions compute fresh bindings; cache fills use tailwind.classValues: plus the real settings key. The actual Go oracle invokes the real private settings-key, compiled reader and rule.Cached implementations. The Adamic helper injects these prerequisites, using a lazy cache factory protocol documented in README.md. It does not supply the prerequisite ports or a production FileCache adapter.

Adamic explicitly refused the first callback capture as cycle-capable, then refused passing a generic function value. These compiler refusals were not credited as mutants. The final implementation passes explicit factory arguments and a capture-free binder, retaining lazy filling without compiler changes or weakened checks.

## Compiled mutants

- AttributeName: replace the Identifier-only name guard with false. Node and sanitized native both execute and agree; Go catches line 1140, got true: versus false:.
- Property Name: replace the computed Identifier/PrivateIdentifier rejection with false. Node and sanitized native both execute and agree; Go catches line 8290, got true:foo versus false:.
- ClassLiteralReaderFor: remove tailwind.classValues: from the cache key. Node and sanitized native both execute and agree; Go catches line 2, got fills:0:1:true versus fills:1:1:true.

Inherited compiling mutants also ran and were caught by the Go comparison: options_json at line 15157 (valid versus invalid), option_schema at line 15166 (valid versus invalid), policy_message at line 22166 (uninterpolated constructor versus sentinel é😀) and strict_options at line 15178 (valid versus invalid). The inherited baseline matched 22,412 lines, ten refusal cases matched and known-gap checks passed.

These compare independent Go results rather than source output alone. Test code refuses a mutant that does not compile, execute or differ semantically. Additional inherited helper mutants and message refusals run in the whole helper package and remain separately credited by the existing report.

## Commands

Run from the repository root, after source /workspace/adamic-tools/env.sh. All test output is redirected to files, never piped.

```
bash cloud/setup.sh > stage1/cohere/lint/helpers/slot02/evidence/setup.log 2>&1
python3 stage1/cohere/lint/helpers/testdata/slot02_capture.py > stage1/cohere/lint/helpers/slot02/evidence/regeneration.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot02/evidence/helpers.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m > stage1/cohere/lint/helpers/slot02/evidence/oracle.log 2>&1
go vet ./... > stage1/cohere/lint/helpers/slot02/evidence/vet.log 2>&1
git diff --check > stage1/cohere/lint/helpers/slot02/evidence/format.log 2>&1
```

Setup: Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s, build cache warm 113s, total 113s, nproc 5. Tools: Go 1.27.1, clang 20.1.8 and Node 24.19.0. Actual environment file is /workspace/adamic-tools/env.sh. A test command initially used system /usr/bin/go and failed with Go: Unknown option: test; sourcing the setup environment fixed it. No test failures are counted as passing evidence.

Whole helper package: PASS 105.059s, with new JSX/property/reader checks taking 9.48s/28.92s/5.16s respectively. Filtered oracle: all six input fixtures pass, 27.453s, native/node cache hits 0, probe misses 6. Vet and whitespace check exit 0 with empty logs. The full repository test gate was not run; the entire touched package plus this filtered oracle was used. No protected compiler files changed. New Adamic files are .a; the inherited options_json.ts is consumed unchanged, and temporary mutant copies rename it .a.

# Dependency handoff and inference

RULES.md lists all 47 consumer rules. These three helpers remove 47 helper dependency entries across 47 distinct rules, not 47 fully ready rules. Only complexity and grouped-accessor-pairs lose their final listed blocker from this slot alone: the frozen readiness count conditionally moves from 46 to 48. Other workers' dependencies are not assumed. readiness.json remains the original inventory rather than silently claiming rules implemented.

Finite corpus agreement is evidence, not proof for every possible program. Rule adapters must preserve the documented AST projection and cache factory contracts. Live Tailwind rule findings require the other workers' readers, cache adapters and design systems and were not exercised here. No PR opened.
