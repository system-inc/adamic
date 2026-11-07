# Nineteen enum-view sites on main after Lower's binding fix

Frozen main: 01813e07a5b062baa5cabd9b5fb7e4ad6c983382.
The exact frozen SHA is recorded in 01-setup.tar.gz/provenance.json and will be
used in the completed result below. This is a fresh worktree and a report-only
branch, typescript/enum-views-19-main. Adaptation 40 and production code are
unchanged by this report.

The comparison carries the previous three columns from report 9683a4e4:
main 63e14a0b, adaptation 40 alone, and B2 at 9bdbe75e. As before, the nineteen
sites mean eighteen debug.ts casts and rule 91 in emitter.ts. The new column
will record actual owning-unit outcomes, separately from scan admission.

The frozen main includes 80168a7db, the entry-reach and const-enum-binding fix,
and intervening changes including adaptations 99r and 99z. All ordinary first
stop changes are reported for the complete frozen snapshot; no causal
attribution solely to the enum fix is assumed.

## Completed pieces

01-setup.tar.gz contains the target and baseline pins, tool hashes, reference
records, and build logs. A stale temporary build directory from the prior setup
was recreated before the successful build; the initial failure is retained.
The target's full apply is running. The latent nineteen-site measurements and
ordinary 79-root comparison will follow in bounded pieces, each committed and
plainly pushed as it finishes.

All raw records are packed in numbered tar.gz archives. SHA256SUMS hashes the
archives, and each archive includes a SHA256SUMS for its raw member files.
