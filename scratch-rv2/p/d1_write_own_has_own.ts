const error = new Error("m");
error.hasOwnProperty = (_key: PropertyKey): boolean => false;
console.log(String(error.hasOwnProperty("message")));
