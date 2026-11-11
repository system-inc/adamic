interface Type { flags: number }
function read(type: Type) {
    return "intrinsicName" in type ? type.intrinsicName : undefined;
}
