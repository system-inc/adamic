// The next production cohere rules, selected by measured finding volume.
import { panic } from 'adamic';
import type { Rules } from './rules.ts';
import { unsafeAssignment } from './rules.ts';
import { types } from './facts.ts';
import type { Types } from './types.ts';
import { Frames, header } from './frames.ts';
import { Repair } from './repair.ts';
import { Suggestion } from './suggestion.ts';
import { unionFlag, intersectionFlag, numberLike, stringLike } from './flags.ts';

export class Volume {
    readonly rules: Rules;
    readonly memberStates: number[];
    constructor(rules: Rules) {
        this.rules = rules;
        this.memberStates = rules.parser.nodes.map(() => 0);
    }
    facts(index: number, question = 'raw-shape'): Types {
        return types(this.rules.ask(index, question), question);
    }
    assignable(index: number, source: number, target: number): boolean {
        const frames = new Frames(this.rules.ask(index, `assignable-types\n${source}\n${target}`));
        header(frames, 'assignable-types');
        const answer = frames.yes();
        frames.end();
        return answer;
    }
    assertion(index: number): void {
        const node = this.rules.parser.node(index);
        const expression = node.children[node.kind === 'AsExpression' ? 0 : 1] ?? panic('missing asserted expression');
        const annotation = node.children[node.kind === 'AsExpression' ? 1 : 0] ?? panic('missing asserted type');
        const senderFacts = this.facts(expression);
        const targetFacts = this.facts(annotation);
        const sender = senderFacts.root();
        const target = targetFacts.root();
        if(sender.id === target.id) {
            return;
        }
        if((target.flags & 1) !== 0 && (sender.flags & 2) !== 0) {
            this.rules.add(
                'no-unsafe-type-assertion',
                'unsafeToAnyTypeAssertion',
                'Unsafe assertion to `any` detected: consider using a more specific type to ensure safety.',
                index,
            );
            return;
        }
        const sourceNode = this.rules.parser.node(expression);
        const emptyMap =
            sourceNode.kind === 'NewExpression' &&
            sourceNode.list === 0 &&
            sourceNode.children.length === 1 &&
            this.rules.parser.node(sourceNode.children[0] ?? -1).kind === 'Identifier' &&
            this.rules.parser.node(sourceNode.children[0] ?? -1).text === 'Map';
        if(unsafeAssignment(senderFacts, sender, targetFacts, target, emptyMap, [])) {
            this.rules.add(
                'no-unsafe-type-assertion',
                'unsafeOfAnyTypeAssertion',
                `Unsafe assertion from ${sender.error ? 'error typed' : '`any`'} detected: consider using type guards or a safer assertion.`,
                index,
            );
            return;
        }
        if(unsafeAssignment(targetFacts, target, senderFacts, sender, false, [])) {
            this.rules.add(
                'no-unsafe-type-assertion',
                'unsafeToAnyTypeAssertion',
                `Unsafe assertion to ${target.error ? 'error typed' : '`any`'} detected: consider using a more specific type to ensure safety.`,
                index,
            );
            return;
        }
        const widened = this.facts(expression, 'widened-shape').root();
        if(this.assignable(index, widened.id, target.id)) {
            return;
        }
        const name = this.rules.name(target, index);
        if((target.flags & 524288) !== 0) {
            if(target.constraint === 0) {
                this.rules.add(
                    'no-unsafe-type-assertion',
                    'unsafeToUnconstrainedTypeAssertion',
                    `Unsafe type assertion: '${name}' could be instantiated with an arbitrary type which could be unrelated to the original type.`,
                    index,
                );
                return;
            }
            if(this.assignable(index, widened.id, target.constraint)) {
                this.rules.add(
                    'no-unsafe-type-assertion',
                    'unsafeTypeAssertionAssignableToConstraint',
                    `Unsafe type assertion: the original type is assignable to the constraint of type '${name}', but '${name}' could be instantiated with a different subtype of its constraint.`,
                    index,
                );
                return;
            }
        }
        this.rules.add(
            'no-unsafe-type-assertion',
            'unsafeTypeAssertion',
            `Unsafe type assertion: type '${name}' is more narrow than the original type.`,
            index,
        );
    }
    member(index: number): number {
        const cached = this.memberStates[index] ?? 0;
        if(cached !== 0) {
            return cached;
        }
        const node = this.rules.parser.node(index);
        const object = node.children[0] ?? panic('missing member object');
        const property = node.children[node.children.length - 1] ?? panic('missing member property');
        const objectNode = this.rules.parser.node(object);
        if(
            ['PropertyAccessExpression', 'ElementAccessExpression'].includes(objectNode.kind) &&
            this.member(object) === 1
        ) {
            this.memberStates[index] = 1;
            return 1;
        }
        const objectType = this.facts(object).root();
        const state = (objectType.flags & 1) !== 0 ? 1 : 2;
        this.memberStates[index] = state;
        if(state === 1) {
            const propertyText = this.rules.text(this.rules.parser.node(property));
            const rendered = node.kind === 'ElementAccessExpression' ? `[${propertyText}]` : `.${propertyText}`;
            this.rules.add(
                'no-unsafe-member-access',
                objectType.error ? 'errorMemberExpression' : 'unsafeMemberExpression',
                objectType.error
                    ? `Unsafe member access ${rendered} on a type that cannot be resolved.`
                    : `Unsafe member access ${rendered} on an \`any\` value.`,
                property,
            );
        }
        return state;
    }
    computed(index: number): void {
        const node = this.rules.parser.node(index);
        const property = this.rules.skip(node.children[node.children.length - 1] ?? panic('missing computed key'));
        const key = this.rules.parser.node(property);
        if(
            [
                'NumericLiteral',
                'StringLiteral',
                'BigIntLiteral',
                'NoSubstitutionTemplateLiteral',
                'RegularExpressionLiteral',
                'TrueKeyword',
                'FalseKeyword',
                'NullKeyword',
                'PrefixUnaryExpression',
                'PostfixUnaryExpression',
            ].includes(key.kind)
        ) {
            return;
        }
        const type = this.facts(property).root();
        if((type.flags & 1) !== 0) {
            const text = `[${this.rules.text(key)}]`;
            this.rules.add(
                'no-unsafe-member-access',
                type.error ? 'errorComputedMemberAccess' : 'unsafeComputedMemberAccess',
                type.error
                    ? `The type of computed name ${text} cannot be resolved.`
                    : `Computed name ${text} resolves to an \`any\` value.`,
                property,
            );
        }
    }
    enumTypes(index: number): number[] {
        const frames = new Frames(this.rules.ask(index, 'enum-types'));
        header(frames, 'enum-types');
        const ids = frames.ids();
        frames.end();
        return ids;
    }
    primitive(facts: Types, flag: number): boolean {
        return facts
            .parts(facts.root())
            .every((id) =>
                facts.parts(facts.type(id), intersectionFlag).some((part) => (facts.type(part).flags & flag) !== 0),
            );
    }
    enumPair(index: number, left: number, right: number, isCase: boolean): void {
        const leftEnums = this.enumTypes(left);
        const rightEnums = this.enumTypes(right);
        if((leftEnums.length === 0 && rightEnums.length === 0) || leftEnums.some((id) => rightEnums.includes(id))) {
            return;
        }
        const first = this.facts(left);
        const second = this.facts(right);
        const ap = first.parts(first.root());
        const bp = second.parts(second.root());
        if(ap.some((id) => bp.includes(id))) {
            return;
        }
        const an = ap.some((id) => (first.type(id).flags & 98304) !== 0 && (first.type(id).flags & 2048) !== 0);
        const as = ap.some((id) => (first.type(id).flags & 98304) !== 0 && (first.type(id).flags & 2048) === 0);
        const bn = bp.some((id) => (second.type(id).flags & 98304) !== 0 && (second.type(id).flags & 2048) !== 0);
        const bs = bp.some((id) => (second.type(id).flags & 98304) !== 0 && (second.type(id).flags & 2048) === 0);
        if(!(
            (an && this.primitive(second, numberLike)) ||
            (as && this.primitive(second, stringLike)) ||
            (bn && this.primitive(first, numberLike)) ||
            (bs && this.primitive(first, stringLike))
        )) {
            return;
        }
        this.rules.add(
            'no-unsafe-enum-comparison',
            isCase ? 'mismatchedCase' : 'mismatchedCondition',
            isCase
                ? "The case statement does not have a shared enum type with the switch predicate. A `case` holding a bare literal where the discriminant is an enum compiles and stops matching the moment the enum's value changes, and nothing reports it: the switch simply falls through. Compare against the enum member instead, which also lets exhaustiveness checking see the case."
                : 'The two values in this comparison do not have a shared enum type. Comparing an enum member against a bare literal type-checks and then stops protecting you: the literal is not tied to the enum, so renaming a member, changing its value, or removing it leaves this comparison compiling and silently false. That is the failure enums exist to prevent. Compare against the enum member itself.',
            index,
        );
    }
    literalParts(facts: Types, id: number): number[] {
        const type = facts.type(id);
        if((type.flags & (unionFlag | intersectionFlag)) === 0) {
            return [id];
        }
        const parts: number[] = [];
        for(const child of type.parts) {
            for(const leaf of this.literalParts(facts, child)) {
                parts.push(leaf);
            }
        }
        return parts;
    }
    exhaustive(index: number): void {
        const node = this.rules.parser.node(index);
        const expression = node.children[0] ?? panic('missing switch expression');
        const cases = this.rules.parser.node(node.children[1] ?? panic('missing case block'));
        const covered: number[] = [];
        let undefinedCovered = false;
        for(const clauseIndex of cases.children) {
            const clause = this.rules.parser.node(clauseIndex);
            if(clause.kind !== 'CaseClause') {
                continue;
            }
            const test = clause.children[0] ?? panic('missing case expression');
            const type = this.facts(test, 'type-shape').root();
            covered.push(type.id);
            if((type.flags & 4) !== 0) {
                undefinedCovered = true;
            }
            this.enumPair(clauseIndex, expression, test, true);
        }
        const facts = this.facts(expression, 'type-shape');
        const names: string[] = [];
        for(const id of this.literalParts(facts, facts.root().id)) {
            const type = facts.type(id);
            if(
                (type.flags & (1024 | 2048 | 4096 | 8192 | 4 | 8 | 16384)) === 0 ||
                covered.includes(id) ||
                (undefinedCovered && (type.flags & 4) !== 0)
            ) {
                continue;
            }
            if((type.flags & (512 | 16384)) !== 0) {
                const frames = new Frames(this.rules.ask(expression, `type-symbol\n${id}`));
                header(frames, 'type-symbol');
                const symbol = frames.field();
                frames.end();
                names.push(symbol === '' ? this.rules.name(type, expression) : `typeof ${symbol}`);
            }
            else {
                names.push(this.rules.name(type, expression));
            }
        }
        if(names.length > 0) {
            this.rules.add(
                'switch-exhaustiveness-check',
                'switchIsNotExhaustive',
                `Switch is not exhaustive. Cases not matched: ${names.join(' | ')}`,
                expression,
            );
        }
    }
    unionFlags(index: number): number {
        const facts = this.facts(index);
        let flags = 0;
        for(const id of facts.parts(facts.root())) {
            flags |= facts.type(id).flags;
        }
        return flags;
    }
    conditional(index: number): boolean {
        let current = index;
        while((this.rules.parents[current] ?? -1) >= 0) {
            const parentIndex = this.rules.parents[current] ?? -1;
            const parent = this.rules.parser.node(parentIndex);
            if(['IfStatement', 'WhileStatement', 'ConditionalExpression'].includes(parent.kind)) {
                if(parent.children[0] === current) {
                    return true;
                }
                if(parent.kind !== 'ConditionalExpression') {
                    return false;
                }
            }
            else if(parent.kind === 'DoStatement') {
                return parent.children[1] === current;
            }
            else if(parent.kind === 'ForStatement') {
                return parent.children[1] === current;
            }
            else if(parent.kind === 'BinaryExpression') {
                const op = this.rules.parser.node(parent.children[1] ?? -1).kind;
                if(
                    !['BarBarToken', 'AmpersandAmpersandToken', 'QuestionQuestionToken'].includes(op) &&
                    !(op === 'CommaToken' && parent.children[2] === current)
                ) {
                    return false;
                }
            }
            else if(parent.kind === 'PrefixUnaryExpression') {
                if(parent.operator !== 'ExclamationToken') {
                    return false;
                }
            }
            else if(parent.kind !== 'ParenthesizedExpression') {
                return false;
            }
            current = parentIndex;
        }
        return false;
    }
    similar(left: number, right: number): boolean {
        const first = this.rules.parser.node(this.rules.skip(left));
        const second = this.rules.parser.node(this.rules.skip(right));
        const am = ['PropertyAccessExpression', 'ElementAccessExpression'].includes(first.kind);
        const bm = ['PropertyAccessExpression', 'ElementAccessExpression'].includes(second.kind);
        if(am && bm) {
            const ap = this.rules.skip(first.children[first.children.length - 1] ?? -1);
            const bp = this.rules.skip(second.children[second.children.length - 1] ?? -1);
            const x = this.rules.parser.node(ap);
            const y = this.rules.parser.node(bp);
            const propertyEqual =
                first.kind === second.kind
                    ? this.similar(ap, bp)
                    : (x.kind === 'StringLiteral' && y.kind === 'Identifier' && x.text === y.text) ||
                      (y.kind === 'StringLiteral' && x.kind === 'Identifier' && x.text === y.text);
            return propertyEqual && this.similar(first.children[0] ?? -1, second.children[0] ?? -1);
        }
        if(first.kind !== second.kind) {
            return false;
        }
        if(['ThisKeyword', 'NullKeyword', 'TrueKeyword', 'FalseKeyword'].includes(first.kind)) {
            return true;
        }
        if(['Identifier', 'StringLiteral', 'BigIntLiteral'].includes(first.kind)) {
            return first.text === second.text;
        }
        return first.kind === 'NumericLiteral' && first.text === second.text;
    }
    memberLike(index: number): boolean {
        return ['Identifier', 'PropertyAccessExpression', 'ElementAccessExpression'].includes(
            this.rules.parser.node(this.rules.skip(index)).kind,
        );
    }
    wrappedText(index: number): string {
        const inner = this.rules.skip(index);
        const parent = this.rules.parents[inner] ?? -1;
        return this.rules.text(
            this.rules.parser.node(
                parent >= 0 && this.rules.parser.node(parent).kind === 'ParenthesizedExpression' ? parent : inner,
            ),
        );
    }
    suggest(
        index: number,
        id: string,
        message: string,
        start: number,
        end: number,
        text: string,
        operator: string,
    ): void {
        const finding = this.rules.add('prefer-nullish-coalescing', id, message, index);
        const repair = new Repair(this.rules.byte(start), this.rules.byte(end), text);
        const suggestion = new Suggestion(
            'suggestNullish',
            `Fix to nullish coalescing operator (\`${operator}\`). Only null and undefined will take the right-hand branch.`,
            [repair],
        );
        finding.suggestions.push(suggestion);
    }
    comments(start: number, end: number, separator: string): string {
        const source = this.rules.scanner.text;
        let text = '';
        for(let at = start; at < end; at++) {
            if(source.slice(at, at + 2) === '//') {
                let last = at + 2;
                while(last < end && source.charCodeAt(last) !== 10 && source.charCodeAt(last) !== 13) {
                    last++;
                }
                text += source.slice(at, last) + separator;
                at = last - 1;
            }
            else if(source.slice(at, at + 2) === '/*') {
                const close = source.indexOf('*/', at + 2);
                if(close >= at && close + 2 <= end) {
                    text += source.slice(at, close + 2) + separator;
                    at = close + 1;
                }
            }
        }
        return text;
    }
    nullish(index: number): void {
        const node = this.rules.parser.node(index);
        if(node.kind === 'BinaryExpression') {
            const operatorIndex = node.children[1] ?? -1;
            const operator = this.rules.parser.node(operatorIndex);
            if(!['BarBarToken', 'BarBarEqualsToken'].includes(operator.kind) || this.conditional(index)) {
                return;
            }
            const left = node.children[0] ?? -1;
            if((this.unionFlags(left) & 31) === 0) {
                return;
            }
            const replacement = operator.kind === 'BarBarToken' ? '??' : '??=';
            const original = operator.kind === 'BarBarToken' ? '||' : '||=';
            const kind = operator.kind === 'BarBarToken' ? 'or' : 'assignment';
            let parentIndex = this.rules.parents[index] ?? -1;
            while(parentIndex >= 0 && this.rules.parser.node(parentIndex).kind === 'ParenthesizedExpression') {
                parentIndex = this.rules.parents[parentIndex] ?? -1;
            }
            let start = this.rules.start(operator);
            let end = operator.end;
            let text = replacement;
            if(
                parentIndex >= 0 &&
                this.rules.parser.node(parentIndex).kind === 'BinaryExpression' &&
                ['BarBarToken', 'BarBarEqualsToken'].includes(
                    this.rules.parser.node(this.rules.parser.node(parentIndex).children[1] ?? -1).kind,
                )
            ) {
                let operand = left;
                while(
                    this.rules.parser.node(this.rules.skip(operand)).kind === 'BinaryExpression' &&
                    ['BarBarToken', 'BarBarEqualsToken'].includes(
                        this.rules.parser.node(this.rules.parser.node(this.rules.skip(operand)).children[1] ?? -1).kind,
                    )
                ) {
                    operand = this.rules.parser.node(this.rules.skip(operand)).children[2] ?? -1;
                }
                const right = node.children[2] ?? -1;
                start = this.rules.start(this.rules.parser.node(operand));
                end = this.rules.parser.node(right).end;
                text = `(${this.rules.text(this.rules.parser.node(operand))} ${replacement} ${this.rules.text(this.rules.parser.node(right))})`;
            }
            this.suggest(
                operatorIndex,
                'preferNullishOverOr',
                `Prefer using nullish coalescing operator (\`${replacement}\`) instead of a logical ${kind} (\`${original}\`), as it is a safer operator. \`${original}\` falls through on every falsy value, so an empty string, a zero and a false all take the right-hand branch alongside null and undefined. Where the left side is nullable, \`${replacement}\` says what was meant and leaves the valid falsy values alone.`,
                start,
                end,
                text,
                replacement,
            );
            return;
        }
        if(node.kind !== 'ConditionalExpression' && node.kind !== 'IfStatement') {
            return;
        }
        let present = node.children[node.kind === 'ConditionalExpression' ? 2 : 0] ?? -1;
        let fallback = -1;
        let before = '';
        let after = '';
        if(node.kind === 'IfStatement') {
            if(node.children.length !== 2) {
                return;
            }
            let statement = node.children[1] ?? -1;
            if(this.rules.parser.node(statement).kind === 'Block') {
                const block = this.rules.parser.node(statement);
                if(block.children.length !== 1) {
                    return;
                }
                statement = block.children[0] ?? -1;
            }
            if(this.rules.parser.node(statement).kind !== 'ExpressionStatement') {
                return;
            }
            const assignment = this.rules.skip(this.rules.parser.node(statement).children[0] ?? -1);
            const assigned = this.rules.parser.node(assignment);
            if(
                assigned.kind !== 'BinaryExpression' ||
                !this.rules.parser.node(assigned.children[1] ?? -1).kind.endsWith('EqualsToken')
            ) {
                return;
            }
            const consequent = this.rules.parser.node(node.children[1] ?? -1);
            before = this.comments(assigned.pos, this.rules.start(assigned), consequent.kind === 'Block' ? '\n' : ' ');
            if(consequent.kind === 'Block') {
                after = this.comments(this.rules.parser.node(statement).end, consequent.end - 1, '\n');
            }
            present = this.rules.skip(assigned.children[0] ?? -1);
            fallback = assigned.children[2] ?? -1;
            if(!this.memberLike(present)) {
                return;
            }
        }
        if(node.kind === 'ConditionalExpression') {
            fallback = node.children[4] ?? -1;
        }
        const test = this.rules.skip(node.children[0] ?? -1);
        const testNode = this.rules.parser.node(test);
        let operator = '';
        const compared: number[] = [];
        if(this.memberLike(test)) {
            operator = '';
        }
        else if(
            testNode.kind === 'PrefixUnaryExpression' &&
            testNode.operator === 'ExclamationToken' &&
            this.memberLike(testNode.children[0] ?? -1)
        ) {
            operator = '!';
        }
        else if(testNode.kind === 'BinaryExpression') {
            const op = this.rules.parser.node(testNode.children[1] ?? -1).kind;
            if(['BarBarToken', 'AmpersandAmpersandToken'].includes(op)) {
                for(const side of [testNode.children[0] ?? -1, testNode.children[2] ?? -1]) {
                    const comparison = this.rules.parser.node(this.rules.skip(side));
                    if(comparison.kind !== 'BinaryExpression') {
                        return;
                    }
                    const half = this.rules.parser.node(comparison.children[1] ?? -1).kind;
                    const strict = op === 'BarBarToken' ? 'EqualsEqualsEqualsToken' : 'ExclamationEqualsEqualsToken';
                    const loose = op === 'BarBarToken' ? 'EqualsEqualsToken' : 'ExclamationEqualsToken';
                    if(half !== strict && half !== loose) {
                        return;
                    }
                    if(operator === '' || half === loose) {
                        operator = half;
                    }
                    compared.push(comparison.children[0] ?? -1);
                    compared.push(comparison.children[2] ?? -1);
                }
            }
            else if(
                [
                    'EqualsEqualsToken',
                    'EqualsEqualsEqualsToken',
                    'ExclamationEqualsToken',
                    'ExclamationEqualsEqualsToken',
                ].includes(op)
            ) {
                operator = op;
                compared.push(testNode.children[0] ?? -1);
                compared.push(testNode.children[2] ?? -1);
            }
            else {
                return;
            }
        }
        else {
            return;
        }
        const positive =
            operator === '' || operator === 'ExclamationEqualsToken' || operator === 'ExclamationEqualsEqualsToken';
        if(node.kind === 'ConditionalExpression' && !positive) {
            present = node.children[4] ?? -1;
            fallback = node.children[2] ?? -1;
        }
        if(node.kind === 'IfStatement' && positive) {
            return;
        }
        let subject = -1;
        let checksNull = false;
        let checksUndefined = false;
        if(compared.length === 0) {
            subject = operator === '!' ? (testNode.children[0] ?? -1) : test;
            if(!this.similar(subject, present) || this.conditional(index) || (this.unionFlags(subject) & 31) === 0) {
                return;
            }
        }
        else {
            for(const operand of compared) {
                const value = this.rules.parser.node(this.rules.skip(operand));
                if(value.kind === 'NullKeyword') {
                    checksNull = true;
                }
                else if(value.kind === 'Identifier' && value.text === 'undefined') {
                    checksUndefined = true;
                }
                else if(this.similar(operand, present)) {
                    if(subject < 0) {
                        subject = this.rules.skip(operand);
                    }
                }
                else {
                    return;
                }
            }
            if(subject < 0 || (!checksNull && !checksUndefined)) {
                return;
            }
            if(checksNull !== checksUndefined && !['EqualsEqualsToken', 'ExclamationEqualsToken'].includes(operator)) {
                const flags = this.unionFlags(subject);
                if((flags & 3) !== 0 || (checksNull && (flags & 4) !== 0) || (checksUndefined && (flags & 8) !== 0)) {
                    return;
                }
            }
        }
        const nullishBranch = this.rules.skip(fallback);
        let right = this.wrappedText(fallback);
        if(
            node.kind === 'ConditionalExpression' &&
            fallback === nullishBranch &&
            (['ConditionalExpression', 'ArrowFunction', 'YieldExpression'].includes(
                this.rules.parser.node(fallback).kind,
            ) ||
                (this.rules.parser.node(fallback).kind === 'BinaryExpression' &&
                    [
                        'EqualsToken',
                        'CommaToken',
                        'QuestionQuestionToken',
                        'PlusEqualsToken',
                        'MinusEqualsToken',
                        'AsteriskEqualsToken',
                        'SlashEqualsToken',
                        'PercentEqualsToken',
                        'BarBarEqualsToken',
                        'AmpersandAmpersandEqualsToken',
                        'QuestionQuestionEqualsToken',
                    ].includes(this.rules.parser.node(this.rules.parser.node(fallback).children[1] ?? -1).kind)))
        ) {
            right = `(${right})`;
        }
        const replacement = node.kind === 'IfStatement' ? '??=' : '??';
        const text = `${before}${this.wrappedText(node.kind === 'IfStatement' ? present : subject)} ${replacement} ${right}${node.kind === 'IfStatement' ? ';' : ''}${after === '' ? '' : ` ${after.slice(0, -1)}`}`;
        this.suggest(
            index,
            node.kind === 'IfStatement' ? 'preferNullishOverAssignment' : 'preferNullishOverTernary',
            node.kind === 'IfStatement'
                ? 'Prefer using nullish coalescing operator (`??=`) instead of an assignment expression, as it is simpler to read. The `if` assigns only when the target is null or undefined, which is exactly what `??=` does, in one statement that names the target once.'
                : 'Prefer using nullish coalescing operator (`??`) instead of a ternary expression, as it is simpler to read. The test checks the value for null and undefined and then names it again in a branch; `??` says the same thing once, so the tested expression and the returned one cannot drift apart.',
            this.rules.start(node),
            node.end,
            text,
            replacement,
        );
    }
    walk(index: number): void {
        const node = this.rules.parser.node(index);
        this.nullish(index);
        if(node.kind === 'SwitchStatement') {
            this.exhaustive(index);
        }
        if(node.kind === 'BinaryExpression') {
            const op = this.rules.parser.node(node.children[1] ?? -1).kind;
            if(
                [
                    'EqualsEqualsToken',
                    'EqualsEqualsEqualsToken',
                    'ExclamationEqualsToken',
                    'ExclamationEqualsEqualsToken',
                    'LessThanToken',
                    'GreaterThanToken',
                    'LessThanEqualsToken',
                    'GreaterThanEqualsToken',
                ].includes(op)
            ) {
                this.enumPair(index, node.children[0] ?? -1, node.children[2] ?? -1, false);
            }
        }
        if(node.kind === 'AsExpression' || node.kind === 'TypeAssertionExpression') {
            this.assertion(index);
        }
        if(node.kind === 'PropertyAccessExpression' || node.kind === 'ElementAccessExpression') {
            this.member(index);
        }
        if(node.kind === 'ElementAccessExpression') {
            this.computed(index);
        }
        for(const child of node.children) {
            this.walk(child);
        }
    }
    run(): void {
        this.walk(this.rules.parser.nodes.length - 1);
        for(const finding of this.rules.findings) {
            finding.sortKey = finding.written();
        }
        this.rules.findings.sort((left, right) =>
            left.sortKey < right.sortKey ? -1 : left.sortKey > right.sortKey ? 1 : 0,
        );
    }
}
