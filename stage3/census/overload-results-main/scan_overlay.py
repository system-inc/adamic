"""Clone the own-side entry overlay and allow later shards to omit repeated scans."""
import json,pathlib,sys
source,output=map(pathlib.Path,sys.argv[1:]);output.mkdir()
meta=json.loads((source/'overlay.json').read_text());keys=[k for k,p in meta['Replace'].items() if 'func latentRun(' in pathlib.Path(p).read_text()];assert len(keys)==1
k=keys[0];s=pathlib.Path(meta['Replace'][k]).read_text()
a='\t\tfunc() {\n\t\t\tdefer func() {\n\t\t\t\tif failure := recover(); failure != nil {'
assert s.count(a)==1
b='\t\t\tfresh().latentScan(file)\n\t\t}()';assert s.count(b)==1
s=s.replace(a,'\t\tif os.Getenv("CENSUS_SKIP_SCAN") != "1" {\n'+a).replace(b,b+'\n\t\t}')
p=output/'units.go';p.write_text(s);meta['Replace'][k]=str(p.resolve());(output/'overlay.json').write_text(json.dumps(meta,indent=2)+'\n')
