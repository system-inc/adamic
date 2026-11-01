# 52: JSON and config values

**Stopped after the full oracle. This branch is not ready for integration.** A real server completion response contradicts the harness LanguageService return contract; the required true check changes two own callers. See the counterexample below.

Adamic task r2628wp. Main is `470e16f8315f7a83c5c7b7f6f2a77d8328db0a25`. Upstream is TypeScript 6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`. Number 41 is occupied; the requested next free number is 52. This adaptation changes JavaScript.

Fourteen public declaration lines lose any. Three caller inputs, both raw members, five protocol slots, plugin config and its change hook, and the localized-messages hook become unknown. convertToObject returns `AdamicJsonRecoveryValue | undefined`. The two existing adaptation-43 definitions become public, preserving their names and recursive domain:

```ts
export type AdamicJsonRecoveryValue = string | number | boolean | null | AdamicJsonRecoveryValue[] | AdamicJsonRecoveryObject;
export interface AdamicJsonRecoveryObject { [key: string]: AdamicJsonRecoveryValue | undefined; }
```

The undefined property value is real parser recovery: invalid object values are retained; invalid array elements are filtered. The value alias itself does not include undefined; the converter's return additionally permits it.

## Baseline-pinned public any

`public-any-pinned.json` names readConfigFile.config and parseConfigFileTextToJson.config. Their honest shape is the published AdamicJsonRecoveryObject. Their contract stays any only because a pinned baseline holds it there. APISample_parseConfig reads exactly `config.config["compilerOptions"]`, virtual line 27, case-input line 41. The expected absence of `APISample_parseConfig.errors.txt` is part of that judged baseline; narrowing the parse return creates it. readConfigFile forwards the same parse result and remains exempt under the ruling.

The Go test `TestJSONValues52PinnedAnySampleReadsRemain` calls t.Parallel and reads the case with git show at stage3/source.json's pin. It invokes neither Go nor a compiler. It fails if an exemption's exact read disappears, refuses added exemption names or a weakened read, and permits entries to be removed. It does not enforce that no other public any survives.

The design's 39 diagnostics split into 29 at the changed unknown outputs and 10 in jsTyping caused solely by narrowing the two config returns. Ruling 2 says nothing is built at those returns' readers. The 10 jsTyping readers are therefore left unchanged, not given redundant adapter checks. `readers.json` accounts for every measured diagnostic and identifies that exception explicitly. This follows the exemption ruling rather than claiming that those 10 diagnostics required runtime edits while their signatures stayed any.

## Reader checks and behavior

The hidden services transfer is src/services/services.ts:1656 in the adapted control:

```ts
maybeSetLocalizedDiagnosticMessages(host.getLocalizedDiagnosticMessages?.bind(host));
```

Upstream strictBindCallApply is false, so bind erases the unknown result. The replacement captures the hook, calls it with the original host as this only when the existing consumer asks for messages, restores unknown at the return boundary, checks the object and every string value, and retains the original dictionary. A witness checks delayed invocation, receiver, one call, object identity and a later dictionary mutation.

Every edit has an exact owner, node kind and before/after text in rules.json. All source plans and their parsed after shapes are validated before any write. A mixed before/after batch is refused, even across files. Every changed file is written once; a second run writes nothing. Cast counts are unchanged in every edited file; this adaptation adds no casts. Earlier adapters and lane manifests are untouched. Adaptation 43 runs before 52; its own guarded signatures are not changed before it runs.

`readers.json` gives each design coordinate and its check class, including the services transfer. Several diagnostics share a checked binding. The classes below name behavior changes for values outside the stated caller invariants. These rejection paths are explicit; a clean oracle alone is not claimed to prove arbitrary external callers compatible.

| Class and sites | Check and preserved behavior | Other values |
| --- | --- | --- |
| Object: commandLineParser raw binding, isSolutionConfig, parseConfig mutation bases; editorServices.ts:2922-2925 | A non-null object or function is checked before indexing. The exact original object is retained, including arrays, function objects and mutations. isSolutionConfig's existing short circuit still runs first. Own config-parser producers return objects. | Primitive values previously boxed or failed later can now throw TypeError at the object check. These are outside the object producer invariant. No such differing own-caller trace is accepted by the oracle. |
| Inherited: commandLineParser.ts:3469-3471 | Checks the extended raw object and every element before mapping an inherited spec array. Existing string arrays and already-set own specs behave identically. ExtendsResult.compileOnSave becomes unknown internally, preserving raw inherited values rather than coercing them to boolean. | Truthy non-array or non-string inherited specs now throw at the check instead of being treated as array-like data or failing later. Such values are outside the string-spec caller invariant. |
| Location: server/session.ts:3504,3507 | Checks file, line, offset and optional project/position fields. Valid requests retain argument identity. The documented internal position-only form is normalized privately with line/offset zero; the handler selects position before those fields, preserving its result. The request itself is unchanged. | Malformed requests outside that schema now fail before the handler. Public arbitrary caller compatibility is not claimed. |
| Response: harness/client.ts:315-316,467 | Checks arrays, required fields and optional fields against the actual completion/reference structures. Checked arrays and items are returned by identity. A missing reference body is still undefined; the private method declaration is corrected to admit it. Witness data comes from stock TypeScript's real language service. | Invalid response bodies now fail before being read or forwarded. The full oracle disproves the claimed completion-details producer invariant: two legal protocol responses contain string tag text. This class is blocked, not accepted. No generic assertion is used to conceal the mismatch. |
| Plugin: harness/harnessLanguageService.ts:630,682 | Checks the config object and its string message at each fixture read. Configured message strings retain their text and timing. | Non-object configs and non-string fixture messages now throw. They are outside the string-message fixture invariant. |
| Localized: services/services.ts:1656 | Checks the unknown host result and its string-valued dictionary, preserving binding, laziness and identity. Undefined remains allowed. | A non-dictionary or non-string-valued host result now throws instead of being stored unchecked. Own localized-message producers supply string dictionaries. |
| Echo: testRunner/unittests/tsserver/session.ts:525 | The HandlerResponse contract admits all non-null values or undefined. These values are forwarded unchanged, including primitives, arrays and objects. | Null now fails at this check; previously a null response would fail later in doOutput's metadata read. Null is outside the echo fixture's response invariant. |

`witnesses.cjs` lifts the actual changed fragments and predicates and executes before/after cases in Node. It checks identities as well as values. `proof.cjs` runs one source mutant per class; each must turn its Node witness red. It also checks a half-applied batch and last-file text drift fail without any writes, and verifies idempotence and zero added casts.

## Public API sanctions

api-sanctions.json has 14 replacement rows and one insertion row containing both full definitions. convertToObject uses kind FunctionDeclaration with exact upstream return type `any`. The insertion anchors after ParseConfigFileHost, kind InterfaceDeclaration, which is where the actual upstream declaration bundler places these definitions. No whole-member workaround or generator modification is used.

Until typescript/api-sanctions-2 lands, proof.cjs applies the rows by hand to a scratch copy of the control declaration, using each row's owner and exact type, and inserts the definitions on new lines at the anchor indentation. The reconstruction matches the generated artifact and the oracle's local api/typescript.d.ts byte for byte (SHA-256 `183fd0568819d08a0a73e9da116394643c2240cbae3e375ffcf3697dfa722f2f`). These future rows are not claimed to be consumed by main's current generator; this is the authorized scratch comparison.

## Measurements

Setup completed successfully with GOPROXY and the requested 1,800-second timeout; its log is on the evidence branch. Control and final full apply run all adaptations in order. Stock checks retain upstream options; no typecheck is skipped. The native measurement uses 99a's 79 sorted src/compiler roots, fixed checker options and prelude. Its comparison preserves full messages, columns and multiplicity; it maps locations only through unchanged source lines, so inserted checks do not manufacture new diagnostic identities.

| Measurement | Control | With 52 |
| --- | ---: | ---: |
| Stock project diagnostics, each of six requested projects | 0 | 0 |
| Native whole-program diagnostics | 112 | 112 |
| New native diagnostics after exact location mapping | | 0 |
| Removed native diagnostics | | 0 |
| Public explicit any tokens | 37 | 23 |
| Full oracle passing | 106,366 | 106,362 |
| Full oracle failing | 1 | 5 |
| Full oracle pending | 0 | 0 |

The control has only the existing public API acknowledgment failure. With 52, two fourslash-server cases fail and produce two extra tsserver log baselines, for four extra failures. The complete moved-baseline list with 52 is:

- `api/typescript.d.ts`
- `tsserver/fourslashServer/completionEntryDetailAcrossFiles01.js`
- `tsserver/fourslashServer/completionEntryDetailAcrossFiles02.js`

`APISample_parseConfig.errors.txt` does not appear in the adapted full run. The isolated source mutant narrows parseConfigFileTextToJson.config, builds successfully, and turns the targeted compiler oracle red: 5 passing, 1 failing, exactly `APISample_parseConfig.errors.txt` newly produced. Thus the named pinned return is exercised, not merely listed.

## Stop: the completion-details contract is false for own callers

Both `tests/cases/fourslash/server/completionEntryDetailAcrossFiles01.ts` and `completionEntryDetailAcrossFiles02.ts` pass a real protocol response whose completion details contain:

```ts
tags: [{ name: "param", text: "p1" }]
```

They explicitly expect that string. The source contracts differ:

- `src/server/protocol.ts:1059` permits JSDocTagInfo.text to be a string or display parts. Its CompletionEntryDetails at :2276 replaces the service tag shape with this protocol shape.
- `src/services/types.ts:1319` permits only SymbolDisplayPart[] for service JSDocTagInfo.text.
- `src/harness/client.ts:117` declares SessionClient implements LanguageService, and :310 declares its result to be the service CompletionEntryDetails. The original any response body conceals the mismatch.
- `src/harness/fourslashImpl.ts:1205` compares the returned tags directly against the expected raw string tags. There is no normalization at this read.

The new predicate is truthful to the declared service type, so it rejects those legal protocol values. A predicate accepting the string as SymbolDisplayPart[] would be false. Normalizing to display parts would change the own callers' explicitly expected tags. Merely widening this private method to protocol.CompletionEntryDetails would cease to satisfy its implements LanguageService contract; a separate harness protocol/service contract repair is needed. That broader decision is left to Adamic rather than hidden behind a cast, a changed test expectation or a new exemption.

`counterexample.cjs` uses the actual oracle response and the exact before/after reader fragments: before returns the same object with text "p1"; after throws TypeError. The fixture and result are retained on the evidence branch as `counterexample-response.json` and `counterexample.json`. The positive Node witnesses still cover the seven classes and each class mutant fails, but their service-generated completion sample did not cover this protocol conversion. The full oracle supplies the discriminating counterexample. No passing-oracle or behavioral-equivalence claim is made for this stopped adaptation.


The report-only public any inventory after adaptation has 23 tokens on 15 declaration lines, including the two pinned config returns. Outside that list are **21 tokens on 13 lines**:

| Excluded class from JSON-DESIGN | Any tokens | Declaration lines |
| --- | ---: | ---: |
| Timer handles and callback tuples | 16 | 8 |
| Nominal marker slots | 2 | 2 |
| Mapped substitution constraint | 1 | 1 |
| Runtime command table | 1 | 1 |
| Opaque compressed host data | 1 | 1 |

No enforcement of removing these excluded classes is introduced. Public explicit any falls from 37 tokens to 23. The two baseline-pinned returns account for 2 of those 23.

## Reproduce

Run cloud/setup.sh with `GOPROXY='https://proxy.golang.org|direct'` and `timeout --kill-after=60s 1800`, logging to a file, then source the environment it prints. Build a control at the recorded main without 52, and a final tree with normal stage3/apply.sh. Install the pinned dependencies and run npm run build in both. Build the main Adamic binary outside tests with go build. Then:

```sh
NODE_PATH=stage3/api/node_modules node stage3/adapt/52-json-values/proof.cjs CONTROL AFTER EVIDENCE > proof.log 2>&1
python3 stage3/adapt/52-json-values/measure.py CONTROL AFTER EVIDENCE --adamic /absolute/path/to/adamic > measure.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=UPSTREAM go test ./stage3/lane -run '^TestJSONValues52PinnedAnySampleReadsRemain$' -count=1 > pinned.log 2>&1
```

Use an absolute NODE_PATH when the working directory differs. The control and final oracle runs are serial, all runners, no filter, four workers, and no concurrent setup, build or measurement workload. Logs, diagnostic streams, hashes and oracle reports belong on branch `typescript/json-values-evidence`, under this directory's evidence subdirectory. They are deliberately excluded from the adaptation branch.
