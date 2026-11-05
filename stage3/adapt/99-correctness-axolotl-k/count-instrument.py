from pathlib import Path
import shutil
src=Path('/workspace/scratch/samemap-baseline')
dst=Path('/workspace/scratch/samemap-count-tree')
shutil.copytree(src,dst,symlinks=True)
p=dst/'src/compiler/builder.ts'
b=p.read_bytes(); old=b'    return next === chain.next ? chain as DiagnosticMessageChain : { ...chain as DiagnosticMessageChain, next };'
assert b.count(old)==1
new='''    const censusGlobal = globalThis as any;
    const census = censusGlobal.__sameMapChainCensus ||= (() => {
        const counts = { noInfo: 0, noInfoNonFull: 0, self: 0, selfNonFull: 0, missing: {} as Record<string, number> };
        const nodeProcess = censusGlobal.process;
        nodeProcess.on("exit", () => nodeProcess.getBuiltinModule("node:fs").writeFileSync(
            `/workspace/scratch/samemap-counts/${nodeProcess.pid}.json`, JSON.stringify(counts)));
        return counts;
    })();
    const missing = ["messageText", "category", "code"].filter(key => !(key in chain));
    census.noInfo++;
    if (missing.length) census.noInfoNonFull++;
    if (next === chain.next) {
        census.self++;
        if (missing.length) {
            census.selfNonFull++;
            const key = missing.join(",");
            census.missing[key] = (census.missing[key] || 0) + 1;
        }
    }
    return next === chain.next ? chain as DiagnosticMessageChain : { ...chain as DiagnosticMessageChain, next };'''.replace('\n','\r\n').encode()
p.write_bytes(b.replace(old,new))
Path('/workspace/scratch/samemap-counts').mkdir()
