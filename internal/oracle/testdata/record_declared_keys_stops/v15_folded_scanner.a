const SyntaxKind = { AbstractKeyword: 128, AccessorKeyword: 129, ConstructorKeyword: 137 } as const;
type KeywordSyntaxKind = (typeof SyntaxKind)[keyof typeof SyntaxKind];
interface MapLike<T> { [index: string]: T; }
const kinds: MapLike<KeywordSyntaxKind> = {
    abstract: SyntaxKind.AbstractKeyword,
    accessor: SyntaxKind.AccessorKeyword,
    ["" + "constructor"]: SyntaxKind.ConstructorKeyword,
};
console.log(`${String(kinds["abstract"])} ${Object.keys(kinds).join(",")}`);
