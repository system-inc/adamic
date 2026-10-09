#!/usr/bin/env python3
# Run with the same Go/clang toolchain for both checkouts; redirect output to a log.
from pathlib import Path
import argparse, json, os, re, subprocess, time

parser = argparse.ArgumentParser(description="Compare runtime concurrency against a baseline checkout, best of five")
parser.add_argument("baseline", type=Path)
parser.add_argument("output", type=Path)
parser.add_argument("--threads", default="1,2,4")
parser.add_argument("--programs", default="nbody,trees,spectral_norm,sort,word_count,tokenizer")
parser.add_argument("--prepare-only", action="store_true")
parser.add_argument("--noise", action="store_true", help="Measure baseline against itself, two interleaved sets of five")
args = parser.parse_args()
root = Path(__file__).resolve().parents[4]
baseline = args.baseline.resolve()
out = args.output.resolve()
out.mkdir(parents=True, exist_ok=True)
threads_to_measure = [int(n) for n in args.threads.split(",")]
for version, checkout in [("main", baseline), ("concurrency", root)]:
    subprocess.run(["go", "build", "-buildvcs=false", "-o", str(out / (version + ".compiler")), "./cmd/adamic"], cwd=checkout, check=True)
flags=['clang','-std=c11','-O2','-pthread','-ffp-contract=off','-fno-optimize-sibling-calls','-I',str(root/'internal/native/runtime'),str(root/'internal/native/testdata/parallel/map.c')]+[str(f) for f in (root/'internal/native/runtime').glob('*.c')]+['-lm','-o',str(out/'map')]
subprocess.run(flags,check=True)
for name in args.programs.split(","):
 for version in ['main','concurrency']:
  compiler=str(out / (version + '.compiler'))
  for counted in [False,True]:
   destination=out/(name+'.'+version+('.count' if counted else ''))
   cmd=[compiler,'build',str(root/'bench'/ (name+'.ts')),'-o',str(destination)]
   if counted:cmd+=['--count']
   subprocess.run(cmd,check=True)
 print('prepared',name,flush=True)

if args.prepare_only:
 raise SystemExit(0)

base=out

def load():
 return list(os.getloadavg())
data={'load_before':load(),'parallel':{},'sequential':{}}
for mode in ['million','strings']:
 samples={str(n):[] for n in threads_to_measure};outputs={}
 for round in range(5):
  for threads in threads_to_measure:
   before=time.perf_counter()
   r=subprocess.run([str(base/'map'),mode,'timed'],env={**os.environ,'ADAMIC_THREADS':str(threads)},capture_output=True,text=True,check=True,timeout=120)
   elapsed=time.perf_counter()-before
   measured=float(re.search(r'map seconds ([0-9.]+)',r.stderr)[1]);samples[str(threads)].append({'map_seconds':measured,'process_seconds':elapsed});outputs[str(threads)]=r.stdout
   print('parallel',mode,'round',round+1,'threads',threads,'map',measured,'wall',elapsed,flush=True)
 if len(set(outputs.values()))!=1:raise RuntimeError('parallel outputs disagree')
 data['parallel'][mode]={'samples':samples,'stdout':next(iter(outputs.values()))}
for name in args.programs.split(","):
 samples={version:[] for version in ['main','concurrency']};outputs={};counts={}
 for round in range(5):
  for version in ['main','concurrency'] if round%2==0 else ['concurrency','main']:
   before=time.perf_counter();r=subprocess.run([str(base/(name+'.'+version))],capture_output=True,text=True,check=True,timeout=120);elapsed=time.perf_counter()-before
   samples[version].append(elapsed);outputs[version]=r.stdout
   print('sequential',name,'round',round+1,version,elapsed,flush=True)
 if len(set(outputs.values()))!=1:raise RuntimeError(name+' outputs disagree')
 for version in ['main','concurrency']:
  r=subprocess.run([str(base/(name+'.'+version+'.count'))],capture_output=True,text=True,check=True,timeout=120)
  counts[version]=r.stderr.strip();print('counts',name,version,r.stderr.strip(),flush=True)
 if counts['main']!=counts['concurrency']:raise RuntimeError(name+' counts disagree')
 data['sequential'][name]={'samples':samples,'counts':counts['main'],'stdout':outputs['main']}
if args.noise:
 data['noise'] = {}
 for name in args.programs.split(","):
  samples = {label: [] for label in ['a', 'b']}
  for round in range(5):
   for label in ['a', 'b'] if round % 2 == 0 else ['b', 'a']:
    before = time.perf_counter()
    subprocess.run([str(base / (name + '.main'))], capture_output=True, check=True, timeout=120)
    samples[label].append(time.perf_counter() - before)
  data['noise'][name] = samples
  print('noise', name, samples, flush=True)
data['load_after']=load()
(base/'measurements.json').write_text(json.dumps(data,indent=2))
print('load before',data['load_before'],'load after',data['load_after'])
print('best of five:')
for name,result in data['parallel'].items():print(name,[(n,min(x['map_seconds'] for x in s),min(x['process_seconds'] for x in s)) for n,s in result['samples'].items()])
for name,result in data['sequential'].items():
 main=min(result['samples']['main']);new=min(result['samples']['concurrency']);print(name,main,new,(new/main-1)*100)


for harness, mode in [("retain", []), ("weak_cost", []), ("weak_cost", ["held"])]:
 samples = {version: [] for version in ["main", "concurrency"]}
 for version, checkout in [("main", baseline), ("concurrency", root)]:
  runtime = checkout / "internal/native/runtime"
  flags = ["clang", "-std=c11", "-O2", "-pthread", "-ffp-contract=off", "-I", str(runtime), str(root / "internal/native/testdata/parallel" / (harness + ".c"))]
  flags += [str(f) for f in runtime.glob("*.c")] + ["-lm", "-o", str(out / (harness + "." + version))]
  subprocess.run(flags, check=True)
 before_load = load()
 for round in range(5):
  for version in ["main", "concurrency"] if round % 2 == 0 else ["concurrency", "main"]:
   result = subprocess.run([str(out / (harness + "." + version))] + mode, capture_output=True, text=True, check=True)
   samples[version].append(float(result.stdout))
 label = harness + ("_held" if mode else "")
 data[label] = {"samples": samples, "load_before": before_load, "load_after": load()}
 print(label, "best", {v: min(s) for v, s in samples.items()}, flush=True)
 if args.noise and harness == "retain":
  noise = {label: [] for label in ["a", "b"]}
  for round in range(5):
   for sample in ["a", "b"] if round % 2 == 0 else ["b", "a"]:
    result = subprocess.run([str(out / (harness + ".main"))], capture_output=True, text=True, check=True)
    noise[sample].append(float(result.stdout))
  data[label]["noise"] = noise
  print('noise retain', noise, flush=True)
(base / "measurements.json").write_text(json.dumps(data, indent=2))
