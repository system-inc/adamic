import pathlib, subprocess, json, shutil, re, collections
b=pathlib.Path('/workspace/scratch')
source=b/'99d-main-after'
mutant=b/'99d-main-mutant'
shutil.copytree(source/'src',mutant/'src')
p=mutant/'src/compiler/commandLineParser.ts'
s=p.read_bytes(); old=b'String(propertyOf(e, "message"))'; assert s.count(old)==1
p.write_bytes(s.replace(old,b'e.message'))
for label in ['before','after','mutant']:
 tree=b/('99d-main-'+label)
 roots=sorted(str(p) for p in (tree/'src/compiler').rglob('*') if p.suffix in ['.ts','.a'])
 with (b/('99d-main-types-'+label+'.stdout')).open('w') as out,(b/('99d-main-types-'+label+'.stderr')).open('w') as err:
  code=subprocess.run([str(b/'99d-adamic'),'types',*roots],stdout=out,stderr=err).returncode
 assert code==1,(label,code)
 print(label,len(roots),code,flush=True)
def read(label):
 text=(b/('99d-main-types-'+label+'.stderr')).read_text()
 text=text.replace(str(b/('99d-main-'+label)),'TREE').replace('../99d-main-'+label,'TREE')
 if label!='before':
  text=re.sub(r'(commandLineParser\.ts:)(\d+)(:\d+:)',lambda m:m[1]+str(int(m[2])-1 if int(m[2])>124 else int(m[2]))+m[3],text)
 rows=[]
 for line in text.splitlines():
  if re.match(r'^.*:\d+:\d+: error TS\d+:',line): rows.append(line)
  else:
   assert rows,repr(line)
   rows[-1]+='\n'+line
 return collections.Counter(rows)
before,after,mutant=map(read,['before','after','mutant'])
removed=list((before-after).elements()); added=list((after-before).elements())
assert len(removed)==1 and not added,(removed,added)
assert 'commandLineParser.ts:2301:' in removed[0] and 'TS18046' in removed[0],removed
assert mutant==before,(list((mutant-before).elements()),list((before-mutant).elements()))
result={'wholeProgram':{'before':sum(before.values()),'after':sum(after.values()),'mutant':sum(mutant.values())},'removed':removed,'added':added,'mutantRestoresExactBeforeMultiset':True,'normalization':'Tree roots replaced; after/mutant commandLineParser positions after the added import map back one line; columns and complete elaboration chains retained.'}
(b/'99d-main-diagnostics.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps(result,indent=2))
