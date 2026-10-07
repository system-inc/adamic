"""Fixed private-Go parity and explicit evidence for the required runtime constructor gap."""
import json,subprocess,tempfile
from pathlib import Path
HERE=Path(__file__).resolve().parent
COMPILER=Path('/workspace/adamic')
EVIDENCE=HERE/'evidence'
EVIDENCE.mkdir(exist_ok=True)
TOOL='/tmp/wave10-regex-adamic'
def run(name,args,cwd=COMPILER):
 out=EVIDENCE/('regex-contract-'+name+'.stdout.txt');err=EVIDENCE/('regex-contract-'+name+'.stderr.txt')
 with out.open('wb') as stdout,err.open('wb') as stderr: result=subprocess.run(args,cwd=cwd,stdout=stdout,stderr=stderr)
 return result.returncode,out.read_bytes(),err.read_bytes()
with tempfile.TemporaryDirectory(prefix='wave10-regex-contract-') as directory:
 scratch=Path(directory);cohere=COMPILER/'cohere';virtual=cohere/'adamic_wave10_fixed_pattern.go'
 overlay=scratch/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(HERE/'testdata/fixed-pattern-oracle.go.txt'),str(cohere/'internal/lint/rules/core/adamic_wave10_fixed_pattern.go'):str(HERE/'testdata/fixed-pattern-export.go.txt')}}))
 status,want,error=run('go',['go','run','-overlay='+str(overlay),str(virtual)],cohere);assert status==0 and not error,(status,error)
 runner=str(COMPILER/'oracle/node.mjs')
 def sides(entry,tag):
  binary=scratch/(tag+'-native');module=scratch/(tag+'.mjs')
  status,_,error=run(tag+'-build',[TOOL,'build',str(entry),'-o',str(binary),'--sanitize']);assert status==0 and not error,(status,error)
  status,source,error=run(tag+'-emit',[TOOL,'js',str(entry)]);assert status==0 and not error,(status,error);module.write_bytes(source)
  results=[]
  for name,args in [('Node',['node','--disable-warning=ExperimentalWarning',runner,str(entry)]),('JS',['node','--disable-warning=ExperimentalWarning',runner,str(module)]),('native',[str(binary)])]:
   status,data,error=run(tag+'-'+name,args);assert status==0 and not error,(name,status,error);results.append((name,data))
  return results
 for name,data in sides(HERE/'testdata/fixed-pattern.a','fixed'):assert data==want,(name,data,want)
 print('fixed table literal: 32 queries, private Go and all three runtimes match',len(want),'bytes')
 source=(HERE/'fixed_pattern.a').read_text();assert source.count('(?![\\s\\S])')==1
 (scratch/'fixed_pattern.a').write_text(source.replace('(?![\\s\\S])','',1))
 (scratch/'driver.a').write_text((HERE/'testdata/fixed-pattern.a').read_text().replace('../fixed_pattern.a','./fixed_pattern.a'))
 for name,data in sides(scratch/'driver.a','anchor-mutant'):assert data!=want,(name,'mutant survived')
 print('absolute_end_removed compiles, exits zero with empty stderr on every runtime; only private-Go output comparison catches it')
 entry=HERE/'testdata/runtime-pattern.a'
 status,data,error=run('dynamic-Node',['node','--disable-warning=ExperimentalWarning',runner,str(entry),'TODO']);assert (status,data,error)==(0,b'true\n',b'')
 for name,args in [('native',[TOOL,'build',str(entry),'-o',str(scratch/'dynamic'),'--sanitize']),('JS',[TOOL,'js',str(entry)])]:
  status,data,error=run('dynamic-'+name,args);assert status!=0 and b'RegExp with a nonconstant pattern' in error,(name,status,data,error)
 print('BLOCKED: required dynamic constructor succeeds on source Node; emitted JS and native lowering refuse RegExp with a nonconstant pattern')
 status,data,error=run('guard-Node',['node','--disable-warning=ExperimentalWarning',runner,str(HERE/'testdata/pattern-refusal.a')]);assert status==70 and b'Go commentPattern option parity' in error,(status,data,error)
 source=(HERE/'pattern.a').read_text().replace("'./fixed_pattern.a'",json.dumps(str(HERE/'fixed_pattern.a')))
 assert source.count('if(!this.fixed)')==1
 (scratch/'pattern.a').write_text(source.replace('if(!this.fixed)','if(false)',1))
 (scratch/'guard.a').write_text((HERE/'testdata/pattern-refusal.a').read_text().replace('../pattern.a','./pattern.a'))
 status,data,error=run('guard-mutant-Node',['node','--disable-warning=ExperimentalWarning',runner,str(scratch/'guard.a')]);assert status==0 and not error and b'true' in data,(status,data,error)
 print('option_parity_guard_removed caught on source Node only; no emitted/native mutant credit because lowering is blocked')

with tempfile.TemporaryDirectory(prefix='wave10-option-dialect-') as directory:
 source=Path(directory)/'main.go';source.write_text((HERE/'testdata/option-dialect.go.txt').read_text())
 status,go,error=run('option-dialect-Go',['go','run',str(source)]);assert status==0 and not error
 status,node,error=run('option-dialect-Node',['node',str(HERE/'testdata/option-dialect.mjs')]);assert status==0 and not error
 assert go==b'false\ntrue\ntrue\n' and node==b'true\nfalse\nSyntaxError\n',(go,node)
 print('BLOCKED: raw option dialect differs: Go ASCII whitespace, no-fallthrough case-insensitivity, Go inline flag syntax')
