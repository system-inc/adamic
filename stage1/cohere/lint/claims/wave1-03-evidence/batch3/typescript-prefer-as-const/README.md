# @typescript-eslint/prefer-as-const: reserved, not ported

Finding range 9..14 points at the literal type. Its fix deletes 7..14 and inserts as const at 22..22. The finding model carries only one replacement at the finding range. The existing Go serializer exits 2 with unexpected fix shape.

The independent full diagnostic observation is in full.txt. legacy.txt preserves
the unchanged shared oracle's refusal. upstream.txt records 7 unchanged
upstream test functions passing (0.013s). The oracle.go adapter selects the
actual pinned Go rule without changing its implementation. testdata contains its
minimal raw TypeScript witness, not an Adamic module.

There is no rule.json, implementation or claimed semantic mutant here. This
folder lives under the claim evidence directory because every subdirectory of
rules must be a complete registered rule; incomplete evidence must not become a
fake registered port or introduce another discovery failure.

Reproduce all three with the sibling no-unnecessary-type-constraint probe.py.
See ../REPORT.md for the command, selection ledger and shared ownership boundary.
