"""Prove the restored production factory option path with a healthy semantic mutant."""
import argparse,json,shutil,subprocess
from pathlib import Path
HERE=Path(__file__).resolve().parent;ROOT=HERE.parents[4]
p=argparse.ArgumentParser();p.add_argument('--scratch',type=Path,required=True);a=p.parse_args();S=a.scratch.resolve();copy=S/'factory-mutant';shutil.copytree(ROOT/'stage1',copy/'stage1',dirs_exist_ok=True)
file=copy/'stage1/cohere/lint/rules/base-security-require-context-access/rule.ts';text=file.read_text();anchor="export function create(context: RuleContext, rawOptions: string = context.settings.text): Rule";assert text.count(anchor)==1;file.write_text(text.replace(anchor,"export function create(context: RuleContext, rawOptions: string = '{}'): Rule",1))
def run(label,cmd):
 with (S/(label+'.log')).open('wb') as out,(S/(label+'.stderr')).open('wb') as err:r=subprocess.run(list(map(str,cmd)),cwd=ROOT,stdout=out,stderr=err)
 assert r.returncode==0 and not (S/(label+'.stderr')).read_bytes(),label
 return (S/(label+'.log')).read_bytes()
source=copy/'stage1/cohere/lint/main.ts';builder=HERE.parent/'nexus-consistency-no-screaming-snake-case/validation_build.go';run('factory-mutant-build',['go','run',builder,source,copy/'native',copy/'emitted.mjs']);want=(S/'Go-covered.log').read_bytes();manifest=S/'manifest.txt'
for backend,cmd in [('Node',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',source,'--manifest',manifest]),('emitted',['node','--disable-warning=ExperimentalWarning',ROOT/'oracle/node.mjs',copy/'emitted.mjs','--manifest',manifest]),('native',[copy/'native','--manifest',manifest])]:
 actual=run('factory-mutant-'+backend,cmd);assert actual!=want,'raw factory option mutant survived '+backend;print('factory_raw_options_ignored',backend,'healthy byte-only catch',flush=True)
(S/'factory-mutant-summary.json').write_text(json.dumps({'name':'factory_raw_options_ignored','healthy':True,'byteOnly':True,'backends':['Node','emitted','sanitized native']},indent=2));print('PASS restored raw factory option check',flush=True)
