"""Independent pinned-Go verification of the eleven owned numeric declarations."""
import json, os, re, subprocess, tempfile
from pathlib import Path
HERE=Path(__file__).resolve().parent
RULES=HERE.parent
SLUGS=['react-no-unsafe','react-self-closing-comp','sort-vars','typescript-ban-tslint-comment','typescript-no-invalid-this','arrow-body-style','max-lines','no-extra-bind','default-case','no-extra-label','no-fallthrough']
COMPILER_ROOT=Path(os.environ.get('ADAMIC_LISTENER_COMPILER_ROOT','/workspace/adamic'))
def run(command,stdout,cwd=COMPILER_ROOT):
    stderr=Path(str(stdout)+'.stderr')
    with Path(stdout).open('wb') as out,stderr.open('wb') as err:
        subprocess.run(command,cwd=cwd,stdout=out,stderr=err,check=True)
    assert stderr.read_bytes()==b'', (command,stderr.read_text())
    return Path(stdout).read_bytes()
with tempfile.TemporaryDirectory(prefix='wave10-listeners-') as directory:
    scratch=Path(directory)
    oracle=scratch/'oracle.go';oracle.write_text((HERE/'testdata/listeners_oracle.go.txt').read_text())
    descriptors=[RULES/slug/'rule.json' for slug in SLUGS]
    want=run(['go','run',str(oracle),*map(str,descriptors)],scratch/'go-output')
    compiler=scratch/'adamic'
    run(['go','build','-o',str(compiler),'./cmd/adamic'],scratch/'compiler-build')
    runner=COMPILER_ROOT/'oracle/node.mjs'
    def driver(mutant=-1):
        source=[]
        for i,(slug,descriptor) in enumerate(zip(SLUGS,descriptors)):
            declaration=RULES/slug/'listeners.a'
            if i==mutant:
                text=declaration.read_text();match=re.search(r'\[(\d+)',text);assert match
                text=text[:match.start(1)]+str(int(match.group(1))+1)+text[match.end(1):]
                declaration=scratch/'mutant.a';declaration.write_text(text)
            source.append('import { syntaxKinds as kinds'+str(i)+' } from '+json.dumps(str(declaration))+';')
        for i,descriptor in enumerate(descriptors):
            name=json.loads(descriptor.read_text())['name']
            source.append('console.log('+json.dumps(name+'\t')+' + kinds'+str(i)+'.join(\',\'));')
        entry=scratch/'driver.a';entry.write_text('\n'.join(source)+'\n');return entry
    for mutant in [-1,*range(len(SLUGS))]:
        entry=driver(mutant);native=scratch/'native';module=scratch/'emitted.mjs'
        run([str(compiler),'build',str(entry),'-o',str(native),'--sanitize'],scratch/'native-build')
        run([str(compiler),'js',str(entry)],module)
        for name,command in [('Node',['node','--disable-warning=ExperimentalWarning',str(runner),str(entry)]),('emitted JS',['node','--disable-warning=ExperimentalWarning',str(runner),str(module)]),('sanitized native',[str(native)])]:
            got=run(command,scratch/'output')
            if mutant<0:assert got==want,(name,got,want)
            else:assert got!=want,(name,SLUGS[mutant],'mutant survived')
        print('baseline matches' if mutant<0 else 'compiling numeric-kind mutant caught: '+SLUGS[mutant])
    print('PASS: 11 declarations match pinned Go on Node, emitted JS and ASan/UBSan native; 11 compiling mutants caught on every runtime; bytes',len(want))
