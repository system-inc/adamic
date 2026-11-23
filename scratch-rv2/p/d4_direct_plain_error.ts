const error = new Error("m");
console.log(`${String(error.hasOwnProperty("message"))} ${String(error.hasOwnProperty("name"))} ${String(error.hasOwnProperty("stack"))}`);
