import { panic } from 'adamic';
import type { RuleContext } from '../../context.ts';
import type { ParseNode } from '../../../../typescript/parser/nodes.ts';
import { messageConsistencyNoConsole } from './messages.ts';

// Reads: cohere internal/types/source_name/source_name.go, TreatedAs and IsTestFile.
function isTestFile(fileName: string): boolean {
    let treated = fileName;
    if(fileName.endsWith('.a')) {
        const stem = fileName.slice(0, -2);
        if(stem !== '' && !stem.endsWith('/')) {
            treated = `${stem}.ts`;
        }
    }
    for(const extension of ['.ts', '.tsx', '.js', '.jsx', '.mts', '.cts']) {
        if(treated.endsWith(`.test${extension}`)) {
            return true;
        }
    }
    return false;
}

export class Rule {
    readonly context: RuleContext;
    readonly testFile: boolean;
    constructor(context: RuleContext) {
        this.context = context;
        // Reads: ConsistencyNoConsole.Run, which installs no listeners for a test file.
        this.testFile = isTestFile(context.parser.path);
    }
    visit(node: ParseNode, index: number): void {
        if(this.testFile) {
            return;
        }
        const receiver = this.context.unwrap(node.children[0] ?? panic('receiver'));
        const key = node.children[node.children.length - 1] ?? panic('property');
        if(this.context.node(receiver).kind === 'Identifier' && this.context.node(receiver).text === 'console') {
            this.context.report(
                index,
                'base/consistency-no-console',
                'consistencyNoConsole',
                messageConsistencyNoConsole(
                    this.context.node(key).kind === 'Identifier' ? this.context.node(key).text : 'log',
                ),
                '',
                '',
                '',
            );
        }
    }
}

export function create(context: RuleContext): Rule {
    return new Rule(context);
}
