# Caller witness validation and mutant

Base: e07f4f064584ad499d76aa5831ad96162b5281ce (unchanged compiler and
adapted source bytes from bdbc73864f82dbc0c51434decf0ffe96d4782767).

The expanded audit passes in legacy, full and speculative modes. It independently
checks the caller unit, call-site columns, checker-printed number argument,
outer/nested inheritance, unused units, deterministic ordering and truncation of
eight distinct calls to five witnesses. Existing gap and byte-mask mutants pass
their rejection checks too.

`LATENT_MUTANT_DROP_INSTANTIATED_BY=1 python3 stage3/census/latent/audit_deferred.py
BINARY OUTPUT` exits 1 at the new `called witnesses` assertion. The boolean
still reads true; the missing witness field is what the assertion rejects.
See witness-mutant.log.txt for the actual assertion and child exit evidence.

The whole statecopy package passes, including TestLatentSpeculativePrivateIRFields.
No lowering field was introduced. The run-owned observation globals are described
in statecopy/POLICY.md.

Fixture streams in all three modes are byte-identical after removing only the
new JSON field, without reserializing any old field. Meter lowering summaries
are identical. fixture-identity.json retains the SHA-256 pairs. The pair reader's
old stable-field check rejects the new metadata; a one-line scratch repair that
excludes instantiated_by produces identical counts and no paired deltas. Its
placement outside latent is being kept separate from this tooling-only change.

The final 79-file and 81-file entry streams both pass byte identity after
removing only instantiated_by. corpus-identity.json retains the raw, stripped
and new SHA-256 hashes. The entry driver alphabetizes map keys; the audit restores
the witness struct field order when checking the recorder's lexical order.
Seven actual meter report artifacts are byte-identical with timing and readiness
inputs fixed; meter-identity.json records those hashes. Complete streams and
paired-reader evidence are retained in the stacked deferred-reach report.
