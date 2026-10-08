// mdast-util-from-markdown and the pinned fork's complete non-MDX extensions.
import { panic } from 'adamic';
import { decodeString } from './decodeString.ts';
import { identifier } from './identifier.ts';
import type { TokenSource } from './tokenSource.ts';
import { MdastArena } from './mdastArena.ts';
import type { MdastNode } from './mdastNode.ts';
import { sliceTokenChunks, serializeTokenChunks } from './tokenizerEvents.ts';
import { tokenEvent, type TokenArena, type TokenEventInterface, type MarkdownTokenInterface } from './tokenArena.ts';

function trimEnding(value: string, start: boolean, end: boolean): string {
    let text = value;
    if(start && text.startsWith('\n')) text = text.slice(1);
    if(end && text.endsWith('\n')) text = text.slice(0, -1);
    return text;
}
function tablePipes(value: string): string {
    const output: string[] = [];
    for(let index = 0; index < value.length; index++) {
        const character = value.slice(index, index + 1);
        const next = value.slice(index + 1, index + 2);
        if(character === '\\' && (next === '\\' || next === '|')) {
            output.push(next === '|' ? '|' : '\\\\');
            index++;
        }
        else output.push(character);
    }
    return output.join('');
}
export class MdastCompiler {
    readonly arena = new MdastArena();
    readonly tokens: TokenArena;
    readonly sources: readonly TokenSource[];
    readonly events: TokenEventInterface[];
    readonly stack: number[] = [];
    readonly openTokens: number[] = [];
    context = 0;
    expectingStart = false;
    flowInside = false;
    mathInside = false;
    slurp = false;
    hardBreak = false;
    inReference = false;
    referenceType = '';
    characterType = '';
    inTable = false;
    constructor(tokens: TokenArena, sources: readonly TokenSource[], events: readonly TokenEventInterface[]) {
        this.tokens = tokens;
        this.sources = sources;
        this.events = events.slice();
    }
    top(): MdastNode {
        return this.arena.node(this.stack[this.stack.length - 1] ?? panic('mdast stack'));
    }
    serialize(token: MarkdownTokenInterface): string {
        return serializeTokenChunks(
            sliceTokenChunks((this.sources[this.context] ?? panic('token source')).chunks, token),
            false,
        );
    }
    enter(type: string, parent: boolean, literal: boolean, id: number): void {
        const next = this.arena.add(type, parent, literal);
        this.top().children.push(next);
        this.stack.push(next);
        this.openTokens.push(id);
        this.top().setStart(this.tokens.token(id).start);
    }
    exit(id: number): void {
        const node = this.arena.node(this.stack.pop() ?? panic('mdast stack'));
        const token = this.tokens.token(id);
        const open = this.openTokens.pop();
        if(open === undefined) panic(`Cannot close \`${token.type}\`: it’s not open`);
        const previous = this.tokens.token(open);
        if(previous.type !== token.type)
            panic(`Cannot close \`${token.type}\`: a different token (\`${previous.type}\`) is open`);
        node.setEnd(token.end);
    }
    buffer(): void {
        this.stack.push(this.arena.add('fragment', true, false));
    }
    resume(): string {
        return this.arena.toString(this.stack.pop() ?? panic('mdast buffer'));
    }
    enterData(token: MarkdownTokenInterface): void {
        const parent = this.top();
        let tail = parent.children[parent.children.length - 1] ?? -1;
        if(tail < 0 || this.arena.node(tail).type !== 'text') {
            tail = this.arena.add('text', false, true);
            this.arena.node(tail).setStart(token.start);
            parent.children.push(tail);
        }
        this.stack.push(tail);
    }
    exitData(token: MarkdownTokenInterface): void {
        const tail = this.arena.node(this.stack.pop() ?? panic('mdast text'));
        tail.value += this.serialize(token);
        tail.setEnd(token.end);
    }
    label(token: MarkdownTokenInterface): void {
        const value = this.resume();
        this.top().label = value;
        this.top().hasLabel = true;
        this.top().identifier = identifier(this.serialize(token));
    }
    finishLink(): void {
        const node = this.top();
        if(this.inReference) {
            node.type += 'Reference';
            node.referenceType = this.referenceType === '' ? 'shortcut' : this.referenceType;
            node.url = '';
            node.title = '';
            node.hasTitle = false;
        }
        else {
            node.identifier = '';
            node.label = '';
            node.hasLabel = false;
        }
        this.referenceType = '';
    }
    paragraph(id: number): void {
        const parentId = this.stack[this.stack.length - 2] ?? -1;
        if(parentId >= 0) {
            const parent = this.arena.node(parentId);
            const node = this.top();
            const headId = node.children[0] ?? -1;
            if(
                parent.type === 'listItem' &&
                parent.hasChecked &&
                headId >= 0 &&
                this.arena.node(headId).type === 'text'
            ) {
                let first = -1;
                for(const child of parent.children) {
                    if(this.arena.node(child).type === 'paragraph') {
                        first = child;
                        break;
                    }
                }
                if(first === this.stack[this.stack.length - 1]) {
                    const head = this.arena.node(headId);
                    head.value = head.value.slice(1);
                    if(head.value === '') node.children.splice(0, 1);
                    else {
                        head.start.column++;
                        head.start.offset++;
                        node.setStart(head.start);
                    }
                }
            }
        }
        this.exit(id);
    }
    onEnter(id: number): void {
        const token = this.tokens.token(id);
        switch(token.type) {
            case 'autolink':
            case 'link':
            case 'literalAutolink':
                this.enter('link', true, false, id);
                return;
            case 'atxHeading':
            case 'setextHeading':
                this.enter('heading', true, false, id);
                return;
            case 'blockQuote':
                this.enter('blockquote', true, false, id);
                return;
            case 'emphasis':
            case 'strong':
            case 'paragraph':
                this.enter(token.type, true, false, id);
                return;
            case 'strikethrough':
                this.enter('delete', true, false, id);
                return;
            case 'codeFenced':
                this.enter('code', false, true, id);
                return;
            case 'codeIndented':
                this.enter('code', false, true, id);
                this.buffer();
                return;
            case 'codeText':
                this.enter('inlineCode', false, true, id);
                this.buffer();
                return;
            case 'htmlFlow':
            case 'htmlText':
                this.enter('html', false, true, id);
                this.buffer();
                return;
            case 'definition':
                this.enter('definition', false, false, id);
                return;
            case 'image':
                this.enter('image', false, false, id);
                return;
            case 'hardBreakEscape':
            case 'hardBreakTrailing':
                this.enter('break', false, false, id);
                return;
            case 'thematicBreak':
                this.enter('thematicBreak', false, false, id);
                return;
            case 'listOrdered':
            case 'listUnordered':
                this.enter('list', true, false, id);
                this.top().ordered = token.type === 'listOrdered';
                this.top().spread = token.spread;
                if(token.type === 'listOrdered') this.expectingStart = true;
                return;
            case 'listItem':
                this.enter('listItem', true, false, id);
                this.top().spread = token.spread;
                return;
            case 'listItemValue':
                if(this.expectingStart) {
                    const node = this.arena.node(this.stack[this.stack.length - 2] ?? panic('list ancestor'));
                    node.startValue = Number.parseInt(this.serialize(token), 10);
                    node.hasStartValue = true;
                    this.expectingStart = false;
                }
                return;
            case 'reference':
                this.referenceType = 'collapsed';
                return;
            case 'autolinkProtocol':
            case 'autolinkEmail':
            case 'characterEscape':
            case 'characterReference':
            case 'codeTextData':
            case 'data':
            case 'codeFlowValue':
            case 'htmlFlowData':
            case 'htmlTextData':
            case 'literalAutolinkEmail':
            case 'literalAutolinkHttp':
            case 'literalAutolinkWww':
                this.enterData(token);
                return;
            case 'codeFencedFenceInfo':
            case 'codeFencedFenceMeta':
            case 'definitionDestinationString':
            case 'definitionLabelString':
            case 'definitionTitleString':
            case 'label':
            case 'referenceString':
            case 'resourceDestinationString':
            case 'resourceTitleString':
            case 'gfmFootnoteCallString':
            case 'gfmFootnoteDefinitionLabelString':
            case 'mathFlowFenceMeta':
                this.buffer();
                return;
            case 'gfmFootnoteCall':
                this.enter('footnoteReference', false, false, id);
                this.top().hasLabel = true;
                return;
            case 'gfmFootnoteDefinition':
                this.enter('footnoteDefinition', true, false, id);
                this.top().hasLabel = true;
                return;
            case 'table':
                this.enter('table', true, false, id);
                for(const align of token.align) this.top().align.push(align === 'none' ? '' : align);
                this.inTable = true;
                return;
            case 'tableData':
            case 'tableHeader':
                this.enter('tableCell', true, false, id);
                return;
            case 'tableRow':
                this.enter('tableRow', true, false, id);
                return;
            case 'mathFlow':
                this.enter('math', false, true, id);
                return;
            case 'mathText':
                this.enter('inlineMath', false, true, id);
                this.buffer();
                return;
            case 'wikiLink':
                this.enter('wikiLink', false, true, id);
                this.top().valueNull = true;
                return;
            case 'liquidNode':
                this.enter('liquidNode', false, false, id);
                this.buffer();
        }
    }
    onExit(id: number): void {
        const token = this.tokens.token(id);
        switch(token.type) {
            case 'atxHeading':
            case 'autolink':
            case 'blockQuote':
            case 'definition':
            case 'emphasis':
            case 'listItem':
            case 'listOrdered':
            case 'listUnordered':
            case 'strong':
            case 'thematicBreak':
            case 'literalAutolink':
            case 'gfmFootnoteCall':
            case 'gfmFootnoteDefinition':
            case 'strikethrough':
            case 'tableData':
            case 'tableHeader':
            case 'tableRow':
            case 'wikiLink':
                this.exit(id);
                return;
            case 'paragraph':
                this.paragraph(id);
                return;
            case 'atxHeadingSequence':
                if(this.top().depth === 0) this.top().depth = this.serialize(token).length;
                return;
            case 'setextHeadingLineSequence':
                this.top().depth = this.serialize(token).startsWith('=') ? 1 : 2;
                return;
            case 'setextHeadingText':
                this.slurp = true;
                return;
            case 'setextHeading':
                this.slurp = false;
                this.exit(id);
                return;
            case 'characterEscapeValue':
            case 'codeFlowValue':
            case 'codeTextData':
            case 'data':
            case 'htmlFlowData':
            case 'htmlTextData':
                this.exitData(token);
                return;
            case 'characterReferenceMarkerHexadecimal':
            case 'characterReferenceMarkerNumeric':
                this.characterType = token.type;
                return;
            case 'characterReferenceValue': {
                let prefix = '&';
                if(this.characterType !== '')
                    prefix = this.characterType === 'characterReferenceMarkerNumeric' ? '&#' : '&#x';
                this.top().value += decodeString(`${prefix + this.serialize(token)};`);
                this.characterType = '';
                return;
            }
            case 'characterReference':
                this.arena.node(this.stack.pop() ?? panic('reference text')).setEnd(token.end);
                return;
            case 'codeFencedFenceInfo':
                this.topAfterResume('lang');
                return;
            case 'codeFencedFenceMeta':
            case 'mathFlowFenceMeta':
                this.topAfterResume('meta');
                return;
            case 'definitionLabelString':
            case 'referenceString':
            case 'gfmFootnoteCallString':
            case 'gfmFootnoteDefinitionLabelString':
                this.label(token);
                if(token.type === 'referenceString') this.referenceType = 'full';
                return;
            case 'definitionTitleString':
            case 'resourceTitleString':
                this.topAfterResume('title');
                return;
            case 'definitionDestinationString':
            case 'resourceDestinationString':
                this.topAfterResume('url');
                return;
            case 'codeFencedFence':
                if(!this.flowInside) {
                    this.buffer();
                    this.flowInside = true;
                }
                return;
            case 'codeFenced': {
                const value = this.resume();
                this.top().value = trimEnding(value, true, true);
                this.flowInside = false;
                this.exit(id);
                return;
            }
            case 'codeIndented': {
                const value = this.resume();
                this.top().value = trimEnding(value, false, true);
                this.exit(id);
                return;
            }
            case 'codeText': {
                const value = this.resume();
                this.top().value = this.inTable ? tablePipes(value) : value;
                this.exit(id);
                return;
            }
            case 'htmlFlow':
            case 'htmlText':
            case 'mathText': {
                const value = this.resume();
                this.top().value = value;
                this.exit(id);
                return;
            }
            case 'hardBreakEscape':
            case 'hardBreakTrailing':
                this.hardBreak = true;
                this.exit(id);
                return;
            case 'lineEnding':
                if(this.hardBreak) {
                    this.arena
                        .node(this.top().children[this.top().children.length - 1] ?? panic('hard break'))
                        .setEnd(token.end);
                    this.hardBreak = false;
                    return;
                }
                if(
                    !this.slurp &&
                    ['emphasis', 'fragment', 'heading', 'paragraph', 'strong', 'delete', 'liquidNode'].includes(
                        this.top().type,
                    )
                ) {
                    this.enterData(token);
                    this.exitData(token);
                }
                return;
            case 'link':
            case 'image':
                this.finishLink();
                this.exit(id);
                return;
            case 'labelText': {
                const node = this.arena.node(this.stack[this.stack.length - 2] ?? panic('label ancestor'));
                const text = this.serialize(token);
                node.label = decodeString(text);
                node.hasLabel = true;
                node.identifier = identifier(text);
                return;
            }
            case 'label': {
                const fragment = this.top();
                const value = this.resume();
                const node = this.top();
                this.inReference = true;
                if(node.type === 'link') {
                    for(const child of fragment.children) node.children.push(child);
                }
                else {
                    node.alt = value;
                    node.hasAlt = true;
                }
                return;
            }
            case 'resource':
                this.inReference = false;
                return;
            case 'autolinkProtocol':
            case 'literalAutolinkHttp':
                this.exitData(token);
                this.top().url = this.serialize(token);
                return;
            case 'autolinkEmail':
            case 'literalAutolinkEmail':
                this.exitData(token);
                this.top().url = `mailto:${this.serialize(token)}`;
                return;
            case 'literalAutolinkWww':
                this.exitData(token);
                this.top().url = `http://${this.serialize(token)}`;
                return;
            case 'table':
                this.exit(id);
                this.inTable = false;
                return;
            case 'taskListCheckValueChecked':
            case 'taskListCheckValueUnchecked': {
                const node = this.arena.node(this.stack[this.stack.length - 2] ?? panic('task ancestor'));
                node.checked = token.type === 'taskListCheckValueChecked';
                node.hasChecked = true;
                return;
            }
            case 'mathFlowFence':
                if(!this.mathInside) {
                    this.buffer();
                    this.mathInside = true;
                }
                return;
            case 'mathFlowValue':
            case 'mathTextData':
                this.enterData(token);
                this.exitData(token);
                return;
            case 'mathFlow': {
                const value = this.resume();
                this.top().value = trimEnding(value, true, true);
                this.exit(id);
                this.mathInside = false;
                return;
            }
            case 'wikiLinkTarget':
                this.top().value = this.serialize(token);
                this.top().valueNull = false;
                return;
            case 'liquidNode':
                this.resume();
                this.top().value = this.serialize(token);
                this.top().literal = true;
                this.exit(id);
        }
    }
    topAfterResume(field: string): void {
        const value = this.resume();
        const node = this.top();
        if(field === 'lang') {
            node.lang = value;
            node.hasLang = true;
        }
        else if(field === 'meta') {
            node.meta = value;
            node.hasMeta = true;
        }
        else if(field === 'title') {
            node.title = value;
            node.hasTitle = true;
        }
        else node.url = value;
    }
    prepareList(start: number, last: number): number {
        let index = start - 1;
        let length = last;
        let balance = -1;
        let spread = false;
        let item = -1;
        let blank = 0;
        let marker = false;
        while(index + 1 <= length) {
            index++;
            const event = this.events[index] ?? panic('list event');
            const token = this.tokens.token(event.token);
            if(token.type === 'listOrdered' || token.type === 'listUnordered' || token.type === 'blockQuote') {
                balance += event.enter ? 1 : -1;
                marker = false;
            }
            else if(token.type === 'lineEndingBlank') {
                if(event.enter) {
                    if(item >= 0 && !marker && balance === 0 && blank === 0) blank = index;
                    marker = false;
                }
            }
            else if(
                ![
                    'linePrefix',
                    'listItemValue',
                    'listItemMarker',
                    'listItemPrefix',
                    'listItemPrefixWhitespace',
                ].includes(token.type)
            )
                marker = false;
            if(
                (balance === 0 && event.enter && token.type === 'listItemPrefix') ||
                (balance === -1 && !event.enter && (token.type === 'listOrdered' || token.type === 'listUnordered'))
            ) {
                if(item >= 0) {
                    let tail = index;
                    let line = 0;
                    while(tail > 0) {
                        tail--;
                        const previous = this.events[tail] ?? panic('list tail');
                        const previousToken = this.tokens.token(previous.token);
                        if(previousToken.type === 'lineEnding' || previousToken.type === 'lineEndingBlank') {
                            if(!previous.enter) continue;
                            if(line !== 0) {
                                this.tokens.token((this.events[line] ?? panic('list line')).token).type =
                                    'lineEndingBlank';
                                spread = true;
                            }
                            previousToken.type = 'lineEnding';
                            line = tail;
                        }
                        else if(
                            ![
                                'linePrefix',
                                'blockQuotePrefix',
                                'blockQuotePrefixWhitespace',
                                'blockQuoteMarker',
                                'listItemIndent',
                            ].includes(previousToken.type)
                        )
                            break;
                    }
                    if(blank !== 0 && (line === 0 || blank < line)) this.tokens.token(item).spread = true;
                    this.tokens.token(item).end =
                        line !== 0
                            ? this.tokens.token((this.events[line] ?? panic('list line')).token).start
                            : token.end;
                    this.events.splice(line !== 0 ? line : index, 0, tokenEvent(false, item, event.context));
                    index++;
                    length++;
                }
                if(token.type === 'listItemPrefix') {
                    item = this.tokens.add('listItem', token.start);
                    this.events.splice(index, 0, tokenEvent(true, item, event.context));
                    index++;
                    length++;
                    blank = 0;
                    marker = true;
                }
            }
        }
        this.tokens.token((this.events[start] ?? panic('list start')).token).spread = spread;
        return length;
    }
    compile(): MdastArena {
        const root = this.arena.add('root', true, false);
        this.stack.push(root);
        const lists: number[] = [];
        for(let index = 0; index < this.events.length; index++) {
            const event = this.events[index] ?? panic('event');
            const token = this.tokens.token(event.token);
            if(token.type === 'listOrdered' || token.type === 'listUnordered') {
                if(event.enter) lists.push(index);
                else index = this.prepareList(lists.pop() ?? panic('list stack'), index);
            }
        }
        for(const event of this.events) {
            this.context = event.context;
            if(event.enter) this.onEnter(event.token);
            else this.onExit(event.token);
        }
        if(this.openTokens.length > 0)
            panic(
                `Cannot close document, a token (\`${this.tokens.token(this.openTokens[this.openTokens.length - 1] ?? panic('open token')).type}\`) is still open`,
            );
        if(this.events.length > 0) {
            this.arena.node(root).setStart(this.tokens.token((this.events[0] ?? panic('root start')).token).start);
            this.arena
                .node(root)
                .setEnd(this.tokens.token((this.events[this.events.length - 2] ?? panic('root end')).token).end);
        }
        else {
            this.arena.node(root).start.line = 1;
            this.arena.node(root).start.column = 1;
            this.arena.node(root).end.line = 1;
            this.arena.node(root).end.column = 1;
        }
        return this.arena;
    }
}
