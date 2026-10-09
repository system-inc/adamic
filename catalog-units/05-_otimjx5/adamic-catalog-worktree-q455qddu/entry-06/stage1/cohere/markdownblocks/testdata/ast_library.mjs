// Run the original pinned fork's exported mdast preprocess on untouched parser facts.
import fs from 'node:fs';
import vm from 'node:vm';
const context={module:{exports:{}},exports:{}};
vm.runInNewContext(fs.readFileSync(process.argv[2]+'/plugins/markdown.js','utf8'),context);
const preprocess=context.module.exports.printers.mdast.preprocess;
if(typeof preprocess!=='function') throw new Error('missing original mdast preprocess');
const decode=text=>text.replace(/\\(.)/g,(_,c)=>c==='n'?'\n':c==='r'?'\r':c==='t'?'\t':c);
const encode=text=>text.replace(/[\\\n\r\t]/g,c=>c==='\\'?'\\\\':c==='\n'?'\\n':c==='\r'?'\\r':'\\t');
const serialize=root=>{
    const nodes=[], ids=new Map(), positions=new Map();
    const walk=node=>{ids.set(node,nodes.length);nodes.push(node);for(const child of node.children??[])walk(child)};
    walk(root);
    return nodes.map(node=>{
        const position=node.position;
        if(position&&!positions.has(position))positions.set(position,positions.size);
        const mask=(node.raw!==undefined?1:0)+(node.children!==undefined?2:0)+(node.ordered?4:0)+(node.isIndented?8:0)+(node.isAligned?16:0)+(node.originalAltText!==undefined?32:0)+('value'in node?64:0)+(node.isCJ?128:0)+(node.hasLeadingPunctuation?256:0)+(node.hasTrailingPunctuation?512:0)+(position?1024:0);
        return ['N',node.type,encode(node.value??''),mask,encode(node.raw??''),encode(node.originalAltText??''),position?positions.get(position):-1,position?.start.offset??0,position?.end.offset??0,position?.start.line??0,position?.end.line??0,position?.start.column??0,position?.end.column??0,node.kind??'',(node.children??[]).map(child=>ids.get(child)).join(',')].join('\t');
    }).join('\n');
};
let source='',tabWidth=4,nodes=[],children=[],positions=new Map();
const output=[];
for(const line of fs.readFileSync(process.argv[3],'utf8').split('\n')) {
    const f=line.split('\t');
    if(f[0]==='A') {source=decode(f[1]);tabWidth=Number(f[2]);nodes=[];children=[];positions=new Map()}
    else if(f[0]==='N') {
        const mask=Number(f[3]),node={type:f[1]};
        if(mask&64)node.value=decode(f[2]);
        if(mask&1)node.raw=decode(f[4]);
        if(mask&32)node.originalAltText=decode(f[5]);
        if(mask&2)node.children=[];
        node.ordered=(mask&4)!==0;node.isIndented=(mask&8)!==0;node.isAligned=(mask&16)!==0;
        node.isCJ=(mask&128)!==0;node.hasLeadingPunctuation=(mask&256)!==0;node.hasTrailingPunctuation=(mask&512)!==0;node.kind=f[13];
        if(mask&1024) {
            const key=Number(f[6]);
            if(!positions.has(key))positions.set(key,{start:{offset:Number(f[7]),line:Number(f[9]),column:Number(f[11])},end:{offset:Number(f[8]),line:Number(f[10]),column:Number(f[12])}});
            node.position=positions.get(key);
        }
        nodes.push(node);children.push(f[14]===''?[]:f[14].split(',').map(Number));
    } else if(f[0]==='F') {
        for(let i=0;i<nodes.length;i++)if(nodes[i].children)nodes[i].children=children[i].map(id=>nodes[id]);
        try {output.push(encode(serialize(preprocess(nodes[0],{originalText:source,tabWidth,parser:'markdown'}))))}
        catch(error) {output.push(encode('E\tmarkdown: '+error.message))}
    } else if(line!=='')throw new Error('unknown AST input');
}
process.stdout.write(output.join('\n')+'\n');
