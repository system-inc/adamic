import pathlib,re,json,difflib,subprocess,time,os
p=pathlib.Path('/tmp/u129/evidence');root=pathlib.Path('stage1/cohere/markdownblocks');groups=json.loads((p/'groups.json').read_text());files={n:(root/n).read_text() for n in ['setup_deadline_markdownblocks_test.go','quote_layout_shards_test.go','structure_layout_shards_test.go','table_layout_shards_test.go','whitespace_layout_split_test.go']}
def section(text,fn):
 m=re.search(r'^func '+re.escape(fn)+r'\(',text,re.M);assert m,fn
 nxt=re.search(r'^func ',text[m.end():],re.M);end=m.end()+nxt.start() if nxt else len(text)
 return m.start(),end
plan=[('S1','setup_deadline_markdownblocks_test.go','markdownLayoutSetupContext','return context.WithCancel(parent)','return context.WithTimeout(parent, time.Nanosecond)',[0]),('S2','quote_layout_shards_test.go','quoteLayoutBuildProduct','filepath.Join(dir, "program.c")','filepath.Join(dir, "missing.c")',[2,3]),('S3','quote_layout_shards_test.go','quoteLayoutNativeProduct','filepath.Join(dir, "port")','filepath.Join(dir, "")',[3]),('S4','structure_layout_shards_test.go','structureLayoutBuildProduct','filepath.Join(dir, "program.c")','filepath.Join(dir, "missing.c")',[4]),('S5','table_layout_shards_test.go','tableLayoutBuildProduct','filepath.Join(directory, "program.c")','filepath.Join(directory, "missing.c")',[5]),('S6','whitespace_layout_split_test.go','whitespaceLayoutBuildProduct','filepath.Join(dir, "program.c")','filepath.Join(dir, "missing.c")',[6,7,8,9]),('S7','whitespace_layout_split_test.go','whitespaceLayoutPrepare','filepath.Join(directory, "products.json")','filepath.Join(directory, "missing.json")',[9]),('S8','whitespace_layout_split_test.go','whitespaceLayoutNativeProduct','filepath.Join(dir, "port")','filepath.Join(dir, "")',[7,9]),('S9','whitespace_layout_split_test.go','whitespaceLayoutPolicySetup','filepath.Join(dir, "port")','filepath.Join(dir, "")',[8,9])]
probes=[('P1','setup_deadline_markdownblocks_test.go','markdownLayoutSetupContext','return nil, func() {}',[0]),('P2','quote_layout_shards_test.go','quoteLayoutBuildProduct','return quoteLayoutProducts{}',[2]),('P3','quote_layout_shards_test.go','quoteLayoutNativeProduct','return ""',[3]),('P4','structure_layout_shards_test.go','structureLayoutBuildProduct','return structureLayoutProducts{}',[4]),('P5','table_layout_shards_test.go','tableLayoutBuildProduct','return tableLayoutProducts{}',[5]),('P6','whitespace_layout_split_test.go','whitespaceLayoutBuildProduct','return whitespaceLayoutProducts{}',[6]),('P7','whitespace_layout_split_test.go','whitespaceLayoutNativeProduct','return ""',[7]),('P8','whitespace_layout_split_test.go','whitespaceLayoutPolicySetup','return whitespaceLayoutProducts{}',[8]),('P9','whitespace_layout_split_test.go','whitespaceLayoutReadyProducts','return whitespaceLayoutProducts{}, whitespaceLayoutProducts{}',[9])]
switched=dict(files);metadata=[]
for id,file,fn,old,new,indices in plan:
 text=files[file];a,b=section(text,fn);pos=text.index(old,a,b);assert pos>=0
 pure=text[:pos]+new+text[pos+len(old):];(p/(id+'.diff')).write_text(''.join(difflib.unified_diff(text.splitlines(True),pure.splitlines(True),fromfile='a/'+str(root/file),tofile='b/'+str(root/file))))
 # Different expression types: context statement versus path expressions.
 if id=='S1':replacement='if markdownAuditMutant("S1") { '+new+' }; '+old
 else:replacement='markdownAuditPath("'+id+'", '+old+', '+new+')'
 a,b=section(switched[file],fn);pos=switched[file].index(old,a,b);switched[file]=switched[file][:pos]+replacement+switched[file][pos+len(old):]
 metadata.append(dict(id=id,kind='construction',file=str(root/file),line=text[:text.index(old,a if a<len(text) else 0)].count('\n')+1,function=fn,old=old,new=new,groups=indices))
for id,file,fn,ret,indices in probes:
 text=files[file];a,b=section(text,fn);brace=text.index('{',a,b);pure=text[:brace+1]+'\n if true { '+ret+' }'+text[brace+1:]
 (p/(id+'.diff')).write_text(''.join(difflib.unified_diff(text.splitlines(True),pure.splitlines(True),fromfile='a/'+str(root/file),tofile='b/'+str(root/file))))
 a,b=section(switched[file],fn);brace=switched[file].index('{',a,b);switched[file]=switched[file][:brace+1]+'\n if markdownAuditMutant("'+id+'") { '+ret+' }'+switched[file][brace+1:]
 metadata.append(dict(id=id,kind='probe',file=str(root/file),line=text[:text.index('{',section(text,fn)[0])].count('\n')+1,function=fn,groups=indices,change=ret))
switched['setup_deadline_markdownblocks_test.go']+='\nfunc markdownAuditMutant(id string) bool { return os.Getenv("ADAMIC_MUTANT") == id }\nfunc markdownAuditPath(id, clean, broken string) string { if markdownAuditMutant(id) { return broken }; return clean }\n'
(p/'plan.json').write_text(json.dumps(metadata,indent=2));results=[]
def run(id,kind,cmd,env=None):
 s=time.monotonic()
 with (p/(id+'-'+kind+'.log')).open('w') as out:r=subprocess.run(cmd,stdout=out,stderr=subprocess.STDOUT,env=env)
 item=dict(id=id,kind=kind,exit=r.returncode,seconds=time.monotonic()-s,command=cmd);results.append(item);(p/'matrix-timings.json').write_text(json.dumps(results,indent=2));print(item,flush=True);return r.returncode
try:
 for file,text in switched.items():(root/file).write_text(text)
 assert run('switch','vet',['go','vet','./stage1/cohere/markdownblocks/'])==0
 # A new cache is necessary because these deliberate construction faults must rerun callbacks.
 for item in metadata:
  indices=item['groups'];members=sum([groups[i][1] for i in indices],[]);id=item['id'];env=dict(os.environ,ADAMIC_MUTANT=id,ADAMIC_BUILD_CACHE_DIR='/tmp/u129/cache/'+id)
  run(id,'matrix',['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/markdownblocks/','-run','^('+'|'.join(members)+')$'],env)
finally:
 for file,text in files.items():(root/file).write_text(text)
