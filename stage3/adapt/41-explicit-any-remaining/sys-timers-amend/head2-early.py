from run import *
import shutil
before=S/'head1-candidate';after=S/'head2-candidate';mutant=S/'head2-mutant'
for tree in [after,mutant]:tree.mkdir();shutil.copytree(before/'src',tree/'src')
old=(before/'src/compiler/sys.ts').read_bytes();text=old
changes=[(b'setTimeout?(callback: (...args: any[]) => void, ms: number, ...args: any[]): any;',b'setTimeout?<A extends unknown[]>(callback: (...args: NoInfer<A>) => void, ms: number, ...args: A): unknown;'),(b'clearTimeout?(timeoutId: any): void;',b'clearTimeout?(timeoutId: unknown): void;'),(b'pollScheduled: boolean;',b'pollScheduled: unknown;'),(b'let pollScheduled: any;',b'let pollScheduled: unknown;')]
for a,b in changes:assert text.count(a)==1;text=text.replace(a,b)
(after/'src/compiler/sys.ts').write_bytes(text)
mutated=text
for a,b in changes[:2]:mutated=mutated.replace(b,a)
(mutant/'src/compiler/sys.ts').write_bytes(mutated)
for label,tree in [('head2-candidate',after),('head2-mutant',mutant)]:
 roots=sorted(str(p) for p in (tree/'src/compiler').rglob('*.ts'));assert len(roots)==79
 assert run(label+'-types',[S/'bin/adamic','types',*roots])==0
 assert run(label+'-latent',[S/'bin/census',tree/'src/compiler',S/(label+'-latent.jsonl')],extra={'LATENT_FULL':'1','LATENT_ONLY_FILE':str(tree/'src/compiler/sys.ts'),'LATENT_ASSERT_NO_OUTPUT':'1','LATENT_TIMES':str(S/(label+'-times.jsonl'))})==0
