# Wave 7 function-return fixture mutant

The fixture now uses the unique pinned declaration `function getDefaultLibFileName(options: CompilerOptions): string;`. No owner in the 262-declaration manifest claims it. It remains a non-overloaded FunctionDeclaration with an explicit return type, sanctioned as a type-only replacement with an undefined union. The test checks its exact changed signature and passes the generated API through the same declaration guard.

The old convertToObject fixture collided with adaptation 52. No adaptation or production generator was changed.

Scratch copy: `/workspace/scratch/C-function-return-mutant`. Command from its lane directory:

```sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/upstream python3 -B -m unittest test_generate_api.ApiSanctionGeneratorTests.test_api_sanction_function_return_applied_and_guard_accepts -v
```

The unmodified relocated test passed. The scratch mutant silently discards FunctionDeclaration return-type edits while retaining other sanctioned edits:

```diff
--- scratch/generate-api.cjs
+++ scratch/generate-api.cjs
@@ -302,7 +302,7 @@
             `${label}: upstream text ${JSON.stringify(target.getText(source))}; before ${JSON.stringify(row.before)}`);
         const edit = { at: target.getStart(source), end: target.end, text: row.after };
         sanctionEdits.push(edit);
-        edits.push(edit);
+        if (row.kind !== 'FunctionDeclaration') edits.push(edit);
     }
     proofPaths.push(relative);
 }
```

It exited 1 with exactly one assertion failure and no errors:

```text
test_api_sanction_function_return_applied_and_guard_accepts (test_generate_api.ApiSanctionGeneratorTests.test_api_sanction_function_return_applied_and_guard_accepts) ... FAIL

======================================================================
FAIL: test_api_sanction_function_return_applied_and_guard_accepts (test_generate_api.ApiSanctionGeneratorTests.test_api_sanction_function_return_applied_and_guard_accepts)
----------------------------------------------------------------------
Traceback (most recent call last):
  File "/workspace/scratch/C-function-return-mutant/stage3/lane/test_generate_api.py", line 228, in test_api_sanction_function_return_applied_and_guard_accepts
    self.assertTrue('function getDefaultLibFileName(options: CompilerOptions): string | undefined;'
AssertionError: False is not true : sanctioned function return was not emitted

----------------------------------------------------------------------
Ran 1 test in 2.614s

FAILED (failures=1)
```

The scratch generator was restored byte for byte. The same targeted test passed again. Full positive, mutant and restored logs are retained in the scratch copy.
