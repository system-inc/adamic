#!/usr/bin/env python3
import json, os, subprocess, tempfile
from pathlib import Path
root=Path(__file__).resolve().parents[3]
evidence=Path(__file__).resolve().parent
base=json.loads((evidence/'provenance.json').read_text())['base']
pattern=(evidence/'matrix.pattern').read_text()
with tempfile.TemporaryDirectory(prefix='defend-checked-views-replay-') as directory:
    scratch=Path(directory)
    for mutant in json.loads((evidence/'plan.json').read_text()):
        mid=mutant['id']
        source=subprocess.check_output(['git','show',base+':'+mutant['file']],cwd=root,text=True)
        assert source.count(mutant['old'])==1
        copy=scratch/(mid+'.go');copy.write_text(source.replace(mutant['old'],mutant['new']))
        overlay=scratch/(mid+'.json');overlay.write_text(json.dumps({'Replace':{str(root/mutant['file']):str(copy)}}))
        package=str(Path(mutant['file']).parent)
        subprocess.run(['go','vet','-overlay='+str(overlay),'./'+package+'/'],cwd=root,check=True)
        env=dict(os.environ, ADAMIC_GATE_UNCACHED='1',ADAMIC_BUILD_CACHE_DIR=str(scratch/'cache'/mid))
        with (evidence/(mid+'-replay.log')).open('w') as log:
            result=subprocess.run(['timeout','120','go','test','-overlay='+str(overlay),'-json','-count=1','-timeout','90s','./internal/oracle/','-run',pattern],cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
        print(mid,'exit',result.returncode)
