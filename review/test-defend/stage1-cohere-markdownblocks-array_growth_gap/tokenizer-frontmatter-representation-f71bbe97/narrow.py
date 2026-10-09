from pathlib import Path
exec(Path(__file__).with_name('run.py').read_text().split('# Bounded clean control precedes')[0])
id,file,a,b,why=plan[0];path=root/file;s=path.read_text()
try:
 path.write_text(s.replace(a,b));run('E1-bounded',controls+['TestTokenizerEvents_003','TestTokenizerEvents_255','TestTokenizerEvents_511'])
finally:path.write_text(s)
