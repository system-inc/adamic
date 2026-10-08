// Default production cohere judgments. The external library supplies facts only.
import { panic, tsgoInspect } from 'adamic';
import type { Parser } from '../../typescript/parser/parser.ts';
import type { Scanner } from '../../typescript/scanner/scanner.ts';
import type { ParseNode } from '../../typescript/parser/nodes.ts';
import { isIdentifierStart, isIdentifierPart } from '../../typescript/scanner/characters.ts';
import { Diagnostic } from './diagnostic.ts';
import { Frames, header } from './frames.ts';
import { types } from './facts.ts';
import { unionFlag, intersectionFlag, numberLike, bigintLike, stringLike, booleanLike, nullish } from './flags.ts';
import type { TypeFact } from './type_fact.ts';
import type { Types } from './types.ts';
import { Parameters } from './parameters.ts';
import type { UnaryMinus } from './unary_minus.ts';

const getterMessage =
    "A getter and its setter name one property, so a caller that reads the property and writes the value straight back has to type-check. The getter here returns a type the setter will not accept, which makes that round trip an error and usually means one of the two annotations is wrong. Widen the setter's parameter, or narrow what the getter returns.";
const mergingMessage =
    'A class and an interface sharing a name are merged into one type, and the members the interface contributes are never initialized by the class constructor. TypeScript does not check them, so reading one compiles and returns undefined at runtime. Give the interface a different name, or declare the members on the class.';
const plusAllowed = 'string, allowing a string + any of: `any`, `boolean`, `null`, `RegExp`, `undefined`';

export function unsafeAssignment(
    senderFacts: Types,
    sender: TypeFact,
    receiverFacts: Types,
    receiver: TypeFact,
    emptyMap: boolean,
    visited: string[],
): boolean {
    if((sender.flags & 1) !== 0 && (receiver.flags & (1 | 2)) === 0) {
        return true;
    }
    const pair = `${sender.id}:${receiver.id}`;
    if(visited.includes(pair)) {
        return false;
    }
    const searched = visited.slice();
    searched.push(pair);
    if(sender.target === 0 || receiver.target === 0 || sender.target !== receiver.target || emptyMap) {
        return false;
    }
    if(sender.arguments.length === 0 || receiver.arguments.length === 0) {
        return false;
    }
    if(sender.arguments.length !== receiver.arguments.length) {
        panic('incompatible checker generic arguments');
    }
    for(let index = 0; index < sender.arguments.length; index++) {
        if(
            unsafeAssignment(
                senderFacts,
                senderFacts.type(sender.arguments[index] ?? 0),
                receiverFacts,
                receiverFacts.type(receiver.arguments[index] ?? 0),
                emptyMap,
                searched,
            )
        ) {
            return true;
        }
    }
    return false;
}

export class Rules {
    inspectLocation: ((path: string, start: number, end: number, kind: string, question: string) => string) | undefined = undefined;
    inspect: ((index: number, question: string) => string) | undefined = undefined;
    readonly program: number;
    readonly path: string;
    readonly parser: Parser;
    readonly scanner: Scanner;
    readonly offsets: readonly number[];
    readonly unary: UnaryMinus;
    readonly parents: number[] = [];
    readonly findings: Diagnostic[] = [];
    queries = 0;
    constructor(
        program: number,
        path: string,
        parser: Parser,
        scanner: Scanner,
        offsets: readonly number[],
        unary: UnaryMinus,
    ) {
        this.program = program;
        this.path = path;
        this.parser = parser;
        this.scanner = scanner;
        this.offsets = offsets;
        this.unary = unary;
    }
    byte(pos: number): number {
        return this.offsets[pos] ?? panic('position outside source');
    }
    start(node: ParseNode): number {
        this.scanner.pos = node.pos;
        this.scanner.scan();
        return this.scanner.start;
    }
    text(node: ParseNode): string {
        return this.scanner.text.slice(this.start(node), node.end);
    }
    ask(index: number, question: string): string {
        if(this.inspect !== undefined) { return this.inspect(index, question); }
        const node = this.parser.node(index);
        this.queries++;
        return tsgoInspect(this.program, this.path, this.byte(node.pos), this.byte(node.end), node.kind, question);
    }
    name(type: TypeFact, index: number): string {
        const frames = new Frames(this.ask(index, `name\n${type.id}`));
        header(frames, 'name');
        const name = frames.field();
        frames.end();
        return name;
    }
    describe(type: TypeFact, index: number): string {
        return type.error ? 'error typed' : `\`${this.name(type, index)}\``;
    }
    add(rule: string, id: string, message: string, index: number): Diagnostic {
        const node = this.parser.node(index);
        const finding = new Diagnostic(rule, id, message, this.byte(this.start(node)), this.byte(node.end));
        this.findings.push(finding);
        return finding;
    }
    skip(index: number, kind = 'ParenthesizedExpression'): number {
        let current = index;
        while(this.parser.node(current).kind === kind) {
            current = this.parser.node(current).children[0] ?? panic('empty parentheses');
        }
        return current;
    }
    links(index: number, parent: number): void {
        this.parents[index] = parent;
        for(const child of this.parser.node(index).children) {
            this.links(child, index);
        }
    }
    key(index: number): string {
        let node = this.parser.node(index);
        if(node.kind === 'ComputedPropertyName') {
            node = this.parser.node(node.children[0] ?? panic('empty accessor key'));
        }
        if(node.kind === 'Identifier' || node.kind === 'PrivateIdentifier') {
            return node.text;
        }
        if(node.kind === 'StringLiteral' || node.kind === 'NumericLiteral') {
            let valid = node.text.length > 0;
            for(let offset = 0; offset < node.text.length; offset++) {
                const code = node.text.codePointAt(offset) ?? 0;
                if(!(offset === 0 ? isIdentifierStart(code) : isIdentifierPart(code))) {
                    valid = false;
                }
                if(code > 65535) {
                    offset++;
                }
            }
            return valid ? node.text : `"${node.text}"`;
        }
        return this.text(node);
    }
    getterPairs(index: number): void {
        const names: string[] = [];
        const getters: number[] = [];
        const setters: number[] = [];
        for(const memberIndex of this.parser.node(index).children) {
            const member = this.parser.node(memberIndex);
            if(member.kind !== 'GetAccessor' && member.kind !== 'SetAccessor') {
                continue;
            }
            if(member.children.some((child) => this.parser.node(child).kind === 'AbstractKeyword')) {
                continue;
            }
            const parameters = member.children.filter((child) => this.parser.node(child).kind === 'Parameter');
            const name =
                member.children.find(
                    (child) =>
                        !this.parser.node(child).kind.endsWith('Keyword') &&
                        this.parser.node(child).kind !== 'Decorator',
                ) ?? panic('accessor without name');
            const annotation = this.annotation(memberIndex);
            if(
                (member.kind === 'GetAccessor' && annotation < 0) ||
                (member.kind === 'SetAccessor' && parameters.length !== 1)
            ) {
                continue;
            }
            const key = this.key(name);
            let at = names.indexOf(key);
            if(at < 0) {
                at = names.length;
                names.push(key);
                getters.push(-1);
                setters.push(-1);
            }
            if(member.kind === 'GetAccessor') {
                getters[at] = memberIndex;
            }
            else {
                setters[at] = parameters[0] ?? -1;
            }
        }
        for(let at = 0; at < names.length; at++) {
            const getter = getters[at] ?? -1;
            const setter = setters[at] ?? -1;
            if(getter < 0 || setter < 0) {
                continue;
            }
            const target = this.parser.node(setter);
            const frames = new Frames(
                this.ask(getter, `assignable\n${this.byte(target.pos)}\n${this.byte(target.end)}\n${target.kind}`),
            );
            header(frames, 'assignable');
            const assignable = frames.yes();
            frames.end();
            if(!assignable) {
                this.add(
                    'related-getter-setter-pairs',
                    'mismatch',
                    getterMessage,
                    this.skip(this.annotation(getter), 'ParenthesizedType'),
                );
            }
        }
    }
    annotation(index: number): number {
        // Accessor children are modifiers/name, parameters, optional return type, body.
        const member = this.parser.node(index);
        const name =
            member.children.find(
                (child) =>
                    !this.parser.node(child).kind.endsWith('Keyword') && this.parser.node(child).kind !== 'Decorator',
            ) ?? -1;
        const afterName = member.children.indexOf(name) + 1;
        for(let at = afterName; at < member.children.length; at++) {
            const child = member.children[at] ?? -1;
            const kind = this.parser.node(child).kind;
            if(kind !== 'Parameter' && kind !== 'TypeParameter' && kind !== 'QuestionToken' && kind !== 'Block') {
                return child;
            }
        }
        return -1;
    }
    merging(index: number): void {
        const node = this.parser.node(index);
        const frames = new Frames(this.ask(index, 'declarations'));
        header(frames, 'declarations');
        const present = frames.yes();
        const count = frames.natural();
        let earliestPos = 9007199254740991;
        let earliestKind = '';
        let start = 0;
        let end = 0;
        for(let list = 0; list < 2; list++) {
            const localPresent = list === 0 ? present : frames.yes();
            const length = list === 0 ? count : frames.natural();
            const use = localPresent && (list === 0 || length > count);
            if(list === 1 && use) {
                earliestPos = 9007199254740991;
                earliestKind = '';
            }
            for(let at = 0; at < length; at++) {
                const kind = frames.field();
                const pos = frames.natural();
                frames.natural();
                frames.field();
                const nameStart = frames.natural();
                const nameEnd = frames.natural();
                if(use && pos < earliestPos) {
                    earliestPos = pos;
                    earliestKind = kind;
                    start = nameStart;
                    end = nameEnd;
                }
            }
        }
        frames.end();
        const opposite = node.kind === 'ClassDeclaration' ? 'InterfaceDeclaration' : 'ClassDeclaration';
        if(present && earliestKind === opposite) {
            this.findings.push(
                new Diagnostic('no-unsafe-declaration-merging', 'unsafeMerging', mergingMessage, start, end),
            );
        }
    }
    plus(index: number): void {
        const node = this.parser.node(index);
        const left = node.children[0] ?? panic('missing plus left');
        const right = node.children[2] ?? panic('missing plus right');
        const leftFacts = types(this.ask(left, 'base-type'), 'base-type');
        const rightFacts = types(this.ask(right, 'base-type'), 'base-type');
        const leftType = leftFacts.root();
        const rightType = rightFacts.root();
        if(leftType.id === rightType.id && (leftType.flags & (numberLike | bigintLike | stringLike)) !== 0) {
            return;
        }
        let complaint = false;
        for(let side = 0; side < 2; side++) {
            const facts = side === 0 ? leftFacts : rightFacts;
            const other = side === 0 ? rightFacts : leftFacts;
            const operand = side === 0 ? left : right;
            const root = facts.root();
            if(facts.has(root, 512 | 16384 | 262144 | 2)) {
                this.plusInvalid(operand, root.name);
                complaint = true;
                continue;
            }
            for(const partID of facts.parts(root)) {
                const part = facts.type(partID);
                const parts = facts.parts(part, (part.flags & intersectionFlag) !== 0 ? intersectionFlag : unionFlag);
                const invalid =
                    part.name === 'RegExp'
                        ? other.has(other.root(), numberLike)
                        : parts.length > 0 && parts.every((id) => (facts.type(id).flags & 1048576) !== 0);
                if(invalid) {
                    this.plusInvalid(operand, part.name);
                    complaint = true;
                }
            }
        }
        if(complaint) {
            return;
        }
        if(
            (leftFacts.has(leftType, numberLike) && rightFacts.has(rightType, bigintLike)) ||
            (rightFacts.has(rightType, numberLike) && leftFacts.has(leftType, bigintLike))
        ) {
            this.add(
                'restrict-plus-operands',
                'bigintAndNumber',
                `Numeric '+' operations must either be both bigints or both numbers. Got \`${leftType.name}\` + \`${rightType.name}\`.`,
                index,
            );
        }
    }
    plusInvalid(index: number, name: string): void {
        this.add(
            'restrict-plus-operands',
            'invalid',
            `Invalid operand for a '+' operation. Operands must each be a number or ${plusAllowed}. Got \`${name}\`.`,
            index,
        );
    }
    arguments(index: number): void {
        const node = this.parser.node(index);
        const arguments_: number[] = [];
        if(node.kind === 'TaggedTemplateExpression') {
            const template = this.parser.node(node.children[node.children.length - 1] ?? panic('missing template'));
            if(template.kind !== 'TemplateExpression') {
                return;
            }
            for(const span of template.children.slice(1)) {
                arguments_.push(this.parser.node(span).children[0] ?? panic('missing interpolation'));
            }
        }
        else {
            if(node.list <= 0) {
                return;
            }
            for(const argument of node.children.slice(node.children.length - node.list)) {
                arguments_.push(argument);
            }
        }
        if(arguments_.length === 0) {
            return;
        }
        const signature = types(this.ask(index, 'signature-shape'), 'signature-shape');
        if(!signature.present || (signature.root().flags & 1) !== 0) {
            return;
        }
        const fixed: number[] = [];
        let rest = 0;
        let restIndex = 0;
        for(let at = 0; at < signature.rests.length; at++) {
            const root = signature.roots[at + 1] ?? panic('missing signature parameter');
            if(signature.rests[at] === true) {
                rest = root;
                restIndex = at;
                break;
            }
            fixed.push(root);
        }
        const parameters = new Parameters(signature, fixed, rest, restIndex);
        if(node.kind === 'TaggedTemplateExpression') {
            parameters.next();
        }
        for(const argumentIndex of arguments_) {
            const argument = this.parser.node(argumentIndex);
            if(argument.kind === 'SpreadElement') {
                const inner = argument.children[0] ?? panic('empty spread');
                const argumentFacts = types(this.ask(inner, 'raw-shape'), 'raw-shape');
                const sender = argumentFacts.root();
                if((sender.flags & 1) !== 0) {
                    this.add(
                        'no-unsafe-argument',
                        'unsafeSpread',
                        `Unsafe spread of an ${this.describe(sender, argumentIndex)} type.`,
                        argumentIndex,
                    );
                }
                else if(
                    sender.array &&
                    sender.arguments.length > 0 &&
                    (argumentFacts.type(sender.arguments[0] ?? 0).flags & 1) !== 0
                ) {
                    const name = argumentFacts.type(sender.arguments[0] ?? 0).error
                        ? 'error'
                        : this.describe(sender, argumentIndex);
                    this.add(
                        'no-unsafe-argument',
                        'unsafeArraySpread',
                        `Unsafe spread of an ${name} array type.`,
                        argumentIndex,
                    );
                }
                else if(sender.tuple >= 0) {
                    for(const member of sender.arguments) {
                        const parameter = parameters.next();
                        if(parameter === 0) {
                            continue;
                        }
                        const type = argumentFacts.type(member);
                        const receiver = signature.type(parameter);
                        if(unsafeAssignment(argumentFacts, type, signature, receiver, false, [])) {
                            this.add(
                                'no-unsafe-argument',
                                'unsafeTupleSpread',
                                `Unsafe spread of a tuple type. The argument is ${type.error ? 'error typed' : `of type \`${this.name(type, argumentIndex)}\``} and is assigned to a parameter of type ${this.describe(receiver, index)}.`,
                                argumentIndex,
                            );
                        }
                    }
                    if((sender.tuple & 12) !== 0) {
                        parameters.consumed = true;
                    }
                }
                continue;
            }
            const parameter = parameters.next();
            if(parameter === 0) {
                continue;
            }
            const argumentFacts = types(this.ask(argumentIndex, 'raw-shape'), 'raw-shape');
            const sender = argumentFacts.root();
            const receiver = signature.type(parameter);
            const callee = argument.children.length > 0 ? this.parser.node(argument.children[0] ?? -1) : argument;
            const emptyMap =
                argument.kind === 'NewExpression' &&
                callee.kind === 'Identifier' &&
                callee.text === 'Map' &&
                argument.list === 0 &&
                argument.children.length === 1;
            if(unsafeAssignment(argumentFacts, sender, signature, receiver, emptyMap, [])) {
                this.add(
                    'no-unsafe-argument',
                    'unsafeArgument',
                    `Unsafe argument of type ${this.describe(sender, argumentIndex)} assigned to a parameter of type ${this.describe(receiver, index)}.`,
                    argumentIndex,
                );
            }
        }
    }
    booleanCompare(index: number): void {
        const node = this.parser.node(index);
        const operator = this.parser.node(node.children[1] ?? panic('missing comparison operator')).kind;
        const positive = operator === 'EqualsEqualsToken' || operator === 'EqualsEqualsEqualsToken';
        if(!positive && operator !== 'ExclamationEqualsToken' && operator !== 'ExclamationEqualsEqualsToken') {
            return;
        }
        const left = node.children[0] ?? panic('missing comparison left');
        const right = node.children[2] ?? panic('missing comparison right');
        let literal = this.parser.node(this.skip(right));
        let expression = this.skip(left);
        if(literal.kind !== 'TrueKeyword' && literal.kind !== 'FalseKeyword') {
            literal = this.parser.node(this.skip(left));
            expression = this.skip(right);
        }
        if(literal.kind !== 'TrueKeyword' && literal.kind !== 'FalseKeyword') {
            return;
        }
        const facts = types(this.ask(expression, 'type-shape'), 'type-shape');
        const type = facts.root();
        const plain = (type.flags & booleanLike) !== 0;
        const parts = facts.parts(type);
        const survivors = parts.filter((id) => (facts.type(id).flags & nullish) === 0);
        const nullable =
            (type.flags & unionFlag) !== 0 &&
            survivors.length > 0 &&
            survivors.length < parts.length &&
            survivors.every((id) => (facts.type(id).flags & booleanLike) !== 0);
        // Both nullable comparison allowances are true in the production defaults.
        if(!plain || nullable) {
            return;
        }
        const message = positive
            ? 'This compares a boolean to a boolean literal, which asks the same question the value already answers. Use the expression on its own.'
            : 'This compares a boolean to a boolean literal to ask whether it is false, which the value already answers. Negate the expression instead.';
        const finding = this.add(
            'no-unnecessary-boolean-literal-compare',
            positive ? 'direct' : 'negated',
            message,
            index,
        );
        let outer = index;
        while(
            (this.parents[outer] ?? -1) >= 0 &&
            this.parser.node(this.parents[outer] ?? -1).kind === 'ParenthesizedExpression'
        ) {
            outer = this.parents[outer] ?? -1;
        }
        const parent = this.parents[outer] ?? -1;
        const wrapped =
            parent >= 0 &&
            this.parser.node(parent).kind === 'PrefixUnaryExpression' &&
            this.parser.node(parent).operator === 'ExclamationToken';
        const mutated = wrapped ? parent : index;
        const inversions = (wrapped ? 1 : 0) + (positive ? 0 : 1) + (literal.kind === 'TrueKeyword' ? 0 : 1);
        let text = this.text(this.parser.node(expression));
        let weak = !this.strong(expression);
        if(inversions % 2 === 1) {
            text = `!${weak ? `(${text})` : text}`;
            weak = false;
        }
        if(weak && this.weakParent(mutated)) {
            text = `(${text})`;
        }
        finding.fixStart = this.byte(this.start(this.parser.node(mutated)));
        finding.fixEnd = this.byte(this.parser.node(mutated).end);
        finding.replacement = text;
    }
    strong(index: number): boolean {
        return [
            'NumericLiteral',
            'BigIntLiteral',
            'StringLiteral',
            'RegularExpressionLiteral',
            'NoSubstitutionTemplateLiteral',
            'TrueKeyword',
            'FalseKeyword',
            'ParenthesizedExpression',
            'Identifier',
            'TypeReference',
            'TypeOperator',
            'ArrayLiteralExpression',
            'ObjectLiteralExpression',
            'PropertyAccessExpression',
            'ElementAccessExpression',
            'CallExpression',
            'NewExpression',
            'TaggedTemplateExpression',
            'ExpressionWithTypeArguments',
        ].includes(this.parser.node(index).kind);
    }
    weakParent(index: number): boolean {
        const parentIndex = this.parents[index] ?? -1;
        if(parentIndex < 0) {
            return false;
        }
        const parent = this.parser.node(parentIndex);
        if(
            [
                'PrefixUnaryExpression',
                'PostfixUnaryExpression',
                'BinaryExpression',
                'ConditionalExpression',
                'AwaitExpression',
            ].includes(parent.kind)
        ) {
            return true;
        }
        return (
            [
                'PropertyAccessExpression',
                'ElementAccessExpression',
                'CallExpression',
                'NewExpression',
                'TaggedTemplateExpression',
            ].includes(parent.kind) && parent.children[0] === index
        );
    }
    walk(index: number): void {
        const node = this.parser.node(index);
        const before = this.unary.findings.length;
        if(node.kind === 'PrefixUnaryExpression' && node.operator === 'MinusToken') {
            this.unary.visit(index);
        }
        for(let at = before; at < this.unary.findings.length; at++) {
            const unary = this.unary.findings[at] ?? panic('missing unary finding');
            this.findings.push(
                new Diagnostic('no-unsafe-unary-minus', unary.id, unary.message, unary.start, unary.end),
            );
        }
        if(['ClassDeclaration', 'ClassExpression', 'InterfaceDeclaration', 'TypeLiteral'].includes(node.kind)) {
            this.getterPairs(index);
        }
        if(node.kind === 'ClassDeclaration' || node.kind === 'InterfaceDeclaration') {
            this.merging(index);
        }
        if(['CallExpression', 'NewExpression', 'TaggedTemplateExpression'].includes(node.kind)) {
            this.arguments(index);
        }
        if(node.kind === 'BinaryExpression') {
            const operator = this.parser.node(node.children[1] ?? panic('missing binary operator')).kind;
            if(operator === 'PlusToken' || operator === 'PlusEqualsToken') {
                this.plus(index);
            }
            this.booleanCompare(index);
        }
        for(const child of node.children) {
            this.walk(child);
        }
    }
    run(): void {
        const root = this.parser.file();
        while(this.parents.length < this.parser.nodes.length) {
            this.parents.push(-1);
        }
        this.links(root, -1);
        const options = new Frames(this.ask(root, 'options'));
        header(options, 'options');
        const strict = options.yes();
        options.end();
        if(!strict) {
            this.findings.push(
                new Diagnostic(
                    'no-unnecessary-boolean-literal-compare',
                    'noStrictNullCheck',
                    'This rule needs the `strictNullChecks` compiler option to tell a boolean from a nullable one. Without it every nullable boolean looks like a plain boolean, so the rule would propose repairs that silently change what the code does on null.',
                    0,
                    0,
                ),
            );
        }
        this.walk(root);
        this.queries += this.unary.queries;
        // Fixes are complete now. Render sort keys once; written() remains live.
        for(const finding of this.findings) {
            finding.sortKey = finding.written();
        }
        this.findings.sort((left, right) => (left.sortKey < right.sortKey ? -1 : left.sortKey > right.sortKey ? 1 : 0));
    }
}
