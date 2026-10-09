import pathlib,re,json,difflib,subprocess,time
root=pathlib.Path('/workspace/adamic'); out=root/'review/test-audit/stage1-cohere-json-audit';out.mkdir(parents=True,exist_ok=True)
files=['internal/lower/object.go','internal/lower/exceptions.go','internal/native/emit_expressions.go','internal/lower/lower.go']+['stage1/cohere/json/'+x+'_test.go' for x in ['audit','builds','chunks','pin','sample','repository','performance']]
base={f:subprocess.check_output(['git','show','origin/main:'+f],cwd=root,text=True) for f in files}
# A Go token scan locates whole bodies without counting braces inside strings/comments.
def body(s,name):
 m=re.search(r'^func (?:\([^\n]*\) )?'+re.escape(name)+r'\(',s,re.M); assert m,name
 start=s.index('{',m.end()); depth=0
 tokens=re.finditer(r'//[^\n]*|/\*[\s\S]*?\*/|"(?:\\.|[^"\\])*"|`[^`]*`|\'(?:\\.|[^\'\\])*\'|[{}]',s[start:])
 for t in tokens:
  if t.group()=='{':depth+=1
  elif t.group()=='}':
   depth-=1
   if depth==0:return start,start+t.end()
 raise ValueError(name)
def replacebody(s,name,value):
 a,b=body(s,name);return s[:a]+'{\n\t'+value+'\n}'+s[b:]
def early(s,name,id,value):
 a,b=body(s,name);return s[:a+1]+'\n\tif auditU102() == "'+id+'" { '+value+' }'+s[a+1:]
def prune(s):
 m=re.search(r'import \(([\s\S]*?)\n\)',s)
 if not m:return s
 rest=s[:m.start()]+s[m.end():]
 lines=[]
 for line in m.group(1).splitlines():
  q=re.search(r'(?:(\w+)\s+)?"([^"]+)"',line)
  if not q: lines.append(line);continue
  alias=q.group(1) or q.group(2).split('/')[-1]
  if re.search(r'\b'+re.escape(alias)+r'\.',rest):lines.append(line)
 return s[:m.start()]+'import (\n'+'\n'.join(lines)+'\n)'+s[m.end():]
plan=[]
def add(id,kind,f,old,new,func=None):
 s=base[f]
 if func:
  a,b=body(s,func);line=s[:a].count('\n')+1; changed=replacebody(s,func,new)
 else:
  assert s.count(old)==1,(id,s.count(old));line=s[:s.index(old)].count('\n')+1;changed=s.replace(old,new)
 changed=prune(changed)
 plan.append(dict(id=id,kind=kind,file=f,line=line,old=old,new=new,function=func,source=changed))
add('M01','production','internal/lower/object.go','"push with other than one value"','"push with other than two values"')
add('M02','production','internal/lower/exceptions.go','"a try around "','"a catch around "')
add('M03','production','internal/native/emit_expressions.go','"(double)%s->length"','"((double)%s->length + 1)"')
loop='''for i, flag := range compilerFlags {
		if strings.HasPrefix(flag, "-ffile-prefix-map=") && strings.HasSuffix(flag, "=/tmp/go-build") {
			compilerFlags[i] = "-ffile-prefix-map=<Go scratch>=/tmp/go-build"
		}
	}'''
add('S01','setup','stage1/cohere/json/builds_test.go',loop,'')
add('S02','setup','stage1/cohere/json/chunks_test.go','count := min(cores, 8, total)','count := min(cores, 7, total)')
add('S03','setup','stage1/cohere/json/sample_test.go','strings.HasPrefix(item.Name, "generated/") || included[item.Name]','included[item.Name]')
add('S04','setup','stage1/cohere/json/chunks_test.go','chunks[index].end = (index + 1) * total / count','chunks[index].end = (index + 1) * total / count - 1')
add('W01','witness','stage1/cohere/json/audit_test.go','body','return true','sameAnswer')
add('W02','witness','stage1/cohere/json/chunks_test.go','body','return run{stdout: single}, nil','joinNativeChunks')
add('W03','witness','stage1/cohere/json/pin_test.go','body','return nil','corpusPinError')
add('W04','witness','stage1/cohere/json/performance_test.go','string(side.result.stdout) == expected','true')
add('W05','witness','stage1/cohere/json/repository_test.go','body','return corpusPin{}, 0, nil','validateCorpus')
for id,f,func,val in [('P01','internal/lower/lower.go','Lower','return nil, nil'),('P02','stage1/cohere/json/builds_test.go','jsonGoToolchain','return nil, nil'),('P03','stage1/cohere/json/chunks_test.go','chunkRanges','return nil'),('P04','stage1/cohere/json/chunks_test.go','joinNativeChunks','return run{}, nil'),('P05','stage1/cohere/json/audit_test.go','sameAnswer','return false'),('P06','stage1/cohere/json/pin_test.go','corpusPinError','return nil'),('P07','stage1/cohere/json/repository_test.go','validateCorpus','return corpusPin{}, 0, nil'),('P08','stage1/cohere/json/sample_test.go','sampledCorpusCases','return nil')]:add(id,'probe',f,'body',val,func)
add('P09','probe','stage1/cohere/json/performance_test.go','string(side.result.stdout) == expected','false')
# Fixed before any changed-code test run.
(out/'plan.txt').write_text(json.dumps([{k:v for k,v in x.items() if k!='source'} for x in plan],indent=2)+'\n')
(out/'functions.txt').write_text('Code under test: lowering.arrayMethodArguments, lowering.exceptions, emitter.evaluate; entry lower.Lower.\nConstruction/check entries: jsonGoToolchain, chunkRanges, joinNativeChunks, sameAnswer, corpusPinError, validateCorpus, sampledCorpusCases, inline stdout equality in TestCachedWidthMutantIsCaught.\nSupport reached in stage1/cohere/json: jsonBuildEnvironment, buildJSONGoOracle, jsonGoOracleUnit, jsonBuildUnit, jsonRunBuildUnit, oracleAnswers, cohereAnswers, writeJSON, bounded, fileCorpus, corpusCases, generatedCases, verifyCorpusPin, pinForCases, caseGroup, gitCorpus, repositoryJSONAtHEAD, lowered, execute, executeResult, onNode, onJavaScriptBackend, natively, leaks, portDirectory, escape, protocol, comparisonError. These are preparation/harness, excluded from production mutation.\nTestProfileSnapshotsAgree did not reach its snapshot entry because artifacts are absent. TestUpstreamNumericSeparatorGap invokes only outside oracles and cannot be production-mutated under this brief.\n')
for x in plan:
 patch=''.join(difflib.unified_diff(base[x['file']].splitlines(True),x['source'].splitlines(True),fromfile='a/'+x['file'],tofile='b/'+x['file']))
 (out/(x['id']+'.diff')).write_text(patch)
# Install switched scratch, preserving all original logic on empty selector.
sw=dict(base)
for x in plan:
 id,f=x['id'],x['file']
 if x['function']:
  sw[f]=early(sw[f],x['function'],id,x['new'])
 elif id=='S01':sw[f]=sw[f].replace(loop,'if auditU102() != "S01" {\n'+loop+'\n}')
 elif id=='S02':sw[f]=sw[f].replace(x['old'],'count := min(cores, 8, total)\n if auditU102() == "S02" { count = min(cores, 7, total) }')
 elif id=='S03':sw[f]=sw[f].replace(x['old'],'(auditU102() != "S03" && strings.HasPrefix(item.Name, "generated/")) || included[item.Name]')
 elif id=='S04':sw[f]=sw[f].replace(x['old'],x['old']+'\n if auditU102() == "S04" { chunks[index].end-- }')
 elif id in ['W04','P09']:
  if id=='W04':sw[f]=sw[f].replace(x['old'],'auditU102() == "W04" || (auditU102() != "P09" && '+x['old']+')')
 else:sw[f]=sw[f].replace(x['old'],'auditU102Text("'+id+'", '+x['old']+', '+x['new']+')')
for f,s in sw.items():(root/f).write_text(s)
helper='''package PACKAGE
import "os"
func auditU102() string { return os.Getenv("ADAMIC_MUTANT") }
func auditU102Text(id, original, changed string) string { if auditU102() == id { return changed }; return original }
'''
for f,pkg in [('internal/lower/audit_u102.go','lower'),('internal/native/audit_u102.go','native'),('stage1/cohere/json/audit_u102_test.go','json')]: (root/f).write_text(helper.replace('PACKAGE',pkg))
subprocess.run(['gofmt','-w']+[str(root/f) for f in files]+[str(root/f) for f in ['internal/lower/audit_u102.go','internal/native/audit_u102.go','stage1/cohere/json/audit_u102_test.go']],check=True)
(out/'switch.diff').write_bytes(subprocess.check_output(['git','diff','--']+files,cwd=root))
pathlib.Path('/tmp/u102/plan.json').write_text(json.dumps(plan))
print('Fixed plan saved: 3 production, 4 setup, 5 witness, 9 probes')
