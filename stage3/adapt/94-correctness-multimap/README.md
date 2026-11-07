# Additive MultiMap helpers before 95

Rule class: JavaScript-changing composition, task b5w3ycg, wave row 7. The approved split first adds exported internal multiMapAddValue and multiMapRemoveValue alongside the legacy private helpers. Their bodies preserve the legacy logic, with this replaced by a Map<K, V[]> parameter and explicit return annotations. Existing functions and call sites are unchanged at this step.

apply.py sorts directory paths lexically: 94-axolotl-property-read, 94-correctness-multimap, 95-checker-axolotl-a. Thus the helpers exist before 95h rewrites its two createProgram calls and before subsequent semantic adapters inspect the intermediate tree.

pin.json pins the complete multiMapRemove insertion boundary through the next createQueue declaration, including the exact added helpers. The adapter plans before writing, refuses drift or partial insertion, preserves LF/CRLF, and replays without writes. It also recognizes the fully finalized 99 tree, checking both exact helpers and the final MultiMap alias.

95h owns its two calls; 99-correctness-core-multimap owns every other call and the final plain Map representation. Interim names remain because a later rename would edit 95h's pinned createProgram. Count, witnesses, complete apply, replay, emitted scope, declarations, types, full site census, and mutant evidence live in ../99-correctness-core-multimap. The completed results and authenticated evidence are indexed in ../99-correctness-core-multimap/PROOF.md.
