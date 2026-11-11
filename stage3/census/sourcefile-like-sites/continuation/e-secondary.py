import pathlib,time,sys,json,gzip
sys.dont_write_bytecode=True
from variants import invoke
from measure import run
p=pathlib.Path(__file__).resolve().parent
while not (p/'e/source-diff.json').exists():time.sleep(2)
invoke(['/workspace/scratch/baobab-e-overloads','/workspace/scratch/baobab-road2-e/src/compiler',str(p.parent/'roots.json'),str(p/'e/native-overload-resolution.json')],p/'e/native-overload-resolution.log',600)
while not (p/'e/sdk-adamic-checker.json').exists():time.sleep(2)
scratch=pathlib.Path('/workspace/scratch/baobab-sdk-e');scratch.mkdir(exist_ok=True)
for name in ['typescript.d.ts','public-adamic-consumer.ts']:
 source=p/'e'/name;data=source.read_bytes() if source.exists() else gzip.decompress(pathlib.Path(str(source)+'.gz').read_bytes());(scratch/('consumer.ts' if name.endswith('consumer.ts') else name)).write_bytes(data)
run('sdk-e',scratch,['consumer.ts'],300,False,1)
print('e secondary workflow child exit=0',flush=True)
