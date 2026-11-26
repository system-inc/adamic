package ir

// SpreadWithSourceStop is the whole run-time NotYet both backends say when a SpreadWith literal's source,
// seen through an object type, is stored as an array, a tuple or a host value at run time, where Node would
// copy its indices or its own keys (Adamic, Nov 26, #k7g6f9h, under #7fg2fp3's form; #n0744ek lifts it).
const SpreadWithSourceStop = "a spread of a value stored as an array or a host value into an object literal isn't built yet (#k7g6f9h)"
