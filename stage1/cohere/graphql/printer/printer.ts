// Port of cohere/internal/format/graphql/print.go and print_helpers.go, composed
// with the existing parser. The reachable shared comment core is below too.
import { panic } from 'adamic';
import { parse, type Document, type GraphNode, type Value, type Comment } from '../parser.ts';
import { Documents, defaults, type SettingsOptions } from './doc.ts';
const absent: Value = { kind: 'Undefined' };
function field(node: GraphNode, key: string): Value {
    for(const property of node.fields) {
        if(property.key === key) return property.value;
    }
    return absent;
}
function child(node: GraphNode, key: string): number {
    const value = field(node, key);
    return value.kind === 'Node' ? value.node : -1;
}
function list(node: GraphNode, key: string): readonly number[] {
    const value = field(node, key);
    return value.kind === 'List' ? value.nodes : [];
}
function flag(node: GraphNode, key: string): boolean {
    const value = field(node, key);
    return value.kind === 'Flag' && value.flag;
}
// The parser's transport form is injective for cooked scalar strings. Decode
// doubled backslashes and \u{HEX}, not GraphQL escapes (already cooked).
export function decoded(text: string): string {
    const parts: string[] = [];
    let run = 0;
    let position = 0;
    while(position < text.length) {
        if(text.charCodeAt(position) !== 92) {
            position++;
            continue;
        }
        parts.push(text.slice(run, position));
        if(text.charCodeAt(position + 1) === 92) {
            parts.push('\\');
            position += 2;
        }
        else {
            if(text.slice(position, position + 3) !== '\\u{') panic('invalid parser string transport');
            let point = 0;
            position += 3;
            while(position < text.length && text.charCodeAt(position) !== 125) {
                const characterCode = text.charCodeAt(position);
                point = point * 16 + (characterCode >= 65 ? characterCode - 55 : characterCode - 48);
                position++;
            }
            if(position === text.length) panic('unterminated parser string transport');
            parts.push(String.fromCodePoint(point));
            position++;
        }
        run = position;
    }
    parts.push(text.slice(run));
    return parts.join('');
}
function textValue(node: GraphNode, key: string): string {
    const value = field(node, key);
    return value.kind === 'Text' ? decoded(value.text) : '';
}
function isNewline(code: number): boolean {
    return code === 10 || code === 13 || code === 0x2028 || code === 0x2029;
}
function spaces(text: string, index: number, back: boolean): number {
    let position = index;
    while(
        position >= 0 &&
        position < text.length &&
        (text.charCodeAt(position) === 32 || text.charCodeAt(position) === 9)
    )
        position += back ? -1 : 1;
    return position;
}
function newline(text: string, index: number, back: boolean): number {
    const code = text.charCodeAt(index);
    if(back && index > 0 && text.charCodeAt(index - 1) === 13 && code === 10) return index - 2;
    if(!back && code === 13 && text.charCodeAt(index + 1) === 10) return index + 2;
    return isNewline(code) ? index + (back ? -1 : 1) : index;
}
function hasNewline(text: string, index: number, back: boolean): boolean {
    const position = spaces(text, index - (back ? 1 : 0), back);
    return position !== newline(text, position, back);
}
function previousEmpty(text: string, index: number): boolean {
    let position = spaces(text, index - 1, true);
    position = newline(text, position, true);
    position = spaces(text, position, true);
    return position !== newline(text, position, true);
}
function nextEmpty(text: string, index: number): boolean {
    let position = index;
    let old = -1;
    // shared isNextLineEmpty skips JS-style inline and trailing comments, even
    // in GraphQL. It deliberately does not skip # comments.
    while(position !== old) {
        old = position;
        while(position < text.length && ',; \t'.includes(text.slice(position, position + 1))) position++;
        if(text.slice(position, position + 2) === '/*') {
            const end = text.indexOf('*/', position + 2);
            if(end >= 0) position = end + 2;
        }
        position = spaces(text, position, false);
    }
    if(text.slice(position, position + 2) === '//') {
        while(position < text.length && text.charCodeAt(position) !== 10 && text.charCodeAt(position) !== 13)
            position++;
    }
    position = newline(text, position, false);
    return position >= -1 && hasNewline(text, position, false);
}
interface AttachmentInterface {
    readonly comment: number;
    readonly node: number;
    readonly placement: string;
}
interface NeighboursInterface {
    readonly enclosing: number;
    readonly preceding: number;
    readonly following: number;
}
class Printer {
    readonly ast: Document;
    readonly source: string;
    readonly docs: Documents;
    readonly attachments: AttachmentInterface[] = [];
    readonly printed: number[] = [];
    // Gap 1: the caller constructs Documents before this constructor runs.
    constructor(ast: Document, source: string, docs: Documents) {
        this.ast = ast;
        this.source = source;
        this.docs = docs;
    }
    node(index: number): GraphNode {
        return this.ast.nodes[index] ?? panic(`missing AST node ${index}`);
    }
    neighbours(index: number, comment: Comment, enclosing: number): NeighboursInterface {
        const children: number[] = [];
        for(const property of this.node(index).fields) {
            if(property.value.kind === 'Node') children.push(property.value.node);
            else if(property.value.kind === 'List') {
                for(const item of property.value.nodes) children.push(item);
            }
        }
        children.sort((leftIndex, rightIndex) => {
            const left = this.node(leftIndex);
            const right = this.node(rightIndex);
            return left.start === right.start ? left.end - right.end : left.start - right.start;
        });
        let left = 0;
        let right = children.length;
        let preceding = -1;
        let following = -1;
        while(left < right) {
            const middle = Math.floor((left + right) / 2);
            const at = children[middle] ?? panic('child');
            const node = this.node(at);
            if(node.start <= comment.start && comment.end <= node.end) return this.neighbours(at, comment, at);
            if(node.end <= comment.start) {
                preceding = at;
                left = middle + 1;
            }
            else if(comment.end <= node.start) {
                following = at;
                right = middle;
            }
            else panic('overlapping GraphQL comment');
        }
        return { enclosing, preceding, following };
    }
    attach(): void {
        for(let position = 0; position < this.ast.comments.length; position++) {
            const comment = this.ast.comments[position] ?? panic('comment');
            const around = this.neighbours(this.ast.root, comment, -1);
            const own = hasNewline(this.source, comment.start, true);
            const end = hasNewline(this.source, comment.end, false);
            let node: number;
            let placement = 'dangling';
            if(own && around.following >= 0) {
                node = around.following;
                placement = 'leading';
            }
            else if(around.preceding >= 0) {
                node = around.preceding;
                placement = 'trailing';
            }
            else if(around.following >= 0) {
                node = around.following;
                placement = 'leading';
            }
            else node = around.enclosing >= 0 ? around.enclosing : this.ast.root;
            // GraphQL has only line comments: with attachment following token attachment non-own-line
            // comment necessarily ends its line, so shared-core remaining ties cannot
            // occur. Keep the invariant loud if the lexer ever gains block comments.
            if(!own && !end && around.preceding >= 0 && around.following >= 0) panic('unexpected GraphQL comment tie');
            this.attachments.push({ comment: position, node, placement });
            this.printed.push(0);
        }
    }
    comment(index: number): number {
        const comment = this.ast.comments[index] ?? panic('comment');
        this.printed[index] = 1;
        return this.docs.text(`#${decoded(comment.value).trimEnd()}`);
    }
    dangling(index: number): number {
        const parts: number[] = [];
        for(const attachment of this.attachments) {
            if(attachment.node === index && attachment.placement === 'dangling')
                parts.push(this.comment(attachment.comment));
        }
        return parts.length === 0
            ? this.docs.text('')
            : this.docs.indent(
                  this.docs.concat([this.docs.line(false, true), this.docs.join(this.docs.line(false, true), parts)]),
              );
    }
    // Generic printer, with comments wrapped as printing.callPluginPrintFunction
    // wraps them. Indexes replace AstPath's reflective selectors.
    print(index: number): number {
        const documents = this.docs;
        if(index < 0) return documents.text('');
        const node = this.node(index);
        let ignored = false;
        for(const attachment of this.attachments) {
            if(
                attachment.node === index &&
                decoded((this.ast.comments[attachment.comment] ?? panic('ignore')).value).trim() === 'prettier-ignore'
            )
                ignored = true;
        }
        let body: number;
        if(ignored) {
            for(let position = 0; position < this.ast.comments.length; position++) {
                const comment = this.ast.comments[position] ?? panic('ignored comment');
                if(comment.start >= node.start && comment.end <= node.end) this.printed[position] = 1;
            }
            body = documents.text(this.source.slice(node.start, node.end));
        }
        else body = this.generic(index);
        const before: number[] = [];
        const after: number[] = [];
        let previousSuffix = false;
        for(const attachment of this.attachments) {
            if(attachment.node !== index || this.printed[attachment.comment] === 1) continue;
            const comment = this.ast.comments[attachment.comment] ?? panic('attached comment');
            if(attachment.placement === 'leading') {
                before.push(this.comment(attachment.comment));
                before.push(documents.line(false, true));
                const position = newline(this.source, spaces(this.source, comment.end, false), false);
                if(hasNewline(this.source, position, false)) before.push(documents.line(false, true));
            }
            else if(attachment.placement === 'trailing') {
                const printed = this.comment(attachment.comment);
                if(previousSuffix || hasNewline(this.source, comment.start, true)) {
                    after.push(
                        documents.suffix(
                            documents.concat([
                                documents.line(false, true),
                                previousEmpty(this.source, comment.start)
                                    ? documents.line(false, true)
                                    : documents.text(''),
                                printed,
                            ]),
                        ),
                    );
                }
                else {
                    after.push(documents.suffix(documents.concat([documents.text(' '), printed])));
                    after.push(documents.breakParent());
                }
                previousSuffix = true;
            }
        }
        return documents.concat([documents.concat(before), body, documents.concat(after)]);
    }
    sequence(index: number, key: string, blank: boolean): number[] {
        const items = list(this.node(index), key);
        const parts: number[] = [];
        for(let position = 0; position < items.length; position++) {
            const at = items[position] ?? panic('sequence');
            const printed = this.print(at);
            parts.push(
                blank && position < items.length - 1 && nextEmpty(this.source, this.node(at).end)
                    ? this.docs.concat([printed, this.docs.line(false, true)])
                    : printed,
            );
        }
        return parts;
    }
    arguments(index: number, key: string, blank: boolean): number {
        const documents = this.docs;
        if(list(this.node(index), key).length === 0) return documents.text('');
        const soft = documents.line(true, false);
        const separator = documents.concat([documents.ifBreak(documents.text(''), documents.text(', ')), soft]);
        return documents.group(
            documents.concat([
                documents.text('('),
                documents.indent(documents.concat([soft, documents.join(separator, this.sequence(index, key, blank))])),
                soft,
                documents.text(')'),
            ]),
        );
    }
    description(index: number): number {
        const documents = this.docs;
        const node = this.node(index);
        const at = child(node, 'description');
        if(at < 0) return documents.text('');
        return documents.concat([
            this.print(at),
            documents.line(false, !(node.kind === 'InputValueDefinition' && !flag(this.node(at), 'block'))),
        ]);
    }
    directives(index: number): number {
        const documents = this.docs;
        const node = this.node(index);
        if(list(node, 'directives').length === 0) return documents.text('');
        const printed = documents.join(documents.line(false, false), this.sequence(index, 'directives', false));
        return node.kind === 'FragmentDefinition' || node.kind === 'OperationDefinition'
            ? documents.group(documents.concat([documents.line(false, false), printed]))
            : documents.concat([
                  documents.text(' '),
                  documents.group(documents.indent(documents.concat([documents.line(true, false), printed]))),
              ]);
    }
    block(index: number, key: string, required: boolean): number {
        const documents = this.docs;
        if(!required && list(this.node(index), key).length === 0) return documents.text('');
        return documents.concat([
            documents.text(' {'),
            documents.indent(
                documents.concat([
                    documents.line(false, true),
                    documents.join(documents.line(false, true), this.sequence(index, key, true)),
                ]),
            ),
            documents.line(false, true),
            documents.text('}'),
        ]);
    }
    generic(index: number): number {
        const documents = this.docs;
        const node = this.node(index);
        const name = this.print(child(node, 'name'));
        const description = this.description(index);
        const directives = this.directives(index);
        switch(node.kind) {
            case 'Document':
                return documents.concat([
                    documents.join(documents.line(false, true), this.sequence(index, 'definitions', true)),
                    documents.line(false, true),
                ]);
            case 'OperationDefinition': {
                const explicit = this.source.slice(node.start, node.start + 1) !== '{';
                const hasName = child(node, 'name') >= 0;
                const parts: number[] = [description];
                if(explicit) parts.push(documents.text(textValue(node, 'operation')));
                if(explicit && hasName) {
                    parts.push(documents.text(' '));
                    parts.push(name);
                }
                if(explicit && !hasName && list(node, 'variableDefinitions').length > 0)
                    parts.push(documents.text(' '));
                parts.push(this.arguments(index, 'variableDefinitions', false));
                parts.push(directives);
                if(explicit || hasName) parts.push(documents.text(' '));
                parts.push(this.print(child(node, 'selectionSet')));
                return documents.concat(parts);
            }
            case 'FragmentDefinition':
                return documents.concat([
                    description,
                    documents.text('fragment '),
                    name,
                    this.arguments(index, 'variableDefinitions', false),
                    documents.text(' on '),
                    this.print(child(node, 'typeCondition')),
                    directives,
                    documents.text(' '),
                    this.print(child(node, 'selectionSet')),
                ]);
            case 'SelectionSet':
                return documents.concat([
                    documents.text('{'),
                    documents.indent(
                        documents.concat([
                            documents.line(false, true),
                            documents.join(documents.line(false, true), this.sequence(index, 'selections', true)),
                        ]),
                    ),
                    documents.line(false, true),
                    documents.text('}'),
                ]);
            case 'Field':
                return documents.group(
                    documents.concat([
                        child(node, 'alias') < 0
                            ? documents.text('')
                            : documents.concat([this.print(child(node, 'alias')), documents.text(': ')]),
                        name,
                        this.arguments(index, 'arguments', true),
                        directives,
                        child(node, 'selectionSet') < 0
                            ? documents.text('')
                            : documents.concat([documents.text(' '), this.print(child(node, 'selectionSet'))]),
                    ]),
                );
            case 'Name':
                return documents.text(textValue(node, 'value'));
            case 'StringValue': {
                const value = textValue(node, 'value');
                if(flag(node, 'block')) {
                    let lines = value.split('"""').join('\\"""').split('\n');
                    if(lines.length === 1) lines = [(lines[0] ?? '').trim()];
                    let empty = true;
                    for(const line of lines) {
                        if(line !== '') empty = false;
                    }
                    const parts: number[] = [documents.text('"""')];
                    if(!empty) {
                        for(const line of lines) parts.push(documents.text(line));
                    }
                    parts.push(documents.text('"""'));
                    return documents.join(documents.line(false, true), parts);
                }
                return documents.text(
                    `"${value.split('\\').join('\\\\').split('"').join('\\"').split('\n').join('\\n')}"`,
                );
            }
            case 'IntValue':
            case 'FloatValue':
            case 'EnumValue':
                return documents.text(textValue(node, 'value'));
            case 'BooleanValue':
                return documents.text(flag(node, 'value') ? 'true' : 'false');
            case 'NullValue':
                return documents.text('null');
            case 'Variable':
                return documents.concat([documents.text('$'), name]);
            case 'ListValue':
            case 'ObjectValue': {
                const object = node.kind === 'ObjectValue';
                const key = object ? 'fields' : 'values';
                const nonempty = list(node, key).length > 0;
                const space = object && nonempty && documents.settings.bracketSpacing ? ' ' : '';
                const soft = documents.line(true, false);
                const separator = documents.concat([documents.ifBreak(documents.text(''), documents.text(', ')), soft]);
                return documents.group(
                    documents.concat([
                        documents.text(object ? '{' : '['),
                        documents.text(space),
                        this.dangling(index),
                        nonempty
                            ? documents.indent(
                                  documents.concat([soft, documents.join(separator, this.sequence(index, key, false))]),
                              )
                            : documents.text(''),
                        soft,
                        documents.ifBreak(documents.text(''), documents.text(space)),
                        documents.text(object ? '}' : ']'),
                    ]),
                );
            }
            case 'ObjectField':
            case 'Argument':
            case 'FragmentArgument':
                return documents.concat([name, documents.text(': '), this.print(child(node, 'value'))]);
            case 'Directive':
                return documents.concat([documents.text('@'), name, this.arguments(index, 'arguments', true)]);
            case 'NamedType':
                return name;
            case 'VariableDefinition':
            case 'InputValueDefinition':
                return documents.concat([
                    description,
                    node.kind === 'VariableDefinition' ? this.print(child(node, 'variable')) : name,
                    documents.text(': '),
                    this.print(child(node, 'type')),
                    child(node, 'defaultValue') < 0
                        ? documents.text('')
                        : documents.concat([documents.text(' = '), this.print(child(node, 'defaultValue'))]),
                    directives,
                ]);
            case 'ObjectTypeDefinition':
            case 'InputObjectTypeDefinition':
            case 'InterfaceTypeDefinition':
            case 'ObjectTypeExtension':
            case 'InputObjectTypeExtension':
            case 'InterfaceTypeExtension': {
                const input = node.kind.startsWith('Input');
                const keyword = input ? 'input' : node.kind.startsWith('Interface') ? 'interface' : 'type';
                const interfaces =
                    list(node, 'interfaces').length === 0
                        ? documents.text('')
                        : documents.concat([
                              documents.text(' implements '),
                              documents.indent(
                                  documents.group(
                                      documents.join(
                                          documents.concat([documents.text(' &'), documents.line(false, false)]),
                                          this.sequence(index, 'interfaces', false),
                                      ),
                                  ),
                              ),
                          ]);
                return documents.concat([
                    node.kind.endsWith('Definition') ? description : documents.text('extend '),
                    documents.text(`${keyword} `),
                    name,
                    interfaces,
                    directives,
                    this.block(index, 'fields', false),
                ]);
            }
            case 'FieldDefinition':
                return documents.concat([
                    description,
                    name,
                    this.arguments(index, 'arguments', true),
                    documents.text(': '),
                    this.print(child(node, 'type')),
                    directives,
                ]);
            case 'DirectiveDefinition':
                return documents.concat([
                    description,
                    documents.text('directive @'),
                    name,
                    this.arguments(index, 'arguments', true),
                    directives,
                    flag(node, 'repeatable') ? documents.text(' repeatable') : documents.text(''),
                    documents.text(' on '),
                    documents.join(documents.text(' | '), this.sequence(index, 'locations', false)),
                ]);
            case 'DirectiveExtension':
                return documents.concat([documents.text('extend directive @'), name, directives]);
            case 'EnumTypeDefinition':
            case 'EnumTypeExtension':
                return documents.concat([
                    description,
                    node.kind.endsWith('Extension') ? documents.text('extend ') : documents.text(''),
                    documents.text('enum '),
                    name,
                    directives,
                    this.block(index, 'values', false),
                ]);
            case 'EnumValueDefinition':
                return documents.concat([description, name, directives]);
            case 'SchemaDefinition':
            case 'SchemaExtension':
                return documents.concat([
                    node.kind === 'SchemaExtension' ? documents.text('extend ') : description,
                    documents.text('schema'),
                    directives,
                    this.block(index, 'operationTypes', node.kind === 'SchemaDefinition'),
                ]);
            case 'OperationTypeDefinition':
                return documents.concat([
                    documents.text(`${textValue(node, 'operation')}: `),
                    this.print(child(node, 'type')),
                ]);
            case 'FragmentSpread':
                return documents.concat([
                    documents.text('...'),
                    name,
                    this.arguments(index, 'arguments', true),
                    directives,
                ]);
            case 'InlineFragment':
                return documents.concat([
                    documents.text('...'),
                    child(node, 'typeCondition') < 0
                        ? documents.text('')
                        : documents.concat([documents.text(' on '), this.print(child(node, 'typeCondition'))]),
                    directives,
                    documents.text(' '),
                    this.print(child(node, 'selectionSet')),
                ]);
            case 'UnionTypeDefinition':
            case 'UnionTypeExtension': {
                const types = this.sequence(index, 'types', false);
                const members =
                    types.length === 0
                        ? documents.text('')
                        : documents.concat([
                              documents.text(' ='),
                              documents.ifBreak(documents.text(''), documents.text(' ')),
                              documents.indent(
                                  documents.concat([
                                      documents.ifBreak(
                                          documents.concat([documents.line(false, false), documents.text('| ')]),
                                          documents.text(''),
                                      ),
                                      documents.join(
                                          documents.concat([documents.line(false, false), documents.text('| ')]),
                                          types,
                                      ),
                                  ]),
                              ),
                          ]);
                return documents.group(
                    documents.concat([
                        description,
                        documents.group(
                            documents.concat([
                                node.kind.endsWith('Extension') ? documents.text('extend ') : documents.text(''),
                                documents.text('union '),
                                name,
                                directives,
                                members,
                            ]),
                        ),
                    ]),
                );
            }
            case 'ScalarTypeDefinition':
            case 'ScalarTypeExtension':
                return documents.concat([
                    description,
                    node.kind.endsWith('Extension') ? documents.text('extend ') : documents.text(''),
                    documents.text('scalar '),
                    name,
                    directives,
                ]);
            case 'NonNullType':
                return documents.concat([this.print(child(node, 'type')), documents.text('!')]);
            case 'ListType':
                return documents.concat([documents.text('['), this.print(child(node, 'type')), documents.text(']')]);
        }
        panic(`Unexpected GraphQL node: ${node.kind}`);
    }
    finish(): string {
        this.attach();
        const root = this.print(this.ast.root);
        for(let position = 0; position < this.printed.length; position++) {
            if(this.printed[position] !== 1) panic(`GraphQL comment not printed: ${position}`);
        }
        return this.docs.print(root);
    }
}
export type ResultType =
    { readonly kind: 'Ok'; readonly text: string } | { readonly kind: 'Error'; readonly message: string };
export function format(source: string, settings: SettingsOptions = defaults): ResultType {
    // cohere/internal/format/native's file wrapper: normalize before parsing and
    // restore the leading BOM after printing. Parser offsets name this normalized text.
    const bom = source.startsWith('\ufeff');
    const text = (bom ? source.slice(1) : source).split('\r\n').join('\n').split('\r').join('\n');
    const result = parse(text);
    if(result.kind === 'Refused') return { kind: 'Error', message: decoded(result.message) };
    const printer = new Printer(result.document, text, new Documents(settings));
    return { kind: 'Ok', text: (bom ? '\ufeff' : '') + printer.finish() };
}
