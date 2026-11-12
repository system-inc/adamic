import pathlib,subprocess,sys,json,hashlib
out=pathlib.Path(__file__).resolve().parent;pins=json.loads((out/'pins.json').read_text());root=pathlib.Path('/workspace/orm-main');build=pathlib.Path('/workspace/orm-build-main')
def git(*args):
 p=subprocess.run(['git',*args],cwd=root,capture_output=True,text=True);print('git child_exit='+str(p.returncode))
 if p.returncode:sys.exit(p.returncode)
 return p.stdout.strip()
def sha(p):return hashlib.sha256(pathlib.Path(p).read_bytes()).hexdigest()
assert git('rev-parse','HEAD')==pins['main'] and not git('status','--porcelain','--untracked-files=no')
for pin in pins['landed'].values():
 p=subprocess.run(['git','merge-base','--is-ancestor',pin,pins['main']],cwd=root);print('ancestry child_exit='+str(p.returncode));p.check_returncode()
source=json.loads((root/'stage3/source.json').read_text());assert git('-C','/workspace/op2-upstream','rev-parse','HEAD')==source['commit']
entry=json.loads((build/'entry-overlay/overlay.json').read_text());scan=json.loads((build/'scan-overlay/overlay.json').read_text());assert all(k.startswith(str(root)+'/') for m in [entry,scan] for k in m['Replace'])
manifest={str(p.relative_to('/workspace/orm-adapted')):sha(p) for p in sorted(pathlib.Path('/workspace/orm-adapted/src').rglob('*')) if p.is_file()}
prior=json.loads((out.parent/'overload-phantom-78933e52/adapted-source-hashes.json').read_text());changed={name:{'previous':prior.get(name),'main':manifest.get(name)} for name in sorted(prior.keys()|manifest.keys()) if prior.get(name)!=manifest.get(name)}
(out/'adapted-source-hashes.json').write_text(json.dumps(manifest,indent=2)+'\n');(out/'input-difference-from-78933e52.json').write_text(json.dumps(changed,indent=2)+'\n');(out/'patch-set.md').write_bytes(pathlib.Path('/workspace/orm-adapted/patch-set.md').read_bytes())
data=dict(main=pins['main'],landed=pins['landed'],report_parent=pins['report_parent'],comparison=pins['comparison'],source=source,submodules=git('submodule','status','--recursive'),shared_adapted_tree='/workspace/orm-adapted',mode=dict(LATENT_FULL='0',LATENT_ASSERT_NO_OUTPUT='1',CENSUS_SHARDS='8',GOMEMLIMIT='3GiB',CENSUS_SKIP_SCAN='1 for shards 2-7; unset for 0-1'),binary_sha256=sha(build/'entry-census'),scan_binary_sha256=sha(build/'entry-census-scan'),overlay_inputs_from_own_tree=True,entry_overlay=entry,scan_overlay=scan,input_matches_prior=not changed,changed_input_files=len(changed))
(out/'provenance.json').write_text(json.dumps(data,indent=2)+'\n');print('main pin and landed fixes verified; source files',len(manifest),'changed input files',len(changed))
