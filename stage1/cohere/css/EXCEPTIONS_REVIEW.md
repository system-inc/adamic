# CSS exception review

All 178 recorded occurrences are accounted for: **(a) 162 deliberate Go departures, (b) 16 candidate Go bugs, (c) 0 port bugs, (d) 0 unexplained**. There are six causes. These counts include 176 printer occurrences and two raw PostCSS surrogate occurrences; the printer JSON has 148 records because identical inputs retain multiple fixture occurrences.

This review changes no implementation, gate, exclusion, or exception. Its base is devtools/gate-inputs-css `bd575e8b590c2862ad3cafdc1c4e25fa7216aca3`. The records are in `cloud/reports/css-gate-inputs/recorded-printer-discrepancies.json` and `recorded-postcss-discrepancies.json` at that base. Evidence below uses Go cohere `7945d102a6c18dd36adf9114a758ce646e8b2359`, fixtures from system-inc/prettier `cb4b33fba24a8428d00e54be85fc886288a374ea`, the cohere fork bundles at the Go pin, and npm Prettier 3.9.6. Raw library comparisons use postcss 8.5.16 and postcss-scss 4.0.9.

## Causes

Source links in this table pin cohere to the requested revision. An implementation that happens to differ is insufficient evidence of intent: class (a) requires an explicit policy or entry-point contract. Class (b) identifies an owner decision or bug candidate, not a demonstrated public CLI failure.

| Cause | Class | Occurrences | Files / cases | Evidence |
| --- | --- | ---: | --- | --- |
| Nonempty YAML front matter is explicitly refused | (a) | 112 | Eleven fixture files listed below, 8 each; three generated inputs, 8 each | [format.go:180–184](https://github.com/system-inc/cohere/blob/7945d102a6c18dd36adf9114a758ce646e8b2359/internal/format/css/format.go#L180) explains that Format has no `textToDoc` to reach the YAML printer, and that silently dropping a failing embed would produce something different from the fork: **“So it refuses instead.”** The port reproduces that refusal. This is an intentional unsupported embedding boundary, including `yaml.less`. |
| Ordinary leading BOM removal at the direct formatter entry point | (a) | 40 | Generated BOM-only, BOM plus `a{}`, and BOM plus `a{b:c}` inputs; no fixture files | [format.go:113–114](https://github.com/system-inc/cohere/blob/7945d102a6c18dd36adf9114a758ce646e8b2359/internal/format/css/format.go#L113) delegates BOM and line-ending normalization to the shared layer; [parser.go:40–42](https://github.com/system-inc/cohere/blob/7945d102a6c18dd36adf9114a758ce646e8b2359/internal/format/css/parser.go#L40) requires an already removed BOM. [postcss/input.go:59–64](https://github.com/system-inc/cohere/blob/7945d102a6c18dd36adf9114a758ce646e8b2359/internal/format/css/postcss/input.go#L59) removes a leading BOM in the parser input. The direct internal formatter omits the BOM where the full Prettier API preserves it. This is defensible as an entry-point boundary, not as full-API equality on unnormalized input. |
| BOM-relative source slicing corrupts comments | (b) | 8 | `tests/format/css/bom/bom.css`, CSS and SCSS modes, both widths and both Prettier variants | Removing the BOM shifts parser offsets, while [format.go:186–190](https://github.com/system-inc/cohere/blob/7945d102a6c18dd36adf9114a758ce646e8b2359/internal/format/css/format.go#L186) keeps the unstripped argument as `OriginalText`. [print.go:67–76](https://github.com/system-inc/cohere/blob/7945d102a6c18dd36adf9114a758ce646e8b2359/internal/format/css/print.go#L67) slices comments from that text, using [print_helpers.go:104–112](https://github.com/system-inc/cohere/blob/7945d102a6c18dd36adf9114a758ce646e8b2359/internal/format/css/print_helpers.go#L104). This can truncate comments and emit replacement characters. No source comment prescribes corruption. The reduced witnesses below separate this effect from intentional BOM removal. The call violates the documented normalized-input precondition, so the owner must decide whether to reject it, normalize consistently, or enforce normalization in the caller. |
| Carriage returns reach the internal printer without shared normalization | (a) | 8 | Generated CSS input `// x\ra{}`, two corpus indices, both widths and both variants | The same explicit [shared-normalization contract](https://github.com/system-inc/cohere/blob/7945d102a6c18dd36adf9114a758ce646e8b2359/internal/format/css/format.go#L113) and [parser precondition](https://github.com/system-inc/cohere/blob/7945d102a6c18dd36adf9114a758ce646e8b2359/internal/format/css/parser.go#L40) apply. [format_test.go:629–631](https://github.com/system-inc/cohere/blob/7945d102a6c18dd36adf9114a758ce646e8b2359/internal/format/css/format_test.go#L629) also normalizes CRLF before calling FormatSCSS. Go and the port preserve the recorded CR where the full Prettier API normalizes it. |
| Missing whitespace-only NBSP fast path | (b) | 8 | Generated single-NBSP input, CSS and SCSS modes, both widths and both variants | [format.go:171–178](https://github.com/system-inc/cohere/blob/7945d102a6c18dd36adf9114a758ce646e8b2359/internal/format/css/format.go#L171) proceeds to parsing; [postcss/tokenize.go:184–190](https://github.com/system-inc/cohere/blob/7945d102a6c18dd36adf9114a758ce646e8b2359/internal/format/css/postcss/tokenize.go#L184) recognizes ASCII whitespace. Full Prettier returns an empty result for this whitespace-only input; Go and the port return Unknown word. No stated policy explains omitting the full API shortcut. The single-code-point witness below is minimal. |
| UTF-16 surrogate cut is deliberately represented at a UTF-8 character boundary | (a) | 2 | Generated `\\😀\u007ca`, raw CSS and SCSS, indices 3114 and 3115 | [postcss/input.go:47–49](https://github.com/system-inc/cohere/blob/7945d102a6c18dd36adf9114a758ce646e8b2359/internal/format/css/postcss/input.go#L47) explicitly maps the second UTF-16 unit to the byte position after the character, keeping the whole character on the left so pieces join. [input.go:70–74](https://github.com/system-inc/cohere/blob/7945d102a6c18dd36adf9114a758ce646e8b2359/internal/format/css/postcss/input.go#L70) implements it; [tokenize.go:307–335](https://github.com/system-inc/cohere/blob/7945d102a6c18dd36adf9114a758ce646e8b2359/internal/format/css/postcss/tokenize.go#L307) and [parser.go:710–711](https://github.com/system-inc/cohere/blob/7945d102a6c18dd36adf9114a758ce646e8b2359/internal/format/css/postcss/parser.go#L710) propagate it into Unknown word and its end location. PostCSS retains only the high surrogate in that token. The port matches Go's documented representation policy; this specific diagnostic boundary is not literal PostCSS parity. |

## Fixture accounting

The YAML cause has eight occurrences for each of these files (CSS and SCSS modes, default and narrow options, npm and fork):

- `tests/format/css/front-matter/embedded-language-formatting/yaml.css`
- `tests/format/css/yaml/comment_after.css`
- `tests/format/css/yaml/dirty.css`
- `tests/format/css/yaml/ignore.css`
- `tests/format/css/yaml/malformed-2.css`
- `tests/format/css/yaml/only_comments.css`
- `tests/format/css/yaml/with_comments.css`
- `tests/format/css/yaml/without-newline-after.css`
- `tests/format/css/yaml/yaml.css`
- `tests/format/scss/yaml/yaml.scss`
- `tests/format/less/yaml/yaml.less`

The last two files have identical inputs, but both occurrences count. Generated YAML cases contribute the remaining 24. All generated inputs are attributed in the records to the corpus in `stage1/cohere/css/testdata/cohere_side_test.go`.

The fixture census is 158 CSS, 90 SCSS and 43 Less files, including one repository CSS file. Extension coverage is not parser coverage: the recorded printer transports invoke CSS or SCSS mode even for `.less`. [parser.go:6–8](https://github.com/system-inc/cohere/blob/7945d102a6c18dd36adf9114a758ce646e8b2359/internal/format/css/parser.go#L6) explicitly says parseLess is not ported. In particular, `yaml.less` proves the YAML refusal on those two modes, not Less conformance.

For an exhaustive mapping back to the printer JSON, number records from one. There are four ordered blocks of 37: default/npm (1–37), default/fork (38–74), narrow/npm (75–111), narrow/fork (112–148). Apply these local positions to each block:

| Cause | Local record positions | Weighted occurrences per block |
| --- | --- | ---: |
| YAML refusal | 3–8, 18–37 | 28 |
| Ordinary BOM | 1, 2, 9–12 | 10 |
| BOM comment corruption | 16, 17 | 2 |
| CR normalization boundary | 13 | 2 |
| NBSP shortcut | 14, 15 | 2 |

These disjoint positions cover all 37 records and 44 occurrences per block: 148 records and 176 occurrences overall. The raw PostCSS record contains the two additional surrogate occurrences. Nothing is excluded or left unexplained.

## Reduced proving inputs

Strings below use JSON escaping; a quoted output denotes literal formatted text, not the transport encoding. Each result was replayed against pinned Go, the port on Node, and a fresh sanitized native build. Both actual Prettier variants agree on the printer witnesses.

| Mode | Input | Go and port | Prettier |
| --- | --- | --- | --- |
| CSS | `"\uFEFF/**/"` | `"\uFEFF/\n"` | `"\uFEFF/**/\n"` |
| SCSS | `"\uFEFF//"` | `"\uFFFD\uFFFD\n"` | `"\uFEFF//\n"` |
| CSS or SCSS | `"\u00A0"` | CssSyntaxError: `Unknown word \u00A0`; line 1, column 1, offset 0; end line 1, column 2, byte offset 2 | `""` |

The CSS witness uses the shortest valid block comment with a leading BOM; the SCSS witness uses its two-character line-comment marker. Without the BOM these comments format normally. These show content corruption, which is why the eight fixture occurrences are separate from the 40 ordinary BOM omissions. The NBSP witness is one code point; the empty input has no discrepancy.

For the raw parser, reducing `"\\😀\u007ca"` to `"\\😀"` retains the discrepancy in both modes. Go and the port report reason `Unknown word \\😀`, end column 4, canonical byte end offset 5. PostCSS reports `Unknown word \\` followed by the unpaired high surrogate `\uD83D`, end column 3, canonical byte end offset 4. Both start at line 1, column 1, offset 0, and end on line 1. The canonical adapter computes byte offsets from JavaScript prefixes; its offset 4 includes the replacement encoding of the isolated surrogate, not a Go UTF-8 character boundary. This is the documented cut policy's diagnostic consequence.

## Validation and limits

Setup ran with `GOPROXY='https://proxy.golang.org|direct'` and `bash cloud/setup.sh --gate-inputs`: successful in 454.935 seconds, `nproc` 5. Fixture installation verified the pinned external counts (157 CSS, 90 SCSS, 43 Less). Go was 1.27.1, Node 24.19.0, clang 20.1.8.

The checked-in gate records originally used cohere 715ba94f. I built a scratch worktree of requested cohere 7945d102 with its exact TypeScript dependency and replayed every recorded input. The production paths responsible for these exceptions did not change between those revisions; empirical replay, rather than that observation alone, establishes that all recorded answers still hold.

Validation commands used the existing printer adapter through a scratch Go overlay (`go test -overlay=... -run '^TestAdamicPrinterCases$' -count=1 -v ./internal/format/css`), the existing `print_library.mjs` against npm and pinned fork bundles, and `oracle/node.mjs` against `print_main.ts`. All **148 records / 176 occurrences** reproduced their recorded Go and Prettier answers exactly in both option modes; the port source matched pinned Go throughout. Fresh `go run ./cmd/adamic build stage1/cohere/css/print_main.ts -o ... --sanitize` and native runs matched Go byte for byte on every recorded printer input in both modes. The reduced printer witnesses also matched Go natively.

A scratch raw-parser Go test, existing `testdata/library.mjs`, and `oracle/node.mjs` against `main.ts` reproduced both original surrogate occurrences and both reduced dialect cases. A fresh sanitized build of `main.ts` matched pinned Go on all four. Native runs kept `ASAN_OPTIONS=detect_leaks=1`; sanitizer stderr was empty. Test and build output remained in local log files.

This is a review of the complete recorded exception set and additional reduced witnesses. It does not claim a fresh replay of the whole 24,076-case gate or arbitrary CSS/SCSS/Less conformance. Class (a) API-boundary cases are defensible only with their named entry-point assumptions. Class (b) cases should go to the Go cohere owner: comment corruption needs a coherent normalization/source-offset policy, and NBSP needs a decision about reproducing the full Prettier whitespace-only shortcut. No port bug was observed in this set, and no check was relaxed to reach that conclusion.
