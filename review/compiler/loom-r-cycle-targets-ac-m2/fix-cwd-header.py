import pathlib,subprocess,json,hashlib
out=pathlib.Path('review/compiler/loom-r-cycle-targets-ac-m2').resolve();root=out.parents[2];path='stage3/namespace-init-sys/cwd.a'
# exec definitions need a shared globals dictionary.
namespace=dict(__file__=str(out/'check-headers.py'));exec((out/'check-headers.py').read_text().split('h=Harness()')[0],namespace)
h=namespace['Harness']();namespace['module'].Gate.aCheck(h,[path]);assert h.exits['a-check']==1 and h.result['a_check'][path]['outcome']=='checked'
(out/'cwd-stale-header.json').write_text(json.dumps(dict(result=h.result,exits=h.exits,failure=h.failure),indent=2)+'\n')
binary=out/'adamic-acheck';before=subprocess.run([str(binary),'c',path],cwd=root,capture_output=True,timeout=90);assert before.returncode==0
p=root/path;original=p.read_bytes();header,body=original.split(b'\n',1);assert header.startswith(b'// a-check: refused');p.write_bytes(body)
after=subprocess.run([str(binary),'c',path],cwd=root,capture_output=True,timeout=90);assert after.returncode==0 and before.stdout==after.stdout
(out/'cwd-header-removal.json').write_text(json.dumps(dict(path=path,removed_header=header.decode(),body_sha256=hashlib.sha256(body).hexdigest(),body_unchanged=True,emitted_c_identical=True,emitted_c_sha256=hashlib.sha256(after.stdout).hexdigest(),before_exit=before.returncode,after_exit=after.returncode),indent=2)+'\n')
print('stale header rejected; remove header only; body and emitted C identical')
