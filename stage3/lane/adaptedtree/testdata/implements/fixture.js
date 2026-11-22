exports.A = class A { a = 1; };
exports.B = class B { b = 2; };

/**
 * @implements {exports.A}
 * @implements {exports.B}
 */
class Both {
    a = 1;
    b = 2;
}
exports.Both = Both;
