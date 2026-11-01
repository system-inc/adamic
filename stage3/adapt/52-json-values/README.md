# 52: JSON and config values

The completion-details read uses the precise harness-local full-reply shape authorized after the reach measurement. It preserves protocol string tags and service numeric-span code actions by identity. All harness flow edits erase to identical JavaScript; the only runtime change in this repair is accepting checked string tag text.

Adamic task r2628wp. Main is `470e16f8315f7a83c5c7b7f6f2a77d8328db0a25`. Upstream is TypeScript 6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`. Number 41 is occupied; the requested next free number is 52. This adaptation changes JavaScript.

Fifteen public declaration lines lose any. Three caller inputs, both raw members, five protocol slots, plugin config and its change hook, and the localized-messages hook become unknown. convertToObject returns `AdamicJsonRecoveryValue | undefined`. readConfigFile.config becomes AdamicJsonRecoveryObject. The two existing adaptation-43 definitions become public, preserving their names and recursive domain:

```ts
export type AdamicJsonRecoveryValue = string | number | boolean | null | AdamicJsonRecoveryValue[] | AdamicJsonRecoveryObject;
export interface AdamicJsonRecoveryObject { [key: string]: AdamicJsonRecoveryValue | undefined; }
```

The undefined property value is real parser recovery: invalid object values are retained; invalid array elements are filtered. The value alias itself does not include undefined; the converter's return additionally permits it.

## Baseline-pinned public any

`public-any-pinned.json` now names only parseConfigFileTextToJson.config. Its honest shape is the published AdamicJsonRecoveryObject. Its contract stays any only because a pinned baseline holds it there. APISample_parseConfig calls `ts.parseConfigFileTextToJson(` at case line 35 and reads exactly `config.config["compilerOptions"]` at case line 41, virtual line 27. The expected absence of `APISample_parseConfig.errors.txt` is part of that judged baseline; narrowing the parse return creates it.

The sample never calls readConfigFile, and tests/cases contains no readConfigFile call. Forwarding the parse result does not give readConfigFile its own pin. Its public config return therefore narrows to AdamicJsonRecoveryObject, and all of its affected jsTyping readers receive real checks. parseConfigFileTextToJson retains any and its own readers remain unchanged under the exemption ruling.

The Go test `TestJSONValues52PinnedAnySampleReadsRemain` calls t.Parallel and reads the case with git show at stage3/source.json's pin. It invokes neither Go nor a compiler. An entry must have both its own `ts.<function>(` call and its exact read in the sample. The list may shrink to zero, has at most one entry, and cannot admit readConfigFile or weaken the baseline/read. `TestJSONValues52ExemptionsRequireTheirOwnCall` is also parallel and pins the distinction: the parse call and its read cannot justify readConfigFile, and neither the call alone nor the read alone is enough. It runs no subprocess.

All ten jsTyping diagnostics in the design are readConfigFile readers. There are no parseConfigFileTextToJson calls in jsTyping, so none of those ten readers remains exempt. `readers.json` accounts for all 39 measured diagnostics and the hidden services transfer.

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
| Response: harness/client.ts:315-316,467 | Checks arrays, required fields and optional fields against the actual completion/reference structures. Checked arrays and items are returned by identity. A missing reference body is still undefined; the private method declaration is corrected to admit it. Witness data comes from stock TypeScript's real language service. | Invalid response bodies now fail before being read or forwarded. CompletionDetailsFull retains service code actions and permits protocol string tags. Its local declaration and check describe that exact full reply; legal string text and display parts are accepted without normalization. |
| Plugin: harness/harnessLanguageService.ts:630,682 | Checks the config object and its string message at each fixture read. Configured message strings retain their text and timing. | Non-object configs and non-string fixture messages now throw. They are outside the string-message fixture invariant. |
| Typing string maps: jsTyping.ts:81,88 | Checks the actual Object.entries values are strings, retaining string input indexing and string-array entries. loadTypesMap keeps its original truthy branch. | Dictionaries or arrays with non-string entries now throw instead of making a falsely string-typed SafeList. A truthy numeric or boolean simpleMap now throws instead of producing an empty map. Undefined safe-list input still throws TypeError. Own map fixtures supply string maps; these invalid values are outside that invariant. |
| Typing dependency keys: jsTyping.ts:213-215,245 | Checks the manifest object before property reads. Each dependency map is checked before getOwnKeys. Strings retain their character-index keys; missing, null, boolean and numeric maps retain their empty key lists. | Object checks can reject absent manifest objects earlier; readConfigFile's parser always supplies an object. Key enumeration is unchanged for every JSON value, including recovery values. |
| Typing package fields: jsTyping.ts:271-290 | Checks the package object, keeps name/types/typings unknown locally, and checks only truthy names and the selected truthy types path immediately before string use. All falsey name skips and types fallback behavior remain unchanged. | Truthy non-string names or selected types paths now throw at the check. These violate the string consumer contract and are outside the own valid-package fixtures. Unused dependency fields are not validated. |
| Localized: services/services.ts:1656 | Checks the unknown host result and its string-valued dictionary, preserving binding, laziness and identity. Undefined remains allowed. | A non-dictionary or non-string-valued host result now throws instead of being stored unchecked. Own localized-message producers supply string dictionaries. |
| Echo: testRunner/unittests/tsserver/session.ts:525 | The HandlerResponse contract admits all non-null values or undefined. These values are forwarded unchanged, including primitives, arrays and objects. | Null now fails at this check; previously a null response would fail later in doOutput's metadata read. Null is outside the echo fixture's response invariant. |

`witnesses.cjs` lifts the actual changed fragments and predicates and executes before/after cases in Node. It checks identities as well as values. `proof.cjs` runs one source mutant per original class, two mutants of the tag-text check (reject legal strings, accept invalid text), and four jsTyping check mutants; each must turn its Node witness red. It also checks a half-applied batch and last-file text drift fail without any writes, and verifies idempotence and zero added casts.

## Public API sanctions

api-sanctions.json has 15 replacement rows and one insertion row containing both full definitions. convertToObject uses kind FunctionDeclaration with exact upstream return type `any`. The insertion anchors after ParseConfigFileHost, kind InterfaceDeclaration, which is where the actual upstream declaration bundler places these definitions. No whole-member workaround or generator modification is used.

Until typescript/api-sanctions-2 lands, proof.cjs applies the rows by hand to a scratch copy of the control declaration, using each row's owner and exact type, and inserts the definitions on new lines at the anchor indentation. The reconstruction matches the generated artifact and the oracle's local api/typescript.d.ts byte for byte (SHA-256 `b8faba99305b76f8f6ea0273807b08dc8ef6b9cada46d563e436f1d943e9a754`). These future rows are not claimed to be consumed by main's current generator; this is the authorized scratch comparison.

## Measurements

Setup completed successfully with GOPROXY and the requested 1,800-second timeout; its log is on the evidence branch. Control and final full apply run all adaptations in order. Stock checks retain upstream options; no typecheck is skipped. The native measurement uses 99a's 79 sorted src/compiler roots, fixed checker options and prelude. Its comparison preserves full messages, columns and multiplicity; it maps locations only through unchanged source lines, so inserted checks do not manufacture new diagnostic identities.

| Measurement | Control | With 52 |
| --- | ---: | ---: |
| Stock project diagnostics, each of six requested projects | 0 | 0 |
| Native whole-program diagnostics | 112 | 112 |
| New native diagnostics after exact location mapping | | 0 |
| Removed native diagnostics | | 0 |
| Public explicit any tokens | 37 | 22 |
| Full oracle passing | 106,366 | 106,366 |
| Full oracle failing | 1 | 1 |
| Full oracle pending | 0 | 0 |

The control and repaired adaptation each have only the standing public API acknowledgment failure. The repaired full oracle has 106,366 passing, 1 failing and 0 pending. Its complete moved-baseline list is:

- `api/typescript.d.ts`

Neither completion-details server baseline moves. The two previously failing cases retain their exact expected string tags. The fresh full apply and the built oracle tree match byte for byte throughout src. Hand-applied sanctions, the generated declaration and the oracle declaration match byte for byte. The incremental API diff remains 135 lines (+19/-15).

`APISample_parseConfig.errors.txt` does not appear in the adapted full run. The isolated source mutant narrows parseConfigFileTextToJson.config, builds successfully, and turns the targeted compiler oracle red: 5 passing, 1 failing, exactly `APISample_parseConfig.errors.txt` newly produced. Thus the named pinned return is exercised, not merely listed.

## Completion-details full reply and harness flows

The harness requests completionEntryDetails-full. In the adapted control, src/server/session.ts:2578 preserves service code actions with numeric-span edits while mapping tags to protocol strings when displayPartsForJSDoc is false. The protocol.CompletionEntryDetails alias instead describes simplified line/offset code actions. The authorized precise harness-local type is therefore:

```ts
export type HarnessCompletionEntryDetails = Omit<CompletionEntryDetails, "tags"> & {
    tags?: readonly protocol.JSDocTagInfo[];
};
export interface HarnessLanguageService extends Omit<LanguageService, "getCompletionEntryDetails"> {
    getCompletionEntryDetails(...args: Parameters<LanguageService["getCompletionEntryDetails"]>): HarnessCompletionEntryDetails | undefined;
}
```

SessionClient implements that interface and declares the precise full-reply return. Native LanguageService remains structurally compatible. The adapter interface, server adapter, fourslash field, memoized proxy, details helper and linked-edit helper carry this local type. The completion baseline result has a type-only observation annotation: tag text is iterable, and its elements may have kind or length. Display parts have kind; string characters have length and no kind, so the original loop retains its exact behavior. The existing proxy assertion stays verbatim, with a variable annotation added. No cast is added, no public declaration is changed by these harness types, and no test expectation is changed.

HARNESS-REACH.md lists all 14 edit locations across the three harness files. There are no testRunner flow edits. The declaration-flow proof removes those type edits from the complete source and compares stock emitted JavaScript: all three files match byte for byte with their runtime checks retained. The sole runtime difference in this repair is the string arm of the tag check. The array and property checks accept every legal full-reply tag value and retain all original values, including numeric-span code actions.

`counterexample.cjs` now pins the previously failing real server response: both before and after return the same object, tags array and string text "p1". Node witnesses additionally cover absent text, empty strings, empty and populated display-part arrays, and nested numeric-span code actions, checking identity throughout. A mutant deleting the string arm rejects the legal full reply. A mutant bypassing the text check accepts malformed values; both make their respective witnesses fail. Invalid null, numeric, boolean, object and malformed-array text is rejected. Such invalid text is outside the legal full-reply shape and is the only new rejection domain discussed here.

The report-only public any inventory after adaptation has 22 tokens on 14 declaration lines, including the one pinned config return. Outside that list are **21 tokens on 13 lines**:

| Excluded class from JSON-DESIGN | Any tokens | Declaration lines |
| --- | ---: | ---: |
| Timer handles and callback tuples | 16 | 8 |
| Nominal marker slots | 2 | 2 |
| Mapped substitution constraint | 1 | 1 |
| Runtime command table | 1 | 1 |
| Opaque compressed host data | 1 | 1 |

No enforcement of removing these excluded classes is introduced. Public explicit any falls from 37 tokens to 22. The one baseline-pinned return accounts for 1 of those 22.

## Reproduce

Run cloud/setup.sh with `GOPROXY='https://proxy.golang.org|direct'` and `timeout --kill-after=60s 1800`, logging to a file, then source the environment it prints. Build a control at the recorded main without 52, and a final tree with normal stage3/apply.sh. Install the pinned dependencies and run npm run build in both. Build the main Adamic binary outside tests with go build. Then:

```sh
NODE_PATH=stage3/api/node_modules node stage3/adapt/52-json-values/proof.cjs CONTROL AFTER EVIDENCE > proof.log 2>&1
python3 stage3/adapt/52-json-values/measure.py CONTROL AFTER EVIDENCE --adamic /absolute/path/to/adamic > measure.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=UPSTREAM go test ./stage3/lane -run '^TestJSONValues52(PinnedAnySampleReadsRemain|ExemptionsRequireTheirOwnCall)$' -count=1 > pinned.log 2>&1
```

Use an absolute NODE_PATH when the working directory differs. The control and final oracle runs are serial, all runners, no filter, four workers, and no concurrent setup, build or measurement workload. Logs, diagnostic streams, hashes and oracle reports belong on branch `typescript/json-values-evidence`, under this directory's evidence subdirectory. They are deliberately excluded from the adaptation branch.
