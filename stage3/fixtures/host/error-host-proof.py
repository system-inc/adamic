import json, subprocess, pathlib, sys
# Run after the re-extraction, with Node 24.19.0 and pinned NODE_PATH.
assert len(sys.argv) in (4, 5), 'usage: python3 error-host-proof.py <repo> <adapted-tree> <logs> [before-observations.json]'
root=pathlib.Path(sys.argv[1]).resolve()
sys.path.insert(0, str(root/'internal/hostconfiguration'))
sys.dont_write_bytecode = True
from configuration import node_for
tree=str(pathlib.Path(sys.argv[2]).resolve())
bucket=root/'stage3/fixtures/host'
out=pathlib.Path(sys.argv[3]).resolve(); out.mkdir(parents=True,exist_ok=True)
rows=json.loads((bucket/'status.json').read_text())
before=json.loads(pathlib.Path(sys.argv[4]).read_text()) if len(sys.argv)==5 else {row['file']:row['node'] for row in rows}
observed={}
for index,row in enumerate(rows):
 binary_node=node_for(row,index+1)
 if binary_node is None:
  observed[row['file']]={'skipped':True,'configuration':row['configuration']}
  continue
 p=subprocess.run([binary_node,'--disable-warning=ExperimentalWarning','oracle/node.mjs',str(bucket/row['file'])],cwd=root,capture_output=True)
 result={'stdout':p.stdout.decode(),'stderr':p.stderr.decode(),'exit':p.returncode}
 (out/(row['file']+'.stdout')).write_bytes(p.stdout)
 (out/(row['file']+'.stderr')).write_bytes(p.stderr)
 assert result==before[row['file']], row['file']+' changed Node observation'
 assert result==row['node'],row['file']+' differs from recorded golden'
 observed[row['file']]=result
(out/'observations.json').write_text(json.dumps(observed,indent=2)+'\n')
fixture=bucket/'25_readDirectory.a'
fixture_index=next(index for index,row in enumerate(rows) if row['file']==fixture.name)
fixture_node=node_for(rows[fixture_index],fixture_index+1)
if fixture_node is None:
 (out/'proof.json').write_text(json.dumps({'status':'skipped','file':fixture.name,'configuration':rows[fixture_index]['configuration'],'observations':observed,'mutantsSkipped':True},indent=2)+'\n')
 sys.exit(0)
original=fixture.read_bytes()
mutants=[]
audit=['node',str(bucket/'source-audit.cjs'),tree]
try:
 for label,old,new in [('condition-cast','if (Error.captureStackTrace)','if ((Error as any).captureStackTrace)'),('call-cast','Error.captureStackTrace(e,','(Error as any).captureStackTrace(e,')]:
  text=original.decode(); assert text.count(old)==1
  fixture.write_text(text.replace(old,new))
  p=subprocess.run(audit,cwd=root,capture_output=True)
  (out/(label+'.log')).write_bytes(p.stdout+p.stderr)
  assert p.returncode!=0 and b'missing or changed copied declaration' in p.stderr,label+' survived audit'
  mutants.append({'mutant':label,'caughtBy':'full source audit','exit':p.returncode})
  fixture.write_bytes(original)
 text=original.decode(); old='if (extensions && !fileExtensionIsOneOf(name, extensions)) continue;'; assert text.count(old)==1
 # Keep relative module imports resolving from the actual fixture directory.
 fixture.write_text(text.replace(old,'if (false) continue;'))
 p=subprocess.run([fixture_node,'--disable-warning=ExperimentalWarning','oracle/node.mjs',str(fixture)],cwd=root,capture_output=True)
 (out/'extension-filter.stdout').write_bytes(p.stdout); (out/'extension-filter.stderr').write_bytes(p.stderr)
 assert p.returncode==0 and not p.stderr and p.stdout.decode()!=before[fixture.name]['stdout'],'extension semantic mutant survived'
 mutants.append({'mutant':'remove extension filter','caughtBy':'exact Node stdout comparison','exit':p.returncode})
finally:
 fixture.write_bytes(original)
p=subprocess.run(audit,cwd=root,capture_output=True)
(out/'restored-audit.log').write_bytes(p.stdout+p.stderr); assert p.returncode==0
proof={'node':subprocess.check_output(['node','--version']).decode().strip(),'fixtures':len(rows),'unchangedBeforeAndGoldens':True,'mutants':mutants,'restoredAuditExit':p.returncode}
skipped=[name for name,result in observed.items() if result.get('skipped')]
if skipped:
 proof.update(skipped=skipped,compared=len(rows)-len(skipped),unchangedBeforeAndGoldens=False)
(out/'proof.json').write_text(json.dumps(proof,indent=2)+'\n')
print('PASS: observed Node rows unchanged; skipped configured rows: '+', '.join(skipped) if skipped else 'PASS: all 25 Node observations unchanged; 2 stale-cast audit mutants and extension-filter semantic mutant caught; restored full audit passes')
