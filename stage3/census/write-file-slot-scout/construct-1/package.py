from pathlib import Path
import shutil,json,gzip,hashlib
r=Path('/workspace/scratch/write-file-construct-run');repo=Path('/workspace/scratch/write-file-slot-scout');dest=repo/'stage3/census/write-file-slot-scout/construct-1';assert not dest.exists();dest.mkdir(parents=True)
assert (r/'collect.exit').read_text().strip()=='0';accept=json.loads((r/'ACCEPTANCE.json').read_text())
# Preserve complete final receipts. Implementation-time attempts remain separate and labelled.
selected=['ACCEPTANCE.json','OWNERSHIP.md','ownership-grep.txt','production-roots.md','code-pin-ledger-identity.json','ledger-check.json','view-pin-continuity.json','final-source-identity.json','construct-mutant-source-identity.json','checks.json','checks-exits.json','actual-stage-proofs.json','actual-small-proofs.json','built-proofs.json','actual-builds.json','mutants.json','view-cli-mutant.json','runtime-keys.json','pairs.json','declarations.json','diffs.json','stock.tsconfig.json','roots.json','roots-control.json','roots-candidate.json','root-profile.json','root-profile.exit','root-driver-receipt.json','root-driver.go.txt','root-overlay.json','frozen-proof.bundle','candidate-entry.stdout','candidate-entry.stderr','candidate-entry.exit','final-apply.log','final-apply.exit','main-apply.log','main-apply.exit','control-apply.log','control-apply.exit','final-npm.log','adamic-types.log','stock.log','public-declarations.log','real-body-witness.log','view-moved-witness.log','witness.json','runtime-keys.log','runtime-keys.exit','99-watch-directory-witness-final.log','99-watch-directory-witness-final.exit','99-scope-mutant-current-shape.log','99-scope-mutant-current-shape.exit','93-main-prefix-known-red.log','93-main-prefix-known-red.exit','93-main-scope-mutant.log','93-main-scope-mutant.exit','construct-contract-final.log','pin93-final.log','pin99-final.log','pin-view.log','pin-construct-final.log','collect.log','collect.exit']
selected += [p.name for p in r.glob('full-sys-*') if p.is_file()]
selected += [p.name for p in r.glob('root-*.jsonl')]+[p.name for p in r.glob('root-*.manifest.json')]+[p.name for p in r.glob('root-*.log')]
selected += [p.name for pattern in ['actual*-*.log','99-*-mutant.log','99-*-refused.log','construct-*-mutant*.log','view-cli-*.log'] for p in r.glob(pattern)]
recipes=['actual-stage-proofs.py','actual-small-proofs.py','actual-built-proofs.py','checks.py','checks.cjs','declarations.cjs','witness.cjs','mutants.py','measure.py','full-sys.py','pair.py','diffs.py','runtime.py','runtime-keys.cjs','build.py','collect.py','package.py']
selected += recipes
manifest=[]
def copy(source,target):
 target.parent.mkdir(parents=True,exist_ok=True);data=source.read_bytes();original=str(source.relative_to(r));sha=hashlib.sha256(data).hexdigest()
 if source.suffix=='.log' and len(data)>1_000_000:
  target=target.with_name(target.name+'.gz');target.write_bytes(gzip.compress(data,mtime=0))
 else:target.write_bytes(data)
 manifest.append({'input':original,'artifact':str(target.relative_to(dest)),'bytes':len(data),'sha256':sha,'artifactSha256':hashlib.sha256(target.read_bytes()).hexdigest()})
for name in sorted(set(selected)):
 p=r/name
 if p.is_file():copy(p,dest/name)
for name in ['ownership-parent','emits','diffs','93-actual-owner-contract']:
 for p in sorted((r/name).rglob('*')):
  if p.is_file() and '/tree/' not in str(p):copy(p,dest/p.relative_to(r))
for name in ['control','candidate']:copy(r/'declarations'/name/'typescript.d.ts',dest/'declarations'/name/'typescript.d.ts')
for p in sorted((r/'lane').rglob('*')):
 if p.is_file() and 'adapted-tree' not in p.relative_to(r/'lane').parts:copy(p,dest/p.relative_to(r))
(dest/'MANIFEST.json').write_text(json.dumps(manifest,indent=2)+'\n')
checks=json.loads((r/'checks-exits.json').read_text());checks=[x for x in checks if x['name']!='99-watch-directory-witness'];checks.append({'name':'99-watch-directory-witness','command':['node','stage3/adapt/99-wasm-crypto-import/proofs/watch-directory-witness.cjs','LANE/adapted-tree'],'exit':0,'receipt':'99-watch-directory-witness-final.log'});(dest/'final-checks-exits.json').write_text(json.dumps(checks,indent=2)+'\n')
(dest/'README.md').write_text('''# Construction adaptation acceptance

Final acceptance index: [ACCEPTANCE.json](ACCEPTANCE.json).
All 79 native outcomes: [production-roots.md](production-roots.md).
Owner routing: [OWNERSHIP.md](OWNERSHIP.md).
Complete census pairing: [pairs.json](pairs.json).
Final check exits: [final-checks-exits.json](final-checks-exits.json),
[actual-stage-proofs.json](actual-stage-proofs.json),
[actual-small-proofs.json](actual-small-proofs.json), [built-proofs.json](built-proofs.json).
The two 93 legacy failures are measured on main itself and documented rather than
claimed as successful mutation tests. Earlier incorrectly parameterized witness
attempts in the raw checks index are superseded by the final index above.

The frozen implementation is 51104a74ba9e981fe93404e21e9c66298e47425f; its merge
with main is 0adcd82e25a88f32c50f625c21a0f127853537c1. The 21 KB frozen-proof.bundle
preserves both commits, requiring only the parent and main named in ACCEPTANCE.json.
Use `git bundle verify frozen-proof.bundle`, then `git fetch frozen-proof.bundle HEAD`
in a clone that contains those parents. The published feature completes proof
text/citations by a documentation-only amendment. Code, pin and ledger hashes
are listed in code-pin-ledger-identity.json; all 747 final fresh source hashes
match the once-tested lane in final-source-identity.json.

Census is full, single-file sys.ts, LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1,
CommonJSBundle, with complete relation trailers and zero diagnostics. Unique
Refused/NotYet findings are paired by kind, phase, reason and position-stripped
message. The raw Boundary disappearance is recorded separately.

Public control/candidate bundles are byte-identical; the stock upstream API diff
is inherited unchanged. Internal declarations lose the same @internal helper.
All actual source, single-file JavaScript and built-bundle diffs are under diffs/.
The real-body witness has 29 equal Node events and 10 equal harness events. Runtime
keys lose exactly the helper, writer name changes, and length/flags remain stable.

Whole lane was run once without filters or expectation changes: PASS, 106,366
passing, 0 pending, only the sanctioned API row. Its raw oracle exit is 1 and lane
exit is 0, as recorded. All logs and verdicts are preserved under lane/.
MANIFEST.json records full uncompressed hashes; large logs may use .gz.
Recipes retain the measured absolute scratch paths for unambiguous provenance.
No adaptation is submitted.
''')
report=repo/'stage3/census/write-file-slot-scout/REPORT.md';old=report.read_text();section='''## Adaptation acceptance on the requested ledger chain

| Road | Target gone | New / changed sys.ts findings | JavaScript change | Public declaration change |
|---|---|---|---|---|
| (a′), adapted in typescript/write-file-construct-1 | Yes, all 73 advance to NotYet sys.ts:1481:57 | 0 / 0; one old finding gone | Yes: named construction wrapper, direct call, no raw slot exposure, internal export removed; harness helper moved | None against parent+main; System remains a method |

Base: a77ca4a82271f18c142d8843ea76604b30d3833d. Re-fetched main:
89614cc0f9d6aa7b7fe96b56c5c6c1c3eace9825. The full 79-root table has 73 moves and
six unchanged outcomes. Full sys.ts: 51 → 50 unique findings, 1 gone / 0 new / 0
changed; old lowering refusal at 1392:49 is gone. Stock tsc and adamic types exit 0
on all 79 roots. StripInternal + upstream public bundler is byte-identical to the
parent+main control (f03adcf915df559d616fcafd75181d3423eaa0f272237482cedbd025525c9516);
stock API differences are inherited unchanged. The first getNodeSystem scan stop
is 1481:57; the removed helper has no remaining sys.ts unit.

The once-run whole lane is PASS: 106,366 passing, 0 pending, only the sanctioned
API row. A separate fresh application of the frozen implementation merged with
main is green; all 747 output source files match that once-tested tree exactly.
Its restoring mutant exactly restores all 747 parent+main source files, preserving
1965, 1533 and the receiver chain. Every owning directory's relevant contracts,
bindings, built witnesses and mutants are recorded with exits; 93's legacy red
checks are also measured on main itself.

The four touched directories and why are in [the ownership audit](construct-1/OWNERSHIP.md).
Pins were regenerated only by their own tools. TestServerHost and createServerHost
stay byte-identical, so their timer-closures owner is not amended. The ledger keeps
the parent's exact bytes with zero parent lines removed.

The real-body witness repeats 29 equal Node events and 10 equal harness events.
Writer name changes from empty to patchedWriteFile, length stays 3, raw writer
never occupies the slot, direct receiver is undefined with no this reads, and own
.call control changes 5 → 0. Runtime ts keys change 2256 → 2255, losing only the
internal helper; remaining order and descriptor flags agree. The accepted reader
conditions remain cited to de29fb07566f3eb82a1895fe6561ce13e61f5139 below: zero
name/length readers among 479 accesses, zero adapted ts helper reads, 27 controls.

[Acceptance index](construct-1/ACCEPTANCE.json), [all 79 roots](construct-1/production-roots.md),
[full census pair](construct-1/pairs.json), [fresh-source identity](construct-1/final-source-identity.json),
[proof inputs and receipts](construct-1/README.md). Nothing submitted.

---

''';report.write_text(section+old)
print(json.dumps({'files':len(manifest),'uncompressedBytes':sum(x['bytes'] for x in manifest),'destination':str(dest)},indent=2))
