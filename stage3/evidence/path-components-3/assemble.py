from pathlib import Path
import gzip,hashlib,json,shutil,subprocess
repo=Path('/workspace/adamic');out=Path(__file__).parent;dest=Path('/workspace/scratch/path-components-main/stage3/evidence/path-components-3')
required=['tip-equivalence','build-adamic-tip','build-latent-tip','reset-before','reset-after','apply-before','apply-after','build-adamic','overlay','build-latent','clean-before','clean-after','build-before','build-after','types-before','types-after','stock-before','stock-after','source-scope','uses-before','uses-after','census-before','census-after','census-delta','compare','prepare-replay','replay40','replay-verify','check-replay','check-restored','check-full-after','planner','projection-four-states']
for n in required:
 r=json.loads((out/(n+'-status.json')).read_text());assert r['exit_code']==0 and not r['stopping_reason'] and r['seconds']<1800,(n,r)
assert json.loads((out/'restoring-mutant-status.json').read_text())['exit_code']==1
before=json.loads((out/'uses-before.json').read_text());after=json.loads((out/'uses-after.json').read_text());assert before['total']==after['total']==13
for data in [before,after]:assert all(u['explicitTypeArguments'] in [None,0] for u in data['uses'])
assert [u for u in before['uses'] if u['line']!=457 or not u['file'].endswith('/path.ts')]==[u for u in after['uses'] if u['line']!=457 or not u['file'].endswith('/path.ts')]
sha=subprocess.check_output(['git','rev-parse','HEAD'],cwd=repo,text=True).strip();base=subprocess.check_output(['git','rev-parse','origin/main'],cwd=repo,text=True).strip()
assert subprocess.check_output(['git','rev-list','--count',base+'..HEAD'],cwd=repo,text=True).strip()=='1'
p=json.loads((out/'provenance.json').read_text());p['appliedMain']=json.loads((out/'tip-equivalence.json').read_text())['appliedMain'];p.update(implementationCommit=sha,main=base,binaries={n:hashlib.sha256((out/n).read_bytes()).hexdigest() for n in ['adamic','latent']},adaptation40={f.name:hashlib.sha256(f.read_bytes()).hexdigest() for f in sorted((repo/'stage3/adapt/40-explicit-any').iterdir()) if f.is_file()})
(out/'provenance.json').write_text(json.dumps(p,indent=2)+'\n')
dest.mkdir()
for f in out.iterdir():
 if f.is_file() and f.name not in ['adamic','latent']:shutil.copy2(f,dest/f.name)
shutil.copytree(out/'latent-overlay',dest/'latent-overlay');assert not list(dest.rglob('*.go'))
(dest/'implementation.patch.gz').write_bytes(gzip.compress(subprocess.check_output(['git','show','--format=fuller','HEAD'],cwd=repo),mtime=0))
census=json.loads((out/'census-delta.json').read_text());rule=next(r for r in json.loads((repo/'stage3/adapt/40-explicit-any/class-rules.json').read_text())['brands'] if r['id']==103)
report=f'''# Generic PathPathComponents interface proof

Implementation: `{sha}` on `typescript/path-components-interface-3`.
Main tip and proof base: `{base}`.
Upstream: TypeScript 6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`.
This is the generic correction to the earlier PathPathComponents amendment.
The earlier evidence remains in ../path-components; this directory is a fresh run.
No production compiler source differs from main, and no lane files change.

The full applies ran at fb1c8d3b72de2686c486cf97685079589308dafd. Main then
advanced only in bridge/tsgo/units_test.go; tip-equivalence.json proves that this
is an excluded external Go test and all compiler-production, adaptation and
emission inputs are byte-identical to the final base. Compiler builds, types and
censuses were refreshed after that rebase. Apply and emission results therefore
validate the exact same inputs as final main, without repeating identical work.

## Contract

Final rule 103 target:
`{rule['after']}`

The planner accepts exactly the upstream any intersection, undefined intersection,
previous non-generic interface and generic interface. Every other owner spelling
is refused. The final interface replays with zero edits. The default fills all
13 existing bare uses, with no consumer edits. declarationProjection parses its
replacement from rule.after, checks rule.owner, derives its upstream base check
from rule.upstream, and retains the emitted scope's modifiers and indentation.
It contains no hardcoded owner or Path array spelling. Eight projection cases
(four states in both module and namespace output) match stock declaration emit.
Planner cases also cover LF/CRLF, unchanged JavaScript and five rejected drifts.

## Proofs

- Both full applies pass from independently reset pinned checkouts, with the
  existing locked dependencies already installed. Only the requested declaration
  differs among 747 source files (source-scope.json).
- Fresh emission builds remove only built/, preserving adapted diagnostic source.
  All ten JavaScript outputs and both public declarations are byte-identical;
  comparison.json records every size and SHA-256 hash. Public d.ts files are
  compared in both lib and built/local. Internal declaration changes are allowed.
- Adamic types exits 0 on all 79 roots before and after. Stock TypeScript under
  Adamic's options and matching prelude, pinned Node and regexp library views
  reports zero diagnostics on both trees. Commands, roots and logs are retained.
- Complete adaptation 40 replay on the recorded generic post-40 tree makes zero
  edits across 747 files. This avoids conflating 40 with downstream owner changes.
- Restoring the **non-generic interface** makes path-components-check.cjs exit 1.
  The generic tree passes before mutation and after restoration; the fully
  composed generic tree also passes.
- PathPathComponents remains internal: absent from the upstream public API golden
  and sanctioned-api.json. Inspection hashes and the full competing-pin scan are
  retained in public-api.json and pin-scan.txt. No other adapter pins it whole.

## All identifier uses

13 uses before and after: one declaration, one overload result, one import,
ten Readonly<PathPathComponents> references, zero casts. Every type reference
has zero explicit type arguments. All consumer rows are unchanged; only the
owner declaration changes from alias to generic interface.

| File | Line:column | Use |
|---|---|---|
'''
for u in after['uses']:report+=f"| {u['file']} | {u['line']}:{u['column']} | {u['kind']} |\n"
report+='''
## Single-file full latent census

Both runs use LATENT_FULL=1, LATENT_ONLY_FILE=<tree>/src/compiler/path.ts and
LATENT_ASSERT_NO_OUTPUT=1, with the same guarded main-tip binary. Raw JSONL,
unit times, environment and child exit records, and .go.txt overlay sources are
retained. Counts are unique (kind, where, reason, text) sites; only tree prefixes
are normalized. This is a path.ts-only measurement.

| Kind | Before | After | Delta |
|---|---:|---:|---:|
'''
for kind in ['NotYet','Refused','Boundary','error','panic']:
 a=census['before']['counts'].get(kind,0);b=census['after']['counts'].get(kind,0);report+=f'| {kind} | {a} | {b} | {b-a:+d} |\n'
report+=f"| All unique sites | {census['before']['uniqueSites']} | {census['after']['uniqueSites']} | {census['delta']:+d} |\n\n"
report+=f"Removed sites: {len(census['removed'])}. Added sites: {len(census['added'])}.\n"
report+='All eleven replacements change only diagnostic rendering from PathPathComponents to PathPathComponents<Path>. Counts and locations are unchanged; census-display-changes.json verifies the exact replacement at each site. The existing getPathComponents overload-result refusal remains.\n'

for action in ['removed','added']:
 for s in census[action]:report+=f"\n- {action}: {s['kind']} at `{s['where']}` — {s['reason']}.\n"
if not census['removed'] and not census['added']:report+='No site changes.\n'
report+='''
Full site text and unit/deferred coverage are in census-delta.json and sites-*.json.
Both census headers report checker_rejected=false and empty diagnostics; the
legacy measurement label saying "checker-rejected program" is not authoritative.

## Process integrity

Every subprocess wrapper reports its child's own exit code. run.py enforces a
1740-second limit and records command, elapsed time, exit code and stopping reason.
types-run.py and census-run.py exec the actual binary. All required children exit
0 except the expected restoring mutant, which exits 1. No measured process runs
for 30 minutes. The first equivalence inspector assumed Go's TestGoFiles field;
this test is external and appears in XTestGoFiles. The corrected inspector checks
both categories and passes; the initial failed inspector record is retained. All Go evidence filenames end .go.txt. No evidence is in adapt/.

Scripts retain the exact /workspace paths used for reproducibility. Source
/workspace/adamic-tools/env.sh for Go/Node and set NODE_PATH to
/workspace/adamic/stage3/api/node_modules for adaptation helpers. comparison.json,
census-delta.json and use inventories are reproduced by compare.py,
census-delta.py and uses.cjs. SHA256.json hashes every retained artifact except
itself. Binaries and the replay tree are omitted; their hashes and source
provenance are retained.
'''
(dest/'REPORT.md').write_text(report)
manifest={f.relative_to(dest).as_posix():hashlib.sha256(f.read_bytes()).hexdigest() for f in sorted(dest.rglob('*')) if f.is_file()}
(dest/'SHA256.json').write_text(json.dumps(manifest,indent=2)+'\n');print(sha);print('retained files='+str(len(manifest)))
