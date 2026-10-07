// Owned comparison driver. --source uses Adamic's parser; default uses Go AST geometry.
import { panic, programArguments, readTextFile, utf8Length } from 'adamic';
import { Parser } from '../../../../typescript/parser/parser.ts';
import { ParseNode } from '../../../../typescript/parser/nodes.ts';
import { Scanner } from '../../../../typescript/scanner/scanner.ts';
import { OptionsJson } from '../../helpers/options_json.ts';
import { Settings } from '../../settings.ts';
import { RuleContext } from '../../context.ts';
import { Rule as Tailwind } from './rule.ts';
import { Rule as Description } from '../eslint-comments-require-description/rule.ts';
import { Rule as Font } from '../next-google-font-display/rule.ts';
function written(text: string): string {
    const pieces: string[] = [];
    for(let index = 0; index < text.length; index++) {
        const code = text.charCodeAt(index);
        pieces.push(code >= 32 && code <= 126 && code !== 92 ? text.slice(index, index + 1) : `\\u${code.toString(16).padStart(4, '0')}`);
    }
    return pieces.join('');
}
const args = programArguments();
const file = readTextFile(args[0] ?? panic('corpus path'));
if(file.kind === 'Error') { panic(file.message); }
const corpus = new OptionsJson(file.text);
const root = corpus.parse();
let count = 0;
for(const row of corpus.node(root).children) {
    const name = corpus.node(corpus.field(row, 'file')).text;
    const chunks: string[] = [];
    for(const chunk of corpus.node(corpus.field(row, 'sourceChunks')).children) { chunks.push(corpus.node(chunk).text); }
    const source = chunks.length > 0 ? chunks.join('') : corpus.node(corpus.field(row, 'source')).text;
    const rule = corpus.node(corpus.field(row, 'rule')).text;
    const parser = new Parser(source, name);
    let tree: number;
    if(args.includes('--source')) {
        if(name.endsWith('.tsx') && source.includes('<')) { panic('NotYet: stage1 JSX parser adapter'); }
        tree = parser.file();
    }
    else {
        for(const raw of corpus.node(corpus.field(row, 'ast')).children) {
            const children: number[] = [];
            for(const child of corpus.node(corpus.field(raw, 'children')).children) { children.push(Number(corpus.node(child).text)); }
            const node = new ParseNode(corpus.node(corpus.field(raw, 'kind')).text, Number(corpus.node(corpus.field(raw, 'pos')).text), Number(corpus.node(corpus.field(raw, 'end')).text), children);
            node.text = corpus.node(corpus.field(raw, 'text')).text;
            parser.nodes.push(node);
        }
        tree = Number(corpus.node(corpus.field(row, 'root')).text);
    }
    const settings = new Settings();
    settings.load(corpus.node(corpus.field(row, 'options')).text);
    const context = new RuleContext(source, parser, new Scanner(source), rule, '', '', false, [], settings);
    const tailwind = new Tailwind(context);
    const description = new Description(context);
    const font = new Font(context);
    const pending = [tree];
    while(pending.length > 0) {
        const index = pending.pop() ?? panic('missing traversal index');
        const node = parser.node(index);
        if(rule === 'structure/tailwind-no-physical-direction') { tailwind.visit(index); }
        else if(rule === '@eslint-community/eslint-comments/require-description' && node.kind === 'SourceFile') { description.visit(index); }
        else if(rule === '@next/next/google-font-display' && (node.kind === 'JsxOpeningElement' || node.kind === 'JsxSelfClosingElement')) { font.visit(index); }
        for(let child = node.children.length - 1; child >= 0; child--) { pending.push(node.children[child] ?? panic('missing child')); }
    }
    context.findings.sort((left, right) => left.start - right.start);
    count += context.findings.length;
    if(!args.includes('--count')) {
        console.log(`case ${written(name)} ${rule}`);
        for(const finding of context.findings) {
            console.log(`${utf8Length(source.slice(0, finding.start))} ${utf8Length(source.slice(0, finding.end))} ${finding.id}\t${written(finding.message)}\t${finding.repair}\t${written(finding.replacement)}\t${written(finding.suggestion)}`);
        }
        console.log(`fixed\t${written(source)}`);
    }
}
if(args.includes('--count')) { console.log(`${count}`); }
