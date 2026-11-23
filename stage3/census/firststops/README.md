# The first-stop census on a cadence

Every census unit's first stop at one commit of main, read as a table, and the table read against the last run's.
It's the map minds steer by on #bbssdcy (#f2kma1s; measured on #yfemexs).

- `run.sh` is the run: the adapted tree, the census tools, both passes side by side, and the table as `firststops.py`'s
  `FS` lines. Ask it sealed at main's sha, at F16 (both passes peak at about 6.7 GiB; a run is about 25 minutes):
  `ahra loom job 'source /opt/adamic-block/environment.sh && bash stage3/census/firststops/run.sh' --kind proof --size F16 --ceiling 60m --at adamic@<sha> ...`
- `fsdiff.py <previous log> <current log> [pass]` reads two runs' logs into one diff. A unit is its file, name, kind and
  ordinal, and its family is `families.py`'s. Every unit present in either run is in exactly one kind: lowers now, stops
  newly, moved family, moved site in one family, reworded at the same site, the same stop at a shifted line, or
  unchanged.
- `families.py` is the list of refusal families, the first rule whose regex matches the whole reason. A reason no rule
  takes is its own family, so when a reason is reworded, the diff lists it as reworded until a rule learns the words.

A kingdom watch asks it when a landing touches `internal/lower`, `internal/ir`, `internal/load`, `stage3/adapt` or
`stage3/census/latent`, one at a time, and posts the diff's head on #bbssdcy with the table attached.
