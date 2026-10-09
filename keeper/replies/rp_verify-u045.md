M06 is caught by the new fixture with an ASan heap-use-after-free, but M08 survives because `TestStatementRegionsAreUsed` reads the unchanged counts golden rather than measuring the compiler. All six runs’ wall times and logs are pushed under `review/test-audit/u045-verify/` on `test-audit/u045-verify`, and the tree is restored.

```json
{
  "main": "946a8f095a",
  "M06": {
    "test": "TestNativeAgreesWithNode/internal/oracle/testdata/borrow_chain_captured_root.a",
    "base": 0,
    "mutant": 1,
    "base2": 0,
    "failing_line": "oracle_test.go:759: the release build: exit codes differ",
    "applied": "clean"
  },
  "M08": {
    "test": "TestStatementRegionsAreUsed",
    "base": 0,
    "mutant": 0,
    "base2": 0,
    "failing_line": null,
    "applied": "clean"
  }
}
```
