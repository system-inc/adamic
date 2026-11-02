import pathlib,subprocess,json,re,collections
base=pathlib.Path('/workspace/scratch')
def run(label,tree):
 roots=sorted(str(p) for p in pathlib.Path(tree,'src/compiler').rglob('*') if p.suffix in ['.ts','.a'])
 with (base/f'types-{label}.stdout').open('w') as out,(base/f'types-{label}.stderr').open('w') as err:
  result=subprocess.run([str(base/'adamic'),'types',*roots],stdout=out,stderr=err)
 assert result.returncode==1,(label,result.returncode)
 text=(base/f'types-{label}.stderr').read_text().replace(tree,'TREE')
 starts=list(re.finditer(r'^.*?:\d+:\d+: (?:error )?TS\d+:',text,re.M))
 assert starts,(label,text[:1000])
 rows=[text[m.start():starts[i+1].start() if i+1<len(starts) else len(text)].rstrip() for i,m in enumerate(starts)]
 (base/f'diagnostics-{label}.json').write_text(json.dumps(rows,indent=2)+'\n')
 return rows,len(roots)
before,n=run('before',str(base/'before'))
after,m=run('after',str(base/'after-final'))
removed=list((collections.Counter(before)-collections.Counter(after)).elements());added=list((collections.Counter(after)-collections.Counter(before)).elements())
report={'rootsBefore':n,'rootsAfter':m,'before':len(before),'after':len(after),'removed':removed,'added':added,'beforeCodes':dict(collections.Counter(re.search(r'TS(\d+):',r)[1] for r in before)),'afterCodes':dict(collections.Counter(re.search(r'TS(\d+):',r)[1] for r in after))}
(base/'measurement.json').write_text(json.dumps(report,indent=2)+'\n')
assert not added,added
assert len(removed)==2,removed
assert all('commandLineParser.ts:' in r for r in removed)
print(json.dumps(report,indent=2))
