
## Tooling adaptations for this run

- `run.py`, `run-one.py`, `analyze.py`, `source-excerpts.py`, `measure.py`, and `validate.py`: use the fresh worktree and unique adapted-tree, binary, output and report paths so the stop record and prior run remain intact.
- `run.py`: remove the unused all-groups scheduler; invoke one bounded `run-one.py` process per root to keep scheduler processes under 30 minutes as well as census children. Child exit codes and process-group cleanup remain recorded.
- `validate.py`: remove the old failed-build artifact requirement and historical result field. Record every status file's actual integer child exit code; require all root children to exit zero, 79 exact sorted roots, two observations per root agreeing in every non-timing field, verified comparison tuples, binary/source hashes, process cleanup and duration bounds. Require phase and owner evidence for every stopped group, including unchanged groups.
- `write-report.py`: update target and baseline identifiers, the prior inspector location, fresh-run build narrative and measurement paths; add the sys.ts/root-order artifact discussion, this change list and actual checkpoint history.
- `package-evidence.py`: accumulate compression metadata across checkpoint batches and retain SHA256 rather than commit a compressed artifact over 20 MB. Raw logs are packaged only after their child exits.
- `assemble.py` is unchanged. `analyze.py` retains the baseline's exact tuple comparison definitions: unchanged, moved to an existing baseline tuple, or new tuple absent from baseline. `source-excerpts.py` retains its source extraction algorithm. The ordinary Go census source and compiler are unchanged, with no overlays.
- Fresh `provenance.json`, `owners.json`, `stop-phases.json`, binary metadata and source evidence describe this run; previous status files and emitter excerpts were not imported as current evidence.
