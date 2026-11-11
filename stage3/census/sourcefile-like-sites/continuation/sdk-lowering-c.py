import pathlib,sys,gzip,json
sys.dont_write_bytecode=True
from measure import run
p=pathlib.Path(__file__).resolve().parent
for label in ['original','c-public','c-public-unstripped','c-public-readonly-sdk-control']:
 target=p/label;scratch=pathlib.Path('/workspace/scratch/baobab-sdk-'+label);scratch.mkdir(exist_ok=True)
 for name in ['typescript.d.ts','public-adamic-consumer.ts']:
  source=target/name
  data=source.read_bytes() if source.exists() else gzip.decompress(pathlib.Path(str(source)+'.gz').read_bytes())
  (scratch/('consumer.ts' if name.endswith('consumer.ts') else name)).write_bytes(data)
 run('sdk-'+label,scratch,['consumer.ts'],300,False,1)
print('SDK lowering controller child exit=0',flush=True)
