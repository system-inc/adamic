// Gap 1: indexOf with a position to search from is refused; without one it lowers.
const text = 'a"b"c';
console.log(`${text.indexOf('"', 2)}`);
