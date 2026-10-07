// Count the actual adapted syntax independently of the immutable downcast ledger.
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const crypto = require('node:crypto');
const ts = require('/home/agent/.cache/adamic-stage3/api/node_modules/typescript');
assert.equal(ts.version, '6.0.3');
const [root, mapFile, output] = process.argv.slice(2);
const mapped = JSON.parse(fs.readFileSync(mapFile));
assert.equal(mapped.length, 2936, 'immutable original downcast denominator');
const expected = new Map(mapped.filter(row => row.adapted).map(row => [`${row.file}:${row.adapted.start}:${row.adapted.end}`, row]));
assert.equal(expected.size, mapped.filter(row => row.adapted).length, 'mapped identities must be unique');
const files = [], rows = [];
function walk(directory) {
 for (const entry of fs.readdirSync(directory, {withFileTypes:true})) {
  const filename = path.join(directory, entry.name);
  if (entry.isDirectory()) walk(filename);
  else if (/\.(ts|a)$/.test(entry.name)) files.push(filename);
 }
}
walk(path.join(root, 'src/compiler'));
for (const filename of files.sort()) {
 const file = path.relative(root, filename), text = fs.readFileSync(filename, 'utf8');
 const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
 function visit(node) {
  if (ts.isAsExpression(node) || ts.isTypeAssertionExpression(node)) {
   const start = Buffer.byteLength(text.slice(0,node.getStart(source))), end = Buffer.byteLength(text.slice(0,node.end));
   const original = expected.get(`${file}:${start}:${end}`);
   if (original) { assert.equal(node.getText(source), original.adapted.text); expected.delete(`${file}:${start}:${end}`); }
   rows.push({file,start,end,text:node.getText(source),syntax:ts.isAsExpression(node)?'as':'angle bracket',generated:file.includes('.generated.'),original_downcast_kind:original?.kind || null});
  }
  ts.forEachChild(node,visit);
 }
 visit(source);
}
assert.equal(expected.size,0,'mapped cast must exist in final adapted tree');
const counts = {};
for (const row of rows) counts[row.generated?'generated':row.syntax] = (counts[row.generated?'generated':row.syntax] || 0)+1;
const result = {typescript:ts.version,syntax_counts:counts,total_assertions:rows.length,non_generated_assertions:rows.filter(row=>!row.generated).length,original_downcast_sites:mapped.length,mapped_original_downcasts:mapped.filter(row=>row.adapted).length,unmapped_original_downcasts:mapped.filter(row=>!row.adapted).map(row=>({file:row.file,line:row.line,kind:row.kind,text:row.text})),source_hashes:Object.fromEntries(files.map(filename=>[path.relative(root,filename),crypto.createHash('sha256').update(fs.readFileSync(filename)).digest('hex')])),note:'Syntax inventory includes const assertions and upcasts. Original downcast kinds come from the immutable 2936-site ledger; adaptations are not reclassified by their declared types.',rows};
fs.writeFileSync(output,JSON.stringify(result,null,2)+'\n');
console.log(JSON.stringify({...result,rows:undefined,source_hashes:undefined},null,2));
