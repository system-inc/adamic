const fs = require('fs');
const path = require('path');
const crypto = require('crypto');
const child = require('child_process');
const pin = process.argv[2];
const ts = require(path.join(pin, 'lib/typescript.js'));
const sha = child.execFileSync('git', ['-C', pin, 'rev-parse', 'HEAD'], {encoding:'utf8'}).trim();
if (sha !== '050880ce59e30b356b686bd3144efe24f875ebc8') throw Error('wrong TypeScript pin');
const ranks = new Set(process.argv[3] ? process.argv[3].split(',').map(Number) : [7,22,31,36,51,52]);
const pairs = JSON.parse(fs.readFileSync(path.join(__dirname,'unknown-callable-pairs-ranked.json'),'utf8')).filter(p=>ranks.has(p.rank));
const rows = [];
const typesText = fs.readFileSync(path.join(pin,'src/compiler/types.ts'),'utf8');
const types = ts.createSourceFile('types.ts',typesText,ts.ScriptTarget.Latest,true);
for (const pair of pairs) {
 const w = pair.witness;
 const bytes = fs.readFileSync(path.join(pin,w.file));
 const source = ts.createSourceFile(w.file,bytes.toString('utf8'),ts.ScriptTarget.Latest,true);
 let read,decl;
 function visit(node) {
  if (ts.isPropertyAccessExpression(node) && node.name.text === pair.field) {
   const lc=source.getLineAndCharacterOfPosition(node.getStart(source));
   if (lc.line+1===w.line && lc.character+1===w.column) read=node;
  }
  ts.forEachChild(node,visit);
 }
 function declaration(node) {
  if (ts.isMethodSignature(node) && node.name.getText(types)===pair.field && node.parent.name && node.parent.name.text===pair.type) decl=node;
  ts.forEachChild(node,declaration);
 }
 visit(source);declaration(types);
 if (!read || !decl) throw Error('missing original witness '+pair.type+'.'+pair.field);
 rows.push({rank:pair.rank,type:pair.type,field:pair.field,candidateReads:pair.reads,witness:w,read:read.getText(source),call:read.parent.getText(source),utf16Start:read.getStart(source),utf16End:read.end,fileSha256:crypto.createHash('sha256').update(bytes).digest('hex'),declaration:decl.getText(types),declarationLine:types.getLineAndCharacterOfPosition(decl.getStart(types)).line+1});
}
fs.writeFileSync(path.join(__dirname,process.argv[4] || 'aggregate-original-witnesses.json'),JSON.stringify({sourceSha:sha,basis:'original declarations and read spans; reduced adjacent helpers and data carriers',members:rows},null,2)+'\n');
console.log('Verified '+rows.length+' original declarations/read spans, '+rows.reduce((n,p)=>n+p.candidateReads,0)+' conservative candidate reads.');
