# Shared lint helpers

The first fan-out prerequisites, ordered by complete helper blockers removed in the inventory's frozen 198-rule cohort. See [HELPERS.md](../HELPERS.md) for the ownership claim and rule lists. These files are Adamic TypeScript, run unchanged on Node and sanitized native; they do not change the existing linter's entry point.

| File | Contract | Consumers covered | Additional rules with no listed helper blocker |
|---|---|---:|---:|
| `options_json.ts` | JSON grammar, decoded strings, numeric lexemes, last duplicate key, acyclic value arena | prerequisite for 103 option consumers | 0 |
| `option_schema.ts` | supported `optionschema.Schema.Validate` acceptance, including local references, deep equality, anyOf/oneOf/allOf/not | 94 complete upstream schemas | 0 until strict target validation is available |
| `strict_options.ts` | `rule.UnmarshalOptions` acceptance for generated target descriptors: exact tagged keys, ASCII case folding for untagged keys, nested structs/lists/maps, null, scalar types, exact signed integer bounds | 97 registered Go target types | 30 with the preceding two helpers |
| `policy_message.ts` | `MessageCatalog.Render` from Go's validated, language-resolved catalog, with strict values and object phrase selections | all 54 policy consumers in the cohort | 16 after options, 46 cumulative |

The 30 and 16 are **helper readiness**, conditional on the inventory's common AST adapter assumption. They are not implemented rules, findings parity claims, or exhaustive proofs over arbitrary configurations. Each rule worker still ports its rule-local decoder, defaults, listeners, spans, fixes and suggestions. `StrictOptions.check` validates the raw target representation; `OptionsJson` exposes parsed values. It does not construct a Go struct, apply rule defaults, merge duplicate objects into an existing target, invoke custom marshalers, or convert a configuration string into a rule's enum. Preserve those operations in the per-rule port. `Settings.load` is not replaced.

```ts
const schema = new OptionSchema(schemaText); // upstream normalized option-list schema
const verdict = schema.check('[{"fixToUnknown":true}]');
const target = new StrictOptions(goTargetDescriptor);
const targetVerdict = target.check('{"fixToUnknown":true}');
const values = new OptionsJson('{"fixToUnknown":true}');
const root = values.parse(); // -1 on failure; inspect values.error before proceeding
```

Schema and target instances can be reused. A schema's unsupported vocabulary is refused at construction; dynamic refusal state resets for the next check. Treat `NotYet:` as a blocked configuration, never as a clean lint result. A parse error is exposed through `OptionsJson.error`. The reader's nesting limit is 512; callers must report that refusal rather than treating it as a lint result. Target descriptors and catalogs must come from the generator, not handwritten guesses.

`PolicyMessage` accepts the JSON catalog exported from **Go's resolved templates**, not the raw policy files. Go performs the source-file naming, language, term, phrase, unused-entry and catalog validation. Runtime rendering selects phrases before substituting values, in sorted placeholder order, including the substitution behavior when values themselves contain placeholders. Catalog replacement is a new instance. No process-global message claim registry, concurrent swap API, or raw catalog loader is ported. The catalog contains both languages; this unit's coverage concerns TypeScript consumers.

Validation uses 23,539 cases generated against the original metadata pin (including the pinned upstream ESLint acceptance/refusal samples). Go cohere provides the verdicts and rendered bytes. The suite also checks ten message invariant failures on Node and native, explicit unsupported regex/custom/Unicode-fold cases, and reusable-state recovery. Metadata drift fails the suite. Fixture definitions are deduplicated and the corpus is compressed; tests expand it only in their temporary directory.

The comparator checks Go acceptance/refusal, not Go's full schema diagnostic prose or decoded struct contents. Policy successful output is byte-for-byte; the ten failure checks also compare panic text and exit 70. Arbitrary malformed catalog input and arbitrary non-ASCII/control-character identifiers in panic messages are not covered. Case generation is bounded; it is not an exhaustive option fuzzer or a replay of every rule's finding corpus.

From the repository root, with the setup environment sourced:

```
python3 stage1/cohere/lint/helpers/testdata/regenerate.py > /tmp/lint-helper-regenerate.log 2>&1
go test ./stage1/cohere/lint/helpers -count=1 -v -timeout=20m > /tmp/lint-helper-tests.log 2>&1
```

The Go overlay adds oracle-only files to cohere without editing its worktree. `catalog.json` and `descriptors.json` were refreshed against cohere `f5d1934a2d7bebe706210cb1cfd01aebff4f8ca7` through the same oracle. The catalog is unchanged. Sixteen target descriptors changed: twelve Tailwind rules gain the selector option shapes in units 6a/6b (`0b892cc9`, `f5d1934a`); `id-length`, `id-match`, `object-shorthand` and `@typescript-eslint/no-require-imports` now describe the JS regular-expression wrapper introduced by `c2e39b75`. The frozen cases and helper implementation are unchanged; their parity checks still run. The original corpus generation and evidence used cohere `7945d102a6c18dd36adf9114a758ce646e8b2359`. See `evidence/helpers-final.log`, `evidence/oracle.log`, `evidence/vet.log` and `REPORT.md` for that original run.
