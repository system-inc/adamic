# Google font display

Ported the URL decision, complete pinned XHTML/numeric entity decoding, messages,
Go adapter, descriptor and JSX listener candidate inside this rule directory.
The generic JSX child layout in the listener is not certified: the shared parser
currently refuses JSX before any visitor executes.

`source /workspace/adamic-tools/env.sh; python3 stage1/cohere/lint/rules/next-google-font-display/validate.py`
compared 276 href cases against the actual unmodified Go GoogleFontDisplay rule
on real TSX. Source Node, emitted JavaScript and ASan/UBSan native each matched
85,014 bytes of diagnostic ids and messages, exited 0 and produced no stderr.
The block-value omission mutant compiled and ran cleanly, then differed from Go
on all three runtimes. The corpus includes duplicate query keys, bare/empty keys,
query ordering, absolute-URL gates, overflow and surrogate numeric entities and
all 253 pinned named XHTML entities. No shared rule body or harness was edited.

Full findings/ranges/fixes parity and findings/second remain blocked, not passed.
The shared parser exits 70 on the real upstream missing-display fixture:
`parser slice expected GreaterThanToken, got Identifier at 32`. See
[the parser log](validation/parser-blocker.log) and
[the decision comparison](validation/decision-comparison.log).

The entry module and imported helpers use the explicitly permitted `.ts`
fallback until shared discovery supports `.a`; the independent probe uses `.a`.
Ordinary lint-package tests also still fail shared profile compilation, and the
shared suggestion serializer blocks the already implemented non-null rules.
This implementation is pushed as a blocked candidate. Work moves on under the
user's instruction to port everything else and identify shared gaps precisely.
