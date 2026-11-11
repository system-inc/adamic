interface Type { flags: number }
function read(type: Type) {
    return "intrinsicName" in type ? type.intrinsicName : undefined;
}

const actual = { flags: 2, intrinsicName: "unknown" };
const result = read(actual);
console.log(typeof result === "string" ? result : "");
