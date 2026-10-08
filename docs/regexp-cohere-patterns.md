# Cohere regex patterns held to Node

Recorded from Adamic `df959ee978da7cfcac45fe0e9ff941b80a6fda58`, cohere `715ba94f3608a6500086b1076ce5cb7e51b836db`, and cohere's TypeScript `8d550c837c90bd1805b047b7eeccc2baac2d5e7a`. TypeScript API 6.0.3 parsed 13425 files. There are 875 static occurrences (868 literals, 7 constant constructors), 17 unresolved constructors, and 7305 parser diagnostics in intentionally malformed and other source fixtures.

v24.19.0 is the external oracle. 238 patterns are rejected by Node; 44938 input strings produce 128592 recorded exec calls. There are 75 source/trace input records and 0 generator parser gaps. Adamic completed 128592 Go comparisons and 385776 native comparisons (bounded, unlimited automatic dispatch, and forced VM), with ASan, UBSan, and LeakSanitizer on native runs.

Extraction, fixture tracing, generation, regeneration commands, limitations, and mutants are described in [the corpus README](../internal/regexp/testdata/cohere/README.md). CI reads the Node observations without invoking Node. Pattern strings, flags, exact UTF-16 units, all occurrence locations, all input provenance, and every observation are in the inventory and compressed testdata. The test compares all cases and pins every observed refusal or disagreement exactly in `issues.json`; it never exempts an entire pattern. Updating this report requires the explicit `ADAMIC_COHERE_RECORD=1` mode.

## Node-valid gaps for the next unit


Scope: 77 implementation/script occurrences and 798 fixture/test occurrences. Of Node's 238 rejections, 141 are invalid literal source tokens whose recovered body/flags would be valid constructors. They do not establish regex compiler defects.

## Unresolved constructors

These are observations of expressions the constant evaluator could not reduce, not proof that all are inherently dynamic.

- `cohere/TypeScript/tools/scripts/tsc/generate-enums.ts:105`: ``new RegExp(`^(${prefix}\\w+)\\s+(?:\\S+\\s*)?=\\s*(.+)$`)`` (nonconstant pattern).
- `cohere/TypeScript/tools/scripts/tsc/generate-enums.ts:108`: ``new RegExp(`^(${prefix}\\w+)$`)`` (nonconstant pattern).
- `cohere/TypeScript/tools/scripts/tsc/generate-enums.ts:237`: ``new RegExp(`^${prefix}`)`` (nonconstant pattern).
- `cohere/TypeScript/tools/scripts/tsc/options.test.ts:176`: ``new RegExp(`Unsupported comparison type for ${name}:`)`` (nonconstant pattern).
- `cohere/TypeScript/tools/scripts/tsc/options.test.ts:295`: ``new RegExp(schema.pattern)`` (nonconstant pattern).
- `cohere/TypeScript/tsc/testdata/fixtures/compiler/commandLineParser.ts:4127`: ``new RegExp(rawExcludeRegex, useCaseSensitiveFileNames ? "" : "i")`` (nonconstant pattern).
- `cohere/TypeScript/tsc/testdata/fixtures/compiler/moduleSpecifiers.ts:138`: ``new RegExp(pattern)`` (nonconstant pattern).
- `cohere/TypeScript/tsc/testdata/fixtures/compiler/moduleSpecifiers.ts:143`: ``new RegExp(pattern)`` (nonconstant pattern).
- `cohere/TypeScript/tsc/testdata/fixtures/compiler/moduleSpecifiers.ts:148`: ``new RegExp(pattern)`` (nonconstant pattern).
- `cohere/TypeScript/tsc/testdata/fixtures/compiler/moduleSpecifiers.ts:154`: ``new RegExp(pattern, flags)`` (nonconstant pattern).
- `cohere/TypeScript/tsc/testdata/fixtures/compiler/parser.ts:10707`: ``new RegExp(`(\\s${name}\\s*=\\s*)(?:(?:'([^']*)')|(?:"([^"]*)"))`, "im")`` (nonconstant pattern).
- `cohere/TypeScript/tsc/testdata/fixtures/compiler/program.ts:5046`: ``new RegExp(directoryPath, "i")`` (nonconstant pattern).
- `cohere/TypeScript/tsc/testdata/fixtures/compiler/utilities.ts:9818`: ``new RegExp(pattern, useCaseSensitiveFileNames ? "" : "i")`` (nonconstant pattern).
- `cohere/TypeScript/tsc/testdata/tests/cases/compiler/duplicateLocalVariable1.ts:43`: ``new RegExp(testcase.errorMessageRegEx)`` (nonconstant pattern).
- `cohere/TypeScript/tsc/testdata/tests/cases/conformance/fixSignatureCaching.ts:326`: ``new RegExp(object[key], 'i')`` (nonconstant pattern).
- `cohere/TypeScript/tsc/testdata/tests/cases/conformance/fixSignatureCaching.ts:346`: ``new RegExp(value, 'i')`` (nonconstant pattern).
- `cohere/TypeScript/tsc/testdata/tests/cases/conformance/fixSignatureCaching.ts:930`: ``new RegExp(pattern, 'i')`` (nonconstant pattern).

## Refusals and disagreements

Each entry names an observed failure. The feature descriptions are triage inferences, not fixes. Agreed rejection of invalid fixtures is listed because the unit asks for every refusal. Wrong acceptance of a Node-invalid pattern is a compiler gap. Source-only rejections mean an invalid literal token was recovered by TypeScript, but its separated body/flags are valid; those are not regex compiler defects.

### Pattern 404, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/asyncJsxArgumentsAttributeName.ts:10:30`.

```text
pattern: ">"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 409, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/checkJsxNotSetError.ts:9:14`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 410, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/checkJsxNotSetError.ts:15:16`.

```text
pattern: ">"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 414, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/contentMapperPerFileExtension.ts:36:26`.

```text
pattern: ">"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 415, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/contextuallyTypedJsxAttribute.ts:17:1`.

```text
pattern: ">"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 416, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/contextuallyTypedJsxAttribute.ts:27:1`.

```text
pattern: ">"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 429, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/doubleUnderscoreReactNamespace.ts:13:23`.

```text
pattern: ">"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 430, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/excessiveStackDepthFlatArray.ts:38:15`.

```text
pattern: "li>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 432, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/excessiveStackDepthFlatArray.ts:42:6`.

```text
pattern: "ul>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 436, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsFileCompilationTypeArgumentSyntaxOfCall.ts:8:15`.

```text
pattern: "Foo>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 437, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsFileCompilationTypeArgumentSyntaxOfCall.ts:9:13`.

```text
pattern: ">"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 438, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsxEmitWithAttributes.ts:50:32`.

```text
pattern: "meta>,"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 439, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsxEmitWithAttributes.ts:51:28`.

```text
pattern: "meta>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 440, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsxFactoryAndReactNamespace.ts:51:32`.

```text
pattern: "meta>,"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 441, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsxFactoryAndReactNamespace.ts:52:28`.

```text
pattern: "meta>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 442, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsxFactoryIdentifier.ts:51:32`.

```text
pattern: "meta>,"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 443, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsxFactoryIdentifier.ts:52:28`.

```text
pattern: "meta>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 444, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsxFactoryIdentifierAsParameter.ts:17:21`.

```text
pattern: ">"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 445, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsxFactoryIdentifierWithAbsentParameter.ts:17:21`.

```text
pattern: ">"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 446, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsxFactoryMissingErrorInsideAClass.ts:10:22`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 447, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsxFactoryNotIdentifierOrQualifiedName.ts:50:32`.

```text
pattern: "meta>,"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 448, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsxFactoryNotIdentifierOrQualifiedName.ts:51:28`.

```text
pattern: "meta>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 449, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsxFactoryNotIdentifierOrQualifiedName2.ts:50:32`.

```text
pattern: "meta>,"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 450, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsxFactoryNotIdentifierOrQualifiedName2.ts:51:28`.

```text
pattern: "meta>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 451, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsxFactoryQualifiedName.ts:51:32`.

```text
pattern: "meta>,"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 452, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsxFactoryQualifiedName.ts:52:28`.

```text
pattern: "meta>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 453, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsxFactoryQualifiedNameResolutionError.ts:17:21`.

```text
pattern: ">"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 454, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsxFactoryQualifiedNameWithEs5.ts:16:26`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 455, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsxImportSourceCheckJs.ts:17:6`.

```text
pattern: ">"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 456, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsxPreserveWithJsInput.ts:11:25`.

```text
pattern: "b>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 457, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsxPreserveWithJsInput.ts:14:21`.

```text
pattern: "c>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 458, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsxPreserveWithJsInput.ts:20:23`.

```text
pattern: "e>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 459, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsxRuntimePragma.ts:8:49`.

```text
pattern: "h1>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 461, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsxRuntimePragma.ts:14:49`.

```text
pattern: "h1>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 463, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsxRuntimePragma.ts:21:49`.

```text
pattern: "h1>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 465, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsxRuntimePragma.ts:29:49`.

```text
pattern: "h1>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 469, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsxSpreadTag.ts:11:19`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 470, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsxSpreadTag.ts:12:25`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 471, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsxSpreadTag.ts:13:1`.

```text
pattern: ">"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 472, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsxSpreadTag.ts:15:19`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 473, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsxSpreadTag.ts:16:35`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 474, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsxSpreadTag.ts:17:1`.

```text
pattern: ">"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 475, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsxSpreadTag.ts:19:19`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 476, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsxSpreadTag.ts:20:39`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 477, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsxSpreadTag.ts:21:1`.

```text
pattern: ">"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 478, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsxSpreadTag.ts:23:19`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 479, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsxSpreadTag.ts:24:41`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 480, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/jsxSpreadTag.ts:25:1`.

```text
pattern: ">"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 489, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/modulePreserve3.ts:14:7`.

```text
pattern: ">"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 490, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/negatedUnicodeSetUnionMayContainStrings.ts:3:11`.

```text
pattern: "[^\\q{xy}b]"
flags: "v"
```

Refusal/diagnostic: regexp: cannot negate a class containing strings at byte 10

Node refusal: Invalid regular expression: /[^\q{xy}b]/dv: Negated character class may contain strings

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 491, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/negatedUnicodeSetUnionMayContainStrings.ts:4:11`.

```text
pattern: "[^b\\q{xy}]"
flags: "v"
```

Refusal/diagnostic: regexp: cannot negate a class containing strings at byte 10

Node refusal: Invalid regular expression: /[^b\q{xy}]/dv: Negated character class may contain strings

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 492, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/negatedUnicodeSetUnionMayContainStrings.ts:5:11`.

```text
pattern: "[^[b\\q{xy}c]]"
flags: "v"
```

Refusal/diagnostic: regexp: cannot negate a class containing strings at byte 13

Node refusal: Invalid regular expression: /[^[b\q{xy}c]]/dv: Negated character class may contain strings

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 494, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/parseJsxElementInUnaryExpressionNoCrash2.ts:5:5`.

```text
pattern: "> <"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 495, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/parseJsxExtends1.ts:10:28`.

```text
pattern: ">"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 496, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/parseJsxExtends2.ts:9:22`.

```text
pattern: ">"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 499, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/reactImportDropped.ts:42:48`.

```text
pattern: ">"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 502, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regExpWithOpenBracketInCharClass.ts:6:3`.

```text
pattern: "[[]"
flags: "v"
```

Refusal/diagnostic: regexp: unterminated character class at byte 3

Node refusal: Invalid regular expression: /[[]/dv: Unterminated character class

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 509, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regexNamedGroupDuplicateNestedInGroup.ts:4:1`.

```text
pattern: "(?<a>x)(?<a>y)"
flags: ""
```

Refusal/diagnostic: regexp: duplicate capture name a at byte 0

Node refusal: Invalid regular expression: /(?<a>x)(?<a>y)/d: Duplicate capture group name

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 510, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regexNamedGroupDuplicateNestedInGroup.ts:7:1`.

```text
pattern: "(?:(?<a>x))(?<a>z)"
flags: ""
```

Refusal/diagnostic: regexp: duplicate capture name a at byte 0

Node refusal: Invalid regular expression: /(?:(?<a>x))(?<a>z)/d: Duplicate capture group name

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 511, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regexNamedGroupDuplicateNestedInGroup.ts:10:1`.

```text
pattern: "(?:(?<a>x)|y)(?<a>z)"
flags: ""
```

Refusal/diagnostic: regexp: duplicate capture name a at byte 0

Node refusal: Invalid regular expression: /(?:(?<a>x)|y)(?<a>z)/d: Duplicate capture group name

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 512, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regexNamedGroupDuplicateNestedInGroup.ts:13:1`.

```text
pattern: "(?=(?<a>x))(?<a>z)"
flags: ""
```

Refusal/diagnostic: regexp: duplicate capture name a at byte 0

Node refusal: Invalid regular expression: /(?=(?<a>x))(?<a>z)/d: Duplicate capture group name

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 513, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regexNamedGroupDuplicateNestedInGroup.ts:16:1`.

```text
pattern: "((?<a>x))(?<a>z)"
flags: ""
```

Refusal/diagnostic: regexp: duplicate capture name a at byte 0

Node refusal: Invalid regular expression: /((?<a>x))(?<a>z)/d: Duplicate capture group name

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 519, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionAnnexB.ts:9:3`.

```text
pattern: "\\q\\u\\i\\c\\k\\_\\f\\o\\x\\-\\j\\u\\m\\p\\s"
flags: "u"
```

Refusal/diagnostic: regexp: invalid identity escape at byte 0

Node refusal: Invalid regular expression: /\q\u\i\c\k\_\f\o\x\-\j\u\m\p\s/du: Invalid escape

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 520, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionAnnexB.ts:10:3`.

```text
pattern: "[\\q\\u\\i\\c\\k\\_\\f\\o\\x\\-\\j\\u\\m\\p\\s]"
flags: "u"
```

Refusal/diagnostic: regexp: invalid identity escape at byte 1

Node refusal: Invalid regular expression: /[\q\u\i\c\k\_\f\o\x\-\j\u\m\p\s]/du: Invalid escape

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 521, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionAnnexB.ts:11:3`.

```text
pattern: "\\P[\\P\\w-_]"
flags: "u"
```

Refusal/diagnostic: regexp: invalid property escape at byte 0

Node refusal: Invalid regular expression: /\P[\P\w-_]/du: Invalid property name

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 532, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionAnnexB.ts:25:3`.

```text
pattern: "{1}??"
flags: ""
```

Refusal/diagnostic: regexp: nothing to repeat at byte 0

Node refusal: Invalid regular expression: /{1}??/d: Nothing to repeat

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 533, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionAnnexB.ts:26:3`.

```text
pattern: "{1,}??"
flags: ""
```

Refusal/diagnostic: regexp: nothing to repeat at byte 0

Node refusal: Invalid regular expression: /{1,}??/d: Nothing to repeat

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 534, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionAnnexB.ts:27:3`.

```text
pattern: "{1,2}??"
flags: ""
```

Refusal/diagnostic: regexp: nothing to repeat at byte 0

Node refusal: Invalid regular expression: /{1,2}??/d: Nothing to repeat

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 535, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionAnnexB.ts:28:3`.

```text
pattern: "{2,1}??"
flags: ""
```

Refusal/diagnostic: regexp: nothing to repeat at byte 0

Node refusal: Invalid regular expression: /{2,1}??/d: Nothing to repeat

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 536, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionAnnexB.ts:31:3`.

```text
pattern: "{??"
flags: "u"
```

Refusal/diagnostic: regexp: incomplete quantifier at byte 0

Node refusal: Invalid regular expression: /{??/du: Lone quantifier brackets

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 537, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionAnnexB.ts:32:3`.

```text
pattern: "{,??"
flags: "u"
```

Refusal/diagnostic: regexp: incomplete quantifier at byte 0

Node refusal: Invalid regular expression: /{,??/du: Lone quantifier brackets

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 538, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionAnnexB.ts:33:3`.

```text
pattern: "{,1??"
flags: "u"
```

Refusal/diagnostic: regexp: incomplete quantifier at byte 0

Node refusal: Invalid regular expression: /{,1??/du: Lone quantifier brackets

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 539, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionAnnexB.ts:34:3`.

```text
pattern: "{1??"
flags: "u"
```

Refusal/diagnostic: regexp: incomplete quantifier at byte 0

Node refusal: Invalid regular expression: /{1??/du: Lone quantifier brackets

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 540, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionAnnexB.ts:35:3`.

```text
pattern: "{1,??"
flags: "u"
```

Refusal/diagnostic: regexp: incomplete quantifier at byte 0

Node refusal: Invalid regular expression: /{1,??/du: Lone quantifier brackets

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 541, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionAnnexB.ts:36:3`.

```text
pattern: "{1,2??"
flags: "u"
```

Refusal/diagnostic: regexp: incomplete quantifier at byte 0

Node refusal: Invalid regular expression: /{1,2??/du: Lone quantifier brackets

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 542, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionAnnexB.ts:37:3`.

```text
pattern: "{2,1??"
flags: "u"
```

Refusal/diagnostic: regexp: incomplete quantifier at byte 0

Node refusal: Invalid regular expression: /{2,1??/du: Lone quantifier brackets

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 543, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionAnnexB.ts:38:3`.

```text
pattern: "{}??"
flags: "u"
```

Refusal/diagnostic: regexp: incomplete quantifier at byte 0

Node refusal: Invalid regular expression: /{}??/du: Lone quantifier brackets

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 544, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionAnnexB.ts:39:3`.

```text
pattern: "{,}??"
flags: "u"
```

Refusal/diagnostic: regexp: incomplete quantifier at byte 0

Node refusal: Invalid regular expression: /{,}??/du: Lone quantifier brackets

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 545, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionAnnexB.ts:40:3`.

```text
pattern: "{,1}??"
flags: "u"
```

Refusal/diagnostic: regexp: incomplete quantifier at byte 0

Node refusal: Invalid regular expression: /{,1}??/du: Lone quantifier brackets

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 546, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionAnnexB.ts:41:3`.

```text
pattern: "{1}??"
flags: "u"
```

Refusal/diagnostic: regexp: nothing to repeat at byte 0

Node refusal: Invalid regular expression: /{1}??/du: Nothing to repeat

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 547, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionAnnexB.ts:42:3`.

```text
pattern: "{1,}??"
flags: "u"
```

Refusal/diagnostic: regexp: nothing to repeat at byte 0

Node refusal: Invalid regular expression: /{1,}??/du: Nothing to repeat

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 548, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionAnnexB.ts:43:3`.

```text
pattern: "{1,2}??"
flags: "u"
```

Refusal/diagnostic: regexp: nothing to repeat at byte 0

Node refusal: Invalid regular expression: /{1,2}??/du: Nothing to repeat

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 549, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionAnnexB.ts:44:3`.

```text
pattern: "{2,1}??"
flags: "u"
```

Refusal/diagnostic: regexp: nothing to repeat at byte 0

Node refusal: Invalid regular expression: /{2,1}??/du: Nothing to repeat

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 550, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionCharacterClassRangeOrder.ts:9:2`.

```text
pattern: "[𝘈-𝘡][𝘡-𝘈]"
flags: ""
```

Refusal/diagnostic: regexp: invalid character class range at byte 10

Node refusal: Invalid regular expression: /[𝘈-𝘡][𝘡-𝘈]/d: Range out of order in character class

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 551, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionCharacterClassRangeOrder.ts:10:2`.

```text
pattern: "[𝘈-𝘡][𝘡-𝘈]"
flags: "u"
```

Refusal/diagnostic: regexp: invalid character class range at byte 21

Node refusal: Invalid regular expression: /[𝘈-𝘡][𝘡-𝘈]/du: Range out of order in character class

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 552, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionCharacterClassRangeOrder.ts:11:2`.

```text
pattern: "[𝘈-𝘡][𝘡-𝘈]"
flags: "v"
```

Refusal/diagnostic: regexp: invalid character class range at byte 21

Node refusal: Invalid regular expression: /[𝘈-𝘡][𝘡-𝘈]/dv: Range out of order in character class

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 553, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionCharacterClassRangeOrder.ts:13:2`.

```text
pattern: "[\\u{1D608}-\\u{1D621}][\\u{1D621}-\\u{1D608}]"
flags: ""
```

Refusal/diagnostic: regexp: invalid character class range at byte 13

Node refusal: Invalid regular expression: /[\u{1D608}-\u{1D621}][\u{1D621}-\u{1D608}]/d: Range out of order in character class

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 554, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionCharacterClassRangeOrder.ts:14:2`.

```text
pattern: "[\\u{1D608}-\\u{1D621}][\\u{1D621}-\\u{1D608}]"
flags: "u"
```

Refusal/diagnostic: regexp: invalid character class range at byte 41

Node refusal: Invalid regular expression: /[\u{1D608}-\u{1D621}][\u{1D621}-\u{1D608}]/du: Range out of order in character class

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 555, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionCharacterClassRangeOrder.ts:15:2`.

```text
pattern: "[\\u{1D608}-\\u{1D621}][\\u{1D621}-\\u{1D608}]"
flags: "v"
```

Refusal/diagnostic: regexp: invalid character class range at byte 41

Node refusal: Invalid regular expression: /[\u{1D608}-\u{1D621}][\u{1D621}-\u{1D608}]/dv: Range out of order in character class

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 556, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionCharacterClassRangeOrder.ts:17:2`.

```text
pattern: "[\\uD835\\uDE08-\\uD835\\uDE21][\\uD835\\uDE21-\\uD835\\uDE08]"
flags: ""
```

Refusal/diagnostic: regexp: invalid character class range at byte 20

Node refusal: Invalid regular expression: /[\uD835\uDE08-\uD835\uDE21][\uD835\uDE21-\uD835\uDE08]/d: Range out of order in character class

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 557, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionCharacterClassRangeOrder.ts:18:2`.

```text
pattern: "[\\uD835\\uDE08-\\uD835\\uDE21][\\uD835\\uDE21-\\uD835\\uDE08]"
flags: "u"
```

Refusal/diagnostic: regexp: invalid character class range at byte 53

Node refusal: Invalid regular expression: /[\uD835\uDE08-\uD835\uDE21][\uD835\uDE21-\uD835\uDE08]/du: Range out of order in character class

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 558, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionCharacterClassRangeOrder.ts:19:2`.

```text
pattern: "[\\uD835\\uDE08-\\uD835\\uDE21][\\uD835\\uDE21-\\uD835\\uDE08]"
flags: "v"
```

Refusal/diagnostic: regexp: invalid character class range at byte 53

Node refusal: Invalid regular expression: /[\uD835\uDE08-\uD835\uDE21][\uD835\uDE21-\uD835\uDE08]/dv: Range out of order in character class

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 566, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionES2025Syntax.ts:20:25`.

```text
pattern: "(?<dup>x)(?<dup>y)"
flags: ""
```

Refusal/diagnostic: regexp: duplicate capture name dup at byte 0

Node refusal: Invalid regular expression: /(?<dup>x)(?<dup>y)/d: Duplicate capture group name

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 570, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionGroupNameSuggestions.ts:2:15`.

```text
pattern: "(?<foo>)\\k<Foo>"
flags: ""
```

Refusal/diagnostic: regexp: unknown capture name at byte 8

Node refusal: Invalid regular expression: /(?<foo>)\k<Foo>/d: Invalid named capture referenced

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 577, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionGroupNameUnicodeEscapes.ts:13:1`.

```text
pattern: "(?<᧔>)\\k<᧔>\\k<\\u19D4>\\k<\\u{19D4}>"
flags: ""
```

Refusal/diagnostic: regexp: invalid capture name at byte 3

Node refusal: Invalid regular expression: /(?<᧔>)\k<᧔>\k<\u19D4>\k<\u{19D4}>/d: Invalid capture group name

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 578, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionGroupNameUnicodeEscapes.ts:14:1`.

```text
pattern: "(?<\\u19D4>)\\k<᧔>\\k<\\u19D4>\\k<\\u{19D4}>"
flags: ""
```

Refusal/diagnostic: regexp: invalid capture name at byte 3

Node refusal: Invalid regular expression: /(?<\u19D4>)\k<᧔>\k<\u19D4>\k<\u{19D4}>/d: Invalid capture group name

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 579, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionGroupNameUnicodeEscapes.ts:15:1`.

```text
pattern: "(?<\\u{19D4}>)\\k<᧔>\\k<\\u19D4>\\k<\\u{19D4}>"
flags: ""
```

Refusal/diagnostic: regexp: invalid capture name at byte 3

Node refusal: Invalid regular expression: /(?<\u{19D4}>)\k<᧔>\k<\u19D4>\k<\u{19D4}>/d: Invalid capture group name

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 589, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionGroupNameUnicodeEscapes.ts:31:1`.

```text
pattern: "(?<𑄽>)\\k<𑄽>\\k<\\u{1113D}>\\k<\\uD804\\uDD3D>"
flags: ""
```

Refusal/diagnostic: regexp: invalid capture name at byte 3

Node refusal: Invalid regular expression: /(?<𑄽>)\k<𑄽>\k<\u{1113D}>\k<\uD804\uDD3D>/d: Invalid capture group name

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 590, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionGroupNameUnicodeEscapes.ts:32:1`.

```text
pattern: "(?<\\u{1113D}>)\\k<𑄽>\\k<\\u{1113D}>\\k<\\uD804\\uDD3D>"
flags: ""
```

Refusal/diagnostic: regexp: invalid capture name at byte 3

Node refusal: Invalid regular expression: /(?<\u{1113D}>)\k<𑄽>\k<\u{1113D}>\k<\uD804\uDD3D>/d: Invalid capture group name

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 591, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionGroupNameUnicodeEscapes.ts:33:1`.

```text
pattern: "(?<\\uD804\\uDD3D>)\\k<𑄽>\\k<\\u{1113D}>\\k<\\uD804\\uDD3D>"
flags: ""
```

Refusal/diagnostic: regexp: invalid capture name at byte 3

Node refusal: Invalid regular expression: /(?<\uD804\uDD3D>)\k<𑄽>\k<\u{1113D}>\k<\uD804\uDD3D>/d: Invalid capture group name

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 595, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionGroupNameUnicodeEscapes.ts:42:1`.

```text
pattern: "(?<⽇>)(?<\\u2F47>)(?<\\u{2F47}>)\\k<⽇>\\k<\\u2F47>\\k<\\u{2F47}>"
flags: ""
```

Refusal/diagnostic: regexp: invalid capture name at byte 3

Node refusal: Invalid regular expression: /(?<⽇>)(?<\u2F47>)(?<\u{2F47}>)\k<⽇>\k<\u2F47>\k<\u{2F47}>/d: Invalid capture group name

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 596, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionGroupNameUnicodeEscapes.ts:43:1`.

```text
pattern: "(?<_⽇>)(?<_\\u2F47>)(?<_\\u{2F47}>)\\k<_⽇>\\k<_\\u2F47>\\k<_\\u{2F47}>"
flags: ""
```

Refusal/diagnostic: regexp: invalid capture name at byte 3

Node refusal: Invalid regular expression: /(?<_⽇>)(?<_\u2F47>)(?<_\u{2F47}>)\k<_⽇>\k<_\u2F47>\k<_\u{2F47}>/d: Invalid capture group name

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 597, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionGroupNameUnicodeEscapes.ts:46:1`.

```text
pattern: "(?<🌚>)(?<\\u{1F31A}>)(?<\\uD83C\\uDF1A>)\\k<🌚>\\k<\\u{1F31A}>\\k<\\uD83C\\uDF1A>"
flags: ""
```

Refusal/diagnostic: regexp: invalid capture name at byte 3

Node refusal: Invalid regular expression: /(?<🌚>)(?<\u{1F31A}>)(?<\uD83C\uDF1A>)\k<🌚>\k<\u{1F31A}>\k<\uD83C\uDF1A>/d: Invalid capture group name

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 598, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionGroupNameUnicodeEscapes.ts:47:1`.

```text
pattern: "(?<_🌚>)(?<_\\u{1F31A}>)(?<_\\uD83C\\uDF1A>)\\k<_🌚>\\k<_\\u{1F31A}>\\k<_\\uD83C\\uDF1A>"
flags: ""
```

Refusal/diagnostic: regexp: invalid capture name at byte 3

Node refusal: Invalid regular expression: /(?<_🌚>)(?<_\u{1F31A}>)(?<_\uD83C\uDF1A>)\k<_🌚>\k<_\u{1F31A}>\k<_\uD83C\uDF1A>/d: Invalid capture group name

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 599, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionGroupNameUnicodeEscapes.ts:50:1`.

```text
pattern: "(?<\\uD800>)(?<\\u{D800}>)\\k<\\uD800>\\k<\\u{D800}>"
flags: ""
```

Refusal/diagnostic: regexp: invalid capture name at byte 3

Node refusal: Invalid regular expression: /(?<\uD800>)(?<\u{D800}>)\k<\uD800>\k<\u{D800}>/d: Invalid capture group name

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 600, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionGroupNameUnicodeEscapes.ts:51:1`.

```text
pattern: "(?<_\\uD800>)(?<_\\u{D800}>)\\k<_\\uD800>\\k<_\\u{D800}>"
flags: ""
```

Refusal/diagnostic: regexp: invalid capture name at byte 3

Node refusal: Invalid regular expression: /(?<_\uD800>)(?<_\u{D800}>)\k<_\uD800>\k<_\u{D800}>/d: Invalid capture group name

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 601, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionGroupNameUnicodeEscapes.ts:54:1`.

```text
pattern: "(?<\\uDFFF>)(?<\\u{DFFF}>)\\k<\\uDFFF>\\k<\\u{DFFF}>"
flags: ""
```

Refusal/diagnostic: regexp: invalid capture name at byte 3

Node refusal: Invalid regular expression: /(?<\uDFFF>)(?<\u{DFFF}>)\k<\uDFFF>\k<\u{DFFF}>/d: Invalid capture group name

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 602, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionGroupNameUnicodeEscapes.ts:55:1`.

```text
pattern: "(?<_\\uDFFF>)(?<_\\u{DFFF}>)\\k<_\\uDFFF>\\k<_\\u{DFFF}>"
flags: ""
```

Refusal/diagnostic: regexp: invalid capture name at byte 3

Node refusal: Invalid regular expression: /(?<_\uDFFF>)(?<_\u{DFFF}>)\k<_\uDFFF>\k<_\u{DFFF}>/d: Invalid capture group name

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 603, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionGroupNameUnicodeEscapes.ts:58:1`.

```text
pattern: "(?<\\u{D800}\\u{DC00}>)\\k<\\u{D800}\\u{DC00}>"
flags: ""
```

Refusal/diagnostic: regexp: invalid capture name at byte 3

Node refusal: Invalid regular expression: /(?<\u{D800}\u{DC00}>)\k<\u{D800}\u{DC00}>/d: Invalid capture group name

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 604, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionGroupNameUnicodeEscapes.ts:59:1`.

```text
pattern: "(?<_\\u{D800}\\u{DC00}>)\\k<_\\u{D800}\\u{DC00}>"
flags: ""
```

Refusal/diagnostic: regexp: invalid capture name at byte 3

Node refusal: Invalid regular expression: /(?<_\u{D800}\u{DC00}>)\k<_\u{D800}\u{DC00}>/d: Invalid capture group name

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 609, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionQuantifierBounds1.ts:7:2`.

```text
pattern: "a{8,7}"
flags: ""
```

Refusal/diagnostic: regexp: quantifier range out of order at byte 1

Node refusal: Invalid regular expression: /a{8,7}/d: numbers out of order in {} quantifier

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 613, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionScanning.ts:5:2`.

```text
pattern: "foo"
flags: "visualstudiocode"
```

Refusal/diagnostic: regexp: invalid flag at byte 4

Node refusal: Invalid flags supplied to RegExp constructor 'visualstudiocode'

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 614, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionScanning.ts:7:2`.

```text
pattern: "(?med-ium:bar)"
flags: ""
```

Refusal/diagnostic: regexp: invalid modifiers at byte 3

Node refusal: Invalid regular expression: /(?med-ium:bar)/d: Invalid group

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 621, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionScanning.ts:15:2`.

```text
pattern: "\\2()(\\12)(foo)\\1\\0[\\0\\1\\01\\123\\08\\8](\\3\\03)\\5\\005\\9\\009"
flags: "u"
```

Refusal/diagnostic: regexp: invalid decimal escape at byte 5

Node refusal: Invalid regular expression: /\2()(\12)(foo)\1\0[\0\1\01\123\08\8](\3\03)\5\005\9\009/du: Invalid escape

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 623, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionScanning.ts:17:2`.

```text
pattern: "(\\k<bar>)\\k<absent>(?<foo>foo)|(?<bar>)((?<foo>)|(bar(?<bar>bar)))"
flags: ""
```

Refusal/diagnostic: regexp: unknown capture name at byte 9

Node refusal: Invalid regular expression: /(\k<bar>)\k<absent>(?<foo>foo)|(?<bar>)((?<foo>)|(bar(?<bar>bar)))/d: Invalid named capture referenced

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 624, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionScanning.ts:19:2`.

```text
pattern: "{}{1,2}_{3}.{4,}?(foo){008}${32,16}\\b{064,128}.+&*?\\???\\n{,256}{\\\\{,"
flags: ""
```

Refusal/diagnostic: regexp: nothing to repeat at byte 28

Node refusal: Invalid regular expression: /{}{1,2}_{3}.{4,}?(foo){008}${32,16}\b{064,128}.+&*?\???\n{,256}{\\{,/d: Nothing to repeat

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 625, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionScanning.ts:21:2`.

```text
pattern: "[-A-Za-z-z-aZ-A\\d_-\\d-.-.\\r-\\n\\w-\\W]"
flags: ""
```

Refusal/diagnostic: regexp: invalid character class range at byte 12

Node refusal: Invalid regular expression: /[-A-Za-z-z-aZ-A\d_-\d-.-.\r-\n\w-\W]/d: Range out of order in character class

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 627, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionScanning.ts:23:2`.

```text
pattern: "\\p{L}\\p{gc=L}\\p{ASCII}\\p{Invalid}[\\p{L}\\p{gc=L}\\P{ASCII}\\p{Invalid}]"
flags: "u"
```

Refusal/diagnostic: regexp: invalid Unicode property at byte 22

Node refusal: Invalid regular expression: /\p{L}\p{gc=L}\p{ASCII}\p{Invalid}[\p{L}\p{gc=L}\P{ASCII}\p{Invalid}]/du: Invalid property name

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 628, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionScanning.ts:24:2`.

```text
pattern: "\\p{L}\\p{gc=L}\\p{ASCII}\\p{Invalid}[\\p{L}\\p{gc=L}\\P{ASCII}\\p{Invalid}]"
flags: "v"
```

Refusal/diagnostic: regexp: invalid Unicode property at byte 22

Node refusal: Invalid regular expression: /\p{L}\p{gc=L}\p{ASCII}\p{Invalid}[\p{L}\p{gc=L}\P{ASCII}\p{Invalid}]/dv: Invalid property name

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 630, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionScanning.ts:26:2`.

```text
pattern: "\\p{InvalidProperty=Value}\\p{=}\\p{sc=}\\P{=foo}[\\p{}\\p\\\\\\P\\P{]\\p{"
flags: "u"
```

Refusal/diagnostic: regexp: invalid Unicode property at byte 0

Node refusal: Invalid regular expression: /\p{InvalidProperty=Value}\p{=}\p{sc=}\P{=foo}[\p{}\p\\\P\P{]\p{/du: Invalid property name

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 631, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionScanning.ts:27:2`.

```text
pattern: "\\p{InvalidProperty=Value}\\p{=}\\p{sc=}\\P{=foo}[\\p{}\\p\\\\\\P\\P{]\\p{"
flags: "v"
```

Refusal/diagnostic: regexp: invalid Unicode property at byte 0

Node refusal: Invalid regular expression: /\p{InvalidProperty=Value}\p{=}\p{sc=}\P{=foo}[\p{}\p\\\P\P{]\p{/dv: Invalid property name

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 633, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionScanning.ts:29:2`.

```text
pattern: "\\p{RGI_Emoji}\\P{RGI_Emoji}[^\\p{RGI_Emoji}\\P{RGI_Emoji}]"
flags: "u"
```

Refusal/diagnostic: regexp: invalid Unicode property at byte 0

Node refusal: Invalid regular expression: /\p{RGI_Emoji}\P{RGI_Emoji}[^\p{RGI_Emoji}\P{RGI_Emoji}]/du: Invalid property name

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 634, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionScanning.ts:30:2`.

```text
pattern: "\\p{RGI_Emoji}\\P{RGI_Emoji}[^\\p{RGI_Emoji}\\P{RGI_Emoji}]"
flags: "v"
```

Refusal/diagnostic: regexp: invalid Unicode property at byte 13

Node refusal: Invalid regular expression: /\p{RGI_Emoji}\P{RGI_Emoji}[^\p{RGI_Emoji}\P{RGI_Emoji}]/dv: Invalid property name

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 636, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionScanning.ts:33:2`.

```text
pattern: "\\c[\\c0\\ca\\cQ\\c\\C]\\c1\\C"
flags: "u"
```

Refusal/diagnostic: regexp: invalid control escape at byte 0

Node refusal: Invalid regular expression: /\c[\c0\ca\cQ\c\C]\c1\C/du: Invalid Unicode escape

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 638, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionScanning.ts:35:2`.

```text
pattern: "\\q\\\\\\`[\\q\\\\\\`[\\Q\\\\\\Q{\\q{foo|bar|baz]\\q{]\\q{"
flags: "u"
```

Refusal/diagnostic: regexp: invalid identity escape at byte 0

Node refusal: Invalid regular expression: /\q\\\`[\q\\\`[\Q\\\Q{\q{foo|bar|baz]\q{]\q{/du: Invalid escape

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 639, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionScanning.ts:36:2`.

```text
pattern: "\\q\\\\\\`[\\q\\\\\\`[\\Q\\\\\\Q{\\q{foo|bar|baz]\\q{]\\q{"
flags: "v"
```

Refusal/diagnostic: regexp: invalid identity escape at byte 0

Node refusal: Invalid regular expression: /\q\\\`[\q\\\`[\Q\\\Q{\q{foo|bar|baz]\q{]\q{/dv: Invalid escape

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 640, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionScanning.ts:38:2`.

```text
pattern: "[a--b[--][\\d++[]]&&[[&0-9--]&&[\\p{L}]--\\P{L}-_-]]&&&\\q{foo}[0---9][&&q&&&\\q{bar}&&]"
flags: ""
```

Refusal/diagnostic: regexp: invalid character class range at byte 4

Node refusal: Invalid regular expression: /[a--b[--][\d++[]]&&[[&0-9--]&&[\p{L}]--\P{L}-_-]]&&&\q{foo}[0---9][&&q&&&\q{bar}&&]/d: Range out of order in character class

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 641, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionScanning.ts:39:2`.

```text
pattern: "[a--b[--][\\d++[]]&&[[&0-9--]&&[\\p{L}]--\\P{L}-_-]]&&&\\q{foo}[0---9][&&q&&&\\q{bar}&&]"
flags: "u"
```

Refusal/diagnostic: regexp: invalid character class range at byte 4

Node refusal: Invalid regular expression: /[a--b[--][\d++[]]&&[[&0-9--]&&[\p{L}]--\P{L}-_-]]&&&\q{foo}[0---9][&&q&&&\q{bar}&&]/du: Range out of order in character class

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 642, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionScanning.ts:40:2`.

```text
pattern: "[a--b[--][\\d++[]]&&[[&0-9--]&&[\\p{L}]--\\P{L}-_-]]&&&\\q{foo}[0---9][&&q&&&\\q{bar}&&]"
flags: "v"
```

Refusal/diagnostic: regexp: unterminated character class at byte 5

Node refusal: Invalid regular expression: /[a--b[--][\d++[]]&&[[&0-9--]&&[\p{L}]--\P{L}-_-]]&&&\q{foo}[0---9][&&q&&&\q{bar}&&]/dv: Invalid set operation in character class

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 643, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionScanning.ts:41:2`.

```text
pattern: "[[^\\P{Decimal_Number}&&[0-9]]&&\\p{L}&&\\p{ID_Continue}--\\p{ASCII}\\p{CWCF}]"
flags: "v"
```

Refusal/diagnostic: regexp: unterminated character class at byte 53

Node refusal: Invalid regular expression: /[[^\P{Decimal_Number}&&[0-9]]&&\p{L}&&\p{ID_Continue}--\p{ASCII}\p{CWCF}]/dv: Invalid set operation in character class

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 644, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionScanning.ts:42:2`.

```text
pattern: "[^\\p{Emoji}\\p{RGI_Emoji}][^\\p{Emoji}--\\p{RGI_Emoji}][^\\p{Emoji}&&\\p{RGI_Emoji}]"
flags: "v"
```

Refusal/diagnostic: regexp: cannot negate a class containing strings at byte 25

Node refusal: Invalid regular expression: /[^\p{Emoji}\p{RGI_Emoji}][^\p{Emoji}--\p{RGI_Emoji}][^\p{Emoji}&&\p{RGI_Emoji}]/dv: Negated character class may contain strings

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 645, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionScanning.ts:43:2`.

```text
pattern: "[^\\p{RGI_Emoji}\\p{Emoji}][^\\p{RGI_Emoji}--\\p{Emoji}][^\\p{RGI_Emoji}&&\\p{Emoji}]"
flags: "v"
```

Refusal/diagnostic: regexp: cannot negate a class containing strings at byte 25

Node refusal: Invalid regular expression: /[^\p{RGI_Emoji}\p{Emoji}][^\p{RGI_Emoji}--\p{Emoji}][^\p{RGI_Emoji}&&\p{Emoji}]/dv: Negated character class may contain strings

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 646, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionScanning.ts:44:2`.

```text
pattern: "[^\\p{RGI_Emoji}\\q{foo}][^\\p{RGI_Emoji}--\\q{foo}][^\\p{RGI_Emoji}&&\\q{foo}]"
flags: "v"
```

Refusal/diagnostic: regexp: cannot negate a class containing strings at byte 23

Node refusal: Invalid regular expression: /[^\p{RGI_Emoji}\q{foo}][^\p{RGI_Emoji}--\q{foo}][^\p{RGI_Emoji}&&\q{foo}]/dv: Negated character class may contain strings

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 647, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionScanning.ts:45:2`.

```text
pattern: "[^\\p{Emoji}[[\\p{RGI_Emoji}]]][^\\p{Emoji}--[[\\p{RGI_Emoji}]]][^\\p{Emoji}&&[[\\p{RGI_Emoji}]]]"
flags: "v"
```

Refusal/diagnostic: regexp: cannot negate a class containing strings at byte 29

Node refusal: Invalid regular expression: /[^\p{Emoji}[[\p{RGI_Emoji}]]][^\p{Emoji}--[[\p{RGI_Emoji}]]][^\p{Emoji}&&[[\p{RGI_Emoji}]]]/dv: Negated character class may contain strings

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 648, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionScanning.ts:46:2`.

```text
pattern: "[^[[\\p{RGI_Emoji}]]\\p{Emoji}][^[[\\p{RGI_Emoji}]]--\\p{Emoji}][^[[\\p{RGI_Emoji}]]&&\\p{Emoji}]"
flags: "v"
```

Refusal/diagnostic: regexp: cannot negate a class containing strings at byte 29

Node refusal: Invalid regular expression: /[^[[\p{RGI_Emoji}]]\p{Emoji}][^[[\p{RGI_Emoji}]]--\p{Emoji}][^[[\p{RGI_Emoji}]]&&\p{Emoji}]/dv: Negated character class may contain strings

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 649, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionScanning.ts:47:2`.

```text
pattern: "[^[[\\p{RGI_Emoji}]]\\q{foo}][^[[\\p{RGI_Emoji}]]--\\q{foo}][^[[\\p{RGI_Emoji}]]&&\\q{foo}]"
flags: "v"
```

Refusal/diagnostic: regexp: cannot negate a class containing strings at byte 27

Node refusal: Invalid regular expression: /[^[[\p{RGI_Emoji}]]\q{foo}][^[[\p{RGI_Emoji}]]--\q{foo}][^[[\p{RGI_Emoji}]]&&\q{foo}]/dv: Negated character class may contain strings

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 650, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionScanning.ts:48:2`.

```text
pattern: "[^\\q{foo|bar|baz}--\\q{foo}--\\q{bar}--\\q{baz}][^\\p{L}--\\q{foo}--[\\q{bar}]--[^[\\q{baz}]]]"
flags: "v"
```

Refusal/diagnostic: regexp: cannot negate a class containing strings at byte 45

Node refusal: Invalid regular expression: /[^\q{foo|bar|baz}--\q{foo}--\q{bar}--\q{baz}][^\p{L}--\q{foo}--[\q{bar}]--[^[\q{baz}]]]/dv: Negated character class may contain strings

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 651, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionScanning.ts:49:2`.

```text
pattern: "[^[[\\q{foo|bar|baz}]]--\\q{foo}--\\q{bar}--\\q{baz}][^[^[^\\p{L}]]--\\q{foo}--[\\q{bar}]--[^[\\q{baz}]]]"
flags: "v"
```

Refusal/diagnostic: regexp: cannot negate a class containing strings at byte 49

Node refusal: Invalid regular expression: /[^[[\q{foo|bar|baz}]]--\q{foo}--\q{bar}--\q{baz}][^[^[^\p{L}]]--\q{foo}--[\q{bar}]--[^[\q{baz}]]]/dv: Negated character class may contain strings

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 652, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionUnicodePropertyValueExpressionSuggestions.ts:2:15`.

```text
pattern: "\\p{ascii}\\p{Sc=Unknown}\\p{sc=unknownX}\\p{Script_Declensions=Inherited}\\p{scx=inherit}"
flags: "u"
```

Refusal/diagnostic: regexp: invalid Unicode property at byte 0

Node refusal: Invalid regular expression: /\p{ascii}\p{Sc=Unknown}\p{sc=unknownX}\p{Script_Declensions=Inherited}\p{scx=inherit}/du: Invalid property name

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 656, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/regularExpressionWithNonBMPFlags.ts:9:20`.

```text
pattern: "(?𝘴𝘪-𝘮:^𝘧𝘰𝘰.)"
flags: "𝘨𝘮𝘶"
```

Refusal/diagnostic: regexp: invalid or duplicate flag at byte 0

Node refusal: Invalid flags supplied to RegExp constructor '𝘨𝘮𝘶d'

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 657, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/shebang.ts:2:3`.

```text
pattern: "usr"
flags: "bin"
```

Refusal/diagnostic: regexp: invalid flag at byte 0

Node refusal: Invalid flags supplied to RegExp constructor 'bind'

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 658, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/shebangBeforeReferences.ts:10:3`.

```text
pattern: "usr"
flags: "bin"
```

Refusal/diagnostic: regexp: invalid flag at byte 0

Node refusal: Invalid flags supplied to RegExp constructor 'bind'

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 659, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/shebangError.ts:4:3`.

```text
pattern: "usr"
flags: "bin"
```

Refusal/diagnostic: regexp: invalid flag at byte 0

Node refusal: Invalid flags supplied to RegExp constructor 'bind'

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 661, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/tsxFragmentChildrenCheck.ts:11:31`.

```text
pattern: "span>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 662, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/tsxFragmentChildrenCheck.ts:26:22`.

```text
pattern: ">"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 663, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/tsxFragmentChildrenCheck.ts:27:29`.

```text
pattern: "span>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 664, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/tsxFragmentChildrenCheck.ts:29:8`.

```text
pattern: ">"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 665, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/tsxResolveExternalModuleExportsTypes.ts:25:19`.

```text
pattern: "h1>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 669, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/unterminatedRegexAtEndOfSource1.ts:2:9`.

```text
pattern: ""
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 672, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/unusedImports13.ts:10:35`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 673, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/unusedImports14.ts:10:35`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 674, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/unusedImports15.ts:11:35`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 675, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/unusedImports16.ts:11:35`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 684, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/compiler/verbatimModuleSyntaxReactReference.ts:16:34`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 701, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/es6/unicodeExtendedEscapes/unicodeExtendedEscapesInRegularExpressions07.ts:5:9`.

```text
pattern: "\\u{110000}"
flags: "gu"
```

Refusal/diagnostic: regexp: invalid Unicode escape at byte 0

Node refusal: Invalid regular expression: /\u{110000}/dgu: Invalid Unicode escape

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 706, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/es6/unicodeExtendedEscapes/unicodeExtendedEscapesInRegularExpressions12.ts:3:9`.

```text
pattern: "\\u{FFFFFFFF}"
flags: "gu"
```

Refusal/diagnostic: regexp: invalid Unicode escape at byte 0

Node refusal: Invalid regular expression: /\u{FFFFFFFF}/dgu: Invalid Unicode escape

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 708, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/es6/unicodeExtendedEscapes/unicodeExtendedEscapesInRegularExpressions14.ts:4:9`.

```text
pattern: "\\u{-DDDD}"
flags: "gu"
```

Refusal/diagnostic: regexp: invalid Unicode escape at byte 0

Node refusal: Invalid regular expression: /\u{-DDDD}/dgu: Invalid Unicode escape

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 711, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/es6/unicodeExtendedEscapes/unicodeExtendedEscapesInRegularExpressions17.ts:3:9`.

```text
pattern: "\\u{r}\\u{n}\\u{t}"
flags: "gu"
```

Refusal/diagnostic: regexp: invalid Unicode escape at byte 0

Node refusal: Invalid regular expression: /\u{r}\u{n}\u{t}/dgu: Invalid Unicode escape

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 713, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/es6/unicodeExtendedEscapes/unicodeExtendedEscapesInRegularExpressions19.ts:3:9`.

```text
pattern: "\\u{}"
flags: "gu"
```

Refusal/diagnostic: regexp: invalid Unicode escape at byte 0

Node refusal: Invalid regular expression: /\u{}/dgu: Invalid Unicode escape

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 730, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsdoc/declarations/jsDeclarationsNonIdentifierInferredNames.ts:13:13`.

```text
pattern: ">"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 731, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsdoc/declarations/jsDeclarationsReactComponents.ts:19:14`.

```text
pattern: ">"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 732, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsdoc/declarations/jsDeclarationsReactComponents.ts:43:10`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 733, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsdoc/declarations/jsDeclarationsReactComponents.ts:62:10`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 734, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsdoc/declarations/jsDeclarationsReactComponents.ts:78:10`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 735, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsdoc/declarations/jsDeclarationsReactComponents.ts:92:15`.

```text
pattern: ">"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 740, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxAttributeInitializer.ts:10:35`.

```text
pattern: ">"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 741, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxAttributeInitializer.ts:11:16`.

```text
pattern: ">"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 742, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxAttributeInitializer.ts:12:2`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 743, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxUnclosedParserRecovery.ts:17:2`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 744, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxUnclosedParserRecovery.ts:21:2`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 745, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxUnclosedParserRecovery.ts:25:2`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 746, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxUnclosedParserRecovery.ts:29:2`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 747, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxUnclosedParserRecovery.ts:33:2`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 748, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxUnclosedParserRecovery.ts:36:11`.

```text
pattern: ""
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 749, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxUnclosedParserRecovery.ts:37:2`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 750, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxUnclosedParserRecovery.ts:41:2`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 751, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxUnclosedParserRecovery.ts:45:2`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 752, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxUnclosedParserRecovery.ts:49:2`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 753, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxUnclosedParserRecovery.ts:54:11`.

```text
pattern: ">"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 754, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxUnclosedParserRecovery.ts:55:2`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 755, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxUnclosedParserRecovery.ts:59:11`.

```text
pattern: ">"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 756, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxUnclosedParserRecovery.ts:60:2`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 757, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxUnclosedParserRecovery.ts:64:11`.

```text
pattern: ">"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 758, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxUnclosedParserRecovery.ts:65:2`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 759, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxUnclosedParserRecovery.ts:70:2`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 760, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxUnclosedParserRecovery.ts:75:2`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 761, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxUnclosedParserRecovery.ts:78:11`.

```text
pattern: ""
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 762, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxUnclosedParserRecovery.ts:80:2`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 763, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxUnclosedParserRecovery.ts:84:11`.

```text
pattern: ">"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 764, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxUnclosedParserRecovery.ts:85:2`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 765, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxUnclosedParserRecovery.ts:89:11`.

```text
pattern: ">"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 766, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxUnclosedParserRecovery.ts:90:2`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 767, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxUnclosedParserRecovery.ts:94:11`.

```text
pattern: ">"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 768, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxUnclosedParserRecovery.ts:95:2`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 769, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxUnclosedParserRecovery.ts:101:2`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 770, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxUnclosedParserRecovery.ts:106:2`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 771, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxUnclosedParserRecovery.ts:111:2`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 772, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxUnclosedParserRecovery.ts:116:2`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 773, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxUnclosedParserRecovery.ts:121:2`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 774, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxUnclosedParserRecovery.ts:124:11`.

```text
pattern: ""
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 775, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxUnclosedParserRecovery.ts:126:2`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 776, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxUnclosedParserRecovery.ts:131:2`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 777, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxUnclosedParserRecovery.ts:136:2`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 778, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/jsxUnclosedParserRecovery.ts:141:2`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 779, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/tsxAttributeInvalidNames.tsx:14:19`.

```text
pattern: ">"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 780, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/tsxEmitSpreadAttribute.ts:9:46`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 781, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/tsxEmitSpreadAttribute.ts:13:55`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 782, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/tsxEmitSpreadAttribute.ts:17:55`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 783, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/tsxEmitSpreadAttribute.ts:21:59`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 784, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/tsxEmitSpreadAttribute.ts:25:72`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 785, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/tsxEmitSpreadAttribute.ts:29:78`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 786, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/tsxEmitSpreadAttribute.ts:33:76`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 787, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/tsxEmitSpreadAttribute.ts:37:66`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 788, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/tsxEmitSpreadAttribute.ts:43:66`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 789, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/tsxEmitSpreadAttribute.ts:47:70`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 790, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/tsxEmitSpreadAttribute.ts:51:60`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 791, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/tsxReactEmitSpreadAttribute.ts:7:46`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 792, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/tsxReactEmitSpreadAttribute.ts:11:55`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 793, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/tsxReactEmitSpreadAttribute.ts:15:55`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 794, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/tsxReactEmitSpreadAttribute.ts:19:59`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 795, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/tsxReactEmitSpreadAttribute.ts:23:72`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 796, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/tsxReactEmitSpreadAttribute.ts:27:78`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 797, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/tsxReactEmitSpreadAttribute.ts:31:20`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 798, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/tsxReactEmitSpreadAttribute.ts:35:76`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 799, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/tsxReactEmitSpreadAttribute.ts:39:66`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 800, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/tsxReactEmitSpreadAttribute.ts:45:68`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 801, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/tsxReactEmitSpreadAttribute.ts:49:70`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 802, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/jsx/tsxReactEmitSpreadAttribute.ts:53:60`.

```text
pattern: "div>"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 803, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/parser/ecmascript2021/numericSeparators/parser.numericSeparators.unicodeEscape.ts:13:13`.

```text
pattern: "u"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 804, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/parser/ecmascript2021/numericSeparators/parser.numericSeparators.unicodeEscape.ts:49:13`.

```text
pattern: "u"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 805, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/parser/ecmascript2021/numericSeparators/parser.numericSeparators.unicodeEscape.ts:85:13`.

```text
pattern: "u"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 806, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/parser/ecmascript2021/numericSeparators/parser.numericSeparators.unicodeEscape.ts:121:14`.

```text
pattern: "u"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 808, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/parser/ecmascript5/MissingTokens/parserMissingToken2.ts:2:1`.

```text
pattern: " b"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 827, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/parser/ecmascript5/RegressionTests/parser579071.ts:2:9`.

```text
pattern: "fo(o"
flags: ""
```

Refusal/diagnostic: regexp: unterminated group at byte 4

Node refusal: Invalid regular expression: /fo(o/d: Unterminated group

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 831, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/parser/ecmascript5/RegressionTests/parser645086_1.ts:2:14`.

```text
pattern: ""
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 833, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/parser/ecmascript5/RegressionTests/parser645086_2.ts:2:15`.

```text
pattern: ""
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.

### Pattern 851, compiler, refusal

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/parser/ecmascript5/RegularExpressions/parserRegularExpressionDivideAmbiguity3.ts:2:8`.

```text
pattern: "regexp"
flags: "a"
```

Refusal/diagnostic: regexp: invalid flag at byte 0

Node refusal: Invalid flags supplied to RegExp constructor 'ad'

Feature needed (inference): Invalid compiler test fixture; Node rejects it too. No missing valid syntax is established.

### Pattern 852, compiler, source-only rejection

Location: `cohere/TypeScript/tsc/testdata/tests/cases/conformance/parser/ecmascript5/RegularExpressions/parserRegularExpressionDivideAmbiguity4.ts:2:5`.

```text
pattern: "notregexp"
flags: ""
```

Refusal/diagnostic: Invalid regular expression: missing /

Node refusal: Invalid regular expression: missing /

Feature needed (inference): Invalid source literal recovered by the TypeScript parser; its body/flags alone are valid. This is a source lexical error, not an Adamic regex compiler defect.
