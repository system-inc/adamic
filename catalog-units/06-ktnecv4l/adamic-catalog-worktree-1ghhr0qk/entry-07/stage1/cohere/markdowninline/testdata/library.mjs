// Direct calls to the pinned original printer: no replacement algorithms or Markdown parsing.
import fs from 'node:fs';
import { createRequire } from 'node:module';
const require = createRequire(import.meta.url);
const library = process.argv[2];
const prettier = require(library);
if(prettier.version !== '3.9.6') throw new Error(`expected Prettier 3.9.6, got ${prettier.version}`);
const plugin = require(library + '/plugins/markdown');
class Path {
    constructor(nodes, indexes) { this.nodes = nodes; this.indexes = indexes; }
    get node() { return this.nodes.at(-1); }
    get parent() { return this.nodes.at(-2); }
    get grandparent() { return this.nodes.at(-3); }
    get index() { return this.indexes.at(-1); }
    get siblings() { return this.parent.children; }
    get previous() { return this.siblings[this.index - 1]; }
    get next() { return this.siblings[this.index + 1]; }
    get isFirst() { return this.index === 0; }
    get isLast() { return this.index === this.siblings.length - 1; }
    findAncestor(predicate) { return this.nodes.slice(0, -1).reverse().find(predicate); }
    hasAncestor(predicate) { return !!this.findAncestor(predicate); }
    callParent(callback) {
        const node = this.nodes.pop();
        const index = this.indexes.pop();
        try { return callback(this); }
        finally { this.nodes.push(node); this.indexes.push(index); }
    }
}
const decode = text => text.replace(/\\([nrt\\])/g, (_, c) => ({n:'\n',r:'\r',t:'\t','\\':'\\'})[c]);
const encode = text => text.replace(/[\\\n\r\t]/g, c => ({'\n':'\\n','\r':'\\r','\t':'\\t','\\':'\\\\'})[c]);
function format(mode, text) {
    const node = {type:'word',value:text};
    const parent = {type:'sentence',children:[node]};
    const root = {type:'paragraph',children:[parent]};
    let index = 0, parentIndex = 0, proseWrap = 'preserve';
    if('wefn'.includes(mode)) {
        root.type = 'emphasis';
        if('we'.includes(mode)) {
            const flank = mode === 'e' ? ' ' : 'a';
            parent.children = [{type:'word',value:flank},node,{type:'word',value:flank}];
            index = 1;
        }
        if(mode === 'n') { root.children = [{type:'sentence',children:[]},parent]; parentIndex = 1; }
    }
    else if(mode === 's') { parent.children = [{type:'whitespace',value:'\n'},node,{type:'whitespace',value:'\n'}]; index = 1; }
    else if(mode === 'p') {}
    else if('ctru'.includes(mode)) {
        node.type = 'inlineCode';
        if('tu'.includes(mode)) root.type = 'tableCell';
        if('ru'.includes(mode)) proseWrap = 'never';
    }
    else if('kv'.includes(mode)) { node.type = 'wikiLink'; if(mode === 'v') proseWrap = 'never'; }
    else if('hi'.includes(mode)) Object.assign(node,{type:'image',alt:text,title:text,url:text});
    else if(mode === 'o') Object.assign(node,{type:'image',alt:'fallback',originalAltText:text,url:text});
    else if('jlq'.includes(mode)) Object.assign(node,{type:'imageReference',alt:text,label:text,referenceType:mode==='j'?'full':mode==='l'?'collapsed':'shortcut'});
    else if(mode === 'b') Object.assign(node,{type:'footnoteReference',label:text});
    else throw new Error('unknown mode');
    const path = new Path([root,parent,node],[parentIndex,index]);
    const document = plugin.printers.mdast.print(path,{parser:'markdown',proseWrap,singleQuote:mode==='i',originalText:''},()=>{throw new Error('unexpected child print');});
    return prettier.doc.printer.printDocToString(document,{printWidth:80,tabWidth:2,useTabs:false}).formatted;
}
for(const line of fs.readFileSync(process.argv[3],'utf8').split('\n')) {
    if(line !== '') process.stdout.write(encode(format(line[0],decode(line.slice(1))))+'\n');
}
