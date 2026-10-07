from pathlib import Path
import subprocess, json, os
rule=Path(__file__).resolve().parents[1];repository=rule.parents[4]
out=Path('/tmp/wave15-quoted-messages');out.mkdir(parents=True,exist_ok=True)
compiler=os.environ.get('ADAMIC_WAVE15_SIXTH_COMPILER','/tmp/wave15-d65-adamic')
def run(args,label,cwd=None):
 p=subprocess.run(args,capture_output=True,cwd=cwd)
 (out/(label+'.stdout')).write_bytes(p.stdout);(out/(label+'.stderr')).write_bytes(p.stderr)
 assert p.returncode==0,(label,p.returncode,p.stderr[:2000]);return p.stdout
(out/'quoted.tsx').write_bytes((rule/'testdata/quoted.jsx-source').read_bytes())
(out/'manifest').write_text(str(out/'quoted.tsx')+'\n')
(out/'tsconfig.json').write_text(json.dumps({'compilerOptions':{'jsx':'preserve','strict':True,'target':'ES2022'},'files':['quoted.tsx']}))
virtual=repository/'cohere/adamic_wave15_quoted.go'
(out/'overlay.json').write_text(json.dumps({'Replace':{str(virtual):str(repository/'stage1/cohere/typeaware/rules/react-jsx-fragments/testdata/oracle.go')}}))
run(['go','build','-overlay',str(out/'overlay.json'),'-o',str(out/'oracle'),str(virtual)],'oracle-build',repository/'cohere')
truth=run([str(out/'oracle'),str(out/'tsconfig.json'),str(out/'manifest')],'go')
expected=b''.join(sorted(line for line in truth.splitlines(keepends=True) if b'react/jsx-no-constructed-context-values' in line))
assert len(expected.splitlines())==4
for sanitize in [False,True]:
 label='messages-asan' if sanitize else 'messages'
 run([compiler,'build',str(rule/'quoted_messages.a'),'-o',str(out/label)]+(['--sanitize'] if sanitize else []),label+'-build')
 actual=run([str(out/label)],label);assert actual==expected,(label,actual,expected)
 print(label,'PASS four production Go findings',len(actual),'bytes',flush=True)
