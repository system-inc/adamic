from pathlib import Path
import hashlib,json,os,shutil,subprocess,sys
repo=Path('/workspace/adamic')
out=repo/'stage3/census/next-stops-37f8314b'
sys.dont_write_bytecode=True
sys.path.insert(0,str(repo/'stage3'))
import apply
key=apply.product_key()
store=Path(os.environ.get('STAGE3_PRODUCT_STORE',str(apply.stage3_cache()/'products')))
manifest=store/(key+'.manifest')
shutil.copyfile(manifest,out/'adapted-product.manifest.json')
def timing(name):
 p=out/(name+'-time.txt')
 return p.read_text().strip() if p.exists() else 'still running at apply checkpoint'
record=dict(compiler_commit=subprocess.check_output(['git','rev-parse','origin/main'],text=True).strip(),typescript_pin=subprocess.check_output(['git','-C','/workspace/typescript-source','rev-parse','HEAD'],text=True).strip(),product_key=key,manifest_sha256=hashlib.sha256(manifest.read_bytes()).hexdigest(),setup=timing('setup'),apply=timing('apply'))
(out/'CHECKPOINT.json').write_text(json.dumps(record,indent=2)+'\n')
print(json.dumps(record))
