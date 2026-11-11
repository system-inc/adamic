import pathlib,sys,time,subprocess,json,shutil,hashlib
sys.dont_write_bytecode=True
from variants import invoke
p=pathlib.Path(__file__).resolve().parent;out=p.parent;repo=out.parents[2]
while 'c public census workflow child exit=0' not in pathlib.Path('/tmp/baobab-c-public-census.log').read_text():time.sleep(2)
for i,command in enumerate([
 ['python3',str(p/'compare-c-sites.py')],['python3',str(p/'analyze-c.py')],
 ['python3',str(p/'analyze-sdk-c.py')],['python3',str(p/'report-sdk-c.py')],['python3',str(p/'report-c.py')],
 ['python3',str(out/'report.py')],['python3',str(out/'seal-inputs.py')],
 ['python3',str(out/'audit.py')],['python3',str(p/'audit-followup.py')],['python3',str(p/'audit-c.py')]]):
 invoke(command,p/'preparation'/('c-final-'+str(i)+'.log'),600)
for source in sorted(pathlib.Path('/tmp').glob('baobab-c-*.log')):
 shutil.copyfile(source,p/'preparation'/source.name)
metadata={'previous_report_sha':'7944b20a23c7e2e40c00e925603c20bd7267a569','main_sha':'d690f276749f91e6b36fe3a9726a5a453fcd7d39','typescript_sha':'050880ce59e30b356b686bd3144efe24f875ebc8','guarded_binary_sha256':hashlib.sha256(pathlib.Path('/workspace/scratch/baobab-latent').read_bytes()).hexdigest(),'cohere_binary_sha256':hashlib.sha256(pathlib.Path('/workspace/scratch/baobab-cohere').read_bytes()).hexdigest(),'production_checker_driver_sha256':hashlib.sha256(pathlib.Path('/workspace/scratch/baobab-production-checker').read_bytes()).hexdigest(),'per_file_bound_seconds':1500,'max_concurrent_file_censuses':2,'public_variant_after_internal_attempts':True,'adamic_whole_src_driver':'production load.Load, exact 604 stock roots, no lowering or output IR','sdk_boundary':'new no-optional-widening refusal, including readonly control','audit':json.loads((p/'c-audit-result.json').read_text())}
(p/'c-provenance.json').write_text(json.dumps(metadata,indent=2)+'\n')
pathlib.Path('/workspace/scratch/baobab-c-ready.json').write_text(json.dumps({'ready':True,'audit':metadata['audit']},indent=2)+'\n')
print('c final preparation child exit=0',flush=True)
