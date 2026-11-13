import datetime, hashlib, json, pathlib, subprocess
root=pathlib.Path('/tmp/census-68x1t2y')
def git(tree,*args):return subprocess.check_output(['git',*args],cwd=tree,text=True).strip()
def sha(path):return hashlib.sha256(path.read_bytes()).hexdigest()
meta=dict(task='68x1t2y',measurement='measured on a checker-rejected program',captured_utc=datetime.datetime.now(datetime.timezone.utc).isoformat(),ae342094=git(root/'tip','rev-parse','ae342094'),note='ae342094 adds a regression test; its parent already contains the rigid typeMapper reset. Exact requested endpoints retained.',process_limit_seconds=1700,configurations={})
for name in ('parent','tip'):
 tree=root/name; adapted=root/('adapted-'+name); out=root/'results'/name
 config=dict(commit=git(tree,'rev-parse','HEAD'),compiler_checkout=str(tree),adapted=str(adapted),source_directory=str(adapted/'src/compiler'),selected_file=str(adapted/'src/compiler/transformers/destructuring.ts'),cohere=git(tree,'rev-parse','HEAD:cohere'),cohere_actual=git(tree/'cohere','rev-parse','HEAD'),typescript_go=git(tree/'cohere/TypeScript','rev-parse','HEAD'),source_pin=json.loads((tree/'stage3/source.json').read_text()),environment=dict(LATENT_FULL='1',LATENT_ONLY_FILE=str(adapted/'src/compiler/transformers/destructuring.ts'),LATENT_ASSERT_NO_OUTPUT='1',LATENT_TIMES=str(out/'times.jsonl')),build_exit_code=0,overlay_exit_code=0,apply_exit_code=0,adapted_product_key=subprocess.check_output(['stage3/apply.sh','--key'],cwd=tree,text=True).strip(),binary_sha256=sha(root/('latent-'+name)),overlay_source_sha256={p.name:sha(p) for p in sorted((root/('overlay-'+name)).glob('*.go'))})
 assert config['cohere']==config['cohere_actual']
 (out/'source-manifest.json').write_text(json.dumps({p.relative_to(adapted).as_posix():sha(p) for p in sorted((adapted/'src').rglob('*')) if p.is_file()},indent=2)+'\n')
 for file in ('patch-set.md',): (out/file).write_bytes((adapted/file).read_bytes())
 (out/'destructuring.ts').write_bytes((adapted/'src/compiler/transformers/destructuring.ts').read_bytes())
 (out/'debug.ts').write_bytes((adapted/'src/compiler/debug.ts').read_bytes())
 meta['configurations'][name]=config
(root/'results/provenance.json').write_text(json.dumps(meta,indent=2)+'\n')
print('provenance and final source manifests captured')
