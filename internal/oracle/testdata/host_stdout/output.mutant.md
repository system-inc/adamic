# stdout write oracle evidence (#grqdgpr)

This backs the stdout-only case added to `TestClosedStdoutEndsAsOnNode`,
the five registered `host_stdout/*.a` fixtures, and their Node, sanitized
native, JavaScript and WASI comparisons. No recorded status is edited.

The runtime mutants copy the real runtime into a temporary directory, compile
it successfully under ASan/UBSan, then run the fixture. A build failure is not
a catch. They do not edit the working tree.

| Mutant | Re-proving test | Comparison that goes red |
|---|---|---|
| Raw output bypasses console buffering | TestStdoutWriteMutantBypassesOrdering | TestStdoutWriteOrderingMatchesNode |
| Raw output passes WTF-8 lone surrogates unchanged | TestStdoutWriteMutantPassesLoneSurrogates | TestStdoutWriteTextMatchesNode |
| Raw writer does not register output teardown | TestStdoutWriteMutantOmitsExitHook | TestStdoutWriteClosedPipeMatchesNode: exit 0 instead of 70 |
| Accept the write's boolean as a value | stdout-mutants.py overlay | TestNodeStdoutWriteRefusals |

With runtime products prepared first:

```sh
ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -count=1 -run '^TestStdoutWrite|^TestWASIAgreesWithNode_017$' -v
python3 internal/lower/testdata/stdout-mutants.py
```

The closed-pipe witness has no console call: nothing else can register the
exit hook for it. The ordering witness prints a buffered console line before
each raw write, so bypassing the shared buffer visibly reverses the bytes.
