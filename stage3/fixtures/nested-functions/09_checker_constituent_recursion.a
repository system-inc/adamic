// a-check: refused the non-null assertion !
// From TypeScript 6.0.3, src/compiler/checker.ts:18832
// From TypeScript 6.0.3, src/compiler/checker.ts:18838
// Census reason: a function inside a function (a closure)

const TypeFlags = { Union: 134217728, Intersection: 268435456, UnionOrIntersection: 402653184 };
interface Type { flags: number; aliasSymbol: string | undefined; origin: Type | undefined; types: Type[]; }
type UnionType = Type;
type UnionOrIntersectionType = Type;
// Standalone adapter for the imported reduceLeft operation.
function reduceLeft(types: Type[], f: (n: number, t: Type) => number, initial: number): number { return types.reduce(f, initial); }
function createTypeChecker() {
    return getConstituentCount;
    function getConstituentCount(type: Type): number {
        return !(type.flags & TypeFlags.UnionOrIntersection) || type.aliasSymbol ? 1 :
            type.flags & TypeFlags.Union && (type as UnionType).origin ? getConstituentCount((type as UnionType).origin!) :
            getConstituentCountOfTypes((type as UnionOrIntersectionType).types);
    }
    function getConstituentCountOfTypes(types: Type[]): number {
        return reduceLeft(types, (n, t) => n + getConstituentCount(t), 0);
    }
}
const leaf: Type = { flags: 4, aliasSymbol: undefined, origin: undefined, types: [] };
const pair: Type = { flags: TypeFlags.Union, aliasSymbol: undefined, origin: undefined, types: [leaf, leaf] };
const intersection: Type = { flags: TypeFlags.Intersection, aliasSymbol: undefined, origin: undefined, types: [pair, leaf] };
const alias: Type = { flags: TypeFlags.Union, aliasSymbol: 'Alias', origin: undefined, types: [leaf, leaf] };
const original: Type = { flags: TypeFlags.Union, aliasSymbol: undefined, origin: pair, types: [leaf] };
const count = createTypeChecker();
console.log(`leaf ${count(leaf)}`);
console.log(`union ${count(pair)}`);
console.log(`intersection ${count(intersection)}`);
console.log(`alias ${count(alias)}`);
console.log(`origin ${count(original)}`);
