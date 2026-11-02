"""Verify immutable inputs and archive the new raw streams without changing A/B."""
import gzip,hashlib,json,pathlib,subprocess
P=pathlib.Path(__file__).resolve().parent
BASE=P.parent
for line in (BASE/'SHA256SUMS').read_text().splitlines():
    sha,name=line.split('  ',1)
    assert hashlib.sha256((BASE/name).read_bytes()).hexdigest()==sha,name
for name,sha in json.loads((BASE/'source-manifest.json').read_text()).items():
    assert hashlib.sha256((pathlib.Path('/workspace/scratch/tree')/name).read_bytes()).hexdigest()==sha,name
assert subprocess.run(['git','-C','/workspace/scratch/compiler-c','diff','--exit-code','--','internal/']).returncode==0
assert json.loads((P/'c-resource.json').read_text())['exit_code']==0
prov=json.loads((P/'provenance.json').read_text())
assert hashlib.sha256(pathlib.Path('/workspace/scratch/census-c').read_bytes()).hexdigest()==prov['binary_sha256']['c']
raw_hashes={}
for name in ['c.jsonl','c-times.jsonl']:
    f=P/name;raw=f.read_bytes();raw_hashes[name]=hashlib.sha256(raw).hexdigest()
    dest=P/(name+'.gz');dest.write_bytes(gzip.compress(raw,mtime=0))
    assert gzip.decompress(dest.read_bytes())==raw
    f.unlink()
old=json.loads((BASE/'provenance.json').read_text())
for sub,side in [('vs-a','a'),('vs-5e5f57d3','b')]:
    delta=json.loads((P/sub/'census-delta.json').read_text())
    assert delta['raw_sha256']['before']==old['raw_stream_sha256'][side+'.jsonl']
    assert delta['raw_sha256']['after']==raw_hashes['c.jsonl']
prov['raw_stream_sha256']=raw_hashes
prov['ledger_byte_equal_to_A']=raw_hashes['c.jsonl']==old['raw_stream_sha256']['a.jsonl']
prov['ledger_byte_equal_to_B']=raw_hashes['c.jsonl']==old['raw_stream_sha256']['b.jsonl']
prov['final_checks']={'all_parent_artifact_checksums_unchanged':True,'all_79_source_hashes_unchanged':True,'internal_source_diff_empty':True,'binary_sha256_unchanged':True,'gzip_crc_and_roundtrip':True,'both_delta_raw_hashes_match':True}
(P/'provenance.json').write_text(json.dumps(prov,indent=2)+'\n')
print(json.dumps(prov['final_checks'],indent=2))
