"""Losslessly intern repeated census metadata, then run the independent audit.

Values are tagged strings/scalars, ordered arrays, or objects with interned key
and value IDs. Sharing is storage only; decoded JSON retains every observation.
"""
import gzip,hashlib,importlib.util,json,pathlib,sys

def encode(value):
 pool=[];seen={}
 def visit(value):
  if isinstance(value,str):entry=('s',value)
  elif isinstance(value,list):entry=('a',tuple(visit(v) for v in value))
  elif isinstance(value,dict):entry=('o',tuple((visit(k),visit(v)) for k,v in value.items()))
  else:entry=('v',type(value).__name__,value)
  if entry in seen:return seen[entry]
  index=len(pool);seen[entry]=index;pool.append(entry);return index
 root=visit(value)
 return {'format':'interned-census-json-v1','root':root,'values':pool}

def decode(artifact):
 assert artifact['format']=='interned-census-json-v1'
 pool=artifact['values'];memo={}
 def visit(index):
  if index in memo:return memo[index]
  entry=pool[index];tag=entry[0]
  if tag=='s':value=entry[1]
  elif tag=='v':value=entry[2]
  elif tag=='a':value=[visit(i) for i in entry[1]]
  elif tag=='o':value={visit(k):visit(v) for k,v in entry[1]}
  else:raise AssertionError('invalid intern tag')
  memo[index]=value;return value
 return visit(artifact['root'])

def json_hash(value):
 digest=hashlib.sha256()
 for piece in json.JSONEncoder(sort_keys=True,separators=(',',':')).iterencode(value):digest.update(piece.encode())
 return digest.hexdigest()

def audit(value,mapped,root):
 path=pathlib.Path(__file__).resolve().parents[1]/'latent/audit.py'
 spec=importlib.util.spec_from_file_location('latent_audit',path);module=importlib.util.module_from_spec(spec);spec.loader.exec_module(module)
 return module.audit(value,mapped,root)

if __name__=='__main__' and sys.argv[1:] == ['--self-test']:
 sample={'flags':[True,False,None,1,1.0], 'text':'λ\n😀', 'copies':[{'same':'value'},{'same':'value'}]}
 artifact=encode(sample);assert json_hash(decode(artifact))==json_hash(sample)
 scalar=encode(sample)
 for i,entry in enumerate(scalar['values']):
  if entry==('v','bool',True):scalar['values'][i]=('v','bool',1);break
 assert decode(scalar)==sample and json_hash(decode(scalar))!=json_hash(sample)
 print('conflate-true-one: canonical JSON hash catches valid metadata mutant')
 array=encode(sample)
 for i,entry in enumerate(array['values']):
  if entry[0]=='a' and entry[1]:array['values'][i]=('a',entry[1][:-1]);break
 assert json_hash(decode(array))!=json_hash(sample)
 print('drop-array-item: canonical JSON hash catches valid metadata mutant')
 print('PASS: lossless mixed scalar, Unicode and repeated-object roundtrip')
 sys.exit(0)

if __name__=='__main__':
 command,input_path,map_path,source_root=sys.argv[1:5]
 mapped=json.loads(pathlib.Path(map_path).read_text());root=pathlib.Path(source_root)
 if command=='pack':
  value=json.loads(pathlib.Path(input_path).read_text());artifact=encode(value)
  artifact['json_sha256']=json_hash(value)
  assert json_hash(decode(artifact))==artifact['json_sha256'],'interning changed an observation'
  counts=audit(value,mapped,root)
  with open(sys.argv[5],'wb') as raw:
   with gzip.GzipFile(fileobj=raw,mode='wb',filename='',mtime=0) as out:out.write(json.dumps(artifact,separators=(',',':')).encode())
  print(json.dumps({'status':'PASS','roundtrip':'all observations equal','values':len(artifact['values']),'counts':counts,'raw_sha256':hashlib.sha256(pathlib.Path(input_path).read_bytes()).hexdigest(),'artifact_bytes':pathlib.Path(sys.argv[5]).stat().st_size},indent=2))
 elif command=='audit':
  with gzip.open(input_path,'rt') as stream:artifact=json.load(stream)
  value=decode(artifact);assert json_hash(value)==artifact['json_sha256'],'artifact changed an observation'
  print(json.dumps({'status':'PASS','counts':audit(value,mapped,root)},indent=2))
 else:raise AssertionError(command)
