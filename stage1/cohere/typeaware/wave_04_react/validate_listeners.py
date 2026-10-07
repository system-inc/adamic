#!/usr/bin/env python3
"""Compare declared named listener sets to production Go maps."""
from pathlib import Path
import argparse,json,os,subprocess,hashlib,re
p=argparse.ArgumentParser();p.add_argument('directory',type=Path);p.add_argument('--adamic',default='/workspace/typeaware-wave-04-landing-current/adamic');a=p.parse_args();own=Path(__file__).resolve().parent;repo=own.parents[3];out=a.directory.resolve();out.mkdir(parents=True,exist_ok=True);runs=[]
def run(label,args):
 with (out/(label+'.stdout')).open('wb') as so,(out/(label+'.stderr')).open('wb') as se:result=subprocess.run([str(x) for x in args],cwd=repo,stdout=so,stderr=se)
 actual=(out/(label+'.stdout')).read_bytes();error=(out/(label+'.stderr')).read_bytes();runs.append(dict(label=label,exit=result.returncode,stdout_bytes=len(actual),stderr_bytes=len(error)))
 if result.returncode:raise RuntimeError(str(runs[-1]))
 return actual,error
subjects=[['constant_casing.a','nexus/consistency_require_constant_casing.go'],['matching_return_type.a','nexus/consistency_require_matching_return_type.go'],['strict_void_return.a','typescript/strict_void_return.go'],['wave_04_next/no_process_exit_after_output.a','nexus/correctness_no_process_exit_after_output.go'],['wave_04_next/no_uncleared_race_timeout.a','nexus/correctness_no_uncleared_race_timeout.go'],['wave_04_next/require_blocking_standard_streams.a','nexus/correctness_require_blocking_standard_streams.go'],['wave_04_react/preserve_manual_memoization.a','react/preserve_manual_memoization.go'],['wave_04_react/purity.a','react/purity.go'],['wave_04_react/refs.a','react/refs.go']]
(out/'subjects.json').write_text(json.dumps(subjects));truth,_=run('go',['go','run',own/'testdata/listeners.go',out/'subjects.json',repo/'cohere/internal/lint/rules',out/'expected.json']);data=json.loads((out/'expected.json').read_text())
driver=''
for i,row in enumerate(data):driver+=f"import {{ listenerKinds as k{i} }} from '{own.parent/row['File']}';\n"
for i,row in enumerate(data):driver+=f"console.log({json.dumps(row['Name'])} + '\t' + k{i}.join(','));\n"
entry=out/'probe.a';entry.write_text(driver)
for variant in ['normal','asan']:
 command=[a.adamic,'build',entry,'-o',out/variant,'--tsgo','/workspace/typeaware-wave-04-landing-current/next/checker-'+variant+'.a']
 if variant=='asan':command+=['--sanitize']
 run('build-'+variant,command);actual,error=run('run-'+variant,[out/variant])
 if actual!=truth or error:raise RuntimeError('named listeners differ '+variant)
actual,error=run('node',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',entry])
if actual!=truth or error:raise RuntimeError('source Node differs')
# The js command rejects checker imports and has no --tsgo option.
for i,row in enumerate(data):
 original=own.parent/row['File'];source=original.read_text();pattern=r'(export const listenerKinds: readonly string\[\] = \[)("[^"]+")';matches=list(re.finditer(pattern,source))
 if len(matches)!=1:raise RuntimeError('nonunique listener declaration')
 match=matches[0];changed=source[:match.start(2)]+json.dumps('Unknown')+source[match.end(2):]
 changed=re.sub(r"from '([^']+)'",lambda m:"from '"+str((original.parent/m[1]).resolve())+"'" if m[1].startswith('.') else m[0],changed)
 module=out/f'mutant-{i}.a';module.write_text(changed);probe=out/f'mutant-{i}-probe.a';probe.write_text(driver.replace(str(original),str(module)))
 run(f'mutant-{i}-build',[a.adamic,'build',probe,'-o',out/f'mutant-{i}','--tsgo','/workspace/typeaware-wave-04-landing-current/next/checker-normal.a']);actual,error=run(f'mutant-{i}-run',[out/f'mutant-{i}'])
 if actual==truth or error or len(actual.splitlines())!=9:raise RuntimeError('metadata mutant survived or failed outside comparison')
(out/'runs.json').write_text(json.dumps(runs,indent=2)+'\n');(out/'source-sha256.json').write_text(json.dumps({str(own.parent/x[0]):hashlib.sha256((own.parent/x[0]).read_bytes()).hexdigest() for x in subjects},indent=2)+'\n')
print(f'PASS named listener declarations: nine rules, {len(truth)} bytes; Go/native/sanitized/source Node; emitted JS blocked by unlinked checker imports; nine compiling comparison-only metadata mutants.')
