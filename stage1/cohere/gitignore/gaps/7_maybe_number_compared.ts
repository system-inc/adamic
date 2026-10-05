// Gap 7, closed at 80c3098: number | undefined compared with a number.
const maybe: number | undefined = [1][0];
console.log(`${maybe === 1}`);
