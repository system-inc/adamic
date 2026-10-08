'use strict';
const fs=require('node:fs'),path=require('node:path'),ts=require('typescript');
const rules=require('./void-rules.json');
function apply(tree) {
    const plans=[];
    for (const file of [...new Set(rules.map(r=>r.file))]) {
        const name=path.join(tree,file),before=fs.readFileSync(name,'utf8'),lines=before.split('\n');
        const source=ts.createSourceFile(file,before,ts.ScriptTarget.Latest,true);
        let count=0;
        for (const r of rules.filter(r=>r.file===file)) {
            const peers=rules.filter(p=>p.file===file && p.before===r.before && p.after===r.after).sort((a,b)=>a.line-b.line);
            const candidates=lines.flatMap((line,i)=>[r.before,r.after].includes(line.replace(/\r$/,''))?[i]:[]);
            if(candidates.length!==peers.length)throw Error('missing or duplicate void owner: '+file+':'+r.line);
            const index=candidates[peers.findIndex(p=>p.id===r.id)],actualLine=index+1;
            const line=lines[index].replace(/\r$/,'');
            if(line===r.after)continue;
            let matches=0;
            function visit(n) {
                if(n.kind===ts.SyntaxKind.VoidExpression && source.getLineAndCharacterOfPosition(n.getStart(source)).line+1===actualLine) {
                    if(ts.SyntaxKind[n.parent.kind]!==r.parent || n.expression.getText(source)!==r.expression)throw Error('void context drift');
                    matches++;
                }
                ts.forEachChild(n,visit);
            }
            visit(source);if(matches!==1)throw Error('missing or duplicate void AST');
            lines[index]=r.after+(lines[index].endsWith('\r')?'\r':'');count++;
        }
        plans.push({name,before,after:lines.join('\n'),count});
    }
    for(const p of plans)if(fs.readFileSync(p.name,'utf8')!==p.before)throw Error('concurrent void edit');
    for(const p of plans)if(p.before!==p.after)fs.writeFileSync(p.name,p.after);
    console.log(JSON.stringify({kind:'void operator',removed:plans.reduce((n,p)=>n+p.count,0)}));
}
module.exports={apply};
