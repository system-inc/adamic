from run import *
import shutil
before=S/'main';after=S/'head1-candidate';mutant=S/'head1-mutant'
for tree in [after,mutant]:tree.mkdir();shutil.copytree(before/'src',tree/'src')
p=after/'src/compiler/sys.ts';text=p.read_bytes()
for head in [b'declare function setTimeout(handler: () => void, timeout: number): NodeJS.Timeout;',b'declare function clearTimeout(handle: NodeJS.Timeout | undefined): void;']:
 assert text.count(head)==1;text=text.replace(head+b'\r\n',b'')
p.write_bytes(text)
for label,tree in [('main',before),('head1-candidate',after),('head1-mutant',mutant)]:
 roots=sorted(str(p) for p in (tree/'src/compiler').rglob('*.ts'));assert len(roots)==79
 assert run(label+'-types',[S/'bin/adamic','types',*roots])==0
 assert run(label+'-latent',[S/'bin/census',tree/'src/compiler',S/(label+'-latent.jsonl')],extra={'LATENT_FULL':'1','LATENT_ONLY_FILE':str(tree/'src/compiler/sys.ts'),'LATENT_ASSERT_NO_OUTPUT':'1','LATENT_TIMES':str(S/(label+'-times.jsonl'))})==0
