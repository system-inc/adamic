package compose

// Ported from eemeli/yaml 2.9.0, dist/schema/yaml-1.1/: the YAML 1.1 schema (schema.js), which a
// document's %YAML 1.1 directive selects, and its tags, six of which (binary, merge, omap, pairs, set,
// timestamp) the core schema also reaches as known tags.

import (
	"encoding/binary"
	"math"
	"regexp"
)

// Ported from bool.js.

var trueTag = &Tag{
	Default: defaultTrue,
	Tag:     "tag:yaml.org,2002:bool",
	Test:    regexp.MustCompile(`^(?:Y|y|[Yy]es|YES|[Tt]rue|TRUE|[Oo]n|ON)$`),
	resolveScalar: func([]uint16, func(string), *Options) (any, error) {
		return newScalar(true), nil
	},
}

var falseTag = &Tag{
	Default: defaultTrue,
	Tag:     "tag:yaml.org,2002:bool",
	Test:    regexp.MustCompile(`^(?:N|n|[Nn]o|NO|[Ff]alse|FALSE|[Oo]ff|OFF)$`),
	resolveScalar: func([]uint16, func(string), *Options) (any, error) {
		return newScalar(false), nil
	},
}

// Ported from int.js.

// removeUnderscores is `text.replace(/_/g, ”)`.
func removeUnderscores(text []uint16) []uint16 {
	result := make([]uint16, 0, len(text))
	for _, unit := range text {
		if unit != '_' {
			result = append(result, unit)
		}
	}
	return result
}

// yaml11IntResolve is int.js's intResolve. intAsBigInt is never set on this path.
func yaml11IntResolve(value []uint16, offset int, radix int) float64 {
	sign := characterAt(value, 0)
	if sign == '-' || sign == '+' {
		offset++
	}
	digits := removeUnderscores(substring(value, offset, len(value)))
	n := parseInt(digits, radix)
	if sign == '-' {
		return -1 * n
	}
	return n
}

var yaml11IntBinTag = &Tag{
	Default: defaultTrue,
	Tag:     "tag:yaml.org,2002:int",
	Format:  "BIN",
	Test:    regexp.MustCompile(`^[-+]?0b[0-1_]+$`),
	resolveScalar: func(value []uint16, _ func(string), _ *Options) (any, error) {
		return yaml11IntResolve(value, 2, 2), nil
	},
}

var yaml11IntOctTag = &Tag{
	Default: defaultTrue,
	Tag:     "tag:yaml.org,2002:int",
	Format:  "OCT",
	Test:    regexp.MustCompile(`^[-+]?0[0-7_]+$`),
	resolveScalar: func(value []uint16, _ func(string), _ *Options) (any, error) {
		return yaml11IntResolve(value, 1, 8), nil
	},
}

var yaml11IntTag = &Tag{
	Default: defaultTrue,
	Tag:     "tag:yaml.org,2002:int",
	Test:    regexp.MustCompile(`^[-+]?[0-9][0-9_]*$`),
	resolveScalar: func(value []uint16, _ func(string), _ *Options) (any, error) {
		return yaml11IntResolve(value, 0, 10), nil
	},
}

var yaml11IntHexTag = &Tag{
	Default: defaultTrue,
	Tag:     "tag:yaml.org,2002:int",
	Format:  "HEX",
	Test:    regexp.MustCompile(`^[-+]?0x[0-9a-fA-F_]+$`),
	resolveScalar: func(value []uint16, _ func(string), _ *Options) (any, error) {
		return yaml11IntResolve(value, 2, 16), nil
	},
}

// Ported from float.js.

var yaml11FloatNaNTag = &Tag{
	Default:       defaultTrue,
	Tag:           "tag:yaml.org,2002:float",
	Test:          regexp.MustCompile(`^(?:[-+]?\.(?:inf|Inf|INF)|\.nan|\.NaN|\.NAN)$`),
	resolveScalar: floatNaNResolve,
}

var yaml11FloatExpTag = &Tag{
	Default: defaultTrue,
	Tag:     "tag:yaml.org,2002:float",
	Format:  "EXP",
	Test:    regexp.MustCompile(`^[-+]?(?:[0-9][0-9_]*)?(?:\.[0-9_]*)?[eE][-+]?[0-9]+$`),
	resolveScalar: func(value []uint16, _ func(string), _ *Options) (any, error) {
		return parseFloat(removeUnderscores(value)), nil
	},
}

var yaml11FloatTag = &Tag{
	Default: defaultTrue,
	Tag:     "tag:yaml.org,2002:float",
	Test:    regexp.MustCompile(`^[-+]?(?:[0-9][0-9_]*)?\.[0-9_]*$`),
	resolveScalar: func(value []uint16, _ func(string), _ *Options) (any, error) {
		node := newScalar(parseFloat(removeUnderscores(value)))
		dot := indexOfUnit(value, '.')
		if dot != -1 {
			f := removeUnderscores(substring(value, dot+1, len(value)))
			if characterAt(f, len(f)-1) == '0' {
				node.MinFractionDigits = len(f)
			}
		}
		return node, nil
	},
}

// Ported from binary.js. Upstream's resolve is `Buffer.from(src, 'base64')`.

var binaryTag = &Tag{
	Default: defaultFalse,
	Tag:     "tag:yaml.org,2002:binary",
	resolveScalar: func(value []uint16, _ func(string), _ *Options) (any, error) {
		return nodeBase64Decode(value), nil
	},
}

// nodeBase64Decode is Node's `Buffer.from(text, 'base64')`. Node first tries a strict decoder (WHATWG
// forgiving-base64) and, when that rejects the input, decodes with its legacy decoder; the two agree
// wherever the strict one accepts, so the legacy decoder alone is ported. It allocates
// base64ByteLength(text) bytes, reads each code unit by its low byte, skips characters outside the
// alphabet (both the standard and the URL-safe one), and stops at the first = or when the buffer is full.
func nodeBase64Decode(text []uint16) Binary {
	// lib/buffer.js base64ByteLength.
	length := len(text)
	if length > 0 && text[length-1] == '=' {
		length--
	}
	if length > 1 && text[length-1] == '=' {
		length--
	}
	destinationLength := int(uint32(length*3) >> 2)
	destination := make([]byte, 0, destinationLength)

	// nbytes Base64DecodedSize: up to two trailing = dropped, then Base64DecodedSizeFast.
	decodedSize := 0
	if size := len(text); size >= 2 {
		if text[size-1] == '=' {
			size--
			if text[size-1] == '=' {
				size--
			}
		}
		if size > 1 {
			decodedSize = size/4*3 + (size%4+1)/2
		}
	}

	// nbytes Base64DecodeFast.
	available := min(destinationLength, decodedSize)
	maxK := available / 3 * 3
	maxI := len(text) / 4 * 4
	i := 0
	for i < maxI && len(destination) < maxK {
		quad := [4]byte{unbase64(text[i]), unbase64(text[i+1]), unbase64(text[i+2]), unbase64(text[i+3])}
		if binary.BigEndian.Uint32(quad[:])&0x80808080 != 0 {
			var proceed bool
			destination, i, proceed = base64DecodeGroupSlow(destination, destinationLength, text, i)
			if !proceed {
				return Binary(destination)
			}
			maxI = i + (len(text)-i)/4*4
		} else {
			destination = append(destination,
				quad[0]<<2|quad[1]>>4, quad[1]<<4|quad[2]>>2, quad[2]<<6|quad[3])
			i += 4
		}
	}
	if i < len(text) && len(destination) < destinationLength {
		destination, _, _ = base64DecodeGroupSlow(destination, destinationLength, text, i)
	}
	return Binary(destination)
}

// base64DecodeGroupSlow is nbytes' Base64DecodeGroupSlow: one group of four alphabet characters,
// skipping others, writing as many bytes as fit. It returns false to stop decoding.
func base64DecodeGroupSlow(destination []byte, destinationLength int, text []uint16, i int) ([]byte, int, bool) {
	var hi, lo byte
	// next is the V macro's loop: the next alphabet character, false at = or the end.
	next := func() bool {
		for {
			c := byte(text[i])
			lo = unbase64(text[i])
			i++
			if lo < 64 {
				return true
			}
			if c == '=' || i >= len(text) {
				return false
			}
		}
	}
	// after is the V macro's tail: false at the end of the input or of the buffer.
	after := func() bool {
		if i >= len(text) || len(destination) >= destinationLength {
			return false
		}
		hi = lo
		return true
	}
	if !next() || !after() {
		return destination, i, false
	}
	if !next() {
		return destination, i, false
	}
	destination = append(destination, (hi&0x3F)<<2|(lo&0x30)>>4)
	if !after() || !next() {
		return destination, i, false
	}
	destination = append(destination, (hi&0x0F)<<4|(lo&0x3C)>>2)
	if !after() || !next() {
		return destination, i, false
	}
	destination = append(destination, (hi&0x03)<<6|lo&0x3F)
	if !after() {
		return destination, i, false
	}
	return destination, i, true
}

// unbase64 is nbytes' table lookup on a code unit's low byte: 0 to 63 for the standard and URL-safe
// alphabets, 255 otherwise.
func unbase64(unit uint16) byte {
	c := byte(unit)
	switch {
	case c >= 'A' && c <= 'Z':
		return c - 'A'
	case c >= 'a' && c <= 'z':
		return c - 'a' + 26
	case c >= '0' && c <= '9':
		return c - '0' + 52
	case c == '+' || c == '-':
		return 62
	case c == '/' || c == '_':
		return 63
	}
	return 255
}

// Ported from merge.js: the tag only. isMergeKey and addMergeToJSMap serve toJS.

var mergeTag = &Tag{
	Default: defaultKey,
	Tag:     "tag:yaml.org,2002:merge",
	Test:    regexp.MustCompile(`^<<$`),
	resolveScalar: func([]uint16, func(string), *Options) (any, error) {
		// `Object.assign(new Scalar(Symbol(MERGE_KEY)), { addToJSMap })`
		return newScalar(&MergeKey{}), nil
	},
}

// Ported from pairs.js: resolvePairs. createPairs serves createNode.

func resolvePairs(seq *Node, onError func(message string)) *Node {
	if IsSeq(seq) {
		for i := 0; i < len(seq.Items); i++ {
			item := seq.Items[i]
			if IsPair(item) {
				continue
			} else if IsMap(item) {
				if len(item.Items) > 1 {
					onError("Each pair must have its own sequence indicator")
				}
				// `item.items[0] || new Pair(new Scalar(null))`: the new key has no range.
				var pair *Node
				if len(item.Items) > 0 {
					pair = item.Items[0]
				} else {
					pair = newPair(newScalar(nil), nil)
				}
				if item.CommentBefore != "" {
					if pair.Key.CommentBefore != "" {
						pair.Key.CommentBefore = item.CommentBefore + "\n" + pair.Key.CommentBefore
					} else {
						pair.Key.CommentBefore = item.CommentBefore
					}
				}
				if item.Comment != "" {
					// `pair.value ?? pair.key`
					cn := pair.Value
					if cn == nil {
						cn = pair.Key
					}
					if cn.Comment != "" {
						cn.Comment = item.Comment + "\n" + cn.Comment
					} else {
						cn.Comment = item.Comment
					}
				}
				item = pair
			}
			if IsPair(item) {
				seq.Items[i] = item
			} else {
				seq.Items[i] = newPair(item, nil)
			}
		}
	} else {
		onError("Expected a sequence for this tag")
	}
	return seq
}

var pairsTag = &Tag{
	Collection:        "seq",
	Default:           defaultFalse,
	Tag:               "tag:yaml.org,2002:pairs",
	resolveCollection: resolvePairs,
}

// Ported from omap.js: the tag and YAMLOMap's construction. YAMLOMap's map methods and toJSON serve the
// document API and toJS.

var omapTag = &Tag{
	Collection: "seq",
	NodeClass:  "YAMLOMap",
	Default:    defaultFalse,
	Tag:        omapTagName,
	resolveCollection: func(seq *Node, onError func(message string)) *Node {
		pairs := resolvePairs(seq, onError)
		// `seenKeys.includes(key.value)` is SameValueZero.
		var seenKeys []any
		for _, item := range pairs.Items {
			key := item.Key
			if IsScalar(key) {
				if containsSameValueZero(seenKeys, key.ScalarValue) {
					onError("Ordered maps must not include duplicate keys: " + javaScriptString(key.ScalarValue))
				} else {
					seenKeys = append(seenKeys, key.ScalarValue)
				}
			}
		}
		return assignCollection("YAMLOMap", pairs)
	},
}

// Ported from set.js: the tag and YAMLSet's construction. YAMLSet's methods serve the document API.

var setTag = &Tag{
	Collection: "map",
	NodeClass:  "YAMLSet",
	Default:    defaultFalse,
	Tag:        setTagName,
	resolveCollection: func(collection *Node, onError func(message string)) *Node {
		if IsMap(collection) {
			if collection.hasAllNullValues(true) {
				return assignCollection("YAMLSet", collection)
			}
			onError("Set items must all have null values")
		} else {
			onError("Expected a mapping for this tag")
		}
		return collection
	},
}

// Ported from timestamp.js.

// parseSexagesimal is upstream's with asBigInt false: the sign, then the colon-separated parts in base
// 60, each read by Number after the underscores are removed.
func parseSexagesimal(text []uint16) float64 {
	sign := characterAt(text, 0)
	parts := text
	if sign == '-' || sign == '+' {
		parts = substring(text, 1, len(text))
	}
	parts = removeUnderscores(parts)
	result := 0.0
	start := 0
	for index := 0; index <= len(parts); index++ {
		if index == len(parts) || parts[index] == ':' {
			result = result*60 + number(parts[start:index])
			start = index + 1
		}
	}
	if sign == '-' {
		return -1 * result
	}
	return result
}

var intTimeTag = &Tag{
	Default: defaultTrue,
	Tag:     "tag:yaml.org,2002:int",
	Format:  "TIME",
	Test:    regexp.MustCompile(`^[-+]?[0-9][0-9_]*(?::[0-5]?[0-9])+$`),
	resolveScalar: func(value []uint16, _ func(string), _ *Options) (any, error) {
		return parseSexagesimal(value), nil
	},
}

var floatTimeTag = &Tag{
	Default: defaultTrue,
	Tag:     "tag:yaml.org,2002:float",
	Format:  "TIME",
	Test:    regexp.MustCompile(`^[-+]?[0-9][0-9_]*(?::[0-5]?[0-9])+\.[0-9_]*$`),
	resolveScalar: func(value []uint16, _ func(string), _ *Options) (any, error) {
		return parseSexagesimal(value), nil
	},
}

var timestampTag = &Tag{
	Default: defaultTrue,
	Tag:     "tag:yaml.org,2002:timestamp",
	// If the time zone is omitted, the timestamp is assumed to be specified in UTC. The time part
	// may be omitted altogether, resulting in a date format. In such a case, the time part is
	// assumed to be 00:00:00Z (start of day, UTC).
	Test: timestampRegexp,
	resolveScalar: func(value []uint16, _ func(string), _ *Options) (any, error) {
		return timestampResolve(value)
	},
}

var timestampRegexp = regexp.MustCompile(`^([0-9]{4})-([0-9]{1,2})-([0-9]{1,2})` + // YYYY-Mm-Dd
	`(?:` + // time is optional
	`(?:t|T|[ \t]+)` + // t | T | whitespace
	`([0-9]{1,2}):([0-9]{1,2}):([0-9]{1,2}(\.[0-9]+)?)` + // Hh:Mm:Ss(.ss)?
	`(?:[ \t]*(Z|[-+][012]?[0-9](?::[0-9]{2})?))?` + // Z | +5 | -03:30
	`)?$`)

func timestampResolve(value []uint16) (any, error) {
	match := timestampRegexp.FindStringSubmatch(asciiView(value))
	if match == nil {
		return nil, errorString("!!timestamp expects a date, starting with yyyy-mm-dd")
	}
	// `match.map(Number)`: an unmatched group is undefined, which Number makes NaN. Every group
	// that participates in a match is at least one character, so an empty group is an unmatched one.
	numbers := make([]float64, len(match))
	for index, group := range match {
		if group == "" && index > 0 {
			numbers[index] = nan()
			continue
		}
		numbers[index] = number(asciiUnits(group))
	}
	year, month, day := numbers[1], numbers[2], numbers[3]
	hour, minute, second := orZero(numbers[4]), orZero(numbers[5]), orZero(numbers[6])
	millisec := 0.0
	if match[7] != "" {
		millisec = number(substr(asciiUnits(match[7]+"00"), 1, 3))
	}
	date := dateUTC(year, month-1, day, hour, minute, second, millisec)
	tz := match[8]
	if tz != "" && tz != "Z" {
		d := parseSexagesimal(asciiUnits(tz))
		if math.Abs(d) < 30 {
			d *= 60
		}
		date -= 60000 * d
	}
	return Date(timeClip(date)), nil
}

// orZero is `n || 0` for a number.
func orZero(n float64) float64 {
	if math.IsNaN(n) || n == 0 {
		return 0
	}
	return n
}

// dateUTC is `Date.UTC(year, month, day, hour, minute, second, millisecond)`: a year from 0 to 99 means
// 1900 plus it, each field is truncated to an integer, and months and days overflow into the next.
func dateUTC(year, month, day, hour, minute, second, millisecond float64) float64 {
	for _, field := range []float64{year, month, day, hour, minute, second, millisecond} {
		if math.IsNaN(field) || math.IsInf(field, 0) {
			return nan()
		}
	}
	yearInteger := math.Trunc(year)
	if yearInteger >= 0 && yearInteger <= 99 {
		year = 1900 + yearInteger
	}
	// MakeDay.
	y, m, dt := math.Trunc(year), math.Trunc(month), math.Trunc(day)
	ym := y + math.Floor(m/12)
	mn := math.Mod(m, 12)
	if mn < 0 {
		mn += 12
	}
	days := daysFromCivil(int64(ym), int64(mn)+1, 1) + dt - 1
	// MakeTime.
	time := math.Trunc(hour)*3600000 + math.Trunc(minute)*60000 + math.Trunc(second)*1000 + math.Trunc(millisecond)
	return timeClip(days*86400000 + time)
}

// daysFromCivil is the number of days from 1970-01-01 to the proleptic Gregorian date.
func daysFromCivil(year int64, month int64, day int64) float64 {
	if month <= 2 {
		year--
	}
	era := year
	if era < 0 {
		era -= 399
	}
	era /= 400
	yearOfEra := year - era*400
	shifted := month + 9
	if month > 2 {
		shifted = month - 3
	}
	dayOfYear := (153*shifted+2)/5 + day - 1
	dayOfEra := yearOfEra*365 + yearOfEra/4 - yearOfEra/100 + dayOfYear
	return float64(era*146097 + dayOfEra - 719468)
}

// timeClip is ECMAScript's TimeClip.
func timeClip(time float64) float64 {
	if math.IsNaN(time) || math.Abs(time) > 8.64e15 {
		return nan()
	}
	return math.Trunc(time) + 0 // + 0 turns -0 into +0, as ToIntegerOrInfinity does
}

var yaml11Schema = []*Tag{
	mapTag,
	seqTag,
	stringTag,
	nullTag,
	trueTag,
	falseTag,
	yaml11IntBinTag,
	yaml11IntOctTag,
	yaml11IntTag,
	yaml11IntHexTag,
	yaml11FloatNaNTag,
	yaml11FloatExpTag,
	yaml11FloatTag,
	binaryTag,
	mergeTag,
	omapTag,
	pairsTag,
	setTag,
	intTimeTag,
	floatTimeTag,
	timestampTag,
}
