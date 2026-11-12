import pathlib,subprocess,time,shutil
root=pathlib.Path('/workspace/or3-report');out=root/'stage3/census/overload-result-02d48211'
while 'paired streams complete; ready for unit analysis' not in pathlib.Path('/tmp/or3-pipeline.log').read_text():
 if 'Traceback' in pathlib.Path('/tmp/or3-pipeline.log').read_text():raise RuntimeError('pipeline failed')
 time.sleep(5)
def run(name,cmd):
 print('start '+name,flush=True)
 with open('/tmp/or3-'+name+'.log','wb') as log:p=subprocess.run(cmd,cwd=root,stdout=log,stderr=subprocess.STDOUT)
 print(name+' child_exit='+str(p.returncode),flush=True);p.check_returncode()
run('analyze',['python3','/workspace/or3-analyze.py'])
run('classify',['python3','/workspace/or3-run.py',str(out/'classification.log'),str(out/'classification.time.json'),'node','/workspace/or3-classify.cjs',str(out/'classification-request.json'),str(out/'overloaded-value-classification.json')])
run('report',['python3','/workspace/or3-report.py'])
for name in ['pipeline.py','post.py','build.sh','run.py','checkpoint.py','analyze.py','classify.cjs','report.py','verify.py']:
 shutil.copyfile('/workspace/or3-'+name,out/name)
run('checkpoint-analysis',['python3','/workspace/or3-checkpoint.py','02d48211 paired report and classifications'])
run('validation',['python3','/workspace/or3-verify.py'])
print('report and validation complete; ready for review',flush=True)
