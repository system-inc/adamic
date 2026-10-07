# Raw runtime context question

The shared dispatcher has one import and one `runtime-context` arm. Implementation
and native decoding live in this owned directory. The dispatcher validates the
exact source/span/kind and leases the checker before calling the adapter.

Questions have exactly two newline-separated words: `runtime-context` and a mode.
The response uses the existing bridge framing: decimal UTF-16 length, newline,
then that many UTF-16 units of text, carried over C as UTF-8. Its header is schema
`1`, `runtime-context`, then the requested mode. All records are raw compiler
observations. No lint verdicts, control-flow graphs, messages, repairs or
suggestions are calculated in Go.

- `origin`: the symbol at the anchor, preserving import alias identity, followed
  by flags, name and declarations.
- `alias-origin`: the same facts after following one import alias. Native code
  chooses whether following it is appropriate for a particular rule.
- `signature`: a resolved call signature's declaration and return-type flags.
  This mode is prepared for the incomplete process-output port.
- `program`: source files and import/export/dynamic import/require resolutions.
  This mode is prepared for the incomplete blocking-streams port.

Declarations include source path, AST kind and byte positions, declaration-file,
default-library and external-module flags, source text, node/function flags, and
ancestors with kinds, names, flags, byte positions, global-augmentation flags and
name kinds. Program edges include kind, type-only/computed flags, specifier and
resolved target. AST positions are compiler byte positions; wire frame lengths
are UTF-16 units. These are different quantities.

The caller's checker lease is reused. No new type handles or owned addresses are
exposed. The archive's released-handle guard runs before dispatch and is tested
with this question and a registry-retention mutant.
