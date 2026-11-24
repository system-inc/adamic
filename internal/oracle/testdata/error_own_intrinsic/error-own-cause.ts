console.log(String(Object.prototype.hasOwnProperty.call(new Error("x", { cause: 1 }), "cause")));
