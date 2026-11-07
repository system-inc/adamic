"""Validate the reported accounting against raw streams and frozen sources."""
import collections,hashlib,json,os,re,subprocess,tarfile
from pathlib import Path
repo=Path('/workspace/adamic');out=Path(__file__).resolve().parent;core=repo/'stage3/census/core-sites-af42d3dc';scratch=Path('/workspace/scratch/after-wave');tree=Path('/workspace/after-wave-adapted')
load=lambda p:json.loads(p.read_text())
rows=[r for p in sorted(out.glob('group-*-rows.json')) for r in load(p)]
assert len(rows)==79 and len({r['roots'][0] for r in rows})==79
assert [len(load(p)) for p in sorted(out.glob('group-*-rows.json'))]==[20,20,20,19]
groups=load(out/'grouped-first-stops.json');assert sum(g['count'] for g in groups)==79
root_names={str(Path(r['roots'][0]).relative_to(tree/'src/compiler')) for r in rows}
assert root_names=={r for g in groups for r in g['roots']}
assert all(len(g['roots'])==g['count'] for g in groups)
for d in [out,core]:
 m=load(d/'adapted-source-manifest.json')
 assert all(hashlib.sha256((tree/r['file']).read_bytes()).hexdigest()==r['sha256'] for r in m)
 assert load(d/'provenance.json')['main_sha']=='af42d3dc1261a3db536064c3dfbb5eaa7e8b6405'
 assert load(d/'provenance.json')['typescript_pin']=='050880ce59e30b356b686bd3144efe24f875ebc8'
 if (d/'raw-streams.tar.gz').exists():
  archive=d/'raw-streams.tar.gz';checksum=(d/'raw-streams.tar.gz.sha256').read_text().split()[0]
  assert hashlib.sha256(archive.read_bytes()).hexdigest()==checksum
  with tarfile.open(archive) as t:
   if d==out:
    for group in range(1,5):
     raw=[json.loads(l) for l in t.extractfile(f'group-{group}.jsonl').read().splitlines()]
     expected=load(out/f'group-{group}-rows.json');assert len(raw)==2*len(expected)
     for i,r in enumerate(expected):
      assert r==raw[2*i]
      assert {k:v for k,v in r.items() if k!='seconds'}=={k:v for k,v in raw[2*i+1].items() if k!='seconds'}
   else:
    raw=[json.loads(l) for l in t.extractfile('raw.jsonl').read().splitlines()]
    assert raw[0]['latent_mode']=='full';assert raw[0]['diagnostics']==[]
    record=next(r for r in raw if r.get('file','').endswith('/core.ts'))
    assert len(record['units'])==len(load(core/'rows.json')['units'])
cr=load(core/'rows.json');sites=cr['sites'];emitters=load(core/'emitters.json');diff=load(core/'comparison.json')
assert len(diff['gone'])+len(diff['remain'])==48
assert len(diff['new'])+len(diff['remain'])==len(sites)
assert set(emitters)=={s['id'] for s in sites}
assert len({tuple(s[k] for k in ['where','kind','reason','phase']) for s in sites})==len(sites)
assert [tuple(map(int,s['where'].rsplit(':',2)[1:])) for s in sites]==sorted(tuple(map(int,s['where'].rsplit(':',2)[1:])) for s in sites)
assert cr['first_walls']==[s['id'] for s in sites if s['instantiation_independent']]
for es in emitters.values():
 for e in es:assert (repo/e['file']).read_text().splitlines()[e['line']-1]==e['source_line']
for excerpt in load(out/'source-excerpts.json'):
 assert excerpt['text'] in Path(excerpt['file']).read_bytes().decode()
 if excerpt['scope_end']-excerpt['scope_start']+1<30:assert excerpt['whole_function']
for binary in load(out/'binaries.json'):
 assert hashlib.sha256(Path(binary['binary']).read_bytes()).hexdigest()==binary['sha256']
assert load(core/'provenance.json')['adapter_count']==94
for where,es in load(out/'new-stop-emitters.json').items():
 for e in es:assert (repo/e['file']).read_text().splitlines()[e['line']-1]==e['source_line']
for overlay in ['latent-overlay','entry-overlay']:
 assert not list((scratch/overlay).glob('*.go'))
 mapping=load(scratch/overlay/'overlay.json')['Replace'];assert all(v.endswith('.go.txt') for v in mapping.values())
statuses=[load(p) for p in sorted(scratch.glob('root-*-status.json'))]
assert len(statuses)==79 and all(s['exit_code']==0 and s['process_group_cleaned'] for s in statuses)
assert load(scratch/'core-census-status.json')['exit_code']==0
# PIDs can be reused, so check a live /proc cmdline for the census executables, not PID existence alone.
live=[]
for p in Path('/proc').iterdir():
 if not p.name.isdigit():continue
 try:cmd=(p/'cmdline').read_bytes().split(b'\0')
 except (FileNotFoundError,PermissionError,ProcessLookupError):continue
 if cmd and cmd[0].decode(errors='replace') in [str(scratch/n) for n in ['census','latent','entry-census']]:live.append(int(p.name))
assert not live,live
result={'ordinary_roots':79,'group_sizes':[20,20,20,19],'ordinary_lower_whole':sum(r['kind']=='accepted' for r in rows),'sites':len(sites),'first_walls':len(cr['first_walls']),'old_sites':48,'gone':len(diff['gone']),'remain':len(diff['remain']),'new':len(diff['new']),'sources_unchanged':True,'all_emitters_checked':True,'generated_go_suffixes_checked':True,'live_census_processes':live,'entry_status':load(out/'root-order.json')['status']}
for d in [out,core]:(d/'validation.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result,indent=2))
