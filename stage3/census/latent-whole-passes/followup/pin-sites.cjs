const fs=require('fs'),ts=require('typescript'),file=process.argv[2],sf=ts.createSourceFile(file,fs.readFileSync(file,'utf8'),ts.ScriptTarget.Latest,true),sites={};
function where(n){const p=sf.getLineAndCharacterOfPosition(n.getStart(sf));return `src/compiler/checker.ts:${p.line+1}:${p.character+1}`}
function visit(n){if(ts.isCallExpression(n)&&n.expression.getText(sf)==='elaborateElementwise'){let s=n.parent;while(s&&!ts.isStatement(s))s=s.parent;if(s)sites[where(s)]=where(n)}ts.forEachChild(n,visit)}visit(sf);console.log(JSON.stringify(sites,null,2));
