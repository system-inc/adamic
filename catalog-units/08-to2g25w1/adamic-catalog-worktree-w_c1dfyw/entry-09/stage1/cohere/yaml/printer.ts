// Go cohere's YAML printer. Positions and links remain arena indices through layout.
import type { UnistContext } from './unistContext.ts';
import { Layout } from './layout.ts';
export class YAMLPrinter {
    context: UnistContext;
    layout = new Layout();
    printedEmpty: number[] = [];
    hard: number;
    soft: number;
    line: number;
    empty: number;
    groups = 0;
    constructor(context: UnistContext) {
        this.context = context;
        this.hard = this.layout.line(false, true, false);
        this.soft = this.layout.line(true, false, false);
        this.line = this.layout.line(false, false, false);
        this.empty = this.layout.text('');
    }
    node(index: number) {
        return this.context.node(index);
    }
    child(index: number, at: number): number {
        return this.node(index).children[at] ?? -1;
    }
    raw(index: number): string {
        return this.context.text.slice(this.context.start(index).offset, this.context.end(index).offset);
    }
    inline(index: number): boolean {
        return (
            index < 0 ||
            ['plain', 'quoteDouble', 'quoteSingle', 'alias', 'flowMapping', 'flowSequence'].includes(
                this.node(index).type,
            )
        );
    }
    single(index: number): boolean {
        return (
            index < 0 ||
            this.node(index).type === 'alias' ||
            (['plain', 'quoteDouble', 'quoteSingle'].includes(this.node(index).type) &&
                this.context.start(index).line === this.context.end(index).line)
        );
    }
    comments(index: number): boolean {
        const node = this.node(index);
        return (
            node.leadingComments.length > 0 ||
            node.middleComments.length > 0 ||
            node.indicatorComment >= 0 ||
            node.trailingComment >= 0 ||
            node.endComments.length > 0
        );
    }
    emptyNode(index: number): boolean {
        return this.node(index).children.length === 0 && !this.comments(index);
    }
    printEnds(index: number): boolean {
        return (
            this.node(index).endComments.length > 0 &&
            !['documentHead', 'documentBody', 'flowMapping', 'flowSequence'].includes(this.node(index).type)
        );
    }
    previousEmpty(offset: number): boolean {
        let index = offset - 1;
        while(
            index >= 0 &&
            (this.context.text.slice(index, index + 1) === ' ' || this.context.text.slice(index, index + 1) === '\t')
        )
            index--;
        if(this.context.text.slice(index, index + 1) === '\n') index--;
        if(this.context.text.slice(index, index + 1) === '\r') index--;
        while(
            index >= 0 &&
            (this.context.text.slice(index, index + 1) === ' ' || this.context.text.slice(index, index + 1) === '\t')
        )
            index--;
        return (
            index >= 0 &&
            (this.context.text.slice(index, index + 1) === '\n' || this.context.text.slice(index, index + 1) === '\r')
        );
    }
    nextEmpty(index: number, parent: number): number {
        const end = this.context.end(index).offset;
        if(this.printedEmpty.includes(end)) return this.empty;
        this.printedEmpty.push(end);
        let count = 0;
        for(let offset = Math.max(0, end - 1); offset < this.context.text.length; offset++) {
            const character = this.context.text.slice(offset, offset + 1);
            if(character === '\n') count++;
            if(count === 1 && character.trim() !== '') return this.empty;
            if(count === 2) return this.printEnds(parent) ? this.empty : this.soft;
        }
        return this.empty;
    }
    last(index: number, ancestors: readonly number[], listPositions: readonly number[]): boolean {
        if(['tag', 'anchor', 'comment'].includes(this.node(index).type)) return false;
        for(let at = 0; at < ancestors.length; at++) {
            const position = listPositions[at] ?? -1;
            if(position >= 0 && position !== this.node(ancestors[at] ?? -1).children.length - 1) return false;
        }
        return true;
    }
    descendant(index: number): number {
        const children = this.node(index).children;
        return children.length === 0 ? index : this.descendant(children[children.length - 1] ?? -1);
    }
    list(
        items: readonly number[],
        ancestors: readonly number[],
        positions: readonly number[],
        childList: boolean,
    ): number[] {
        const result: number[] = [];
        for(let index = 0; index < items.length; index++)
            result.push(
                this.print(
                    items[index] ?? -1,
                    ancestors,
                    childList ? positions.concat([index]) : positions.concat([-1]),
                ),
            );
        return result;
    }
    print(index: number, ancestors: readonly number[], positions: readonly number[]): number {
        if(index < 0) return this.empty;
        const node = this.node(index);
        const parent = ancestors[ancestors.length - 1] ?? -1;
        const path = ancestors.concat([index]);
        const parts: number[] = [];
        if(node.type !== 'mappingValue' && node.leadingComments.length > 0) {
            parts.push(this.layout.join(this.hard, this.list(node.leadingComments, path, positions, false)));
            parts.push(this.hard);
        }
        if(node.tag >= 0) parts.push(this.print(node.tag, path, positions.concat([-1])));
        if(node.tag >= 0 && node.anchor >= 0) parts.push(this.layout.text(' '));
        if(node.anchor >= 0) parts.push(this.print(node.anchor, path, positions.concat([-1])));
        let nextEmpty = this.empty;
        if(
            ['mapping', 'sequence', 'comment', 'directive', 'mappingItem', 'sequenceItem'].includes(node.type) &&
            !this.last(index, ancestors, positions)
        )
            nextEmpty = this.nextEmpty(index, parent);
        if(node.tag >= 0 || node.anchor >= 0)
            parts.push(
                ['sequence', 'mapping'].includes(node.type) && node.middleComments.length === 0
                    ? this.hard
                    : this.layout.text(' '),
            );
        if(node.middleComments.length > 0) {
            parts.push(node.middleComments.length === 1 ? this.empty : this.hard);
            parts.push(this.layout.join(this.hard, this.list(node.middleComments, path, positions, false)));
            parts.push(this.hard);
        }
        let ignored = false;
        const ignoreComments =
            node.type === 'documentBody' ? this.node(this.child(parent, 0)).endComments : node.leadingComments;
        if(ignoreComments.length > 0)
            ignored = this.node(ignoreComments[ignoreComments.length - 1] ?? -1).value.trim() === 'prettier-ignore';
        if(ignored) {
            const raw = this.raw(index).trimEnd().split('\n');
            const docs: number[] = [];
            for(let line = 0; line < raw.length; line++) {
                if(line > 0) docs.push(this.layout.line(false, true, true));
                docs.push(this.layout.text(raw[line] ?? ''));
            }
            parts.push(this.layout.concat(docs));
        }
        else parts.push(this.layout.group(this.printNode(index, path, positions), -1, false));
        if(node.trailingComment >= 0 && node.type !== 'document' && node.type !== 'documentHead') {
            const suffix: number[] = [
                this.layout.text(node.type === 'mappingValue' && this.child(index, 0) < 0 ? '' : ' '),
            ];
            const grandparent = ancestors[ancestors.length - 3] ?? -1;
            if(!(
                this.node(parent).type === 'mappingKey' &&
                this.node(grandparent).type === 'mapping' &&
                this.inline(index)
            ))
                suffix.push(this.layout.make('break', []));
            suffix.push(this.print(node.trailingComment, path, positions.concat([-1])));
            parts.push(this.layout.make('suffix', [this.layout.concat(suffix)]));
        }
        if(this.printEnds(index)) {
            const ends: number[] = [];
            for(const comment of node.endComments)
                ends.push(
                    this.layout.concat([
                        this.previousEmpty(this.context.start(comment).offset) ? this.hard : this.empty,
                        this.print(comment, path, positions.concat([-1])),
                    ]),
                );
            parts.push(
                this.layout.align(
                    node.type === 'sequenceItem' ? 2 : 0,
                    this.layout.concat([this.hard, this.layout.join(this.hard, ends)]),
                ),
            );
        }
        parts.push(nextEmpty);
        return this.layout.concat(parts);
    }
    scalar(type: string, content: string): number {
        const lines = content.split('\n');
        const docs: number[] = [];
        for(let index = 0; index < lines.length; index++) {
            let line = lines[index] ?? '';
            if(index > 0 && index < lines.length - 1) line = line.trim();
            else if(index === 0 && index < lines.length - 1) line = line.trimEnd();
            else if(index > 0) line = line.trimStart();
            docs.push(line === '' ? this.layout.make('fill', []) : this.layout.make('fill', [this.layout.text(line)]));
        }
        return this.layout.join(this.hard, docs);
    }
    printChild(index: number, at: number, path: readonly number[], positions: readonly number[]): number {
        return this.print(this.child(index, at), path, positions.concat([at]));
    }
    printNode(index: number, path: readonly number[], positions: readonly number[]): number {
        const node = this.node(index);
        const parent = path[path.length - 2] ?? -1;
        if(node.type === 'root') {
            const last = this.node(this.descendant(index));
            const hard = !(['blockLiteral', 'blockFolded'].includes(last.type) && last.chomping === 'keep');
            const parts: number[] = [];
            for(let at = 0; at < node.children.length; at++) {
                const document = node.children[at] ?? -1;
                const next = node.children[at + 1] ?? -1;
                const head = this.child(next, 0);
                if(at > 0) parts.push(this.hard);
                parts.push(this.printChild(index, at, path, positions));
                if(
                    this.node(document).documentEndMarker ||
                    this.node(document).trailingComment >= 0 ||
                    (next >= 0 && (this.node(head).children.length > 0 || this.node(head).endComments.length > 0))
                ) {
                    if(hard) parts.push(this.hard);
                    parts.push(this.layout.text('...'));
                    if(this.node(document).trailingComment >= 0)
                        parts.push(
                            this.layout.concat([
                                this.layout.text(' '),
                                this.print(
                                    this.node(document).trailingComment,
                                    path.concat([document]),
                                    positions.concat([at, -1]),
                                ),
                            ]),
                        );
                }
            }
            if(hard) parts.push(this.hard);
            return this.layout.concat(parts);
        }
        if(node.type === 'document') {
            const head = this.child(index, 0);
            const body = this.child(index, 1);
            const parts: number[] = [];
            if(
                node.directivesEndMarker ||
                this.node(head).children.length > 0 ||
                this.node(head).endComments.length > 0 ||
                this.node(head).trailingComment >= 0
            ) {
                if(this.node(head).children.length > 0 || this.node(head).endComments.length > 0)
                    parts.push(this.printChild(index, 0, path, positions));
                const marker: number[] = [this.layout.text('---')];
                if(this.node(head).trailingComment >= 0)
                    marker.push(
                        this.layout.concat([
                            this.layout.text(' '),
                            this.print(this.node(head).trailingComment, path.concat([head]), positions.concat([0, -1])),
                        ]),
                    );
                parts.push(this.layout.concat(marker));
            }
            if(this.node(body).children.length > 0 || this.node(body).endComments.length > 0)
                parts.push(this.printChild(index, 1, path, positions));
            return this.layout.join(this.hard, parts);
        }
        if(node.type === 'documentHead')
            return this.layout.join(
                this.hard,
                this.list(node.children, path, positions, true).concat(
                    this.list(node.endComments, path, positions, false),
                ),
            );
        if(node.type === 'documentBody') {
            let separator = this.empty;
            if(node.children.length > 0 && node.endComments.length > 0) {
                const last = this.node(this.descendant(index));
                if(['blockFolded', 'blockLiteral'].includes(last.type)) {
                    if(last.chomping !== 'keep') separator = this.layout.concat([this.hard, this.hard]);
                }
                else
                    separator =
                        this.node(node.children[node.children.length - 1] ?? -1).type === 'mapping' &&
                        this.previousEmpty(this.context.start(node.endComments[0] ?? -1).offset)
                            ? this.layout.concat([this.hard, this.hard])
                            : this.hard;
            }
            return this.layout.concat([
                this.layout.join(this.hard, this.list(node.children, path, positions, true)),
                separator,
                this.layout.join(this.hard, this.list(node.endComments, path, positions, false)),
            ]);
        }
        if(node.type === 'directive') return this.layout.text(`%${[node.name].concat(node.parameters).join(' ')}`);
        if(node.type === 'comment') return this.layout.text(`#${node.value}`);
        if(node.type === 'alias') return this.layout.text(`*${node.value}`);
        if(node.type === 'anchor') return this.layout.text(`&${node.value}`);
        if(node.type === 'tag') return this.layout.text(this.raw(index));
        if(node.type === 'plain') return this.scalar(node.type, this.raw(index));
        if(node.type === 'quoteDouble' || node.type === 'quoteSingle') {
            let raw = this.raw(index).slice(1, -1);
            let otherEscape = false;
            for(let at = 0; at + 1 < raw.length; at++)
                if(raw.slice(at, at + 1) === '\\' && raw.slice(at + 1, at + 2) !== '"') otherEscape = true;
            let quote = '"';
            if((node.type === 'quoteSingle' && raw.includes('\\')) || (node.type === 'quoteDouble' && otherEscape))
                quote = node.type === 'quoteSingle' ? "'" : '"';
            else if(raw.includes('"')) {
                quote = "'";
                if(node.type === 'quoteDouble') raw = raw.split('\\"').join('"').split("'").join("''");
            }
            else if(raw.includes("'")) {
                quote = '"';
                if(node.type === 'quoteSingle') raw = raw.split("''").join("'");
            }
            return this.layout.concat([this.layout.text(quote), this.scalar(node.type, raw), this.layout.text(quote)]);
        }
        if(node.type === 'blockFolded' || node.type === 'blockLiteral') return this.block(index, path, positions);
        if(node.type === 'mapping' || node.type === 'sequence')
            return this.layout.join(this.hard, this.list(node.children, path, positions, true));
        if(node.type === 'sequenceItem')
            return this.layout.concat([
                this.layout.text('- '),
                this.layout.align(2, this.printChild(index, 0, path, positions)),
            ]);
        if(node.type === 'mappingKey' || node.type === 'mappingValue' || node.type === 'flowSequenceItem')
            return this.printChild(index, 0, path, positions);
        if(node.type === 'mappingItem' || node.type === 'flowMappingItem')
            return this.mapping(index, parent, path, positions);
        if(node.type === 'flowMapping' || node.type === 'flowSequence') return this.flow(index, path, positions);
        this.context.fail(`Unexpected node type "${node.type}" in YAML`);
        return this.empty;
    }
    mapping(index: number, parent: number, path: readonly number[], positions: readonly number[]): number {
        const node = this.node(index);
        const key = this.child(index, 0);
        const value = this.child(index, 1);
        const keyContent = this.child(key, 0);
        const valueContent = this.child(value, 0);
        const emptyKey = this.emptyNode(key);
        const emptyValue = this.emptyNode(value);
        if(emptyKey && emptyValue) return this.layout.text(': ');
        const printedKey = this.print(key, path, positions.concat([0]));
        const space = this.layout.text(this.node(keyContent).type === 'alias' ? ' ' : '');
        if(emptyValue) {
            if(node.type === 'flowMappingItem' && this.node(parent).type === 'flowMapping') return printedKey;
            if(
                node.type === 'mappingItem' &&
                this.single(keyContent) &&
                this.node(keyContent).trailingComment < 0 &&
                this.node(this.node(parent).tag).value !== 'tag:yaml.org,2002:set'
            )
                return this.layout.concat([printedKey, space, this.layout.text(':')]);
            return this.layout.concat([this.layout.text('? '), this.layout.align(2, printedKey)]);
        }
        const printedValue = this.print(value, path, positions.concat([1]));
        if(emptyKey) return this.layout.concat([this.layout.text(': '), this.layout.align(2, printedValue)]);
        if(this.node(value).leadingComments.length > 0 || !this.inline(keyContent)) {
            const parts: number[] = [this.layout.text('? '), this.layout.align(2, printedKey), this.hard];
            for(const comment of this.node(value).leadingComments) {
                parts.push(this.print(comment, path.concat([value]), positions.concat([1, -1])));
                parts.push(this.hard);
            }
            parts.push(this.layout.text(': '));
            parts.push(this.layout.align(2, printedValue));
            return this.layout.concat(parts);
        }
        if(
            this.single(keyContent) &&
            this.single(valueContent) &&
            this.node(keyContent).leadingComments.length === 0 &&
            this.node(keyContent).middleComments.length === 0 &&
            this.node(keyContent).trailingComment < 0 &&
            this.node(key).endComments.length === 0 &&
            this.node(valueContent).leadingComments.length === 0 &&
            this.node(valueContent).middleComments.length === 0 &&
            this.node(value).endComments.length === 0
        )
            return this.layout.concat([printedKey, space, this.layout.text(': '), printedValue]);
        const implicit: number[] = [space, this.layout.text(':')];
        const content = this.node(valueContent);
        if(
            this.node(value).endComments.length > 0 &&
            ['flowMapping', 'flowSequence'].includes(content.type) &&
            content.children.length === 0
        )
            implicit.push(this.layout.text(' '));
        else if(
            content.leadingComments.length > 0 ||
            (this.node(value).endComments.length > 0 &&
                valueContent >= 0 &&
                !['mapping', 'sequence'].includes(content.type)) ||
            (this.node(parent).type === 'mapping' &&
                this.node(keyContent).trailingComment >= 0 &&
                this.inline(valueContent)) ||
            (['mapping', 'sequence'].includes(content.type) && content.tag < 0 && content.anchor < 0)
        )
            implicit.push(this.hard);
        else if(valueContent >= 0) implicit.push(this.line);
        else if(this.node(value).trailingComment >= 0) implicit.push(this.layout.text(' '));
        implicit.push(printedValue);
        const implicitValue = this.layout.align(2, this.layout.concat(implicit));
        if(
            this.single(keyContent) &&
            this.node(keyContent).leadingComments.length === 0 &&
            this.node(keyContent).middleComments.length === 0 &&
            this.node(keyContent).trailingComment < 0 &&
            this.node(key).endComments.length === 0
        )
            return this.layout.group(this.layout.concat([printedKey, implicitValue]), -1, true);
        const group = this.groups;
        this.groups++;
        const groupedKey = this.layout.group(
            this.layout.concat([
                this.layout.ifBreak(this.layout.text('? '), this.empty, -1),
                this.layout.group(this.layout.align(2, printedKey), group, false),
            ]),
            -1,
            false,
        );
        const explicitValue = this.layout.concat([
            this.hard,
            this.layout.text(': '),
            this.layout.align(2, printedValue),
        ]);
        return this.layout.group(
            this.layout.concat([groupedKey, this.layout.ifBreak(explicitValue, implicitValue, group)]),
            -1,
            true,
        );
    }
    flow(index: number, path: readonly number[], positions: readonly number[]): number {
        const node = this.node(index);
        const map = node.type === 'flowMapping';
        const spacing = map && node.children.length > 0 ? this.line : this.soft;
        const contents: number[] = [spacing];
        for(let at = 0; at < node.children.length; at++) {
            const child = node.children[at] ?? -1;
            contents.push(this.print(child, path, positions.concat([at])));
            if(at + 1 < node.children.length) {
                const next = node.children[at + 1] ?? -1;
                contents.push(this.layout.text(','));
                contents.push(this.line);
                if(this.context.start(child).line !== this.context.start(next).line)
                    contents.push(this.nextEmpty(child, index));
            }
            else contents.push(this.empty);
        }
        contents.push(this.layout.ifBreak(this.layout.text(','), this.empty, -1));
        if(node.endComments.length > 0)
            contents.push(
                this.layout.concat([
                    this.hard,
                    this.layout.join(this.hard, this.list(node.endComments, path, positions, false)),
                ]),
            );
        else contents.push(this.empty);
        const last = node.children[node.children.length - 1] ?? -1;
        const closing =
            last >= 0 &&
            this.node(last).type === 'flowMappingItem' &&
            this.emptyNode(this.child(last, 0)) &&
            this.emptyNode(this.child(last, 1))
                ? this.empty
                : spacing;
        return this.layout.concat([
            this.layout.text(map ? '{' : '['),
            this.layout.align(2, this.layout.concat(contents)),
            closing,
            this.layout.text(map ? '}' : ']'),
        ]);
    }
    block(index: number, path: readonly number[], positions: readonly number[]): number {
        const node = this.node(index);
        let parentIndent = 0;
        for(const ancestor of path.slice(0, -1))
            if(this.node(ancestor).type === 'sequence' || this.node(ancestor).type === 'mapping') parentIndent++;
        const last = this.last(index, path.slice(0, -1), positions);
        const parts: number[] = [this.layout.text(node.type === 'blockFolded' ? '>' : '|')];
        if(node.indent >= 0) parts.push(this.layout.text(String(node.indent)));
        if(node.chomping !== 'clip') parts.push(this.layout.text(node.chomping === 'keep' ? '+' : '-'));
        if(node.indicatorComment >= 0)
            parts.push(
                this.layout.concat([
                    this.layout.text(' '),
                    this.print(node.indicatorComment, path, positions.concat([-1])),
                ]),
            );
        let content = '';
        if(this.context.start(index).line !== this.context.end(index).line) {
            const raw = this.raw(index);
            content = raw.slice(raw.indexOf('\n') + 1);
        }
        const contents: number[] = [];
        if(content !== '') {
            let leading = node.indent >= 0 ? node.indent - 1 + parentIndent : -1;
            const rawLines = content.split('\n');
            if(node.indent < 0)
                for(const line of rawLines) {
                    let count = 0;
                    while(line.slice(count, count + 1) === ' ') count++;
                    const character = line.slice(count, count + 1);
                    if(character !== '' && character !== '\r') {
                        leading = count;
                        break;
                    }
                }
            let lines: string[] = [];
            for(const line of rawLines) lines.push(leading < 0 ? '' : line.slice(leading));
            if(node.chomping === 'keep') {
                if(content.endsWith('\n') && (lines[lines.length - 1] ?? '') === '') lines = lines.slice(0, -1);
            }
            else {
                let count = 0;
                for(let at = lines.length - 1; at >= 0; at--) {
                    if((lines[at] ?? '').split(' ').join('').split('\t').join('') !== '') break;
                    count++;
                }
                if(count > 0) lines = lines.slice(0, lines.length - (count >= 2 && !last ? count - 1 : count));
            }
            for(let at = 0; at < lines.length; at++) {
                const line = lines[at] ?? '';
                if(at === 0) contents.push(this.hard);
                contents.push(
                    line === '' ? this.layout.make('fill', []) : this.layout.make('fill', [this.layout.text(line)]),
                );
                if(at !== lines.length - 1)
                    contents.push(
                        line === '' ? this.hard : this.layout.align(-998, this.layout.line(false, true, true)),
                    );
                else if(node.chomping === 'keep' && last)
                    contents.push(this.layout.align(-999, this.layout.line(false, true, line !== '')));
            }
        }
        parts.push(
            node.indent < 0
                ? this.layout.align(-1, this.layout.align(2, this.layout.concat(contents)))
                : this.layout.align(
                      -999,
                      this.layout.align(node.indent - 1 + parentIndent, this.layout.concat(contents)),
                  ),
        );
        return this.layout.concat(parts);
    }
    format(root: number): string {
        return this.layout.print(this.print(root, [], []));
    }
}
