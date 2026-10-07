# Google Font Display partial port

The owned .a rule contains the complete diagnostic decision, first matching literal href semantics, case-insensitive attribute matching, intrinsic link discrimination, query order and XHTML entity decoding. It is registered for both JSX opening forms using the numeric child layout published on origin/codex/stage1-jsx-lint. That parser branch is not integrated here; no shared source was edited.

`validate-selector.py --scratch /tmp/wave07-font --compiler /tmp/wave07-typescript` compares messages against the unmodified Go rule parsing real JSX. All 16 upstream source cases plus 11 query/entity cases match on source Node, emitted JavaScript and ASan/UBSan native: 27 rows, 4,228 identical bytes. Adamic receives externally extracted attribute data in this bounded selector comparison. It does not prove independently parsed JSX, element ranges or complete rule parity.

The block-display-accepted mutant compiles and executes successfully on all three Adamic runtimes, and only the Go byte comparison rejects its missing diagnostic. Clang's default nesting limit initially rejected the 252-case entity switch; bounded owned lookup functions fixed compilation. That failure is not credited as a mutant.

The full owned profile builds. A real `<link href="https://fonts.googleapis.com/css2?family=Krona+One" />` gives a positive Go diagnostic but source Node, emitted JavaScript and sanitized native each exit 70: `parser slice expected GreaterThanToken, got Identifier at 32`. See evidence/full-parser-results.txt. JSX tree construction blocks full findings/ranges, compiler/stage1/upstream whole-rule comparisons and comparable whole-rule throughput. It is not replaced with a source-text heuristic. The shared harness also only discovers `.ts.txt` witnesses and routes them through a TS filename; JSX witness support remains an integration requirement.

Selector-only rates including startup, best of five 1,000-finding runs: native 122,721.49, Node 14,690.49, Go 84,020.24 findings/s. Go parses JSX while Adamic receives attribute data; these rates do not establish a whole-rule performance advantage.

Run the full profile with WAVE07_PROFILE pointing to profile.a and WAVE07_ARTIFACTS pointing to a scratch directory, then `go test ./stage1/cohere/lint/rules/next-google-font-display -run '^TestCompileProfiles$' -count=1 -v`. oracle/node.mjs executes either profile.a or the emitted.mjs artifact. Run native with the same manifest.

All upstream Google Font Display tests pass (0.010s). The registry, touched Go packages and filtered uncached input oracle pass. The full repository gate was not run. Setup initially failed because a new descriptor was observed before its witness existed; after adding owned witnesses, setup succeeded in 20s on 5 processors. No new rules were claimed during this completion.
