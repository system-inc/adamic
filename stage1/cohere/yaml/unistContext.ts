// yaml-unist-parser 3.2.0's transforms. Arena indices preserve the identity of shared points and positions.
import type { Composer } from './composer.ts';
import type { CSTParser } from './cstParser.ts';
import { Token } from './cst.ts';
import { ComposedNode } from './composedNode.ts';
import { UnistNode, hasChildrenField, hasLeadingCommentsField, hasTrailingCommentField } from './unistNode.ts';
import { UnistDocumentData } from './unistDocumentData.ts';
import { UnistPoint } from './unistPoint.ts';
import { UnistPosition } from './unistPosition.ts';

export class UnistContext {
    nodes: UnistNode[] = [];
    points: UnistPoint[] = [];
    positions: UnistPosition[] = [];
    comments: number[] = [];
    errorName = '';
    errorMessage = '';
    text: string;
    parser: CSTParser;
    composer: Composer;
    constructor(text: string, parser: CSTParser, composer: Composer) {
        this.text = text;
        this.parser = parser;
        this.composer = composer;
    }
    token(index: number): Token {
        return index < 0 ? new Token('', 0, 0, '', false) : this.parser.get(index);
    }
    composed(index: number): ComposedNode {
        return index < 0 ? new ComposedNode('', '') : this.composer.node(index);
    }
    node(index: number): UnistNode {
        return this.nodes[index] ?? new UnistNode('', -1);
    }
    point(index: number): UnistPoint {
        return this.points[index] ?? new UnistPoint(0, 0, 0);
    }
    position(index: number): UnistPosition {
        return this.positions[this.node(index).position] ?? new UnistPosition(-1, -1);
    }
    start(index: number): UnistPoint {
        return this.point(this.position(index).startPoint);
    }
    end(index: number): UnistPoint {
        return this.point(this.position(index).endPoint);
    }
    fail(message: string): void {
        if(this.errorName !== '') return;
        this.errorName = 'Error';
        this.errorMessage = message;
    }
    typeError(receiver: string, field: string): void {
        if(this.errorName !== '') return;
        this.errorName = 'TypeError';
        this.errorMessage = `Cannot read properties of ${receiver} (reading '${field}')`;
    }
    offset(offset: number): number {
        let line = 0;
        for(let index = 0; index < this.parser.lineStarts.length; index++) {
            if((this.parser.lineStarts[index] ?? 0) > offset) break;
            line = index + 1;
        }
        const column = line === 0 ? offset : offset - (this.parser.lineStarts[line - 1] ?? 0) + 1;
        const result = this.points.length;
        this.points.push(new UnistPoint(line, column, offset));
        return result;
    }
    makePosition(start: number, end: number): number {
        const result = this.positions.length;
        this.positions.push(new UnistPosition(start, end));
        return result;
    }
    range(start: number, end: number): number {
        return this.makePosition(this.offset(start), this.offset(end));
    }
    make(type: string, position: number): number {
        const result = this.nodes.length;
        this.nodes.push(new UnistNode(type, position));
        return result;
    }
    tokens(list: readonly number[]): number[] {
        const result: number[] = [];
        for(const index of list) {
            const type = this.token(index).type;
            if(type !== 'space' && type !== 'newline') result.push(index);
        }
        return result;
    }
    property(index: number): boolean {
        return ['comment', 'tag', 'anchor'].includes(this.token(index).type);
    }
    comment(index: number): number {
        const token = this.token(index);
        const result = this.make('comment', this.range(token.offset, token.offset + token.source.length));
        this.node(result).value = token.source.slice(1);
        this.comments.push(result);
        return result;
    }
    extractComments(list: readonly number[]): number[] {
        const rest: number[] = [];
        for(const index of this.tokens(list)) {
            if(this.token(index).type === 'comment') this.comment(index);
            else rest.push(index);
        }
        return rest;
    }
    /** @mutates target Content properties are assigned to this newly allocated syntax node. */
    content(target: UnistNode, composed: number, props: readonly number[]): void {
        let first = -1;
        for(const index of props) {
            const token = this.token(index);
            const node = this.composed(composed);
            if(token.type === 'tag' || token.type === 'anchor') {
                if(first < 0) first = token.offset;
                if(composed < 0) {
                    this.typeError('null', token.type);
                    return;
                }
                const property = this.make(token.type, this.range(token.offset, token.offset + token.source.length));
                if(token.type === 'anchor') {
                    this.node(property).value = node.anchor;
                    target.anchor = property;
                }
                else {
                    let value = node.tag;
                    if(value === '') value = token.source.slice(token.source.startsWith('!!') ? 2 : 1);
                    if(value === '!') value = 'tag:yaml.org,2002:str';
                    this.node(property).value = value;
                    target.tag = property;
                }
            }
            else if(token.type === 'comment') {
                const comment = this.comment(index);
                if(first >= 0 && first <= token.offset) {
                    if(composed < 0) {
                        this.typeError('null', 'range');
                        return;
                    }
                    if(node.range.length === 0) {
                        this.typeError('undefined', '0');
                        return;
                    }
                    if(token.offset + token.source.length <= (node.range[0] ?? 0)) target.middleComments.push(comment);
                }
            }
            else this.fail(`Unexpected content property token type: ${token.type}`);
        }
    }
    empty(index: number, props: readonly number[]): boolean {
        if(index < 0) return true;
        const node = this.composed(index);
        if(node.range.length === 0) {
            this.typeError('undefined', '0');
            return true;
        }
        if(node.range[0] !== node.range[1]) return false;
        for(const token of props) if(this.token(token).type !== 'comment') return false;
        return true;
    }
    transform(index: number, props: readonly number[]): number {
        if(index < 0 || this.errorName !== '') return -1;
        const node = this.composed(index);
        if(node.kind === 'Map' || node.kind === 'Seq') return this.collection(index, props);
        let type: string;
        if(node.kind === 'Alias') type = 'alias';
        else if(node.kind === 'Scalar') {
            if(node.type === 'PLAIN') type = 'plain';
            else if(node.type === 'QUOTE_DOUBLE') type = 'quoteDouble';
            else if(node.type === 'QUOTE_SINGLE') type = 'quoteSingle';
            else if(node.type === 'BLOCK_FOLDED') type = 'blockFolded';
            else if(node.type === 'BLOCK_LITERAL') type = 'blockLiteral';
            else {
                this.fail(`Unexpected scalar type: ${node.type === '' ? 'undefined' : node.type}`);
                return -1;
            }
        }
        else {
            this.fail('Unexpected unknown node type');
            return -1;
        }
        let start = node.range[0] ?? 0;
        let end = node.range[1] ?? 0;
        const source = this.token(node.sourceToken);
        const block = type === 'blockFolded' || type === 'blockLiteral';
        if(type === 'plain' && start === end) {
            let last = start - 1;
            while(last >= 0 && last < this.text.length && this.text.slice(last, last + 1).trim() === '') last--;
            start = last + 1;
            end = start;
        }
        else if(type === 'alias' && node.sourceToken < 0) {
            this.typeError('undefined', 'end');
            return -1;
        }
        else if(block && source.type !== 'block-scalar') {
            this.fail('Expected block scalar srcToken');
            return -1;
        }
        else if(type === 'plain' && source.type !== 'scalar') {
            this.fail('Expected plain scalar srcToken');
            return -1;
        }
        else if(type === 'quoteDouble' && source.type !== 'double-quoted-scalar') {
            this.fail('Expected double-quoted scalar srcToken');
            return -1;
        }
        else if(type === 'quoteSingle' && source.type !== 'single-quoted-scalar') {
            this.fail('Expected single-quoted scalar srcToken');
            return -1;
        }
        else if(!block) {
            for(const token of this.extractComments(source.end)) {
                const context = type === 'alias' ? 'alias' : type === 'plain' ? 'plain scalar' : 'quote value';
                this.fail(`Unexpected token type in ${context} end: ${this.token(token).type}`);
            }
        }
        const result = this.make(type, this.range(start, end));
        this.node(result).value = type === 'plain' && start === end ? '' : node.source;
        if(block) {
            let header = -1;
            for(const token of this.tokens(source.props)) {
                if(this.token(token).type === 'comment') this.node(result).indicatorComment = this.comment(token);
                else if(this.token(token).type === 'block-scalar-header') header = token;
                else this.fail(`Unexpected token type in block value end: ${this.token(token).type}`);
            }
            if(header < 0) this.fail('Expected block scalar header token');
            this.node(result).chomping = 'clip';
            for(const character of this.token(header).source.slice(1).split('')) {
                if(character === '+') this.node(result).chomping = 'keep';
                else if(character === '-') this.node(result).chomping = 'strip';
                else if(character >= '0' && character <= '9') this.node(result).indent = Number(character);
            }
        }
        this.content(this.node(result), index, props);
        return result;
    }
    itemValue(index: number, props: readonly number[]): number {
        const node = this.composed(index);
        if(node.kind !== 'Pair') {
            if(this.empty(index, props)) {
                this.extractComments(props);
                return -1;
            }
            return this.transform(index, props);
        }
        const item = this.pair(index, node.sourceItemCollection, node.sourceItem, false);
        if(item < 0) return -1;
        const result = this.make('mapping', this.node(item).position);
        this.node(result).children.push(item);
        this.content(this.node(result), node.key, props);
        return result;
    }
    collection(index: number, props: readonly number[]): number {
        const node = this.composed(index);
        const map = node.kind === 'Map';
        const source = this.token(node.sourceToken);
        const expected = node.flow ? 'flow-collection' : map ? 'block-map' : 'block-seq';
        if(source.type !== expected) {
            this.fail(
                node.flow
                    ? `Expected flow-collection CST node for flow ${map ? 'map' : 'sequence'}`
                    : `Expected block ${map ? 'mapping' : 'sequence'} srcToken`,
            );
            return -1;
        }
        const children: number[] = [];
        for(let itemIndex = 0; itemIndex < node.items.length; itemIndex++) {
            const itemNodeIndex = node.items[itemIndex] ?? -1;
            const itemNode = this.composed(itemNodeIndex);
            const raw = source.items[itemIndex];
            if(raw === undefined) {
                this.typeError('undefined', 'start');
                return -1;
            }
            if(map || itemNode.kind === 'Pair')
                children.push(this.pair(itemNodeIndex, node.sourceToken, itemIndex, node.flow));
            else if(
                node.flow &&
                itemNode.kind === 'Map' &&
                itemNode.sourceToken < 0 &&
                itemNode.items.length === 1 &&
                this.composed(itemNode.items[0] ?? -1).sourceItemCollection === node.sourceToken &&
                this.composed(itemNode.items[0] ?? -1).sourceItem === itemIndex
            ) {
                children.push(this.pair(itemNode.items[0] ?? -1, node.sourceToken, itemIndex, true));
            }
            else {
                const itemProps: number[] = [];
                let indicator = -1;
                for(const token of this.tokens(raw.start)) {
                    if(this.property(token)) itemProps.push(token);
                    else if(this.token(token).type === (node.flow ? 'comma' : 'seq-item-ind')) indicator = token;
                    else this.fail(`Unexpected token type in sequence item start: ${this.token(token).type}`);
                }
                const value = node.flow
                    ? this.transform(itemNodeIndex, itemProps)
                    : this.itemValue(itemNodeIndex, itemProps);
                if(value < 0 && indicator < 0) {
                    this.typeError('null', 'position');
                    return -1;
                }
                const token = this.token(indicator);
                const start =
                    !node.flow && indicator >= 0 ? this.offset(token.offset) : this.position(value).startPoint;
                const end =
                    value >= 0 ? this.position(value).endPoint : this.offset(token.offset + token.source.length);
                const item = this.make(node.flow ? 'flowSequenceItem' : 'sequenceItem', this.makePosition(start, end));
                if(value >= 0) this.node(item).children.push(value);
                children.push(item);
            }
            if(this.errorName !== '') return -1;
        }
        for(let item = node.items.length; item < source.items.length; item++) {
            for(const token of this.extractComments(source.items[item]?.start ?? [])) {
                if(node.flow && this.token(token).type === 'comma') continue;
                this.fail(`Unexpected token type in collection item start: ${this.token(token).type}`);
            }
        }
        let position: number;
        if(node.flow) {
            let close = -1;
            const closing = map ? 'flow-map-end' : 'flow-seq-end';
            for(const token of this.extractComments(source.end)) {
                if(this.token(token).type === closing) close = token;
                else this.fail(`Unexpected token type in flow ${map ? 'map' : 'seq'} end: ${this.token(token).type}`);
            }
            if(close < 0) {
                this.fail(`Expected ${closing} token`);
                return -1;
            }
            const end = this.token(close);
            position = this.range(this.token(source.flowStart).offset, end.offset + end.source.length);
        }
        else {
            if(children.length === 0) {
                this.typeError('undefined', 'position');
                return -1;
            }
            position = this.makePosition(
                this.position(children[0] ?? -1).startPoint,
                this.position(children[children.length - 1] ?? -1).endPoint,
            );
        }
        const result = this.make(
            node.flow ? (map ? 'flowMapping' : 'flowSequence') : map ? 'mapping' : 'sequence',
            position,
        );
        this.node(result).children = children;
        this.content(this.node(result), index, props);
        return result;
    }
    pair(index: number, collection: number, itemIndex: number, flow: boolean): number {
        const raw = this.token(collection).items[itemIndex];
        if(raw === undefined) {
            this.typeError('undefined', 'start');
            return -1;
        }
        const pair = this.composed(index);
        const keyProps: number[] = [];
        const valueProps: number[] = [];
        let explicit = -1;
        let colon = -1;
        for(const token of this.tokens(raw.start)) {
            const type = this.token(token).type;
            if(this.property(token)) keyProps.push(token);
            else if(type === 'explicit-key-ind') explicit = token;
            else if(type !== 'comma') this.fail(`Unexpected token type in collection item start: ${type}`);
        }
        for(const token of this.tokens(raw.sep)) {
            const type = this.token(token).type;
            if(this.property(token)) valueProps.push(token);
            else if(type === 'map-value-ind') colon = token;
            else this.fail(`Unexpected token type in collection item sep: ${type}`);
        }
        const first = explicit >= 0 ? explicit : raw.key >= 0 ? raw.key : colon >= 0 ? colon : raw.value;
        if(first < 0) {
            this.typeError('undefined', 'offset');
            return -1;
        }
        const keyStart = this.token(first).offset;
        let keyEnd = keyStart;
        if(raw.key >= 0) {
            if(pair.key < 0) {
                this.typeError('null', 'range');
                return -1;
            }
            if(this.composed(pair.key).range.length === 0) {
                this.typeError('undefined', '1');
                return -1;
            }
            keyEnd = this.composed(pair.key).range[1] ?? 0;
        }
        else if(explicit >= 0) keyEnd = this.token(explicit).offset + this.token(explicit).source.length;
        let valueStart = -1;
        if(pair.value >= 0)
            valueStart =
                colon >= 0
                    ? this.token(colon).offset
                    : raw.value >= 0
                      ? this.token(raw.value).offset
                      : (this.composed(pair.value).range[0] ?? 0);
        let keyContent = -1;
        let valueContent = -1;
        if(this.empty(pair.key, keyProps)) this.extractComments(keyProps);
        else keyContent = this.transform(pair.key, keyProps);
        if(this.empty(pair.value, valueProps)) this.extractComments(valueProps);
        else valueContent = this.transform(pair.value, valueProps);
        if(keyContent >= 0) keyEnd = this.end(keyContent).offset;
        const key = this.make('mappingKey', this.range(keyStart, keyEnd));
        if(keyContent >= 0) this.node(key).children.push(keyContent);
        let value: number;
        if(valueContent >= 0 || valueStart >= 0) {
            const start = valueStart >= 0 ? valueStart : this.start(valueContent).offset;
            const end = valueContent >= 0 ? this.end(valueContent).offset : valueStart + 1;
            value = this.make('mappingValue', this.range(start, end));
            if(valueContent >= 0) this.node(value).children.push(valueContent);
        }
        else
            value = this.make(
                'mappingValue',
                this.makePosition(this.position(key).endPoint, this.position(key).endPoint),
            );
        const result = this.make(
            flow ? 'flowMappingItem' : 'mappingItem',
            this.makePosition(this.position(key).startPoint, this.position(value).endPoint),
        );
        this.node(result).children.push(key);
        this.node(result).children.push(value);
        return result;
    }
    pointText(offset: number): string {
        const point = this.point(this.offset(offset));
        return `${point.line}:${point.column}`;
    }
    documents(): number[] {
        const data: UnistDocumentData[] = [];
        let buffer: number[] = [];
        let before: number[] = [];
        let current = -1;
        for(const index of this.tokens(this.parser.roots)) {
            const token = this.token(index);
            if(token.type === 'document') {
                if(data.length >= this.composer.documents.length) {
                    this.fail(`Unexpected 'document' token at ${this.pointText(token.offset)}`);
                    return [];
                }
                const item = new UnistDocumentData();
                item.source = index;
                item.document = data.length;
                item.before = before.concat(buffer);
                data.push(item);
                current = data.length - 1;
                before = [];
                buffer = [];
            }
            else if(token.type === 'comment') buffer.push(index);
            else if(token.type === 'directive') {
                before = before.concat(buffer);
                before.push(index);
                buffer = [];
            }
            else if(token.type === 'doc-end') {
                const item = data[current];
                if(item === undefined || item.end >= 0) {
                    this.fail(`Unexpected 'doc-end' token at ${this.pointText(token.offset)}`);
                    return [];
                }
                item.after = buffer.slice();
                buffer = [];
                item.end = index;
            }
        }
        if(before.length > 0) {
            const first = this.token(before[0] ?? -1);
            this.fail(`Unexpected '${first.type}' token at ${this.pointText(first.offset)}`);
            return [];
        }
        if(buffer.length > 0) {
            if(current < 0) {
                const item = new UnistDocumentData();
                item.document = data.length;
                item.before = before.concat(buffer);
                data.push(item);
                current = data.length - 1;
                buffer = [];
            }
            const item = data[current];
            if(item !== undefined) item.after = item.after.concat(buffer);
        }
        const result: number[] = [];
        for(const item of data) result.push(this.document(item));
        return result;
    }
    document(data: UnistDocumentData): number {
        const document = this.composer.documents[data.document];
        if(document === undefined) {
            this.typeError('undefined', 'contents');
            return -1;
        }
        const directives: number[] = [];
        let candidates: number[] = [];
        let last = -1;
        for(const index of data.before) {
            const token = this.token(index);
            if(token.type === 'comment') {
                const comment = this.comment(index);
                if(
                    last >= 0 &&
                    this.end(last).line === this.start(comment).line &&
                    this.node(last).trailingComment < 0
                ) {
                    this.node(last).trailingComment = comment;
                    this.position(last).endPoint = this.position(comment).endPoint;
                }
                else candidates.push(comment);
            }
            else {
                const directive = this.make('directive', this.range(token.offset, token.offset + token.source.length));
                const words: string[] = [];
                for(const word of token.source.trim().split('\t').join(' ').split(' '))
                    if(word !== '') words.push(word);
                this.node(directive).name = (words[0] ?? '').slice(1);
                this.node(directive).parameters = words.slice(1);
                directives.push(directive);
                last = directive;
                candidates = [];
            }
        }
        let between: number[] = [];
        let marker = -1;
        for(const index of this.tokens(this.token(data.source).start)) {
            between.push(index);
            if(marker < 0 && this.token(index).type === 'doc-start') {
                for(const token of between)
                    if(this.token(token).type === 'comment') candidates.push(this.comment(token));
                between = [];
                marker = index;
            }
        }
        const composed = this.composed(document.contents);
        let headStart =
            marker >= 0
                ? this.token(marker).offset
                : document.contents >= 0
                  ? (composed.range[0] ?? 0)
                  : (document.range[0] ?? 0);
        const headEnd = marker >= 0 ? headStart + this.token(marker).source.length : headStart;
        if(directives.length > 0) headStart = this.start(directives[0] ?? -1).offset;
        const head = this.make('documentHead', this.range(headStart, headEnd));
        this.node(head).children = directives;
        if(marker >= 0) {
            this.node(head).endComments = candidates;
            const first = between[0] ?? -1;
            if(
                first >= 0 &&
                this.token(first).type === 'comment' &&
                this.point(this.offset(this.token(first).offset)).line === this.end(head).line
            ) {
                this.node(head).trailingComment = this.comment(first);
                between = between.slice(1);
            }
        }
        const props: number[] = [];
        for(const index of between) {
            if(this.property(index)) props.push(index);
            else this.fail(`Unexpected token type: ${this.token(index).type}`);
        }
        for(const index of this.extractComments(data.after))
            this.fail(`Unexpected token type: ${this.token(index).type}`);
        const endComments: number[] = [];
        const trailing: number[] = [];
        if(data.source >= 0) {
            for(const index of this.tokens(this.token(data.source).end.concat(this.token(data.end).end))) {
                if(this.token(index).type !== 'comment') {
                    this.fail(`Unexpected token type: ${this.token(index).type}`);
                    continue;
                }
                const comment = this.comment(index);
                if(data.end < 0) endComments.push(comment);
                else {
                    const line = this.point(this.offset(this.token(data.end).offset)).line;
                    if(this.start(comment).line === line) trailing.push(comment);
                    else if(this.start(comment).line < line) endComments.push(comment);
                }
            }
        }
        if(trailing.length > 1)
            this.fail(
                `Unexpected multiple document trailing comments at ${this.start(trailing[1] ?? -1).line}:${this.start(trailing[1] ?? -1).column}`,
            );
        let hasContent = document.contents >= 0 && (composed.range[0] ?? 0) < (composed.range[1] ?? 0);
        if(document.contents >= 0)
            for(const token of props)
                if(this.token(token).type === 'tag' || this.token(token).type === 'anchor') hasContent = true;
        let content = -1;
        if(hasContent) content = this.transform(document.contents, props);
        else
            for(const token of this.extractComments(props))
                this.fail(`Unexpected token type in empty document body: ${this.token(token).type}`);
        let bodyEnd = data.end >= 0 ? Math.max(0, this.token(data.end).offset - 1) : (document.range[2] ?? 0);
        if(data.end < 0) {
            while(bodyEnd < this.text.length && this.text.slice(bodyEnd, bodyEnd + 1).trim() === '') bodyEnd++;
            if(bodyEnd >= this.text.length) bodyEnd = this.text.length;
        }
        if(this.text.slice(bodyEnd - 1, bodyEnd) === '\r') bodyEnd--;
        let bodyStart = content >= 0 ? this.start(content).offset : bodyEnd;
        if(marker >= 0) {
            const markerEnd = this.token(marker).offset + this.token(marker).source.length + 1;
            if(bodyStart < markerEnd && markerEnd <= bodyEnd) bodyStart = markerEnd;
        }
        const body = this.make('documentBody', this.range(bodyStart, bodyEnd));
        this.node(body).endComments = endComments;
        if(content >= 0) this.node(body).children.push(content);
        const end =
            data.end >= 0
                ? this.offset(this.token(data.end).offset + this.token(data.end).source.length)
                : this.position(body).endPoint;
        const result = this.make('document', this.makePosition(this.position(head).startPoint, end));
        this.node(result).children.push(head);
        this.node(result).children.push(body);
        this.node(result).directivesEndMarker = marker >= 0;
        this.node(result).documentEndMarker = data.end >= 0;
        this.node(result).trailingComment = trailing[trailing.length - 1] ?? -1;
        return result;
    }
    parents(index: number, parent: number): void {
        if(index < 0) return;
        const node = this.node(index);
        for(const child of node.children) this.parents(child, index);
        this.parents(node.anchor, index);
        this.parents(node.tag, index);
        for(const child of node.leadingComments) this.parents(child, index);
        for(const child of node.middleComments) this.parents(child, index);
        this.parents(node.indicatorComment, index);
        this.parents(node.trailingComment, index);
        for(const child of node.endComments) this.parents(child, index);
        node.parent = parent;
    }
    explicit(index: number): boolean {
        const position = this.position(index);
        return (
            position.startPoint !== position.endPoint &&
            (this.node(index).children.length === 0 ||
                this.start(index).offset !== this.start(this.node(index).children[0] ?? -1).offset)
        );
    }
    ownsEnd(index: number, comment: number): boolean {
        const node = this.node(index);
        const start = this.start(index);
        const end = this.end(index);
        const commentStart = this.start(comment);
        const commentEnd = this.end(comment);
        if(
            start.offset < commentStart.offset &&
            end.offset > commentEnd.offset &&
            (node.type === 'flowMapping' || node.type === 'flowSequence')
        )
            return (
                node.children.length === 0 ||
                commentStart.line > this.end(node.children[node.children.length - 1] ?? -1).line
            );
        if(commentEnd.offset < end.offset) return false;
        if(node.type === 'sequenceItem') return commentStart.column > start.column;
        if(node.type === 'mappingKey' || node.type === 'mappingValue')
            return (
                commentStart.column > this.start(node.parent).column &&
                (node.children.length === 0 ||
                    (node.children.length === 1 &&
                        !['blockFolded', 'blockLiteral'].includes(this.node(node.children[0] ?? -1).type))) &&
                (node.type === 'mappingValue' || this.explicit(index))
            );
        return false;
    }
    /**
     * @mutates leading The attachment table is populated for each source line.
     * @mutates trailing The attachment table is populated for each source line.
     * @mutates ends The attachment table is populated for each source line.
     */
    table(index: number, leading: number[], trailing: number[], ends: number[]): void {
        const node = this.node(index);
        const start = this.start(index);
        const end = this.end(index);
        if(start.offset === end.offset) return;
        if(hasLeadingCommentsField(node.type)) {
            const previous = leading[start.line - 1] ?? -1;
            if(previous < 0 || start.column < this.start(previous).column) leading[start.line - 1] = index;
        }
        if(
            hasTrailingCommentField(node.type) &&
            end.column > 1 &&
            node.type !== 'document' &&
            node.type !== 'documentHead'
        ) {
            const previous = trailing[end.line - 1] ?? -1;
            if(previous < 0 || end.column >= this.end(previous).column) trailing[end.line - 1] = index;
        }
        if(!['root', 'document', 'documentHead', 'documentBody'].includes(node.type)) {
            const lines = [end.line];
            if(start.line !== end.line) lines.push(start.line);
            for(const line of lines) {
                const previous = ends[line - 1] ?? -1;
                if(previous < 0 || end.column >= this.end(previous).column) ends[line - 1] = index;
            }
        }
        if(hasChildrenField(node.type)) for(const child of node.children) this.table(child, leading, trailing, ends);
    }
    attach(root: number): void {
        this.parents(root, -1);
        const leading: number[] = [];
        const trailing: number[] = [];
        const ends: number[] = [];
        const comments: number[] = [];
        for(let line = 0; line < this.end(root).line; line++) {
            leading.push(-1);
            trailing.push(-1);
            ends.push(-1);
            comments.push(-1);
        }
        for(const comment of this.comments) comments[this.start(comment).line - 1] = comment;
        this.table(root, leading, trailing, ends);
        const unattached: number[] = [];
        for(const comment of this.comments) if(this.node(comment).parent < 0) unattached.push(comment);
        let documentIndex = 0;
        for(const comment of unattached) {
            const documents = this.node(root).children;
            while(
                documentIndex < documents.length - 1 &&
                this.start(comment).line > this.end(documents[documentIndex] ?? -1).line
            )
                documentIndex++;
            this.attachOne(comment, documents[documentIndex] ?? -1, leading, trailing, ends, comments);
        }
    }
    attachOne(
        comment: number,
        document: number,
        leading: readonly number[],
        trailing: readonly number[],
        ends: readonly number[],
        comments: readonly number[],
    ): void {
        const line = this.start(comment).line;
        const target = trailing[line - 1] ?? -1;
        if(target >= 0) {
            if(this.node(target).trailingComment >= 0) {
                this.fail(`Unexpected multiple trailing comment at ${line}:${this.start(comment).column}`);
                return;
            }
            this.parents(comment, target);
            this.node(target).trailingComment = comment;
            return;
        }
        for(let currentLine = line; currentLine >= this.start(document).line; currentLine--) {
            let current = ends[currentLine - 1] ?? -1;
            if(current < 0) {
                const previous = comments[currentLine - 1] ?? -1;
                if(currentLine !== line && previous >= 0) current = this.node(previous).parent;
                else continue;
            }
            if(current < 0) {
                this.typeError('undefined', 'type');
                return;
            }
            if(this.node(current).type === 'sequence' || this.node(current).type === 'mapping')
                current = this.node(current).children[0] ?? -1;
            if(this.node(current).type === 'mappingItem') {
                const key = this.node(current).children[0] ?? -1;
                current = this.explicit(key) ? key : (this.node(current).children[1] ?? -1);
            }
            while(current >= 0) {
                if(this.ownsEnd(current, comment)) {
                    this.parents(comment, current);
                    this.node(current).endComments.push(comment);
                    return;
                }
                current = this.node(current).parent;
            }
            break;
        }
        for(let currentLine = line + 1; currentLine <= this.end(document).line; currentLine++) {
            const next = leading[currentLine - 1] ?? -1;
            if(next >= 0) {
                this.parents(comment, next);
                this.node(next).leadingComments.push(comment);
                return;
            }
        }
        const body = this.node(document).children[1] ?? -1;
        this.parents(comment, body);
        this.node(body).endComments.push(comment);
    }
    extendStart(index: number, point: number): void {
        if(this.point(point).offset < this.start(index).offset) this.position(index).startPoint = point;
    }
    extendEnd(index: number, point: number): void {
        if(this.point(point).offset > this.end(index).offset) this.position(index).endPoint = point;
    }
    update(index: number): void {
        const node = this.node(index);
        if(index < 0 || !hasChildrenField(node.type)) return;
        for(const child of node.children) this.update(child);
        if(node.type === 'document') {
            const head = this.position(node.children[0] ?? -1);
            const body = this.position(node.children[1] ?? -1);
            if(this.point(head.startPoint).offset === this.point(head.endPoint).offset) {
                head.startPoint = body.startPoint;
                head.endPoint = body.startPoint;
            }
            else if(this.point(body.startPoint).offset === this.point(body.endPoint).offset) {
                body.startPoint = head.endPoint;
                body.endPoint = head.endPoint;
            }
        }
        if(node.endComments.length > 0) {
            this.extendStart(index, this.position(node.endComments[0] ?? -1).startPoint);
            this.extendEnd(index, this.position(node.endComments[node.endComments.length - 1] ?? -1).endPoint);
        }
        if(node.children.length > 0) {
            const first = node.children[0] ?? -1;
            const last = node.children[node.children.length - 1] ?? -1;
            this.extendStart(index, this.position(first).startPoint);
            this.extendEnd(index, this.position(last).endPoint);
            if(this.node(first).leadingComments.length > 0)
                this.extendStart(index, this.position(this.node(first).leadingComments[0] ?? -1).startPoint);
            if(this.node(first).tag >= 0) this.extendStart(index, this.position(this.node(first).tag).startPoint);
            if(this.node(first).anchor >= 0) this.extendStart(index, this.position(this.node(first).anchor).startPoint);
            if(this.node(last).trailingComment >= 0)
                this.extendEnd(index, this.position(this.node(last).trailingComment).endPoint);
        }
    }
    parse(): number {
        const children = this.documents();
        if(this.errorName !== '') return -1;
        for(let index = 1; index < this.comments.length; index++) {
            const comment = this.comments[index] ?? -1;
            let target = index;
            while(target > 0 && this.start(this.comments[target - 1] ?? -1).offset > this.start(comment).offset) {
                this.comments[target] = this.comments[target - 1] ?? -1;
                target--;
            }
            this.comments[target] = comment;
        }
        const root = this.make('root', this.range(0, this.text.length));
        this.node(root).children = children;
        this.node(root).comments = this.comments;
        this.attach(root);
        if(this.errorName !== '') return -1;
        this.update(root);
        return root;
    }
}
