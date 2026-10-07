interface TypeParser {
    readonly type: (minimum: number, conditional: boolean) => number;
}
class ParserLike {
    type(minimum = 0, conditional = true): number {
        return conditional ? minimum + 1 : minimum;
    }
}
function parse(parser: TypeParser): number {
    return parser.type(0, true);
}
console.log(parse(new ParserLike()).toString());
