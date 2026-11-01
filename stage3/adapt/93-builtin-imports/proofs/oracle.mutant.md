Mutant evidence for #he8tysb head 1's oracle baseline guard, following 92.

Mutant: append an extra line to the actual completed baseline.diff without
changing either API snapshot. The command that executed the probe was:

    BUILTIN_PROOF_RESUME_AFTER=1 python3 stage3/adapt/93-builtin-imports/proofs/oracles.py

The full serial recipe also runs this probe without the resume switch.
It calls the unmodified stage3/lane/check.py api_check, exactly as 92 did.
The observed red is:

    baseline.diff does not exactly describe the API snapshots (extra, missing, or changed diff lines)

baseline-mutant.json records that rejection. The driver restores the exact
original bytes in finally. Restored command:

    python3 stage3/lane/check.py /workspace/scratch/adapt-93-proof/lane-after

Exit 0, exact verdict:

    PASS stage3 landing lane [causes: timeout=0, baseline-content=1, baseline-missing=0, exception=0, other=0]

The extra line is not sanctioned. The single baseline-content observation in
the passing verdict is main's existing public API failure. No expected count,
API sanction, shared runner or checker source is changed. Complete raw output
is retained in logs/lane-after-restored.log.gz, together with both full suites.
No baseline or source mutant remains.
