Built: JSX StringAttributeValue, Collapse isHexDigit and Node.IsContainer, one helper per .a file; 18 dependency entries across 12 rules.\
Commits: claims 6ed891f, caa9400, 3a2ecc5; implementations 02bb022, d020214, 9750420; final report follows.\
Commands and outputs: helper package PASS 241.009s; vet exit 0; filtered uncached external oracle PASS 1.011s; setup previously 75s, nproc 5.\
Mutants: all five new semantic mutants and one byte-domain refusal mutant caught; all 25 prior/inherited mutants rerun and caught.\
Not covered: whole-rule findings/fixes, dynamic fixture construction, cross-slot callback integration and the full repository test gate; zero additional fully helper-ready rules.

# Slot 01 fourth batch report

## Delivered behavior

`jsx_string_attribute_value.a` preserves ordered first-match behavior. It rejects wrong attribute containers, skips spreads and unnamed/namespaced properties, and returns absent immediately for a matching missing or expression initializer. Later duplicates cannot rescue such a match. A string literal is decoded through the separately owned UnescapeStringLiteralText callback and returns found even for an empty value. AttributeName and the selected name matcher are also explicit callbacks over the same parser-derived arena. Every expected result comes from the actual Go StringAttributeValue, including the wrong-kind control; actual Go supplies dependency callback observations independently.

`collapse_is_hex_digit.a` accepts exactly ASCII 0-9, a-f and A-F over Go's byte domain. Integer range checks refuse unsupported numeric inputs rather than silently applying byte semantics to a wider number. `collapse_node_is_container.a` accepts exactly rule, at-rule, context and at-root. Declaration, comment, empty, case variants and unknown kinds return false. Child-list presence does not determine the result. The receiver must be a non-null typed node, matching Go's nil receiver panic.

All new Adamic files use .a. Each production helper occupies its own file. No shared registration generator, shared rule test harness, compiler ownership file, existing rule directory or frozen readiness ledger was changed. Slot-owned Go overlays call the pinned cohere helpers without modifying cohere. There were no previously claimed rule ports in this helper unit. Previous nine helper implementations were already tested and pushed before this batch began.

## Claim selection and publication

Before each claim, fetched all origin codex/lint-helpers* branches and checked every claims file. All larger named helpers were reserved. The seven-consumer strict-option entry is a per-rule schema gap rather than a named shared helper, as established in the base report. The selected helpers tied the largest unclaimed named count at six. Each claim was pushed before its implementation; the JSX and hex implementations were tested and pushed before claiming the next helper.

JSX claim 6ed891f at 01:19:38 UTC precedes slot 02 972f103 at 01:20:12 and slot 04 2053731 at 01:20:22. Both later claimants explicitly withdrew their duplicate. Final wildcard fetch checked six remote branches and five claim files; the base branch's older comment bundle remains reserved outside claims/. No other active claim for either Collapse leaf was observed. `evidence/slot01-wave4/claims.json` retains the fetched claims and their branch SHAs. No further helpers were claimed.

Claims: 6ed891f, caa9400 and 3a2ecc5. Implementations: 02bb022, d020214 and 9750420. Publication uses ordinary commits and pushes to codex/lint-helpers-01; no PR, force push or history rewriting.

## Commands and retained observations

Every test command redirects stdout/stderr directly to a log. Paths below are relative to the helper directory; commands ran at repository root with the full `stage1/cohere/lint/helpers/` prefix.

- `source /workspace/adamic-tools/env.sh` in build shells.
- `go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=20m > evidence/slot01-wave4/final.log 2>&1`: PASS 241.009s, 28 top-level tests.
- `go vet ./... > evidence/slot01-wave4/vet.log 2>&1`: exit 0, empty log.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m > evidence/slot01-wave4/oracle.log 2>&1`: PASS 1.011s, six fixtures and six probe misses.
- Initial per-helper and combined successful fourth-batch runs are retained as jsx-initial.log, hex-initial.log, complete-initial.log and targeted.log. The six-test targeted run passed in 20.688s; final.log includes the subsequently strengthened actual-Go wrong-kind observation.
- `gofmt -l` over the slot-owned Go files and `git diff --check`: empty output.

No setup rerun was needed. This unit's setup completed in 75s on five processors: go ready 0s, clang ready 1s, node ready 1s, submodules ready 1s, build cache warm 75s, done 75s. cgroup cpu.max is 400000 100000 and memory 17.6 GB. The initial moving-tree setup failure and successful retry remain documented in SLOT01_REPORT.md and its retained setup evidence.

The JSX corpus reads all six consuming fixture files: 571 complete fixture string expressions, 576 source batches, 4,690 helper queries plus one wrong-kind control, 9,382 expected output lines. The Go parser finds JSX attributes; exact and ignoring-case matchers are tried against standard and actual attribute names. Controls cover spread/namespaced properties, empty strings, bare attributes, expression initializers, case and later duplicates. All outputs match source Node, sanitized native and emitted JavaScript.

The hexadecimal corpus reads all six consuming fixture files: 625 fixture strings, all their UTF-8 bytes and all 256 byte values, 27,316 actual-Go predicate observations. All outputs match source Node, sanitized native and emitted JavaScript. The additional number 256 is an explicit unsupported-domain refusal test, not a representable Go byte input.

The container corpus reads those same six files: 625 fixture strings, 697 actual-Go observations, including 54 nodes traversed from actual Go ParseCSS and explicit known/unknown kind controls. 145 source batches fail CSS parsing; these failures are recorded and do not count as helper comparisons or helper failures. Every raw string still receives an actual-Go kind predicate query. All 697 outputs match source Node, sanitized native and emitted JavaScript.

These are helper corpora, statically extracting complete Go string expressions including concatenations, labels and messages. They do not execute entire consuming-rule harnesses or dynamically generated fixtures. Sanitized native baselines and semantic mutants require clean stderr and successful execution. A compiler refusal, panic or sanitizer error does not count as a semantic catch.

## Every fourth-batch mutant

| Helper | Mutation | Observation that caught it |
|---|---|---|
| StringAttributeValue | continue after first matching non-string initializer | output line 3: mutant true, actual Go false |
| StringAttributeValue | bypass UnescapeStringLiteralText callback | output line 64: mutant `&#47;about`, actual Go `/about` |
| StringAttributeValue | continue after first matching nil initializer | output line 75: mutant true, actual Go false |
| isHexDigit | exclude uppercase F | output line 64: mutant false, actual Go true |
| Node.IsContainer | omit context kind | output line 3: mutant false, actual Go true |
| isHexDigit domain contract | remove integer-byte guard | compiled mutant successfully accepts 256 and prints false; source/native/emitted-JavaScript baselines refuse with the expected byte-domain error |

The first five compile and run successfully with empty stderr, and actual-Go stdout comparisons catch them. The sixth is a contract/refusal check: Go's byte parameter cannot represent 256, so no invented Go boolean observation is used. The mutated native program compiles and exits successfully; accepting 256 is the witnessed failure.

The full helper package also reruns all 25 prior/inherited mutants documented in SLOT01_REPORT.md, SLOT01_WAVE2_REPORT.md and SLOT01_WAVE3_REPORT.md. These include JSON control/oneOf/interpolation/unknown-field validation; path normalization; class expression detection; defaults and all three array freshness checks; factory pattern order/validity/false attributes; cache bypass/unbound/empty entries; quote stripping/short ranges; attribute configured membership/false map entries; entity overflow/surrogate/named table; component ASCII restriction/range stride. The retained final.log records their compiling witnesses. Earlier entity surrogate observation was strengthened to UTF-16 code units after a console-only mutant survived, as documented in the third report.

## Consumer dependency increments

### github.com/system-inc/cohere/internal/lint/ecmascript/jsx.StringAttributeValue

- `@next/next/google-font-display`
- `@next/next/google-font-preconnect`
- `@next/next/next-script-for-ga`
- `@next/next/no-css-tags`
- `@next/next/no-html-link-for-pages`
- `@next/next/no-page-custom-font`

### github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.isHexDigit

- `better-tailwindcss/enforce-canonical-classes`
- `better-tailwindcss/enforce-consistent-class-order`
- `better-tailwindcss/enforce-consistent-variant-order`
- `better-tailwindcss/enforce-shorthand-classes`
- `better-tailwindcss/no-conflicting-classes`
- `better-tailwindcss/no-unknown-classes`

### github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.*Node.IsContainer

- `better-tailwindcss/enforce-canonical-classes`
- `better-tailwindcss/enforce-consistent-class-order`
- `better-tailwindcss/enforce-consistent-variant-order`
- `better-tailwindcss/enforce-shorthand-classes`
- `better-tailwindcss/no-conflicting-classes`
- `better-tailwindcss/no-unknown-classes`

The batch removes 18 dependency entries across 12 distinct consumers. Neither the batch alone nor its addition to all previous slot 01 helpers makes an additional rule fully helper-ready. The cumulative slot-only fully helper-ready rules remain structure/network-no-direct-fetch, structure/react-hook-require-result-naming and structure/storage-no-direct-local-storage. This is helper availability accounting against the frozen ledger, not a claim that those rule ports are implemented.

`slot01_wave4_readiness.json` names every consumer and lists residual dependencies after this batch and after all twelve slot 01 helpers. Concurrent workers' unmerged helpers are not silently credited.

No whole-rule findings, spans, fixes, suggestion serialization, shared profile compilation or arbitrary cross-slot adapter integration was verified. JSX decoder/name matcher integration remains explicit. Invalid manually assembled arenas beyond the documented index refusals and a nil container receiver are outside the typed adapter contract. The full repository go test ./... gate was not run; verification covers the entire touched helper package, repository-wide vet and a filtered uncached external oracle.
