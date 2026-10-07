const fs=require('node:fs'),path=require('node:path');
const tree=path.resolve(process.argv[2]);
if(!fs.existsSync(path.join(tree,'slice.json')))throw new Error('adaptation 82 requires a declaration slice');
const file=path.join(tree,'src/compiler/scanner.ts');let text=fs.readFileSync(file,'utf8'),changes=0;
for(const [name,args,call] of [['tryScan','callback: () => T','callback'],['lookAhead','callback: () => T','callback'],['scanRange','start: number, length: number, callback: () => T','start, length, callback']]) {
 const before=`        ${name},`,after=`        ${name}: <T>(${args}): T => ${name}(${call}),`;
 if(text.includes(before)){text=text.replace(before,after);changes++;}
}
if(changes)fs.writeFileSync(file,text);
console.log(JSON.stringify({wrappers:changes}));
