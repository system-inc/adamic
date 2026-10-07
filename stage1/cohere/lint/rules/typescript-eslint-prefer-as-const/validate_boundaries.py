#!/usr/bin/env python3
"""Prove both explicit repair-contract refusals can fail, without shared edits."""
import json, os, shutil, subprocess, sys, tempfile
from pathlib import Path
root=Path(__file__).resolve().parents[5]
owned=Path(__file__).resolve().parent
scratch=Path(tempfile.mkdtemp(prefix='wave14-boundaries-'))
log=(owned/'evidence/boundaries.log').open('w',buffering=1)
names=['no-unnecessary-type-constraint','prefer-as-const','prefer-enum-initializers','no-extra-non-null-assertion','no-confusing-non-null-assertion','no-unnecessary-parameter-property-assignment']
def note(text):print(text,file=log,flush=True)
def observe(args,cwd=root):
 out=scratch/('out-'+str(observe.count));err=scratch/('err-'+str(observe.count));observe.count+=1
 with out.open('wb') as stdout,err.open('wb') as stderr:
  result=subprocess.run([str(a) for a in args],cwd=cwd,stdout=stdout,stderr=stderr)
 return result.returncode,out.read_bytes(),err.read_bytes()
observe.count=0
def clean(args,cwd=root):
 result=observe(args,cwd=cwd)
 if result[0] or result[2]:raise RuntimeError(repr(args)+' '+repr(result))
 return result[1]
def build_tool(source,virtual,target,cwd):
 cwd=cwd.resolve()
 overlay=scratch/(target.name+'-overlay.json');overlay.write_text(json.dumps({'Replace':{str(cwd/virtual):str(source)}}))
 clean(['go','build','-overlay='+str(overlay),'-o',target,cwd/virtual],cwd=cwd)
def runner(tree,path):
 imports="import { panic, programArguments } from 'adamic';\n"
 imports+=f"import {{ Parser }} from '{tree}/stage1/typescript/parser/parser.ts';\nimport {{ Scanner }} from '{tree}/stage1/typescript/scanner/scanner.ts';\nimport {{ RuleContext }} from '{tree}/stage1/cohere/lint/context.ts';\nimport {{ Settings }} from '{tree}/stage1/cohere/lint/settings.ts';\n"
 for i,name in enumerate(names):imports+=f"import {{ create as create{i} }} from '{tree}/stage1/cohere/lint/rules/typescript-eslint-{name}/rule.a';\n"
 imports+="const selected = programArguments()[0] ?? panic('missing selection');\nconst context = new RuleContext('', new Parser('', 'empty.ts'), new Scanner(''), selected, '', '', false, [], new Settings());\n"
 for i,name in enumerate(names):imports+=f"if(selected === '@typescript-eslint/{name}') {{ create{i}(context); }}\n"
 imports+="console.log('accepted');\n";path.write_text(imports)
def artifacts(source,label):
 binary=scratch/(label+'-native');js=scratch/(label+'.mjs');clean([builder,source,binary,js,'sanitized']);return [('Node',node+[source]),('native',[binary]),('emitted JavaScript',node+[js])]
def variant(label):
 tree=scratch/label;shutil.copytree(root/'stage1',tree/'stage1',ignore=shutil.ignore_patterns('evidence','.generated'));return tree
try:
 note('scratch='+str(scratch))
 builder=scratch/'builder';build_tool(owned/'build.go.txt','wave14_boundary_build.go',builder,root)
 node=['node','--disable-warning=ExperimentalWarning',root/'oracle/node.mjs']
 source=scratch/'refusal.a';runner(root,source);baseline=artifacts(source,'refusal')
 expected=b'adamic: panic: NotYet: shared lint Finding must carry independent edit ranges, fix arrays and suggestion arrays; use the owned complete runner\n'
 for name in names:
  for side,args in baseline:
   result=observe(args+['@typescript-eslint/'+name]);assert result==(70,b'',expected),repr(result)
  note(name+': registered factory refuses unsupported shared contract on all three runtimes')
 tree=variant('shared-mutant');repair=tree/'stage1/cohere/lint/rules/typescript-eslint-prefer-as-const/repairs.a'
 text=repair.read_text();assert text.count('if(context.enabled(name))')==1
 repair.write_text(text.replace('if(context.enabled(name))',"if(context.enabled(name) && name === 'unsupported-mutant')"))
 source=scratch/'shared-mutant.a';runner(tree,source)
 for side,args in artifacts(source,'shared-mutant'):
  for name in names:assert observe(args+['@typescript-eslint/'+name])==(0,b'accepted\n',b'')
  note('shared refusal mutant compiles and exits 0; baseline comparison catches it on '+side)
 witness=root/'stage1/cohere/lint/rules/typescript-eslint-no-unnecessary-parameter-property-assignment/gaps/unsupported-byte-boundary.ts.txt'
 manifest=scratch/'unicode.manifest';manifest.write_text(str(witness)+'\t@typescript-eslint/no-unnecessary-parameter-property-assignment\t/repository/source/unicode.ts\n')
 existing=Path(sys.argv[1]) if len(sys.argv)>1 else None
 if existing:
  baseline=[('Node',node+[owned/'backlog_runner.a']),('native',[existing/'native']),('emitted JavaScript',node+[existing/'emitted.mjs'])];oracle=existing/'oracle'
 else:
  baseline=artifacts(owned/'backlog_runner.a','unicode-baseline');oracle=scratch/'oracle';build_tool(owned/'backlog_oracle.go.txt','wave14_boundary_oracle.go',oracle,root/'cohere')
 expected=b'adamic: panic: NotYet: parameter-property suggestion splits a UTF-8 character\n'
 for side,args in baseline:
  result=observe(args+[manifest]);assert result==(70,b'case 0\n',expected),repr(result)
  note('split UTF-8 suggestion explicitly refused on '+side)
 want=clean([oracle,manifest])
 tree=variant('unicode-mutant');rule=tree/'stage1/cohere/lint/rules/typescript-eslint-no-unnecessary-parameter-property-assignment/rule.a'
 text=rule.read_text();assert text.count('charCodeAt(node.end) > 127')==1;rule.write_text(text.replace('charCodeAt(node.end) > 127','charCodeAt(node.end) > 65535'))
 source=tree/'stage1/cohere/lint/rules/typescript-eslint-prefer-as-const/backlog_runner.a'
 for side,args in artifacts(source,'unicode-mutant'):
  got=clean(args+[manifest]);assert got!=want
  note('Unicode boundary mutant compiles, exits 0 without stderr; only Go byte comparison catches it on '+side)
 note('PASS repair-contract and UTF-8-boundary refusal checks and compiling mutants')
except Exception as error:note('FAIL '+repr(error));raise
finally:log.close()
