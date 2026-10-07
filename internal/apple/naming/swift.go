package naming

import (
	"fmt"
	"strings"
)

// This file ports the synchronous Swift 3 naming transformations from
// swift/lib/Basic/StringExtras.cpp (omitNeedlessWords and its helpers),
// swift/lib/ClangImporter/ImportName.cpp (selector/initializer labels), and
// swift/lib/ClangImporter/ImportEnumInfo.cpp (common word/plural prefix),
// commit f79ab15f11ffea6cad859da22b17da6cd9252167 (swift-6.0-RELEASE).
// Copyright (c) 2014 - 2017 Apple Inc. and the Swift project authors.
// Apache-2.0 WITH Swift-exception; see THIRD_PARTY_NOTICES.md.
// Modified: LLVM ranges become Go word slices; clang facts are supplied by the
// caller. Async/error/overlay rewriting is outside this synchronous layer.

type omissionType struct {
	name, element                      string
	defaultArgument, boolean, function bool
}

func omission(t Type) omissionType {
	s := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(t.Spelling), "*"))
	switch s {
	case "id":
		s = "AnyObject"
	case "BOOL", "bool", "_Bool":
		s = "Bool"
	case "NSInteger", "NSUInteger", "int", "unsigned int":
		s = "Int"
	}
	return omissionType{name: s, element: t.Element, boolean: s == "Bool", function: t.Function != nil || strings.Contains(s, "(^")}
}
func matchWord(n, t string) bool {
	if len(n) > len(t) || n == "" {
		return false
	}
	if strings.EqualFold(n, t) {
		return true
	}
	if strings.HasSuffix(strings.ToLower(t), strings.ToLower(n)) && !lower(t[len(t)-len(n)]) {
		for i := 0; i < len(n); i++ {
			if lower(t[i]) || t[i] == '_' {
				return false
			}
		}
		return true
	}
	if strings.HasPrefix(strings.ToLower(t), strings.ToLower(n)) {
		for i := len(n); i < len(t); i++ {
			if !digit(t[i]) {
				return false
			}
		}
		return true
	}
	return false
}
func skipSuffix(s string) string {
	for _, suffix := range []string{"Type", "Ref", "Mask", "_t"} {
		if strings.HasSuffix(s, suffix) && len(s) > len(suffix) {
			return strings.TrimSuffix(s, suffix)
		}
	}
	if strings.HasSuffix(s, "D") {
		i := len(s) - 1
		for i > 0 && digit(s[i-1]) {
			i--
		}
		if i < len(s)-1 {
			return s[:i]
		}
	}
	return s
}
func matchLeading(n, t string) string {
	nw, tw := words(n), words(t)
	if len(nw) == 0 {
		return n
	}
	for i := len(tw) - 1; i >= 0; i-- {
		if matchWord(nw[0], tw[i]) {
			if len(nw) < len(tw)-i {
				return n
			}
			for j := 1; j < len(tw)-i; j++ {
				if !matchWord(nw[j], tw[i+j]) {
					return n
				}
			}
			return strings.Join(nw[len(tw)-i:], "")
		}
	}
	return n
}

// matchBack returns the first matching word, anchored at the name's end.
func matchBack(n string, t omissionType) int {
	nw, tw := words(n), words(t.name)
	i, j := len(nw)-1, len(tw)-1
	for i >= 0 && j >= 0 {
		if matchWord(nw[i], tw[j]) {
			i--
			j--
			continue
		}
		if (strings.EqualFold(nw[i], "Indexes") || strings.EqualFold(nw[i], "Indices")) && tw[j] == "Set" && j > 0 && matchWord("Index", tw[j-1]) {
			i--
			j -= 2
			continue
		}
		if strings.EqualFold(nw[i], "Index") && (matchWord("Int", tw[j]) || matchWord("Integer", tw[j])) {
			i--
			j--
			continue
		}
		if strings.EqualFold(tw[j], "Object") && strings.EqualFold(nw[i], "Value") && i > 0 && strings.EqualFold(nw[i-1], "Object") {
			i -= 2
			j--
			continue
		}
		if t.element != "" && len(nw[i]) > 2 && strings.HasSuffix(nw[i], "s") {
			shortened := strings.TrimSuffix(strings.Join(nw[:i+1], ""), "s")
			next := trailing(shortened, omissionType{name: t.element}, "partial", nil)
			if next != shortened {
				i = len(words(next)) - 1
				continue
			}
		}
		if i == len(nw)-1 {
			short := skipSuffix(t.name)
			if short != t.name {
				t.name = short
				tw = words(short)
				j = len(tw) - 1
				continue
			}
		}
		break
	}
	return i + 1
}
func vacuous(s string) bool {
	return strings.Contains("|get|for|set|using|with|", "|"+strings.ToLower(s)+"|")
}
func memberName(s string) bool {
	switch s {
	case "init", "Protocol", "self", "Type":
		return false
	}
	return s != ""
}
func propertyMatch(s string, properties []string) bool {
	s = swiftLower(s)
	for _, p := range properties {
		if p == s || p == s+"s" || p == s+"es" || (strings.HasSuffix(s, "y") && p == strings.TrimSuffix(s, "y")+"ies") {
			return true
		}
	}
	return false
}
func trailing(n string, t omissionType, role string, properties []string) string {
	if n == "" || t.name == "" {
		return n
	}
	ws := words(n)
	i := matchBack(n, t)
	if i == len(ws) {
		return n
	}
	if i == 0 {
		if role == "first" || role == "partial" {
			return ""
		}
		return n
	}
	if i == len(ws)-1 && ws[i] == "Error" {
		return n
	}
	previous := speech(ws[i-1])
	if role != "property" {
		if previous == "" {
			return n
		}
		if role == "base" {
			if previous == "preposition" && i == 1 {
				return n
			}
			if previous != "preposition" && propertyMatch(strings.Join(ws[i:], ""), properties) {
				return n
			}
		}
	}
	next := strings.Join(ws[:i], "")
	if (role == "base" || role == "property") && (!memberName(next) || vacuous(next)) {
		return n
	}
	return next
}
func omitSelf(n string, t omissionType, properties []string) string {
	if t.name == "" {
		return n
	}
	t.element = ""
	ws := words(n)
	for end := len(ws); end > 0; end-- {
		i := matchBack(strings.Join(ws[:end], ""), t)
		if i != end {
			if i == 0 || (end == len(ws) && i == end-1 && ws[i] == "Error") || speech(ws[i-1]) != "verb" || propertyMatch(strings.Join(ws[i:end], ""), properties) {
				return n
			}
			next := strings.Join(ws[:i], "") + strings.Join(ws[end:], "")
			if !memberName(next) || vacuous(next) {
				return n
			}
			return next
		}
		for {
			short := skipSuffix(t.name)
			if short == t.name {
				break
			}
			t.name = short
		}
	}
	return n
}
func splitBase(n string, t omissionType, parameter string) (string, string) {
	ws := words(n)
	if len(ws) == 0 {
		return n, ""
	}
	if t.boolean && ws[len(ws)-1] == "Animated" {
		return strings.TrimSuffix(n, "Animated"), "animated"
	}
	if ws[0] == "set" || (parameter == "sender" && strings.HasSuffix(t.name, "Object")) {
		return n, ""
	}
	i := -1
	for j := len(ws) - 1; j >= 0; j-- {
		if speech(ws[j]) == "preposition" {
			i = j
			break
		}
	}
	if i < 0 {
		return n, ""
	}
	if strings.EqualFold(ws[i], "of") {
		for j := i - 1; j >= 0; j-- {
			if speech(ws[j]) == "preposition" {
				if !strings.EqualFold(ws[j], "of") && !strings.EqualFold(ws[j], "for") {
					i = j
				}
				break
			}
		}
	}
	pre := strings.ToLower(ws[i])
	before, after := "", ""
	if i > 0 {
		before = strings.ToLower(ws[i-1])
	}
	if i+1 < len(ws) {
		after = strings.ToLower(ws[i+1])
	}
	if (pre == "in" && before == "plug") || (pre == "with" && (after == "error" || after == "no")) || ((pre == "to" || pre == "from") && after == "backing") || (pre == "to" && after == "visible") {
		return n, ""
	}
	start := i
	if (before == "compatible" && pre == "with") || (before == "best" && pre == "matching") || (before == "according" && pre == "to") || (before == "bound" && pre == "by") || (before == "separated" && pre == "by") {
		start--
	}
	argumentStart, baseEnd := start, start
	if i+1 == len(ws)-1 && (ws[i+1] == "X" || ws[i+1] == "Y" || ws[i+1] == "Z") {
		argumentStart = i + 1
		baseEnd = i + 1
	}
	if (pre == "with" || pre == "using") && after != "zone" && (t.defaultArgument || t.function) {
		argumentStart = i + 1
		baseEnd = start
	}
	if baseEnd == 0 {
		return n, ""
	}
	base := strings.Join(ws[:baseEnd], "")
	if (!memberName(ws[0]) && baseEnd <= 1) || (vacuous(ws[0]) && baseEnd <= 2) {
		return n, ""
	}
	return base, strings.Join(ws[argumentStart:], "")
}
func omitNeedlessWords(base string, labels []string, d Declaration) (string, []string) {
	context, result := omission(Type{Spelling: d.Parent}), omission(d.Result)
	if d.Result.Spelling == "instancetype" {
		result = context
	}
	same := context.name != "" && context.name == result.name
	if same {
		next := matchLeading(base, context.name)
		ws := words(next)
		if len(ws) > 1 && speech(ws[0]) == "preposition" {
			base = next
			if ws[0] == "By" && strings.HasSuffix(ws[1], "ing") {
				base = strings.Join(ws[1:], "")
			}
		}
	}
	if d.Kind != Property {
		base = omitSelf(base, context, d.PropertyNames)
	}
	if len(d.Parameters) == 0 {
		if same {
			base = trailing(base, result, "property", d.PropertyNames)
		}
	} else {
		if words(base)[0] == "set" {
			base = trailing(base, context, "property", d.PropertyNames)
		}
		first := omission(d.Parameters[0].Type)
		first.defaultArgument = d.Parameters[0].DefaultArgument
		if labels[0] == "" {
			base, labels[0] = splitBase(base, first, d.Parameters[0].Name)
		}
		for i, p := range d.Parameters {
			t := omission(p.Type)
			role := "subsequent"
			if i == 0 && labels[i] == "" {
				role = "base"
			} else if i == 0 && base != "init" && !p.DefaultArgument {
				role = "first"
			}
			if role == "base" {
				base = trailing(base, t, role, d.PropertyNames)
			} else {
				labels[i] = trailing(labels[i], t, role, nil)
			}
		}
	}
	base = swiftLower(base)
	for i, l := range labels {
		labels[i] = swiftLower(l)
	}
	return base, labels
}

// Apple's Foundation API notes remove NS for these reference types. The URL/URLRequest
// bridges are included for the task's documented names; other value-type
// overlays (String, Array, Date, etc.) are not synthesized here.
// Keep the table explicit so a new SDK rename cannot silently change a surface.
var foundationNames = map[string]string{
	"NSBlockOperation":                "BlockOperation",
	"NSBundle":                        "Bundle",
	"NSByteCountFormatter":            "ByteCountFormatter",
	"NSDateComponentsFormatter":       "DateComponentsFormatter",
	"NSDateFormatter":                 "DateFormatter",
	"NSDateIntervalFormatter":         "DateIntervalFormatter",
	"NSEnergyFormatter":               "EnergyFormatter",
	"NSFileHandle":                    "FileHandle",
	"NSFileManager":                   "FileManager",
	"NSFormatter":                     "Formatter",
	"NSHTTPURLResponse":               "HTTPURLResponse",
	"NSLengthFormatter":               "LengthFormatter",
	"NSMassFormatter":                 "MassFormatter",
	"NSMeasurementFormatter":          "MeasurementFormatter",
	"NSNotificationCenter":            "NotificationCenter",
	"NSNumberFormatter":               "NumberFormatter",
	"NSOperation":                     "Operation",
	"NSOperationQueue":                "OperationQueue",
	"NSPersonNameComponentsFormatter": "PersonNameComponentsFormatter",
	"NSProcessInfo":                   "ProcessInfo",
	"NSProgress":                      "Progress",
	"NSRunLoop":                       "RunLoop",
	"NSScanner":                       "Scanner",
	"NSThread":                        "Thread",
	"NSTimer":                         "Timer",
	"NSURL":                           "URL",
	"NSURLRequest":                    "URLRequest",
	"NSURLResponse":                   "URLResponse",
	"NSURLSession":                    "URLSession",
	"NSURLSessionConfiguration":       "URLSessionConfiguration",
	"NSURLSessionDataTask":            "URLSessionDataTask",
	"NSURLSessionDownloadTask":        "URLSessionDownloadTask",
	"NSURLSessionTask":                "URLSessionTask",
	"NSURLSessionUploadTask":          "URLSessionUploadTask",
	"NSUndoManager":                   "UndoManager",
	"NSUserDefaults":                  "UserDefaults",
}

func importedType(n, framework string) string {
	if framework == "Foundation" {
		if name, ok := foundationNames[n]; ok {
			return name
		}
	}
	return n
}
func parentName(d Declaration) string {
	if d.ParentSwiftName != "" {
		return d.ParentSwiftName
	}
	return importedType(d.Parent, d.Framework)
}

func swiftName(d Declaration) (string, string, []string, error) {
	callable := isCallable(d.Kind) && d.AccessorProperty == ""
	var base string
	var labels []string
	context := parentName(d)
	if explicitName(d) != "" {
		name := explicitName(d)
		if callable {
			left := strings.IndexByte(name, '(')
			if left < 0 || !strings.HasSuffix(name, ")") {
				return "", "", nil, fmt.Errorf("invalid Swift callable name %q", name)
			}
			base = name[:left]
			args := name[left+1 : len(name)-1]
			if args != "" {
				if !strings.HasSuffix(args, ":") {
					return "", "", nil, fmt.Errorf("invalid Swift labels %q", name)
				}
				labels = strings.Split(strings.TrimSuffix(args, ":"), ":")
			}
			if len(labels) != len(d.Parameters) {
				return "", "", nil, fmt.Errorf("Swift label count does not match parameters for %s", d.Name)
			}
		} else {
			base = name
		}
		if dot := strings.LastIndexByte(base, '.'); dot >= 0 {
			context = base[:dot]
			base = base[dot+1:]
		}
	} else if callable {
		if d.Kind == CFunction {
			base = d.Name
			labels = make([]string, len(d.Parameters))
		} else {
			count := strings.Count(d.Name, ":")
			if count != len(d.Parameters) || (count > 0 && !strings.HasSuffix(d.Name, ":")) {
				return "", "", nil, fmt.Errorf("selector/parameter mismatch for %q", d.Name)
			}
			pieces := strings.Split(strings.TrimSuffix(d.Name, ":"), ":")
			if !identifier(pieces[0]) {
				return "", "", nil, fmt.Errorf("invalid first selector piece %q", pieces[0])
			}
			base = pieces[0]
			labels = make([]string, count)
			for i := 1; i < count; i++ {
				labels[i] = pieces[i]
			}
			factoryTail := matchLeading(base, d.Parent)
			resultType := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(d.Result.Spelling), "*"))
			factory := d.Kind == ClassMethod && factoryTail != base && (resultType == d.Parent || resultType == "instancetype")
			if factory {
				base = "init" + factoryTail
			}
			if strings.HasPrefix(base, "init") && (len(base) == 4 || !lower(base[4])) {
				tail := strings.TrimPrefix(base, "init")
				tail = strings.TrimPrefix(tail, "With")
				base = "init"
				if count == 0 && tail != "" {
					return "", "", nil, fmt.Errorf("zero-argument named initializer %s needs an explicit Swift name", d.Name)
				}
				if count > 0 {
					labels[0] = swiftLower(tail)
				}
			}
		}
		if d.Kind != CFunction && !d.KeepNeedlessWords {
			base, labels = omitNeedlessWords(base, labels, d)
		}
	} else if isCase(d.Kind) {
		if len(d.Enumerators) == 0 {
			return "", "", nil, fmt.Errorf("%s needs all enum siblings", d.Name)
		}
		prefix := enumPrefix(d.Parent, d.Enumerators)
		base = strings.TrimPrefix(d.Name, prefix)
		if base == d.Name && prefix != "" {
			base = strings.TrimPrefix(d.Name, commonWordPrefix(prefix, d.Name))
		}
		base = swiftLower(base)
	} else {
		base = importedType(d.Name, d.Framework)
		if d.AccessorProperty != "" {
			base = d.AccessorProperty
		}
		if d.Kind == Property {
			if omission(d.Result).boolean && d.Getter != "" {
				base = d.Getter
			}
			// Swift property omission uses context type, not arbitrary result type.
			base, _ = omitNeedlessWords(base, nil, d)
		}
	}
	if !identifier(base) {
		return "", "", nil, fmt.Errorf("invalid imported identifier %q", base)
	}
	for i, l := range labels {
		if l == "" {
			labels[i] = "_"
		} else if l != "_" && !identifier(l) {
			return "", "", nil, fmt.Errorf("invalid Swift label %q", l)
		}
	}
	if d.RefinedForSwift {
		if base == "init" && len(labels) > 0 {
			labels[0] = "__" + labels[0]
		} else {
			base = "__" + base
		}
	}
	name := base
	if context != "" {
		name = context + "." + name
	}
	if callable {
		name += "("
		for _, l := range labels {
			name += l + ":"
		}
		name += ")"
	}
	return name, base, labels, nil
}

func commonWordPrefix(a, b string) string {
	aw, bw := words(a), words(b)
	i := 0
	for i < len(aw) && i < len(bw) && aw[i] == bw[i] {
		i++
	}
	if (i < len(aw) && !identifier(aw[i])) || (i < len(bw) && !identifier(bw[i])) {
		if i > 0 {
			i--
		}
	}
	return strings.Join(aw[:i], "")
}
func commonPluralPrefix(singular, plural string) string {
	common := commonWordPrefix(singular, plural)
	if common == singular || !strings.HasSuffix(plural, "s") {
		return common
	}
	rest := strings.TrimPrefix(singular, common)
	ws := words(rest)
	if len(ws) == 0 {
		return common
	}
	candidate := common + ws[0]
	if strings.HasPrefix(plural, candidate+"s") || strings.HasPrefix(plural, candidate+"es") || (strings.HasSuffix(candidate, "y") && strings.HasPrefix(plural, strings.TrimSuffix(candidate, "y")+"ies")) {
		return candidate
	}
	return common
}
func enumPrefix(parent string, cases []Enumerator) string {
	var eligible []string
	available := false
	for _, c := range cases {
		if c.SwiftName == "" && !c.Deprecated && !c.Unavailable {
			available = true
		}
	}
	for _, c := range cases {
		if c.SwiftName != "" || (available && (c.Deprecated || c.Unavailable)) {
			continue
		}
		eligible = append(eligible, c.Name)
	}
	if len(eligible) == 0 {
		return ""
	}
	prefix := eligible[0]
	for _, n := range eligible[1:] {
		prefix = commonWordPrefix(prefix, n)
	}
	if prefix == "" {
		return ""
	}
	check := prefix
	delta := 0
	if check[0] == 'k' && (len(check) == 1 || upper(check[1])) {
		check = check[1:]
		delta = 1
	}
	common := commonPluralPrefix(check, parent)
	if len(common) < len(check) && check[len(common)] == '_' {
		delta++
	}
	return prefix[:len(common)+delta]
}
