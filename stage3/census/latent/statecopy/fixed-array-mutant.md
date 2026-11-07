# Fixed-array value-copy mutant

Mutation: replace the element copier identity comparison in fixedArrayCopiesByValue with `return true`, accepting every fixed array by value. The production tree is unchanged; a Go build overlay applies the mutant.

Command: `go test -overlay=/tmp/container-fit-statecopy-mutant/overlay.json ./stage3/census/latent/statecopy -run '^TestLatentStatecopyRefusesFixedArrayNeedingCopy$' -count=1 -v`. Exit: 1, as required. The named struct-array case turns red, alongside the other owned-element cases.

```text
=== RUN   TestLatentStatecopyRefusesFixedArrayNeedingCopy
=== PAUSE TestLatentStatecopyRefusesFixedArrayNeedingCopy
=== CONT  TestLatentStatecopyRefusesFixedArrayNeedingCopy
=== RUN   TestLatentStatecopyRefusesFixedArrayNeedingCopy/entryfalse
    guard_test.go:330: got <nil>, want refusal state copy: fixed array [2]entry requires an explicit copier
=== RUN   TestLatentStatecopyRefusesFixedArrayNeedingCopy/entrytrue
    guard_test.go:330: got <nil>, want refusal state copy: mutable named container entries requires an explicit copier
=== RUN   TestLatentStatecopyRefusesFixedArrayNeedingCopy/map[int]intfalse
    guard_test.go:330: got <nil>, want refusal state copy: fixed array [2]map[int]int requires an explicit copier
=== RUN   TestLatentStatecopyRefusesFixedArrayNeedingCopy/map[int]inttrue
    guard_test.go:330: got <nil>, want refusal state copy: mutable named container entries requires an explicit copier
=== RUN   TestLatentStatecopyRefusesFixedArrayNeedingCopy/[]intfalse
    guard_test.go:330: got <nil>, want refusal state copy: fixed array [2][]int requires an explicit copier
=== RUN   TestLatentStatecopyRefusesFixedArrayNeedingCopy/[]inttrue
    guard_test.go:330: got <nil>, want refusal state copy: mutable named container entries requires an explicit copier
=== RUN   TestLatentStatecopyRefusesFixedArrayNeedingCopy/ir.Valuefalse
    guard_test.go:330: got <nil>, want refusal state copy: fixed array [2]ir.Value requires an explicit copier
=== RUN   TestLatentStatecopyRefusesFixedArrayNeedingCopy/ir.Valuetrue
    guard_test.go:330: got <nil>, want refusal state copy: mutable named container entries requires an explicit copier
--- FAIL: TestLatentStatecopyRefusesFixedArrayNeedingCopy (0.00s)
    --- FAIL: TestLatentStatecopyRefusesFixedArrayNeedingCopy/entryfalse (0.00s)
    --- FAIL: TestLatentStatecopyRefusesFixedArrayNeedingCopy/entrytrue (0.00s)
    --- FAIL: TestLatentStatecopyRefusesFixedArrayNeedingCopy/map[int]intfalse (0.00s)
    --- FAIL: TestLatentStatecopyRefusesFixedArrayNeedingCopy/map[int]inttrue (0.00s)
    --- FAIL: TestLatentStatecopyRefusesFixedArrayNeedingCopy/[]intfalse (0.00s)
    --- FAIL: TestLatentStatecopyRefusesFixedArrayNeedingCopy/[]inttrue (0.00s)
    --- FAIL: TestLatentStatecopyRefusesFixedArrayNeedingCopy/ir.Valuefalse (0.00s)
    --- FAIL: TestLatentStatecopyRefusesFixedArrayNeedingCopy/ir.Valuetrue (0.00s)
FAIL
FAIL	github.com/system-inc/adamic/stage3/census/latent/statecopy	0.003s
FAIL
```

Restored full subtree: `GOFLAGS=-buildvcs=false go test ./stage3/census/latent/... -count=1`, exit 0. VCS stamping was disabled because the initial isolated checkout shared its dependency by symlink. The dependency was then made a real worktree at the identical pins.

```text
ok  	github.com/system-inc/adamic/stage3/census/latent/importclean	0.990s
ok  	github.com/system-inc/adamic/stage3/census/latent/overlaytext	0.016s
ok  	github.com/system-inc/adamic/stage3/census/latent/refusalrewrite	1.122s
?   	github.com/system-inc/adamic/stage3/census/latent/refusalrewrite/cmd	[no test files]
?   	github.com/system-inc/adamic/stage3/census/latent/replay	[no test files]
?   	github.com/system-inc/adamic/stage3/census/latent/replay/worker	[no test files]
ok  	github.com/system-inc/adamic/stage3/census/latent/statecopy	4.104s
ok  	github.com/system-inc/adamic/stage3/census/latent/statementrewrite	3.497s
ok  	github.com/system-inc/adamic/stage3/census/latent/timing	0.029s
?   	github.com/system-inc/adamic/stage3/census/latent/tool	[no test files]
```

Meter overlay: `python3 stage3/census/latent/make_overlay.py "$PWD" /workspace/container-fit-statecopy-overlay`, gofmt, then `go build -buildvcs=false -overlay=/workspace/container-fit-statecopy-overlay/overlay.json -o /workspace/container-fit-statecopy-latent ./stage3/census/latent/tool`: exit 0.
