# Bounded matrix selection

M1 and M2: Jsx.identifier, reached by member/name rejection fixtures and jsxManifest fixtures. M3: Jsx.text, reached by jsxManifest child fixtures. The conservative JSX matrix runs both rejection rows, TestJsxNode and TestJsxNative, including rows whose negative fixtures do not call text. TestJsxScannerMutants uses the same fixtures, but its clean precondition failures are recorded separately from witness verdicts.

M4: countTree, called by main.run only with --count. In the requested slice that means TestPerformance, TestWholePerformance, and the clean precondition of TestNodeCountCheckCatchesMutant. The witness's precondition is not a production kill.

P1: main.run, the executable entry processing each source. The seven output/rejection rows call this entry. Witnesses are probed only as clean preconditions, which cannot prove the witness comparator. Setup constructs a binary without invoking run and has no run-entry probe.

TestExpressionsAgree exercises expression parser code with .ts inputs rather than JSX. It is evaluated with W1, as its body expects three built-in mutants to disagree. Function inventory is static and intentionally conservative; no claim of dynamic coverage is made.

Commands are logged in run-status.json. Each bounded matrix cell runs one top-level row to avoid one Go timeout aborting later observations. The original full-package timeout is in baseline.log. Kills outside these explicitly named rows remain unknown.
