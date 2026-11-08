import pathlib,subprocess,os,time,re,sys,json
out=pathlib.Path('/tmp/fa-report/stage3/census/find-ancestor-5a86e3b3')
# Wait for setup to close and apply to report its own exit before running census.
while not (out/'apply.status.json').exists() or not pathlib.Path('/tmp/fa-setup.status').exists():time.sleep(2)
assert json.loads((out/'apply.status.json').read_text())['exit']==0
assert pathlib.Path('/tmp/fa-setup.status').read_text().strip()=='0'
(out/'setup.log').write_bytes(pathlib.Path('/tmp/fa-setup.log').read_bytes())
subprocess.run(['git','rev-parse','HEAD'],cwd='/tmp/fa-report/cohere',check=True)
head=pathlib.Path('/tmp/fa-head/cohere')
if not head.is_symlink():
 if head.exists():head.rmdir()
 head.symlink_to('/tmp/fa-report/cohere',target_is_directory=True)
for side,root in [('parent','/tmp/fa-report'),('head','/tmp/fa-head')]:
 text=pathlib.Path(root+'/go.mod').read_text();pairs=re.findall(r'^\s*(\S+) => (\./cohere\S*)',text,re.M)
 replacements={m:str(pathlib.Path('/tmp/fa-report')/p) for m,p in pairs if m!='github.com/microsoft/TypeScript/tsc'}
 pathlib.Path('/tmp/fa-'+side+'.work').write_text('go 1.27\n\ntoolchain go1.27.1\n\nuse (\n'+root+'\n/tmp/fa-report/cohere/TypeScript/tsc\n)\n\nreplace (\n'+''.join(m+' => '+p+'\n' for m,p in replacements.items())+')\n')
rc=subprocess.run([sys.executable,str(out/'build.py')]).returncode
print('build child_exit='+str(rc),flush=True);assert rc==0
rc=subprocess.run([sys.executable,str(out/'control.py')],stdout=open('/tmp/fa-control.log','w'),stderr=subprocess.STDOUT).returncode
print('control child_exit='+str(rc),flush=True)
if rc:(out/'control-failed.json').write_text(json.dumps({'exit':rc}))
raise SystemExit(rc)
