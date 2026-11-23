Task dp24r5m: 99-correctness-tracing-dumptypes replaces two explicit-any reads
with getObjectFlags and an Intrinsic flag guard plus checked IntrinsicType read.
The whole lane records 0 calls across 398 shards and 797 Node processes; three
generateTrace projects visit 84,642 types, including 46 real objectFlags values
outside ObjectFlagsType, with 0 differences at all five masks or intrinsicName.
Written types are byte-identical, 79 roots have 0 diagnostics, both restoring
mutants restore their sites, every JavaScript hunk stays inside dumpTypes, and
2,256 runtime export keys are identical. See its PROOF.md. No upstream issue filed.

- #baobab, Nov 11 2026 23:29Z, Adamic: retire 62 and 63's widened fs views
  using `fs.writeSync(typesFd, JSON.stringify(descriptor)!)` and
  `fs.writeFileSync(legendPath, JSON.stringify(legend)!)`. This applies the
  established postfix non-null check ruling (docs/rulings.md:50): evaluate once,
  reject only null and undefined, exit 70, counted. Whole-owner pins and literal
  guards require the const descriptor object and namespace-level const legend
  array, with no declared toJSON and no binding reassignment. Amend 99's
  dumpTypes pins to the new spelling; its subtype read changes are preserved.
  Evidence is on typescript/tracing-fs-views-evidence, outside stage3/adapt.
