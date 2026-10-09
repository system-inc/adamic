package policy

import (
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"
)

// TestSwiftMessagesTypeTheValuesAndTheOptions: a Swift message becomes a function whose parameters are its
// text's values and phrases in the order the sentence reads them, its terms in their Swift words, its
// string phrase in place, and its object phrase an enum whose cases carry their own values.
func TestSwiftMessagesTypeTheValuesAndTheOptions(t *testing.T) {
	t.Parallel()
	catalog, err := LoadMessages(fstest.MapFS{"consistency-no-thing.json": {Data: []byte(`{
  "rules": { "TypeScript": "base/consistency-no-thing", "Swift": "cohere-swift/consistency-no-thing" },
  "phrases": {
    "reasoning": "A thing hides what it is.",
    "exitCount": { "one": "", "several": ", the first of {{count}}" }
  },
  "messages": {
    "thing": {
      "text": "This is [[subject]] named {{name}}<<exitCount>>. <<reasoning>>",
      "terms": { "subject": { "TypeScript": "a {{kind}}", "Swift": "an enum" } }
    },
    "otherThing": { "text": "Another is {{name}}. <<reasoning>>" }
  }
}`)}})
	if err != nil {
		t.Fatal(err)
	}
	source, err := swiftMessagesSource(catalog, "digest")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`static let sourceDigest = "digest"`,
		`"ConsistencyNoThing.otherThing",` + "\n" + `        "ConsistencyNoThing.thing",`,
		"    /* cohere-swift/consistency-no-thing */\n    enum ConsistencyNoThing {",
		`static func otherThing(name: String) -> Message {` + "\n" + `            Message(` + "\n" +
			`                id: "otherThing",` + "\n" + `                text:` + "\n" +
			`                    #"Another is \#(name). A thing hides what it is."#,` + "\n" + `            )`,
		"        enum ExitCount {\n            case one\n            case several(count: String)\n        }",
		`static func thing(name: String, exitCount: ExitCount) -> Message {`,
		"case .one:\n                        #\"\"#",
		"case .several(let count):\n                        #\", the first of \\#(count)\"#",
		"return Message(\n                id: \"thing\",\n                text:\n" +
			`                    #"This is an enum named \#(name)\#(exitCountText). A thing hides what it is."#,`,
	} {
		if !strings.Contains(string(source), want) {
			t.Errorf("the source does not hold %q:\n%s", want, source)
		}
	}
	if strings.Contains(string(source), "{{kind}}") || strings.Contains(string(source), "kind") {
		t.Errorf("the TypeScript term's value reached Swift:\n%s", source)
	}
}

// TestSwiftMessagesRefuseAValueWithTwoPlaces: a value in the text and in an option of the same message
// would be two parameters for one name, so generation stops and names it.
func TestSwiftMessagesRefuseAValueWithTwoPlaces(t *testing.T) {
	t.Parallel()
	catalog, err := LoadMessages(fstest.MapFS{"consistency-no-thing.json": {Data: []byte(`{
  "rules": { "Swift": "cohere-swift/consistency-no-thing" },
  "phrases": { "exitCount": { "one": "", "several": " of {{count}}" } },
  "messages": { "thing": { "text": "{{count}} things<<exitCount>>." } }
}`)}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := swiftMessagesSource(catalog, "digest"); err == nil || !strings.Contains(err.Error(), `"count"`) {
		t.Errorf("a value with two places was generated (%v)", err)
	}
}

// TestSwiftLiteralsOutrunTheText: a `"#` in the text stays text, a line break is the raw string's escape,
// and a value interpolates through the same delimiter.
func TestSwiftLiteralsOutrunTheText(t *testing.T) {
	t.Parallel()
	literal, err := swiftInterpolatedLiteral("a \"# b\n{{name}}")
	if err != nil {
		t.Fatal(err)
	}
	if want := `##"a "# b\##n\##(name)"##`; literal != want {
		t.Errorf("got %s, expected %s", literal, want)
	}
}

// TestOneEditReachesBothEngines: one edited entry shows in the Go rule's rendered message and in the
// Swift source generated from the same files. The Swift test theFindingIsTheCatalogsMessage closes the
// chain from that source to the finding a user sees.
func TestOneEditReachesBothEngines(t *testing.T) {
	t.Parallel()
	edited := fstest.MapFS{}
	err := fs.WalkDir(MessageFiles(), ".", func(name string, entry fs.DirEntry, walkError error) error {
		if walkError != nil || entry.IsDir() {
			return walkError
		}
		data, err := fs.ReadFile(MessageFiles(), name)
		edited[name] = &fstest.MapFile{Data: data}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	const original = "which names no declared failure."
	const changed = "which names no failure anyone declared."
	file := edited["consistency-no-bare-throw.json"]
	if file == nil || !strings.Contains(string(file.Data), original) {
		t.Fatalf("bare-throw's shared clause %q is not in its file, so the edit would not apply", original)
	}
	file.Data = []byte(strings.Replace(string(file.Data), original, changed, 1))
	catalog, err := LoadMessages(edited)
	if err != nil {
		t.Fatal(err)
	}

	goText := catalog.Render(MessageHandle{Rule: "base/consistency-no-bare-throw", Id: "bareThrow"}, map[string]string{"constructor": "Error"})
	if !strings.Contains(goText, changed) {
		t.Errorf("the Go rule renders %q", goText)
	}

	source, err := swiftMessagesSource(catalog, "digest")
	if err != nil {
		t.Fatal(err)
	}
	function := string(source)[strings.Index(string(source), "static func bareThrow("):]
	if !strings.Contains(function[:strings.Index(function, "\n        }")], changed) {
		t.Errorf("the Swift function does not carry the edit:\n%s", source)
	}
}

// TestALongSignatureTakesOneParameterToALine: past the house format's 120 columns, the signature breaks
// as the format breaks it, so the formatter leaves the generated file as written.
func TestALongSignatureTakesOneParameterToALine(t *testing.T) {
	t.Parallel()
	catalog, err := LoadMessages(fstest.MapFS{"consistency-no-thing.json": {Data: []byte(`{
  "rules": { "Swift": "cohere-swift/consistency-no-thing" },
  "messages": {
    "thing": { "text": "{{target}} lacks {{features}}; add {{settings}} because {{reasons}}, {{explanation}}." },
    "short": { "text": "{{target}} is short." }
  }
}`)}})
	if err != nil {
		t.Fatal(err)
	}
	source, err := swiftMessagesSource(catalog, "digest")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"        static func thing(\n            target: String,\n            features: String,\n            settings: String,\n            reasons: String,\n            explanation: String,\n        ) -> Message {",
		"        static func short(target: String) -> Message {",
	} {
		if !strings.Contains(string(source), want) {
			t.Errorf("the source does not hold %q:\n%s", want, source)
		}
	}
}

// TestATwinWithNoSwiftMessageGetsNoType: a twin file whose messages are all TypeScript's names the Swift
// rule, and generates nothing for it rather than an empty enum.
func TestATwinWithNoSwiftMessageGetsNoType(t *testing.T) {
	t.Parallel()
	catalog, err := LoadMessages(fstest.MapFS{"consistency-no-thing.json": {Data: []byte(`{
  "rules": { "TypeScript": "nexus/consistency-no-thing", "Swift": "cohere-swift/consistency-no-thing" },
  "messages": { "thing": { "text": "A thing.", "languages": ["TypeScript"] } }
}`)}})
	if err != nil {
		t.Fatal(err)
	}
	source, err := swiftMessagesSource(catalog, "digest")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(source), "ConsistencyNoThing") {
		t.Errorf("a Swift rule with no Swift message got a type:\n%s", source)
	}
}
