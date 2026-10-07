import importlib.util,os,pathlib,subprocess,tempfile
repo=pathlib.Path('/workspace/adamic');reports=repo/'cloud/reports/node-pin';fixture=sorted(pathlib.Path('/tmp/adamic-gate').glob('node-path-proof-*/repository'))[-1]
original=(repo/'cloud/setup.sh').read_text().replace('cloudSource="$repository/cloud"','cloudSource="/workspace/adamic/cloud"')
# Simulate drift after successful preparation, so only the final check can catch it.
original=original.replace('source "$tools/env.sh"','source "$tools/env.sh"\nln -sf /tmp/adamic-gate/node-drift-tools/bin/node "$tools/bin/node"')
condition='[ "$finalNodeVersion" = "$expectedNodeVersion" ] || { echo "setup: node got $finalNodeVersion, want $expectedNodeVersion" >&2; exit 1; }'
assert condition in original
env=dict(os.environ,ADAMIC_TOOLS='/tmp/adamic-gate/node-pin-tools',ADAMIC_SETUP_REPOSITORY=str(fixture),GOCACHE='/home/agent/.cache/go-build',GOMODCACHE='/tmp/adamic-gate/setup-modules-proof/with',GONOPROXY='github.com/klauspost/compress')
with tempfile.TemporaryDirectory(prefix='node-final-check-',dir=repo/'cloud') as temporary:
 for name,text in [('final-drift-refused',original),('missing-final-check',original.replace(condition,':'))]:
  script=pathlib.Path(temporary)/'setup.sh';script.write_text(text.replace('repository=$(cd -- "$scriptDirectory/.." && pwd)','repository=/workspace/adamic'))
  with (reports/(name+'.log')).open('wb') as log:result=subprocess.run(['bash',str(script)],env=env,stdout=log,stderr=subprocess.STDOUT,timeout=120)
  log=(reports/(name+'.log')).read_text()
  if name=='final-drift-refused':assert result.returncode!=0 and 'got v24.21.0, want v24.19.0' in log,log
  else:
   assert result.returncode==0 and log.splitlines()[-1]=='setup: node v24.21.0',log
   try:assert result.returncode!=0,'setup must reject post-prepare Node drift'
   except AssertionError as failure:print('missing-final-check mutant caught:',failure)
   else:raise AssertionError('mutant survived')
spec=importlib.util.spec_from_file_location('node',repo/'cloud/setup-node.py');node=importlib.util.module_from_spec(spec);spec.loader.exec_module(node);print(node.prepare('/tmp/adamic-gate/node-pin-tools'))
print('PASS: final exact-version refusal fails only when the check is removed; restored pinned Node')
