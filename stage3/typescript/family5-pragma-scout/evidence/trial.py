from pathlib import Path
import sys

root = Path(sys.argv[1]) / "src/compiler"
parser = root / "parser.ts"
types = root / "types.ts"
s = types.read_text()
needle = 'export type PragmaPseudoMap = '
a = s.index(needle)
b = s.index('\n', a)
s = s[:a] + '''export type PragmaNamesWithKind<F extends PragmaKindFlags> = {
    [K in keyof ConcretePragmaSpecs]: ConcretePragmaSpecs[K]["kind"] extends F ? K : never;
}[keyof ConcretePragmaSpecs];

/** @internal */
export type ParsedPragmaArguments<K extends keyof ConcretePragmaSpecs> =
    K extends "ts-check" | "ts-nocheck" ? Record<never, never> :
    K extends "amd-dependency" ? { path: string | undefined; name?: string | undefined; } :
    K extends "amd-module" ? { name: string | undefined; } :
    K extends "reference" ? {
        [P in keyof PragmaArgumentType<K>]: P extends "types" | "lib" | "path"
            ? PragmaArgumentType<K>[P]
            : PragmaArgumentType<K>[P] | undefined;
    } :
    { [P in keyof PragmaArgumentType<K>]?: PragmaArgumentType<K>[P]; };

/** @internal */
export type PragmaPseudoMap = { [K in keyof ConcretePragmaSpecs]: { arguments: ParsedPragmaArguments<K>; range: CommentRange; }; };''' + s[b:]
types.write_text(s)
s = parser.read_text()
s = s.replace('    type PragmaPseudoMap,', '    type ParsedPragmaArguments,\n    type PragmaNamesWithKind,\n    type PragmaPseudoMap,')
a = s.index('function extractPragmas(')
b = s.index('/** @internal */', a)
part = s[a:b]
part = part.replace('as keyof PragmaPseudoMap;', 'as XmlPragmaName;', 1)
part = part.replace('as keyof PragmaPseudoMap;', 'as PositionalPragmaName;', 1)
import re
part, n = re.subn(r'const argument: \{ \[index: string\]: string \| \{ value: string; pos: number; end: number; \}; \} = \{\};', 'const argument: XmlPragmaBuilder = {};', part)
assert n == 1
part = part.replace('getNamedArgRegEx(arg.name)', 'getNamedArgRegEx(arg.name)')
part = part.replace('argument[arg.name]', 'argument[arg.name as keyof XmlPragmaBuilder]')
part = part.replace("as keyof XmlPragmaBuilder] = {", "as \"types\" | \"lib\" | \"path\"] = {").replace("as keyof XmlPragmaBuilder] = value", "as \"path\" | \"name\" | \"no-default-lib\" | \"resolution-mode\" | \"preserve\"] = value")
part = part.replace('arguments: argument, range', 'arguments: argument as ParsedPragmaArguments<XmlPragmaName>, range', 1)
part, n = re.subn(r'\{ \[index: string\]: string(?: \| undefined)?; \}', 'PositionalPragmaBuilder', part)
assert n == 2
part = part.replace('argMap[argument.name]', 'argMap[argument.name as keyof PositionalPragmaBuilder]')
part = part.replace('arguments: argument, range', 'arguments: argument as ParsedPragmaArguments<PositionalPragmaName>, range', 1)
aliases = '''type XmlPragmaName = PragmaNamesWithKind<PragmaKindFlags.TripleSlashXML>;
type PositionalPragmaName = PragmaNamesWithKind<PragmaKindFlags.SingleLine | PragmaKindFlags.MultiLine>;
type PragmaBuilder<T> = {
    [K in (T extends unknown ? keyof T : never)]?: (T extends unknown ? K extends keyof T ? T[K] : never : never) | undefined;
};
type XmlPragmaBuilder = PragmaBuilder<ParsedPragmaArguments<XmlPragmaName>>;
type PositionalPragmaBuilder = PragmaBuilder<ParsedPragmaArguments<PositionalPragmaName>>;

'''
parser.write_text(s[:a] + aliases + part + s[b:])
