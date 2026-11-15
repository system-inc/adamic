const pick = (...values: number[]): number => values.length;
const values = [5, 3, 9, 1];
values.sort((left: number, right: number) => left - right);
console.log(values.join(",") + " " + String(pick(1, 2)));
