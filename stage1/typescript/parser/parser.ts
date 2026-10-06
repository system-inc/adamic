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
        const id = this.make(this.scanner.kind, this.scanner.fullStart);
        this.next();
        this.node(id).end = this.scanner.fullStart;
        return id;
    }
    expect(kind: string): void {
        if(this.scanner.kind !== kind) {
            panic(`parser slice expected ${kind}, got ${this.scanner.kind} at ${this.scanner.start}`);
        }
        this.next();
    }
    literal(): number {
        const id = this.make(this.scanner.kind, this.scanner.fullStart);
        const node = this.node(id);
        node.text = this.scanner.value;
        node.literalFlags =
            this.scanner.flags &
            (this.scanner.kind === 'StringLiteral'
                ? 72716
                : this.scanner.kind === 'RegularExpressionLiteral'
                  ? 4
                  : this.scanner.kind === 'NoSubstitutionTemplateLiteral'
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
        switch(this.scanner.kind) {
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
            case 'OpenParenToken': {
                this.next();
                const expression = this.expression();
                this.expect('CloseParenToken');
                return this.make('ParenthesizedExpression', pos, [expression]);
            }
            default:
                if(this.scanner.kind === 'Identifier' || this.scanner.kind.endsWith('Keyword')) {
                    return this.identifier();
                }
                panic(`parser slice unsupported primary ${this.scanner.kind} at ${this.scanner.start}`);
        }
    }
    unary(): number {
        const pos = this.scanner.fullStart;
        const operator = this.scanner.kind;
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
        const operand = this.primary();
        const postfix = this.scanner.kind;
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
        for(;;) {
            this.scanner.rescanGreater();
            const operator = this.scanner.kind;
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
        const operator = this.scanner.kind;
        if(operator === 'EqualsToken' || (operator.endsWith('EqualsToken') && precedence(operator) < 0)) {
            const token = this.token();
            const right = this.assignment();
            return this.make('BinaryExpression', pos, [left, token, right]);
        }
        if(operator === 'QuestionToken') {
            const question = this.token();
            const yes = this.assignment();
            if(this.scanner.kind !== 'ColonToken') {
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
        while(this.scanner.kind === 'CommaToken') {
            const comma = this.token();
            const right = this.assignment();
            left = this.make('BinaryExpression', pos, [left, comma, right]);
        }
        return left;
    }
    file(): void {
        while(this.scanner.kind !== 'EndOfFile') {
            if(this.scanner.kind === 'SemicolonToken') {
                this.next();
                continue;
            }
            this.roots.push(this.expression());
            if(this.scanner.kind === 'SemicolonToken') {
                this.next();
            }
            else if(this.scanner.kind !== 'EndOfFile' && (this.scanner.flags & 1) === 0) {
                panic(`parser slice unsupported statement at ${this.scanner.start}`);
            }
        }
    }
}
