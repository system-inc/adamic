import subprocess,pathlib,json,time
p=pathlib.Path('/tmp/defend-fuzz');names=['TestOneSeedOneProgram','TestGeneratedProgramsCheckAndLower','TestRegexProgramsPassTheChecker','TestOverridesShapesAndLower','rest']
for name in names:
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/fuzz/','-coverpkg=./internal/fuzz','-coverprofile='+str(p/(name+'.cover')),'-run','.'if name=='rest'else'^'+name+'$']
 if name=='rest':cmd+=['-skip','^TestOneSeedOneProgram$']
 with(p/(name+'.log')).open('w')as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT)
 print(name,r.returncode,flush=True)
def blocks(name):
 result={}
 for line in(p/(name+'.cover')).read_text().splitlines()[1:]:
  span,n,count=line.split();result[span]=int(count)
 return result
comparisons=[('TestOneSeedOneProgram','rest'),('TestGeneratedProgramsCheckAndLower','TestRegexProgramsPassTheChecker'),('TestRegexProgramsPassTheChecker','TestOverridesShapesAndLower')]
data={a:dict(compared_with=b,exclusive_blocks=[k for k,v in blocks(a).items()if v and not blocks(b).get(k)])for a,b in comparisons}
(p/'coverage-differences.json').write_text(json.dumps(data,indent=2));print(json.dumps(data),flush=True)
