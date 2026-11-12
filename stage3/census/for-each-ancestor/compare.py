from pathlib import Path
import difflib,json
out=Path('/workspace/adamic/stage3/census/for-each-ancestor');scratch=Path('/workspace/scratch/ancestor')
s=json.loads((out/'stock.json').read_text());baseline=s['candidates']['baseline'];summary={}
for name,row in s['candidates'].items():
 changed=[f for f in sorted(set(baseline['outputs'])|set(row['outputs'])) if baseline['outputs'].get(f)!=row['outputs'].get(f)]
 diff=''
 for f in changed:
  a=scratch/'emit-baseline'/f;b=scratch/('emit-'+name)/f
  diff+=''.join(difflib.unified_diff(a.read_text().splitlines(True) if a.exists() else [], b.read_text().splitlines(True) if b.exists() else [],fromfile='baseline/'+f,tofile=name+'/'+f))
 (out/(name+'-javascript.diff')).write_text(diff)
 sources=['src/compiler/path.ts','src/compiler/watchUtilities.ts','src/typingsInstallerCore/typingsInstaller.ts']
 diff=''
 for f in sources:
  a=(scratch/'baseline'/f).read_text();b=(scratch/name/f).read_text() if name!='baseline' else a
  diff+=''.join(difflib.unified_diff(a.splitlines(True),b.splitlines(True),fromfile='baseline/'+f,tofile=name+'/'+f))
 (out/(name+'-source.diff')).write_text(diff)
 summary[name]={'roots':len(row['roots']),'diagnostics':len(row['diagnostics']),'emitted_js_files':len(row['outputs']),'javascript_byte_identical':not changed,'changed_js':changed}
(out/'comparison.json').write_text(json.dumps(summary,indent=2)+'\n');print(json.dumps(summary,indent=2))
