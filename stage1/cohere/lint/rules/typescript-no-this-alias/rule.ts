import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import { OptionsJson } from '../../helpers/options_json.ts';
import { assignment, destructure } from './messages.ts';
export class Rule {
    readonly context: RuleContext;
    readonly reportDestructuring: boolean;
    readonly allowed: readonly string[];
    constructor(context: RuleContext, reportDestructuring: boolean, allowed: readonly string[]) {
        this.context = context;
        this.reportDestructuring = reportDestructuring;
        this.allowed = allowed;
    }
    visit(index: number): void {
        if(!['.ts', '.tsx', '.mts', '.cts'].some((extension) => this.context.parser.path.endsWith(extension))) {
            return;
        }
        const node = this.context.node(index);
        let target = node.children[0] ?? panic('alias target');
        const right = node.children[node.children.length - 1] ?? panic('alias initializer');
        let pattern = node.kind === 'VariableDeclaration';
        if(this.context.node(right).kind !== 'ThisKeyword') {
            return;
        }
        if(node.kind === 'BinaryExpression') {
            const operator = this.context.node(node.children[1] ?? panic('assignment operator')).kind;
            if(
                ![
                    'EqualsToken',
                    'PlusEqualsToken',
                    'MinusEqualsToken',
                    'AsteriskEqualsToken',
                    'AsteriskAsteriskEqualsToken',
                    'SlashEqualsToken',
                    'PercentEqualsToken',
                    'LessThanLessThanEqualsToken',
                    'GreaterThanGreaterThanEqualsToken',
                    'GreaterThanGreaterThanGreaterThanEqualsToken',
                    'AmpersandEqualsToken',
                    'BarEqualsToken',
                    'CaretEqualsToken',
                    'AmpersandAmpersandEqualsToken',
                    'BarBarEqualsToken',
                    'QuestionQuestionEqualsToken',
                ].includes(operator)
            ) {
                return;
            }
            const kind = this.context.node(target).kind;
            pattern = kind === 'ArrayLiteralExpression' || kind === 'ObjectLiteralExpression';
            if(!pattern) {
                while(
                    [
                        'ParenthesizedExpression',
                        'AsExpression',
                        'SatisfiesExpression',
                        'TypeAssertionExpression',
                        'NonNullExpression',
                    ].includes(this.context.node(target).kind)
                ) {
                    const wrapper = this.context.node(target);
                    target =
                        wrapper.children[wrapper.kind === 'TypeAssertionExpression' ? 1 : 0] ??
                        panic('assignment wrapper');
                }
            }
        }
        const kind = this.context.node(target).kind;
        if(kind === 'Identifier') {
            if(this.allowed.includes(this.context.node(target).text)) {
                return;
            }
            this.context.report(target, '@typescript-eslint/no-this-alias', 'thisAssignment', assignment, '', '', '');
        }
        else if(
            this.reportDestructuring &&
            pattern &&
            [
                'ObjectBindingPattern',
                'ArrayBindingPattern',
                'ObjectLiteralExpression',
                'ArrayLiteralExpression',
            ].includes(kind)
        ) {
            this.context.report(target, '@typescript-eslint/no-this-alias', 'thisDestructure', destructure, '', '', '');
        }
    }
}
export function create(context: RuleContext): Rule {
    let report = false;
    let allowed: string[] = [];
    if(!context.enabled('@typescript-eslint/no-this-alias')) {
        return new Rule(context, report, allowed);
    }
    const json = new OptionsJson(context.settings.text === '' ? 'null' : context.settings.text);
    const root = json.parse();
    if(root < 0) {
        panic('invalid this-alias JSON');
    }
    const object = json.node(root);
    if(object.kind === 'null') {
        return new Rule(context, report, allowed);
    }
    if(object.kind !== 'object') {
        panic('this-alias options must be an object');
    }
    let correct: string[] = [];
    let hasCorrect = false;
    let alias: string[] = [];
    for(let optionIndex = 0; optionIndex < object.keys.length; optionIndex++) {
        const key = object.keys[optionIndex] ?? panic('option key');
        const value = json.node(object.children[optionIndex] ?? panic('option value'));
        if(key === 'allowDestructuring' || key.toLowerCase() === 'reportdestructuring') {
            if(value.kind === 'null') {
                if(key === 'allowDestructuring') {
                    report = false;
                }
                continue;
            }
            if(value.kind !== 'boolean') {
                panic('this-alias option must be boolean');
            }
            report = key === 'allowDestructuring' ? value.text === 'false' : value.text === 'true';
        }
        else if(key === 'allowNames' || key.toLowerCase() === 'allowednames') {
            if(value.kind === 'null') {
                if(key === 'allowNames') {
                    alias = [];
                }
                else {
                    correct = [];
                    hasCorrect = false;
                }
                continue;
            }
            if(value.kind !== 'array') {
                panic('this-alias names must be an array');
            }
            const names: string[] = [];
            for(const child of value.children) {
                const name = json.node(child);
                if(name.kind === 'null') {
                    names.push('');
                }
                else if(name.kind === 'string') {
                    names.push(name.text);
                }
                else {
                    panic('this-alias name must be a string');
                }
            }
            if(key === 'allowNames') {
                alias = names;
            }
            else {
                correct = names;
                hasCorrect = true;
            }
        }
        else {
            panic('unknown this-alias option');
        }
    }
    allowed = hasCorrect ? correct : alias;
    return new Rule(context, report, allowed);
}
