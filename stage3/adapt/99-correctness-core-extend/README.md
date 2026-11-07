Current recut: origin/main 698965fe66aed945d18efa320a594a8a8b4df2f8.
The earlier f0271c6a7 results below are historical until replaced by the recut proof.

# extend and copyProperties

Task b5w3ycg, wave row 6. Type-only candidate on origin/main
f0271c6a7. Runtime shape measurement and complete proof are in progress.

The candidate changes only the complete owners in compiler/core.ts. It removes
F6 through F10 in extend and F11 in copyProperties, whose original refusal is
"a cast the runtime can't check" at cast_proof.go:188.

extend now takes and returns string-indexed records whose values may be absent.
The empty record and both indexed loops need no cast. Its two compiler callers
already pass CompilerOptions or WatchOptions records. The return no longer
claims that an enumerable-own-string copy constructs every member of arbitrary
T1 & T2, including symbols and non-enumerable properties.

copyProperties gives both arguments the same type parameter. The for-in key is
therefore a key of the writable target and source. Both compiler callers pass a
WatchHost source and a host that extends WatchHost. No caller is edited.

The emitted JavaScript is pinned byte-identical by the adapter, with comments
kept and removed. The six any casts disappear without an unchecked replacement.
No JavaScript-changing road is implemented, so no ledger entry is required.
The exact full owner text and hashes are in owners.json. A changed, missing,
duplicate, nested or mixed owner refuses before any write. Replay performs no
writes and preserves bytes and mtime. LF, CRLF and unrelated text survive.

The baseline ownership search found only the historical declined-sites inventory
in adaptation 41, which does not pin or edit these owners. No prior owner is
amended. The pinned upstream is TypeScript 6.0.3,
050880ce59e30b356b686bd3144efe24f875ebc8.
