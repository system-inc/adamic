'use strict';
const fs=require('node:fs'),path=require('node:path'),ts=require('typescript');
const root=process.argv[2],rows=[];
for(const relative of fs.readdirSync(root+'/src',{recursive:true}).filter(name=>name.endsWith('.ts')).sort()) {
    const file=path.join(root,'src',relative),text=fs.readFileSync(file,'utf8');
    if(!text.includes('takeWhile')) continue;
    const source=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true);
    function visit(node) {
        if(ts.isIdentifier(node)&&node.text==='takeWhile') {
            const pos=source.getLineAndCharacterOfPosition(node.getStart(source));
            rows.push({file:'src/'+relative,line:pos.line+1,column:pos.character+1,
                call:ts.isCallExpression(node.parent)&&node.parent.expression===node,
                context:node.parent.getText(source)});
        }
        ts.forEachChild(node,visit);
    }
    visit(source);
}
console.log(JSON.stringify({files:'all src/**/*.ts',references:rows,calls:rows.filter(r=>r.call)},null,2));
