import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import type { ParseNode } from '../../../../typescript/parser/nodes.ts';
import {
    messageMethodSignatureStyleConvertToMethod,
    messageMethodSignatureStyleErrorMethod,
    messageMethodSignatureStyleErrorProperty,
} from './messages.ts';

const rule = '@typescript-eslint/method-signature-style';

export class Rule {
    readonly context: RuleContext;
    constructor(context: RuleContext) {
        this.context = context;
    }
    visit(node: ParseNode, index: number): void {
        const style = this.context.settings.read('style', 'property');
        const key = this.signatureKey(index);
        if(key === '') {
            return;
        }
        if(style === 'property' && node.kind === 'MethodSignature') {
            this.method(index, key);
        }
        if(style !== 'property' && node.kind === 'PropertySignature') {
            this.property(index, key);
        }
    }
    // method rewrites a method signature as a property holding a function type, merging an overload group
    // into one intersection at its first member.
    method(index: number, key: string): void {
        let ancestor = this.context.parent(index);
        let inModule = false;
        while(ancestor >= 0) {
            if(this.context.node(ancestor).kind === 'ModuleDeclaration') {
                inModule = true;
                break;
            }
            ancestor = this.context.parent(ancestor);
        }
        const type = this.returnType(index);
        const skip = inModule || (type >= 0 && this.containsKind(type, 'ThisType'));
        const message = messageMethodSignatureStyleErrorMethod;
        if(skip) {
            this.report(index, 'errorMethod', message);
            return;
        }
        const owner = this.context.parent(index);
        const group =
            owner < 0
                ? [index]
                : this.context
                      .node(owner)
                      .children.filter(
                          (child) =>
                              this.context.node(child).kind === 'MethodSignature' && this.signatureKey(child) === key,
                      );
        if(group.length > 1) {
            if(group[0] !== index) {
                this.report(index, 'errorMethod', message);
                return;
            }
            const parts: string[] = [];
            for(const member of group) {
                const parameters = this.parametersText(member);
                if(parameters === '') {
                    this.report(index, 'errorMethod', message);
                    return;
                }
                const returnType = this.returnType(member);
                parts.push(`(${parameters} => ${returnType < 0 ? 'any' : this.context.raw(returnType)})`);
            }
            const last = group[group.length - 1] ?? panic('last overload');
            const finding = this.context.report(
                index,
                rule,
                'errorMethod',
                message,
                'fix',
                `${key}: ${parts.join(' & ')}${this.delimiter(last)}`,
                '',
            );
            finding.editEnd = this.context.node(last).end;
            return;
        }
        const parameters = this.parametersText(index);
        if(parameters === '') {
            this.report(index, 'errorMethod', message);
            return;
        }
        this.context.report(
            index,
            rule,
            'errorMethod',
            message,
            'fix',
            `${key}: ${parameters} => ${type < 0 ? 'any' : this.context.raw(type)}${this.delimiter(index)}`,
            '',
        );
    }
    // property rewrites a property holding a function type as a method signature, offering rather than
    // applying it when the property is readonly, since a method cannot be.
    property(index: number, key: string): void {
        const type = this.returnType(index);
        if(type < 0 || this.context.node(type).kind !== 'FunctionType') {
            return;
        }
        const message = messageMethodSignatureStyleErrorProperty;
        const parameters = this.parametersText(type);
        if(parameters === '') {
            this.report(index, 'errorProperty', message);
            return;
        }
        const returnType = this.returnType(type);
        const readonly = this.context.has(index, 'ReadonlyKeyword');
        this.context.report(
            index,
            rule,
            'errorProperty',
            message,
            readonly ? 'suggestion' : 'fix',
            `${key}${parameters}: ${returnType < 0 ? 'any' : this.context.raw(returnType)}${this.delimiter(index)}`,
            readonly ? messageMethodSignatureStyleConvertToMethod : '',
        );
    }
    report(index: number, id: string, message: string): void {
        this.context.report(index, rule, id, message, '', '', '');
    }
    containsKind(index: number, kind: string): boolean {
        return (
            this.context.node(index).kind === kind ||
            this.context.node(index).children.some((child) => this.containsKind(child, kind))
        );
    }
    returnType(index: number): number {
        const node = this.context.node(index);
        const last = node.children[node.children.length - 1] ?? -1;
        if(last < 0 || last === this.context.name(index)) {
            return -1;
        }
        const kind = this.context.node(last).kind;
        return kind === 'Parameter' ||
            kind === 'TypeParameter' ||
            kind === 'QuestionToken' ||
            kind === 'ReadonlyKeyword'
            ? -1
            : last;
    }
    parametersText(index: number): string {
        const parameters = this.context.children(index, 'Parameter');
        const generics = this.context.children(index, 'TypeParameter');
        const floor = this.context.start(index);
        const end = this.context.node(index).end;
        let result = '()';
        if(parameters.length > 0) {
            const first = parameters[0] ?? panic('first parameter');
            const last = parameters[parameters.length - 1] ?? panic('last parameter');
            const opening = this.context.source.slice(0, this.context.start(first)).lastIndexOf('(');
            const closing = this.context.source.indexOf(')', this.context.node(last).end);
            if(opening < floor || closing < 0 || closing >= end) {
                return '';
            }
            result = this.context.source.slice(opening, closing + 1);
        }
        if(generics.length > 0) {
            const first = generics[0] ?? panic('first generic');
            const last = generics[generics.length - 1] ?? panic('last generic');
            const opening = this.context.source.slice(0, this.context.start(first)).lastIndexOf('<');
            const closing = this.context.source.indexOf('>', this.context.node(last).end);
            if(opening < floor || closing < 0 || closing >= end) {
                return '';
            }
            result = this.context.source.slice(opening, closing + 1) + result;
        }
        return result;
    }
    delimiter(index: number): string {
        const last = this.context.source[this.context.node(index).end - 1] ?? '';
        return last === ';' || last === ',' ? last : '';
    }
    signatureKey(index: number): string {
        const name = this.context.name(index);
        return name < 0 ? '' : this.context.raw(name) + (this.context.has(index, 'QuestionToken') ? '?' : '');
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
