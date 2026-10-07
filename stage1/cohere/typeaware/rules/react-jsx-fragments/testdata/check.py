from pathlib import Path
import subprocess,time,os,json
repository=Path(__file__).resolve().parents[6]
root=repository/'stage1/cohere/typeaware/rules'
out=Path(os.environ.get('ADAMIC_WAVE15_SIXTH_ARTIFACTS', '/workspace/wave15-sixth'))
out.mkdir(parents=True, exist_ok=True)
compiler=os.environ.get('ADAMIC_WAVE15_SIXTH_COMPILER', '/workspace/wave15-sixth-adamic')
for fixture in ['fragments','undef','context']:
 (out/(fixture+'.tsx')).write_bytes((Path(__file__).parent/(fixture+'.jsx-source')).read_bytes())
(out/'controls.manifest').write_text(''.join(str(out/(name+'.tsx'))+'\n' for name in ['fragments','undef','context']))
(out/'tsconfig.json').write_text(json.dumps({'compilerOptions':{'target':'ES2022','module':'ESNext','strict':True,'jsx':'preserve'},'files':['fragments.tsx','undef.tsx','context.tsx']}))
(out/'overlay.json').write_text(json.dumps({'Replace':{str(repository/'cohere/adamic_wave15_sixth_oracle.go'):str(Path(__file__).parent/'oracle.go')}}))
build=subprocess.run(['go','build','-overlay',str(out/'overlay.json'),'-o',str(out/'oracle'),str(repository/'cohere/adamic_wave15_sixth_oracle.go')],cwd=repository/'cohere',capture_output=True)
(out/'oracle-build.stdout').write_bytes(build.stdout); (out/'oracle-build.stderr').write_bytes(build.stderr)
assert build.returncode == 0, build.stderr.decode()
def run(args,label):
 start=time.perf_counter_ns(); p=subprocess.run(args,capture_output=True)
 (out/(label+'.stdout')).write_bytes(p.stdout); (out/(label+'.stderr')).write_bytes(p.stderr)
 if p.returncode: raise RuntimeError(f'{label}: exit {p.returncode}: {p.stderr.decode()[:1500]}')
 return p.stdout,time.perf_counter_ns()-start
truth={}
for option in ['', '--element','--globals']:
 data,elapsed=run([str(out/'oracle'),str(out/'tsconfig.json'),str(out/'controls.manifest')]+([option] if option else []),'go'+(option or '-default'))
 current=''
 for line in data.splitlines(keepends=True):
  fields=line.rstrip().split(bytes([9]))
  if fields[0]==b'file': current=fields[1].decode()
  elif len(fields)>=7: truth.setdefault((current,fields[2].decode(),option),[]).append(line)
 print('Go',option or 'default','whole process ns',elapsed,flush=True)
subjects=[('react-jsx-fragments','fragments.tsx','react/jsx-fragments',''),('react-jsx-no-undef','undef.tsx','react/jsx-no-undef',''),('react-jsx-no-constructed-context-values','context.tsx','react/jsx-no-constructed-context-values','')]
for dirname,fixture,name,_ in subjects:
 d=root/dirname; expected=b''.join(sorted(truth.get((str(out/fixture),name,''),[]))); (out/(dirname+'-expected.stdout')).write_bytes(expected)
 for sanitizer in [False,True]:
  suffix='-asan' if sanitizer else ''; binary=out/(dirname+suffix)
  run([compiler,'build',str(d/'controls.a'),'-o',str(binary)]+(['--sanitize'] if sanitizer else []),dirname+suffix+'-build')
  actual,elapsed=run([str(binary)],dirname+suffix)
  assert actual==expected,(dirname,suffix,expected[:250],actual[:250])
  print(dirname+suffix,'PASS',len(actual),'bytes',len(actual.splitlines()),'findings; process ns',elapsed,flush=True)
 for option in (['--element'] if name=='react/jsx-fragments' else ['--globals'] if name=='react/jsx-no-undef' else []):
  expectedOption=b''.join(sorted(truth.get((str(out/fixture),name,option),[])))
  actual,elapsed=run([str(out/dirname),option],dirname+option)
  assert actual==expectedOption,(dirname,option)
  print(dirname,option,'PASS',len(actual),'bytes',flush=True)
 mutations={
 'react-jsx-fragments':("node.propertyName === 'Fragment'","node.propertyName === 'NeverFragment'"),
 'react-jsx-no-undef':('return first < 97 || first > 122;','return first < 97 && first > 122;'),
 'react-jsx-no-constructed-context-values':("['object', 'array',","['changed object', 'array',")}
 before,after=mutations[dirname]; original=(d/'rule.a').read_text(); assert before in original
 (d/'mutant.a').write_text(original.replace(before,after,1)); (d/'mutant_controls.a').write_text((d/'controls.a').read_text().replace("'./rule.a'","'./mutant.a'"))
 try:
  binary=out/(dirname+'-mutant'); run([compiler,'build',str(d/'mutant_controls.a'),'-o',str(binary)],dirname+'-mutant-build')
  actual,elapsed=run([str(binary)],dirname+'-mutant')
  assert actual!=expected,'mutant escaped'
  first=next((i for i,(a,b) in enumerate(zip(expected,actual)) if a!=b),min(len(expected),len(actual)))
  print(dirname,'mutant compiled, exit 0, independent Go bytes catch byte',first,flush=True)
 finally:
  (d/'mutant.a').unlink(); (d/'mutant_controls.a').unlink()

for dirname,flag,message in [
 ('react-jsx-fragments','--missing-facts','requires checker declaration facts'),
 ('react-jsx-no-undef','--missing-facts','requires checker declaration-file facts'),
 ('react-jsx-no-constructed-context-values','--gap','requires numeric value-tree, binding and memo stability adapters')]:
 d=root/dirname
 observed=subprocess.run([str(out/dirname),flag],capture_output=True)
 (out/(dirname+'-refusal.stdout')).write_bytes(observed.stdout); (out/(dirname+'-refusal.stderr')).write_bytes(observed.stderr)
 assert observed.returncode==70 and message.encode() in observed.stderr,(dirname,observed.returncode,observed.stderr)
 print(dirname,'refusal PASS exit 70',flush=True)
 original=(d/'rule.a').read_text()
 if dirname=='react-jsx-no-constructed-context-values':
  needle="panic('jsx-no-constructed-context-values requires numeric value-tree, binding and memo stability adapters');"
  replacement="return '';"
 else:
  needle='!node.factsReady'; replacement='(node.factsReady && !node.factsReady)'
 assert needle in original
 (d/'mutant.a').write_text(original.replace(needle,replacement,1))
 (d/'mutant_controls.a').write_text((d/'controls.a').read_text().replace("'./rule.a'","'./mutant.a'"))
 try:
  binary=out/(dirname+'-refusal-mutant')
  run([compiler,'build',str(d/'mutant_controls.a'),'-o',str(binary)],dirname+'-refusal-mutant-build')
  result,elapsed=run([str(binary),flag],dirname+'-refusal-mutant')
  print(dirname,'refusal mutant compiled, exit 0, caught by required exit 70',flush=True)
 finally:
  (d/'mutant.a').unlink(); (d/'mutant_controls.a').unlink()
