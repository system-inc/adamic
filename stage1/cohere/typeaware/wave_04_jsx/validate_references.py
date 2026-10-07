#!/usr/bin/env python3
"""Compare numeric AST helpers with actual production Go AST helpers."""
from pathlib import Path
import argparse,json,os,subprocess,re,hashlib
p=argparse.ArgumentParser();p.add_argument('directory',type=Path);p.add_argument('--adamic',default='/workspace/typeaware-wave-04-landing-f801/adamic');a=p.parse_args();own=Path(__file__).resolve().parent;repo=own.parents[3];out=a.directory.resolve();out.mkdir(parents=True,exist_ok=True);runs=[]
def run(label,args,cwd=repo,env=None):
 with (out/(label+'.stdout')).open('wb') as stdout,(out/(label+'.stderr')).open('wb') as stderr:
  result=subprocess.run([str(x) for x in args],cwd=cwd,env=env,stdout=stdout,stderr=stderr)
 actual=(out/(label+'.stdout')).read_bytes();error=(out/(label+'.stderr')).read_bytes();runs.append(dict(label=label,exit=result.returncode,stdout_bytes=len(actual),stderr_bytes=len(error)))
 if result.returncode:raise RuntimeError(str(runs[-1]))
 return actual,error
virtual=repo/'cohere/internal/lint/rules/react/wave04_jsx_references_test.go';overlay=out/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(own/'testdata/references_test.go')}}))
run('go',['go','test','-overlay',overlay,'./internal/lint/rules/react','-run','^TestWave04JsxReferences$','-count=1','-v'],cwd=repo/'cohere',env=dict(os.environ,WAVE04_JSX_REFERENCES=str(out/'controls.json')))
data=json.loads((out/'controls.json').read_text());truth=data['Expected'].encode();(out/'expected.txt').write_bytes(truth)
driver=f"import {{ FragmentDeclaration }} from '{own}/jsx_fragments/declaration.a';\nimport {{ NumericTag }} from '{own}/numeric_tag.a';\nimport {{ JsxFragments }} from '{own}/jsx_fragments/rule.a';\nimport {{ JsxNoUndef }} from '{own}/jsx_no_undef/rule.a';\nimport {{ JsxNoConstructedContextValues }} from '{own}/jsx_no_constructed_context_values/rule.a';\nconst fragments = new JsxFragments(); const undef = new JsxNoUndef(); const context = new JsxNoConstructedContextValues();\nconst tree: NumericTag[] = []; const declarations: FragmentDeclaration[] = [];\n"
for i,node in enumerate(data['Nodes']):driver+=f"const args{i}: number[] = {json.dumps(node.get('Arguments') or [])}; tree.push(new NumericTag({i},{node['Kind']},{json.dumps(node['Text'],ensure_ascii=False)},{node['Receiver']},{node['Name']},0,1,args{i}));\n"
for node in data['Nodes']:driver+=f"declarations.push(new FragmentDeclaration({json.dumps(node['KindName'])},{json.dumps(node['Text'],ensure_ascii=False)},{node['Parent']},{node['Name']},{node['PropertyName']},{node['Initializer']},{node['ModuleSpecifier']}));\n"
for i,node in enumerate(data['Nodes']):
 # The driver fetches each root once and hands it to the helpers.
 driver+=f'{{ const node = tree[{i}]; if(node !== undefined) {{\nconsole.log(`reference ${{undef.reference(node,tree)}}`);\n'
 if node['Kind']!=79:driver+='console.log(`fragment ${fragments.qualifiedFragment(node,tree)}`);\n'
 driver+='console.log(`factory ${context.createContextCallee(node,tree)}`);\nconsole.log(`source ${fragments.fragmentSource(node,tree)}`);\nconsole.log(`value ${context.valueExpression(node,tree)}`);\n} }\n'
 driver+=f'{{ const declaration = declarations[{i}]; if(declaration !== undefined) {{ console.log(`declaration ${{fragments.declarationIsFragment(declaration,declarations,tree)}}`); console.log(`module ${{fragments.importModuleName(declaration,declarations)}}`); }} }}\n'
entry=out/'probe.a';entry.write_text(driver)
for variant in ['normal','asan']:
 command=[a.adamic,'build',entry,'-o',out/variant]
 if variant=='asan':command+=['--sanitize']
 run('build-'+variant,command);actual,error=run('run-'+variant,[out/variant]);assert actual==truth and not error,variant
actual,error=run('source-node',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',entry]);assert actual==truth and not error
js,_=run('emit',[a.adamic,'js',entry]);(out/'probe.js').write_bytes(js);actual,error=run('emitted-js',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',out/'probe.js']);assert actual==truth and not error
mutants=[('jsx_fragments',"imported.text === 'Fragment'","imported.text === 'Other'"),('jsx_fragments',"this.importModuleName(node,declarations) === 'react'","this.importModuleName(node,declarations) === 'preact'"),('jsx_fragments',"pattern.kind !== 'ObjectBindingPattern'","pattern.kind !== 'ArrayBindingPattern'"),('jsx_fragments',"clause.kind !== 'ImportClause'","clause.kind !== 'NamedImports'"),('jsx_fragments',"receiver.text === 'React'","receiver.text === 'Other'"),('jsx_fragments',"argument.text === 'react'","argument.text === 'preact'"),('jsx_fragments','argument.kind === 14','argument.kind === 8'),('jsx_no_undef','return root.kind === 79 ? root.identity : -1;','return root.kind === 79 && this.componentName(root.text) ? root.identity : -1;'),('jsx_no_constructed_context_values','while(receiver.kind === 218)','while(receiver.kind === 8)'),('jsx_no_constructed_context_values',"name.text !== 'value'","name.text !== 'Value'"),('jsx_no_constructed_context_values','initializer.kind !== 295','initializer.kind !== 8'),('jsx_no_constructed_context_values','if(initializer === undefined || initializer.kind !== 295) { return -1; }','if(initializer === undefined || initializer.kind !== 295) { continue; }')]
for mutant_index,(folder,before,after) in enumerate(mutants):
 module=own/folder/'rule.a';source=module.read_text();assert source.count(before)==1
 source=re.sub(r"from '(\.[^']+)'",lambda m:"from '"+str((module.parent/m[1]).resolve())+"'",source.replace(before,after))
 label=folder+'-'+str(mutant_index);path=out/(label+'-mutant.a');path.write_text(source);probe=out/(label+'-probe.a');probe.write_text(driver.replace(str(module),str(path)))
 run(label+'-mutant-build',[a.adamic,'build',probe,'-o',out/(label+'-mutant'),'--sanitize']);actual,error=run(label+'-mutant-run',[out/(label+'-mutant')]);assert actual!=truth and not error,'mutant survived or failed outside comparison'
(out/'runs.json').write_text(json.dumps(runs,indent=2)+'\n');(out/'source-sha256.json').write_text(json.dumps({str(x.relative_to(own)):hashlib.sha256(x.read_bytes()).hexdigest() for x in [own/folder/'rule.a' for folder in ['jsx_fragments','jsx_no_undef','jsx_no_constructed_context_values']]+[own/'numeric_tag.a',own/'jsx_fragments/declaration.a',own/'testdata/references_test.go',Path(__file__).resolve()]},indent=2)+'\n')
print(f'PASS partial JSX references: {len(data["Nodes"])} numeric AST nodes, {len(truth.splitlines())} records, {len(truth)} bytes; Go/native/sanitizers/source Node/emitted JS; twelve comparison-only mutants. Full source parity remains blocked.')
