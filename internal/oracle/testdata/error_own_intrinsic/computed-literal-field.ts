const unrelated = { ["k"]: 0 };
const error = new Error("message");
console.log(String(error.hasOwnProperty("message")));
console.log(String(error.hasOwnProperty("name")));
console.log(String(error.hasOwnProperty("stack")));
