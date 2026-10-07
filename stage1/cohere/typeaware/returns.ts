import type { Rules } from './rules.ts';
import type { Shadow } from './shadow.ts';
import { types } from './facts.ts';
import { Frames, header } from './frames.ts';
import type { Types } from './types.ts';
import type { TypeFact } from './type_fact.ts';
import { Completion } from './completion.ts';
import { Diagnostic } from './diagnostic.ts';

export class Returns {
    readonly rules: Rules;
    readonly shadow: Shadow;
    readonly pending: number[] = [];
    constructor(rules: Rules, shadow: Shadow) {
        this.rules = rules;
        this.shadow = shadow;
    }
    scope(index: number): boolean {
        return [
            'FunctionDeclaration',
            'FunctionExpression',
            'ArrowFunction',
            'MethodDeclaration',
            'GetAccessor',
            'SetAccessor',
            'Constructor',
        ].includes(this.rules.parser.node(index).kind);
    }
    count(index: number, id: number): number {
        const frames = new Frames(this.rules.ask(index, `call-count\n${id}`));
        header(frames, 'call-count');
        const count = frames.natural();
        frames.end();
        return count;
    }
    thenable(index: number, facts: Types, subject: TypeFact): boolean {
        const apparent = types(this.rules.ask(index, `apparent-shape\n${subject.id}`), 'apparent-shape');
        for(const part of apparent.parts(apparent.root())) {
            const property = types(this.rules.ask(index, `property-shape\n${part}\nthen`), 'property-shape');
            if(!property.present) {
                continue;
            }
            for(const then of property.parts(property.root())) {
                const parameters = types(this.rules.ask(index, `call-parameters\n${then}`), 'call-parameters');
                for(const parameter of parameters.roots) {
                    for(const callback of parameters.parts(parameters.type(parameter))) {
                        if(this.count(index, callback) > 0) {
                            return true;
                        }
                    }
                }
            }
        }
        return false;
    }
    promiseVoid(index: number, facts: Types, subject: TypeFact): boolean {
        let current = subject;
        for(let depth = 0; depth < 32; depth++) {
            if(current.target === 0 || current.arguments.length === 0 || !this.thenable(index, facts, current)) {
                return false;
            }
            current = facts.type(current.arguments[0] ?? 0);
            if(facts.has(current, 16)) {
                return true;
            }
        }
        return false;
    }
    suppress(index: number): boolean {
        if(!this.scope(index)) {
            return false;
        }
        const facts = types(this.rules.ask(index, 'call-returns'), 'call-returns');
        const async = this.rules.parser
            .node(index)
            .children.some((child) => this.rules.parser.node(child).kind === 'AsyncKeyword');
        return facts.roots.some((id) =>
            async ? this.promiseVoid(index, facts, facts.type(id)) : facts.has(facts.type(id), 16),
        );
    }
    collect(index: number): void {
        for(const child of this.rules.parser.node(index).children) {
            if(this.scope(child)) {
                continue;
            }
            if(this.rules.parser.node(child).kind === 'ReturnStatement') {
                this.pending.push(child);
            }
            this.collect(child);
        }
    }
    key(index: number): number {
        const node = this.rules.parser.node(index);
        if(node.kind === 'ArrowFunction' || node.kind === 'SourceFile') {
            return -1;
        }
        if(node.kind === 'FunctionExpression') {
            const parentIndex = this.rules.parents[index] ?? -1;
            if(
                parentIndex >= 0 &&
                ['PropertyAssignment', 'PropertyDeclaration'].includes(this.rules.parser.node(parentIndex).kind)
            ) {
                return this.key(parentIndex);
            }
        }
        for(const child of node.children) {
            if(
                ['Identifier', 'PrivateIdentifier', 'StringLiteral', 'NumericLiteral', 'ComputedPropertyName'].includes(
                    this.rules.parser.node(child).kind,
                )
            ) {
                return child;
            }
            if(this.rules.parser.node(child).kind === 'Parameter' || this.rules.parser.node(child).kind === 'Block') {
                break;
            }
        }
        return -1;
    }
    name(index: number, capital: boolean): string {
        const node = this.rules.parser.node(index);
        if(node.kind === 'SourceFile') {
            return capital ? 'Program' : 'program';
        }
        if(node.kind === 'Constructor') {
            return capital ? 'Constructor' : 'constructor';
        }
        const words: string[] = [];
        if(node.children.some((child) => this.rules.parser.node(child).kind === 'StaticKeyword')) {
            words.push('static');
        }
        const key = this.key(index);
        const privateName = key >= 0 && this.rules.parser.node(key).kind === 'PrivateIdentifier';
        if(privateName) {
            words.push('private');
        }
        if(node.children.some((child) => this.rules.parser.node(child).kind === 'AsyncKeyword')) {
            words.push('async');
        }
        if(node.children.some((child) => this.rules.parser.node(child).kind === 'AsteriskToken')) {
            words.push('generator');
        }
        const parent = this.rules.parents[index] ?? -1;
        if(node.kind === 'GetAccessor') {
            words.push('getter');
        }
        else if(node.kind === 'SetAccessor') {
            words.push('setter');
        }
        else if(
            node.kind === 'MethodDeclaration' ||
            (node.kind === 'FunctionExpression' &&
                parent >= 0 &&
                ['PropertyAssignment', 'PropertyDeclaration'].includes(this.rules.parser.node(parent).kind))
        ) {
            words.push('method');
        }
        else {
            if(node.kind === 'ArrowFunction') {
                words.push('arrow');
            }
            words.push('function');
        }
        if(key >= 0) {
            const named = this.rules.parser.node(key);
            if(privateName) {
                words.push(named.text);
            }
            else if(named.kind === 'ComputedPropertyName') {
                const inner = this.rules.parser.node(named.children[0] ?? -1);
                if(['StringLiteral', 'NumericLiteral', 'NoSubstitutionTemplateLiteral'].includes(inner.kind)) {
                    words.push(`'${inner.text}'`);
                }
            }
            else if(named.text !== '') {
                words.push(`'${named.text}'`);
            }
        }
        const name = words.join(' ');
        return capital ? name.slice(0, 1).toUpperCase() + name.slice(1) : name;
    }
    truthy(index: number): boolean {
        if(index < 0) {
            return true;
        }
        const node = this.rules.parser.node(this.rules.skip(index));
        if(['TrueKeyword', 'RegularExpressionLiteral'].includes(node.kind)) {
            return true;
        }
        if(node.kind === 'NumericLiteral') {
            return node.text !== '0';
        }
        if(node.kind === 'StringLiteral') {
            return node.text !== '';
        }
        if(node.kind === 'BigIntLiteral') {
            const first = ['0x', '0X', '0b', '0B', '0o', '0O'].includes(node.text.slice(0, 2)) ? 2 : 0;
            for(let at = first; at < node.text.length - 1; at++) {
                if(!['0', '_'].includes(node.text.slice(at, at + 1))) {
                    return true;
                }
            }
            return false;
        }
        return false;
    }
    throwable(index: number): boolean {
        const node = this.rules.parser.node(index);
        if(this.scope(index)) {
            return false;
        }
        if(
            [
                'CallExpression',
                'NewExpression',
                'PropertyAccessExpression',
                'ElementAccessExpression',
                'TaggedTemplateExpression',
                'YieldExpression',
            ].includes(node.kind)
        ) {
            return true;
        }
        if(node.kind === 'Identifier') {
            const parentIndex = this.rules.parents[index] ?? -1;
            if(parentIndex < 0) {
                return false;
            }
            const parent = this.rules.parser.node(parentIndex);
            if(
                [
                    'LabeledStatement',
                    'BreakStatement',
                    'ContinueStatement',
                    'ArrayBindingPattern',
                    'ImportSpecifier',
                    'ExportSpecifier',
                    'ImportClause',
                    'NamespaceImport',
                    'NamespaceExport',
                    'CatchClause',
                ].includes(parent.kind)
            ) {
                return false;
            }
            if(parent.kind === 'BindingElement') {
                return (
                    !parent.children.some((child) => this.rules.parser.node(child).kind === 'DotDotDotToken') &&
                    this.shadow.name(parentIndex) === index &&
                    this.rules.parser.node(this.rules.parents[parentIndex] ?? -1).kind === 'ObjectBindingPattern'
                );
            }
            if(
                [
                    'FunctionDeclaration',
                    'FunctionExpression',
                    'ArrowFunction',
                    'ClassDeclaration',
                    'ClassExpression',
                    'VariableDeclaration',
                    'PropertyAssignment',
                    'PropertyDeclaration',
                    'MethodDeclaration',
                    'GetAccessor',
                    'SetAccessor',
                    'EnumDeclaration',
                    'ModuleDeclaration',
                    'InterfaceDeclaration',
                    'TypeAliasDeclaration',
                ].includes(parent.kind)
            ) {
                return this.key(parentIndex) !== index;
            }
            return true;
        }
        return node.children.some((child) => this.throwable(child));
    }
    sequence(children: readonly number[]): Completion {
        const result = new Completion(true);
        for(const child of children) {
            if(result.normal) {
                const next = this.flow(child);
                result.normal = false;
                result.merge(next);
            }
        }
        return result;
    }
    jump(index: number, continueJump: boolean): number {
        const node = this.rules.parser.node(index);
        const labelIndex = node.children[0] ?? -1;
        const label = labelIndex >= 0 ? this.rules.parser.node(labelIndex).text : '';
        let parent = this.rules.parents[index] ?? -1;
        while(parent >= 0) {
            const statement = this.rules.parser.node(parent);
            if(
                label !== '' &&
                statement.kind === 'LabeledStatement' &&
                this.rules.parser.node(statement.children[0] ?? -1).text === label
            ) {
                if(!continueJump) {
                    return parent;
                }
                let target = statement.children[1] ?? -1;
                while(this.rules.parser.node(target).kind === 'LabeledStatement') {
                    target = this.rules.parser.node(target).children[1] ?? -1;
                }
                return target;
            }
            if(
                label === '' &&
                (['WhileStatement', 'DoStatement', 'ForStatement', 'ForInStatement', 'ForOfStatement'].includes(
                    statement.kind,
                ) ||
                    (!continueJump && statement.kind === 'SwitchStatement'))
            ) {
                return parent;
            }
            if(this.scope(parent)) {
                break;
            }
            parent = this.rules.parents[parent] ?? -1;
        }
        return -1;
    }
    flow(index: number): Completion {
        const node = this.rules.parser.node(index);
        if(node.kind === 'ReturnStatement') {
            const result = new Completion(false);
            result.thrown = node.children.some((child) => this.throwable(child));
            return result;
        }
        if(node.kind === 'ThrowStatement') {
            const result = new Completion(false);
            result.thrown = true;
            return result;
        }
        if(node.kind === 'BreakStatement' || node.kind === 'ContinueStatement') {
            const result = new Completion(false);
            if(node.kind === 'BreakStatement') {
                result.breaks.push(this.jump(index, false));
            }
            else {
                result.continues.push(this.jump(index, true));
            }
            return result;
        }
        if(node.kind === 'Block' || node.kind === 'SourceFile') {
            return this.sequence(node.children);
        }
        if(node.kind === 'IfStatement') {
            const result = this.flow(node.children[1] ?? -1);
            if(this.throwable(node.children[0] ?? -1)) {
                result.thrown = true;
            }
            result.merge(node.children.length > 2 ? this.flow(node.children[2] ?? -1) : new Completion(true));
            return result;
        }
        if(['WhileStatement', 'DoStatement', 'ForStatement', 'ForOfStatement', 'ForInStatement'].includes(node.kind)) {
            const bodyIndex =
                node.kind === 'DoStatement'
                    ? (node.children[0] ?? -1)
                    : (node.children[node.children.length - 1] ?? -1);
            const body = this.flow(bodyIndex);
            let condition =
                node.kind === 'WhileStatement'
                    ? (node.children[0] ?? -1)
                    : node.kind === 'DoStatement'
                      ? (node.children[1] ?? -1)
                      : -1;
            if(node.kind === 'ForStatement') {
                this.rules.scanner.pos = this.rules.start(node);
                this.rules.scanner.scan();
                this.rules.scanner.scan();
                this.rules.scanner.scan();
                let nesting = 0;
                let semi = 0;
                let conditionStart = -1;
                while(
                    this.rules.scanner.start < this.rules.parser.node(bodyIndex).pos &&
                    this.rules.scanner.kind !== 'EndOfFileToken'
                ) {
                    if(semi === 1 && conditionStart < 0 && this.rules.scanner.kind !== 'SemicolonToken') {
                        conditionStart = this.rules.scanner.start;
                    }
                    if(['OpenParenToken', 'OpenBracketToken', 'OpenBraceToken'].includes(this.rules.scanner.kind)) {
                        nesting++;
                    }
                    if(['CloseParenToken', 'CloseBracketToken', 'CloseBraceToken'].includes(this.rules.scanner.kind)) {
                        nesting--;
                    }
                    if(nesting === 0 && this.rules.scanner.kind === 'SemicolonToken') {
                        semi++;
                    }

                    this.rules.scanner.scan();
                }
                if(conditionStart >= 0) {
                    condition =
                        node.children.find(
                            (child) =>
                                child !== bodyIndex &&
                                this.rules.start(this.rules.parser.node(child)) === conditionStart,
                        ) ?? -1;
                }
            }
            const canSkip = ['ForOfStatement', 'ForInStatement'].includes(node.kind) || !this.truthy(condition);
            const result = new Completion(
                body.breaks.includes(index) ||
                    (node.kind === 'DoStatement'
                        ? (body.normal || body.continues.includes(index)) && canSkip
                        : canSkip),
            );
            result.outward(body, index);
            return result;
        }
        if(node.kind === 'SwitchStatement') {
            const clauses = this.rules.parser.node(node.children[1] ?? -1).children;
            const result = new Completion(
                !clauses.some((clause) => this.rules.parser.node(clause).kind === 'DefaultClause'),
            );
            let following = new Completion(true);
            for(let at = clauses.length - 1; at >= 0; at--) {
                const clause = this.rules.parser.node(clauses[at] ?? -1);
                const flow = this.sequence(clause.children.slice(clause.kind === 'CaseClause' ? 1 : 0));
                if(flow.normal) {
                    flow.normal = false;
                    flow.merge(following);
                }
                following = flow;
                if(flow.normal || flow.breaks.includes(index)) {
                    result.normal = true;
                }
                result.outward(flow, index);
            }
            return result;
        }
        if(node.kind === 'TryStatement') {
            let result = this.flow(node.children[0] ?? -1);
            const catchIndex = node.children.find((child) => this.rules.parser.node(child).kind === 'CatchClause');
            if(catchIndex !== undefined && (result.normal || result.thrown)) {
                const caught = this.rules.parser.node(catchIndex);
                result.thrown = false;
                result.merge(this.flow(caught.children[caught.children.length - 1] ?? -1));
            }
            const finalIndex =
                node.children.length > 1 &&
                this.rules.parser.node(node.children[node.children.length - 1] ?? -1).kind === 'Block'
                    ? (node.children[node.children.length - 1] ?? -1)
                    : -1;
            if(finalIndex >= 0) {
                const final = this.flow(finalIndex);
                if(final.normal) {
                    final.normal = false;
                    result.merge(final);
                }
                else {
                    result = final;
                }
            }
            return result;
        }
        if(node.kind === 'LabeledStatement') {
            const flow = this.flow(node.children[1] ?? -1);
            const result = new Completion(flow.normal || flow.breaks.includes(index));
            result.outward(flow, index);
            return result;
        }
        if(node.kind === 'WithStatement') {
            return this.flow(node.children[1] ?? -1);
        }
        const result = new Completion(true);
        result.thrown = !this.scope(index) && this.throwable(index);
        return result;
    }
    judge(index: number): void {
        const node = this.rules.parser.node(index);
        while(this.pending.length > 0) {
            this.pending.pop();
        }
        this.collect(index);
        let returns = this.pending.slice();
        if(returns.length === 0) {
            return;
        }
        if(returns.some((child) => this.rules.parser.node(child).children.length === 0) && this.suppress(index)) {
            returns = returns.filter((child) => this.rules.parser.node(child).children.length > 0);
        }
        if(returns.length === 0) {
            return;
        }
        const first = this.rules.parser.node(returns[0] ?? -1).children.length > 0;
        for(const child of returns.slice(1)) {
            if(this.rules.parser.node(child).children.length > 0 !== first) {
                const id = first ? 'missingReturnValue' : 'unexpectedReturnValue';
                const tail = first
                    ? 'expected a return value. This returns nothing while another return in the same function returns a value. The caller receives undefined from one path and a value from another, and nothing in the signature says which.'
                    : 'expected no return value. This returns a value while another return in the same function returns nothing. A reader who saw the bare return will not expect a value to come back from here.';
                this.rules.add('consistent-return', id, `${this.name(index, true)} ${tail}`, child);
            }
        }
        const key = this.key(index);
        if(
            !first ||
            node.kind === 'Constructor' ||
            (['FunctionDeclaration', 'FunctionExpression'].includes(node.kind) &&
                key >= 0 &&
                (this.rules.parser.node(key).text.charCodeAt(0) > 127 ||
                    this.rules.parser.node(key).text.slice(0, 1) !==
                        this.rules.parser.node(key).text.slice(0, 1).toLowerCase()))
        ) {
            return;
        }
        const body = node.kind === 'SourceFile' ? index : (node.children[node.children.length - 1] ?? -1);
        if(!this.flow(body).normal) {
            return;
        }
        const message = `Expected to return a value at the end of ${this.name(index, false)}. Some paths through this function return a value and some run off the end, which returns undefined. A caller reading one branch cannot tell which kind of function this is, so the undefined arrives somewhere far from here.`;
        if(node.kind === 'SourceFile') {
            this.rules.findings.push(new Diagnostic('consistent-return', 'missingReturn', message, 0, 0));
        }
        else {
            const span =
                node.kind === 'ArrowFunction'
                    ? (node.children.find((child) => this.rules.parser.node(child).kind === 'EqualsGreaterThanToken') ??
                      index)
                    : key >= 0
                      ? key
                      : index;
            this.rules.add('consistent-return', 'missingReturn', message, span);
        }
    }
    run(): void {
        for(let index = 0; index < this.rules.parser.nodes.length; index++) {
            if(this.scope(index) || this.rules.parser.node(index).kind === 'SourceFile') {
                this.judge(index);
            }
        }
    }
}
