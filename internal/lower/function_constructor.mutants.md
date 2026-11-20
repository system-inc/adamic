# Mutants for a function constructor writing a field on one path (#837zjkt slice 2)

Each is one edit to function_constructor.go, applied alone on a house box, the pins run with -v, then put back.

## 1. The window's layout kept present

In windowLayout, old:

```go
	if settled && slices.Equal(names, window.Order) {
```

New:

```go
	if true || settled && slices.Equal(names, window.Order) {
```

Red (J-hqz7wgx, each row's own `--- FAIL` line):

- TestFunctionConstructorOnePathSkippedStopsAtTheRead: the JavaScript backend exits 0 printing `7 -1 0`, where the checked read of end must stop with exit 70.
- TestFunctionConstructorOnePathSkippedKeysAgree: the JavaScript backend prints `kind,pos,end` and `true`, where Node prints `kind,pos` and `false`.
- TestFunctionConstructorBranchesAgree/different_fields_on_each_side: the JavaScript backend prints `checker,flags,id checker,flags,id` and `true true 0`, where Node prints `checker,flags flags,id` and `false false 0`.
- TestFunctionConstructorRegionAbsentField: the JavaScript backend prints `kind,pos,end`, where Node prints `kind,pos`.
- TestWindowLayout/one_path and TestWindowLayout/every_path,_in_another_order: each field laid present.
- TestFunctionConstructorReorderedWindowConstructs: the swapped window's fields laid present, and Object.keys reads `flags,checker,id`.

TestFunctionConstructorOnePathWrittenAgrees stays green: with verbose set every field is written, so a present layout reads what Node reads.
