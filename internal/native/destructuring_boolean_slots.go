package native

// Optional booleans occupy a non-reference field slot as three exact numbers.
// Function arguments keep their ordinary present/value struct; the codec takes
// that struct once so a producing call can never be evaluated twice.
const destructuringBooleanSlots = `
static double adamic_destructuring_boolean_pack(adamic_maybe_boolean value) {
    return !value.present ? 2.0 : value.boolean ? 1.0 : 0.0;
}
static adamic_maybe_boolean adamic_destructuring_boolean_unpack(double packed) {
    return (adamic_maybe_boolean){packed != 2.0, packed == 1.0};
}

`
