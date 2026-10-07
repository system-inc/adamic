Built hexValue, UnescapeStringLiteralText and parameterNodes in separate .a files: 25 dependency edges, 16 distinct consumers.  
Commits: c584145 and be16b65 implement text helpers; 38079aa claims parameterNodes; final parameter implementation and evidence are committed after the gate.  
Commands: complete slot-package gate PASS 152.213s (2,440,871 comparison lines); vet/gofmt clean; six uncached input-oracle fixtures PASS 1.516s; regeneration identical.  
Mutants: uppercase F removed, incorrect entity substitution, copied parameter list and ignored nil-list guard; compiling native variants caught.  
Not covered: full repository gate, complete rule findings/fixes/suggestions, arbitrary invalid UTF-8 strings and decoder implementation.  

All earlier six helpers were tested and pushed before this batch began. Every claim was published before implementation. Selection checked every origin codex/lint-helpers* branch. The initial attributeValues reservation collided with earlier slot 01 ad07122 (00:59:17 UTC), preceding ours 5142004 (00:59:39 UTC); our passing duplicate was removed and is not counted or delivered.

hexValue was claimed in 2df34e5 after every 11- and 10-consumer dependency became reserved. Its consumer count is nine. UnescapeStringLiteralText was claimed in c584145 at 01:03:44 UTC, preceding slot 04 502d072 at 01:03:48 UTC, so slot 03 retains it. decodeEntity belongs to slot 01 and remains an explicit callback. MatchExactly was claimed in the API-published be16b65, but slot 05 470ddd3 at 01:09:43 UTC precedes be16b65 at 01:11:33 UTC. The passing duplicate was removed and is not counted. After API claim refresh and a recovered normal Git fetch, parameterNodes was the sole highest-count unclaimed named helper at seven and was claimed in 38079aa before code.

## Publication recovery

Shell Git fetch began failing with fatal: could not read Username for https://github.com; retries, including an anonymous public-repository fetch, also failed. The connected GitHub API still reports push permission and successfully reads the same repository. We refreshed all six branch claims directories and every present claims file through that API. The shared branch has no claims directory (404), and its comments reservation in HELPERS.md remains excluded.

Publication uses immutable blobs, a tree built on the current parent, a new commit and a non-force branch update with expected SHA. Every uploaded blob and complete tree SHA is checked against the local Git index. Remote commit headers are reconstructed and checked against the actual commit SHA before installation; the local branch advances only from its exact parent. No history is rewritten and no PR is opened. Shell Git fetch/push subsequently recovered, so parameterNodes and final evidence use ordinary fast-forward Git commits and pushes.

## Observation scope

The capture records 694 actual runtime source inputs across all 16 consumers. Upstream next/structure capture-package tests PASS (capture_gate_exit=0). The helper oracle parses those sources with the pinned typescript-go parser, reads JSX text from AsJsxText().Text and normalizes paths with tspath.NormalizePath. It retains decoded identifiers, string literals and template text, plus complete source strings for scanning controls.

Hex conversion compares all Unicode code points plus rune extremes. Text scanning exercises every named XHTML entity from Go, decimal/hex references, numeric overflow, surrogate replacement, unknown/invalid/unfinished entities, nested ampersands, one-pass replacement, controls and supplementary Unicode. Parameter lists exercise all function-like lists found in each consuming rule fixture, with nil/present-empty, duplicate-node, nil-node and backing-storage mutation controls. Node indices encode identity; fixed-length arrays encode Go slices for these read/range consumers, with structural growth outside the adapter contract. Decoder calls and results are independently observed in Go through an oracle-only overlay; the scanner does not reimplement the separately owned entity decoder.

The first unescape oracle attempts failed because Node.Text does not handle JsxText and parser filenames must normalize Windows separators. The adapter was corrected to use the actual dedicated field and path API. A separate emitted-JavaScript attempt lacked the adamic runtime loader; it now uses the existing oracle/node.mjs runner. None of these attempts is counted as a pass. No shared harness or compiler file was edited.

Each mutant compiles and exits 0 without stderr before its semantic output is compared with Go. Native sanitizers remain enabled. Initial parameter replacement test: PASS 28.655s, 1,117,147 lines across all four executions; an additional explicit duplicate/nil-node identity control is included in the final gate. Hex and unescape witnesses remain line 43 (-1 versus 15) and 1114212 (literal copy body versus copyright character). Parameter mutants change shared-write observations or return nodes for an absent list. The final witness lines are recorded below.

Readiness removes 25 dependencies from 16 distinct rules, with zero final helper blockers removed by this batch alone. Its residual ledger subtracts only this slot's nine delivered symbols and does not assume other branches have merged. This is helper readiness, not rule implementation.

## Exact consumers

Both text.hexValue and text.UnescapeStringLiteralText serve these nine rules:

- @next/next/google-font-display
- @next/next/google-font-preconnect
- @next/next/next-script-for-ga
- @next/next/no-before-interactive-script-outside-document
- @next/next/no-css-tags
- @next/next/no-html-link-for-pages
- @next/next/no-page-custom-font
- @next/next/no-unwanted-polyfillio
- structure/boundary-no-project-theme-value

structure.parameterNodes serves these seven rules:

- structure/network-require-hook-options-parameter
- structure/network-require-hook-request-suffix
- structure/network-require-hook-variables-type
- structure/next-require-api-parameter-name
- structure/react-component-no-destructuring
- structure/react-component-require-properties-parameter
- structure/react-component-require-properties-type-suffix

## Commands and environment

All test output is written directly to evidence logs; no test output is piped. Source /workspace/adamic-tools/env.sh before commands. Original setup remains valid: go ready 0s, clang ready 1s, node ready 1s, submodules ready 1s, cache warm 106s, done 106s. nproc again printed 5. Cohere pin: 715ba94f3608a6500086b1076ce5cb7e51b836db.

- python3 stage1/cohere/lint/helpers/slot03/batch3/testdata/regenerate.py: evidence/regenerate.log and capture.log.
- go test ./stage1/cohere/lint/helpers/slot03 -count=1 -v -timeout=20m: evidence/final.log.
- go vet ./...: evidence/vet.log.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout=20m: evidence/oracle.log.
- gofmt -l cmd internal stage1/cohere/lint/helpers/slot03: evidence/gofmt.log.
- Repeated capture and SHA-256 equality: evidence/reproducibility.log.

## Final gate observations

The complete slot package passes in 152.213s. Baselines compare 1,115,085 original, 208,636 second-batch and 1,117,150 third-batch output lines: 2,440,871 total. All compare real Go, Node source and sanitized native; the third batch also compares emitted JavaScript through the existing Node runtime loader.

All eleven compiling native semantic mutants are caught: original component name (line 159), whitespace (847), listener kind (1114948), listener freshness (1114949); second-batch intrinsic kind (24), hole-boundary inheritance (179230), dispatcher route (185959); third-batch uppercase F (43), entity substitution (1114212), parameter shared storage (1115385), absent parameter list (1115375). The missing-consumer coverage mutant is also rejected. The parameter copy mutant's Go observation is -1,-1,3,2 versus the mutant's 2,-1,3,2, proving shared storage, duplicate identity and nil-node handling together.

Whole-repository go vet and the recorded gofmt scan produce empty logs. The uncached filtered input oracle passes all six fixtures in 1.516s, with zero cache hits and six probe misses. Repeated capture regenerates the compressed fixture and coverage manifest byte for byte. The full repository test gate and complete consuming-rule implementations were not run or delivered. External decoder ownership and fixed-length parameter-list adapter limits remain explicit above. No shared harness, registration or compiler file was changed.
