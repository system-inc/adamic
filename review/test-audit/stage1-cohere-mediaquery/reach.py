import pathlib,re,subprocess,json
root=pathlib.Path('/workspace/adamic');p=root/'review/test-audit/stage1-cohere-mediaquery';scratch=pathlib.Path('/tmp/u132/traced');scratch.mkdir(exist_ok=True)
entries=[]
for src in (root/'stage1/cohere/mediaquery').glob('*.ts'):
 s=src.read_text(); inserts=[]
 for m in re.finditer(r'(?:export )?function (\w+)\(|\bconstructor\(',s):
  name=m.group(1) or 'MediaNode.constructor';a=s.index('{',m.end())+1;inserts.append((a,'\nconsole.error('+json.dumps(name)+');\n'));entries.append(dict(function=name,file=str(src.relative_to(root)),line=s[:m.start()].count('\n')+1))
 for a,text in reversed(inserts):s=s[:a]+text+s[a:]
 old="const resetNode = (): MediaQueryElement => ({ before: '', after: '', value: '', type: '', sourceIndex: 0 });"
 if old in s:
  s=s.replace(old,"const resetNode = (): MediaQueryElement => { console.error('resetNode'); return { before: '', after: '', value: '', type: '', sourceIndex: 0 }; };")
  entries.append(dict(function='resetNode',file=str(src.relative_to(root)),line=src.read_text().splitlines().index('\t'+old)+1))
 (scratch/src.name).write_text(s)
cmd=['node','--disable-warning=ExperimentalWarning',str(root/'oracle/node.mjs')]
for label,entry in [('original',root/'stage1/cohere/mediaquery/main.ts'),('traced',scratch/'main.ts')]:
 with (p/(label+'-stdout.log')).open('w') as stdout,(p/(label+'-stderr.log')).open('w') as stderr:r=subprocess.run(cmd+[str(entry),'/tmp/u132/cases.txt'],cwd=root,stdout=stdout,stderr=stderr);assert r.returncode==0
assert (p/'original-stdout.log').read_bytes()==(p/'traced-stdout.log').read_bytes()
seen=set((p/'traced-stderr.log').read_text().splitlines());assert all(x['function'] in seen for x in entries)
(p/'port-functions-reached.json').write_text(json.dumps(entries,indent=2)+'\n')
with (p/'lower-functions-reached.txt').open('w') as f:
 for line in (p/'lower-functions-all.txt').read_text().splitlines():
  if not line.startswith('total:') and float(line.split()[-1].rstrip('%'))>0:f.write(line+'\n')
(p/'reach-proof.json').write_text(json.dumps(dict(port_functions=len(entries),all_observed=True,stdout_identical=True,command=cmd+['/tmp/u132/traced/main.ts','/tmp/u132/cases.txt']),indent=2)+'\n')
print('Port functions observed:',len(entries))
