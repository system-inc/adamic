# Tailwind utility arm constructors

Three separate .a files port the original Go colorArm, themeArm and widthArm constructors. Import their named exports with the matching camel-case function name. UtilityArm in model.a carries every upstream field.

colorArm(keys, keysPresent) preserves the caller's array identity, namespace order and duplicates. keysPresent distinguishes Go nil from an empty non-nil slice: when false, pass an empty placeholder array. themeArm(key) and widthArm(key, suffix, refusesModifier) each allocate a fresh single-key array. widthArm sets BareValuePositiveInteger and preserves the exact suffix and modifier flag; colorArm alone sets isColor. All other fields retain Go's zero values, including an absent InferTypes slice. inferTypesPresent=false means its empty placeholder array represents absence.

No resolution, inference, CSS emission or bare-value handler is implemented by these constructors. A consumer must supply those separate helpers before it can resolve a utility. The type carries enum spellings as upstream values, without a regex or a handwritten matcher.

The owned harness overlays thin exports into pinned Go cohere without changing any original function. It captures every string literal in every inventory consumer test file and uses those as constructor arguments alongside namespace, nil/empty, duplicate, Unicode, NUL, newline, suffix and flag controls. These are helper parameter projections, not full rule findings or execution of the consumer rules. The comparisons cover all fields, nil flags, array lengths/order, color aliases and fresh single-key arrays.

Run with the setup environment sourced:

```
ADAMIC_GATE_UNCACHED=1 go test ./stage1/cohere/lint/helpers/slot05/batch30 -count=1 -v -timeout=20m > /tmp/lint05-batch30-helpers.log 2>&1
```

Set ADAMIC_SLOT05_BATCH30_EVIDENCE to an existing directory to retain raw generated corpus/verdict/coverage files. Committed evidence is compressed losslessly. See REPORT.md for observed counts and limits.
