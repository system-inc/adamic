from pathlib import Path
import json,subprocess,os,time,difflib,shlex
ROOT=Path('/workspace/adamic');P=Path(__file__).resolve().parent;PKG='./stage1/cohere/lint/helpers/'
subjects=['TestMessageRefusalsMatchGo','TestKnownGapsAreExplicit'];sub='TestHelpersMatchCohere'
def go_cov(name):
 d={}
 for s in (P/(name+'.cover')).read_text().splitlines()[1:]:
  pos,stm,c=s.rsplit(' ',2);d[pos]=int(c)
 return d
def port_cov(name):
 covered={}
 for s in json.loads((P/(name+'-v8-port.json')).read_text()):
  file=Path(s['url']).name;src=(P/(file+'.transformed.js')).read_text();mask=bytearray(len(src))
  ranges=[r for f in s['functions'] for r in f['ranges']]
  for r in sorted(ranges,key=lambda r:r['endOffset']-r['startOffset'],reverse=True):
   a,b=r['startOffset'],r['endOffset'];mask[a:b]=bytes([r['count']>0])*(b-a)
  if file not in covered:covered[file]=mask
  else:covered[file]=bytearray(a|b for a,b in zip(covered[file],mask))
 return covered
comp={};b=go_cov(sub);port_b=port_cov(sub)
for name in subjects:
 a=go_cov(name);port_a=port_cov(name);ranges=[]
 for file,mask in port_a.items():
  source=(P/(file+'.transformed.js')).read_text();other=port_b.get(file,bytearray(len(mask)));unique=[i for i,(x,y) in enumerate(zip(mask,other)) if x and not y]
  spans=[]
  for i in unique:
   if spans and spans[-1][1]==i:spans[-1][1]=i+1
   else:spans.append([i,i+1])
  for start,end in spans:
   if source[start:end].strip():ranges.append({'file':file,'transformed_start':start,'transformed_end':end,'snippet':source[start:end]})
 comp[name]={'subsumer':sub,'exclusive_go_blocks':[k for k,v in a.items() if v and not b.get(k,0)],'exclusive_port_ranges':ranges}
(P/'coverage-diffs.json').write_text(json.dumps(comp,indent=2))
print([(n,len(x['exclusive_go_blocks']),len(x['exclusive_port_ranges'])) for n,x in comp.items()],flush=True)
plan=[{'id':'R1','file':'stage1/cohere/lint/helpers/policy_message.ts','old':'if(template < 0) {','new':'if(template < -1) {','aim':'Missing message must retain the Go-specific refusal, rather than falling through to an invalid arena read.'},{'id':'G1','file':'stage1/cohere/lint/helpers/strict_options.ts','old':"        if(kind === 'unsupported') { this.unsupported = 'custom or unsupported Go option type'; return false; }\n",'new':'','aim':'Unsupported custom descriptor must return NotYet even when the input is null.'}]
for m in plan:
 source=(ROOT/m['file']).read_text();assert source.count(m['old'])==1;m['line']=source.split(m['old'])[0].count('\n')+1
 (P/(m['id']+'.diff')).write_text(''.join(difflib.unified_diff(source.splitlines(True),source.replace(m['old'],m['new']).splitlines(True),fromfile='a/'+m['file'],tofile='b/'+m['file'])))
(P/'plan.json').write_text(json.dumps(plan,indent=2))
for m in plan:
 file=ROOT/m['file'];source=file.read_text()
 try:
  file.write_text(source.replace(m['old'],m['new']));env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/lint-helper-defense/cache/'+m['id']
  cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s',PKG,'-run','.'];start=time.monotonic()
  with (P/(m['id']+'.log')).open('w') as output:r=subprocess.run(cmd,cwd=ROOT,env=env,stdout=output,stderr=subprocess.STDOUT)
  info={'mutant':m['id'],'command':cmd,'env':{'ADAMIC_BUILD_CACHE_DIR':env['ADAMIC_BUILD_CACHE_DIR']},'exit':r.returncode,'wall':time.monotonic()-start};(P/(m['id']+'-run.json')).write_text(json.dumps(info,indent=2));print(m['id'],r.returncode,info['wall'],flush=True)
 finally:file.write_text(source)
