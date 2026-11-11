import pathlib,time,sys,json,hashlib,shutil
sys.dont_write_bytecode=True
from variants import invoke
p=pathlib.Path(__file__).resolve().parent;out=p.parent
for log,marker in [('/tmp/baobab-e-corrected-controller.log','e census workflow child exit=0'),('/tmp/baobab-e-corrected-aux.log','e auxiliary workflow child exit=0'),('/tmp/baobab-e-secondary.log','e secondary workflow child exit=0')]:
 while marker not in pathlib.Path(log).read_text():time.sleep(2)
for i,cmd in enumerate([
 ['python3',str(p/'analyze-e.py')],['python3',str(p/'report-e.py')],['python3',str(out/'report.py')],['python3',str(out/'seal-inputs.py')],
 ['python3',str(out/'audit.py')],['python3',str(p/'audit-followup.py')],['python3',str(p/'audit-c.py')],['python3',str(p/'audit-e.py')]]):invoke(cmd,p/'preparation'/('e-final-'+str(i)+'.log'),600)
for log in pathlib.Path('/tmp').glob('baobab-e-*.log'):shutil.copyfile(log,p/'preparation'/log.name)
metadata={'previous_report_sha':'e87b5ec71092a42bf931b8f0f0a6b8840d3718b6','main_sha':'d690f276749f91e6b36fe3a9726a5a453fcd7d39','typescript_sha':'050880ce59e30b356b686bd3144efe24f875ebc8','guarded_binary_sha256':hashlib.sha256(pathlib.Path('/workspace/scratch/baobab-latent').read_bytes()).hexdigest(),'cohere_binary_sha256':hashlib.sha256(pathlib.Path('/workspace/scratch/baobab-cohere').read_bytes()).hexdigest(),'native_overload_probe_sha256':hashlib.sha256(pathlib.Path('/workspace/scratch/baobab-e-overloads').read_bytes()).hexdigest(),'per_file_bound_seconds':1500,'max_concurrent_file_censuses':2,'SourceFileLike_unchanged_including_documentation':True,'source_tree':'/workspace/scratch/baobab-road2-e','abandoned_controls':['e-unsuppressed-control','e-displaced-doc-control'],'audit':json.loads((p/'e-audit-result.json').read_text())}
(p/'e-provenance.json').write_text(json.dumps(metadata,indent=2)+'\n')
pathlib.Path('/workspace/scratch/baobab-e-ready.json').write_text(json.dumps({'ready':True,'audit':metadata['audit']},indent=2)+'\n')
print('e final preparation child exit=0',flush=True)
