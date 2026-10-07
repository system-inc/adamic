from pathlib import Path
import sys,json,re
repo=Path(__file__).resolve().parents[4]
root=Path(sys.argv[1]);root.mkdir(parents=True,exist_ok=True)
source=(repo/'cohere/internal/lint/rules/core/prefer_rest_params_test.go').read_text()
cases=re.findall(r'\{"[^"\n]+", ("(?:[^"\\]|\\.)*")\}',source)
texts=[json.loads(s) for s in cases]
texts += ['function outer() { arguments; var bar = () => arguments; }','function café() {\r\n /* 😀 */ arguments[0];\r\n}']
hooks=(repo/'cohere/internal/lint/rules/react/exhaustive_deps_test.go').read_text()
for value in re.findall(r'"(?:[^"\\]|\\.)*"|`[^`]*`',hooks):
 try: text=value[1:-1] if value.startswith('`') else json.loads(value)
 except ValueError: continue
 if 'function ' in text and any(hook in text for hook in ['useEffect(', 'useCallback(', 'useMemo(', 'useLayoutEffect(', 'useImperativeHandle(', 'React.use']): texts.append(text)
regex=(repo/'cohere/internal/lint/rules/core/prefer_regex_literals_test.go').read_text()
for value in re.findall(r'source:\s*("(?:[^"\\]|\\.)*"|`[^`]*`)',regex):
 texts.append(value[1:-1] if value.startswith('`') else json.loads(value))
texts += ["const R = RegExp; new R('a');", "const {RegExp: R} = globalThis; R('a');", "const G = globalThis; G['Reg'+'Exp']('a');", "RegExp = fake; RegExp('a');", "function f(){ const R = RegExp; (R)('a'); }", "RegExp(String.raw`[\\n/]`);", "RegExp('');"]
texts += [
 "const {RegExp: R = fake} = globalThis; R('a');",
 "let R; ({RegExp: R} = globalThis); R('a');",
 "let R; ({RegExp: R = fake} = globalThis); R('a');",
 "const {RegExp} = globalThis; RegExp('a');",
 "const G = globalThis; const {RegExp: R} = G; R('a');",
 "const R = RegExp; const r = R; r('b');",
 "(RegExp || fake)('a');",
 "(fake, RegExp)('a');",
 "function C(p){let [s,setS]=useState(0);setS=p;useEffect(()=>setS(1),[]);}",
 "function C(){const [s,setS]=React.useState<number>(0);useEffect(()=>setS(1),[]);}",
 "function C(){const [s,setS]=React.useState<number>(0);useEffect(()=>setS(1));}",
 "function C(p){ useEffect(()=>log(p?.a),[p.a]); }",
 "function C(p){ const callback=()=>log(p);useEffect(callback,[]); }",
 "RegExp(String.raw`[\\\\n/]`);",
]
paths=[]
for i,text in enumerate(texts):
 p=root/f'rest-{i:03}.a';p.write_text(text+'\nexport {};\n');paths.append(str(p))
(root/'controls.manifest').write_text('\n'.join(paths)+'\n')
(root/'tsconfig.json').write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ESNext','module':'ESNext'},'files':[paths[0]]}))
print(len(paths),'controls')
