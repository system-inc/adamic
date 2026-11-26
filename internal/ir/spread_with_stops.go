package ir

// SpreadWithSourceStop is the whole run-time NotYet both backends say when a SpreadWith literal's source is
// stored as anything but a plain object or a class instance at run time, an array, a tuple, a typed array, a Map,
// a Set, a closure or a host value among them, whose own keys Node would copy (Adamic, Nov 26, #k7g6f9h, under
// #7fg2fp3's form; #n0744ek lifts it). It's a positive test on both backends, so a storage added later stops.
const SpreadWithSourceStop = "a spread of a value not stored as a plain object into an object literal isn't built yet (#k7g6f9h)"
