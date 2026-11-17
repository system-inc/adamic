import os,sys,subprocess,pathlib,json,time
s=pathlib.Path('/workspace/overload-visitor-scratch'); name=sys.argv[1];file=sys.argv[2] if len(sys.argv)>2 else 'visitorPublic.ts';out=s/(name+'-'+file.replace('/','__')+'-full.jsonl');env=dict(os.environ,LATENT_FULL='1',LATENT_ASSERT_NO_OUTPUT='1',LATENT_ONLY_FILE=str(s/name/'src/compiler'/file))
start=time.monotonic()
with open(str(out)+'.log','w') as f:
 p=subprocess.run([str(s/'latent'),str(s/name/'src/compiler'),str(out)],env=env,stdout=f,stderr=f)
meta={'exit':p.returncode,'seconds':time.monotonic()-start,'file':env['LATENT_ONLY_FILE'],'mode':'CommonJSBundle','LATENT_FULL':1,'LATENT_ASSERT_NO_OUTPUT':1}
pathlib.Path(str(out)+'.meta.json').write_text(json.dumps(meta,indent=2)+'\n');print(json.dumps(meta),flush=True);sys.exit(p.returncode)
