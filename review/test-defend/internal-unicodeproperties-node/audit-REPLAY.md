Starting commit: 60397548dd8a9494a7d2aaa874b7648b8e625607.

Apply one diffs/MNN.diff at a time and run go vet ./internal/unicodeproperties/. They are standalone production mutations without a switch. S01 is the explicitly permitted setup construction edit; W01 is the explicitly permitted witness reporting weakening. Both also pass go vet. P01 is an empty-answer probe, selected by ADAMIC_MUTANT=P01, not a production mutant.

run.py preserves exact go argv and wall times in runs.jsonl. Its frozen plan is plan.json; its switched source is saved as fixtures, not active Go files. matrix.json contains complete selected-row outcomes for all thirteen production columns. matrix-rows.json lists the ten caller rows. The full baseline and narrowed clean baselines are separate logs. Five-row medians each use three count=1 commands, and the subsumer has three additional timings.

survivor.py replays the M02 metadata witness with no switch. survivor-clean.log and survivor-M02.log show the before/after values. Restore the diff after replay. No native product cache key is needed: these are Go table/lookup mutants, not compiler or stage1 port mutations.
