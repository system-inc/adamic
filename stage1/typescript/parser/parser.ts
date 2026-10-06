// Expression descent follows cohere/TypeScript/tsc/internal/parser/parser.go.
import { panic } from 'adamic';
import { Scanner } from '../scanner/scanner.ts';
import { ParseNode } from './nodes.ts';

export function precedence(kind: string): number {
    switch(kind) {
        case 'QuestionQuestionToken':
            return 4;
        case 'BarBarToken':
            return 5;
        case 'AmpersandAmpersandToken':
            return 6;
        case 'BarToken':
            return 7;
        case 'CaretToken':
            return 8;
        case 'AmpersandToken':
            return 9;
        case 'EqualsEqualsToken':
        case 'ExclamationEqualsToken':
        case 'EqualsEqualsEqualsToken':
        case 'ExclamationEqualsEqualsToken':
            return 10;
        case 'LessThanToken':
        case 'GreaterThanToken':
        case 'LessThanEqualsToken':
        case 'GreaterThanEqualsToken':
        case 'InstanceOfKeyword':
        case 'InKeyword':
        case 'AsKeyword':
        case 'SatisfiesKeyword':
            return 11;
        case 'LessThanLessThanToken':
        case 'GreaterThanGreaterThanToken':
        case 'GreaterThanGreaterThanGreaterThanToken':
            return 12;
        case 'PlusToken':
        case 'MinusToken':
            return 13;
        case 'AsteriskToken':
        case 'SlashToken':
        case 'PercentToken':
            return 14;
        case 'AsteriskAsteriskToken':
            return 15;
        default:
            return -1;
    }
}

export class Parser {
    readonly scanner: Scanner;
    readonly nodes: ParseNode[] = [];
    readonly roots: number[] = [];
    constructor(text: string) {
        this.scanner = new Scanner(text);
        this.next();
    }
    kind(): string {
        return this.scanner.kind;
    }
    next(): void {
        this.scanner.scan();
    }
    node(index: number): ParseNode {
        return this.nodes[index] ?? panic('missing node');
    }
    make(kind: string, pos: number, children: number[] = []): number {
        this.nodes.push(new ParseNode(kind, pos, this.scanner.fullStart, children));
        return this.nodes.length - 1;
    }
    token(): number {
        const id = this.make(this.kind(), this.scanner.fullStart);
        this.next();
        this.node(id).end = this.scanner.fullStart;
        return id;
    }
    expect(kind: string): void {
        if(this.kind() !== kind) {
            panic(`parser slice expected ${kind}, got ${this.kind()} at ${this.scanner.start}`);
        }
        this.next();
    }
    literal(): number {
        const id = this.make(this.kind(), this.scanner.fullStart);
        const node = this.node(id);
        node.text = this.scanner.value;
        node.literalFlags =
            this.scanner.flags &
            (this.kind() === 'StringLiteral'
                ? 72716
                : this.kind() === 'RegularExpressionLiteral'
                  ? 4
                  : this.kind() === 'NoSubstitutionTemplateLiteral'
                    ? 7180
                    : 25584);
        this.next();
        node.end = this.scanner.fullStart;
        return id;
    }
    identifier(): number {
        const pos = this.scanner.fullStart;
        const text = this.scanner.value;
        this.next();
        const id = this.make('Identifier', pos);
        this.node(id).text = text;
        return id;
    }
    primary(): number {
        const pos = this.scanner.fullStart;
        switch(this.kind()) {
            case 'NumericLiteral':
            case 'BigIntLiteral':
            case 'StringLiteral':
            case 'NoSubstitutionTemplateLiteral':
                return this.literal();
            case 'SlashToken':
            case 'SlashEqualsToken':
                this.scanner.rescanSlash();
                return this.literal();
            case 'ThisKeyword':
            case 'SuperKeyword':
            case 'NullKeyword':
            case 'TrueKeyword':
            case 'FalseKeyword':
                return this.token();
            case 'PrivateIdentifier': {
                const id = this.identifier();
                const old = this.node(id);
                this.nodes[id] = new ParseNode('PrivateIdentifier', old.pos, old.end, []);
                this.node(id).text = old.text;
                return id;
            }
            case 'OpenBracketToken':
                return this.array();
            case 'OpenBraceToken':
                return this.object();
            case 'TemplateHead':
                return this.template();
            case 'NewKeyword': {
                this.next();
                if(this.kind() === 'DotToken') {
                    this.next();
                    const name = this.identifier();
                    const id = this.make('MetaProperty', pos, [name]);
                    this.node(id).operator = 'NewKeyword';
                    return id;
                }
                const target = this.suffix(this.primary(), false);
                const children = [target];
                let list = -1;
                let trailing = false;
                if(this.kind() === 'OpenParenToken') {
                    const args = this.arguments();
                    for(const argument of args) {
                        children.push(argument);
                    }
                    list = args.length;
                    trailing = this.lastTrailing;
                }
                const id = this.make('NewExpression', pos, children);
                this.node(id).list = list;
                this.node(id).trailing = trailing;
                return id;
            }
            case 'ImportKeyword': {
                if(this.peek() === 'DotToken') {
                    this.next();
                    this.next();
                    const name = this.identifier();
                    const id = this.make('MetaProperty', pos, [name]);
                    this.node(id).operator = 'ImportKeyword';
                    return id;
                }
                return this.token();
            }
            case 'OpenParenToken': {
                this.next();
                const expression = this.expression();
                this.expect('CloseParenToken');
                return this.make('ParenthesizedExpression', pos, [expression]);
            }
            default:
                if(this.kind() === 'Identifier' || this.kind().endsWith('Keyword')) {
                    return this.identifier();
                }
                panic(`parser slice unsupported primary ${this.kind()} at ${this.scanner.start}`);
        }
    }
    lastTrailing = false;
    peek(): string {
        const pos = this.scanner.pos;
        const start = this.scanner.start;
        const fullStart = this.scanner.fullStart;
        const value = this.scanner.value;
        const kind = this.kind();
        const flags = this.scanner.flags;
        const errors = this.scanner.errors.length;
        this.next();
        const result = this.kind();
        this.scanner.pos = pos;
        this.scanner.start = start;
        this.scanner.fullStart = fullStart;
        this.scanner.value = value;
        this.scanner.kind = kind;
        this.scanner.flags = flags;
        this.scanner.errors.splice(errors);
        return result;
    }
    optionalChain(index: number): boolean {
        const node = this.node(index);
        if(node.optional) {
            return true;
        }
        if(node.kind !== 'NonNullExpression') {
            return false;
        }
        const child = node.children[0] ?? panic('non-null expression without child');
        if(this.optionalChain(child)) {
            node.optional = true;
            return true;
        }
        return false;
    }
    spread(): number {
        const pos = this.scanner.fullStart;
        this.next();
        const expression = this.assignment();
        return this.make('SpreadElement', pos, [expression]);
    }
    arguments(): number[] {
        this.expect('OpenParenToken');
        const result: number[] = [];
        let trailing = false;
        while(this.kind() !== 'CloseParenToken') {
            result.push(this.kind() === 'DotDotDotToken' ? this.spread() : this.assignment());
            trailing = this.kind() === 'CommaToken';
            if(!trailing) {
                break;
            }
            this.next();
        }
        this.expect('CloseParenToken');
        this.lastTrailing = trailing;
        return result;
    }
    suffix(expression: number, allowCall: boolean): number {
        let left = expression;
        const pos = this.node(left).pos;
        while(true) {
            let question = -1;
            if(this.kind() === 'QuestionDotToken') {
                const after = this.peek();
                if(!allowCall && after === 'OpenParenToken') {
                    break;
                }
                question = this.token();
            }
            if(
                this.kind() === 'DotToken' ||
                (question >= 0 &&
                    (this.kind() === 'Identifier' ||
                        this.kind().endsWith('Keyword') ||
                        this.kind() === 'PrivateIdentifier'))
            ) {
                if(question < 0) {
                    this.next();
                }
                const children = [left];
                if(question >= 0) {
                    children.push(question);
                }
                if(this.kind() === 'PrivateIdentifier') {
                    const id = this.primary();
                    children.push(id);
                }
                else {
                    children.push(this.identifier());
                }
                const optional = question >= 0 || this.optionalChain(left);
                left = this.make('PropertyAccessExpression', pos, children);
                this.node(left).optional = optional;
                continue;
            }
            if(this.kind() === 'OpenBracketToken') {
                this.next();
                const argument = this.expression();
                this.expect('CloseBracketToken');
                const children = [left];
                if(question >= 0) {
                    children.push(question);
                }
                children.push(argument);
                const optional = question >= 0 || this.optionalChain(left);
                left = this.make('ElementAccessExpression', pos, children);
                this.node(left).optional = optional;
                continue;
            }
            if(allowCall && this.kind() === 'OpenParenToken') {
                const args = this.arguments();
                const trailing = this.lastTrailing;
                const children = [left];
                if(question >= 0) {
                    children.push(question);
                }
                for(const argument of args) {
                    children.push(argument);
                }
                const optional = question >= 0 || this.optionalChain(left);
                left = this.make('CallExpression', pos, children);
                this.node(left).optional = optional;
                this.node(left).list = args.length;
                this.node(left).trailing = trailing;
                continue;
            }
            if(this.kind() === 'NoSubstitutionTemplateLiteral' || this.kind() === 'TemplateHead') {
                const template = this.kind() === 'TemplateHead' ? this.template() : this.literal();
                const children = [left];
                if(question >= 0) {
                    children.push(question);
                }
                children.push(template);
                const optional = question >= 0 || this.node(left).optional;
                left = this.make('TaggedTemplateExpression', pos, children);
                this.node(left).optional = optional;
                continue;
            }
            if(question >= 0) {
                panic('optional chain missing member');
            }
            if(this.kind() === 'ExclamationToken' && (this.scanner.flags & 1) === 0) {
                this.next();
                left = this.make('NonNullExpression', pos, [left]);
                continue;
            }
            return left;
        }
        return left;
    }
    templatePart(): number {
        const id = this.make(this.kind(), this.scanner.fullStart);
        const node = this.node(id);
        node.text = this.scanner.value;
        node.literalFlags = this.scanner.flags & 7180;
        node.raw = this.scanner.text.slice(
            this.scanner.start + 1,
            this.scanner.pos - ((this.scanner.flags & 4) !== 0 ? 0 : this.kind() === 'TemplateTail' ? 1 : 2),
        );
        this.next();
        node.end = this.scanner.fullStart;
        return id;
    }
    template(): number {
        const pos = this.scanner.fullStart;
        const children = [this.templatePart()];
        while(true) {
            const spanPos = this.scanner.fullStart;
            const expression = this.expression();
            if(this.kind() !== 'CloseBraceToken') {
                panic('template missing closing brace');
            }
            this.scanner.rescanTemplate();
            const tail = this.kind() === 'TemplateTail';
            const literal = this.templatePart();
            children.push(this.make('TemplateSpan', spanPos, [expression, literal]));
            if(tail) {
                break;
            }
        }
        return this.make('TemplateExpression', pos, children);
    }
    array(): number {
        const pos = this.scanner.fullStart;
        this.next();
        const multiLine = (this.scanner.flags & 1) !== 0;
        const children: number[] = [];
        let trailing = false;
        while(this.kind() !== 'CloseBracketToken') {
            children.push(
                this.kind() === 'CommaToken'
                    ? this.make('OmittedExpression', this.scanner.fullStart)
                    : this.kind() === 'DotDotDotToken'
                      ? this.spread()
                      : this.assignment(),
            );
            trailing = this.kind() === 'CommaToken';
            if(!trailing) {
                break;
            }
            this.next();
        }
        this.expect('CloseBracketToken');
        const id = this.make('ArrayLiteralExpression', pos, children);
        this.node(id).list = children.length;
        this.node(id).trailing = trailing;
        this.node(id).multiLine = multiLine;
        return id;
    }
    propertyName(): number {
        if(this.kind() === 'StringLiteral' || this.kind() === 'NumericLiteral') {
            return this.literal();
        }
        if(this.kind() === 'OpenBracketToken') {
            const pos = this.scanner.fullStart;
            this.next();
            const expression = this.expression();
            this.expect('CloseBracketToken');
            return this.make('ComputedPropertyName', pos, [expression]);
        }
        return this.identifier();
    }
    object(): number {
        const pos = this.scanner.fullStart;
        this.next();
        const multiLine = (this.scanner.flags & 1) !== 0;
        const properties: number[] = [];
        let trailing = false;
        while(this.kind() !== 'CloseBraceToken') {
            const start = this.scanner.fullStart;
            if(this.kind() === 'DotDotDotToken') {
                this.next();
                const expression = this.assignment();
                properties.push(this.make('SpreadAssignment', start, [expression]));
            }
            else {
                const name = this.propertyName();
                const children = [name];
                if(this.kind() === 'QuestionToken' || this.kind() === 'ExclamationToken') {
                    children.push(this.token());
                }
                const assignment = this.kind() === 'ColonToken';
                if(assignment) {
                    this.next();
                    children.push(this.assignment());
                }
                else if(this.kind() === 'EqualsToken') {
                    children.push(this.token());
                    children.push(this.assignment());
                }
                properties.push(
                    this.make(assignment ? 'PropertyAssignment' : 'ShorthandPropertyAssignment', start, children),
                );
            }
            trailing = this.kind() === 'CommaToken';
            if(!trailing) {
                break;
            }
            this.next();
        }
        this.expect('CloseBraceToken');
        const id = this.make('ObjectLiteralExpression', pos, properties);
        this.node(id).list = properties.length;
        this.node(id).trailing = trailing;
        this.node(id).multiLine = multiLine;
        return id;
    }
    unary(): number {
        const pos = this.scanner.fullStart;
        const operator = this.kind();
        if(
            operator === 'PlusToken' ||
            operator === 'MinusToken' ||
            operator === 'TildeToken' ||
            operator === 'ExclamationToken' ||
            operator === 'PlusPlusToken' ||
            operator === 'MinusMinusToken'
        ) {
            this.next();
            const operand = this.unary();
            const id = this.make('PrefixUnaryExpression', pos, [operand]);
            this.node(id).operator = operator;
            return id;
        }
        if(operator === 'DeleteKeyword' || operator === 'VoidKeyword' || operator === 'TypeOfKeyword') {
            this.next();
            const operand = this.unary();
            return this.make(
                operator === 'DeleteKeyword'
                    ? 'DeleteExpression'
                    : operator === 'VoidKeyword'
                      ? 'VoidExpression'
                      : 'TypeOfExpression',
                pos,
                [operand],
            );
        }
        const operand = this.suffix(this.primary(), true);
        const postfix = this.kind();
        if((postfix === 'PlusPlusToken' || postfix === 'MinusMinusToken') && (this.scanner.flags & 1) === 0) {
            this.next();
            const id = this.make('PostfixUnaryExpression', pos, [operand]);
            this.node(id).operator = postfix;
            return id;
        }
        return operand;
    }
    binary(minimum: number): number {
        const pos = this.scanner.fullStart;
        let left = this.unary();
        while(true) {
            this.scanner.rescanGreater();
            const operator = this.kind();
            const rank = precedence(operator);
            if(rank < minimum || (rank === minimum && operator !== 'AsteriskAsteriskToken')) {
                break;
            }
            const token = this.token();
            const right = this.binary(rank);
            left = this.make('BinaryExpression', pos, [left, token, right]);
        }
        return left;
    }
    assignment(): number {
        const pos = this.scanner.fullStart;
        let left = this.binary(0);
        const operator = this.kind();
        if(operator === 'EqualsToken' || (operator.endsWith('EqualsToken') && precedence(operator) < 0)) {
            const token = this.token();
            const right = this.assignment();
            return this.make('BinaryExpression', pos, [left, token, right]);
        }
        if(operator === 'QuestionToken') {
            const question = this.token();
            const yes = this.assignment();
            if(this.kind() !== 'ColonToken') {
                panic('conditional missing colon');
            }
            const colon = this.token();
            const no = this.assignment();
            left = this.make('ConditionalExpression', pos, [left, question, yes, colon, no]);
        }
        return left;
    }
    expression(): number {
        const pos = this.scanner.fullStart;
        let left = this.assignment();
        while(this.kind() === 'CommaToken') {
            const comma = this.token();
            const right = this.assignment();
            left = this.make('BinaryExpression', pos, [left, comma, right]);
        }
        return left;
    }
    file(): void {
        while(this.kind() !== 'EndOfFile') {
            if(this.kind() === 'SemicolonToken') {
                this.next();
                continue;
            }
            this.roots.push(this.expression());
            if(this.kind() === 'SemicolonToken') {
                this.next();
            }
            else if(this.kind() !== 'EndOfFile' && (this.scanner.flags & 1) === 0) {
                panic(`parser slice unsupported statement at ${this.scanner.start}`);
            }
        }
    }
}
