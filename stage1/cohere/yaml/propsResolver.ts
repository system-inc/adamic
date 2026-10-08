import type { CSTParser } from './cstParser.ts';
import type { ScalarResolution } from './scalarResolution.ts';
import { ComposeError } from './composeError.ts';
import { Props } from './props.ts';
export class PropsResolver {
    parser: CSTParser;
    diagnostics: ScalarResolution;
    constructor(parser: CSTParser, diagnostics: ScalarResolution) {
        this.parser = parser;
        this.diagnostics = diagnostics;
    }
    resolve(
        tokens: readonly number[],
        flow: string,
        indicator: string,
        next: number,
        offset: number,
        parentIndent: number,
        startOnNewline: boolean,
    ): Props {
        const result = new Props();
        let atNewline = startOnNewline;
        let hasSpace = startOnNewline;
        let commentSeparator = '';
        let requireSpace = false;
        let tab = -1;
        let startSet = false;
        for(const index of tokens) {
            const token = this.parser.get(index);
            if(requireSpace) {
                if(token.type !== 'space' && token.type !== 'newline' && token.type !== 'comma')
                    this.diagnostics.error(
                        token.offset,
                        'MISSING_CHAR',
                        'Tags and anchors must be separated from the next token by white space',
                    );
                requireSpace = false;
            }
            if(tab >= 0) {
                if(atNewline && token.type !== 'comment' && token.type !== 'newline')
                    this.diagnostics.tokenError(
                        this.parser.get(tab),
                        'TAB_AS_INDENT',
                        'Tabs are not allowed as indentation',
                    );
                tab = -1;
            }
            if(token.type === indicator) {
                if(result.anchor >= 0 || result.tag >= 0)
                    this.diagnostics.tokenError(
                        token,
                        'BAD_PROP_ORDER',
                        `Anchors and tags must be after the ${token.source} indicator`,
                    );
                if(result.found >= 0)
                    this.diagnostics.tokenError(
                        token,
                        'UNEXPECTED_TOKEN',
                        `Unexpected ${token.source} in ${flow === '' ? 'collection' : flow}`,
                    );
                result.found = index;
                atNewline = indicator === 'seq-item-ind' || indicator === 'explicit-key-ind';
                hasSpace = false;
                continue;
            }
            switch(token.type) {
                case 'space':
                    if(
                        flow === '' &&
                        (indicator !== 'doc-start' || next < 0 || this.parser.get(next).type !== 'flow-collection') &&
                        token.source.includes('\t')
                    )
                        tab = index;
                    hasSpace = true;
                    break;
                case 'comment': {
                    if(!hasSpace)
                        this.diagnostics.tokenError(
                            token,
                            'MISSING_CHAR',
                            'Comments must be separated from other tokens by white space characters',
                        );
                    let comment = token.source.slice(1);
                    if(comment === '') comment = ' ';
                    result.comment += result.comment === '' ? comment : commentSeparator + comment;
                    commentSeparator = '';
                    atNewline = false;
                    break;
                }
                case 'newline':
                    if(atNewline) {
                        if(result.comment !== '') result.comment += token.source;
                        else if(result.found < 0 || indicator !== 'seq-item-ind') result.spaceBefore = true;
                    }
                    else commentSeparator += token.source;
                    atNewline = true;
                    result.hasNewline = true;
                    if(result.anchor >= 0 || result.tag >= 0) result.newlineAfterProp = index;
                    hasSpace = true;
                    break;
                case 'anchor':
                case 'tag': {
                    const anchor = token.type === 'anchor';
                    if(anchor) {
                        if(result.anchor >= 0)
                            this.diagnostics.tokenError(
                                token,
                                'MULTIPLE_ANCHORS',
                                'A node can have at most one anchor',
                            );
                        if(token.source.endsWith(':')) {
                            const warningOffset = token.offset + token.source.length - 1;
                            result.warnings.push(
                                new ComposeError(
                                    warningOffset,
                                    warningOffset + 1,
                                    'BAD_ALIAS',
                                    'Anchor ending in : is ambiguous',
                                ),
                            );
                        }
                        result.anchor = index;
                    }
                    else {
                        if(result.tag >= 0)
                            this.diagnostics.tokenError(token, 'MULTIPLE_TAGS', 'A node can have at most one tag');
                        result.tag = index;
                    }
                    if(!startSet) {
                        result.start = token.offset;
                        startSet = true;
                    }
                    atNewline = false;
                    hasSpace = false;
                    requireSpace = true;
                    break;
                }
                case 'comma':
                    if(flow !== '') {
                        if(result.comma >= 0)
                            this.diagnostics.tokenError(token, 'UNEXPECTED_TOKEN', `Unexpected , in ${flow}`);
                        result.comma = index;
                        atNewline = false;
                        hasSpace = false;
                        break;
                    }
                    this.diagnostics.tokenError(token, 'UNEXPECTED_TOKEN', `Unexpected ${token.type} token`);
                    atNewline = false;
                    hasSpace = false;
                    break;
                default:
                    this.diagnostics.tokenError(token, 'UNEXPECTED_TOKEN', `Unexpected ${token.type} token`);
                    atNewline = false;
                    hasSpace = false;
            }
        }
        result.end = offset;
        if(tokens.length > 0) {
            const last = this.parser.get(tokens[tokens.length - 1] ?? -1);
            result.end = last.offset + last.source.length;
        }
        if(requireSpace && next >= 0) {
            const token = this.parser.get(next);
            if(
                token.type !== 'space' &&
                token.type !== 'newline' &&
                token.type !== 'comma' &&
                (token.type !== 'scalar' || token.source !== '')
            )
                this.diagnostics.error(
                    token.offset,
                    'MISSING_CHAR',
                    'Tags and anchors must be separated from the next token by white space',
                );
        }
        if(
            tab >= 0 &&
            ((atNewline && this.parser.get(tab).indent <= parentIndent) ||
                (next >= 0 &&
                    (this.parser.get(next).type === 'block-map' || this.parser.get(next).type === 'block-seq')))
        )
            this.diagnostics.tokenError(this.parser.get(tab), 'TAB_AS_INDENT', 'Tabs are not allowed as indentation');
        if(!startSet) result.start = result.end;
        return result;
    }
}
