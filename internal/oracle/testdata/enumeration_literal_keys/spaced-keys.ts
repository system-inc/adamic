const unrelated = { "#": 1 };
const keys = { "room 2": "north", "room 1": "south" };
console.log(Object.values({ "room 2": "north", "room 1": "south" }).join(","));
console.log(Object.keys(keys).join(","));
