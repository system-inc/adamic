### unittests:: sys:: symlinkWatching: (exception)

```text
with ts.sys::
         watchFile using polling
           watchFile using polling:

      AssertionError: expected 'export const x = 100;' to equal 'export const x = 10;'
      + expected - actual

      -export const x = 100;
      +export const x = 10;
      
      at Context.<anonymous> (src/testRunner/unittests/sys/symlinkWatching.ts:48:49)



file:///workspace/scratch/step34-artifacts/e41a430a94a8b7b011d31cd1c9aaabfda03d4f8c9dc3f67d13eb8859aa434b24/tree/scripts/build/utils.mjs:42
                        new ExecError(exitCode);
                        ^

ExecError: Process exited with code: 1
    at ChildProcess.<anonymous> (file:///workspace/scratch/step34-artifacts/e41a430a94a8b7b011d31cd1c9aaabfda03d4f8c9dc3f67d13eb8859aa434b24/tree/scripts/build/utils.mjs:42:25)
    at ChildProcess.emit (node:events:509:28)
    at ChildProcess._handle.onexit (node:internal/child_process:295:12) {
  exitCode: 1
}

Node.js v24.19.0
```

Stack top: at Context.<anonymous> (src/testRunner/unittests/sys/symlinkWatching.ts:48:49)

### unittests:: sys:: symlinkWatching:: with ts.sys:: watchFile using polling watchFile using polling (exception)

```text
expected 'export const x = 100;' to equal 'export const x = 10;'
```

Elapsed: 1574.312 ms; timeout: 40000 ms.

Stack top:     at Context.<anonymous> (src/testRunner/unittests/sys/symlinkWatching.ts:48:49)
