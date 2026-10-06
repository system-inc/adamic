// Conversion reads indexed TypeScript nodes. Unrepresented syntax fails before output.
import { panic } from 'adamic';
import type { Parser } from '../../typescript/parser/parser.ts';
import { Scanner } from '../../typescript/scanner/scanner.ts';
import { keywords, punctuators } from '../../typescript/scanner/tokens.ts';
import type { ParseNode } from '../../typescript/parser/nodes.ts';
import { Arena } from './arena.ts';
import { Value, absent, boolValue, childValue, listValue, numberValue, stringValue, templateValue } from './values.ts';
import { unitOffsets } from './strippedText.ts';

export class Converter {
    readonly parser: Parser;
    readonly arena = new Arena();
    readonly text: string;
    readonly offsets: readonly number[];
    readonly scanner: Scanner;
    readonly starts: number[] = [];
    readonly spellings = new Map<string, string>();
    constructor(parser: Parser, text: string) {
        this.parser = parser;
        this.text = text;
        this.offsets = unitOffsets(text);
        this.scanner = new Scanner(text);
        for(const [spelling, kind] of keywords) {
            this.spellings.set(kind, spelling);
        }
        for(const [spelling, kind] of punctuators) {
            this.spellings.set(kind, spelling);
        }
    }
    ts(id: number): ParseNode {
        return this.parser.node(id);
    }
    kind(id: number): string {
        return id < 0 ? '' : this.ts(id).kind;
    }
    child(id: number, index: number): number {
        return this.ts(id).children[index] ?? -1;
    }
    startUnit(id: number): number {
        const node = this.ts(id);
        this.scanner.pos = node.pos;
        this.scanner.scan();
        return this.scanner.start;
    }
    byte(unit: number): number {
        return this.offsets[unit] ?? panic('AST range outside source');
    }
    start(id: number): number {
        return this.byte(this.startUnit(id));
    }
    raw(id: number): string {
        return this.text.slice(this.startUnit(id), this.ts(id).end);
    }
    create(id: number, type: string): number {
        return this.arena.newNode(type, this.start(id), this.byte(this.ts(id).end));
    }
    set(id: number, key: string, value: Value): void {
        this.arena.node(id).set(key, value);
    }
    converted(id: number, pattern = false, parent = -1): Value {
        return childValue(this.convert(id, pattern, parent));
    }
    list(ids: readonly number[], pattern = false, parent = -1): Value {
        const result: number[] = [];
        for(const id of ids) {
            result.push(this.convert(id, pattern, parent));
        }
        return listValue(result);
    }
    identifier(id: number): number {
        const result = this.create(id, 'Identifier');
        this.set(result, 'decorators', listValue([]));
        this.set(result, 'name', stringValue(this.ts(id).text));
        this.set(result, 'optional', boolValue(false));
        this.set(result, 'typeAnnotation', absent());
        return result;
    }
    patternFields(id: number): void {
        this.set(id, 'decorators', listValue([]));
        this.set(id, 'optional', boolValue(false));
        this.set(id, 'typeAnnotation', absent());
    }
    annotation(id: number): number {
        const result = this.arena.newNode(
            'TSTypeAnnotation',
            this.byte(this.ts(id).pos - 1),
            this.byte(this.ts(id).end),
        );
        this.set(result, 'typeAnnotation', this.converted(id));
        return result;
    }
    bind(id: number, type: number): number {
        const result = this.convert(id);
        if(type >= 0) {
            const annotation = this.annotation(type);
            this.set(result, 'typeAnnotation', childValue(annotation));
            this.arena.node(result).end = Math.max(this.arena.node(result).end, this.arena.node(annotation).end);
        }
        return result;
    }
    property(id: number, pattern: boolean): number {
        const node = this.ts(id);
        const name = this.child(id, 0);
        const shorthand = node.kind === 'ShorthandPropertyAssignment';
        if(shorthand && node.children.length !== 1) {
            return panic('ESTree converter does not yet represent shorthand defaults');
        }
        const result = this.create(id, 'Property');
        this.set(result, 'computed', boolValue(this.kind(name) === 'ComputedPropertyName'));
        this.set(result, 'key', this.converted(name));
        this.set(result, 'kind', stringValue('init'));
        this.set(result, 'method', boolValue(false));
        this.set(result, 'optional', boolValue(false));
        this.set(result, 'shorthand', boolValue(shorthand));
        this.set(result, 'value', this.converted(shorthand ? name : this.child(id, 1), pattern, id));
        return result;
    }
    body(ids: readonly number[], parent: number, directives: boolean): Value {
        const result: number[] = [];
        let allowed = directives;
        for(const id of ids) {
            if(this.kind(id) === 'EndOfFile') {
                continue;
            }
            const child = this.convert(id, false, parent);
            if(child < 0) {
                continue;
            }
            const expression = this.arena.node(child).child('expression');
            if(allowed && this.kind(id) === 'ExpressionStatement' && this.kind(this.child(id, 0)) === 'StringLiteral') {
                const raw = this.arena.node(expression).string('raw');
                this.set(child, 'directive', stringValue(raw.slice(1, -1)));
            }
            else {
                allowed = false;
            }
            result.push(child);
        }
        return listValue(result);
    }
    chain(result: number, source: number): number {
        const node = this.arena.node(result);
        const key =
            node.type === 'CallExpression' ? 'callee' : node.type === 'MemberExpression' ? 'object' : 'expression';
        const child = node.child(key);
        let chain = false;
        if(
            this.arena.type(child) === 'ChainExpression' &&
            this.kind(this.child(source, 0)) !== 'ParenthesizedExpression'
        ) {
            this.set(result, key, this.arena.node(child).get('expression'));
            chain = true;
        }
        if(!node.bool('optional') && !chain) {
            return result;
        }
        const wrapper = this.arena.newNode('ChainExpression', node.start, node.end);
        this.set(wrapper, 'expression', childValue(result));
        return wrapper;
    }
    convert(id: number, pattern = false, parent = -1): number {
        if(id < 0) {
            return -1;
        }
        const node = this.ts(id);
        const first = this.child(id, 0);
        const second = this.child(id, 1);
        const third = this.child(id, 2);
        if(node.kind === 'SourceFile') {
            const result = this.create(id, 'Program');
            this.set(result, 'body', this.body(node.children, id, true));
            this.set(result, 'comments', absent());
            this.set(result, 'sourceType', stringValue('module'));
            this.set(result, 'tokens', absent());
            return result;
        }
        if(node.kind === 'Block' || node.kind === 'ModuleBlock') {
            const result = this.create(id, node.kind === 'Block' ? 'BlockStatement' : 'TSModuleBlock');
            const directives =
                node.kind === 'ModuleBlock' ||
                [
                    'FunctionDeclaration',
                    'FunctionExpression',
                    'ArrowFunction',
                    'MethodDeclaration',
                    'GetAccessor',
                    'SetAccessor',
                    'Constructor',
                ].includes(this.kind(parent));
            this.set(result, 'body', this.body(node.children, id, directives));
            return result;
        }
        if(node.kind === 'Identifier') {
            return this.identifier(id);
        }
        if(node.kind === 'PrivateIdentifier') {
            const result = this.create(id, 'PrivateIdentifier');
            this.set(result, 'name', stringValue(node.text.slice(1)));
            return result;
        }
        if(node.kind === 'OmittedExpression') {
            return -1;
        }
        if(
            node.kind === 'ParenthesizedExpression' ||
            node.kind === 'ParenthesizedType' ||
            node.kind === 'ComputedPropertyName'
        ) {
            return this.convert(first, false, parent);
        }
        if(node.kind === 'ExpressionStatement') {
            const result = this.create(id, 'ExpressionStatement');
            this.set(result, 'directive', absent());
            this.set(result, 'expression', this.converted(first, false, id));
            return result;
        }
        if(node.kind === 'VariableStatement') {
            if(node.children.length !== 1) {
                return panic('ESTree converter does not yet represent variable modifiers');
            }
            const result = this.convert(first);
            this.arena.node(result).start = this.start(id);
            this.arena.node(result).end = this.byte(node.end);
            return result;
        }
        if(node.kind === 'VariableDeclarationList') {
            const result = this.create(id, 'VariableDeclaration');
            this.set(result, 'declarations', this.list(node.children, false, id));
            this.set(result, 'declare', boolValue(false));
            this.set(
                result,
                'kind',
                stringValue(
                    node.semantic === '1'
                        ? 'let'
                        : node.semantic === '2'
                          ? 'const'
                          : node.semantic === '4'
                            ? 'using'
                            : node.semantic === '6'
                              ? 'await using'
                              : 'var',
                ),
            );
            return result;
        }
        if(node.kind === 'VariableDeclaration') {
            let index = 1;
            let definite = false;
            if(this.kind(node.children[index] ?? -1) === 'ExclamationToken') {
                definite = true;
                index++;
            }
            let type = -1;

            if(index < node.children.length) {
                const next = node.children[index] ?? -1;
                this.scanner.pos = this.ts(next).pos;
                this.scanner.scan();
                const previous = this.text.slice(this.ts(first).end, this.scanner.start);
                if(previous.includes(':')) {
                    type = next;
                    index++;
                }
            }
            const init = node.children[index] ?? -1;
            const result = this.create(id, 'VariableDeclarator');
            this.set(result, 'definite', boolValue(definite));
            this.set(result, 'id', childValue(this.bind(first, type)));
            this.set(result, 'init', this.converted(init));
            return result;
        }
        if(node.kind === 'ArrayLiteralExpression' || node.kind === 'ObjectLiteralExpression') {
            const array = node.kind === 'ArrayLiteralExpression';
            const result = this.create(
                id,
                array ? (pattern ? 'ArrayPattern' : 'ArrayExpression') : pattern ? 'ObjectPattern' : 'ObjectExpression',
            );
            if(pattern) {
                this.set(result, 'decorators', listValue([]));
                if(!array) {
                    this.set(result, 'optional', boolValue(false));
                }
            }
            this.set(result, array ? 'elements' : 'properties', this.list(node.children, pattern, id));
            if(pattern) {
                if(array) {
                    this.set(result, 'optional', boolValue(false));
                }
                this.set(result, 'typeAnnotation', absent());
            }
            return result;
        }
        if(node.kind === 'PropertyAssignment' || node.kind === 'ShorthandPropertyAssignment') {
            return this.property(id, pattern);
        }
        if(node.kind === 'SpreadElement' || node.kind === 'SpreadAssignment') {
            const result = this.create(id, pattern ? 'RestElement' : 'SpreadElement');
            this.set(result, 'argument', this.converted(first, pattern));
            if(pattern) {
                this.patternFields(result);
                this.set(result, 'value', absent());
            }
            return result;
        }
        if(node.kind === 'BinaryExpression') {
            const operator = this.raw(second);
            const assignment = [
                '=',
                '+=',
                '-=',
                '*=',
                '**=',
                '/=',
                '%=',
                '<<=',
                '>>=',
                '>>>=',
                '&=',
                '|=',
                '^=',
                '&&=',
                '||=',
                '??=',
            ].includes(operator);
            const logical = ['&&', '||', '??'].includes(operator);
            if(operator === ',') {
                const left = this.convert(first);
                const expressions: number[] = [];
                if(this.arena.type(left) === 'SequenceExpression' && this.kind(first) !== 'ParenthesizedExpression') {
                    for(const expression of this.arena.node(left).list('expressions')) {
                        expressions.push(expression);
                    }
                }
                else {
                    expressions.push(left);
                }
                expressions.push(this.convert(third));
                const result = this.create(id, 'SequenceExpression');
                this.set(result, 'expressions', listValue(expressions));
                return result;
            }
            const result = this.create(
                id,
                assignment
                    ? pattern
                        ? 'AssignmentPattern'
                        : 'AssignmentExpression'
                    : logical
                      ? 'LogicalExpression'
                      : 'BinaryExpression',
            );
            if(pattern && assignment) {
                this.set(result, 'decorators', listValue([]));
                this.set(result, 'left', this.converted(first, true, id));
                this.set(result, 'optional', boolValue(false));
                this.set(result, 'right', this.converted(third));
                this.set(result, 'typeAnnotation', absent());
            }
            else {
                this.set(result, 'operator', stringValue(operator));
                this.set(result, 'left', this.converted(first, assignment, id));
                this.set(result, 'right', this.converted(third));
            }
            return result;
        }
        if(node.kind === 'PropertyAccessExpression' || node.kind === 'ElementAccessExpression') {
            const question = this.kind(second) === 'QuestionDotToken';
            const result = this.create(id, 'MemberExpression');
            this.set(result, 'computed', boolValue(node.kind === 'ElementAccessExpression'));
            this.set(result, 'object', this.converted(first));
            this.set(result, 'optional', boolValue(question));
            this.set(result, 'property', this.converted(question ? third : second));
            return this.chain(result, id);
        }
        if(node.kind === 'CallExpression' || node.kind === 'NewExpression') {
            const count = Math.max(0, node.list);
            const args = node.children.slice(node.children.length - count);
            const prefix = node.children.slice(1, node.children.length - count);
            const question = prefix.length > 0 && this.kind(prefix[0] ?? -1) === 'QuestionDotToken';
            if(prefix.length > (question ? 1 : 0)) {
                return panic('ESTree converter requires unrepresented type-argument list ranges');
            }
            const result = this.create(id, node.kind);
            if(this.kind(first) === 'ImportKeyword') {
                this.arena.node(result).type = 'ImportExpression';
                this.set(result, 'options', this.converted(args[1] ?? -1));
                this.set(result, 'source', this.converted(args[0] ?? -1));
                return result;
            }
            this.set(result, 'arguments', this.list(args));
            this.set(result, 'callee', this.converted(first));
            if(node.kind === 'CallExpression') {
                this.set(result, 'optional', boolValue(question));
            }
            this.set(result, 'typeArguments', absent());
            return node.kind === 'CallExpression' ? this.chain(result, id) : result;
        }
        if(
            node.kind === 'PrefixUnaryExpression' ||
            node.kind === 'PostfixUnaryExpression' ||
            node.kind === 'DeleteExpression' ||
            node.kind === 'VoidExpression' ||
            node.kind === 'TypeOfExpression'
        ) {
            const operator =
                node.operator !== ''
                    ? (this.spellings.get(node.operator) ?? panic('missing unary operator'))
                    : node.kind === 'DeleteExpression'
                      ? 'delete'
                      : node.kind === 'VoidExpression'
                        ? 'void'
                        : 'typeof';
            const result = this.create(
                id,
                operator === '++' || operator === '--' ? 'UpdateExpression' : 'UnaryExpression',
            );
            this.set(result, 'argument', this.converted(first));
            this.set(result, 'operator', stringValue(operator));
            this.set(result, 'prefix', boolValue(node.kind !== 'PostfixUnaryExpression'));
            return result;
        }
        if(
            node.kind === 'StringLiteral' ||
            node.kind === 'NumericLiteral' ||
            node.kind === 'BigIntLiteral' ||
            node.kind === 'RegularExpressionLiteral' ||
            node.kind === 'TrueKeyword' ||
            node.kind === 'FalseKeyword' ||
            node.kind === 'NullKeyword'
        ) {
            const result = this.create(id, 'Literal');
            const raw = this.raw(id);
            if(node.kind === 'BigIntLiteral') {
                this.set(result, 'bigint', stringValue(node.text.slice(0, -1)));
            }
            this.set(result, 'raw', stringValue(raw));
            if(node.kind === 'RegularExpressionLiteral') {
                const value = new Value('regex');
                const slash = raw.lastIndexOf('/');
                value.text = raw.slice(1, slash);
                value.cooked = raw.slice(slash + 1);
                this.set(result, 'regex', value);
            }
            this.set(
                result,
                'value',
                node.kind === 'StringLiteral'
                    ? stringValue(node.text)
                    : node.kind === 'NumericLiteral'
                      ? numberValue(Number(node.text))
                      : node.kind === 'TrueKeyword' || node.kind === 'FalseKeyword'
                        ? boolValue(node.kind === 'TrueKeyword')
                        : absent(),
            );
            return result;
        }
        if(
            node.kind === 'NoSubstitutionTemplateLiteral' ||
            node.kind === 'TemplateExpression' ||
            node.kind === 'TemplateLiteralType'
        ) {
            const result = this.create(
                id,
                node.kind === 'TemplateLiteralType' ? 'TSTemplateLiteralType' : 'TemplateLiteral',
            );
            const expressions: number[] = [];
            const quasis: number[] = [];
            if(node.kind === 'NoSubstitutionTemplateLiteral') {
                quasis.push(this.template(id, true));
            }
            else {
                quasis.push(this.template(first, false));
                for(const span of node.children.slice(1)) {
                    expressions.push(this.convert(this.child(span, 0)));
                    quasis.push(this.template(this.child(span, 1), this.kind(this.child(span, 1)) === 'TemplateTail'));
                }
            }
            if(node.kind === 'TemplateLiteralType') {
                this.set(result, 'quasis', listValue(quasis));
                this.set(result, 'types', listValue(expressions));
            }
            else {
                this.set(result, 'expressions', listValue(expressions));
                this.set(result, 'quasis', listValue(quasis));
            }
            return result;
        }
        if(node.kind === 'TypeReference') {
            if(node.children.length !== 1) {
                return panic('ESTree converter requires unrepresented type-argument list ranges');
            }
            const result = this.create(id, 'TSTypeReference');
            this.set(result, 'typeArguments', absent());
            this.set(result, 'typeName', this.converted(first));
            return result;
        }
        if(node.kind === 'ThisKeyword') {
            const result = this.create(id, 'ThisExpression');
            return result;
        }
        if(node.kind === 'SuperKeyword') {
            const result = this.create(id, 'Super');
            return result;
        }
        if(node.kind === 'ThisType') {
            const result = this.create(id, 'TSThisType');
            return result;
        }
        if(node.kind === 'EmptyStatement') {
            const result = this.create(id, 'EmptyStatement');
            return result;
        }
        if(node.kind === 'DebuggerStatement') {
            const result = this.create(id, 'DebuggerStatement');
            return result;
        }
        if(node.kind === 'QualifiedName') {
            const result = this.create(id, 'TSQualifiedName');
            this.set(result, 'left', this.converted(this.child(id, 0), false, id));
            this.set(result, 'right', this.converted(this.child(id, 1), false, id));
            return result;
        }
        if(node.kind === 'ReturnStatement') {
            const result = this.create(id, 'ReturnStatement');
            this.set(result, 'argument', this.converted(this.child(id, 0), false, id));
            return result;
        }
        if(node.kind === 'ThrowStatement') {
            const result = this.create(id, 'ThrowStatement');
            this.set(result, 'argument', this.converted(this.child(id, 0), false, id));
            return result;
        }
        if(node.kind === 'BreakStatement') {
            const result = this.create(id, 'BreakStatement');
            this.set(result, 'label', this.converted(this.child(id, 0), false, id));
            return result;
        }
        if(node.kind === 'ContinueStatement') {
            const result = this.create(id, 'ContinueStatement');
            this.set(result, 'label', this.converted(this.child(id, 0), false, id));
            return result;
        }
        if(node.kind === 'IfStatement') {
            const result = this.create(id, 'IfStatement');
            this.set(result, 'alternate', this.converted(this.child(id, 2), false, id));
            this.set(result, 'consequent', this.converted(this.child(id, 1), false, id));
            this.set(result, 'test', this.converted(this.child(id, 0), false, id));
            return result;
        }
        if(node.kind === 'WhileStatement') {
            const result = this.create(id, 'WhileStatement');
            this.set(result, 'body', this.converted(this.child(id, 1), false, id));
            this.set(result, 'test', this.converted(this.child(id, 0), false, id));
            return result;
        }
        if(node.kind === 'DoStatement') {
            const result = this.create(id, 'DoWhileStatement');
            this.set(result, 'body', this.converted(this.child(id, 0), false, id));
            this.set(result, 'test', this.converted(this.child(id, 1), false, id));
            return result;
        }
        if(node.kind === 'WithStatement') {
            const result = this.create(id, 'WithStatement');
            this.set(result, 'body', this.converted(this.child(id, 1), false, id));
            this.set(result, 'object', this.converted(this.child(id, 0), false, id));
            return result;
        }
        if(node.kind === 'LabeledStatement') {
            const result = this.create(id, 'LabeledStatement');
            this.set(result, 'body', this.converted(this.child(id, 1), false, id));
            this.set(result, 'label', this.converted(this.child(id, 0), false, id));
            return result;
        }
        if(node.kind === 'ConditionalExpression') {
            const result = this.create(id, 'ConditionalExpression');
            this.set(result, 'alternate', this.converted(this.child(id, 4), false, id));
            this.set(result, 'consequent', this.converted(this.child(id, 2), false, id));
            this.set(result, 'test', this.converted(this.child(id, 0), false, id));
            return result;
        }
        if(node.kind === 'AwaitExpression') {
            const result = this.create(id, 'AwaitExpression');
            this.set(result, 'argument', this.converted(this.child(id, 0), false, id));
            return result;
        }
        if(node.kind === 'NonNullExpression') {
            const result = this.create(id, 'TSNonNullExpression');
            this.set(result, 'expression', this.converted(this.child(id, 0), false, id));
            return this.chain(result, id);
        }
        if(node.kind === 'ArrayType') {
            const result = this.create(id, 'TSArrayType');
            this.set(result, 'elementType', this.converted(this.child(id, 0), false, id));
            return result;
        }
        if(node.kind === 'IndexedAccessType') {
            const result = this.create(id, 'TSIndexedAccessType');
            this.set(result, 'indexType', this.converted(this.child(id, 1), false, id));
            this.set(result, 'objectType', this.converted(this.child(id, 0), false, id));
            return result;
        }
        if(node.kind === 'ConditionalType') {
            const result = this.create(id, 'TSConditionalType');
            this.set(result, 'checkType', this.converted(this.child(id, 0), false, id));
            this.set(result, 'extendsType', this.converted(this.child(id, 1), false, id));
            this.set(result, 'falseType', this.converted(this.child(id, 3), false, id));
            this.set(result, 'trueType', this.converted(this.child(id, 2), false, id));
            return result;
        }
        if(node.kind === 'AsExpression') {
            const result = this.create(id, 'TSAsExpression');
            this.set(result, 'expression', this.converted(this.child(id, 0), false, id));
            this.set(result, 'typeAnnotation', this.converted(this.child(id, 1), false, id));
            return result;
        }
        if(node.kind === 'SatisfiesExpression') {
            const result = this.create(id, 'TSSatisfiesExpression');
            this.set(result, 'expression', this.converted(this.child(id, 0), false, id));
            this.set(result, 'typeAnnotation', this.converted(this.child(id, 1), false, id));
            return result;
        }
        if(node.kind === 'TypeAssertionExpression') {
            const result = this.create(id, 'TSTypeAssertion');
            this.set(result, 'expression', this.converted(this.child(id, 1), false, id));
            this.set(result, 'typeAnnotation', this.converted(this.child(id, 0), false, id));
            return result;
        }
        if(node.kind === 'OptionalType') {
            const result = this.create(id, 'TSOptionalType');
            this.set(result, 'typeAnnotation', this.converted(this.child(id, 0), false, id));
            return result;
        }
        if(node.kind === 'RestType') {
            const result = this.create(id, 'TSRestType');
            this.set(result, 'typeAnnotation', this.converted(this.child(id, 0), false, id));
            return result;
        }
        if(node.kind === 'InferType') {
            const result = this.create(id, 'TSInferType');
            this.set(result, 'typeParameter', this.converted(this.child(id, 0), false, id));
            return result;
        }
        if(node.kind === 'Decorator') {
            const result = this.create(id, 'Decorator');
            this.set(result, 'expression', this.converted(this.child(id, 0), false, id));
            return result;
        }
        if(node.kind === 'ExternalModuleReference') {
            const result = this.create(id, 'TSExternalModuleReference');
            this.set(result, 'expression', this.converted(this.child(id, 0), false, id));
            return result;
        }
        if(node.kind === 'NamespaceExportDeclaration') {
            const result = this.create(id, 'TSNamespaceExportDeclaration');
            this.set(result, 'id', this.converted(this.child(id, 0), false, id));
            return result;
        }
        if(node.kind === 'ImportAttribute') {
            const result = this.create(id, 'ImportAttribute');
            this.set(result, 'key', this.converted(this.child(id, 0), false, id));
            this.set(result, 'value', this.converted(this.child(id, 1), false, id));
            return result;
        }
        if(node.kind === 'EnumMember') {
            const result = this.create(id, 'TSEnumMember');
            this.set(result, 'id', this.converted(this.child(id, 0), false, id));
            this.set(result, 'initializer', this.converted(this.child(id, 1), false, id));
            return result;
        }
        if(node.kind === 'AnyKeyword') {
            return this.create(id, 'TSAnyKeyword');
        }
        if(node.kind === 'BigIntKeyword') {
            return this.create(id, 'TSBigIntKeyword');
        }
        if(node.kind === 'BooleanKeyword') {
            return this.create(id, 'TSBooleanKeyword');
        }
        if(node.kind === 'NeverKeyword') {
            return this.create(id, 'TSNeverKeyword');
        }
        if(node.kind === 'NumberKeyword') {
            return this.create(id, 'TSNumberKeyword');
        }
        if(node.kind === 'ObjectKeyword') {
            return this.create(id, 'TSObjectKeyword');
        }
        if(node.kind === 'StringKeyword') {
            return this.create(id, 'TSStringKeyword');
        }
        if(node.kind === 'SymbolKeyword') {
            return this.create(id, 'TSSymbolKeyword');
        }
        if(node.kind === 'UnknownKeyword') {
            return this.create(id, 'TSUnknownKeyword');
        }
        if(node.kind === 'VoidKeyword') {
            return this.create(id, 'TSVoidKeyword');
        }
        if(node.kind === 'UndefinedKeyword') {
            return this.create(id, 'TSUndefinedKeyword');
        }
        if(node.kind === 'IntrinsicKeyword') {
            return this.create(id, 'TSIntrinsicKeyword');
        }
        if(node.kind === 'AbstractKeyword') {
            return this.create(id, 'TSAbstractKeyword');
        }
        if(node.kind === 'LiteralType') {
            if(this.kind(first) === 'NullKeyword') {
                return this.create(first, 'TSNullKeyword');
            }
            const result = this.create(id, 'TSLiteralType');
            this.set(result, 'literal', this.converted(first));
            return result;
        }
        if(
            node.kind === 'UnionType' ||
            node.kind === 'IntersectionType' ||
            node.kind === 'TupleType' ||
            node.kind === 'TypeLiteral'
        ) {
            const result = this.create(
                id,
                node.kind === 'UnionType'
                    ? 'TSUnionType'
                    : node.kind === 'IntersectionType'
                      ? 'TSIntersectionType'
                      : node.kind === 'TupleType'
                        ? 'TSTupleType'
                        : 'TSTypeLiteral',
            );
            this.set(
                result,
                node.kind === 'TupleType' ? 'elementTypes' : node.kind === 'TypeLiteral' ? 'members' : 'types',
                this.list(node.children, false, id),
            );
            return result;
        }
        if(node.kind === 'TypeOperator') {
            const result = this.create(id, 'TSTypeOperator');
            this.set(
                result,
                'operator',
                stringValue(this.spellings.get(node.operator) ?? panic('unknown type operator')),
            );
            this.set(result, 'typeAnnotation', this.converted(first));
            return result;
        }
        return panic(`ESTree conversion not yet represented: ${node.kind} at ${this.start(id)}`);
    }
    template(id: number, tail: boolean): number {
        const result = this.create(id, 'TemplateElement');
        this.set(result, 'tail', boolValue(tail));
        this.set(result, 'value', templateValue(this.ts(id).raw, this.ts(id).text, true));
        return result;
    }
}
