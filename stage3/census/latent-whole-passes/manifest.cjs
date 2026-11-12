// Count top-level first-error units in the entry driver's exact source order.
const fs=require('fs'),ts=require('typescript');
const records=fs.readFileSync(process.argv[2],'utf8').trim().split('\n').map(JSON.parse);
const head=records[0];
const files=process.argv.includes('--ordered')?records.filter(r=>r.file).map(r=>r.file):head.sources;
let start=0;
// Driver sources is sorted for reporting. Actual traversal puts entry first, then resolved
// checker order; use a census with all file rows for exact ordinal mapping instead.
for(const file of files || []) {
 const sf=ts.createSourceFile(file,fs.readFileSync(file,'utf8'),ts.ScriptTarget.Latest,true);
 console.log(JSON.stringify({file,top_level_units:sf.statements.length,start,end:start+sf.statements.length,first_kind:sf.statements.length?ts.SyntaxKind[sf.statements[0].kind]:null}));
 start+=sf.statements.length;
}
