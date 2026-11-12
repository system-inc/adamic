"""Scratch-only first-error unit selection; no production source changes."""
import json,pathlib,sys
root=pathlib.Path(sys.argv[1]);meta=json.loads((root/'overlay.json').read_text())
paths=[pathlib.Path(p) for p in meta['Replace'].values() if 'func latentRun(' in pathlib.Path(p).read_text()]
assert len(paths)==1
p=paths[0];s=p.read_text()
a='\tvar facts *latentProgramFacts';assert s.count(a)==1
s=s.replace(a,''' // Census-only selection is after eligibility and generic deferral. All
 // declarations and the full project stay registered for each selected unit.
 shard, shards, ordinal := 0, 1, 0
 if raw := os.Getenv("CENSUS_SHARDS"); raw != "" {
  if _, err := fmt.Sscan(raw, &shards); err != nil || shards < 1 { panic("invalid CENSUS_SHARDS") }
  if _, err := fmt.Sscan(os.Getenv("CENSUS_SHARD"), &shard); err != nil || shard < 0 || shard >= shards { panic("invalid CENSUS_SHARD") }
 }
'''+a)
a='\t\t\t\tdefer latentUnitTimer.Start(record.File, unit.Where)()';assert s.count(a)==1
# Do not execute or append unselected attempted units. Deferred and skipped units
# remain in every shard, so every shard's witness metadata can be merged.
b='\t\t\tfunc() {\n'+a;assert s.count(b)==1
s=s.replace(b,'''   selected := ordinal % shards == shard
   ordinal++
   if !selected { continue }
'''+b)
p.write_text(s)
(root/'sharding.go.txt').write_text(s)
