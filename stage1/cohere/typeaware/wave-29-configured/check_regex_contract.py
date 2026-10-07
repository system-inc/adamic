#!/usr/bin/env python3
"""Required option constructor, explicit native gap and independent dialect witnesses."""
import json,os,subprocess,sys
from pathlib import Path
s=Path(__file__).resolve().parent;r=s.parents[3];d=Path(sys.argv[1]).resolve();d.mkdir(parents=True,exist_ok=True)
c=Path(os.environ.get('ADAMIC_COMPILER','/workspace/wave29-regex-contract-adamic'));commands=[]
def run(name,args,cwd=r,expected=0):
 with (d/(name+'.stdout')).open('wb') as out,(d/(name+'.stderr')).open('wb') as err:p=subprocess.run([str(x) for x in args],cwd=cwd,stdout=out,stderr=err)
 commands.append(dict(name=name,args=[str(x) for x in args],exit=p.returncode));(d/'commands.json').write_text(json.dumps(commands,indent=2)+'\n');assert p.returncode==expected,(name,p.returncode)
 return (d/(name+'.stdout')).read_bytes()
probe=s/'gaps/dynamic_regexp.a';node=['node','--disable-warning=ExperimentalWarning',r/'oracle/node.mjs']
assert run('dynamic-node',node+[probe,'TODO'])==b'true\n'
run('dynamic-native',[c,'build',probe,'-o',d/'dynamic'],expected=1)
assert b"RegExp with a nonconstant pattern" in (d/'dynamic-native.stderr').read_bytes()
run('configured-native',[c,'build',s/'runner.a','-o',d/'configured','--tsgo','/workspace/wave29-next-checker.a'],expected=1)
assert b"RegExp with a nonconstant pattern" in (d/'configured-native.stderr').read_bytes()
print('Required option path runs on Node; dynamic probe and configured source refuse nonconstant RegExp natively',flush=True)
# A constant-pattern mutant compiles and exits zero but ignores the runtime option.
mutant=d/'constant-mutant.a';mutant.write_text(probe.read_text().replace("new RegExp(pattern, 'u')","new RegExp('TODO', 'u')"))
run('mutant-build',[c,'build',mutant,'-o',d/'mutant'])
assert run('dynamic-node-miss',node+[probe,'NO'])==b'false\n'
assert run('mutant',[d/'mutant','NO'])==b'true\n' and not(d/'mutant.stderr').read_bytes()
print('Constant-pattern mutant compiles, exits 0, empty stderr; Node runtime-option bytes catch it',flush=True)
# Use the actual owned option factory on Node, never a handwritten evaluator.
js=d/'dialect.a'
patterns=[r'\Afoo\z','^foo$','.',r'(?i)^foo$','^[^_]+$'];names=['foo','foo\n','\r','FOO','foo']
js.write_text("import { optionPattern } from '"+str(s/'option_pattern.a')+"';\nconst patterns = "+json.dumps(patterns)+";\nconst names = "+json.dumps(names)+";\n"+"for(let index = 0; index < patterns.length; index++) { try { console.log(`${index} match ${optionPattern(patterns[index] ?? '').test(names[index] ?? '')}`); } catch { console.log(`${index} invalid`); } }\n")
go=run('dialect-go',['go','run',s/'testdata/regexp_semantics.go']);observed=run('dialect-node',node+[js]);assert go==b'0 match true\n1 match false\n2 match true\n3 match true\n4 match true\n'
assert observed==b'0 invalid\n1 match false\n2 match false\n3 invalid\n4 match true\n'
print('Actual JS option factory differs from production Go regex dialect on the preserved five witnesses',flush=True)
assert "RegexpProgram" not in (s.parent/'id_match.a').read_text()
assert not (s.parent/'regexp_program.a').exists()
print('PASS: VM removed; JS option contract and native/dialect blockers proven; full configured findings remain blocked',flush=True)
