// Port of cohere/internal/format/yaml/cst/parser.go and yaml 2.9.0's parser.
import { panic } from 'adamic';
import { Token, tokenType } from './cst.ts';
import { CollectionItem } from './collectionItem.ts';
import { Lexer } from './lexer.ts';
function pair(start: number[], key: number, sep: number[]): CollectionItem {
    const item = new CollectionItem(start);
    item.key = key;
    item.keyPresent = true;
    item.sep = sep;
    item.sepPresent = true;
    return item;
}
export class CSTParser {
    /** @mutates target append source tokens into the owning list */
    append(target: number[], source: readonly number[]): void {
        for(const value of source) target.push(value);
    }

    nodes: Token[] = [];
    roots: number[] = [];
    lineStarts: number[] = [];
    atNewLine = true;
    atScalar = false;
    indent = 0;
    offset = 0;
    onKeyLine = false;
    stack: number[] = [];
    source = '';
    type = '';
    lexer = new Lexer();
    get(index: number): Token {
        return this.nodes[index] ?? panic(`missing CST token ${index}`);
    }
    peek(count: number): number {
        return this.stack[this.stack.length - count] ?? -1;
    }
    last(token: Token): CollectionItem {
        return token.items[token.items.length - 1] ?? panic('missing CST item');
    }
    make(type: string, offset: number, indent: number, source: string, indentPresent: boolean): number {
        const index = this.nodes.length;
        this.nodes.push(new Token(type, offset, indent, source, indentPresent));
        return index;
    }
    sourceToken(): number {
        return this.make(this.type, this.offset, this.indent, this.source, true);
    }
    error(message: string): number {
        const index = this.make('error', this.offset, 0, this.source, false);
        this.get(index).message = message;
        return index;
    }
    includes(list: readonly number[], type: string): boolean {
        for(const index of list) if(this.get(index).type === type) return true;
        return false;
    }
    nonEmpty(list: readonly number[]): number {
        for(let index = 0; index < list.length; index++) {
            const type = this.get(list[index] ?? panic('missing token')).type;
            if(type !== 'space' && type !== 'comment' && type !== 'newline') return index;
        }
        return -1;
    }
    flowToken(index: number): boolean {
        if(index < 0) return false;
        const type = this.get(index).type;
        return (
            type === 'alias' ||
            type === 'scalar' ||
            type === 'single-quoted-scalar' ||
            type === 'double-quoted-scalar' ||
            type === 'flow-collection'
        );
    }
    previousProps(parent: Token): number[] {
        if(parent.type === 'document') return parent.start;
        if(parent.type === 'block-map') {
            const item = this.last(parent);
            return item.sepPresent ? item.sep : item.start;
        }
        if(parent.type === 'block-seq') return this.last(parent).start;
        return [];
    }
    /** @mutates previous transfer the final start properties out of the owning list */
    firstKeyProps(previous: number[]): number[] {
        let index = previous.length - 1;
        while(index >= 0) {
            const type = this.get(previous[index] ?? panic('missing previous token')).type;
            if(
                type === 'doc-start' ||
                type === 'explicit-key-ind' ||
                type === 'map-value-ind' ||
                type === 'seq-item-ind' ||
                type === 'newline'
            )
                break;
            index--;
        }
        index++;
        while(index < previous.length && this.get(previous[index] ?? panic('missing previous token')).type === 'space')
            index++;
        return previous.splice(index, previous.length);
    }
    /** @mutates token normalize the parser-owned flow sequence before attaching it */
    fixFlowSequence(token: Token): void {
        if(this.get(token.flowStart).type !== 'flow-seq-start') return;
        for(const item of token.items) {
            if(
                !item.sepPresent ||
                item.value >= 0 ||
                this.includes(item.start, 'explicit-key-ind') ||
                this.includes(item.sep, 'map-value-ind')
            )
                continue;
            if(item.key >= 0) item.value = item.key;
            item.key = -1;
            item.keyPresent = false;
            if(this.flowToken(item.value)) {
                const value = this.get(item.value);
                if(value.endPresent) this.append(value.end, item.sep);
                else {
                    value.end = item.sep;
                    value.endPresent = true;
                }
            }
            else this.append(item.start, item.sep);
            item.sep = [];
            item.sepPresent = false;
        }
    }
    commentLessIndented(list: readonly number[], indent: number): boolean {
        for(const index of list) {
            const token = this.get(index);
            if(token.type === 'comment' && token.indent >= indent) return false;
        }
        return true;
    }
    pop(error: number): void {
        let index = error;
        if(index < 0) index = this.stack.pop() ?? -1;
        if(index < 0) {
            this.roots.push(this.error('Tried to pop an empty stack'));
            return;
        }
        if(this.stack.length === 0) {
            this.roots.push(index);
            return;
        }
        const token = this.get(index);
        const top = this.get(this.peek(1));
        if(token.type === 'block-scalar') token.indent = top.indentPresent ? top.indent : 0;
        else if(token.type === 'flow-collection' && top.type === 'document') token.indent = 0;
        if(token.type === 'flow-collection') this.fixFlowSequence(token);
        switch(top.type) {
            case 'document':
                top.value = index;
                break;
            case 'block-scalar':
                top.props.push(index);
                break;
            case 'block-map': {
                const item = this.last(top);
                if(item.value >= 0) {
                    top.items.push(pair([], index, []));
                    this.onKeyLine = true;
                    return;
                }
                if(item.sepPresent) item.value = index;
                else {
                    item.key = index;
                    item.keyPresent = true;
                    item.sep = [];
                    item.sepPresent = true;
                    this.onKeyLine = !item.explicitKey;
                    return;
                }
                break;
            }
            case 'block-seq': {
                const item = this.last(top);
                if(item.value >= 0) {
                    const next = new CollectionItem([]);
                    next.value = index;
                    top.items.push(next);
                }
                else item.value = index;
                break;
            }
            case 'flow-collection': {
                const item = top.items[top.items.length - 1];
                if(item === undefined || item.value >= 0) top.items.push(pair([], index, []));
                else if(item.sepPresent) item.value = index;
                else {
                    item.key = index;
                    item.keyPresent = true;
                    item.sep = [];
                    item.sepPresent = true;
                }
                return;
            }
            default:
                this.pop(-1);
                this.pop(index);
        }
        if(
            (top.type === 'document' || top.type === 'block-map' || top.type === 'block-seq') &&
            (token.type === 'block-map' || token.type === 'block-seq')
        ) {
            const last = token.items[token.items.length - 1];
            if(
                last !== undefined &&
                !last.sepPresent &&
                last.value < 0 &&
                last.start.length > 0 &&
                this.nonEmpty(last.start) < 0 &&
                (token.indent === 0 || this.commentLessIndented(last.start, token.indent))
            ) {
                if(top.type === 'document') {
                    top.end = last.start;
                    top.endPresent = true;
                }
                else top.items.push(new CollectionItem(last.start));
                token.items.pop();
            }
        }
    }
    newLines(): void {
        let newline = this.source.indexOf('\n') + 1;
        while(newline !== 0) {
            this.lineStarts.push(this.offset + newline);
            newline = this.source.indexOf('\n', newline) + 1;
        }
    }
    flowScalar(type: string): number {
        this.newLines();
        return this.make(type, this.offset, this.indent, this.source, true);
    }
    map(start: number[], key: number, sep: number[], offset: number, indent: number): number {
        const index = this.make('block-map', offset, indent, '', true);
        this.get(index).items.push(pair(start, key, sep));
        return index;
    }
    startBlockValue(parent: Token): number {
        switch(this.type) {
            case 'alias':
            case 'scalar':
            case 'single-quoted-scalar':
            case 'double-quoted-scalar':
                return this.flowScalar(this.type);
            case 'block-scalar-header': {
                const index = this.make('block-scalar', this.offset, this.indent, '', true);
                this.get(index).props.push(this.sourceToken());
                return index;
            }
            case 'flow-map-start':
            case 'flow-seq-start': {
                const index = this.make('flow-collection', this.offset, this.indent, '', true);
                this.get(index).flowStart = this.sourceToken();
                this.get(index).endPresent = true;
                return index;
            }
            case 'seq-item-ind': {
                const index = this.make('block-seq', this.offset, this.indent, '', true);
                this.get(index).items.push(new CollectionItem([this.sourceToken()]));
                return index;
            }
            case 'explicit-key-ind': {
                this.onKeyLine = true;
                const start = this.firstKeyProps(this.previousProps(parent));
                start.push(this.sourceToken());
                const index = this.make('block-map', this.offset, this.indent, '', true);
                const item = new CollectionItem(start);
                item.explicitKey = true;
                this.get(index).items.push(item);
                return index;
            }
            case 'map-value-ind': {
                this.onKeyLine = true;
                const start = this.firstKeyProps(this.previousProps(parent));
                return this.map(start, -1, [this.sourceToken()], this.offset, this.indent);
            }
        }
        return -1;
    }
    stream(): void {
        switch(this.type) {
            case 'directive-line':
                this.roots.push(this.make('directive', this.offset, 0, this.source, false));
                return;
            case 'byte-order-mark':
            case 'space':
            case 'comment':
            case 'newline':
                this.roots.push(this.sourceToken());
                return;
            case 'doc-mode':
            case 'doc-start': {
                const index = this.make('document', this.offset, 0, '', false);
                if(this.type === 'doc-start') this.get(index).start.push(this.sourceToken());
                this.stack.push(index);
                return;
            }
        }
        this.roots.push(this.error(`Unexpected ${this.type} token in YAML stream`));
    }
    /** @mutates token extend the document being built by this parser */
    document(token: Token): void {
        if(token.value >= 0) {
            this.lineEnd(token);
            return;
        }
        switch(this.type) {
            case 'doc-start':
                if(this.nonEmpty(token.start) >= 0) {
                    this.pop(-1);
                    this.step();
                }
                else token.start.push(this.sourceToken());
                return;
            case 'anchor':
            case 'tag':
            case 'space':
            case 'comment':
            case 'newline':
                token.start.push(this.sourceToken());
                return;
        }
        const value = this.startBlockValue(token);
        if(value >= 0) this.stack.push(value);
        else this.roots.push(this.error(`Unexpected ${this.type} token in YAML document`));
    }
    /** @mutates token transfer the scalar end properties to the new mapping */
    scalar(token: Token): void {
        if(this.type !== 'map-value-ind') {
            this.lineEnd(token);
            return;
        }
        const start = this.firstKeyProps(this.previousProps(this.get(this.peek(2))));
        let sep: number[] = [];
        if(token.endPresent) sep = token.end;
        sep.push(this.sourceToken());
        token.end = [];
        token.endPresent = false;
        this.onKeyLine = true;
        this.stack[this.stack.length - 1] = this.map(start, this.peek(1), sep, token.offset, token.indent);
    }
    /** @mutates token extend the parser-owned block scalar */
    blockScalar(token: Token): void {
        switch(this.type) {
            case 'space':
            case 'comment':
            case 'newline':
                token.props.push(this.sourceToken());
                return;
            case 'scalar':
                token.source = this.source;
                this.atNewLine = true;
                this.indent = 0;
                this.newLines();
                this.pop(-1);
                return;
            default:
                this.pop(-1);
                this.step();
        }
    }
    indentedComment(start: readonly number[], indent: number): boolean {
        if(this.type !== 'comment' || this.indent <= indent) return false;
        for(const index of start) {
            const type = this.get(index).type;
            if(type !== 'newline' && type !== 'space') return false;
        }
        return true;
    }
    /** @mutates collection append a new collection item when the value ended */
    /** @mutates item append source properties to the current item */
    appendItemEnd(collection: Token, item: CollectionItem): void {
        if(item.value < 0) {
            (item.sepPresent ? item.sep : item.start).push(this.sourceToken());
            return;
        }
        const value = this.get(item.value);
        const last = value.end[value.end.length - 1] ?? -1;
        if(last >= 0 && this.get(last).type === 'comment') value.end.push(this.sourceToken());
        else collection.items.push(new CollectionItem([this.sourceToken()]));
    }
    /** @mutates collection move its last item comments to the previous value */
    moveIndentedComment(collection: Token, item: CollectionItem): boolean {
        if(!this.indentedComment(item.start, collection.indent)) return false;
        const previous = collection.items[collection.items.length - 2];
        if(previous === undefined || previous.value < 0) return false;
        const value = this.get(previous.value);
        if(!value.endPresent) return false;
        this.append(value.end, item.start);
        value.end.push(this.sourceToken());
        collection.items.pop();
        return true;
    }
    /** @mutates token extend the parser-owned mapping */
    blockMap(token: Token): void {
        const item = this.last(token);
        switch(this.type) {
            case 'newline':
                this.onKeyLine = false;
                this.appendItemEnd(token, item);
                return;
            case 'space':
            case 'comment':
                if(item.value >= 0) token.items.push(new CollectionItem([this.sourceToken()]));
                else if(item.sepPresent) item.sep.push(this.sourceToken());
                else if(!this.moveIndentedComment(token, item)) item.start.push(this.sourceToken());
                return;
        }
        if(this.indent < token.indent) {
            this.pop(-1);
            this.step();
            return;
        }
        const atMapIndent = !this.onKeyLine && this.indent === token.indent;
        const atNextItem = atMapIndent && (item.sepPresent || item.explicitKey) && this.type !== 'seq-item-ind';
        let start: number[] = [];
        if(atNextItem && item.sepPresent && item.value < 0) {
            let newlines: number[] = [];
            for(let index = 0; index < item.sep.length; index++) {
                const source = this.get(item.sep[index] ?? panic('missing separator'));
                switch(source.type) {
                    case 'newline':
                        newlines.push(index);
                        break;
                    case 'space':
                        break;
                    case 'comment':
                        if(source.indent > token.indent) newlines = [];
                        break;
                    default:
                        newlines = [];
                }
            }
            if(newlines.length >= 2)
                start = item.sep.splice(newlines[1] ?? panic('missing second newline'), item.sep.length);
        }
        switch(this.type) {
            case 'anchor':
            case 'tag':
                if(atNextItem || item.value >= 0) {
                    start.push(this.sourceToken());
                    token.items.push(new CollectionItem(start));
                    this.onKeyLine = true;
                }
                else (item.sepPresent ? item.sep : item.start).push(this.sourceToken());
                return;
            case 'explicit-key-ind': {
                if(!item.sepPresent && !item.explicitKey) {
                    item.start.push(this.sourceToken());
                    item.explicitKey = true;
                }
                else if(atNextItem || item.value >= 0) {
                    start.push(this.sourceToken());
                    const next = new CollectionItem(start);
                    next.explicitKey = true;
                    token.items.push(next);
                }
                else {
                    const index = this.make('block-map', this.offset, this.indent, '', true);
                    const next = new CollectionItem([this.sourceToken()]);
                    next.explicitKey = true;
                    this.get(index).items.push(next);
                    this.stack.push(index);
                }
                this.onKeyLine = true;
                return;
            }
            case 'map-value-ind': {
                if(item.explicitKey) {
                    if(!item.sepPresent) {
                        if(this.includes(item.start, 'newline')) {
                            item.key = -1;
                            item.keyPresent = true;
                            item.sep = [this.sourceToken()];
                            item.sepPresent = true;
                        }
                        else
                            this.stack.push(
                                this.map(
                                    this.firstKeyProps(item.start),
                                    -1,
                                    [this.sourceToken()],
                                    this.offset,
                                    this.indent,
                                ),
                            );
                    }
                    else if(item.value >= 0) token.items.push(pair([], -1, [this.sourceToken()]));
                    else if(this.includes(item.sep, 'map-value-ind'))
                        this.stack.push(this.map(start, -1, [this.sourceToken()], this.offset, this.indent));
                    else if(this.flowToken(item.key) && !this.includes(item.sep, 'newline')) {
                        const keyStart = this.firstKeyProps(item.start);
                        const key = item.key;
                        const sep = item.sep;
                        sep.push(this.sourceToken());
                        item.key = -1;
                        item.keyPresent = false;
                        item.sep = [];
                        item.sepPresent = false;
                        this.stack.push(this.map(keyStart, key, sep, this.offset, this.indent));
                    }
                    else if(start.length > 0) {
                        item.sep = item.sep.concat(start);
                        item.sep.push(this.sourceToken());
                    }
                    else item.sep.push(this.sourceToken());
                }
                else if(!item.sepPresent) {
                    item.key = -1;
                    item.keyPresent = true;
                    item.sep = [this.sourceToken()];
                    item.sepPresent = true;
                }
                else if(item.value >= 0 || atNextItem) token.items.push(pair(start, -1, [this.sourceToken()]));
                else if(this.includes(item.sep, 'map-value-ind'))
                    this.stack.push(this.map([], -1, [this.sourceToken()], this.offset, this.indent));
                else item.sep.push(this.sourceToken());
                this.onKeyLine = true;
                return;
            }
            case 'alias':
            case 'scalar':
            case 'single-quoted-scalar':
            case 'double-quoted-scalar': {
                const scalar = this.flowScalar(this.type);
                if(atNextItem || item.value >= 0) {
                    token.items.push(pair(start, scalar, []));
                    this.onKeyLine = true;
                }
                else if(item.sepPresent) this.stack.push(scalar);
                else {
                    item.key = scalar;
                    item.keyPresent = true;
                    item.sep = [];
                    item.sepPresent = true;
                    this.onKeyLine = true;
                }
                return;
            }
            default: {
                const value = this.startBlockValue(token);
                if(value >= 0) {
                    if(this.get(value).type === 'block-seq') {
                        if(!item.explicitKey && item.sepPresent && !this.includes(item.sep, 'newline')) {
                            this.pop(this.error('Unexpected block-seq-ind on same line with key'));
                            return;
                        }
                    }
                    else if(atMapIndent) token.items.push(new CollectionItem(start));
                    this.stack.push(value);
                    return;
                }
            }
        }
        this.pop(-1);
        this.step();
    }
    /** @mutates token extend the parser-owned sequence */
    blockSequence(token: Token): void {
        const item = this.last(token);
        switch(this.type) {
            case 'newline':
                if(item.value >= 0) this.appendItemEnd(token, item);
                else item.start.push(this.sourceToken());
                return;
            case 'space':
            case 'comment':
                if(item.value >= 0) token.items.push(new CollectionItem([this.sourceToken()]));
                else if(!this.moveIndentedComment(token, item)) item.start.push(this.sourceToken());
                return;
            case 'anchor':
            case 'tag':
                if(item.value < 0 && this.indent > token.indent) {
                    item.start.push(this.sourceToken());
                    return;
                }
                break;
            case 'seq-item-ind':
                if(this.indent !== token.indent) break;
                if(item.value >= 0 || this.includes(item.start, 'seq-item-ind'))
                    token.items.push(new CollectionItem([this.sourceToken()]));
                else item.start.push(this.sourceToken());
                return;
        }
        if(this.indent > token.indent) {
            const value = this.startBlockValue(token);
            if(value >= 0) {
                this.stack.push(value);
                return;
            }
        }
        this.pop(-1);
        this.step();
    }
    /** @mutates token extend the parser-owned flow collection */
    flowCollection(token: Token): void {
        const item = token.items[token.items.length - 1];
        if(this.type === 'flow-error-end') {
            while(true) {
                this.pop(-1);
                const top = this.peek(1);
                if(top < 0 || this.get(top).type !== 'flow-collection') break;
            }
            return;
        }
        if(token.end.length === 0) {
            switch(this.type) {
                case 'comma':
                case 'explicit-key-ind':
                    if(item === undefined || item.sepPresent)
                        token.items.push(new CollectionItem([this.sourceToken()]));
                    else item.start.push(this.sourceToken());
                    return;
                case 'map-value-ind':
                    if(item === undefined || item.value >= 0) token.items.push(pair([], -1, [this.sourceToken()]));
                    else if(item.sepPresent) item.sep.push(this.sourceToken());
                    else {
                        item.key = -1;
                        item.keyPresent = true;
                        item.sep = [this.sourceToken()];
                        item.sepPresent = true;
                    }
                    return;
                case 'space':
                case 'comment':
                case 'newline':
                case 'anchor':
                case 'tag':
                    if(item === undefined || item.value >= 0)
                        token.items.push(new CollectionItem([this.sourceToken()]));
                    else (item.sepPresent ? item.sep : item.start).push(this.sourceToken());
                    return;
                case 'alias':
                case 'scalar':
                case 'single-quoted-scalar':
                case 'double-quoted-scalar': {
                    const scalar = this.flowScalar(this.type);
                    if(item === undefined || item.value >= 0) token.items.push(pair([], scalar, []));
                    else if(item.sepPresent) this.stack.push(scalar);
                    else {
                        item.key = scalar;
                        item.keyPresent = true;
                        item.sep = [];
                        item.sepPresent = true;
                    }
                    return;
                }
                case 'flow-map-end':
                case 'flow-seq-end':
                    token.end.push(this.sourceToken());
                    return;
            }
            const value = this.startBlockValue(token);
            if(value >= 0) this.stack.push(value);
            else {
                this.pop(-1);
                this.step();
            }
            return;
        }
        const parent = this.get(this.peek(2));
        if(
            parent.type === 'block-map' &&
            ((this.type === 'map-value-ind' && parent.indent === token.indent) ||
                (this.type === 'newline' && !this.last(parent).sepPresent))
        ) {
            this.pop(-1);
            this.step();
            return;
        }
        if(this.type === 'map-value-ind' && parent.type !== 'flow-collection') {
            const start = this.firstKeyProps(this.previousProps(parent));
            this.fixFlowSequence(token);
            const sep = token.end.splice(1, token.end.length);
            sep.push(this.sourceToken());
            this.onKeyLine = true;
            this.stack[this.stack.length - 1] = this.map(start, this.peek(1), sep, token.offset, token.indent);
            return;
        }
        this.lineEnd(token);
    }
    /** @mutates token append the document end properties */
    documentEnd(token: Token): void {
        if(this.type === 'doc-mode') return;
        token.endPresent = true;
        token.end.push(this.sourceToken());
        if(this.type === 'newline') this.pop(-1);
    }
    /** @mutates token append trailing source tokens */
    lineEnd(token: Token): void {
        switch(this.type) {
            case 'comma':
            case 'doc-start':
            case 'doc-end':
            case 'flow-seq-end':
            case 'flow-map-end':
            case 'map-value-ind':
                this.pop(-1);
                this.step();
                return;
        }
        if(this.type === 'newline') this.onKeyLine = false;
        token.endPresent = true;
        token.end.push(this.sourceToken());
        if(this.type === 'newline') this.pop(-1);
    }
    step(): void {
        const index = this.peek(1);
        if(this.type === 'doc-end' && (index < 0 || this.get(index).type !== 'doc-end')) {
            while(this.stack.length > 0) this.pop(-1);
            this.stack.push(this.make('doc-end', this.offset, 0, this.source, false));
            return;
        }
        if(index < 0) {
            this.stream();
            return;
        }
        const top = this.get(index);
        switch(top.type) {
            case 'document':
                this.document(top);
                return;
            case 'alias':
            case 'scalar':
            case 'single-quoted-scalar':
            case 'double-quoted-scalar':
                this.scalar(top);
                return;
            case 'block-scalar':
                this.blockScalar(top);
                return;
            case 'block-map':
                this.blockMap(top);
                return;
            case 'block-seq':
                this.blockSequence(top);
                return;
            case 'flow-collection':
                this.flowCollection(top);
                return;
            case 'doc-end':
                this.documentEnd(top);
                return;
        }
        this.pop(-1);
    }
    next(source: string): void {
        this.source = source;
        if(this.atScalar) {
            this.atScalar = false;
            this.step();
            this.offset += source.length;
            return;
        }
        const type = tokenType(source);
        if(type === '') {
            this.pop(this.error(`Not a YAML token: ${source}`));
            this.offset += source.length;
            return;
        }
        if(type === 'scalar') {
            this.atNewLine = false;
            this.atScalar = true;
            this.type = 'scalar';
            return;
        }
        this.type = type;
        this.step();
        switch(type) {
            case 'newline':
                this.atNewLine = true;
                this.indent = 0;
                this.lineStarts.push(this.offset + source.length);
                break;
            case 'space':
                if(this.atNewLine && source.slice(0, 1) === ' ') this.indent += source.length;
                break;
            case 'explicit-key-ind':
            case 'map-value-ind':
            case 'seq-item-ind':
                if(this.atNewLine) this.indent += source.length;
                break;
            case 'doc-mode':
            case 'flow-error-end':
                return;
            default:
                this.atNewLine = false;
        }
        this.offset += source.length;
    }
    parse(source: string, incomplete = false): number[] {
        const start = this.roots.length;
        if(this.offset === 0) this.lineStarts.push(0);
        for(const lexeme of this.lexer.lex(source, incomplete)) this.next(lexeme);
        if(!incomplete) while(this.stack.length > 0) this.pop(-1);
        return this.roots.slice(start);
    }
}
