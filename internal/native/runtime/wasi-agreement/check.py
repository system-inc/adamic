import pathlib, subprocess, shutil, os, json, sys
repo=pathlib.Path(__file__).resolve().parents[4]; root=pathlib.Path(os.environ.get('ADAMIC_AGREEMENT_DIRECTORY', '/tmp/wasi-agreement-evidence'));root.mkdir(exist_ok=True)
clang=str(pathlib.Path(os.environ['WASI_SYSROOT']).parents[1]/'bin/clang')
flags=['--target=wasm32-wasi','--sysroot='+os.environ['WASI_SYSROOT'],'-DADAMIC_TARGET_WASI=1','-std=c11','-Wall','-Wextra','-Werror','-pedantic','-O2','-ffp-contract=off','-fno-optimize-sibling-calls']
fixtures=['closures_throw','stack_forever','stack_tail_call','write_stdout_order','write_stderr_order']
def run(args):
 r=subprocess.run(args,cwd=repo,capture_output=True,timeout=180)
 if r.returncode: raise RuntimeError(str(args)+'\n'+r.stderr.decode())
 return r.stdout
def observe(args,combined=False):
 r=subprocess.run(args,cwd=repo,stdout=subprocess.PIPE,stderr=subprocess.STDOUT if combined else subprocess.PIPE,timeout=30)
 return dict(exit=r.returncode,stdout=r.stdout.decode(),stderr='' if combined else r.stderr.decode())
node=['node','--disable-warning=ExperimentalWarning']
compiler=root/'adamic';run(['go','build','-o',str(compiler),'./cmd/adamic'])
rt=root/'runtime';shutil.copytree(repo/'internal/native/runtime',rt,dirs_exist_ok=True)
objects=root/'objects';objects.mkdir(exist_ok=True)
for p in sorted(rt.glob('*.c')):
 print('compile',p.name,flush=True);run([clang,*flags,'-I',str(rt),'-c',str(p),'-o',str(objects/(p.name+'.o'))])
def build(name):
 src=root/(name+'.c');src.write_bytes(run([str(compiler),'c','internal/oracle/testdata/'+name+'.a']))
 out=root/(name+'.wasm');run([clang,*flags,'-Wno-unused-variable','-Wno-unused-but-set-variable','-Wno-unused-function','-Wno-unused-parameter','-Wno-self-assign','-I',str(rt),'-Wl,-z,stack-size=131072',str(src),*[str(p) for p in sorted(objects.glob('*.o'))],'-lm','-o',str(out)])
 return out
observations={}
for name in fixtures:
 source=str(repo/'internal/oracle/testdata'/(name+'.a')); wasm=build(name)
 expected=observe([*node,str(repo/'oracle/node.mjs'),source]);actual=observe([*node,str(repo/'oracle/wasi.mjs'),str(wasm)])
 observations[name]={'node':expected,'fixed':actual}
 print(name,json.dumps(observations[name]),flush=True)
 assert expected==actual,name
 if name.startswith('write_'):
  expected=observe([*node,str(repo/'oracle/node.mjs'),source],True);actual=observe([*node,str(repo/'oracle/wasi.mjs'),str(wasm)],True)
  observations[name]['combined']={'node':expected,'fixed':actual};assert expected==actual,name+' combined'
(root/'observations.json').write_text(json.dumps(observations,indent=2))
# Restore each old mechanism independently, compile it successfully, and require the
# independent source oracle to detect the runtime change rather than a compiler warning.
for file,names in [('sort.c',['closures_throw']),('adamic.h',['stack_forever','stack_tail_call']),('input.c',['write_stdout_order','write_stderr_order'])]:
 saved=(rt/file).read_bytes();(rt/file).write_bytes(run(['git','show','68446568595eaf38af9035d22793bab74729951e:internal/native/runtime/'+file]))
 if file.endswith('.c'):run([clang,*flags,'-I',str(rt),'-c',str(rt/file),'-o',str(objects/(file+'.o'))])
 for name in names:
  wasm=build(name);actual=observe([*node,str(repo/'oracle/wasi.mjs'),str(wasm)])
  observations[name]['mutant']=actual;print('mutant',file,name,json.dumps(actual),flush=True)
  assert actual!=observations[name]['node'],'SURVIVED '+file
 (rt/file).write_bytes(saved)
 if file.endswith('.c'):run([clang,*flags,'-I',str(rt),'-c',str(rt/file),'-o',str(objects/(file+'.o'))])
(root/'observations.json').write_text(json.dumps(observations,indent=2))
print('PASS: five separate-stream comparisons, two combined-stream comparisons, three independent mutants (five witnesses)',flush=True)
