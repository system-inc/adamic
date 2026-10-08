# Next.js helper package and no-document-import-in-page

Claim: `lint-helpers/ecmascript-nextjs`, first pushed at `3880100c`; consuming blockers added at `6321c851`. Base: `origin/area/stage1-lint` (`8236e199`). Triage: `d7ab0bc4`, rank 9. The earlier eligible packages `rules/react` and `ecmascript/text` were already claimed. The post-claim fetch found no competing nextjs package claim.

The package covers all nine nextjs dependencies in the triage ledger. Every helper is independently exported in its own file and composes with the others without supplied Go facts or callbacks.

| Go symbol | Adamic symbol | File relative to `stage1/cohere/lint/helpers/nextjs/` |
|---|---|---|
| IsDocumentFile | isDocumentFile | is_document_file.a |
| lastSeparator | lastSeparator | last_separator.a |
| splitPath | splitPath | split_path.a |
| IsDocumentPage | isDocumentPage | is_document_page.a |
| IsInApplicationDirectory | isInApplicationDirectory | is_in_application_directory.a |
| IsInPagesDirectory | isInPagesDirectory | is_in_pages_directory.a |
| splitSegments | splitSegments | split_segments.a |
| buildRouteFileContracts | buildRouteFileContracts | build_route_file_contracts.a |
| RouteContractExports | routeContractExports | route_contract_exports.a |

`IsDocumentFile`, `lastSeparator`, and `splitPath` reuse the decisions and separator algorithm from `origin/codex/lint-helpers-03`, respectively `slot03/batch20/is_document_file.a`, `slot03/batch20/last_separator.a`, and `slot03/batch21/split_path.a`. Their injected dependency callbacks are replaced with direct imports. The last-separator result remains a UTF-8 byte offset; splitPath converts that offset to the runtime's string slicing coordinate. Capture inputs are valid Unicode paths. The partial byte-array interface's arbitrary invalid UTF-8 domain is not claimed for this string interface.

Go oracle pin: `cohere` `7945d102a6c18dd36adf9114a758ce646e8b2359`. `nextjs/testdata/capture.py` overlays wrappers around unchanged Go helper bodies and records every call from the ten current consuming rules' upstream suites. The frozen ledger names eight; a source search adds both Nexus route-contract consumers. Package tests and explicit path controls supplement the real consuming-rule calls. `counts.json` separates consumer calls, all calls, and deduplicated asserted source/path/options combinations. Cases contain paths and Go outputs, not substituted Go callback answers.

The helper driver compares canonical booleans, byte offsets, UTF-8 path bytes, segments, and sorted map key sets byte for byte. Contracts are exposed as name arrays, with undefined distinguishing Go's nil map; the Go caller contract forbids changing returned sets. The final capture contains 572 unique upstream cases, 1,274 consumer helper calls, and 1,832 total helper calls. The dedicated helper test checks Node source, emitted JavaScript, ASan/UBSan native, and one compiling semantic mutant per helper on all three backends.

The registered rule uses `IsDocumentPage`, scans only static import declarations, compares the parsed module specifier with `next/document`, and reports the complete declaration. It retains the unchanged Go rule through its own oracle adapter, exact message, witness, and module-equality mutant. Upstream coverage: 34 unique asserted cases; selected and all-rule modes yield 68 rows and 28,078 identical bytes on Go, Node, emitted JavaScript, and sanitized native (see `evidence/upstream.log`).

The package alone removes every ledger helper blocker for `@next/next/no-document-import-in-page` and `@next/next/no-head-import-in-document`. With the other packages in the landing sequence it also supplies blockers for no-before-interactive-script-outside-document, no-head-element, no-page-custom-font, no-styled-jsx-in-document, no-typos, and structure/next-no-near-miss-route-export. Their other blockers remain as listed in the claim. No additional unported helper package or compiler gap stops this package.

Validation commands (source `/workspace/adamic-tools/env.sh`):

```
go test ./stage1/cohere/lint/helpers -run TestNextjs -count=1 -v -timeout=20m
ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-corpus go test ./stage1/cohere/lint/helpers ./stage1/cohere/lint -count=1 -v -timeout=60m
```

The compiler input checkout is pinned to `050880ce59e30b356b686bd3144efe24f875ebc8` (TypeScript 6.0.3). Full lint validation includes generated/owned witnesses, upstream selected/all-rule rows, compiler and stage1 sources, shards, and compiling rule mutants. Performance/profile tests remain opt-in; they do not add a distinct parity input set. Native lint builds use the harness's checker archive. The helper driver uses a supported explicit lexical sort comparator.

Post-merge certification: merged `origin/area/stage1-lint` at `cd56db1d` (including `f0ccab34`) as merge `49e12ac2`. The full lint package passes in 1800.424 seconds. Discovery checks all 233 captured JSX sources without a shared inventory edit. Upstream parity covers 4,610 unique cases; compiler/stage1 parity covers 927 files and 30,523,048 identical bytes; shards cover 4,785 rows and 34,069,244 identical bytes. All 84 registered rule mutants pass on Node, emitted JavaScript, and sanitized native. The focused nine-helper recertification passes in 48.876 seconds with the same 572 upstream cases and 1,832 calls.

See `evidence/merged-lint-summary.md`, the lossless `merged-lint.log.gz`, and `merged-helpers.log`. The earlier attempt and its successful targeted continuations remain recorded; its fixed JSX-inventory failure is resolved by upstream discovery. No count patch or owned shared-file edit is included.

The requested no-head-import-in-document follow-up is not free: `origin/codex/stage1-nextjs-lint` still claims it, although its status says it is unimplemented. `evidence/next-head-availability.md` records that check and the original claim. No second rule directory or branch has been taken pending release of that ownership.
