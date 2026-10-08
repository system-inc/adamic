const fs = require('node:fs'), path = require('node:path');
const tree=path.resolve(process.argv[2]);
if (!fs.existsSync(path.join(tree,'slice.json'))) throw new Error('adaptation 57 requires a declaration slice');
const file=path.join(tree,'src/compiler/scanner.ts');
let text=fs.readFileSync(file,'utf8'), changes=0;
if (text.includes('var text = textInitial!;')) {
    const first=text.indexOf('var text = textInitial!;'), call=text.indexOf('setText(text, start, length);',first);
    const between=text.slice(first+'var text = textInitial!;'.length,call);
    if (/\btext\b/.test(between.replace(/\/\/[^\r\n]*|\/\*[\s\S]*?\*\//g, ''))) throw new Error('text is referenced before setText');
    text=text.slice(0,call)+'setText(textInitial, start, length);'+text.slice(call+'setText(text, start, length);'.length);
    text=text.replace('var text = textInitial!;','var text: string = undefined!;'); changes+=2;
}
text=text.replaceAll('languageVersion! >= ScriptTarget.ES2015',()=>{changes++; return 'languageVersion !== undefined && languageVersion >= ScriptTarget.ES2015';});
if (changes) fs.writeFileSync(file,text);
console.log(JSON.stringify({changes}));
