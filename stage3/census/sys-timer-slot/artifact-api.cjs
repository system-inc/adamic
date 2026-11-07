'use strict';
const fs=require('fs'),ts=require('typescript');
const [root,out]=process.argv.slice(2);if(ts.version!=='6.0.3')throw Error(ts.version);
function inventory(file){
 const sf=ts.createSourceFile(file,fs.readFileSync(root+'/'+file,'utf8'),ts.ScriptTarget.Latest,true),timers=[];
 function visit(n){if(n.name&&['setTimeout','clearTimeout','setImmediate','clearImmediate'].includes(n.name.text)){const l=sf.getLineAndCharacterOfPosition(n.getStart());let anys=0;function count(x){if(x.kind===ts.SyntaxKind.AnyKeyword)anys++;ts.forEachChild(x,count);}count(n);timers.push({file,line:l.line+1,start:n.getStart(),end:n.end,name:n.name.text,owner:n.parent.name?.text,text:n.getText(),anyTokens:anys});}ts.forEachChild(n,visit);}visit(sf);return timers;
}
const timers=inventory('built/local/typescript.d.ts'),referenceTimers=inventory('tests/baselines/reference/api/typescript.d.ts');
fs.writeFileSync(out,JSON.stringify({typescript:ts.version,timers,referenceTimers,anyTokens:timers.reduce((a,x)=>a+x.anyTokens,0)},null,2)+'\n');console.log(JSON.stringify({declarations:timers.length,anyTokens:timers.reduce((a,x)=>a+x.anyTokens,0)}));
