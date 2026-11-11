interface Type { flags: number }
interface IntrinsicType extends Type { intrinsicName: string }
function read(type: Type): string | undefined {
    return type.flags & 402431 ? (type as IntrinsicType).intrinsicName : undefined;
}

const actual = { flags: 2, intrinsicName: "unknown" };
console.log(read(actual) ?? "");
