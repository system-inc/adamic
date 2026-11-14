import pathlib,subprocess,sys,json
s=pathlib.Path('/workspace/scratch/path-mint-618');tree=pathlib.Path(sys.argv[1]);out=pathlib.Path(sys.argv[2]);roots=sorted((tree/'src/compiler').rglob('*.ts'));assert len(roots)==79
with out.open('w') as f:r=subprocess.run([str(s/'adamic'),'types',*map(str,roots)],stdout=f,stderr=subprocess.STDOUT)
out.with_suffix('.receipt.json').write_text(json.dumps({'roots':len(roots),'exit':r.returncode,'command':[str(s/'adamic'),'types',*map(str,roots)]},indent=2)+'\n');sys.exit(r.returncode)
