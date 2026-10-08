#!/usr/bin/env python3
import json,os,pathlib,re,subprocess,tempfile
here=pathlib.Path(__file__).resolve().parent
repo=here.parents[5];lint=repo/'stage1/cohere/lint';owned=here.parent
slug=owned.name;descriptor=json.loads((owned/'rule.json').read_text());name=descriptor['name']
def run(args,env=None):
 p=subprocess.run(args,cwd=repo,capture_output=True,env=env)
 assert p.returncode==0 and p.stderr==b'',(args,p.returncode,p.stderr.decode(errors='replace'))
 return p.stdout
with tempfile.TemporaryDirectory(prefix='cfg-native-mutant-') as td:
 temp=pathlib.Path(td);copied=temp/'lint';copied.mkdir()
 for src in lint.rglob('*'):
  if not src.is_file() or src.suffix not in ('.a','.ts'):continue
  rel=src.relative_to(lint)
  if any(p in ('testdata','evidence','validation') for p in rel.parts):continue
  dst=copied/rel;dst.parent.mkdir(parents=True,exist_ok=True)
  def external(m):
   target=m[2]
   if not target.startswith('.'):return m[0]
   resolved=(src.parent/target).resolve()
   if resolved.is_relative_to(lint):return m[0]
   return m[1]+str(resolved)+m[3]
  text=re.sub(r"(\bfrom\s+['\"])([^'\"]+)(['\"])",external,src.read_text())
  dst.write_text(text)
 change=json.loads((owned/'mutant.json').read_text());entry=change.get('file','rule.a');file=copied/'rules'/slug/entry
 text=file.read_text();assert text.count(change['from'])==1;file.write_text(text.replace(change['from'],change['to']))
 fixture=next(here.glob('*.ts.txt'));source=temp/'witness.ts';source.write_bytes(fixture.read_bytes())
 options=fixture.with_name(fixture.name.removesuffix('.ts.txt')+'.options.json')
 settings=options.read_text().strip() if options.exists() else '{}'
 manifest=temp/'manifest';manifest.write_text(str(source)+'\t'+name+'\t\t\t\t'+settings+'\n')
 runner=repo/'oracle/node.mjs'
 normal=run(['node','--no-warnings',str(runner),str(lint/'main.ts'),'--manifest',str(manifest)])
 mutant=run(['node','--no-warnings',str(runner),str(copied/'main.ts'),'--manifest',str(manifest)])
 assert normal!=mutant,'mutant survived on Node'
 binary=temp/'mutant'
 archive=temp/'checker.a'
 run(['go','build','-buildmode=c-archive','-o',str(archive),'./bridge/tsgo/archive'],dict(os.environ,GOMAXPROCS='4',CC='clang',CGO_CFLAGS='-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all'))
 run(['go','run','./cmd/adamic','build',str(copied/'main.ts'),'-o',str(binary),'--sanitize','--tsgo',str(archive)])
 native=run([str(binary),'--manifest',str(manifest)])
 assert native==mutant,'sanitized native mutant differs from Node'
 print(name+': mutant caught on Node and sanitized native; mutated outputs identical')
