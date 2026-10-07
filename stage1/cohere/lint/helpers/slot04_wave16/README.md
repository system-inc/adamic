# Utility resolution helpers

Each `.a` file contains one claimed Go helper port:

- `modifier_value.a`: `modifierAsValue` maps explicit nil presence to an absent value, converts only `arbitrary` modifiers to arbitrary values, and clears fraction and typehint. Unknown modifier kinds become named, as Go does.
- `arbitrary_argument.a`: `resolveArbitraryArgument<T>` accepts an already bracketed argument, uses wildcard first, trusts an existing typehint, and otherwise asks the supplied inference function. It returns the supplied parser's nodes unchanged on success, and an empty node list with `ok: false` on failure. Go nil slices are represented by empty lists; the explicit success flag preserves resolution failure.
- `variable_reference.a`: `variableReference` looks up the stored key before calling prefixing and escaping, then emits a fallback only for a nonempty value carrying reference bit 2. An absent key returns an empty string with `ok: false`.

ParseValue, InferDataType, PrefixKey and escapeCSSIdentifier remain separately owned dependencies. Their callbacks must implement the corresponding Go behavior. This batch does not substitute JS parsing or infer a different grammar. The generic node type allows callers to retain their own value-node arena and identity representation. The modifier view's presence flag distinguishes nil from an explicitly present empty modifier.

Run from the repository root with the setup environment sourced:

```
python3 stage1/cohere/lint/helpers/slot04_wave16/testdata/generate.py > /tmp/slot04-wave16-generate.log 2>&1
python3 stage1/cohere/lint/helpers/slot04_wave16/testdata/capture.py > /tmp/slot04-wave16-capture.log 2>&1
go test -count=1 -v ./stage1/cohere/lint/helpers/slot04_wave16 > /tmp/slot04-wave16-tests.log 2>&1
```

The private Go oracle is added through temporary overlays. Consumer capture also uses temporary overlays, leaving the shared harness and pinned cohere worktree unchanged. The differential driver supplies real Go inference and parser projections; parsed nodes are projected through ValueToCss for comparison. This verifies resolver decisions and printed values, not arbitrary parser implementations or every node field. Valid Unicode strings are covered; malformed UTF-8 and malformed unbracketed arguments are outside the caller contract exercised here.
