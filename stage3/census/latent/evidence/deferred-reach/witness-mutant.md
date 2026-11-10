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

At this checkpoint, full 79-file and 81-file entry comparisons are still running.
Their final evidence belongs in the stacked deferred-reach report. This document
does not claim those long-run comparisons have already completed.
