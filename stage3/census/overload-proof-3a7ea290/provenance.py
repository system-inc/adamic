import pathlib,subprocess,sys,json,hashlib
out=pathlib.Path(__file__).resolve().parent

def git(*args):
 p=subprocess.run(['git',*args],cwd='/workspace/op2-report',capture_output=True,text=True);print('git child_exit='+str(p.returncode));
 if p.returncode:sys.exit(p.returncode)
 return p.stdout.strip()
def sha(p):return hashlib.sha256(pathlib.Path(p).read_bytes()).hexdigest()
base=git('rev-parse','17c6378d');head=git('rev-parse','3a7ea290');assert git('rev-parse',head+'^')==base
paths=['stage3/source.json','stage3/apply.py','stage3/apply.sh','stage3/api','stage3/adapt']
assert not git('diff','--name-only',base,head,'--',*paths)
assert not git('diff','--name-only',base,head,'--','stage3/census','stage3/meter')
assert git('-C','/workspace/op2-upstream','rev-parse','HEAD')=='050880ce59e30b356b686bd3144efe24f875ebc8'
data=dict(base=base,head=head,prior_census_commit=git('rev-parse','83e07d2f'),upstream=git('-C','/workspace/op2-upstream','rev-parse','HEAD'),adaptation_changed=False,adaptation_paths_checked=paths,shared_adapted_tree='/workspace/op2-adapted',mode=dict(LATENT_FULL='0',LATENT_ASSERT_NO_OUTPUT='1',CENSUS_SHARDS='8',GOMEMLIMIT='3GiB',CENSUS_SKIP_SCAN='1 for shards 2-7; unset for 0-1'),head_changed_paths=git('diff','--name-only',base,head).splitlines())
for label,pin in [('base',base),('head',head)]:
 root=pathlib.Path('/workspace/op2-'+label);build=pathlib.Path('/workspace/op2-build-'+label)
 assert git('-C',str(root),'rev-parse','HEAD')==pin
 assert not git('-C',str(root),'status','--porcelain','--untracked-files=no')
 meta=json.loads((build/'entry-overlay/overlay.json').read_text())
 assert all(k.startswith(str(root)+'/') for k in meta['Replace'])
 scanmeta=json.loads((build/'scan-overlay/overlay.json').read_text());assert all(k.startswith(str(root)+'/') for k in scanmeta['Replace'])
 data.setdefault('builds',{})[label]=dict(scan_binary_sha256=sha(build/'entry-census-scan'),scan_overlay=scanmeta,submodules=git('-C',str(root),'submodule','status','--recursive'),binary_sha256=sha(build/'entry-census'),overlay_inputs_from_own_tree=True,entry_overlay=meta)
data['head_proof_sha256']=sha('/workspace/op2-head/internal/lower/census_overload_proof.go')
manifest={str(p.relative_to('/workspace/op2-adapted')):sha(p) for p in sorted(pathlib.Path('/workspace/op2-adapted/src').rglob('*')) if p.is_file()}
(out/'adapted-source-hashes.json').write_text(json.dumps(manifest,indent=2)+'\n')
(out/'patch-set.md').write_bytes(pathlib.Path('/workspace/op2-adapted/patch-set.md').read_bytes())
(out/'provenance.json').write_text(json.dumps(data,indent=2)+'\n')
print('source_files',len(manifest))
