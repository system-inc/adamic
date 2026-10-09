package mediaquery

import (
	"errors"
	"unicode/utf8"

	"github.com/system-inc/cohere/internal/format/estree"
)

// postcss-media-query-parser 0.2.3, dist/parsers.js.
//
// # Offsets
//
// JavaScript indexes the string by UTF-16 unit; this walks bytes. Every test the library makes on a
// character compares it with an ASCII character, or asks whether it is whitespace (all of which is in
// the BMP), and no element ever ends between the two halves of a surrogate pair, so walking bytes (and
// runes where whitespace is asked) takes the same decisions. Every sourceIndex is a sum of the offsets
// and lengths of pieces of the input, so computed on bytes it comes out in bytes, which is what the
// brief asks for at the boundary.
//
// # Failures
//
// The library has two ways not to return. parseMediaFeature reads modesEntered[-1].mode, a TypeError,
// once a "}" pops more modes than were pushed; that is an error here with V8's message. And
// parseMediaList loops forever on a `url(` whose parentheses never close (string[i] runs past the end
// and is undefined forever); that cannot be ported, so it is an error here too, named for what it is.

// errorReadingMode is the TypeError V8 throws on modesEntered[lastModeIndex].mode with lastModeIndex -1.
var errorReadingMode = errors.New("Cannot read properties of undefined (reading 'mode')")

// errorUnclosedUrl stands for the infinite loop on an unclosed `url(`, where the library never returns.
var errorUnclosedUrl = errors.New("postcss-media-query-parser never returns on an unclosed url(: its parentheses loop reads past the end forever")

// mode is one entry of modesEntered. character is the quote a string mode was entered with.
type mode struct {
	mode                 string
	isCalculationEnabled bool
	character            byte
}

/**
 * Parses a media feature expression, e.g. `max-width: 10px`, `(color)`
 *
 * @param {string} string - the source expression string, can be inside parens
 * @param {Number} index - the index of `string` in the overall input
 *
 * @return {Array} an array of Nodes, the first element being a media feature,
 *    the secont - its value (may be missing)
 */
func parseMediaFeature(text string, index int) ([]*estree.Node, error) {
	modesEntered := []mode{{
		mode:      "normal",
		character: 0,
	}}
	result := []*estree.Node{}
	lastModeIndex := 0
	mediaFeature := ""
	var colon *estree.Node
	var mediaFeatureValue *estree.Node
	indexLocal := index

	stringNormalized := text
	// Strip trailing parens (if any), and correct the starting index
	if len(text) > 0 && text[0] == '(' && text[len(text)-1] == ')' {
		stringNormalized = text[1 : len(text)-1]
		indexLocal++
	}

	for i := 0; i < len(stringNormalized); i++ {
		character := stringNormalized[i]

		// If entering/exiting a string
		if character == '\'' || character == '"' {
			if modesEntered[lastModeIndex].isCalculationEnabled {
				modesEntered = append(modesEntered, mode{
					mode:                 "string",
					isCalculationEnabled: false,
					character:            character,
				})
				lastModeIndex++
			} else if modesEntered[lastModeIndex].mode == "string" && modesEntered[lastModeIndex].character == character && (i == 0 || stringNormalized[i-1] != '\\') {
				modesEntered = modesEntered[:len(modesEntered)-1]
				lastModeIndex--
			}
		}

		// If entering/exiting interpolation
		if character == '{' {
			modesEntered = append(modesEntered, mode{
				mode:                 "interpolation",
				isCalculationEnabled: true,
			})
			lastModeIndex++
		} else if character == '}' {
			modesEntered = modesEntered[:len(modesEntered)-1]
			lastModeIndex--
		}

		// The next line reads modesEntered[lastModeIndex].mode, which throws once a "}" has popped the
		// last mode. lastModeIndex is always len(modesEntered)-1: they move together until this throw.
		if lastModeIndex < 0 {
			return nil, errorReadingMode
		}

		// If a : is met outside of a string, function call or interpolation, than
		// this : separates a media feature and a value
		if modesEntered[lastModeIndex].mode == "normal" && character == ':' {
			mediaFeatureValueString := stringNormalized[i+1:]
			mediaFeatureValueBefore := leadingWhitespace(mediaFeatureValueString)
			mediaFeatureValue = estree.New("value", 0, 0,
				"before", mediaFeatureValueBefore,
				"after", trailingWhitespace(mediaFeatureValueString),
				"value", trim(mediaFeatureValueString),
			)
			// +1 for the colon
			mediaFeatureValue.Set("sourceIndex", len(mediaFeatureValueBefore)+i+1+indexLocal)
			colon = estree.New("colon", 0, 0,
				"sourceIndex", i+indexLocal,
				"after", mediaFeatureValueBefore,
				"value", ":",
			)
			break
		}

		// The byte itself, as text: string(character) made it a rune, so each byte of a character outside ASCII
		// became a character of its own, and a no-break space before a feature read as "\u00c2\u00a0", which is
		// not whitespace (found by @system_adamic's stream P2).
		mediaFeature += stringNormalized[i : i+1]
	}

	// Forming a media feature node
	mediaFeatureBefore := leadingWhitespace(mediaFeature)
	mediaFeatureAfter := trailingWhitespace(mediaFeature)
	mediaFeatureNode := estree.New("media-feature", 0, 0,
		"before", mediaFeatureBefore,
		"after", mediaFeatureAfter,
		"value", trim(mediaFeature),
	)
	mediaFeatureNode.Set("sourceIndex", len(mediaFeatureBefore)+indexLocal)
	result = append(result, mediaFeatureNode)

	if colon != nil {
		colon.Set("before", mediaFeatureAfter)
		result = append(result, colon)
	}

	if mediaFeatureValue != nil {
		result = append(result, mediaFeatureValue)
	}

	return result, nil
}

// mediaQueryElement is the plain object parseMediaQuery builds an element in (resetNode's shape plus
// the type, sourceIndex and nodes it may gain) before handing it to Node or Container.
type mediaQueryElement struct {
	before      string
	after       string
	value       string
	nodeType    string
	sourceIndex int
	nodes       []*estree.Node
}

/**
 * Parses a media query, e.g. `screen and (color)`, `only tv`
 *
 * @param {string} string - the source media query string
 * @param {Number} index - the index of `string` in the overall input
 *
 * @return {Array} an array of Nodes and Containers
 */
func parseMediaQuery(text string, index int) ([]*estree.Node, error) {
	result := []*estree.Node{}

	// How many timies the parser entered parens/curly braces
	localLevel := 0
	// Has any keyword, media type, media feature expression or interpolation
	// ('element' hereafter) started
	insideSomeValue := false

	resetNode := func() mediaQueryElement {
		return mediaQueryElement{
			before: "",
			after:  "",
			value:  "",
		}
	}

	node := resetNode()

	// Go walks runes where JavaScript walks units: the whitespace tests need the whole character.
	for i := 0; i < len(text); {
		decoded, size := utf8.DecodeRuneInString(text[i:])
		character := text[i : i+size]
		// If not yet entered any element
		if !insideSomeValue {
			if isWhitespace(decoded) {
				// A whitespace
				// Don't form 'after' yet; will do it later
				node.before += character
			} else {
				// Not a whitespace - entering an element
				// Expression start
				if character == "(" {
					node.nodeType = "media-feature-expression"
					localLevel++
				}
				node.value = character
				node.sourceIndex = index + i
				insideSomeValue = true
			}
		} else {
			// Already in the middle of some alement
			node.value += character

			// Here parens just increase localLevel and don't trigger a start of
			// a media feature expression (since they can't be nested)
			// Interpolation start
			if character == "{" || character == "(" {
				localLevel++
			}
			// Interpolation/function call/media feature expression end
			if character == ")" || character == "}" {
				localLevel--
			}
		}

		// If exited all parens/curlies and the next symbol
		if insideSomeValue && localLevel == 0 && (character == ")" || i+size == len(text) || startsWithWhitespace(text[i+size:])) {
			if node.value == "not" || node.value == "only" || node.value == "and" {
				node.nodeType = "keyword"
			}
			// if it's an expression, parse its contents
			if node.nodeType == "media-feature-expression" {
				nodes, err := parseMediaFeature(node.value, node.sourceIndex)
				if err != nil {
					return nil, err
				}
				node.nodes = nodes
			}
			if node.nodes != nil {
				result = append(result, newContainer(node.after, node.before, node.nodeType, node.value, node.sourceIndex, node.nodes))
			} else {
				result = append(result, newNode(node.after, node.before, node.nodeType, node.value, node.sourceIndex))
			}
			node = resetNode()
			insideSomeValue = false
		}

		i += size
	}

	// Now process the result array - to specify undefined types of the nodes
	// and specify the `after` prop
	for i := 0; i < len(result); i++ {
		node := result[i]
		if i > 0 {
			result[i-1].Set("after", node.Get("before"))
		}

		// Node types. Might not be set because contains interpolation/function
		// calls or fully consists of them
		if node.Type() == "" {
			if i > 0 {
				// only `and` can follow an expression
				if result[i-1].Type() == "media-feature-expression" {
					node.SetType("keyword")
					continue
				}
				// Anything after 'only|not' is a media type
				if result[i-1].String("value") == "not" || result[i-1].String("value") == "only" {
					node.SetType("media-type")
					continue
				}
				// Anything after 'and' is an expression
				if result[i-1].String("value") == "and" {
					node.SetType("media-feature-expression")
					continue
				}

				if result[i-1].Type() == "media-type" {
					// if it is the last element - it might be an expression
					// or 'and' depending on what is after it
					if i+1 >= len(result) {
						node.SetType("media-feature-expression")
					} else if result[i+1].Type() == "media-feature-expression" {
						node.SetType("keyword")
					} else {
						node.SetType("media-feature-expression")
					}
				}
			}

			if i == 0 {
				// `screen`, `fn( ... )`, `#{ ... }`. Not an expression, since then
				// its type would have been set by now
				if i+1 >= len(result) {
					node.SetType("media-type")
					continue
				}

				// `screen and` or `#{...} (max-width: 10px)`
				if result[i+1].Type() == "media-feature-expression" || result[i+1].Type() == "keyword" {
					node.SetType("media-type")
					continue
				}
				if i+2 < len(result) {
					// `screen and (color) ...`
					if result[i+2].Type() == "media-feature-expression" {
						node.SetType("media-type")
						result[i+1].SetType("keyword")
						continue
					}
					// `only screen and ...`
					if result[i+2].Type() == "keyword" {
						node.SetType("keyword")
						result[i+1].SetType("media-type")
						continue
					}
				}
				if i+3 < len(result) {
					// `screen and (color) ...`
					if result[i+3].Type() == "media-feature-expression" {
						node.SetType("keyword")
						result[i+1].SetType("media-type")
						result[i+2].SetType("keyword")
						continue
					}
				}
			}
		}
	}
	return result, nil
}

/**
 * Parses a media query list. Takes a possible `url()` at the start into
 * account, and divides the list into media queries that are parsed separately
 *
 * @param {string} string - the source media query list string
 *
 * @return {Array} an array of Nodes/Containers
 */
func parseMediaList(text string) ([]*estree.Node, error) {
	result := []*estree.Node{}
	interimIndex := 0
	levelLocal := 0

	// Check for a `url(...)` part (if it is contents of an @import rule)
	if doesHaveUrl, urlBefore := matchUrlStart(text); doesHaveUrl > 0 {
		i := doesHaveUrl
		parenthesesLevel := 1
		for parenthesesLevel > 0 {
			// string[i] past the end is undefined, and the loop never ends.
			if i >= len(text) {
				return nil, errorUnclosedUrl
			}
			character := text[i]
			if character == '(' {
				parenthesesLevel++
			}
			if character == ')' {
				parenthesesLevel--
			}
			i++
		}
		// result.unshift into an empty result.
		result = append([]*estree.Node{newNode(
			leadingWhitespace(text[i:]),
			urlBefore,
			"url",
			trim(text[:i]),
			len(urlBefore),
		)}, result...)
		interimIndex = i
	}

	// Start processing the media query list
	for i := interimIndex; i < len(text); i++ {
		character := text[i]

		// Dividing the media query list into comma-separated media queries
		// Only count commas that are outside of any parens
		// (i.e., not part of function call params list, etc.)
		if character == '(' {
			levelLocal++
		}
		if character == ')' {
			levelLocal--
		}
		if levelLocal == 0 && character == ',' {
			mediaQueryString := text[interimIndex:i]
			spaceBefore := leadingWhitespace(mediaQueryString)
			nodes, err := parseMediaQuery(mediaQueryString, interimIndex)
			if err != nil {
				return nil, err
			}
			result = append(result, newContainer(
				trailingWhitespace(mediaQueryString),
				spaceBefore,
				"media-query",
				trim(mediaQueryString),
				interimIndex+len(spaceBefore),
				nodes,
			))
			interimIndex = i + 1
		}
	}

	mediaQueryString := text[interimIndex:]
	spaceBefore := leadingWhitespace(mediaQueryString)
	nodes, err := parseMediaQuery(mediaQueryString, interimIndex)
	if err != nil {
		return nil, err
	}
	result = append(result, newContainer(
		trailingWhitespace(mediaQueryString),
		spaceBefore,
		"media-query",
		trim(mediaQueryString),
		interimIndex+len(spaceBefore),
		nodes,
	))

	return result, nil
}

// matchUrlStart is `/^(\s*)url\s*\(/.exec(string)`: the match's length (0 for no match, since a match
// is at least four characters long) and its first group.
func matchUrlStart(text string) (int, string) {
	before := leadingWhitespace(text)
	rest := text[len(before):]
	if len(rest) < 3 || rest[:3] != "url" {
		return 0, ""
	}
	afterUrl := rest[3:]
	afterUrl = afterUrl[len(leadingWhitespace(afterUrl)):]
	if len(afterUrl) == 0 || afterUrl[0] != '(' {
		return 0, ""
	}
	return len(text) - len(afterUrl) + 1, before
}
