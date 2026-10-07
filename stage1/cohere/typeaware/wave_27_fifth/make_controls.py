from pathlib import Path
import sys,json,re
repo=Path(__file__).resolve().parents[4];root=Path(sys.argv[1]);root.mkdir(parents=True,exist_ok=True)
texts=[]
for name in ['require_await','symbol_description','valid_typeof']:
 source=(repo/f'cohere/internal/lint/rules/core/{name}_test.go').read_text()
 # Drop Go comments before reading Go source string tokens.
 source=re.sub(r'^\s*//[^\n]*','',source,flags=re.M)
 tokens=re.findall(r'"(?:[^"\\]|\\.)*"|`[^`]*`',source)
 prelude=''
 if name=='require_await':prelude=re.search(r'prelude := `([^`]*)`',source).group(1)
 for token in tokens:
  try:text=token[1:-1] if token.startswith('`') else json.loads(token)
  except ValueError:continue
  if any(s in text for s in ['async ','Symbol(', 'Symbol =','Symbol {};','Symbol:', 'typeof ', 'new Symbol','(Symbol)','Symbol.','isNaN();','Number();','parseInt();','String();','Boolean();','Array();','Object();','Date();','await using','for await']):
   if 'useMemo(function' in text and 'function useMemo' not in text:text=prelude+text
   texts.append(text)
texts += [
 "function identity<T>(value:T):T{return value;} identity?.(async()=>1);",
 "const f=async()=>async()=>await work();",
 "const f=async()=>class C{async run(){await work();}};",
 "import {Symbol} from './ambient.a'; Symbol();",
 "import {undefined} from './ambient.a'; typeof value === undefined;",
 "typeof v !== (undefined); typeof v != 42n; typeof v === /a/u; typeof v == true;",
 "/* 😀 */ const café = (Symbol)();\r\n typeof café === 'nul';",
 "async function f(){class C{async g(){await work();}} return 1;}",
 "async function f(){;}",
 "type F=()=>Promise<number>; const f:F=async()=>1;",
 "type F=()=>number|Promise<number>; const f:F=async()=>1;",
 "function pick<T>(a:T,b:T):T{return a;} const f=pick(()=>Promise.resolve(1),async()=>1);",
 "function pick<T>(a:T,b:T):T{return a;} const f=pick(()=>1,async()=>1);",
 "function take<T>(...f:((n:number)=>T)[]){return f;} take(async()=>1);",
 "function take<T extends ()=>Promise<void>>(f:T){return f;} take(async()=>{return;});",
 "function take<T>(f:[()=>T]){} take([async()=>1]);",
 "function take<T>(f:{[key:string]:()=>T}){} take({run:async()=>1});",
 "function identity<T>(value:T):T{return value;} identity({async run(){return 1;}});",
 "interface I {run():Promise<number>;} class C implements I {async run(){return 1;}}",
 "class B{run():Promise<number>{return Promise.resolve(1);}} class C extends B{async run(){return 1;}}",
 "type P={then:(cb:()=>void)=>void}; const f:(()=>P)=async()=>1;",
 "type P={then:(cb:string)=>void}; const f:(()=>P)=async()=>1;",
]
paths=[]
for i,text in enumerate(texts):
 p=root/f'control-{i:03}.a';p.write_text(text+'\nexport {};\n');paths.append(str(p))
(root/'ambient.a').write_text('export declare const Symbol: SymbolConstructor; export declare const undefined: string;\n')
(root/'controls.manifest').write_text('\n'.join(paths)+'\n');(root/'tsconfig.json').write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ESNext','module':'ESNext'},'files':[paths[0]]}))
print(len(paths),'control candidates')
