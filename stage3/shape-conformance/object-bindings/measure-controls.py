"""Capture full raw control observations with pinned source and binary hashes."""
import hashlib,json,pathlib,subprocess,sys
root=pathlib.Path.cwd();binary=pathlib.Path(sys.argv[1]);out=pathlib.Path(sys.argv[2]);runs=[]
for label in ['clean','host']:
 scratch=pathlib.Path('/tmp/shape-object-binding-measure-'+label);(scratch/'src/compiler').mkdir(parents=True,exist_ok=True)
 source=root/'stage3/shape-conformance/object-bindings'/(label+'.a');entry=scratch/'src/compiler/main.a';entry.write_bytes(source.read_bytes());mapped=scratch/'map.json';result=scratch/'result.json'
 subprocess.run(['node','stage3/shape-conformance/latent/fixture-sites.cjs',str(entry),str(mapped)],check=True)
 subprocess.run([str(binary),str(scratch),str(mapped),str(result)],check=True)
 r=json.loads(result.read_text());assert len(r['diagnostics'])==1 and 'TS2322' in r['diagnostics'][0]
 runs.append({'label':label,'source_sha256':hashlib.sha256(source.read_bytes()).hexdigest(),'map':json.loads(mapped.read_text()),'result':r})
out.write_text(json.dumps({'binary_sha256':hashlib.sha256(binary.read_bytes()).hexdigest(),'runs':runs},indent=2)+'\n')
print('PASS full control observations: '+str(sum(len(r['result']['sites']) for r in runs))+' sites',flush=True)
