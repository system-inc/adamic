# Never constructs retain Refused diagnostics (#53r37a3)

`TestNeverRefusalMutantsAreCaught` uses `buildcache.GoTestMutated` to replace one
Refused return in `never_refused.go` with NotYet, leaving its recognition guard
unchanged. Each child runs only its corresponding rows of
`TestNeverConstructsAreRefused`, in both `.a` and `.ts`.

| Mutant | Child rows | Catch text |
| --- | --- | --- |
| any returns NotYet | any | Never construct must be Refused with its fix |
| global eval returns NotYet | eval | Never construct must be Refused with its fix |
| global Function returns NotYet | new_function, new_function_statement, function_call | Never construct must be Refused with its fix |
| non-tuple array length write returns NotYet | array_length | Never construct must be Refused with its fix |

The catch is the `t.Fatalf` in `TestNeverConstructsAreRefused`. Controls cover a
user eval method, a user Function class, object and class length fields, the
admitted tuple-length no-op, and unknown narrowed before use. Checker pins keep
TS1215 for a local eval binding and TS2540 for readonly/typed-array length writes.

Historical meter/census snapshots are not rewritten. Their old reason remains in
owners.json beside the four new reasons. The exact oracle stdout_handle_view
stops are live readers and now expect the Refused text.
