# Merged member slot mutant

On Linux, a Go overlay removed the call to `reserveMergedFields` in
`internal/lower/class_inheritance.go`, leaving `f` off the instance layout.
The overlay affected only that test run.

`go test -overlay <overlay.json> ./internal/oracle -run '^TestMergedMemberSlot$/^write$' -count=1 -v`

Pin 1 failed. Native exited 70 with empty stdout and this stderr:

```text
adamic: panic: compiler bug: a field the checker proved is there is missing
```

Node and Adamic's JavaScript backend printed `m/string a,f` and exited 0.

Repeated on the branch based on `423a6a9b`. The overlay leaves the shared
class-slot decision in place and removes the constructor layout call, so the
pin reaches the original native failure.
