"""Preserve small logs, source inputs, run commands and hashes, never compiled binaries."""
import pathlib,gzip,json,hashlib
ROOT=pathlib.Path(__file__).resolve().parents[4];OWN=pathlib.Path(__file__).resolve().parent
OUT=OWN/'validation';OUT.mkdir(exist_ok=True)
def collect(base,label):
    target=OUT/label;target.mkdir(exist_ok=True)
    for file in sorted(base.rglob('*')):
        if not file.is_file():continue
        if file.stat().st_size>10_000_000:continue
        if file.suffix not in ['.stdout','.stderr','.log','.json','.manifest','.tsx','.go','.a','.mjs']:continue
        if file.suffix=='.a' and file.stat().st_size>1_000_000:continue
        relative=file.relative_to(base);dest=target/(str(relative)+'.gz');dest.parent.mkdir(parents=True,exist_ok=True)
        dest.write_bytes(gzip.compress(file.read_bytes(),mtime=0))
collect(pathlib.Path('/workspace/wave-26-attributes'),'new-and-gates')
collect(pathlib.Path('/workspace/wave-26-react-regreen'),'prior15')
for directory in ['wave-26-static-validation-pass','wave-26-render-validation-final','wave-26-effect-validation-final','wave-26-rawhir','wave-26-react-reporting-final','wave-26-react-node-landing']:
    collect(pathlib.Path('/workspace')/directory,directory)
hashes={}
for file in sorted(OWN.rglob('*')):
    if file.is_file() and OUT not in file.parents and file.suffix in ['.a','.py','.go','.mjs']:
        hashes[str(file.relative_to(ROOT))]=hashlib.sha256(file.read_bytes()).hexdigest()
for name in ['jsx_structure.go','jsx_structure_test.go','symbol_locations.go','symbol_locations_test.go','facts.go']:
    file=ROOT/'bridge/tsgo/checker'/name;hashes[str(file.relative_to(ROOT))]=hashlib.sha256(file.read_bytes()).hexdigest()
(OUT/'source-hashes.json').write_text(json.dumps(hashes,indent=2)+'\n')
streams={}
S=pathlib.Path('/workspace/wave-26-attributes')
for corpus in ['controls','compiler','repository']:
    streams[corpus]={mode:hashlib.sha256((S/(corpus+'-'+mode+'.stdout')).read_bytes()).hexdigest() for mode in ['go','native','asan']}
(OUT/'diagnostic-hashes.json').write_text(json.dumps(streams,indent=2)+'\n')
