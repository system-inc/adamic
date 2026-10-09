import pathlib,subprocess,os,time,json
root=pathlib.Path(__file__).resolve().parents[3];out=root/'review/compiler/fuzz-lower-guards';env=os.environ.copy();env['GOMAXPROCS']='4';rows=[]
plans=[('M16','internal/fuzz','TestReduceRejectsDifferentRefusal'),('M05','internal/fuzz','TestExactSignatureRejectsSuffix'),('M14','internal/fuzz','TestBytesSharedRejects63Bytes'),('M04','internal/lower','TestArgumentsLengthReadKeepsReaderFact')]
def run(label,package,test,expected):
 cmd=['go','test','./'+package,'-run','^'+test+'$','-count=1','-parallel=4','-timeout=90s','-json'];start=time.monotonic()
 with (out/(label+'.jsonl')).open('w') as log:r=subprocess.run(cmd,cwd=root,env=env,stdout=log,stderr=log,timeout=120)
 events=[json.loads(s) for s in (out/(label+'.jsonl')).read_text().splitlines() if s.startswith('{')];terminal=[e for e in events if e.get('Test')==test and e.get('Action') in ['pass','fail','skip']]
 row=dict(label=label,command=cmd,exit=r.returncode,seconds=time.monotonic()-start,result=terminal[-1] if terminal else None);rows.append(row);print(row,flush=True)
 assert row['result'] and row['result']['Action']==expected,row
 assert (r.returncode==0)==(expected=='pass'),row
for mutant,package,test in plans:
 run(mutant+'-baseline',package,test,'pass')
 patch=out/(mutant+'.diff');subprocess.run(['git','apply','--check',str(patch)],cwd=root,check=True,timeout=15);subprocess.run(['git','apply',str(patch)],cwd=root,check=True,timeout=15)
 try:run(mutant+'-mutant',package,test,'fail')
 finally:subprocess.run(['git','apply','--reverse',str(patch)],cwd=root,check=True,timeout=15)
 run(mutant+'-restored',package,test,'pass')
(out/'results.json').write_text(json.dumps(rows,indent=2)+'\n')
