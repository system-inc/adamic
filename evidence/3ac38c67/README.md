# A–C at 3ac38c67

Source: `3ac38c67f2745703f82644d9116240a03ec48e64`, the user-requested replacement for f95deb36 focused checks.

| Step | Exit | Result |
|---|---:|---|
| A gofmt / vet | 0 / 0 | Empty formatting output; vet passes. |
| B generic-method and NodeWatch pins | 1 | Chdir and exit both fail for .a and .ts at load. All other generic-method subtests and all NodeWatch tests pass. |
| C remove ` && l.generic(source)` | 1 | Same four rows red at load; no green-to-red proof. |
| Restore byte-for-byte / Git diff | 0 / 0 | Exact, source worktree clean. |

## Full baseline failure messages

```text
    generic_method_view_test.go:72: Load: /tmp/adamic-gate/TestAGenericMethodSeenThroughAnotherViewIsNotYeta_host_method_a1900388022/001/main.ts:4:71: error TS2591: Cannot find name 'process'. Do you need to install type definitions for node? Try `npm i --save-dev @types/node` and then add 'node' to the types field in your tsconfig.
    generic_method_view_test.go:72: Load: /tmp/adamic-gate/TestAGenericMethodSeenThroughAnotherViewIsNotYeta_host_method_a2731467029/001/main.a:4:71: error TS2591: Cannot find name 'process'. Do you need to install type definitions for node? Try `npm i --save-dev @types/node` and then add 'node' to the types field in your tsconfig.
    generic_method_view_test.go:72: Load: /tmp/adamic-gate/TestAGenericMethodSeenThroughAnotherViewIsNotYeta_host_method_a2003939606/001/main.ts:4:59: error TS2591: Cannot find name 'process'. Do you need to install type definitions for node? Try `npm i --save-dev @types/node` and then add 'node' to the types field in your tsconfig.
    generic_method_view_test.go:72: Load: /tmp/adamic-gate/TestAGenericMethodSeenThroughAnotherViewIsNotYeta_host_method_a46164934/001/main.a:4:59: error TS2591: Cannot find name 'process'. Do you need to install type definitions for node? Try `npm i --save-dev @types/node` and then add 'node' to the types field in your tsconfig.
```

## Full mutant failure messages

```text
    generic_method_view_test.go:72: Load: /tmp/adamic-gate/TestAGenericMethodSeenThroughAnotherViewIsNotYeta_host_method_a3660661643/001/main.a:4:71: error TS2591: Cannot find name 'process'. Do you need to install type definitions for node? Try `npm i --save-dev @types/node` and then add 'node' to the types field in your tsconfig.
    generic_method_view_test.go:72: Load: /tmp/adamic-gate/TestAGenericMethodSeenThroughAnotherViewIsNotYeta_host_method_a388344340/001/main.ts:4:71: error TS2591: Cannot find name 'process'. Do you need to install type definitions for node? Try `npm i --save-dev @types/node` and then add 'node' to the types field in your tsconfig.
    generic_method_view_test.go:72: Load: /tmp/adamic-gate/TestAGenericMethodSeenThroughAnotherViewIsNotYeta_host_method_a3438035181/001/main.a:4:59: error TS2591: Cannot find name 'process'. Do you need to install type definitions for node? Try `npm i --save-dev @types/node` and then add 'node' to the types field in your tsconfig.
    generic_method_view_test.go:72: Load: /tmp/adamic-gate/TestAGenericMethodSeenThroughAnotherViewIsNotYeta_host_method_a1205328494/001/main.ts:4:59: error TS2591: Cannot find name 'process'. Do you need to install type definitions for node? Try `npm i --save-dev @types/node` and then add 'node' to the types field in your tsconfig.
```

The failing rows are `a host method a view makes generic` (chdir) and `a host method a view makes generic, exit`, for both extensions. They fail before the lowerer judges the new guard. Red rows alone do not satisfy the requested mutant proof when baseline has the same loading failure. The mutant patch and raw logs/exits are adjacent; no test or compiler repair was attempted.

The results were reported immediately after C, quoting full chdir messages. D/E stayed at f95deb36, per the original instruction and the announced scope; [full D/E report](../f95deb36/README.md). D exits 1 on the two original chdir rows; E exits 0 with complete-corpus pass and zero new refusals. The compiler implementation is identical between these two commits; the replacement adds only three test lines.
