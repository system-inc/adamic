// Rewrite only a scratch compiler tree, using locations from the stock checker census.
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const assert = require('node:assert/strict');
const [rootArg, outArg, censusArg] = process.argv.slice(2);
const root=path.resolve(rootArg),out=path.resolve(outArg),census=JSON.parse(fs.readFileSync(censusArg,'utf8'));
assert.equal(ts.version,'6.0.3');
assert.ok(!fs.existsSync(out),'scratch destination must be new');
fs.cpSync(path.join(root,'src/compiler'),path.join(out,'src/compiler'),{recursive:true});
const edits=[];
const printer=ts.createPrinter({newLine:ts.NewLineKind.LineFeed});
for(const info of census.files) {
    const file=path.join(root,info.file),text=fs.readFileSync(file,'utf8');
    const sf=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true,ts.ScriptKind.TS);
    const accesses=new Map(census.accesses.filter(a=>a.file===info.file).map(a=>[`${a.start}:${a.end}`,a]));
    const calls=new Map([...census.creations,...census.other_returning_calls].filter(a=>a.file===info.file).map(a=>[`${a.start}:${a.end}`,a]));
    const transformations=ts.transform(sf,[context=>{
        const f=context.factory;
        const helper=(name,args)=>f.createCallExpression(f.createPropertyAccessExpression(f.createIdentifier('globalThis'),name),undefined,args);
        const str=x=>f.createStringLiteral(x);
        const site=n=>{const p=sf.getLineAndCharacterOfPosition(n.getStart(sf));return `${info.file}:${p.line+1}:${p.character+1}`;};
        const unwrapped=n=>{while(ts.isAsExpression(n)||ts.isParenthesizedExpression(n)||ts.isNonNullExpression(n))n=n.expression;return n;};
        function visitor(n) {
            if(ts.isBinaryExpression(n)&&n.operatorToken.kind===ts.SyntaxKind.EqualsToken) {
                const left=unwrapped(n.left),a=accesses.get(`${left.getStart(sf)}:${left.end}`);
                let genericRange=false;
                if(info.file==='src/compiler/utilities.ts' && ts.isPropertyAccessExpression(left) && ['pos','end'].includes(left.name.text)) {
                    for(let p=n.parent;p;p=p.parent)if(ts.isFunctionDeclaration(p)) {genericRange=['setTextRangePos','setTextRangeEnd'].includes(p.name?.text);break;}
                }
                if((a?.mode==='write'||genericRange)&&(ts.isPropertyAccessExpression(left)||ts.isElementAccessExpression(left))) {
                    edits.push({file:info.file,...a,operation:'write',site:site(n)});
                    return helper('__nodearrayWrite',[ts.visitNode(left.expression,visitor),ts.isPropertyAccessExpression(left)?str(left.name.text):ts.visitNode(left.argumentExpression,visitor),ts.visitNode(n.right,visitor),str(site(n))]);
                }
            }
            const a=accesses.get(`${n.getStart(sf)}:${n.end}`);
            if(a?.mode==='read'&&(ts.isPropertyAccessExpression(n)||ts.isElementAccessExpression(n))) {
                edits.push({file:info.file,operation:'read',site:site(n),field:a.field});
                return helper('__nodearrayRead',[ts.visitNode(n.expression,visitor),ts.isPropertyAccessExpression(n)?str(n.name.text):ts.visitNode(n.argumentExpression,visitor),str(site(n))]);
            }
            const c=calls.get(`${n.getStart(sf)}:${n.end}`);
            const child=ts.visitEachChild(n,visitor,context);
            if(c && (ts.isCallExpression(n)||ts.isAsExpression(n)||ts.isTypeAssertionExpression(n)||ts.isArrayLiteralExpression(n))) {
                edits.push({file:info.file,operation:'observe',site:site(n),category:c.category});
                const phase=c.category==='NodeArray.slice (ordinary array result)'?'slice result':c.category.includes('assertion')?'promotion':c.category.startsWith('createNodeArray')?'factory result':c.category.startsWith('setTextRange')?'range result':'other result';
                return helper('__nodearrayObserve',[child,str(phase),str(site(n))]);
            }
            return child;
        }
        return node=>ts.visitNode(node,visitor);
    }]);
    fs.writeFileSync(path.join(out,info.file),printer.printFile(transformations.transformed[0]));
    transformations.dispose();
}
fs.writeFileSync(path.join(out,'instrumentation.json'),JSON.stringify(edits,null,2)+'\n');
console.log(JSON.stringify({rewritten_files:census.files.length,operations:edits.length,by_operation:edits.reduce((a,e)=>(a[e.operation]=(a[e.operation]||0)+1,a),{})},null,2));
