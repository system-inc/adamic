import pathlib,json,urllib.parse,subprocess
p=pathlib.Path('review/test-defend/stage1-typescript-scanner');source=pathlib.Path('stage1/typescript/scanner');files=['characters.ts','tokens.ts','scanner.ts','main.ts'];original={n:(source/n).read_bytes() for n in files};sides={}
for side in ['family','snapshot']:
 scripts={n:[] for n in files}
 for log in (pathlib.Path('/workspace/scanner-defend-tmp/v8')/side).glob('*.json'):
  for script in json.loads(log.read_text())['result']:
   if not script['url'].startswith('file:'):continue
   path=pathlib.Path(urllib.parse.unquote(urllib.parse.urlparse(script['url']).path))
   if path.name not in files:continue
   if not all((path.parent/n).exists() and (path.parent/n).read_bytes()==original[n] for n in files):continue
   scripts[path.name].append(script)
 sides[side]=scripts
 (p/(side+'-v8-port.json')).write_text(json.dumps(scripts,indent=2))
report=[]
for name in files:
 ranges={side:[[r for f in script['functions'] for r in f['ranges']] for script in scripts[name]] for side,scripts in sides.items()}
 points=sorted({r[k] for process in ranges.values() for rs in process for r in rs for k in ['startOffset','endOffset']})
 def executed(side,position):
  for rs in ranges[side]:
   enclosing=[r for r in rs if r['startOffset']<=position<r['endOffset']]
   if enclosing and min(enclosing,key=lambda r:r['endOffset']-r['startOffset'])['count']>0:return True
  return False
 exclusive=[]
 for a,b in zip(points,points[1:]):
  if executed('family',a) and not executed('snapshot',a):exclusive.append([a,b])
 row=dict(file=name,family_processes=len(ranges['family']),snapshot_processes=len(ranges['snapshot']),family_only_executed_js_utf16_ranges=exclusive)
 report.append(row);print(row)
(p/'coverage-diffs.json').write_text(json.dumps(report,indent=2))
for name in ['REPORT.md','rows.json','inventory.md','function-inventory.txt','matrix.json','menu.json','REPLAY.md']:
 (p/('prior-'+name)).write_bytes(subprocess.check_output(['git','show','origin/test-audit/stage1-typescript-scanner:review/test-audit/stage1-typescript-scanner/'+name]))
(p/'audit-commit.txt').write_text(subprocess.check_output(['git','rev-parse','origin/test-audit/stage1-typescript-scanner'],text=True))
