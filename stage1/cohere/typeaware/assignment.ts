import { Diagnostic } from './diagnostic.ts';
import type { Rules } from './rules.ts';
import { unsafeAssignment } from './rules.ts';
import type { Volume } from './volume.ts';
import type { Shadow } from './shadow.ts';
import type { Types } from './types.ts';
import { types } from './facts.ts';

export class Assignment {
    readonly rules: Rules;
    readonly volume: Volume;
    readonly shadow: Shadow;
    constructor(rules: Rules, volume: Volume, shadow: Shadow) {
        this.rules = rules;
        this.volume = volume;
        this.shadow = shadow;
    }
    property(index: number, id: number, name: string): Types {
        return types(this.rules.ask(index, `property-shape\n${id}\n${name}`), 'property-shape');
    }
    anyArray(facts: Types, id: number): boolean {
        const root = facts.type(id);
        return root.array && root.arguments.length > 0 && (facts.type(root.arguments[0] ?? 0).flags & 1) !== 0;
    }
    destructure(target: number, facts: Types, id: number, sender: number): boolean {
        const node = this.rules.parser.node(target);
        const source = facts.type(id);
        const array = ['ArrayBindingPattern', 'ArrayLiteralExpression'].includes(node.kind);
        const object = ['ObjectBindingPattern', 'ObjectLiteralExpression'].includes(node.kind);
        if(!array && !object) {
            return false;
        }
        if(array && this.anyArray(facts, id)) {
            this.rules.add(
                'no-unsafe-assignment',
                'unsafeArrayPattern',
                'Unsafe array destructuring of an `any` array value.',
                target,
            );
            return false;
        }
        if(array && source.tuple < 0) {
            return true;
        }
        let reported = false;
        for(let at = 0; at < node.children.length; at++) {
            const element = node.children[at] ?? -1;
            const entry = this.rules.parser.node(element);
            if(
                ['SpreadElement', 'SpreadAssignment', 'OmittedExpression'].includes(entry.kind) ||
                entry.children.some((child) => this.rules.parser.node(child).kind === 'DotDotDotToken')
            ) {
                continue;
            }
            let name = element;
            if(entry.kind === 'BindingElement') {
                const initializer = this.shadow.initializer(element);
                const children = entry.children.filter(
                    (child) => child !== initializer && this.rules.parser.node(child).kind !== 'DotDotDotToken',
                );
                name = children[children.length - 1] ?? -1;
            }
            else if(entry.kind === 'PropertyAssignment') {
                name = entry.children[1] ?? -1;
            }
            if(name < 0) {
                continue;
            }
            if(array) {
                const member = source.arguments[at] ?? 0;
                if(member === 0) {
                    continue;
                }
                const type = facts.type(member);
                if((type.flags & 1) !== 0) {
                    this.rules.add(
                        'no-unsafe-assignment',
                        'unsafeArrayPatternFromTuple',
                        `Unsafe array destructuring of a tuple element with an ${type.error ? 'error typed' : '`any`'} value.`,
                        name,
                    );
                    reported = true;
                }
                else if(
                    [
                        'ArrayBindingPattern',
                        'ArrayLiteralExpression',
                        'ObjectBindingPattern',
                        'ObjectLiteralExpression',
                    ].includes(this.rules.parser.node(name).kind)
                ) {
                    reported = this.destructure(name, facts, member, sender);
                }
            }
            else {
                let key = entry.children[0] ?? element;
                if(entry.kind === 'ShorthandPropertyAssignment') {
                    key = entry.children[0] ?? -1;
                }
                if(this.rules.parser.node(key).kind === 'ComputedPropertyName') {
                    key = this.rules.parser.node(key).children[0] ?? -1;
                    if(
                        !['StringLiteral', 'NumericLiteral', 'NoSubstitutionTemplateLiteral'].includes(
                            this.rules.parser.node(key).kind,
                        )
                    ) {
                        continue;
                    }
                }
                const property = this.property(sender, id, this.rules.parser.node(key).text);
                if(!property.present) {
                    continue;
                }
                const type = property.root();
                if((type.flags & 1) !== 0) {
                    this.rules.add(
                        'no-unsafe-assignment',
                        'unsafeObjectPattern',
                        `Unsafe object destructuring of a property with an ${type.error ? 'error typed' : '`any`'} value.`,
                        name,
                    );
                    reported = true;
                }
                else if(
                    [
                        'ArrayBindingPattern',
                        'ArrayLiteralExpression',
                        'ObjectBindingPattern',
                        'ObjectLiteralExpression',
                    ].includes(this.rules.parser.node(name).kind)
                ) {
                    reported = this.destructure(name, property, type.id, sender);
                }
            }
        }
        return reported;
    }
    check(target: number, sender: number, reporting: number, mode: number): void {
        const senderFacts = this.volume.facts(sender);
        const source = senderFacts.root();
        let receiverFacts = this.volume.facts(target);
        if(mode === 2) {
            const context = this.volume.facts(target, 'contextual-shape');
            if(context.present) {
                receiverFacts = context;
            }
        }
        const receiver = receiverFacts.root();
        if((source.flags & 1) !== 0 && (receiver.flags & 2) === 0) {
            const finding = this.rules.add(
                'no-unsafe-assignment',
                'anyAssignment',
                `Unsafe assignment of an ${source.error ? 'error typed' : '`any`'} value.`,
                reporting,
            );
            const node = this.rules.parser.node(reporting);
            if(
                node.kind === 'Parameter' &&
                node.children.some((child) =>
                    ['PrivateKeyword', 'PublicKeyword', 'ProtectedKeyword'].includes(
                        this.rules.parser.node(child).kind,
                    ),
                )
            ) {
                // A parameter property's modifiers are outside the assignment's span.
                this.rules.findings.pop();
                this.rules.findings.push(
                    new Diagnostic(
                        'no-unsafe-assignment',
                        finding.id,
                        finding.message,
                        this.rules.byte(this.rules.start(this.rules.parser.node(target))),
                        finding.end,
                    ),
                );
            }
            return;
        }
        const senderNode = this.rules.parser.node(sender);
        const emptyMap =
            senderNode.kind === 'NewExpression' &&
            senderNode.list === 0 &&
            senderNode.children.length === 1 &&
            this.rules.parser.node(senderNode.children[0] ?? -1).text === 'Map';
        if(mode !== 0 && unsafeAssignment(senderFacts, source, receiverFacts, receiver, emptyMap, [])) {
            this.rules.add(
                'no-unsafe-assignment',
                'unsafeAssignment',
                `Unsafe assignment of type \`${this.rules.name(source, sender)}\` to a variable of type \`${this.rules.name(receiver, target)}\`.`,
                reporting,
            );
            return;
        }
        this.destructure(target, senderFacts, source.id, sender);
    }
    target(index: number): boolean {
        let current = index;
        while((this.rules.parents[current] ?? -1) >= 0) {
            const parentIndex = this.rules.parents[current] ?? -1;
            const parent = this.rules.parser.node(parentIndex);
            if(parent.kind === 'BinaryExpression') {
                return (
                    this.rules.parser.node(parent.children[1] ?? -1).kind === 'EqualsToken' &&
                    parent.children[0] === current
                );
            }
            if(
                ![
                    'PropertyAssignment',
                    'ShorthandPropertyAssignment',
                    'ObjectLiteralExpression',
                    'ArrayLiteralExpression',
                ].includes(parent.kind)
            ) {
                return false;
            }
            current = parentIndex;
        }
        return false;
    }
    visit(index: number): void {
        const node = this.rules.parser.node(index);
        if(['VariableDeclaration', 'PropertyDeclaration', 'Parameter', 'BindingElement'].includes(node.kind)) {
            const sender = this.shadow.initializer(index);
            const target = this.shadow.name(index);
            if(sender >= 0 && target >= 0) {
                const mode =
                    node.kind === 'Parameter' ||
                    node.kind === 'BindingElement' ||
                    node.children.indexOf(sender) - node.children.indexOf(target) > 1
                        ? 1
                        : 0;
                this.check(target, sender, index, mode);
            }
        }
        if(node.kind === 'BinaryExpression' && this.rules.parser.node(node.children[1] ?? -1).kind === 'EqualsToken') {
            this.check(node.children[0] ?? -1, node.children[2] ?? -1, index, 1);
        }
        if(node.kind === 'PropertyAssignment' && !this.target(index)) {
            this.check(node.children[0] ?? -1, node.children[1] ?? -1, index, 2);
        }
        if(node.kind === 'ShorthandPropertyAssignment' && node.children.length === 1 && !this.target(index)) {
            this.check(node.children[0] ?? -1, node.children[0] ?? -1, index, 2);
        }
        if(
            node.kind === 'SpreadElement' &&
            this.rules.parser.node(this.rules.parents[index] ?? -1).kind === 'ArrayLiteralExpression'
        ) {
            const facts = this.volume.facts(node.children[0] ?? -1);
            const type = facts.root();
            if((type.flags & 1) !== 0 || this.anyArray(facts, type.id)) {
                this.rules.add(
                    'no-unsafe-assignment',
                    'unsafeArraySpread',
                    `Unsafe spread of an ${type.error ? 'error typed' : '`any`'} value in an array.`,
                    index,
                );
            }
        }
        for(const child of node.children) {
            this.visit(child);
        }
    }
    run(): void {
        this.visit(this.rules.parser.nodes.length - 1);
    }
}
