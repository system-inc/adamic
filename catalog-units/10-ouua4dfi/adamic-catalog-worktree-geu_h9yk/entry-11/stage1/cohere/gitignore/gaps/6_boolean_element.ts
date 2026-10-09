// Gap 6, closed at 80c3098: an element of a boolean[], which is boolean | undefined.
const flags: boolean[] = [true, false];
const first = flags[0];
console.log(`${flags[0] === true} ${first === true} ${flags[2] === undefined}`);
