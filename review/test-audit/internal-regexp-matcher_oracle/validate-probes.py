import pathlib,subprocess,json,time
D=pathlib.Path('review/test-audit/internal-regexp-matcher_oracle'); originals={p:p.read_text() for p in pathlib.Path('internal/regexp').glob('*.go')};results=[]
for diff in sorted(D.glob('P*.diff'))+[D/'W01.diff']:
 t=time.monotonic();subprocess.run(['git','apply','--check',str(diff)],check=True);subprocess.run(['git','apply',str(diff)],check=True)
 with (D/('vet-'+diff.stem+'.log')).open('w') as out:rc=subprocess.call(['timeout','90','go','vet','./internal/regexp/'],stdout=out,stderr=subprocess.STDOUT)
 results.append(dict(diff=diff.name,exit=rc,wall_seconds=round(time.monotonic()-t,3)))
 for p,s in originals.items():p.write_text(s)
 for p in pathlib.Path('internal/regexp').glob('*.go'):
  if p not in originals:p.unlink()
 assert rc==0,(diff,rc)
(D/'probe-validation.json').write_text(json.dumps(results,indent=2))
