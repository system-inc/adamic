// From TypeScript 6.0.3, src/compiler/utilities.ts:8280
// Census reason: type assertion sites
// Support types and driver are reduced scaffolding; the upstream function is unchanged.

type ObjectFlags = number;
const TypeFlags = { ObjectFlagsType: 403963917 } as const;
interface Type { readonly flags: number; }
interface ObjectFlagsType extends Type { readonly objectFlags: ObjectFlags; }

export function getObjectFlags(type: Type): ObjectFlags {
    return type.flags & TypeFlags.ObjectFlagsType ? (type as ObjectFlagsType).objectFlags : 0;
}

const object: ObjectFlagsType = { flags: 1048576, objectFlags: 2 };
console.log(String(getObjectFlags(object)));
console.log(String(getObjectFlags({ flags: 32 })));
