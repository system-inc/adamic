export class UnistNode {
    parent = -1;
    children: number[] = [];
    value = '';
    tag = -1;
    anchor = -1;
    middleComments: number[] = [];
    leadingComments: number[] = [];
    trailingComment = -1;
    endComments: number[] = [];
    chomping = '';
    indent = -1;
    indicatorComment = -1;
    directivesEndMarker = false;
    documentEndMarker = false;
    name = '';
    parameters: string[] = [];
    comments: number[] = [];
    type: string;
    position: number;
    constructor(type: string, position: number) {
        this.type = type;
        this.position = position;
    }
}

export function hasLeadingCommentsField(type: string): boolean {
    return [
        'alias',
        'blockFolded',
        'blockLiteral',
        'directive',
        'flowMapping',
        'flowSequence',
        'flowMappingItem',
        'mappingItem',
        'mappingValue',
        'mapping',
        'plain',
        'quoteDouble',
        'quoteSingle',
        'sequenceItem',
        'sequence',
    ].includes(type);
}
export function hasTrailingCommentField(type: string): boolean {
    return [
        'alias',
        'directive',
        'document',
        'documentHead',
        'flowMapping',
        'flowSequence',
        'mappingKey',
        'mappingValue',
        'plain',
        'quoteDouble',
        'quoteSingle',
        'sequenceItem',
    ].includes(type);
}
export function hasChildrenField(type: string): boolean {
    return [
        'root',
        'document',
        'documentHead',
        'documentBody',
        'flowMapping',
        'flowSequence',
        'flowMappingItem',
        'flowSequenceItem',
        'mappingItem',
        'mappingKey',
        'mappingValue',
        'mapping',
        'sequenceItem',
        'sequence',
    ].includes(type);
}
