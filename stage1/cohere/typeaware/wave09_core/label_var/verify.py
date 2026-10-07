"""Isolated validation; does not modify shared registrations or harness files."""
import hashlib, json, os, pathlib, statistics, subprocess, sys, time
root = pathlib.Path(__file__).resolve().parent
repo = root.parents[4]
work = pathlib.Path(os.environ.get("ADAMIC_WAVE09_CORE_ARTIFACTS", "/workspace/wave-09-core"))
work.mkdir(parents=True, exist_ok=True)

def run(name, args, cwd=repo, env=None, expected=0):
    start = time.monotonic()
    with (work / (name + ".stdout")).open("wb") as out, (work / (name + ".stderr")).open("wb") as err:
        result = subprocess.run(list(map(str, args)), cwd=cwd, env=env, stdout=out, stderr=err)
    elapsed = time.monotonic() - start
    print(name, "exit", result.returncode, "seconds", round(elapsed, 6), flush=True)
    if result.returncode != expected:
        raise RuntimeError(f"{name}: expected {expected}, got {result.returncode}; see logs")
    return (work / (name + ".stdout")).read_bytes(), elapsed

def write(name, text):
    path = work / name
    path.write_text(text)
    return path

def archive(name, sanitized=False, retained=False, fact_mutant=False):
    shared = repo / "bridge/tsgo/checker/facts.go"
    text = shared.read_text()
    anchor = "switch mode {"
    assert text.count(anchor) == 1
    text = text.replace(anchor, anchor + '\n\tcase "scope-value-symbols": return scopeValueSymbols(c, node, question)')
    replacements = {str(shared): str(write(name + "-dispatch.go", text)),
                    str(repo / "bridge/tsgo/checker/wave09_scope_value_symbols.go"): str(root / "testdata/scope_value_symbols.go")}
    if retained:
        path = repo / "bridge/tsgo/archive/main.go"
        text = path.read_text()
        anchor = "delete(programs.live, uint64(handle))"
        assert text.count(anchor) == 1
        replacements[str(path)] = str(write(name + "-retained.go", text.replace(anchor, "// Mutant retains released handle.")))
    if fact_mutant:
        text = (root / "testdata/scope_value_symbols.go").read_text()
        anchor = "ast.SymbolFlagsValue"
        assert text.count(anchor) == 1
        replacements[str(repo / "bridge/tsgo/checker/wave09_scope_value_symbols.go")] = str(write(name + "-fact.go", text.replace(anchor, "ast.SymbolFlagsVariable")))
    overlay = write(name + "-overlay.json", json.dumps({"Replace": replacements}))
    output = work / (name + ".a")
    env = os.environ.copy()
    if sanitized:
        env.update(CC="clang", CGO_CFLAGS="-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all")
    run(name + "-archive", ["go", "build", "-overlay", overlay, "-buildmode=c-archive", "-o", output, "./bridge/tsgo/archive"], env=env)
    return output

def build(name, source, library, sanitized=False):
    args = [work / "adamic", "build", source, "-o", work / name, "--tsgo", library]
    if sanitized: args.append("--sanitize")
    run(name + "-build", args)
    return work / name

def compare(name, oracle, native, config, manifest):
    want, go_time = run(name + "-go", [oracle, config, manifest])
    got, native_time = run(name + "-native", [native, config, manifest])
    assert not (work / (name + "-native.stderr")).read_bytes(), "native stderr is not empty"
    if got != want:
        diff = next((i for i, (a,b) in enumerate(zip(got,want)) if a != b), min(len(got),len(want)))
        raise AssertionError(f"{name} byte {diff}: native {got[max(0,diff-30):diff+200]!r}; Go {want[max(0,diff-30):diff+200]!r}")
    print(name, "identical bytes", len(want), want.splitlines()[-1], flush=True)
    return want, native_time, go_time

run("stage0", ["go", "build", "-o", work / "adamic", "./cmd/adamic"])
library = archive("checker")
native = build("label-var", root / "main.a", library)
virtual = repo / "cohere/adamic_wave09_next_oracle.go"
overlay = write("oracle-overlay.json", json.dumps({"Replace": {str(virtual): str(root / "testdata/oracle.go")}}))
run("oracle-build", ["go", "build", "-overlay", overlay, "-o", work / "oracle", virtual], cwd=repo / "cohere")
oracle = work / "oracle"
run("fixtures-build", ["go", "build", "-o", work / "fixtures", root / "testdata/fixtures.go"])
fixtures,_ = run("fixtures", [work / "fixtures", repo / "cohere/internal/lint/rules/core/no_label_var_test.go"])
sources = json.loads(fixtures)
sources += [
 "/* 世界 🌍 */\r\nlet 漢=1; 漢: { break 漢; }\r\n",
 "interface x {} x: {}",
 "type x=number;x: {}",
 "enum x {A} x: {}",
 "namespace x {export const a=1;} x: {}",
 "let x=1;function f(){x:{break x;}}",
 "function f(){x:{}var x;}",
 "x: {let x=1;}",
 "x: for(let x=0;;){break x;}",
 "import {x} from './helper.js';x:{}",
 "import type {X} from './helper.js';X:{}",
 "try{}catch(x){x:{break x;}}",
 "const {x}={x:1};x:{}",
 "const [x]=[1];x:{}",
]
prelude = ''
write('helper.a', 'export const x=1;export interface X {}')
paths = [str(write(f"control-{i:03}.a", prelude + source + "\nexport {};\n")) for i,source in enumerate(sources)]
config = write("tsconfig.json", json.dumps({"compilerOptions":{"strict":True,"target":"ES2022","module":"NodeNext","moduleResolution":"NodeNext","lib":["es2022","dom"]},"files":["root.d.ts"]}))
write("root.d.ts", "declare const ambientValue:number;\n")
manifest = write("controls.manifest", "\n".join(paths)+"\n")
valid,_ = run("controls-filter", [oracle, config, manifest, "--valid-sources"])
manifest.write_bytes(valid)
# Preserve every rejected fixture and its parser error; never call these covered.
supported=[]
rejected=[]
for i,path in enumerate(valid.decode().splitlines()):
    single=write("single.manifest",path+"\n")
    with (work/f"parse-{i}.stdout").open("wb") as out,(work/f"parse-{i}.stderr").open("wb") as err:
        result=subprocess.run([str(native),str(config),str(single)],stdout=out,stderr=err)
    if result.returncode==0: supported.append(path)
    elif result.returncode==70 and b"parser slice" in (work/f"parse-{i}.stderr").read_bytes():
        rejected.append({"file":path,"source":pathlib.Path(path).read_text(),"error":(work/f"parse-{i}.stderr").read_text()})
    else: raise RuntimeError(f"unexpected fixture failure: {path}")
write("parser-rejections.json",json.dumps(rejected,ensure_ascii=False,indent=2)+"\n")
manifest.write_text("\n".join(supported)+"\n")
print("controls parser supported",len(supported),"rejected",len(rejected),flush=True)
truth,_,_ = compare("controls", oracle,native,config,manifest)
assert b"\tno-label-var\t" in truth
sanitized_library = archive("checker-asan", True)
sanitized = build("label-var-asan",root / "main.a",sanitized_library,True)
compare("controls-asan",oracle,sanitized,config,manifest)

# A compiling semantic mutant reverses label scope membership.
mutant_source = work / "mutant-source"
mutant_source.mkdir(exist_ok=True)
for path in root.glob("*.a"):
    text = path.read_text().replace("from '../../", "from '" + str(repo / "stage1/cohere/typeaware") + "/")
    text = text.replace(str(repo / "stage1/cohere/typeaware") + "/../../typescript/", str(repo / "stage1/typescript") + "/")
    if path.name == "no_label_var.a":
        anchor = "scope.names.includes(label)"
        assert text.count(anchor)==1
        text = text.replace(anchor,"!scope.names.includes(label)")
    (mutant_source / path.name).write_text(text)
mutant = build("label-var-mutant",mutant_source / "main.a",library)
got,_ = run("label-var-mutant-run",[mutant,config,manifest])
assert got != truth and not (work / "label-var-mutant-run.stderr").read_bytes()
print("label-var mutant caught only by Go bytes",flush=True)
fact_library = archive("checker-fact-mutant",fact_mutant=True)
fact_mutant = build("fact-mutant",root / "main.a",fact_library)
got,_ = run("fact-mutant-run",[fact_mutant,config,manifest])
assert got != truth and not (work / "fact-mutant-run.stderr").read_bytes()
print("scope-value-symbols mutant caught only by Go bytes",flush=True)

released = write("released.a", "import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);console.log(tsgoInspect(program,file,0,5,'LabeledStatement','scope-value-symbols'));\n")
probe = write("probe.a","x: {}")
stale = build("released",released,library)
run("released-run",[stale,config,probe],expected=70)
assert (work / "released-run.stderr").read_bytes()==b"adamic: panic: invalid or released checker handle\n"
retained = archive("checker-retained",retained=True)
stale_mutant = build("released-mutant",released,retained)
run("released-mutant-run",[stale_mutant,config,probe])
print("retained handle mutant caught by required panic 70",flush=True)
measurements={}
for corpus,corpus_config,corpus_manifest in [
 ("compiler",pathlib.Path('/workspace/TypeScript-050880ce59e30b356b686bd3144efe24f875ebc8/src/compiler/tsconfig.json'),pathlib.Path('/workspace/wave-09-validation/compiler.manifest')),
 ("repository",repo/'tsconfig.json',pathlib.Path('/workspace/wave-09-validation/repository.manifest'))]:
    compare(corpus,oracle,native,corpus_config,corpus_manifest)
    compare(corpus+"-asan",oracle,sanitized,corpus_config,corpus_manifest)
    measurements[corpus]=[]
    for round_ in range(3):
        expected,native_time,go_time=compare(f"{corpus}-bench-{round_}",oracle,native,corpus_config,corpus_manifest)
        measurements[corpus].append({"native":native_time,"go":go_time,"bytes":len(expected),"sha256":hashlib.sha256(expected).hexdigest()})
    print(corpus,"medians",statistics.median(x['native'] for x in measurements[corpus]),statistics.median(x['go'] for x in measurements[corpus]),flush=True)
write("measurements.json",json.dumps(measurements,indent=2)+"\n")
print("PASS label-var independent full bytes, mutants, stale handles, sanitizers",flush=True)
