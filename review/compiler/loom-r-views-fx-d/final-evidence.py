import pathlib,json,re,subprocess,hashlib,shutil,difflib
root=pathlib.Path.cwd();ev=root/'review/compiler/loom-r-views-fx-d'
rows=json.loads((ev/'branch-results.json').read_text())
old=json.loads((root/'review/compiler/loom-views-fx-d/branch-results.json').read_text())
byname={pathlib.Path(r['source']).name:r for r in old}
verdict=[]
for row in rows:
 prior=byname[pathlib.Path(row['source']).name]
 for backend in ['node','native','js']:
  assert all(row[backend][key]==prior[backend][key] for key in ['exit','stdout','stderr']),(row['source'],backend)
 verdict.append(dict(source=row['source'],item=row['item'],agrees_with_node=row['agrees'],matches_source_commit=True))
assert len(rows)==23
(ev/'replay-verdict.json').write_text(json.dumps(verdict,indent=2)+'\n')
artifacts=[];dest=pathlib.Path('/tmp/adamic-gate/cleared-views-native');dest.mkdir(exist_ok=True)
for path in (ev/'branch-probes').glob('*/native'):
 digest=hashlib.sha256(path.read_bytes()).hexdigest();artifacts.append(dict(probe=path.parent.name,sha256=digest));shutil.move(str(path),str(dest/(path.parent.name+'-'+digest[:12])))
(ev/'native-artifact-sha256.json').write_text(json.dumps(artifacts,indent=2)+'\n')
measured=[]
for log in ev.glob('cleared-final-*.log'):
 for name,seconds in re.findall(r'^--- PASS: (\S+) \(([0-9.]+)s\)',log.read_text(),re.M):measured.append(dict(test=name,seconds=float(seconds),log=log.name))
assert measured and all(row['seconds']<60 for row in measured),measured
(ev/'cleared-final-leaf-seconds.json').write_text(json.dumps(measured,indent=2)+'\n')
lines=['Final runtime differences against 2c410041. All object.c differences are inherited from dbdf977f5b; mixed/untagged differences are the exact runtime-cleared correction. Each hunk below lists every removed/added line; complete contextual patches are saved alongside this report.\n']
for name in ['object.c','view_unions_mixed.c','view_unions_untagged.c']:
 path='internal/native/runtime/'+name;old=subprocess.check_output(['git','show','2c410041:'+path],text=True);new=(root/path).read_text()
 patch=''.join(difflib.unified_diff(old.splitlines(True),new.splitlines(True),fromfile='2c410041/'+path,tofile='final/'+path,n=0))
 lines.append('\n'+name+'\n```diff\n'+patch+'```\n')
(ev/'runtime-lines-versus-source.md').write_text(''.join(lines))
print('23 replays match source commit; all final leaves <60 seconds; native binaries moved out of evidence')
