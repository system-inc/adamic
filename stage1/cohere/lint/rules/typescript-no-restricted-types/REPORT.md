Built the previously reserved no-restricted-types rule and all Google font display policy logic as .a modules.
Checks: 52 restricted-type upstream cases plus 3 witnesses, 209 compiler/stage1 files, and all 16 font JSX fixtures through an explicit Go JSX data adapter.
Outputs: restricted upstream 14,731 identical bytes; compiler/stage1 12,248,892; font policy 3,083, on Node, emitted JavaScript and sanitized native.
Mutants: reporting an allowed restricted type, and accepting display=block, both compile and differ only at output comparison on all three runtimes.
Uncovered: full JSX source parsing for fonts, shared integration, exhaustive malformed option errors, the full repository gate and representative font throughput.

The original no-restricted-types reservation is now implemented: keyword gates, name and generic reference checks, empty tuple/literal gates, Go Unicode whitespace stripping, raw and captured resolved ban data, exact custom messages, automatic replacements and each independent suggestion. The private reporter now prints every automatic fix separately from suggestions. Its existing three non-null/alias witnesses still pass (15,111 identical bytes). No shared repository file changed.

NoRestrictedTypes remains dependent on shared .a integration and a complete reporter. Its ordinary main driver refuses explicitly through the existing owned reporting guard. The new upstream shared harness is origin/codex/lint-harness-dot-a at 2650ad59. It is not merged here. In its single-suggestion shorthand, suggestion IDs are omitted, and automatic fix metadata is overwritten when a finding also carries suggestions. The owned independent detailed reporter retains both categories and all IDs. Upstream compiler inputs were all 77 TypeScript v6.0.3 compiler files at 050880ce59e30b356b686bd3144efe24f875ebc8, plus this branch's stage1 .ts/.a sources. Configured bans make the compiler comparison nontrivial; default options deliberately ban nothing.

The font policy is nested in font-policy/ to avoid advertising an unimplemented registered JSX listener. evaluate consumes the shared JSX contract: intrinsic tag, decoded first string href attribute and presence. The complete HTTPS stylesheet prefix gate, first matching query parameter, missing/bare parameter judgment, allowed/unknown values and exact messages are ported. The independent Go adapter parses all 16 original JSX fixture sources using Go cohere, extracts actual jsx.ElementParts/StringAttributeValue results, and runs the unmodified Go GoogleFontDisplay rule for the answer. This certifies rule policy against real JSX inputs but does not certify Adamic's parser or JSX helper. Shared JSX parsing still refuses the original source probe at offset 32; its earlier three-runtime evidence remains in the existing gap report. No source-level parity, fixed-source comparison or representative findings/s is claimed for this blocked source pipeline. Go offers no font fixes. Neither a zero-finding placeholder nor a synthetic parser is registered.

Commands, from the root after source /workspace/adamic-tools/env.sh:

```sh
python3 stage1/cohere/lint/rules/typescript-no-restricted-types/prepare-validation.py /tmp/wave02-font-retry
GOFLAGS=-overlay=/tmp/wave02-font-retry/overlay.json go test ./stage1/cohere/lint -run '^TestRestrictedTypesParity$' -count=1 -v -timeout 10m > /tmp/wave02-restricted-parity.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-02-typescript GOFLAGS=-overlay=/tmp/wave02-fourth/overlay.json go test ./stage1/cohere/lint -run '^(TestRestrictedTypesCompiler|TestRestrictedTypesMutant|TestRestrictedTypesThroughput|TestThirdWitnesses)$' -count=1 -v -timeout 15m > /tmp/wave02-restricted-final.log 2>&1
GOFLAGS=-overlay=/tmp/wave02-fourth-font-fixed/overlay.json go test ./stage1/cohere/lint -run '^(TestFontPolicyParity|TestRestrictedTypesThroughput)$' -count=1 -v -timeout 10m > /tmp/wave02-font-policy-final.log 2>&1
GOFLAGS=-overlay=/tmp/wave02-font-retry/overlay.json go test ./stage1/cohere/lint -run '^TestFontPolicyParity$' -count=1 -v -timeout 10m > /tmp/wave02-font-policy-retry.log 2>&1
```

Parity passed 24.730s. The combined compiler/mutant/witness command ran those three tests successfully (32.22s, 15.73s, 16.40s) but then failed registration for a temporarily unregistered font directory. The directory was moved under this owned rule. Throughput then passed 5.94s; font adapter compilation initially failed because its Go build path was relative to the wrong directory. Correcting the private path made the font comparison and mutant pass 8.791s. Raw failures and final evidence are preserved, not counted as green runs. No shared generator or test harness was edited.

One run on 1,000 repeated owned witnesses, 5,000 findings: native 63,279.18 findings/s (79.014931ms), Node 29,427.94 (169.906547ms), Go 141,106.39 (35.434256ms). This is synthetic rule workload, not compiler speed. Native correctness runs use ASan/UBSan; timings use optimized native. Setup with the prior owned overlay passed: go/clang/Node/submodules 0s, cache warm 16s, done 16s, nproc 5, cgroup four CPUs. The normal unintegrated harness is not certified. Exhaustive malformed configuration decoding and whitespace-normalized duplicate key ordering are outside this evidence.

## Selected-rule option isolation

The later grouped-accessor-pairs raw positional options exposed an owned
constructor bug: a disabled restricted-types rule still decoded another rule's
configuration. Its constructor now loads bans only when its context enables
this rule. No shared factory or harness changed. The raw options comparison
initially failed explicitly with exit 70, then passed after this correction.
The original 52 restricted-type upstream cases plus four configured witnesses
still match Go on Node, emitted JavaScript and sanitized native: 16,565 bytes,
PASS in 27.87s. The combined raw-options/restricted regression command passed
in 50.446s. Log: evidence/option-isolation.log. Raw positional and captured
exception-map corner cases are also retained in the owned id-length validator.
