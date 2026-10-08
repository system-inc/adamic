// yaml-unist-parser's composer options: strict, source tokens, merge, non-unique keys.
import { panic } from 'adamic';
import { Token } from './cst.ts';
import type { CSTParser } from './cstParser.ts';
import type { Props } from './props.ts';
import { PropsResolver } from './propsResolver.ts';
import { ScalarResolver } from './scalar.ts';
import { ScalarResolution } from './scalarResolution.ts';
import { ComposedNode } from './composedNode.ts';
import { ComposedDocument } from './composedDocument.ts';
import { Directives } from './directives.ts';
import { ComposeError } from './composeError.ts';
import { schemaTags } from './schemaTags.ts';
// Schema patterns are fixed and matching only reads their compiled trees.
const coreTags = schemaTags('1.2');
const legacyTags = schemaTags('1.1');
export class Composer {
    parser: CSTParser;
    nodes: ComposedNode[] = [];
    documents: ComposedDocument[] = [];
    directives = new Directives();
    document = new ComposedDocument(new Directives());
    documentPresent = false;
    atDirectives = false;
    prelude: string[] = [];
    atRoot = true;
    atKey = false;
    scalar: ScalarResolver;
    coreTags = coreTags;
    legacyTags = legacyTags;
    constructor(parser: CSTParser) {
        this.parser = parser;
        this.scalar = new ScalarResolver(parser);
    }
    node(index: number): ComposedNode {
        return this.nodes[index] ?? panic('missing composed YAML node');
    }
    make(kind: string, className: string): number {
        this.nodes.push(new ComposedNode(kind, className));
        return this.nodes.length - 1;
    }
    error(offset: number, code: string, message: string): void {
        this.document.diagnostics.error(offset, code, message);
    }
    tokenError(index: number, code: string, message: string): void {
        const token = this.parser.get(index);
        const length =
            token.type === 'document' ||
            token.type === 'block-map' ||
            token.type === 'block-seq' ||
            token.type === 'flow-collection'
                ? 1
                : token.source.length;
        this.document.diagnostics.errors.push(new ComposeError(token.offset, token.offset + length, code, message));
    }
    rangeError(index: number, code: string, message: string): void {
        const range = this.node(index).range;
        this.document.diagnostics.errors.push(new ComposeError(range[0] ?? 0, range[1] ?? 0, code, message));
    }
    warning(index: number, code: string, message: string): void {
        const token = this.parser.get(index);
        this.document.warnings.push(new ComposeError(token.offset, token.offset + token.source.length, code, message));
    }
    properties(
        tokens: readonly number[],
        flow: string,
        indicator: string,
        next: number,
        offset: number,
        parentIndent: number,
        newline: boolean,
    ): Props {
        const resolver = new PropsResolver(this.parser, this.document.diagnostics);
        const props = resolver.resolve(tokens, flow, indicator, next, offset, parentIndent, newline);
        for(const warning of props.warnings) this.document.warnings.push(warning);
        return props;
    }
    collectDirectives(): void {
        for(const error of this.document.directives.errors) this.document.diagnostics.errors.push(error);
        for(const warning of this.document.directives.warnings) this.document.warnings.push(warning);
        this.document.directives.errors = [];
        this.document.directives.warnings = [];
    }
    tagName(index: number): string {
        if(index < 0) return '';
        const name = this.document.directives.tagName(this.parser.get(index));
        this.collectDirectives();
        return name;
    }
    /** @mutates node Scalar values and formatting metadata are resolved into the newly allocated node. */
    scalarValue(node: ComposedNode, tagName: string, errorSource: number, plain: boolean): void {
        node.valueKind = 'string';
        node.stringValue = node.source;
        const tags = this.document.directives.version === '1.1' ? this.legacyTags : this.coreTags;
        let resolve = '';
        for(const tag of tags) {
            if(tagName !== '') {
                if(tag.tag !== tagName) continue;
            }
            else if(!plain || (tag.defaultKey && !this.atKey)) continue;
            if(tag.pattern.test(node.source)) {
                resolve = tag.resolve;
                node.format = tag.format;
                break;
            }
        }
        if(tagName === '!' || tagName === 'tag:yaml.org,2002:str') return;
        if(tagName !== '' && resolve === '') {
            if(tagName === 'tag:yaml.org,2002:binary') {
                node.valueKind = 'binary';
                return;
            }
            if(tagName === 'tag:yaml.org,2002:merge' && this.document.directives.version === '1.2') {
                node.valueKind = 'symbol';
                return;
            }
            if(tagName === 'tag:yaml.org,2002:timestamp' && this.document.directives.version === '1.2') {
                for(const tag of this.legacyTags)
                    if(tag.resolve === 'timestampTag') {
                        if(tag.pattern.test(node.source)) node.valueKind = 'date';
                        else
                            this.tokenError(
                                errorSource,
                                'TAG_RESOLVE_FAILED',
                                '!!timestamp expects a date, starting with yyyy-mm-dd',
                            );
                    }
                return;
            }
            if(this.prototypeTag(tagName))
                this.tokenError(errorSource, 'TAG_RESOLVE_FAILED', 'tag.resolve is not a function');
            else this.warning(errorSource, 'TAG_RESOLVE_FAILED', `Unresolved tag: ${tagName}`);
            return;
        }
        if(resolve === 'nullTag') node.valueKind = 'null';
        else if(resolve === 'boolTag' || resolve === 'trueTag' || resolve === 'falseTag') {
            node.valueKind = 'boolean';
            node.booleanValue =
                resolve === 'trueTag' || (resolve === 'boolTag' && node.source.slice(0, 1).toLowerCase() === 't');
        }
        else if(resolve === 'mergeTag') node.valueKind = 'symbol';
        else if(resolve === 'timestampTag') node.valueKind = 'date';
        else if(resolve !== '') {
            node.valueKind = 'number';
            let source = node.source;
            if(resolve.startsWith('yaml11') || node.format === 'TIME') source = source.split('_').join('');
            if(resolve.endsWith('NaNTag'))
                node.numberValue =
                    source.slice(-3).toLowerCase() === 'nan' ? NaN : source.startsWith('-') ? -Infinity : Infinity;
            else if(node.format === 'TIME') {
                let sign = 1;
                if(source.startsWith('-')) sign = -1;
                if(source.startsWith('-') || source.startsWith('+')) source = source.slice(1);
                for(const part of source.split(':')) node.numberValue = node.numberValue * 60 + Number(part);
                node.numberValue *= sign;
            }
            else if(node.format === 'BIN' || node.format === 'OCT' || node.format === 'HEX') {
                let sign = 1;
                if(source.startsWith('-')) sign = -1;
                if(source.startsWith('-') || source.startsWith('+')) source = source.slice(1);
                const prefix = resolve === 'yaml11IntOctTag' ? 1 : 2;
                node.numberValue =
                    sign *
                    Number.parseInt(source.slice(prefix), node.format === 'BIN' ? 2 : node.format === 'OCT' ? 8 : 16);
            }
            else if(resolve.includes('Int')) node.numberValue = Number.parseInt(source, 10);
            else {
                node.numberValue = Number.parseFloat(source);
                const dot = node.source.indexOf('.');
                if(dot >= 0 && node.format === '') {
                    const fraction = node.source
                        .slice(dot + 1)
                        .split('_')
                        .join('');
                    if(fraction.endsWith('0')) node.minFractionDigits = fraction.length;
                }
            }
        }
    }
    composeScalar(token: Token, tag: number, sourceIndex: number): number {
        const resolved =
            token.type === 'block-scalar' ? this.scalar.block(token, this.atRoot) : this.scalar.flow(token);
        for(const error of resolved.errors) this.document.diagnostics.errors.push(error);
        const index = this.make('Scalar', 'Scalar');
        const node = this.node(index);
        node.range = resolved.range;
        node.source = resolved.value;
        node.sourcePresent = true;
        node.type = resolved.type;
        node.comment = resolved.comment;
        node.tag = this.tagName(tag);
        this.scalarValue(node, node.tag, tag >= 0 ? tag : sourceIndex, token.type === 'scalar');
        return index;
    }
    empty(offset: number, before: readonly number[], props: Props): number {
        let position = offset;
        for(let index = before.length - 1; index >= 0; index--) {
            const token = this.parser.get(before[index] ?? -1);
            if(token.type === 'space' || token.type === 'comment' || token.type === 'newline')
                position -= token.source.length;
            else {
                index++;
                while(index < before.length && this.parser.get(before[index] ?? -1).type === 'space') {
                    position += this.parser.get(before[index] ?? -1).source.length;
                    index++;
                }
                break;
            }
        }
        const index = this.composeScalar(new Token('scalar', position, -1, '', true), props.tag, -1);
        const node = this.node(index);
        if(props.anchor >= 0) {
            node.anchorPresent = true;
            node.anchor = this.parser.get(props.anchor).source.slice(1);
            if(node.anchor === '') this.tokenError(props.anchor, 'BAD_ALIAS', 'Anchor cannot be an empty string');
        }
        node.spaceBefore = props.spaceBefore;
        if(props.comment !== '') {
            node.comment = props.comment;
            node.range[2] = props.end;
        }
        return index;
    }
    composeNode(index: number, props: Props): number {
        const token = this.parser.get(index);
        let result = -1;
        let sourcePresent = true;
        if(token.type === 'alias') {
            result = this.make('Alias', 'Alias');
            const node = this.node(result);
            node.source = token.source.slice(1);
            node.sourcePresent = true;
            if(node.source === '') this.error(token.offset, 'BAD_ALIAS', 'Alias cannot be an empty string');
            if(token.source.endsWith(':') && token.source.length > 1)
                this.document.warnings.push(
                    new ComposeError(
                        token.offset + token.source.length - 1,
                        token.offset + token.source.length,
                        'BAD_ALIAS',
                        'Alias ending in : is ambiguous',
                    ),
                );
            const resolved = new ScalarResolution();
            const end = token.offset + token.source.length;
            const nodeEnd = this.scalar.end(token.end, end, true, resolved);
            for(const error of resolved.errors) this.document.diagnostics.errors.push(error);
            node.range = [token.offset, end, nodeEnd];
            node.comment = resolved.comment;
            if(props.anchor >= 0 || props.tag >= 0)
                this.tokenError(index, 'ALIAS_PROPS', 'An alias node must not specify any properties');
        }
        else if(
            token.type === 'scalar' ||
            token.type === 'single-quoted-scalar' ||
            token.type === 'double-quoted-scalar' ||
            token.type === 'block-scalar'
        )
            result = this.composeScalar(token, props.tag, index);
        else if(token.type === 'block-map' || token.type === 'block-seq' || token.type === 'flow-collection')
            result = this.collection(index, props);
        else {
            this.tokenError(
                index,
                'UNEXPECTED_TOKEN',
                token.type === 'error' ? token.message : `Unsupported token (type: ${token.type})`,
            );
            sourcePresent = false;
        }
        if(result < 0) return this.empty(token.offset, [], props);
        const node = this.node(result);
        if(token.type !== 'alias' && props.anchor >= 0) {
            node.anchorPresent = true;
            node.anchor = this.parser.get(props.anchor).source.slice(1);
            if(node.anchor === '') this.tokenError(props.anchor, 'BAD_ALIAS', 'Anchor cannot be an empty string');
        }
        node.spaceBefore = props.spaceBefore;
        if(props.comment !== '') {
            if(token.type === 'scalar' && token.source === '') node.comment = props.comment;
            else node.commentBefore = props.comment;
        }
        if(sourcePresent) node.sourceToken = index;
        return result;
    }
    newline(index: number): boolean {
        if(index < 0) return false;
        const token = this.parser.get(index);
        if(
            token.type === 'alias' ||
            token.type === 'scalar' ||
            token.type === 'single-quoted-scalar' ||
            token.type === 'double-quoted-scalar'
        ) {
            if(token.source.includes('\n')) return true;
            for(const end of token.end) if(this.parser.get(end).type === 'newline') return true;
            return false;
        }
        if(token.type === 'flow-collection') {
            for(const item of token.items) {
                for(const start of item.start) if(this.parser.get(start).type === 'newline') return true;
                for(const separator of item.sep) if(this.parser.get(separator).type === 'newline') return true;
                if(this.newline(item.key) || this.newline(item.value)) return true;
            }
            return false;
        }
        return true;
    }
    pair(key: number, value: number, collection: number, item: number): number {
        const index = this.make('Pair', 'Pair');
        this.node(index).key = key;
        this.node(index).value = value;
        this.node(index).sourceItemCollection = collection;
        this.node(index).sourceItem = item;
        return index;
    }
    appendComment(index: number, comment: string): void {
        if(comment === '') return;
        const node = this.node(index);
        node.comment = node.comment === '' ? comment : `${node.comment}\n${comment}`;
    }
    collection(index: number, props: Props): number {
        const token = this.parser.get(index);
        const name = this.tagName(props.tag);
        if(token.type === 'block-seq') {
            let last = props.anchor;
            if(props.tag >= 0 && (last < 0 || this.parser.get(props.tag).offset > this.parser.get(last).offset))
                last = props.tag;
            if(
                last >= 0 &&
                (props.newlineAfterProp < 0 ||
                    this.parser.get(props.newlineAfterProp).offset < this.parser.get(last).offset)
            )
                this.tokenError(last, 'MISSING_CHAR', 'Missing newline after block sequence props');
        }
        const map =
            token.type === 'block-map' ||
            (token.type === 'flow-collection' && this.parser.get(token.flowStart).source === '{');
        const result =
            token.type === 'block-map'
                ? this.blockMap(index)
                : token.type === 'block-seq'
                  ? this.blockSequence(index)
                  : this.flowCollection(index);
        if(result < 0) return -1;
        const node = this.node(result);
        node.tag = name === '!' ? `tag:yaml.org,2002:${map ? 'map' : 'seq'}` : name;
        // Explicit non-generic collection tags are resolved after the CST collection has been composed.
        this.collectionTag(result, props.tag);
        return result;
    }
    blockMap(index: number): number {
        const token = this.parser.get(index);
        const result = this.make('Map', 'YAMLMap');
        this.atRoot = false;
        let offset = token.offset;
        let commentEnd = -1;
        let itemIndex = 0;
        for(const item of token.items) {
            const properties = this.properties(
                item.start,
                '',
                'explicit-key-ind',
                item.key >= 0 ? item.key : (item.sep[0] ?? -1),
                offset,
                token.indent,
                true,
            );
            const implicit = properties.found < 0;
            if(implicit) {
                if(item.key >= 0) {
                    const key = this.parser.get(item.key);
                    if(key.type === 'block-seq')
                        this.error(
                            offset,
                            'BLOCK_AS_IMPLICIT_KEY',
                            'A block sequence may not be used as an implicit map key',
                        );
                    else if(key.indentPresent && key.indent !== token.indent)
                        this.error(offset, 'BAD_INDENT', 'All mapping items must start at the same column');
                }
                if(properties.anchor < 0 && properties.tag < 0 && !item.sepPresent) {
                    commentEnd = properties.end;
                    this.appendComment(result, properties.comment);
                    itemIndex++;
                    continue;
                }
                if(properties.newlineAfterProp >= 0 || this.newline(item.key))
                    this.tokenError(
                        item.key >= 0 ? item.key : (item.start[item.start.length - 1] ?? -1),
                        'MULTILINE_IMPLICIT_KEY',
                        'Implicit keys need to be on a single line',
                    );
            }
            else if(this.parser.get(properties.found).indent !== token.indent)
                this.error(offset, 'BAD_INDENT', 'All mapping items must start at the same column');
            this.atKey = true;
            const key =
                item.key >= 0
                    ? this.composeNode(item.key, properties)
                    : this.empty(properties.end, item.start, properties);
            this.atKey = false;
            const valueProps = this.properties(
                item.sep,
                '',
                'map-value-ind',
                item.value,
                this.node(key).range[2] ?? 0,
                token.indent,
                item.key < 0 || this.parser.get(item.key).type === 'block-scalar',
            );
            offset = valueProps.end;
            let value = -1;
            if(valueProps.found >= 0) {
                if(implicit) {
                    if(item.value >= 0 && this.parser.get(item.value).type === 'block-map' && !valueProps.hasNewline)
                        this.error(
                            offset,
                            'BLOCK_AS_IMPLICIT_KEY',
                            'Nested mappings are not allowed in compact mappings',
                        );
                    if(properties.start < this.parser.get(valueProps.found).offset - 1024)
                        this.rangeError(
                            key,
                            'KEY_OVER_1024_CHARS',
                            'The : indicator must be at most 1024 chars after the start of an implicit block mapping key',
                        );
                }
                value =
                    item.value >= 0
                        ? this.composeNode(item.value, valueProps)
                        : this.empty(offset, item.sep, valueProps);
                offset = this.node(value).range[2] ?? 0;
            }
            else {
                if(implicit)
                    this.rangeError(key, 'MISSING_CHAR', 'Implicit map keys need to be followed by map values');
                this.appendComment(key, valueProps.comment);
            }
            this.node(result).items.push(this.pair(key, value, index, itemIndex));
            itemIndex++;
        }
        if(commentEnd > 0 && commentEnd < offset)
            this.error(commentEnd, 'IMPOSSIBLE', 'Map comment with trailing content');
        this.node(result).range = [token.offset, offset, commentEnd < 0 ? offset : commentEnd];
        return result;
    }
    blockSequence(index: number): number {
        const token = this.parser.get(index);
        const result = this.make('Seq', 'YAMLSeq');
        this.atRoot = false;
        this.atKey = false;
        let offset = token.offset;
        let commentEnd = -1;
        for(const item of token.items) {
            const props = this.properties(item.start, '', 'seq-item-ind', item.value, offset, token.indent, true);
            if(props.found < 0) {
                if(props.anchor >= 0 || props.tag >= 0 || item.value >= 0) {
                    if(item.value >= 0 && this.parser.get(item.value).type === 'block-seq')
                        this.error(props.end, 'BAD_INDENT', 'All sequence items must start at the same column');
                    else this.error(offset, 'MISSING_CHAR', 'Sequence item without - indicator');
                }
                else {
                    commentEnd = props.end;
                    if(props.comment !== '') this.node(result).comment = props.comment;
                    continue;
                }
            }
            const value =
                item.value >= 0 ? this.composeNode(item.value, props) : this.empty(props.end, item.start, props);
            offset = this.node(value).range[2] ?? 0;
            this.node(result).items.push(value);
        }
        this.node(result).range = [token.offset, offset, commentEnd < 0 ? offset : commentEnd];
        return result;
    }
    isBlock(index: number): boolean {
        if(index < 0) return false;
        const token = this.parser.get(index);
        return token.type === 'block-map' || token.type === 'block-seq';
    }
    flowCollection(index: number): number {
        const token = this.parser.get(index);
        const map = this.parser.get(token.flowStart).source === '{';
        const name = map ? 'flow map' : 'flow sequence';
        const result = this.make(map ? 'Map' : 'Seq', map ? 'YAMLMap' : 'YAMLSeq');
        this.node(result).flow = true;
        const root = this.atRoot;
        this.atRoot = false;
        this.atKey = false;
        let offset = token.offset + this.parser.get(token.flowStart).source.length;
        let itemIndex = 0;
        for(const item of token.items) {
            const props = this.properties(
                item.start,
                name,
                'explicit-key-ind',
                item.key >= 0 ? item.key : (item.sep[0] ?? -1),
                offset,
                token.indent,
                false,
            );
            if(props.found < 0) {
                if(props.anchor < 0 && props.tag < 0 && !item.sepPresent && item.value < 0) {
                    if(itemIndex === 0 && props.comma >= 0)
                        this.tokenError(props.comma, 'UNEXPECTED_TOKEN', `Unexpected , in ${name}`);
                    else if(itemIndex < token.items.length - 1)
                        this.error(props.start, 'UNEXPECTED_TOKEN', `Unexpected empty item in ${name}`);
                    this.appendComment(result, props.comment);
                    offset = props.end;
                    itemIndex++;
                    continue;
                }
                if(!map && this.newline(item.key))
                    this.tokenError(
                        item.key,
                        'MULTILINE_IMPLICIT_KEY',
                        'Implicit keys of flow sequence pairs need to be on a single line',
                    );
            }
            if(itemIndex === 0) {
                if(props.comma >= 0) this.tokenError(props.comma, 'UNEXPECTED_TOKEN', `Unexpected , in ${name}`);
            }
            else {
                if(props.comma < 0) this.error(props.start, 'MISSING_CHAR', `Missing , between ${name} items`);
                if(props.comment !== '') {
                    let previousComment = '';
                    for(const start of item.start) {
                        const property = this.parser.get(start);
                        if(property.type === 'comma' || property.type === 'space') continue;
                        if(property.type === 'comment') previousComment = property.source.slice(1);
                        break;
                    }
                    if(previousComment !== '') {
                        let previous = this.node(result).items[this.node(result).items.length - 1] ?? -1;
                        if(previous < 0) {
                            this.tokenError(
                                index,
                                'RESOURCE_EXHAUSTION',
                                "Cannot read properties of undefined (reading 'comment')",
                            );
                            return -1;
                        }
                        if(this.node(previous).kind === 'Pair')
                            previous =
                                this.node(previous).value >= 0 ? this.node(previous).value : this.node(previous).key;
                        this.appendComment(previous, previousComment);
                        props.comment = props.comment.slice(previousComment.length + 1);
                    }
                }
            }
            if(!map && !item.sepPresent && props.found < 0) {
                const value =
                    item.value >= 0 ? this.composeNode(item.value, props) : this.empty(props.end, item.sep, props);
                this.node(result).items.push(value);
                offset = this.node(value).range[2] ?? 0;
                if(this.isBlock(item.value))
                    this.rangeError(
                        value,
                        'BLOCK_IN_FLOW',
                        'Block collections are not allowed within flow collections',
                    );
            }
            else {
                this.atKey = true;
                const key =
                    item.key >= 0 ? this.composeNode(item.key, props) : this.empty(props.end, item.start, props);
                if(this.isBlock(item.key))
                    this.rangeError(key, 'BLOCK_IN_FLOW', 'Block collections are not allowed within flow collections');
                this.atKey = false;
                const valueProps = this.properties(
                    item.sep,
                    name,
                    'map-value-ind',
                    item.value,
                    this.node(key).range[2] ?? 0,
                    token.indent,
                    false,
                );
                if(valueProps.found >= 0) {
                    if(!map && props.found < 0) {
                        for(const separator of item.sep) {
                            if(separator === valueProps.found) break;
                            if(this.parser.get(separator).type === 'newline') {
                                this.tokenError(
                                    separator,
                                    'MULTILINE_IMPLICIT_KEY',
                                    'Implicit keys of flow sequence pairs need to be on a single line',
                                );
                                break;
                            }
                        }
                        if(props.start < this.parser.get(valueProps.found).offset - 1024)
                            this.tokenError(
                                valueProps.found,
                                'KEY_OVER_1024_CHARS',
                                'The : indicator must be at most 1024 chars after the start of an implicit flow sequence key',
                            );
                    }
                }
                else if(item.value >= 0) {
                    if(this.parser.get(item.value).source.startsWith(':'))
                        this.tokenError(item.value, 'MISSING_CHAR', `Missing space after : in ${name}`);
                    else this.error(valueProps.start, 'MISSING_CHAR', `Missing , or : between ${name} items`);
                }
                let value = -1;
                if(item.value >= 0) value = this.composeNode(item.value, valueProps);
                else if(valueProps.found >= 0) value = this.empty(valueProps.end, item.sep, valueProps);
                if(value >= 0) {
                    if(this.isBlock(item.value))
                        this.rangeError(
                            value,
                            'BLOCK_IN_FLOW',
                            'Block collections are not allowed within flow collections',
                        );
                }
                else this.appendComment(key, valueProps.comment);
                const pair = this.pair(key, value, index, itemIndex);
                if(map) this.node(result).items.push(pair);
                else {
                    const wrapper = this.make('Map', 'YAMLMap');
                    this.node(wrapper).flow = true;
                    this.node(wrapper).items.push(pair);
                    const endRange = this.node(value >= 0 ? value : key).range;
                    this.node(wrapper).range = [this.node(key).range[0] ?? 0, endRange[1] ?? 0, endRange[2] ?? 0];
                    this.node(result).items.push(wrapper);
                }
                offset = value >= 0 ? (this.node(value).range[2] ?? 0) : valueProps.end;
            }
            itemIndex++;
        }
        const expected = map ? '}' : ']';
        const first = token.end[0] ?? -1;
        let endTokens = token.end.slice(1);
        let end = offset;
        if(first >= 0 && this.parser.get(first).source === expected)
            end = this.parser.get(first).offset + this.parser.get(first).source.length;
        else {
            const title = map ? 'Flow map' : 'Flow sequence';
            this.error(
                offset,
                root ? 'MISSING_CHAR' : 'BAD_INDENT',
                root
                    ? `${title} must end with a ${expected}`
                    : `${title} in block collection must be sufficiently indented and end with a ${expected}`,
            );
            if(first >= 0 && this.parser.get(first).source.length !== 1) endTokens = token.end;
        }
        const resolved = new ScalarResolution();
        const nodeEnd = this.scalar.end(endTokens, end, true, resolved);
        for(const error of resolved.errors) this.document.diagnostics.errors.push(error);
        this.appendComment(result, resolved.comment);
        this.node(result).range = [token.offset, end, nodeEnd];
        return result;
    }
    sameValue(first: number, second: number): boolean {
        const left = this.node(first);
        const right = this.node(second);
        if(left.kind !== 'Scalar' || right.kind !== 'Scalar' || left.valueKind !== right.valueKind) return false;
        if(left.valueKind === 'number')
            return (
                left.numberValue === right.numberValue ||
                (Number.isNaN(left.numberValue) && Number.isNaN(right.numberValue))
            );
        if(left.valueKind === 'string') return left.stringValue === right.stringValue;
        if(left.valueKind === 'boolean') return left.booleanValue === right.booleanValue;
        return left.valueKind === 'null';
    }
    valueText(index: number): string {
        const node = this.node(index);
        if(node.valueKind === 'null') return 'null';
        if(node.valueKind === 'number') return String(node.numberValue);
        if(node.valueKind === 'boolean') return node.booleanValue ? 'true' : 'false';
        return node.stringValue;
    }
    prototypeTag(tagName: string): boolean {
        return [
            '__proto__',
            '__defineGetter__',
            '__defineSetter__',
            '__lookupGetter__',
            '__lookupSetter__',
            'constructor',
            'hasOwnProperty',
            'isPrototypeOf',
            'propertyIsEnumerable',
            'toLocaleString',
            'toString',
            'valueOf',
        ].includes(tagName);
    }
    collectionTag(index: number, tokenIndex: number): void {
        const node = this.node(index);
        if(
            tokenIndex < 0 ||
            node.tag === '' ||
            node.tag === '!' ||
            node.tag === `tag:yaml.org,2002:${node.kind === 'Map' ? 'map' : 'seq'}`
        )
            return;
        if(node.tag === 'tag:yaml.org,2002:set' && node.kind === 'Map') {
            node.className = 'YAMLSet';
            for(const item of node.items) {
                const value = this.node(item).value;
                if(value >= 0) {
                    const scalar = this.node(value);
                    if(
                        scalar.kind !== 'Scalar' ||
                        scalar.valueKind !== 'null' ||
                        scalar.commentBefore !== '' ||
                        scalar.comment !== '' ||
                        scalar.tag !== ''
                    ) {
                        this.tokenError(tokenIndex, 'TAG_RESOLVE_FAILED', 'Set items must all have null values');
                        break;
                    }
                }
            }
        }
        else if(
            (node.tag === 'tag:yaml.org,2002:pairs' || node.tag === 'tag:yaml.org,2002:omap') &&
            node.kind === 'Seq'
        ) {
            if(node.tag === 'tag:yaml.org,2002:omap') node.className = 'YAMLOMap';
            for(let itemIndex = 0; itemIndex < node.items.length; itemIndex++) {
                const item = this.node(node.items[itemIndex] ?? -1);
                if(item.kind === 'Pair') continue;
                if(item.kind === 'Map') {
                    if(item.items.length > 1)
                        this.tokenError(
                            tokenIndex,
                            'TAG_RESOLVE_FAILED',
                            'Each pair must have its own sequence indicator',
                        );
                    let pair = item.items[0] ?? -1;
                    if(pair < 0) pair = this.pair(this.make('Scalar', 'Scalar'), -1, -1, -1);
                    if(item.commentBefore !== '') {
                        const key = this.node(this.node(pair).key);
                        key.commentBefore =
                            key.commentBefore === ''
                                ? item.commentBefore
                                : `${item.commentBefore}\n${key.commentBefore}`;
                    }
                    if(item.comment !== '') {
                        const value = this.node(pair).value;
                        const key = this.node(pair).key;
                        const receiver = this.node(value >= 0 ? value : key);
                        receiver.comment =
                            receiver.comment === '' ? item.comment : `${item.comment}\n${receiver.comment}`;
                    }
                    node.items[itemIndex] = pair;
                }
                else node.items[itemIndex] = this.pair(node.items[itemIndex] ?? -1, -1, -1, -1);
            }
            if(node.tag === 'tag:yaml.org,2002:omap') {
                const keys: number[] = [];
                for(const pair of node.items) {
                    const key = this.node(pair).key;
                    if(this.node(key).kind !== 'Scalar') continue;
                    let duplicate = false;
                    for(const previous of keys) if(this.sameValue(previous, key)) duplicate = true;
                    if(duplicate)
                        this.tokenError(
                            tokenIndex,
                            'TAG_RESOLVE_FAILED',
                            `Ordered maps must not include duplicate keys: ${this.valueText(key)}`,
                        );
                    else keys.push(key);
                }
            }
        }
        else {
            let expects = '';
            if(node.tag === 'tag:yaml.org,2002:set') expects = 'map';
            if(node.tag === 'tag:yaml.org,2002:pairs' || node.tag === 'tag:yaml.org,2002:omap') expects = 'seq';
            if(
                node.tag === 'tag:yaml.org,2002:binary' ||
                node.tag === 'tag:yaml.org,2002:merge' ||
                node.tag === 'tag:yaml.org,2002:timestamp'
            )
                expects = 'scalar';
            const prototype = this.prototypeTag(node.tag);
            if(prototype) expects = 'scalar';
            if(expects !== '' && (prototype || this.document.directives.version === '1.2'))
                this.warning(
                    tokenIndex,
                    'BAD_COLLECTION_TYPE',
                    `${prototype ? 'undefined' : node.tag} used for ${node.kind === 'Map' ? 'map' : 'seq'} collection, but expects ${expects}`,
                );
            else this.warning(tokenIndex, 'TAG_RESOLVE_FAILED', `Unresolved tag: ${node.tag}`);
        }
    }
    decorate(after: boolean): void {
        let comment = '';
        let afterEmptyLine = false;
        let atComment = false;
        for(let index = 0; index < this.prelude.length; index++) {
            const source = this.prelude[index] ?? '';
            if(source.startsWith('#')) {
                let text = source.slice(1);
                if(text === '') text = ' ';
                comment = comment === '' ? text : `${comment}${afterEmptyLine ? '\n\n' : '\n'}${text}`;
                atComment = true;
                afterEmptyLine = false;
            }
            else if(source.startsWith('%')) {
                if(!(this.prelude[index + 1] ?? '').startsWith('#')) index++;
                atComment = false;
            }
            else {
                if(!atComment) afterEmptyLine = true;
                atComment = false;
            }
        }
        if(comment !== '') {
            if(after)
                this.document.comment = this.document.comment === '' ? comment : `${this.document.comment}\n${comment}`;
            else if(afterEmptyLine || this.document.directives.docStart || this.document.contents < 0)
                this.document.commentBefore = comment;
            else {
                let target = this.document.contents;
                if(
                    (this.node(target).kind === 'Map' || this.node(target).kind === 'Seq') &&
                    !this.node(target).flow &&
                    this.node(target).items.length > 0
                ) {
                    target = this.node(target).items[0] ?? -1;
                    if(this.node(target).kind === 'Pair') target = this.node(target).key;
                }
                const node = this.node(target);
                node.commentBefore = node.commentBefore === '' ? comment : `${comment}\n${node.commentBefore}`;
            }
        }
        this.prelude = [];
    }
    compose(endOffset: number): void {
        for(const index of this.parser.roots) {
            const token = this.parser.get(index);
            if(token.type === 'directive') {
                this.directives.add(token);
                this.prelude.push(token.source);
                this.atDirectives = true;
            }
            else if(token.type === 'document') {
                const previous = this.document;
                this.document = new ComposedDocument(this.directives.atDocument());
                for(const error of this.directives.errors) this.document.diagnostics.errors.push(error);
                for(const warning of this.directives.warnings) this.document.warnings.push(warning);
                this.directives.errors = [];
                this.directives.warnings = [];
                this.atRoot = true;
                this.atKey = false;
                const props = this.properties(
                    token.start,
                    '',
                    'doc-start',
                    token.value >= 0 ? token.value : (token.end[0] ?? -1),
                    token.offset,
                    0,
                    true,
                );
                this.document.directives.docStart = props.found >= 0;
                if(props.found >= 0 && token.value >= 0 && this.isBlock(token.value) && !props.hasNewline)
                    this.error(
                        props.end,
                        'MISSING_CHAR',
                        'Block collection cannot start on same line with directives-end marker',
                    );
                this.document.contents =
                    token.value >= 0 ? this.composeNode(token.value, props) : this.empty(props.end, token.start, props);
                const contentEnd = this.node(this.document.contents).range[2] ?? 0;
                const resolved = new ScalarResolution();
                const end = this.scalar.end(token.end, contentEnd, false, resolved);
                for(const error of resolved.errors) this.document.diagnostics.errors.push(error);
                this.document.comment = resolved.comment;
                this.document.range = [token.offset, contentEnd, end];
                if(this.atDirectives && !this.document.directives.docStart)
                    this.tokenError(index, 'MISSING_CHAR', 'Missing directives-end/doc-start indicator line');
                this.decorate(false);
                if(this.documentPresent) this.documents.push(previous);
                this.documentPresent = true;
                this.atDirectives = false;
            }
            else if(token.type === 'comment' || token.type === 'newline') this.prelude.push(token.source);
            else if(token.type === 'error') {
                // The stream token source is quoted as JSON by the original composer.
                const message = token.source === '' ? token.message : `${token.message}: ${this.quoted(token.source)}`;
                if(this.atDirectives || !this.documentPresent)
                    this.directives.errors.push(
                        new ComposeError(token.offset, token.offset + token.source.length, 'UNEXPECTED_TOKEN', message),
                    );
                else this.tokenError(index, 'UNEXPECTED_TOKEN', message);
            }
            else if(token.type === 'doc-end') {
                if(!this.documentPresent)
                    this.directives.errors.push(
                        new ComposeError(
                            token.offset,
                            token.offset + token.source.length,
                            'UNEXPECTED_TOKEN',
                            'Unexpected doc-end without preceding document',
                        ),
                    );
                else {
                    this.document.directives.docEnd = true;
                    const resolved = new ScalarResolution();
                    const end = this.scalar.end(token.end, token.offset + token.source.length, true, resolved);
                    for(const error of resolved.errors) this.document.diagnostics.errors.push(error);
                    this.decorate(true);
                    this.document.comment =
                        resolved.comment === ''
                            ? this.document.comment
                            : this.document.comment === ''
                              ? resolved.comment
                              : `${this.document.comment}\n${resolved.comment}`;
                    this.document.range[2] = end;
                }
            }
            else if(token.type !== 'byte-order-mark' && token.type !== 'space')
                this.directives.errors.push(
                    new ComposeError(
                        token.offset,
                        token.offset + token.source.length,
                        'UNEXPECTED_TOKEN',
                        `Unsupported token ${token.type}`,
                    ),
                );
        }
        if(this.documentPresent) {
            for(const error of this.directives.errors) this.document.diagnostics.errors.push(error);
            for(const warning of this.directives.warnings) this.document.warnings.push(warning);
            this.decorate(true);
            this.documents.push(this.document);
        }
        else {
            this.document = new ComposedDocument(this.directives.atDocument());
            for(const error of this.directives.errors) this.document.diagnostics.errors.push(error);
            for(const warning of this.directives.warnings) this.document.warnings.push(warning);
            if(this.atDirectives) this.error(endOffset, 'MISSING_CHAR', 'Missing directives-end indicator line');
            this.document.range = [0, endOffset, endOffset];
            this.decorate(false);
            this.documents.push(this.document);
        }
    }
    quoted(source: string): string {
        const parts: string[] = ['"'];
        for(let index = 0; index < source.length; index++) {
            const character = source.slice(index, index + 1);
            const code = source.charCodeAt(index);
            if(character === '"' || character === '\\') parts.push(`\\${character}`);
            else if(character === '\b') parts.push('\\b');
            else if(character === '\f') parts.push('\\f');
            else if(character === '\n') parts.push('\\n');
            else if(character === '\r') parts.push('\\r');
            else if(character === '\t') parts.push('\\t');
            else if(code < 32 || (code >= 0xd800 && code <= 0xdfff)) {
                const next = source.charCodeAt(index + 1);
                if(code >= 0xd800 && code <= 0xdbff && next >= 0xdc00 && next <= 0xdfff) {
                    parts.push(source.slice(index, index + 2));
                    index++;
                }
                else parts.push(`\\u${code.toString(16).padStart(4, '0')}`);
            }
            else parts.push(character);
        }
        parts.push('"');
        return parts.join('');
    }
}
