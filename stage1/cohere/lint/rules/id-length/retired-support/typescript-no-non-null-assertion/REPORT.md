Built: three `.a` rule candidates with exact findings and complete suggestion objects; regular-driver integration remains blocked.\
Commits: claim `cd4a6ec6`; assertion `1d548728`; optional-chain assertion `1e88c391`; this aliases `a55ed6c9`.\
Checks: 145 cohere cases, owned witnesses and 208 compiler/stage1 files (624 rule/file pairs, 38,294,173 identical bytes) on source Node, emitted JavaScript and sanitized native.\
Mutants: preceding-byte removal, omitted suggestion edits and omitted aliases all compile, exit cleanly and fail only output comparison on all three runtimes.\
Uncovered: shared `.a` registration and full suggestion reporting, exhaustive malformed-option diagnostics, synthesized ASTs, complete repository runtime gate and earlier batch gaps.

This batch is `@typescript-eslint/no-non-null-asserted-optional-chain`, `@typescript-eslint/no-non-null-assertion`, and `@typescript-eslint/no-this-alias`. Existing work was pushed before fetching all origin heads. Main observed: `ef3d907ecdc4c771b016f7d9c52372def057a340`. The published 46 helper-ready entries were already ported on main or recorded in origin claims. Selection then used the inventory's ordered syntax-only entries, excluding type-information and binding-only rules. It examined 320 origin refs and 39 Markdown files under their claims directories, matching complete public names with punctuation boundaries. `cd4a6ec6` reserved these three and was pushed before code. No subsequent rules are claimed.

Ahra's correction arrived after the comparisons and mutants were running. All implementation, models, adapters, witnesses, validation artifacts, logs and this report are inside the three owned rule directories. The claim update preceded the correction. No shared repository generator, harness, finding, settings, traversal, parser, lowerer or native file was edited. The existing `.a` files remain `.a`; the correction permits temporary `.ts` but does not require rewriting validated `.a` modules. Shared `.a`/emitted-JavaScript harness work belongs to `codex/lint-harness-dot-a`.

The separate blocker is the regular finding/reporting contract. Its scalar Finding has one repair at the diagnostic's range. Go's optional-chain assertion suggests removing only the final operator; the general assertion can suggest two distinct edits, with preserved trivia between them. Those edits cannot be represented faithfully by that scalar contract. The owned SuggestedFinding stores complete typed suggestion and fix lists without changing the base Finding. The owned detailed driver emits every suggestion ID, message, edit range and replacement text. For witnesses and upstream cases, it also applies each suggestion independently and emits the resulting bytes. The Go output adapter does the same from the unmodified Go rule diagnostics and their actual Fix objects. It does not wrap or alter a Go rule, replace its fixes, or normalize its semantic result. The primary formatter and automatic-fix engine are Go cohere's actual implementations.

The ordinary driver explicitly refuses selection of either assertion rule without the detailed reporter, rather than dropping suggestion metadata. TestThirdReportingRefusal proves that source Node, emitted JavaScript and ASan/UBSan native all exit 70 with `non-null rules require a driver that reports complete suggestion edits`. This remains an integration blocker even after `.a` discovery lands. This-alias has no suggestion-model dependency in its implementation and only needs shared `.a` registration support. Per Ahra's instruction, this worker stops here and does not edit shared files to remove either blocker.

The scratch overlay reads shared Go sources and patches temporary copies only. It enables `.a` discovery/copying and emitted-JavaScript comparison, selects the owned detailed driver and independent Go output adapter, preserves upstream filename extensions, and adds the actual absent-optional-node test family to capture. Its checked-in preparation script reproduces the overlay. These are private validation artifacts, not applied shared changes or a proposed shared dispatch list.

Observed comparisons:

- TestThirdWitnesses: PASS, 15,111 identical bytes, 14.53s; replay using the checked-in preparation script PASS, 14.642s. The rows include selected and all-rule executions on owned witnesses. A Unicode witness tests byte offsets and Go's raw dot search into an intervening comment.
- TestThirdUpstream: PASS, 145 unique cases for the three rules, 103,034 identical bytes, 27.30s. Capture collected 524 cases across the registered descriptors before selecting the three new rules. Original rule tests and the absent-optional-node tests ran. No selected upstream case was excluded.
- TestThirdCompiler: PASS, TypeScript v6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`, every compiler `.ts` file and every stage1 `.ts`/`.a` file on this branch. 208 files, 624 selected rule/file pairs, 38,294,173 identical bytes, 84.528s. No source was excluded. Complete suggestion edits are compared; printing the entire rewritten compiler file for every suggestion is avoided. All three rules have no automatic fixes, and unchanged automatic-fix output is checked against Go's actual converging fix engine.
- TestThirdMutants: PASS, 41.64s. Each variant compiles and runs with exit 0 and empty stderr on Node, emitted JavaScript and ASan/UBSan native. The assertion-operator mutant changes `edit 56 57` to `edit 55 56`; the general-assertion mutant removes the suggestion and its edits; the alias mutant removes identifier findings. Only byte comparison catches them.
- TestThirdReportingRefusal: PASS, 13.38s, matched explicit exit-70 refusal on all three runtimes.
- Registry package: PASS, deterministic regeneration and 11 malformed-descriptor rejection mutants. Filtered input oracle: PASS, six input programs, 0.175s, six probe-cache hits. This worker gate used the permitted cache; it is not an uncached integration gate.

The first owned witness run caught a missed compound assignment because the parser stores BinaryExpression as left/operator/right child indexes rather than an operator field. The candidate was corrected to read the token child and the right-hand child. Type assertions use type/expression child indexes and are unwrapped accordingly. The initial failure remains in evidence/witnesses.log; passing results are in evidence/baseline.log and evidence/reproduce.log. Observed Go behavior also searches for the first raw dot after an assertion, including dots in comments; the Unicode witness and candidate deliberately preserve that behavior. This report does not claim it is a desirable rewrite.

Toolchain setup used the private `.a` overlay and succeeded: Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s, build cache warm 18s, done 18s. `nproc` printed 5; cgroup quota is 4 CPUs. Versions: Go 1.27.1, clang 20.1.8, Node 24.19.0. Environment: `/workspace/adamic-tools/env.sh`. Setup warms compilation only and does not constitute a runtime test gate.

One end-to-end throughput run per cell, including process startup, file reads and parsing. Native throughput uses an optimized build; parity and mutants use sanitizers.

| Rule on 208 compiler/stage1 files | Findings | Native findings/s | Node findings/s | Go findings/s |
| --- | ---: | ---: | ---: | ---: |
| no-non-null-asserted-optional-chain | 7 | 5.01 | 8.08 | 33.06 |
| no-non-null-assertion | 1,123 | 908.89 | 1,241.36 | 5,305.18 |
| no-this-alias | 0 | 0.00 | 0.00 | 0.00 |

| Rule on 1,000 repeated owned witnesses | Findings | Native findings/s | Node findings/s | Go findings/s |
| --- | ---: | ---: | ---: | ---: |
| no-non-null-asserted-optional-chain | 1,000 | 27,038.98 | 7,055.73 | 45,595.32 |
| no-non-null-assertion | 1,000 | 30,770.70 | 6,995.67 | 49,328.28 |
| no-this-alias | 2,000 | 61,554.69 | 13,766.53 | 82,489.31 |

The alias source corpus pass took native 1.276s, Node 1.055s and Go 0.206s. Zero findings do not establish positive findings throughput. Repeated-witness measurements are synthetic and are kept separate from real-source measurements. They are not extrapolations or stable multi-run benchmarks.

Reproduction from the repository root (every test writes directly to a log):

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/lint/rules/typescript-no-non-null-assertion/prepare-validation.py /tmp/wave02-third
export GOFLAGS=-overlay=/tmp/wave02-third/overlay.json
export ADAMIC_TYPESCRIPT_SOURCE=/path/to/TypeScript
# The compiler checkout must be 050880ce59e30b356b686bd3144efe24f875ebc8.
ADAMIC_TOOLS=/workspace/adamic-tools bash cloud/setup.sh > /tmp/third-setup.log 2>&1
go test ./stage1/cohere/lint -run '^(TestThirdWitnesses|TestThirdUpstream)$' -count=1 -v -timeout 15m > /tmp/third-baseline.log 2>&1
go test ./stage1/cohere/lint -run '^TestThirdCompiler$' -count=1 -v -timeout 20m > /tmp/third-compiler.log 2>&1
go test ./stage1/cohere/lint -run '^(TestThirdMutants|TestThirdThroughput)$' -count=1 -v -timeout 15m > /tmp/third-mutants-throughput.log 2>&1
go test ./stage1/cohere/lint -run '^(TestThirdReportingRefusal|TestThirdSourceThroughput)$' -count=1 -v -timeout 15m > /tmp/third-refusal-source-bench.log 2>&1
go test ./stage1/cohere/lint/registry -count=1 -v > /tmp/third-registry.log 2>&1
go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout 10m > /tmp/third-oracle.log 2>&1
```

All actual output logs are in evidence/. No full repository runtime gate, exhaustive arbitrary/malformed configuration comparison, exact configuration-error prose, synthesized malformed AST proof or earlier-batch gap resolution is claimed. Captured this-alias options are the actual Go decoded values; its adapter also supports the real upstream wire decoder. Ordinary shared-driver integration is explicitly blocked as described above. No pull request is opened and no additional rule is reserved.
