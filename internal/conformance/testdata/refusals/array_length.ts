// Spec: `length` can't be assigned.
// Expect: refused assigning a non-tuple array's length
const xs: number[] = [1,2]; xs.length = 1;
