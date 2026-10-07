const fs=require('node:fs'),path=require('node:path');
const tree=path.resolve(process.argv[2]);
if(!fs.existsSync(path.join(tree,'slice.json')))throw new Error('adaptation 83 requires a declaration slice');
const file=path.join(tree,'src/compiler/scanner.ts');let text=fs.readFileSync(file,'utf8'),changes=0;
text=text.replace(/(\? EscapeSequenceScanningFlags\.(?:ReportErrors|AnnexB|AnyUnicodeMode|AtomEscape) : )0/g,(_,prefix)=>{changes++;return prefix+'(EscapeSequenceScanningFlags.String & EscapeSequenceScanningFlags.ReportErrors)';});
if(changes)fs.writeFileSync(file,text);
console.log(JSON.stringify({zeroBranches:changes}));
