# Binding views

TypeScript pin: 050880ce59e30b356b686bd3144efe24f875ebc8 (v6.0.3).
Base Adamic commit: 7534c1d7575db02bb7b6512119e0685987b7e333.

Apply scout ty5myjt's `resolution-readonly.patch`, with the file's established
`/** @internal */ export interface` convention: introduce an internal
`ResolutionReadView extends ResolutionWithResolvedFileName` with readonly
`resolvedFileName: string | undefined`, and use it in R's constraints and default.
The original interface stays unchanged. The wrapper T, runtime bodies, calls,
ordering and values stay the same.

The companion diagnostics overloads are an **amendment to
[95-checker-axolotl-h](../95-checker-axolotl-h/README.md)**, whose whole
`createProgram` pin encloses `getDiagnosticsHelper`. Its `diagnostics-overloads.json`
pins the complete helper overload group and complete enclosing owner before and
after that adaptation. It applies the scout's exact non-generic
DiagnosticWithLocation and Diagnostic overloads over the same implementation.
There is no second adapter claiming that enclosing owner here.

The first ownership grep covered both files under stage3/adapt. No earlier
adaptation pins any of the edited resolution declarations whole. `owners.json`
pins the complete ResolutionCache interface, GetResolutionWithResolvedFileName
alias and createResolutionCache function, plus the inserted view. A stricter
complete-file text and SHA-256 pin also covers all surrounding resolution bytes.
The final program.ts pin covers the unchanged original interface there. `mutant-owners.json` pins both complete final files for exact
restoration. Unknown text, duplicate declarations, partial adaptation, mixed line
endings and source changes during planning fail before writes. CRLF resolution
input is preserved. A replay writes zero files. Both JavaScript comment modes
are asserted before any type-only write.

`node adapt.cjs TREE` applies the resolution portion after the 95h amendment.
`node mutant.cjs TREE` restores both original final files. The mutant is a
measurement tool, not part of normal apply. The requested type-only ledger line
in stage3/upstream/LEDGER.md cites #ty5myjt and Compiler's accept.

The initial unexported helper changed declaration export modifiers. A two-half
bisect isolated resolutionCache.ts; the internal exported form gives byte-identical
upstream declarations.

[PROOF.md](PROOF.md) records the proof protocol and evidence destination. Commands, logs, censuses, declaration
outputs and contract controls belong on the
[evidence branch](https://github.com/system-inc/adamic/tree/typescript/binding-views-evidence/stage3/census/binding-views).
Nothing is submitted to Loom.
