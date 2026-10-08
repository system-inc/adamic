"""Check pending rehearsal receipts without declaring a native pass."""
import argparse,hashlib,json,re
from pathlib import Path

def read(path):return json.loads(path.read_text())
def diagnostic(text):
    line=text.splitlines()[0]
    match=re.fullmatch(r'(?:adamic: )?(.+):(\d+):(\d+): (.+)',line)
    assert match,'unlocated diagnostic'
    file,line,column,message=match.groups()
    return 'src/'+file.rsplit('/src/',1)[1],int(line),int(column),message

def verify(directory,source=None):
    meta=read(directory/'provenance.json');rows=read(directory/'stops.json');logs=directory/'logs'
    assert meta['status']=='pending','rehearsal status must remain pending'
    assert meta['native_tsc_produced'] is False,'no native tsc produced'
    assert meta['compiler_sha']==(logs/'compiler-commit.txt').read_text().strip()==meta['train_sha'],'compiler pin'
    assert meta['stop_count']==len(rows)==16,'stop population'
    closure=read(directory/'closure.json')['files'];manifest=read(directory/'source-hashes.json')
    assert len(closure)==81 and len(set(closure))==81 and set(closure)==set(manifest),'closure population'
    outside=[f for f in closure if not f.startswith('src/compiler/')]
    assert sorted(outside)==read(directory/'outside-compiler.json')==['src/tsc/_namespaces/ts.ts','src/tsc/tsc.ts'],'outside compiler closure'
    if source:
        for file,digest in manifest.items():assert hashlib.sha256((source/file).read_bytes()).hexdigest()==digest,'source hash: '+file
    assert (logs/'node-entry.stdout').read_text()=='Version 6.0.3\n' and not (logs/'node-entry.stderr').read_bytes(),'Node entry observation'
    pristine=(logs/'pristine-split-0.stderr').read_text()
    for split in (0,1):assert int((logs/f'pristine-split-{split}.exit').read_text())==1,'pristine build exit'
    assert (logs/'pristine-split-0.stderr').read_bytes()==(logs/'pristine-split-1.stderr').read_bytes(),'pristine split stderr'
    golden=['undefined\n','p1\n','ok\n','undefined\n','undefined\n','ok\n','ok\n','ok\n','real\n']+['ok\n']*7
    for number,row in enumerate(rows,1):
        assert row['ordinal']==number and row['status']=='pending' and row['phase']=='checker','stop identity/status'
        assert row['outside_compiler'] is False and row['owner']['production_fix_claimed'] is False,'owner scope'
        assert row['original_line']>0 and row['original_column']>0,'original coordinates'
        assert row['message_in_pristine']==(row['message'] in pristine),'pristine diagnostic membership'
        if number in (9,14):assert row['owner']['workstream']=='scratch continuation / inference' and not row['message_in_pristine'],'scratch inference attribution'
        else:assert row['owner']['candidate_branch']=='codex/stricter-options-next' and row['owner']['integration']=='conflicted and aborted','checked-options owner'
        for split in (0,1):
            stem=f'{number:02d}-split-{split}'
            assert int((logs/(stem+'.exit')).read_text())==row[f'split_{split}_exit']==1,'native exit'
            parsed=diagnostic((logs/(stem+'.stderr')).read_text())
            assert parsed==(row['file'],row['line'],row['column'],row['message']),'native stop diagnostic'
        for suffix in ('stdout','stderr'):
            assert (logs/f'{number:02d}-split-0.{suffix}').read_bytes()==(logs/f'{number:02d}-split-1.{suffix}').read_bytes(),'split bytes'
        stem=Path(row['probe']).stem
        assert int((logs/(stem+'-node.exit')).read_text())==row['node_exit']==0,'Node witness exit'
        assert not (logs/(stem+'-node.stderr')).read_bytes(),'Node witness stderr'
        assert (logs/(stem+'-node.stdout')).read_text()==row['node_stdout']==golden[number-1],'Node witness bytes'
        assert int((logs/(stem+'-native.exit')).read_text())==row['native_exit']==1,'native witness exit'
        message=re.sub(r'^.+:\d+:\d+: ', '',(logs/(stem+'-native.stderr')).read_text().splitlines()[0])
        assert message==row['probe_message'],'witness diagnostic'
        assert row['probe_exact_message_match']==(message==row['message']),'witness exactness'
        assert re.match(r'error TS\d+:',message).group()==re.match(r'error TS\d+:',row['message']).group(),'witness diagnostic family'
    assert diagnostic(pristine)==(rows[0]['file'],rows[0]['original_line'],rows[0]['original_column'],rows[0]['message']),'first pristine stop'
    replacements=read(directory/'replacements.json')
    assert [r['ordinal'] for r in replacements]==list(range(1,16)),'replacement population'
    for r in replacements:assert r['replacement']=='{ throw new Error("tsc-entry scratch placeholder"); }' and r['end']>r['start'],'throwing body replacement'

if __name__=='__main__':
    parser=argparse.ArgumentParser();parser.add_argument('evidence',type=Path,nargs='?',default=Path(__file__).resolve().parent);parser.add_argument('--source-tree',type=Path);args=parser.parse_args()
    verify(args.evidence,args.source_tree)
    print('Evidence consistent: 16 checker stops, both split modes agree, 16 Node witnesses; rehearsal remains pending')
