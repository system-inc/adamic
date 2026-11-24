package ir

// The two run-time NotYet stops of lower's checked dictionary view (cast_dictionary_view.go, #2v3xtgm), each a
// whole keyless message both backends say alike, in the shape cmd/adamic-admission-delta's runtimeNotYetKeyless
// lists (" isn't built yet").
const (
	// DictionaryViewRecordStop is a value cast from object to a dictionary of unknown values that isn't a record:
	// a fixed object, a class instance, an array, a function, a host value. Adamic's ruled phrase (#2v3xtgm,
	// 04:14Z).
	DictionaryViewRecordStop = "a cast to an index-signature type of a value not held as a record isn't built yet"
	// DictionaryViewReferencesStop is a record whose values aren't held as references (a Record<string, number>
	// or a Record<string, boolean> seen as object), which no read as a Union takes. Its words wait on Adamic's,
	// so a re-word is this one line.
	DictionaryViewReferencesStop = "a cast to an index-signature type of a record whose values aren't held as references isn't built yet"
)
