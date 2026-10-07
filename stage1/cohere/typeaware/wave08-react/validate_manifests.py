"""Listener-only manifests against the existing independent numeric kind oracle."""
import argparse,json,pathlib,re,subprocess
parser=argparse.ArgumentParser();parser.add_argument('--artifacts',required=True);parser.add_argument('--stage0',required=True);parser.add_argument('--archive',required=True);parser.add_argument('--asan-archive',required=True)
args=parser.parse_args();source=pathlib.Path(__file__).resolve().parent;repo=source.parents[3];out=pathlib.Path(args.artifacts).resolve();out.mkdir(parents=True,exist_ok=True)
with open(out/'declarations.log','wb') as log:
 p=subprocess.run(['python3',str(source/'validate_listeners.py'),'--artifacts',str(out/'declarations'),'--stage0',args.stage0,'--archive',args.archive,'--asan-archive',args.asan_archive],stdout=log,stderr=log)
assert p.returncode==0
# The Go constant oracle printed each module followed by its numeric subscriptions.
rows={};current=''
for line in (out/'declarations/go.stdout').read_text().splitlines():
 if line.endswith('.a'):current=line;rows[current]=[]
 else:rows[current].append(int(line))
enum=repo/'cohere/TypeScript/packages/typescript/src/enums/syntaxKind.enum.ts';values={name:int(value) for name,value in re.findall(r'\b(\w+)\s*=\s*(\d+)',enum.read_text())}
modules={'no-loop-func':'no_loop_func.a','no-require-imports':'no_require_imports.a','correctness-no-import-cycle-load-time-read':'no_import_cycle_load_time_read.a','correctness-no-process-exit-after-output':'wave08-next/no_process_exit_after_output.a','correctness-no-uncleared-race-timeout':'wave08-next/no_uncleared_race_timeout.a','correctness-require-blocking-standard-streams':'wave08-next/require_blocking_standard_streams.a','globals':'wave08-react/globals.a','immutability':'wave08-react/immutability.a','no-deriving-state-in-effects':'wave08-react/no_deriving_state_in_effects.a'}
def check(path,data):
 assert set(data)=={'name','kinds'} and data['name'].split('/')[-1]==path.parent.name
 kinds=data['kinds'];assert kinds and len(kinds)==len(set(kinds))
 actual=[values[kind] for kind in kinds];assert actual==rows[modules[path.parent.name]],path
paths=sorted((source/'listeners').glob('*/rule.json'));assert len(paths)==9
for path in paths:check(path,json.loads(path.read_text()))
witness=next(path for path in paths if path.parent.name=='no-require-imports');mutant=json.loads(witness.read_text());mutant['kinds'][0]='NewExpression';(out/'mutant-rule.json').write_text(json.dumps(mutant,indent=2)+'\n')
caught=False
try:check(witness,mutant)
except AssertionError:caught=True
assert caught
print('PASS nine rule.json kinds match native declarations and Go numeric kinds; valid NewExpression manifest mutant caught by comparison',flush=True)
