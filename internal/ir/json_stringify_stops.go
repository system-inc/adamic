package ir

// JSONErrorWriteStop is the whole stop shared by both stringify backends when an Error's
// name, or a message absent at construction, became own and enumerable (Adamic, Nov 13, #z4a9p9e).
const JSONErrorWriteStop = "JSON.stringify of an Error whose name, or a message it was made without, was written isn't built yet"
