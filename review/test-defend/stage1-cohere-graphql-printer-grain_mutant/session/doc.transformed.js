import { panic } from 'adamic';
import { stringWidth } from '../../json/width.ts';
export const defaults = {
    printWidth: 120,
    tabWidth: 4,
    useTabs: false,
    bracketSpacing: true
};
export class Documents {
    nodes = [];
    settings;
    constructor(settings){
        this.settings = settings;
    }
    get(index) {
        return this.nodes[index] ?? panic(`missing doc ${index}`);
    }
    add(kind, text, parts, broken) {
        let mustBreak = broken;
        if (kind !== 'ifBreak' && kind !== 'suffix') {
            for (const part of parts){
                if (this.get(part).broken) {
                    mustBreak = true;
                }
            }
        }
        const index = this.nodes.length;
        this.nodes.push({
            kind,
            text,
            parts,
            broken: mustBreak,
            width: stringWidth(text)
        });
        return index;
    }
    text(text) {
        return this.add('text', text, [], false);
    }
    concat(parts) {
        return this.add('concat', '', parts, false);
    }
    join(separator, parts) {
        const joined = [];
        for(let position = 0; position < parts.length; position++){
            if (position > 0) joined.push(separator);
            joined.push(parts[position] ?? panic('join'));
        }
        return this.concat(joined);
    }
    indent(child) {
        return this.add('indent', '', [
            child
        ], false);
    }
    group(child) {
        return this.add('group', '', [
            child
        ], false);
    }
    line(soft, hard) {
        return this.add('line', soft ? '' : ' ', [], hard);
    }
    ifBreak(broken, flat) {
        return this.add('ifBreak', '', [
            broken,
            flat
        ], false);
    }
    suffix(child) {
        return this.add('suffix', '', [
            child
        ], false);
    }
    breakParent() {
        return this.add('breakParent', '', [], true);
    }
    fits(candidate, rest, width) {
        let remaining = width;
        const stack = [
            candidate
        ];
        let restIndex = rest.length;
        while(remaining >= 0){
            if (stack.length === 0) {
                if (restIndex === 0) return true;
                restIndex--;
                stack.push(rest[restIndex] ?? panic('rest'));
                continue;
            }
            const command = stack.pop() ?? panic('fit');
            const node = this.get(command.doc);
            if (node.kind === 'suffix' || node.kind === 'breakParent') continue;
            if (node.kind === 'text') {
                remaining -= node.width;
                continue;
            }
            if (node.kind === 'line') {
                if (!command.flat || node.broken) return true;
                remaining -= node.width;
                continue;
            }
            if (node.kind === 'ifBreak') {
                stack.push({
                    doc: node.parts[command.flat ? 1 : 0] ?? panic('ifBreak'),
                    indent: 0,
                    flat: command.flat
                });
                continue;
            }
            const flat = node.kind === 'group' && node.broken ? false : command.flat;
            for(let position = node.parts.length - 1; position >= 0; position--)stack.push({
                doc: node.parts[position] ?? panic('fit child'),
                indent: 0,
                flat
            });
        }
        return false;
    }
    print(root) {
        const stack = [
            {
                doc: root,
                indent: 0,
                flat: false
            }
        ];
        const output = [];
        const suffixes = [];
        let column = 0;
        let pendingIndent = 0;
        while(stack.length > 0 || suffixes.length > 0){
            if (stack.length === 0) {
                while(suffixes.length > 0)stack.push(suffixes.pop() ?? panic('suffix'));
            }
            const command = stack.pop() ?? panic('print');
            const node = this.get(command.doc);
            if (node.kind === 'text') {
                if (node.text !== '' && pendingIndent > 0) {
                    output.push((this.settings.useTabs ? '\t' : ' ').repeat(this.settings.useTabs ? pendingIndent / this.settings.tabWidth : pendingIndent));
                    pendingIndent = 0;
                }
                output.push(node.text);
                column += node.width;
            } else if (node.kind === 'suffix') {
                suffixes.push({
                    doc: node.parts[0] ?? panic('suffix child'),
                    indent: command.indent,
                    flat: command.flat
                });
            } else if (node.kind === 'line') {
                if (suffixes.length > 0 && (!command.flat || node.broken)) {
                    stack.push(command);
                    while(suffixes.length > 0)stack.push(suffixes.pop() ?? panic('suffix'));
                    continue;
                }
                if (command.flat && !node.broken) {
                    output.push(node.text);
                    column += node.width;
                } else {
                    while(output.length > 0){
                        const tail = output.pop() ?? panic('trim');
                        let end = tail.length;
                        while(end > 0 && (tail.charCodeAt(end - 1) === 32 || tail.charCodeAt(end - 1) === 9))end--;
                        const trimmed = tail.slice(0, end);
                        if (trimmed !== '') {
                            output.push(trimmed);
                            break;
                        }
                    }
                    output.push('\n');
                    pendingIndent = command.indent;
                    column = command.indent;
                }
            } else if (node.kind === 'group') {
                const child = node.parts[0] ?? panic('group');
                const candidate = {
                    doc: child,
                    indent: command.indent,
                    flat: true
                };
                const flat = !node.broken && (command.flat || this.fits(candidate, stack, this.settings.printWidth - column));
                stack.push({
                    doc: child,
                    indent: command.indent,
                    flat
                });
            } else if (node.kind === 'ifBreak') {
                stack.push({
                    doc: node.parts[command.flat ? 1 : 0] ?? panic('branch'),
                    indent: command.indent,
                    flat: command.flat
                });
            } else {
                const indent = node.kind === 'indent' ? command.indent + this.settings.tabWidth : command.indent;
                for(let position = node.parts.length - 1; position >= 0; position--)stack.push({
                    doc: node.parts[position] ?? panic('child'),
                    indent,
                    flat: command.flat
                });
            }
        }
        return output.join('');
    }
}
