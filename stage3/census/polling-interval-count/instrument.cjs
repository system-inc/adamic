// Apply only to a disposable measurement copy, never stage3/adapt.
const fs=require('fs'),path=require('path'),ts=require('../../api/node_modules/typescript');
const root=path.resolve(process.argv[2]), census=JSON.parse(fs.readFileSync(process.argv[3],'utf8'));
const byFile=new Map();for(const r of census.reads){const a=byFile.get(r.file)||[];a.push(r);byFile.set(r.file,a);}
for(const [file,rs] of byFile){const name=path.join(root,file),text=fs.readFileSync(name,'utf8'),sf=ts.createSourceFile(name,text,ts.ScriptTarget.Latest,true);const edits=[];
 const add=(n,kind,id)=>edits.push({start:n.getStart(sf),end:n.end,text:`(globalThis as unknown as { __pollingMeasure: <T>(kind: string, id: string, value: T) => T }).__pollingMeasure(${JSON.stringify(kind)}, ${id}, (${n.getText(sf)}))`});
 function walk(n){if(ts.isPropertyAccessExpression(n)||ts.isElementAccessExpression(n)){const line=sf.getLineAndCharacterOfPosition(n.getStart(sf)).line+1;const r=rs.find(r=>r.line===line&&r.expression===n.getText(sf));if(r)add(n,'read',JSON.stringify(`${file}:${line}:${r.expression}`));}
 if(file==='src/compiler/sys.ts'){
 if(ts.isFunctionDeclaration(n)&&n.name?.text==='setCustomPollingValues')edits.push({start:n.body.getStart(sf)+1,end:n.body.getStart(sf)+1,text:'\n(globalThis as unknown as { __pollingMeasure: <T>(kind: string, id: string, value: T) => T }).__pollingMeasure("call", "setCustomPollingValues", 1);\n'});
 if(ts.isBinaryExpression(n)&&n.operatorToken.kind===ts.SyntaxKind.EqualsToken){const lhs=n.left.getText(sf);if(['levels[level]','pollingChunkSize','unchangedPollThresholds'].includes(lhs))add(n,'write',lhs==='levels[level]'?'"enum:" + level':JSON.stringify(lhs));}
 if(ts.isCallExpression(n)&&n.expression.getText(sf)==='system.getEnvironmentVariable'&&n.arguments[0]?.getText(sf).includes('envVar'))add(n,'env',n.arguments[0].getText(sf));
 }ts.forEachChild(n,walk);}
 walk(sf);let out=text;for(const e of edits.sort((a,b)=>b.start-a.start))out=out.slice(0,e.start)+e.text+out.slice(e.end);fs.writeFileSync(name,out);
}
