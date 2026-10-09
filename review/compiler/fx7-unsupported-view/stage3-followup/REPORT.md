Kept the all-or-nothing compiler refusal and updated only the stage3 fixture status toward fix-forward 5, item 142.
Merged main c4c59914 in 812eb7a3; delivery commit is the commit containing this report.
Fixture leaf passes in 0.10s; full lower passes in 42.106s; TestCallTargetReaders passes in 6.440s.
Revert mutant fails the stage0 status assertion in 0.11s, while its Node assertion passes.
No native comparison is possible for this fixture: the reverted compiler also stops before native emission.

The refusal is right. The fixture casts Node to ObjectLiteralExpression at
stage3/fixtures/assertions/09_interface_kind.a:15:10. Its properties member is
readonly number[]. internal/lower/view_unions_dispatch.go:74 marks array
contracts Unsupported with views-v3: array element kind, and
internal/lower/view_contracts.go:15 has no registered array hook. A supported
ordinary array representation does not provide a checked view lowering.
The compiler check was not narrowed, so p37 and p21 remain refused. The full
lower run includes those tests and the supported-only Node agreement control.

Observation: the old recorded status was NotYet at the member read, not
Compiles. After reverting the original compiler fix using the saved
../revert.diff, building the old compiler succeeds. Building the exact fixture
with that compiler exits 1 at 15:9 with the same NotYet diagnostic. There is
no native executable or native stdout to compare. Node exits 0 with exactly
true\nfalse\nfalse\n and empty stderr. Thus this is a status migration from
read-site NotYet to creation-site Refused, not evidence of a native output
mismatch in this particular fixture. The original p37 mismatch evidence is
preserved in the parent report.

The source fixture, recorded Node behavior and provenance are unchanged. Its
status now pins the exact creation-site diagnostic, member name, member type,
and fix text. No new fixture was added, so counts.md needs no new row. No Go
test file was changed. Full packages other than the requested internal/lower
were not run; unrelated Loom lint units were not changed.

Reproduced the integration red after merging current main. An initial test
started while merge was running failed package compilation; the post-merge
reproduction and all subsequent observations used the completed merge.
See COMMANDS.txt and the adjacent logs for commands and raw evidence.

Lane checks pass: 1.4 s, gofmt and tools on 6 Go files, t.Parallel on
2 test packages, vet 2 packages. The first lane attempt lacked gofmt on PATH;
sourcing the toolchain environment fixed it.
