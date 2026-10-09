import pathlib,subprocess,time,json,os
r=pathlib.Path('/workspace/adamic');p=r/'review/test-audit/stage1-cohere-tsprinter-class_interface_gap';res=[]
def run(id,cmd):
 t=time.monotonic()
 with (p/(id+'.log')).open('w') as f:code=subprocess.run(cmd,cwd=r,stdout=f,stderr=subprocess.STDOUT).returncode
 res.append(dict(id=id,command=cmd,seconds=time.monotonic()-t,exit=code));(p/'finish-runs.json').write_text(json.dumps(res,indent=2));print(id,code,res[-1]['seconds'],flush=True);return code
f=r/'internal/lower/lower.go';s=f.read_text();subprocess.run(['git','apply',str(p/'P-Lower.diff')],cwd=r,check=True)
try:run('P-Lower-vet',['timeout','90','go','vet','./internal/lower/'])
finally:f.write_text(s)
regex='^(TestExpressionsAgainstGoAndPrettier_000|TestMutants_000)$'
code=run('baseline-family-members',['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/tsprinter/','-run',regex])
if code:raise SystemExit('narrowed family baseline not green; no family mutation inference')
f=r/'stage1/cohere/tsprinter/doc.ts';s=f.read_text();subprocess.run(['git','apply',str(p/'M1.diff')],cwd=r,check=True)
try:
 os.environ['ADAMIC_BUILD_CACHE_DIR']='/workspace/u140-cache/M1-family'
 run('M1-family-member',['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/tsprinter/','-run','^TestExpressionsAgainstGoAndPrettier_000$'])
finally:f.write_text(s)
