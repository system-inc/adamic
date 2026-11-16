import subprocess as p, pathlib, time, os, json
s=pathlib.Path('/workspace/scratch/watch1532-recut');r=s/'work';t=s/'composed';env=dict(os.environ,NODE_PATH=str(r/'stage3/api/node_modules'))
ref='refs/stage3/93-builtin-imports-'
while p.run(['git','-C',str(t),'rev-parse','--verify',ref+'after'],stdout=p.DEVNULL,stderr=p.DEVNULL).returncode: time.sleep(2)
for state in ['before','after']:
 d=s/('93-'+state);d.mkdir();archive=p.Popen(['git','-C',str(t),'archive',ref+state],stdout=p.PIPE);p.run(['tar','-x','-C',str(d)],stdin=archive.stdout,check=True);archive.stdout.close();assert archive.wait()==0
with (s/'logs/93-real-pin.log').open('w') as f: result=p.run(['node',str(r/'stage3/adapt/93-builtin-imports/adapt.cjs'),'--pin',str(s/'93-before')],env=env,stdout=f,stderr=p.STDOUT)
assert result.returncode==0
script="""const fs=require('fs'),ts=require('typescript'),assert=require('assert/strict'),crypto=require('crypto');const p=require(process.argv[1]);const raw=fs.readFileSync(process.argv[2],'utf8');const sf=ts.createSourceFile('sys.ts',raw,ts.ScriptTarget.Latest,true);let owners=[];function visit(n){if(ts.isFunctionDeclaration(n)&&n.name?.text==='getNodeSystem')owners.push(n);ts.forEachChild(n,visit)}visit(sf);assert.equal(owners.length,1);const text=owners[0].getText(sf);assert.equal(text,p[0].ownerAfterText);assert.equal(crypto.createHash('sha256').update(text).digest('hex'),p[0].ownerAfterSha256);console.log(JSON.stringify({snapshotTextAndHashMatch:true,hasWriter:text.includes('const patchedWriteFile'),hasInterval:text.includes('pollingInterval: number | undefined')}));"""
with (s/'logs/93-snapshot-match.log').open('w') as f:p.run(['node','-e',script,str(r/'stage3/adapt/93-builtin-imports/proofs/targets.json'),str(s/'93-after/src/compiler/sys.ts')],env=env,stdout=f,stderr=p.STDOUT,check=True)
print('93 pin and single-step snapshot match',flush=True)
