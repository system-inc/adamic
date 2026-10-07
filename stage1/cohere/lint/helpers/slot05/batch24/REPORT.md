# Slot 05 batch 24 report

Three helpers are complete in separate .a files: regexp.rewrite, regexsyntax.IsHexDigit and regexsyntax.AllHexDigits. The trio removes ten prerequisite occurrences across seven rules and no final blocker. Cumulative slot 05: 68 helpers, 365 occurrences across 73 consumers, 50 helper-ready rules under the frozen AST-adapter assumptions. No rule is declared ported.

Claim fa546b148e44f9b76cd198bd4ad5329d79236180 was pushed before all new source. All twenty origin helper branches and every claim tree were read before selection and refreshed before publication; evidence/ownership.json confirms no competing claim. Higher-count comments leaves remain owned by the shared bundle. rewrite had the highest remaining concrete fan-out, four consumers; both hex helpers tied the next highest, three each. Main b6b1538b0cebc4ba6741ac34f1aedb60293c1d06 and area d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898 remain current ancestors. Only codex/lint-helpers-05 is pushed; integration owns main and area branches.

## Behavior and consumers

rewrite preserves Go's output bytes, exactness, errors, escape dispatch, modifier option stack and external dependency argument/order contract. It translates syntax for the external regexp2 engine; it writes no matcher. UTF-8 decoding/encoding, capture counting, escapes, class construction, case widening and group checks remain explicit dependencies. Source, atom handles, outputs and errors use reversible byte encodings. No byte offset is used as a finding's UTF-16 position. The hex functions preserve ASCII byte ranges, nonempty validation and inspection of every byte. See README.md for the callback boundary.

rewrite removes one prerequisite from @next/next/no-html-link-for-pages, @typescript-eslint/no-empty-object-type, no-restricted-exports and no-restricted-imports. Each hex helper removes one prerequisite from no-control-regex, no-regex-spaces and no-useless-escape. CONSUMERS.md and readiness.json list exact mappings. These counts describe dependency removal, not whole-rule diagnostics, fixes or suggestions.

## Observed checks

Environment source: /workspace/adamic-tools/env.sh. Go 1.27.1, clang 20.1.8, Node 24.19.0; nproc 5, cgroup four-core quota, 17.6 GB. Setup succeeded:

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (41s)
setup: done in 41s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Every command wrote its output directly to a log:

```
bash cloud/setup.sh > /tmp/lint05-batch24-setup.log 2>&1
ADAMIC_SLOT05_BATCH24_EVIDENCE="$PWD/stage1/cohere/lint/helpers/slot05/batch24/evidence" ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch24 -count=1 -v -timeout=20m > /tmp/lint05-batch24-complete.log 2>&1
go vet ./... > /tmp/lint05-batch24-final-vet.log 2>&1
gofmt -l cmd internal stage1/cohere/lint/helpers/slot05/batch24 > /tmp/lint05-batch24-final-format.log
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v > /tmp/lint05-batch24-oracle.log 2>&1
```

Complete package PASS 247.634s: rewrite 233.64s, IsHexDigit 9.30s, AllHexDigits 4.69s. All 38,373 baseline helper calls match pinned actual Go cohere, source Node, emitted JavaScript on Node and native under ASan/UBSan. Vet and formatting have empty logs and exit zero. The six filtered uncached Node input fixtures PASS 1.133s, zero cache hits and six probe misses. The initial pooled ten-mutant gate also passed, 125.521s; the final gate supersedes it with 21 variants.

rewrite: all 905 distinct captured strings, including empty, from all four consumers, plus boundary sources, give 1,433 sources and 22,928 calls under all sixteen flags. Go test-file string occurrences per consumer are 386, 67, 483 and 1,270. Both hex corpora retain all 584 distinct captured strings from their three consumers (42, 338 and 473 occurrences respectively). IsHexDigit has 14,339 byte calls, including all 256 direct bytes and every byte in captured/boundary sources. AllHexDigits has 1,106 source calls. Coverage JSON lists every upstream file. Captured strings include options, prose and incomplete programs; they are helper inputs, not a rule finding replay. Invalid UTF-8 and every byte are covered without converting byte offsets to UTF-16.

The private overlay replaces dependency call names inside the original Go rewrite body and removes its now-unused UTF-8 import. Every dependency executes Go's actual original function. Pooling deduplicates 21,681 dependency records without removing any row. Every one of the 22,928 unpooled source/options/byte input rows was independently checked against the final pooled input. The answer builder is unchanged. Go case-extra iteration may emit equivalent class members in different orders across fresh oracle processes, observed as BA versus AB; exact per-generation output and trace remain the comparison target.

## Compiling semantic mutants

All 21 variants compile to native, execute successfully with empty stderr, then differ from actual Go. No compile failure, warning, refusal, panic, sanitizer failure or nonzero exit earns mutant credit. evidence/mutants.json gives each replacement, exact input row, first output line and both outputs.

Fourteen rewrite variants cover named-reference exactness, numeric-reference exactness, word/nonword class negation, assertion suffix position, class negation, modifier flags, Unicode case-widening arguments, partial output on an escape error, dotAll translation, nested option restoration, property-escape exactness, original versus current group-check options, class exactness and multiline start anchors. Two argument variants are detected by the dependency trace even where the resulting helper value can agree; their argument/order contract is part of the helper result comparison. All other variants produce output/error/exactness differences.

Four IsHexDigit variants exclude 9, A, a or f; all are caught by the corresponding exhaustive byte input. Three AllHexDigits variants accept empty, invert byte validation or omit the final byte; the empty and nonhex-byte controls catch them. The complete log records every exact witness. Prior 269 witnesses are preserved by identical retained helper/compiler/runtime/oracle Git object lists, documented in retained-input-identity.json; they were not freshly rerun this unit.

## Findings and limits

The first oracle overlay failed on an unused UTF-8 import; the next driver failed Adamic's string-only console typing. Both were fixed in owned test files. The unpooled corpus repeated 78 MB of dependency JSON and made native parsing slow; that run was stopped without claiming a pass. Pooling reduced it to 7.8 MB while retaining every input row, dependency and expected result within its generation. No corpus was trimmed. Failed and stopped logs are retained as such.

Not covered: regex compilation or matching, full rule findings/positions/fixes/suggestions, integration into rule modules, arbitrary configurations or an exhaustive pattern language proof. Specialized dependency implementations remain externally supplied. The full repository gate and its seventeen required stage 1 external-input checks were not run; none was skipped or credited as passing. No shared harness, rule registry or compiler file was edited. No unfinished claim and no fourth helper reserved.
