import pathlib,subprocess,sys
root=pathlib.Path(sys.argv[1])
roots=sorted(str(p) for p in (root/'src/compiler').rglob('*.ts'))
assert len(roots)==79,len(roots)
p=subprocess.run(['/workspace/scratch/overload-transformers-runtime/adamic','types',*roots])
sys.exit(p.returncode)
