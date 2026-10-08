package naming

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
)

type fixture struct {
	Declaration Declaration `json:"declaration"`
	Swift       string      `json:"swift"`
	Adamic      string      `json:"adamic"`
	Module      string      `json:"module"`
	Source      string      `json:"source"`
}

func fixtures(t *testing.T) []fixture {
	t.Helper()
	data, err := os.ReadFile("testdata/declarations.json")
	if err != nil {
		t.Fatal(err)
	}
	var result []fixture
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result); err != nil {
		t.Fatal(err)
	}
	return result
}
func TestDocumentationFixtures(t *testing.T) {
	t.Parallel()
	rows := fixtures(t)
	if len(rows) < 150 {
		t.Fatalf("only %d documentation fixtures", len(rows))
	}
	frameworks := map[string]int{}
	for _, f := range rows {
		frameworks[f.Declaration.Framework]++
		t.Run(Identity(f.Declaration), func(t *testing.T) {
			t.Parallel()
			if !strings.HasPrefix(f.Source, "https://developer.apple.com/documentation/") {
				t.Fatal("missing Apple documentation provenance")
			}
			got, err := Name(f.Declaration)
			if err != nil {
				t.Fatal(err)
			}
			if got.SwiftName != f.Swift {
				t.Errorf("Swift got %q want %q (oracle %s)", got.SwiftName, f.Swift, f.Source)
			}
			if got.Name != f.Adamic {
				t.Errorf("Adamic got %q want %q", got.Name, f.Adamic)
			}
			if got.Module != f.Module {
				t.Errorf("module got %q want %q", got.Module, f.Module)
			}
			if got.Original != f.Declaration.Name {
				t.Error("original spelling was lost")
			}
		})
	}
	for _, f := range []string{"AppKit", "Foundation", "UIKit", "CoreGraphics"} {
		if frameworks[f] < 10 {
			t.Errorf("too few %s fixtures: %d", f, frameworks[f])
		}
	}
}
func TestFixtureBijection(t *testing.T) {
	t.Parallel()
	var declarations []Declaration
	for _, f := range fixtures(t) {
		declarations = append(declarations, f.Declaration)
	}
	surface, err := Build(declarations)
	if err != nil {
		t.Fatal(err)
	}
	if len(surface.Forward) != len(declarations) || len(surface.Reverse) != len(declarations) {
		t.Fatal("surface is not bijective")
	}
	for _, d := range declarations {
		id := Identity(d)
		out := surface.Forward[id]
		if surface.Reverse[ExportKey(d, out)] != id {
			t.Fatalf("cannot reverse %s", id)
		}
	}
	for i, j := 0, len(declarations)-1; i < j; i, j = i+1, j-1 {
		declarations[i], declarations[j] = declarations[j], declarations[i]
	}
	reversed, err := Build(declarations)
	if err != nil || !reflect.DeepEqual(surface, reversed) {
		t.Fatal("input order changed the surface", err)
	}
}
func TestCollisionsAreRejected(t *testing.T) {
	t.Parallel()
	cases := map[string][]Declaration{
		"two enum prefixes": {
			{Kind: OptionConstant, Name: "NSFirstStatusReady", Parent: "NSFirstStatus", ParentSwiftName: "NSWindow.Status", Framework: "AppKit", Enumerators: []Enumerator{{Name: "NSFirstStatusReady"}, {Name: "NSFirstStatusWaiting"}}},
			{Kind: OptionConstant, Name: "NSSecondStatusReady", Parent: "NSSecondStatus", ParentSwiftName: "NSWindow.Status", Framework: "AppKit", Enumerators: []Enumerator{{Name: "NSSecondStatusReady"}, {Name: "NSSecondStatusWaiting"}}},
		},
		"property and method": {
			{Kind: Property, Name: "reload", Parent: "NSWindow", Framework: "AppKit"},
			{Kind: InstanceMethod, Name: "reload", Parent: "NSWindow", Framework: "AppKit"},
		},
		"overloads of one shape": {
			{Kind: InstanceMethod, Name: "showText:", Parent: "NSWindow", Framework: "AppKit", Parameters: []Parameter{{Name: "text", Type: Type{Spelling: "NSString *", Nullability: Nonnull}}}, SwiftName: "show(_:)"},
			{Kind: InstanceMethod, Name: "showMessage:", Parent: "NSWindow", Framework: "AppKit", Parameters: []Parameter{{Name: "message", Type: Type{Spelling: "NSString *", Nullability: Nonnull}}}, SwiftName: "show(_:)"},
		},
		"normalization": {
			{Kind: Property, Name: "contentRect", Parent: "NSWindow", Framework: "AppKit"},
			{Kind: Property, Name: "contentRectangle", Parent: "NSWindow", Framework: "AppKit"},
		},
		"module export": {
			{Kind: Class, Name: "NSURLSession", Framework: "Foundation"},
			{Kind: Class, Name: "URLSession", Framework: "Foundation"},
		},
		"duplicate original": {
			{Kind: Class, Name: "NSWindow", Framework: "AppKit"},
			{Kind: Class, Name: "NSWindow", Framework: "AppKit"},
		},
	}
	for name, ds := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := Build(ds); err == nil {
				t.Fatal("collision accepted")
			}
			ds[0], ds[1] = ds[1], ds[0]
			if _, err := Build(ds); err == nil {
				t.Fatal("reversed collision accepted")
			}
		})
	}
	// A class and its instances are two namespaces, and overloads of different shapes are told apart
	// by a call's arguments.
	if _, err := Build([]Declaration{
		{Kind: ClassMethod, Name: "reload", Parent: "NSWindow", Framework: "AppKit"},
		{Kind: InstanceMethod, Name: "reload", Parent: "NSWindow", Framework: "AppKit"},
		{Kind: InstanceMethod, Name: "showText:", Parent: "NSWindow", Framework: "AppKit", Parameters: []Parameter{{Name: "text", Type: Type{Spelling: "NSString *", Nullability: Nonnull}}}, SwiftName: "show(_:)"},
		{Kind: InstanceMethod, Name: "showCount:", Parent: "NSWindow", Framework: "AppKit", Parameters: []Parameter{{Name: "count", Type: Type{Spelling: "NSInteger", Nullability: Nonnull}}}, SwiftName: "show(_:)"},
	}); err != nil {
		t.Fatal(err)
	}
	// Identical words in different frameworks and different owning types are safe.
	_, err := Build([]Declaration{{Kind: Class, Name: "NSView", Framework: "AppKit"}, {Kind: Class, Name: "UIView", Framework: "UIKit"}, {Kind: Property, Name: "title", Parent: "NSWindow", Framework: "AppKit"}, {Kind: Property, Name: "title", Parent: "NSMenu", Framework: "AppKit"}})
	if err != nil {
		t.Fatal(err)
	}
}
func TestArgumentLayout(t *testing.T) {
	t.Parallel()
	for _, f := range fixtures(t) {
		d := f.Declaration
		if d.Name != "setFrame:display:animate:" && d.Name != "dataTaskWithURL:completionHandler:" && d.Name != "dismissViewControllerAnimated:completion:" {
			continue
		}
		out, err := Name(d)
		if err != nil {
			t.Fatal(err)
		}
		var positional []int
		var options []string
		for _, a := range out.Arguments.Positional {
			positional = append(positional, a.Index)
		}
		for _, a := range out.Arguments.Options {
			options = append(options, a.Name)
		}
		wantPos := []int(nil)
		wantOptions := []string{"with", "completionHandler"}
		if d.Name == "setFrame:display:animate:" {
			wantPos = []int{0}
			wantOptions = []string{"display", "animate"}
		}
		if d.Name == "dismissViewControllerAnimated:completion:" {
			wantOptions = []string{"animated", "completion"}
		}
		if !reflect.DeepEqual(positional, wantPos) || !reflect.DeepEqual(options, wantOptions) {
			t.Fatalf("%s layout %v %v", d.Name, positional, options)
		}
		all := append(append([]Argument{}, out.Arguments.Positional...), out.Arguments.Options...)
		for i, a := range all {
			if a.Index != i {
				t.Fatal("marshalling reordered arguments")
			}
		}
	}
	constructor := Declaration{Kind: InstanceMethod, Name: "initWithContentRect:styleMask:backing:defer:", Parent: "NSWindow", Framework: "AppKit", Parameters: []Parameter{
		{Name: "contentRect", Type: Type{Spelling: "NSRect"}},
		{Name: "styleMask", Type: Type{Spelling: "NSWindowStyleMask", OptionCases: []string{"Titled", "Closable"}}},
		{Name: "backing", Type: Type{Spelling: "NSBackingStoreType"}},
		{Name: "defer", Type: Type{Spelling: "BOOL"}},
	}, Result: Type{Spelling: "instancetype", Nullability: Nonnull}}
	out, err := Name(constructor)
	if err != nil {
		t.Fatal(err)
	}
	if out.Name != "constructor" || out.SwiftName != "NSWindow.init(contentRect:styleMask:backing:defer:)" || !out.Arguments.Constructor || len(out.Arguments.Positional) != 0 {
		t.Fatalf("constructor: %#v", out)
	}
	if out.Arguments.Options[0].Name != "contentRectangle" || out.Arguments.Options[1].Type != "readonly ('Titled' | 'Closable')[]" || out.Type != "Window" {
		t.Fatalf("constructor: %#v", out)
	}
	explicit := Declaration{Kind: InstanceMethod, Name: "consumeURL:options:", Parent: "NSWindow", Framework: "AppKit", SwiftName: "consume(_:options:)", Parameters: []Parameter{{Name: "URL", Type: Type{Spelling: "NSURL *", Nullability: Nonnull, Object: true}}, {Name: "options", Type: Type{Spelling: "NSUInteger"}}}}
	out, err = Name(explicit)
	if err != nil {
		t.Fatal(err)
	}
	if out.Name != "consume" || out.Arguments.Positional[0].Name != "url" || out.Arguments.Options[0].Name != "options" {
		t.Fatal(out)
	}
}
func TestTypeMapping(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		input Type
		want  string
	}{
		{Type{Spelling: "id"}, "unknown | null"}, {Type{Spelling: "CGFloat"}, "number"}, {Type{Spelling: "NSInteger"}, "number"}, {Type{Spelling: "NSUInteger"}, "number"}, {Type{Spelling: "double"}, "number"}, {Type{Spelling: "float"}, "number"}, {Type{Spelling: "BOOL"}, "boolean"},
		{Type{Spelling: "NSString *"}, "string | null"}, {Type{Spelling: "NSString *", Nullability: Nullable}, "string | null"}, {Type{Spelling: "NSString *", Nullability: Nonnull}, "string"},
		{Type{Spelling: "NSURL *", Nullability: Nonnull, Object: true}, "Url"}, {Type{Spelling: "NSHTTPURLResponse *", Object: true}, "HttpUrlResponse | null"},
		{Type{Spelling: "CGRect"}, "Rectangle"}, {Type{Spelling: "CGColorRef", Reference: true}, "ColorRef | null"},
		{Type{Spelling: "void (^)(void)", Function: &FunctionType{Result: Type{Spelling: "void"}}}, "(() => void) | null"},
		{Type{Spelling: "NSWindowStyleMask", OptionCases: []string{"Titled", "Closable"}}, "readonly ('Titled' | 'Closable')[]"},
	} {
		got, err := MapType(tt.input)
		if err != nil || got != tt.want {
			t.Errorf("%#v = %q, %v want %q", tt.input, got, err, tt.want)
		}
	}
	for _, bad := range []Type{{Spelling: "char **"}, {Spelling: "CGFloat *"}, {Spelling: "NSRect *"}, {Spelling: "void *"}, {Spelling: "BOOL", Nullability: "maybe"}, {Spelling: "Mask", OptionCases: []string{"URL", "Url"}}} {
		if _, err := MapType(bad); err == nil {
			t.Errorf("unsupported type accepted %#v", bad)
		}
	}
}
func TestNamesAndWordBoundaries(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		input, want string
		pascal      bool
	}{
		{"NSWindow", "Window", true}, {"URLSession", "UrlSession", true}, {"JSONDecoder", "JsonDecoder", true}, {"NSURL", "Url", true}, {"HTTPURLResponse", "HttpUrlResponse", true},
		{"identifier", "identifier", false}, {"taskID", "taskIdentifier", false}, {"contentRect", "contentRectangle", false}, {"max", "max", false}, {"min", "min", false}, {"index", "index", false},
		{"CAAnimation", "Animation", true}, {"AVPlayer", "Player", true}, {"CFRunLoop", "RunLoop", true}, {"URLIsValid", "urlIsValid", false}, {"NSNotFound", "NotFound", true},
	} {
		if got := normalize(tt.input, tt.pascal); got != tt.want {
			t.Errorf("%s = %s want %s", tt.input, got, tt.want)
		}
	}
	if got := commonWordPrefix("NSFooBar", "NSFooBas"); got != "NSFoo" {
		t.Fatal(got)
	}
	if got := enumPrefix("NSStatus", []Enumerator{{Name: "NSStatusReady"}, {Name: "NSStatusReallyReady"}}); got != "NSStatus" {
		t.Fatal(got)
	}
	if got := enumPrefix("NSStatus", []Enumerator{{Name: "NSStatusReady"}}); got != "NSStatus" {
		t.Fatal("singleton ate case", got)
	}
}
func TestErrorsAndMetadata(t *testing.T) {
	t.Parallel()
	good := Declaration{Kind: Property, Name: "hidden", Parent: "UIView", Framework: "UIKit", Getter: "isHidden", Setter: "setHidden:", Result: Type{Spelling: "BOOL"}}
	out, err := Name(good)
	if err != nil || out.Name != "isHidden" || out.Getter != "isHidden" || out.Setter != "setHidden:" {
		t.Fatal(out, err)
	}
	refined := Declaration{Kind: InstanceMethod, Name: "readValue", Parent: "NSWindow", Framework: "AppKit", RefinedForSwift: true}
	out, err = Name(refined)
	if err != nil || !out.RefinedForSwift || out.SwiftName != "NSWindow.__readValue()" || out.Name != "readValue" {
		t.Fatal(out, err)
	}
	for _, bad := range []Declaration{
		{Kind: EnumConstant, Name: "NSStatusUnknown", Parent: "NSStatus", Framework: "Foundation", SwiftName: "_"},
		{Kind: "oops", Name: "x", Framework: "AppKit"}, {Kind: Class, Name: "NSWindow", Framework: "../appkit"},
		{Kind: InstanceMethod, Name: "setFrame:", Parent: "NSWindow", Framework: "AppKit"},
		{Kind: InstanceMethod, Name: ":", Parent: "NSWindow", Framework: "AppKit", Parameters: []Parameter{{Type: Type{Spelling: "BOOL"}}}},
		{Kind: EnumConstant, Name: "NSFooReady", Parent: "NSFoo", Framework: "Foundation"},
		{Kind: InstanceMethod, Name: "x:", Parent: "NSWindow", Framework: "AppKit", SwiftName: "x(_:_:)", Parameters: []Parameter{{Type: Type{Spelling: "BOOL"}}}},
		{Kind: InstanceMethod, Name: "x:y:", Parent: "NSWindow", Framework: "AppKit", SwiftName: "x(_:_:)", Parameters: []Parameter{{Type: Type{Spelling: "BOOL"}}, {Type: Type{Spelling: "BOOL"}}}},
		{Kind: Property, Name: "title", Framework: "AppKit"},
	} {
		if _, err := Name(bad); err == nil {
			t.Errorf("bad declaration accepted %#v", bad)
		}
	}
}

func TestPortSafeguards(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, typ, role, want string
		properties            []string
	}{
		{"setObject", "NSObject", "base", "setObject", nil},
		{"addOperation", "NSOperation", "base", "addOperation", []string{"operations"}},
		{"insertObject", "NSObject", "base", "insert", nil},
		{"atIndex", "Int", "subsequent", "at", nil},
		{"Object", "NSObject", "first", "", nil},
		{"Object", "NSObject", "subsequent", "Object", nil},
		{"withError", "NSError", "subsequent", "withError", nil},
		{"withIndices", "NSIndexSet", "subsequent", "with", nil},
		{"usingObjectValue", "AnyObject", "subsequent", "using", nil},
	} {
		if got := trailing(tt.name, omissionType{name: tt.typ}, tt.role, tt.properties); got != tt.want {
			t.Errorf("trailing %s: %s want %s", tt.name, got, tt.want)
		}
	}
	for _, tt := range []struct {
		name                   string
		typ                    omissionType
		parameter, base, label string
	}{
		{"dismissAnimated", omissionType{name: "Bool", boolean: true}, "animated", "dismiss", "animated"},
		{"setObjectForKey", omissionType{name: "AnyObject"}, "object", "setObjectForKey", ""},
		{"makeKeyAndOrderFront", omissionType{name: "AnyObject"}, "sender", "makeKeyAndOrderFront", ""},
		{"moveToX", omissionType{name: "CGFloat"}, "x", "moveTo", "X"},
		{"dataTaskWithURL", omissionType{name: "NSURL"}, "url", "dataTask", "WithURL"},
		{"performWithBlock", omissionType{name: "Block", function: true}, "block", "perform", "Block"},
		{"copyWithZone", omissionType{name: "NSZone"}, "zone", "copy", "WithZone"},
		{"plugIn", omissionType{name: "NSObject"}, "object", "plugIn", ""},
		{"convertToBacking", omissionType{name: "NSRect"}, "rect", "convertToBacking", ""},
	} {
		base, label := splitBase(tt.name, tt.typ, tt.parameter)
		if base != tt.base || label != tt.label {
			t.Errorf("split %s: %s %s want %s %s", tt.name, base, label, tt.base, tt.label)
		}
	}
	if got := trailing("withViews", omissionType{name: "NSArray", element: "NSView"}, "subsequent", nil); got != "with" {
		t.Fatal("collection element omission", got)
	}
	if got := commonPluralPrefix("NSCategory", "NSCategories"); got != "NSCategory" {
		t.Fatal("plural prefix", got)
	}
	if got := enumPrefix("CGLineCap", []Enumerator{{Name: "kCGLineCapButt"}, {Name: "kCGLineCapRound"}, {Name: "AnOldName", Deprecated: true}, {Name: "Custom", SwiftName: "custom"}}); got != "kCGLineCap" {
		t.Fatal("available enum prefix", got)
	}
	if got := enumPrefix("CGLineCap", []Enumerator{{Name: "CGLineCap_Butt"}, {Name: "CGLineCap_Round"}}); got != "CGLineCap_" {
		t.Fatal("underscore prefix", got)
	}
	d := Declaration{Kind: Category, Name: "Drawing", Parent: "NSView", Framework: "AppKit"}
	out, err := Name(d)
	if err != nil || out.Name != "Drawing" || out.Module != "apple/appkit/view" {
		t.Fatal(out, err)
	}
	d = Declaration{Kind: InstanceMethod, Name: "title", Parent: "NSWindow", Framework: "AppKit", AccessorProperty: "title"}
	if _, err := Build([]Declaration{d}); err == nil {
		t.Fatal("raw accessor emitted as independent member")
	}
	d = Declaration{Kind: ClassMethod, Name: "colorWithRed:green:blue:alpha:", Parent: "UIColor", Framework: "UIKit", Result: Type{Spelling: "UIColor *", Object: true, Nullability: Nonnull}, Parameters: []Parameter{{Type: Type{Spelling: "CGFloat"}}, {Type: Type{Spelling: "CGFloat"}}, {Type: Type{Spelling: "CGFloat"}}, {Type: Type{Spelling: "CGFloat"}}}}
	out, err = Name(d)
	if err != nil || out.SwiftName != "UIColor.init(red:green:blue:alpha:)" || !out.Arguments.Constructor {
		t.Fatal(out, err)
	}
	d = Declaration{Kind: Class, Name: "NSGeometry", Framework: "AppKit", SwiftName: "NSWindow.Geometry"}
	out, err = Name(d)
	if err != nil || out.Module != "apple/appkit/window" || out.Name != "Geometry" {
		t.Fatal(out, err)
	}
}

func TestGlobalNameBijection(t *testing.T) {
	t.Parallel()
	pairs := map[string]string{"NSError": "FoundationError", "NSDate": "FoundationDate", "NSString": "FoundationString", "NSNumber": "FoundationNumber", "NSArray": "FoundationArray", "NSSet": "FoundationSet", "NSDictionary": "FoundationDictionary", "NSObject": "FoundationObject", "NSProxy": "FoundationProxy"}
	var declarations []Declaration
	for original, want := range pairs {
		d := Declaration{Kind: Class, Name: original, Framework: "Foundation"}
		o, err := Name(d)
		if err != nil || o.Name != want || !strings.HasSuffix(o.Module, "/"+kebab(want)) {
			t.Fatalf("%s: %#v %v", original, o, err)
		}
		declarations = append(declarations, d)
	}
	surface, err := Build(declarations)
	if err != nil || len(surface.Reverse) != len(pairs) {
		t.Fatal("global surface is not bijective", err)
	}
	for _, d := range declarations {
		o := surface.Forward[Identity(d)]
		if surface.Reverse[ExportKey(d, o)] != Identity(d) {
			t.Fatal("global reverse lookup lost", d)
		}
	}
	if _, err := Build([]Declaration{{Kind: Class, Name: "NSError", Framework: "Foundation"}, {Kind: Class, Name: "FoundationError", Framework: "Foundation"}}); err == nil {
		t.Fatal("prefixed global collision accepted")
	}
	data, err := os.ReadFile("testdata/es2024-globals.json")
	if err != nil {
		t.Fatal(err)
	}
	var globals []string
	if err := json.Unmarshal(data, &globals); err != nil {
		t.Fatal(err)
	}
	if len(globals) != 194 {
		t.Fatal("global audit incomplete")
	}
	for _, global := range globals {
		if got := typeName(global, "Foundation"); got != "Foundation"+normalize(global, true) {
			t.Errorf("unreserved global %s: %s", global, got)
		}
	}
	got, err := MapType(Type{Spelling: "NSError *", Object: true, Nullability: Nonnull, Framework: "Foundation"})
	if err != nil || got != "FoundationError" {
		t.Fatal(got, err)
	}
}

func TestES2024GlobalAudit(t *testing.T) {
	t.Parallel()
	libraries := filepath.Join("..", "..", "..", "cohere", "TypeScript", "tsc", "internal", "bundled", "libs")
	reference := regexp.MustCompile(`<reference lib="([^"]+)"`)
	declaration := regexp.MustCompile(`(?m)^(?:declare\s+)?(?:var|const|function|class|interface|type|namespace)\s+([A-Za-z_$][\w$]*)`)
	seen := map[string]bool{}
	names := map[string]bool{}
	var visit func(string)
	visit = func(name string) {
		if seen[name] {
			return
		}
		seen[name] = true
		source, err := os.ReadFile(filepath.Join(libraries, "lib."+name+".d.ts"))
		if err != nil {
			t.Fatal(err)
		}
		for _, match := range reference.FindAllSubmatch(source, -1) {
			visit(string(match[1]))
		}
		for _, match := range declaration.FindAllSubmatch(source, -1) {
			names[string(match[1])] = true
		}
	}
	visit("es2024")
	actual := []string{}
	for name := range names {
		actual = append(actual, name)
	}
	sort.Strings(actual)
	source, err := os.ReadFile("testdata/es2024-globals.json")
	if err != nil {
		t.Fatal(err)
	}
	var expected []string
	if err := json.Unmarshal(source, &expected); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("pinned ECMAScript globals changed: got %v want %v", actual, expected)
	}
}
