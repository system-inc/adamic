# Completion-details harness reach

Main remains `470e16f8315f7a83c5c7b7f6f2a77d8328db0a25`. The reach was measured against prior implementation `d3aeaecadaef1fdf57640ea14ca56458c5c5f4b5`. This records the scratch reach measurement subsequently authorized and implemented in rules.json. The jsTyping readers were unchanged at this historical harness measurement. The later own-pin correction checks all ten readConfigFile readers, as recorded in README.md.

## A second full-reply distinction

`src/harness/client.ts:428` in the scratch candidate requests `CompletionDetailsFull`, not the simplified command. In the adapted control, `src/server/session.ts:2567` implements that command. Its fullResult arm at :2578 retains service code actions and only maps tags when displayPartsForJSDoc is false. The simplified arm additionally maps code actions.

`src/server/protocol.ts:2276` declares CompletionEntryDetails with protocol CodeAction[] at :2278. Protocol CodeAction at :2142 contains FileCodeEdits and line/offset CodeEdit values. Service CodeAction has numeric-span TextChange values. Therefore the protocol alias does not describe the full command's returned code actions. Using it literally sends the mismatch into fourslash's existing numeric-span readers at :3362, :3369, :3708 and :3711.

The literal protocol return and a common native/protocol service interface produced 12 stock diagnostics, all in src/harness/fourslashImpl.ts: 1205, 2511, 2512 (two), 2518, 3362, 3369, 3708 (two), 3711 (three). No public declaration or testRunner edit was made.

## Measured and authorized precise full-reply shape

This harness-local type accurately describes the full reply while admitting every legal protocol tag value:

```ts
export type HarnessCompletionEntryDetails = Omit<CompletionEntryDetails, "tags"> & {
    tags?: readonly protocol.JSDocTagInfo[];
};
export interface HarnessLanguageService extends Omit<LanguageService, "getCompletionEntryDetails"> {
    getCompletionEntryDetails(...args: Parameters<LanguageService["getCompletionEntryDetails"]>): HarnessCompletionEntryDetails | undefined;
}
```

The native service is structurally assignable. The completion guard retains service code-action checks and changes its tag check to accept string text or checked display parts. The tags are readonly only in the local observation type, allowing the existing deepEqual read with readonly expected tags. No returned value is changed.

The alternative reaches 14 edit locations in three files:

| File | Locations in the previous adapted candidate | Edits |
| --- | --- | ---: |
| src/harness/client.ts | tag predicate :114; details predicate :147; two new local declarations before SessionClient :232; class implements :232; method return :425 | 6 |
| src/harness/harnessLanguageService.ts | adapter interface :125; server adapter method :805 | 2 |
| src/harness/fourslashImpl.ts | languageService field :246; memoWrap signature :452; proxy annotation :460; details helper return :1793; baselineCompletions result annotation :2484; linked-edit helper parameter :3779 | 6 |

The result annotation exposes only what baselineCompletions reads. Its tag text is Iterable<{ kind?: string; length?: number }>, which both strings and display-part arrays satisfy. String characters have length and no kind, so the existing loop skips them exactly as it does today. No guard or normalization is added to that loop. The existing proxy assertion is left verbatim; its new variable annotation carries the flow type. No new cast is introduced.

Stock TypeScript 6.0.3 `tsc -b src --force --pretty false`, retaining all upstream project options, exits 0 for this scratch alternative. Its compiler log is empty. Emitted JavaScript for the adapters and fourslash is byte-identical. Client JavaScript is byte-identical after accounting for exactly the added string arm in the tag guard. No expectation, public source declaration, adaptation rule or testRunner source was edited.

This is within the requested size and harness boundary. The user authorized the precise full-reply shape after this measurement. It is now implemented in the adaptation, with identity witnesses and two new check mutants. README.md records the final measurements and oracle result. Evidence retains both historical reach runs and the final proof.
