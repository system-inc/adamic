"""Owned continuation validation; does not edit the shared harness or dispatcher."""
import argparse, gzip, hashlib, json, os, pathlib, re, shutil, subprocess, time

REPO=pathlib.Path(__file__).resolve().parents[4]
OWN=pathlib.Path(__file__).resolve().parent
NAMES=['no-new-func','no-new-native-nonconstructor','no-new-wrappers']

def main():
 p=argparse.ArgumentParser();p.add_argument('--scratch',required=True);p.add_argument('--compiler');p.add_argument('--initial',action='store_true');args=p.parse_args()
 scratch=pathlib.Path(args.scratch).resolve();scratch.mkdir(parents=True,exist_ok=True);records=[]
 def run(name,command,env=None,expect=0):
  start=time.monotonic()
  with (scratch/(name+'.stdout')).open('wb') as out,(scratch/(name+'.stderr')).open('wb') as err:
   result=subprocess.run(command,cwd=REPO,stdout=out,stderr=err,env=env)
  record={'name':name,'command':[str(x) for x in command],'exit':result.returncode,'seconds':time.monotonic()-start};records.append(record)
  (scratch/'runs.json').write_text(json.dumps(records,indent=2)+'\n')
  if result.returncode!=expect:raise RuntimeError(str(record)+'\n'+(scratch/(name+'.stderr')).read_text())
  return (scratch/(name+'.stdout')).read_bytes()
 def overlay(name,replacements):
  path=scratch/(name+'.json');path.write_text(json.dumps({'Replace':replacements}));return path
 archive=scratch/'checker.a';asan_archive=scratch/'checker-asan.a';stage0=scratch/'adamic';native=scratch/'native';asan=scratch/'native-asan';oracle=scratch/'oracle'
 run('stage0-build',['go','build','-o',stage0,'./cmd/adamic'])
 run('archive-build',['go','build','-buildmode=c-archive','-o',archive,'./bridge/tsgo/archive'])
 run('native-build',[stage0,'build',OWN/'suite.a','-o',native,'--tsgo',archive])
 virtual=REPO/'cohere/adamic_wave06nextoracle.go'
 go_overlay=overlay('oracle-overlay',{str(virtual):str(OWN/'testdata/oracle.go')})
 run('oracle-build',['go','-C',REPO/'cohere','build','-overlay',go_overlay,'-o',oracle,virtual])
 anchor=scratch/'anchor.d.ts';anchor.write_text('export {};')
 config=scratch/'tsconfig.json';config.write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ES2022','module':'ESNext','moduleResolution':'bundler','allowArbitraryExtensions':True,'lib':['ES2022','DOM']},'files':[str(anchor)]}))
 helper=scratch/'helpers.a';helper.write_text('export class Function{};export class Symbol{};export class BigInt{};export class String{};export class Number{};export class Boolean{};')
 examples=['new Function("return 1")', 'Function("return 1")', '(Function)("return 1")', 'new ((Function))("return 1")', 'Function.call(null,"return 1")', 'Function.apply(null,["return 1"])', 'Function.bind(null,"return 1")()', 'Function["call"](null,"return 1")', 'Function[`bind`](null,"return 1")', '(Function?.call)(null,"return 1")', 'Function?.("return 1")', 'Function[("call")](null,"return 1")', 'Function.toString()', 'Function.bind', 'foo.Function.call(null)', '(foo,Function).call(null)', 'Function.Function.call(null)', 'foo().Function.call(null)', 'Function[method](null)', 'Function["ca"+"ll"](null)', 'new Function;', 'class Function{};new Function()', 'function f(Function:any){Function("x")}', 'function f(){class Function{}};new Function("x")', 'const alias=Function;new alias("x")', 'function Function(){};Function("x")', 'const Function=foo;Function.call(null)', 'import {Function} from "./helpers.a";new Function("x")', 'new Symbol(1)', 'new (Symbol)(1)', 'new ((Symbol))(1)', 'Symbol(1)', 'new Symbol', 'function f(Symbol:any){return new Symbol(1)}', 'class Symbol{};new Symbol(1)', 'function f(){class Symbol{}};new Symbol(1)', 'new globalThis.Symbol(1)', 'const Symbol=foo;new Symbol(1)', 'import {Symbol} from "./helpers.a";new Symbol(1)', '{const Symbol=foo;new Symbol(1)}new Symbol(1)', 'if(foo){new Symbol(1)}else{var Symbol=foo}', 'new foo(Symbol)', 'new BigInt(1)', 'new (BigInt)(1)', 'new ((BigInt))(1)', 'BigInt(1)', 'new BigInt', 'function f(BigInt:any){return new BigInt(1)}', 'class BigInt{};new BigInt(1)', 'function f(){class BigInt{}};new BigInt(1)', 'new globalThis.BigInt(1)', 'const BigInt=foo;new BigInt(1)', 'import {BigInt} from "./helpers.a";new BigInt(1)', '{const BigInt=foo;new BigInt(1)}new BigInt(1)', 'if(foo){new BigInt(1)}else{var BigInt=foo}', 'new foo(BigInt)', 'new String(1)', 'new (String)(1)', 'new ((String))(1)', 'String(1)', 'new String', 'function f(String:any){return new String(1)}', 'class String{};new String(1)', 'function f(){class String{}};new String(1)', 'new globalThis.String(1)', 'const String=foo;new String(1)', 'import {String} from "./helpers.a";new String(1)', '{const String=foo;new String(1)}new String(1)', 'if(foo){new String(1)}else{var String=foo}', 'new foo(String)', 'new Number(1)', 'new (Number)(1)', 'new ((Number))(1)', 'Number(1)', 'new Number', 'function f(Number:any){return new Number(1)}', 'class Number{};new Number(1)', 'function f(){class Number{}};new Number(1)', 'new globalThis.Number(1)', 'const Number=foo;new Number(1)', 'import {Number} from "./helpers.a";new Number(1)', '{const Number=foo;new Number(1)}new Number(1)', 'if(foo){new Number(1)}else{var Number=foo}', 'new foo(Number)', 'new Boolean(1)', 'new (Boolean)(1)', 'new ((Boolean))(1)', 'Boolean(1)', 'new Boolean', 'function f(Boolean:any){return new Boolean(1)}', 'class Boolean{};new Boolean(1)', 'function f(){class Boolean{}};new Boolean(1)', 'new globalThis.Boolean(1)', 'const Boolean=foo;new Boolean(1)', 'import {Boolean} from "./helpers.a";new Boolean(1)', '{const Boolean=foo;new Boolean(1)}new Boolean(1)', 'if(foo){new Boolean(1)}else{var Boolean=foo}', 'new foo(Boolean)', '/* global Boolean:off */ new Boolean(false)', 'new Number(1);new Boolean(0);new String("x");', '// π😀\r\nnew String("é")\r\nnew Symbol("μ");', '(Function!)("return 1")', '(Function as any)("return 1")', 'Function[("call" as const)](null,"return 1")']
 examples += ["Function?.['call'](null,'x')", "Function?.[(\"call\")](null,'x')", r"new \u0053tring('x')", r"\u0046unction('x')", r"Function['c\x61ll'](null,'x')"]
 controls=[]
 for i,example in enumerate(examples):
  file=scratch/('control-'+str(i)+'.a');file.write_text('declare const foo:any;declare const method:string;'+example+';\nexport {};');controls.append(file)
 manifest=scratch/'controls.manifest';manifest.write_text(''.join(str(x)+'\n' for x in controls))
 def compare(name,binary,configuration,inputs,timed=False):
  env=os.environ.copy()
  if timed:env['ADAMIC_TSGO_TIMING']='1'
  native_output=run(name+'-native',[binary,configuration,inputs],env)
  go_output=run(name+'-go',[oracle,configuration,inputs])
  if native_output!=go_output:
   a=native_output.decode().splitlines();b=go_output.decode().splitlines()
   import difflib
   difference='\n'.join(difflib.unified_diff(b,a,fromfile='Go',tofile='native'))
   (scratch/(name+'-diff.txt')).write_text(difference)
   raise AssertionError(difference)
  print(name, len(native_output),'identical bytes',native_output.decode().splitlines()[-1],flush=True)
  return go_output
 truth=compare('controls',native,config,manifest)
 for name in NAMES:assert name.encode() in truth,name+' missing positive control'
 if args.initial:return
 env=os.environ.copy();env.update(CC='clang',CGO_CFLAGS='-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all')
 run('archive-asan-build',['go','build','-buildmode=c-archive','-o',asan_archive,'./bridge/tsgo/archive'],env)
 run('native-asan-build',[stage0,'build',OWN/'suite.a','-o',asan,'--tsgo',asan_archive,'--sanitize'])
 compare('controls-asan',asan,config,manifest)
 assert (scratch/'controls-asan-native.stderr').read_bytes()==b''
 for folder,old,new in [('no_new_func','!g.global(target)','false'),('no_new_native_nonconstructor',',callee);',',at);'),('no_new_wrappers','g.skip(g.child(at))','g.child(at)')]:
  directory=scratch/(folder+'-source');shutil.copytree(OWN,directory,dirs_exist_ok=True)
  for file in directory.rglob('*.a'):
   original=OWN/file.relative_to(directory)
   text=file.read_text();text=re.sub(r"(['\"])([^'\"]+\.(?:ts|a))\1",lambda m:repr(m.group(2) if m.group(2).endswith('.a') and (original.parent/m.group(2)).resolve().is_relative_to(OWN) else str((original.parent/m.group(2)).resolve())),text)
   if file==directory/folder/'index.a':assert text.count(old)==1;text=text.replace(old,new)
   file.write_text(text)
  binary=scratch/(folder+'-mutant');run(folder+'-mutant-build',[stage0,'build',directory/'suite.a','-o',binary,'--tsgo',archive])
  output=run(folder+'-mutant-run',[binary,config,manifest]);assert output!=truth and (scratch/(folder+'-mutant-run.stderr')).read_bytes()==b''
  print(folder,'semantic mutant exits 0, Go byte comparison catches it',flush=True)
 # The new question follows the archive's existing released-handle boundary.
 probe=scratch/'released.a';probe.write_text("import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';const a=programArguments();const p=tsgoProgram(a[0]??'',[a[1]??'']);tsgoRelease(p);console.log(tsgoInspect(p,a[1]??'',0,1,'CallExpression','wave06-declarations\\nsymbol'));")
 released=scratch/'released';run('released-build',[stage0,'build',probe,'-o',released,'--tsgo',archive]);run('released-run',[released,config,controls[0]],expect=70)
 assert (scratch/'released-run.stderr').read_text()=='adamic: panic: invalid or released checker handle\n'
 # Mutation of the Go fact projection, never of the oracle.
 helper=REPO/'bridge/tsgo/checker/wave06_declarations.go';source=helper.read_text();needle='out.yes(source.IsDeclarationFile)';assert source.count(needle)==1
 mutated=scratch/'question.go';mutated.write_text(source.replace(needle,'out.yes(!source.IsDeclarationFile)'))
 fact_overlay=overlay('question-mutant-overlay',{str(helper):str(mutated)})
 fact_archive=scratch/'question-mutant.a';run('question-mutant-archive',['go','build','-overlay',fact_overlay,'-buildmode=c-archive','-o',fact_archive,'./bridge/tsgo/archive'])
 fact_binary=scratch/'question-mutant';run('question-mutant-build',[stage0,'build',OWN/'suite.a','-o',fact_binary,'--tsgo',fact_archive]);output=run('question-mutant-run',[fact_binary,config,manifest]);assert output!=truth and (scratch/'question-mutant-run.stderr').read_bytes()==b''
 print('fact mutant exits 0, Go finding bytes catch the wrong declaration-file flag; released handle panic 70 verified',flush=True)
 if args.compiler:
  compiler=pathlib.Path(args.compiler).resolve();pin=subprocess.check_output(['git','-C',compiler,'rev-parse','HEAD'],text=True).strip();assert pin=='050880ce59e30b356b686bd3144efe24f875ebc8'
  for name,base,configuration in [('compiler',compiler,compiler/'src/compiler/tsconfig.json'),('repository',REPO,REPO/'tsconfig.json')]:
   relative=(OWN.parent/'validation-coverage'/(name+'.manifest')).read_text().splitlines();inputs=scratch/(name+'.manifest');inputs.write_text(''.join(str(base/x)+'\n' for x in relative if x))
   compare(name,native,configuration,inputs);compare(name+'-asan',asan,configuration,inputs);assert (scratch/(name+'-asan-native.stderr')).read_bytes()==b''
   compare(name+'-timed',native,configuration,inputs,True)
 print('PASS',flush=True)

if __name__=='__main__':main()
