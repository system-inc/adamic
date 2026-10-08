import pathlib,subprocess,concurrent.futures,sys,hashlib
root=pathlib.Path(__file__).resolve().parents[3]; out=pathlib.Path('/tmp/empty-objects')/sys.argv[1];out.mkdir(parents=True,exist_ok=True)
flags=['-std=c11','-Wall','-Wextra','-Werror','-pedantic','-Wno-unused-variable','-Wno-unused-but-set-variable','-Wno-unused-function','-Wno-unused-parameter','-Wno-self-assign','-ffp-contract=off','-fno-optimize-sibling-calls','-O2']
def build(p):
 for counted in [False,True]:
  dest=out/(p.stem+('-count' if counted else '')+'.o')
  subprocess.run(['clang',*flags,*(['-DADAMIC_COUNT'] if counted else []),'-c',str(p),'-o',str(dest)],check=True)
with concurrent.futures.ThreadPoolExecutor(4) as pool:list(pool.map(build,sorted((root/'internal/native/runtime').glob('*.c'))))
print('objects',len(list(out.glob('*.o'))))
if sys.argv[1]=='after':
 before=out.parent/'before';diff=[p.name for p in out.glob('*.o') if p.read_bytes()!=(before/p.name).read_bytes()];print('differing',diff);assert not diff
