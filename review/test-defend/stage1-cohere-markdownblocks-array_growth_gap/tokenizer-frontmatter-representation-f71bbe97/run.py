from pathlib import Path
import subprocess,json,time,os,re,difflib
root=Path('/workspace/adamic');p=Path(__file__).resolve().parent;pkg='./stage1/cohere/markdownblocks/'
def cov(name):
 d={}
 for s in (p/(name+'.cover')).read_text().splitlines()[1:]:
  pos,stm,c=s.rsplit(' ',2);d[pos]=(int(stm),int(c))
 return d
pairs=[('TestTokenizerEvents_003','TestMicromarkInputChunks'),('TestFrontMatterStage','TestParserRepresentationProbes'),('TestParserRepresentationProbes','TestFrontMatterStage')];diffs={}
for a,b in pairs:
 x,y=cov(a),cov(b);diffs[a]={'against':b,'exclusive_blocks':[k for k,v in x.items() if v[1]>0 and y.get(k,(0,0))[1]==0]}
(p/'coverage-diffs.json').write_text(json.dumps(diffs,indent=2))
plan=[('E1','stage1/cohere/markdownblocks/tokenizerEvents.ts','this.consumed = true;','this.consumed = false;','Tokenizer consumed-state contract'),('F1','stage1/cohere/markdownblocks/frontmatter.ts',"text.indexOf('\\n', 3)","text.indexOf('\\n', 4)",'Opening newline immediately following the delimiter'),('P1','internal/lower/object.go','if len(arguments) != 1 {\n\t\t\treturn nil, true, l.notYet(node, "push with other than one value")','if len(arguments) < 1 {\n\t\t\treturn nil, true, l.notYet(node, "push with other than one value")','Accept multiple push arguments and silently lower only the first')]
for id,file,a,b,why in plan:
 s=(root/file).read_text();assert s.count(a)==1
 (p/(id+'.diff')).write_text(''.join(difflib.unified_diff(s.splitlines(True),s.replace(a,b).splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
(p/'plan.json').write_text(json.dumps([{'mutant':i,'file':f,'line':(root/f).read_text().split(a)[0].count('\n')+1,'from':a,'to':b,'aim':w} for i,f,a,b,w in plan],indent=2))
controls=['TestMicromarkInputChunks','TestFrontMatterStage','TestParserRepresentationProbes','TestArrayGrowthWitness','TestMarkdownASTPreprocessing','TestMarkdownLeafComposition_000','TestMarkdownLeafComposition_001','TestMarkdownLeafComposition_002','TestMarkdownLeafComposition_003','TestMdastIdentifierScalars','TestOptionalStringGap']
# Only valid current names enter the matrix.
scope=json.loads((p/'scope.json').read_text())['tests'];controls=[n for n in controls if n in scope]
def run(id,names):
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s',pkg,'-run','^('+'|'.join(names)+')$'];env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/markdown-defense/cache/'+id
 start=time.monotonic()
 with (p/(id+'.log')).open('w') as out:r=subprocess.run(cmd,cwd=root,env=env,stdout=out,stderr=subprocess.STDOUT)
 result={'id':id,'command':cmd,'cache':env['ADAMIC_BUILD_CACHE_DIR'],'rows':names,'exit':r.returncode,'wall':time.monotonic()-start}
 (p/(id+'-run.json')).write_text(json.dumps(result,indent=2));print(id,r.returncode,result['wall'],flush=True)
 return result
# Bounded clean control precedes all source mutations.
r=run('clean-bounded',controls+['TestTokenizerEvents_003']);assert r['exit']==0
for id,file,a,b,why in plan:
 path=root/file;s=path.read_text()
 try:
  path.write_text(s.replace(a,b))
  if file.endswith('.go'):
   with (p/(id+'-vet.log')).open('w') as out:rc=subprocess.run(['go','vet','./internal/lower/'],cwd=root,stdout=out,stderr=subprocess.STDOUT).returncode
   assert rc==0
  names=controls+([n for n in scope if re.fullmatch(r'TestTokenizerEvents_\d+',n)] if id=='E1' else ['TestTokenizerEvents_003'])
  run(id,names)
 finally:path.write_text(s)
