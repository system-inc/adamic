import pathlib,subprocess,sys,json,hashlib,re
out=pathlib.Path(__file__).resolve().parent
c=json.loads((out/'composition.json').read_text())
def git(*args):
 p=subprocess.run(['git',*args],cwd='/workspace/op2-report',capture_output=True,text=True);print('git child_exit='+str(p.returncode))
 if p.returncode:sys.exit(p.returncode)
 return p.stdout.strip()
def sha(p):return hashlib.sha256(pathlib.Path(p).read_bytes()).hexdigest()
base,head=c['base'],c['head'];paths=['stage3/source.json','stage3/apply.py','stage3/apply.sh','stage3/api','stage3/adapt']
for name,pin in [('step2b',c['step2b']),('step3',c['step3'])]:
 assert git('rev-parse',pin+'^')==c['step2']
 assert not git('diff','--name-only',pin+'^',pin,'--',*paths)
assert not git('diff','--name-only',base,head,'--',*paths)
assert not git('diff','--name-only',base,head,'--','stage3/census','stage3/meter')
assert c['merges']['base_path']['parents']==[c['main'],c['path']]
assert c['merges']['base_step2']['parents']==[c['merges']['base_path']['commit'],c['step2']]
assert c['merges']['head_step2b']['parents']==[base,c['step2b']]
assert c['merges']['head_step3']['parents']==[c['merges']['head_step2b']['commit'],c['step3']]
for name,m in c['merges'].items():
 assert git('show','-s','--format=%P',m['commit']).split()==m['parents']
 assert git('rev-parse',m['commit']+'^{tree}')==m['tree'] and not m['conflicts']
log_names=['base-merge-path','base-merge-step2','head-merge-step2b','head-merge-step3']
for name in log_names:
 d=json.loads(pathlib.Path('/tmp/oph-'+name+'.time.json').read_text());assert d['exit_code']==0 and not d['timed_out']
upstream=git('-C','/workspace/op2-upstream','rev-parse','HEAD');assert upstream=='050880ce59e30b356b686bd3144efe24f875ebc8'
data=dict(base=base,head=head,composition=c,upstream=upstream,adaptation_changed_at_head=False,adaptation_paths_checked=paths,shared_adapted_tree='/workspace/oph-adapted',mode=dict(LATENT_FULL='0',LATENT_ASSERT_NO_OUTPUT='1',CENSUS_SHARDS='8',GOMEMLIMIT='3GiB',CENSUS_SKIP_SCAN='1 for shards 2-7; unset for 0-1'),head_changed_paths=git('diff','--name-only',base,head).splitlines(),builds={})
for label,pin in [('base',base),('head',head)]:
 root=pathlib.Path('/workspace/oph-'+label);build=pathlib.Path('/workspace/oph-build-'+label)
 assert git('-C',str(root),'rev-parse','HEAD')==pin
 assert not git('-C',str(root),'status','--porcelain','--untracked-files=no')
 entry=json.loads((build/'entry-overlay/overlay.json').read_text());scan=json.loads((build/'scan-overlay/overlay.json').read_text())
 assert all(k.startswith(str(root)+'/') for m in [entry,scan] for k in m['Replace'])
 data['builds'][label]=dict(submodules=git('-C',str(root),'submodule','status','--recursive'),binary_sha256=sha(build/'entry-census'),scan_binary_sha256=sha(build/'entry-census-scan'),overlay_inputs_from_own_tree=True,entry_overlay=entry,scan_overlay=scan)
data['go_evidence_sha256']={p.name:sha(p) for p in out.glob('*-head.go.txt')}
manifest={str(p.relative_to('/workspace/oph-adapted')):sha(p) for p in sorted(pathlib.Path('/workspace/oph-adapted/src').rglob('*')) if p.is_file()}
(out/'adapted-source-hashes.json').write_text(json.dumps(manifest,indent=2)+'\n')
(out/'patch-set.md').write_bytes(pathlib.Path('/workspace/oph-adapted/patch-set.md').read_bytes())
p=pathlib.Path('/workspace/oph-adapted/src/compiler/path.ts');text=p.read_text();matches=list(re.finditer(r'export\s+interface\s+PathPathComponents\b[^\{]*\{[^}]*\}',text));assert len(matches)==1
fragment=matches[0].group();normalized=' '.join(fragment.split());assert 'T extends Path = Path' in normalized and 'extends Array<T>' in normalized and '__pathComponensBrand' in normalized
(out/'path-components-interface.json').write_text(json.dumps(dict(source='src/compiler/path.ts',source_sha256=sha(p),line=text[:matches[0].start()].count('\n')+1,text=fragment),indent=2)+'\n')
(out/'provenance.json').write_text(json.dumps(data,indent=2)+'\n');print('source files',len(manifest),'both scratch compositions and PathPathComponents interface verified')
