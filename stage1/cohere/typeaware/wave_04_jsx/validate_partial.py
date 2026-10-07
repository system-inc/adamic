#!/usr/bin/env python3
"""Compare partial JSX kernels to pinned production Go helpers."""
from pathlib import Path
import argparse,json,subprocess,os,re
p=argparse.ArgumentParser();p.add_argument('directory',type=Path);p.add_argument('--adamic',default='/workspace/typeaware-wave-04-landing-f801/adamic');a=p.parse_args()
own=Path(__file__).resolve().parent;repo=own.parents[3];out=a.directory.resolve();out.mkdir(parents=True,exist_ok=True);runs=[]
def run(label,args,cwd=repo,env=None,expected=0):
 with (out/(label+'.stdout')).open('wb') as so,(out/(label+'.stderr')).open('wb') as se:
  result=subprocess.run([str(x) for x in args],cwd=cwd,env=env,stdout=so,stderr=se)
 actual=(out/(label+'.stdout')).read_bytes();error=(out/(label+'.stderr')).read_bytes();runs.append(dict(label=label,exit=result.returncode,stdout_bytes=len(actual),stderr_bytes=len(error)))
 if result.returncode!=expected:raise RuntimeError(str(runs[-1]))
 return actual,error
folders=['jsx_fragments','jsx_no_constructed_context_values','jsx_no_undef']
overlay=out/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(repo/'cohere/internal/lint/rules/react/wave04_jsx_kernels_test.go'):str(own/'testdata/kernels_test.go')}}))
run('go',['go','test','-overlay',overlay,'./internal/lint/rules/react','-run','^TestWave04JsxKernels$','-count=1','-v'],cwd=repo/'cohere',env=dict(os.environ,WAVE04_JSX_KERNELS=str(out/'controls.json')))
data=json.loads((out/'controls.json').read_text());truth=data['Expected'].encode();(out/'expected.txt').write_bytes(truth)
for folder,kinds in zip(folders,data['Listeners']):
 manifest=json.loads((own/folder/'rule.json').read_text());assert manifest['kinds']==sorted(kinds)
driver=f"import {{ FragmentNode, JsxFragments }} from '{own}/jsx_fragments/rule.a';\nimport {{ JsxNoUndef }} from '{own}/jsx_no_undef/rule.a';\nimport {{ JsxNoConstructedContextValues }} from '{own}/jsx_no_constructed_context_values/rule.a';\nconst undef = new JsxNoUndef(); const context = new JsxNoConstructedContextValues(); const fragments = new JsxFragments();\n"
for name in data['Names']:driver+='console.log(`name ${undef.componentName('+json.dumps(name,ensure_ascii=False)+')}`);\n'
for kind in data['Kinds']:driver+=f'console.log(`construction ${{context.constructionKind(new FragmentNode({kind},4,17))}}`);\n'
for name in data['FunctionNames']:driver+='console.log(`function ${context.functionRemedy('+json.dumps(name)+')}`);\n'
for kind in data['Kinds'][:10]:driver+=f'console.log(context.reportInline(new FragmentNode({kind},4,17),3).written());\n'
driver+='console.log(fragments.report(new FragmentNode(286,4,17),false).written());\nconsole.log(fragments.report(new FragmentNode(289,4,17),true).written());\nconsole.log(undef.report(new FragmentNode(79,4,17)).written());\n'
entry=out/'probe.a';entry.write_text(driver)
for variant in ['normal','asan']:
 command=[a.adamic,'build',entry,'-o',out/variant]
 if variant=='asan':command+=['--sanitize']
 run('build-'+variant,command);actual,error=run('run-'+variant,[out/variant]);assert actual==truth and not error,variant
actual,error=run('source-node',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',entry]);assert actual==truth and not error
js,_=run('emit',[a.adamic,'js',entry]);(out/'probe.js').write_bytes(js);actual,error=run('emitted-js',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',out/'probe.js']);assert actual==truth and not error
mutants=[('jsx_fragments','node.start, node.end,','node.start, node.end + 1,'),('jsx_no_undef','first < 97 || first > 122','first < 98 || first > 122'),('jsx_no_constructed_context_values',"case 211: return 'object';","case 211: return 'array';")]
for folder,before,after in mutants:
 module=own/folder/'rule.a';source=module.read_text();assert source.count(before)==1
 source=source.replace("from '../syntax_node.a'",f"from '{own}/syntax_node.a'").replace("from '../../diagnostic.ts'",f"from '{own.parent}/diagnostic.ts'").replace("from '../jsx_fragments/rule.a'",f"from '{own}/jsx_fragments/rule.a'")
 path=out/(folder+'-mutant.a');path.write_text(source.replace(before,after));probe=out/(folder+'-probe.a');probe.write_text(driver.replace(str(module),str(path)))
 run(folder+'-mutant-build',[a.adamic,'build',probe,'-o',out/(folder+'-mutant'),'--sanitize']);actual,error=run(folder+'-mutant-run',[out/(folder+'-mutant')]);assert actual!=truth and not error,'mutant survived or failed outside comparison'
# Every absent source adapter refuses explicitly; dropping the refusal is caught.
for folder,cls in [('jsx_fragments','JsxFragments'),('jsx_no_undef','JsxNoUndef'),('jsx_no_constructed_context_values','JsxNoConstructedContextValues')]:
 module=own/folder/'rule.a';probe=out/(folder+'-refusal.a');source=f"import {{ FragmentNode }} from '{own}/jsx_fragments/rule.a';\nimport {{ {cls} }} from '{module}';\nnew {cls}().run(new FragmentNode(286,4,17));\n";probe.write_text(source)
 run(folder+'-refusal-build',[a.adamic,'build',probe,'-o',out/(folder+'-refusal')]);actual,error=run(folder+'-refusal-run',[out/(folder+'-refusal')],expected=70);assert not actual and b'NotYet:' in error
 # Removing the refusal compiles and exits zero; the required exit-70 check kills it.
 changed=re.sub(r"panic\(`NotYet:[^`]+`\);",'return;',module.read_text())
 assert changed!=module.read_text()
 changed=changed.replace("from '../syntax_node.a'",f"from '{own}/syntax_node.a'").replace("from '../../diagnostic.ts'",f"from '{own.parent}/diagnostic.ts'")
 mutant=out/(folder+'-refusal-mutant.a');mutant.write_text(changed);mutant_probe=out/(folder+'-refusal-mutant-probe.a');mutant_probe.write_text(source.replace(str(module),str(mutant)))
 run(folder+'-refusal-mutant-build',[a.adamic,'build',mutant_probe,'-o',out/(folder+'-refusal-mutant')]);actual,error=run(folder+'-refusal-mutant-run',[out/(folder+'-refusal-mutant')]);assert not actual and not error
(out/'runs.json').write_text(json.dumps(runs,indent=2)+'\n')
print(f'PASS partial JSX kernels: {len(truth.splitlines())} records, {len(truth)} bytes; Go/native/sanitizers/source Node/emitted JS; three comparison-only mutants; three source-refusal mutants. Full rule parity is not covered.')
