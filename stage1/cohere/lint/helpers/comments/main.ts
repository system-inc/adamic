import { panic, programArguments, readTextFile } from 'adamic';
import { OptionsJson } from '../options_json.ts';
import { Parser } from '../../../../typescript/parser/parser.ts';
import { ParseNode, written } from '../../../../typescript/parser/nodes.ts';
import { all } from './all.ts';
import { forFile, FileCommentCache } from './for_file.ts';
import { canBeginAt } from './can_begin_at.ts';
import { sortByPosition } from './sort_by_position.ts';
function read(path: string): string {
    const result = readTextFile(path);
    if(result.kind === 'Error') {
        panic(result.message);
    }
    return result.text;
}
const args = programArguments();
const corpus = new OptionsJson(read(args[0] ?? panic('corpus path')));
const root = corpus.parse();
for(const row of corpus.node(root).children) {
    const source = corpus.node(corpus.field(row, 'source')).text;
    const name = corpus.node(corpus.field(row, 'name')).text;
    console.log(`case ${written(name)}`);
    const parser = new Parser(source, name.endsWith('.tsx') ? 'source.tsx' : 'source.ts');
    const adapted = corpus.field(row, 'ast');
    // The current parser has no JSX support and can reinterpret an element as
    // another syntax shape. Do not treat that tree as a validated JSX adapter.
    if(adapted < 0 && name.endsWith('.tsx')) {
        panic('NotYet: stage-1 JSX parser adapter');
    }
    let tree: number;
    if(adapted < 0) {
        tree = parser.file();
    }
    else {
        for(const raw of corpus.node(adapted).children) {
            const kind = corpus.node(corpus.field(raw, 'kind')).text;
            const pos = Number(corpus.node(corpus.field(raw, 'pos')).text);
            const end = Number(corpus.node(corpus.field(raw, 'end')).text);
            const children: number[] = [];
            for(const child of corpus.node(corpus.field(raw, 'children')).children) {
                children.push(Number(corpus.node(child).text));
            }
            parser.nodes.push(new ParseNode(kind, pos, end, children));
        }
        tree = Number(corpus.node(corpus.field(row, 'root')).text);
    }
    const cache = new FileCommentCache(parser, tree);
    const first = forFile(parser, tree, cache);
    const second = forFile(parser, tree, cache);
    console.log(
        `cache ${first === second ? 1 : 0} ${all(undefined, -1).length} ${forFile(undefined, -1, undefined).length}`,
    );
    console.log(`cached ${first.length} ${forFile(parser, tree, undefined).length}`);
    for(const comment of first) {
        console.log(
            `${comment.start} ${comment.end} ${comment.isBlock ? 1 : 0} ${comment.startLine} ${comment.startColumn} ${comment.endLine}\t${written(comment.text)}`,
        );
    }
    const comments = all(parser, tree);
    console.log(`comments ${comments.length}`);
    for(const comment of comments) {
        console.log(
            `${comment.start} ${comment.end} ${comment.isBlock ? 1 : 0} ${comment.startLine} ${comment.startColumn} ${comment.endLine}\t${written(comment.text)}`,
        );
    }
    comments.reverse();
    sortByPosition(comments);
    console.log(`sorted ${comments.map((comment) => comment.start.toString()).join(',')}`);
    let guard = '';
    for(let position = 0; position <= source.length; position++) {
        // Go positions are bytes; only scalar boundaries are offered here.
        const code = source.charCodeAt(position);
        if(code >= 56320 && code <= 57343) {
            continue;
        }
        guard += canBeginAt(source, position) ? '1' : '0';
    }
    console.log(`guard ${guard}`);
}
