"""Compare a clean-base compiler with the normal build, including environment toggles."""
import hashlib
import os
from pathlib import Path
import subprocess
import sys
import tarfile

repository = Path(sys.argv[1]).resolve()
output = Path(sys.argv[2]).resolve()
output.mkdir(parents=True,exist_ok=True)
base = output/'base'
base.mkdir(exist_ok=True)
archive = output/'base.tar'
with archive.open('wb') as destination:
    subprocess.run(['git','archive','a5630a90','go.mod','go.work','cmd','internal','bridge'],cwd=repository,stdout=destination,check=True)
with tarfile.open(archive) as source:
    source.extractall(base,filter='data')
if not (base/'cohere').exists():
    (base/'cohere').symlink_to(repository/'cohere',target_is_directory=True)
with (output/'production-diff.log').open('w') as log:
    subprocess.run(['git','diff','--exit-code','a5630a90','--','cmd','internal','bridge','go.mod','go.sum','go.work'],cwd=repository,stdout=log,stderr=subprocess.STDOUT,check=True)
for name, tree in [('base',base),('normal',repository)]:
    with (output/(name+'-build.log')).open('w') as log:
        subprocess.run(['go','build','-buildvcs=false','-o',str(output/(name+'-adamic')),'./cmd/adamic'],cwd=tree,stdout=log,stderr=subprocess.STDOUT,check=True)
fixture = repository/'internal/oracle/testdata/functions.a'
results={}
for name,binary,extra in [('base',output/'base-adamic',{}),('overlay-off',output/'normal-adamic',{}),('flag-on-production',output/'normal-adamic',{'LATENT_SPECULATIVE':'1'})]:
    env=dict(os.environ)
    for variable in ['LATENT_SPECULATIVE','LATENT_FULL','LATENT_MUTANT_NO_STUBS']:
        env.pop(variable,None)
    env.update(extra)
    for backend in ['c','js']:
        with (output/(name+'.'+backend)).open('wb') as stdout, (output/(name+'.'+backend+'.err')).open('wb') as stderr:
            process=subprocess.run([str(binary),backend,str(fixture)],cwd=repository,env=env,stdout=stdout,stderr=stderr)
        assert process.returncode==0,(name,backend)
        results[(name,backend)]=((output/(name+'.'+backend)).read_bytes(),(output/(name+'.'+backend+'.err')).read_bytes(),process.returncode)
for backend in ['c','js']:
    expected=results[('base',backend)]
    assert results[('overlay-off',backend)]==expected
    assert results[('flag-on-production',backend)]==expected
    mutant=(expected[0]+b'/* census stub leaked */\n',expected[1],expected[2])
    try:
        assert mutant==expected
    except AssertionError:
        print(backend+' leaked-stub-output mutant caught by byte comparison')
    else:
        raise AssertionError('output comparison mutant survived')
    print(f'PASS: base, overlay off, and speculative flag on production: identical {backend} ({len(expected[0])} bytes, sha256={hashlib.sha256(expected[0]).hexdigest()})')
