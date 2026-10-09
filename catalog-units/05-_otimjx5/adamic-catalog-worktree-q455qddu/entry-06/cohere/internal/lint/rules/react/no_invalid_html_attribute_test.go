package react

import (
	"strings"
	"testing"

	rule_testing "github.com/system-inc/cohere/internal/lint/testing"
)

// noInvalidHtmlAttributeFile is where the fixtures pretend to live.
//
// A `.tsx` name because every JSX case needs a parser that reads JSX. It is NOT load-bearing:
// this rule has no file gate, and TestNoInvalidHtmlAttributeHasNoFileGate pins that by running
// one reporting createElement source under two extensions.
const noInvalidHtmlAttributeFile = "/repository/source/NoInvalidHtmlAttribute.tsx"

// noInvalidHtmlAttributeCase is one row of upstream's corpus.
type noInvalidHtmlAttributeCase struct {
	// name is the corpus list and index the row came from, so a failure names a case that can
	// be found in upstream's own file rather than a number local to this table.
	name string

	// source is upstream's `code`, byte for byte.
	source string

	// ids are the message ids upstream's installed 7.37.5 build produced, in order. Every case
	// in this corpus reports at most one finding, which is asserted rather than assumed by
	// TestNoInvalidHtmlAttributeCorpusIsSingleFinding below.
	ids []string
}

// noInvalidHtmlAttributeFiresCases are the rows upstream reports on.
var noInvalidHtmlAttributeFiresCases = []noInvalidHtmlAttributeCase{
	{
		name:   "invalid-0",
		source: "<a rel=\"alternatex\"></a>",
		ids:    []string{"neverValid"},
	},
	{
		name:   "invalid-1",
		source: "React.createElement(\"a\", { rel: \"alternatex\" })",
		ids:    []string{"neverValid"},
	},
	{
		name:   "invalid-2",
		source: "React.createElement(\"a\", { rel: [\"alternatex\"] })",
		ids:    []string{"neverValid"},
	},
	{
		name:   "invalid-3",
		source: "<a rel=\"alternatex alternate\"></a>",
		ids:    []string{"neverValid"},
	},
	{
		name:   "invalid-4",
		source: "React.createElement(\"a\", { rel: \"alternatex alternate\" })",
		ids:    []string{"neverValid"},
	},
	{
		name:   "invalid-5",
		source: "React.createElement(\"a\", { rel: [\"alternatex alternate\"] })",
		ids:    []string{"neverValid"},
	},
	{
		name:   "invalid-6",
		source: "<a rel=\"alternate alternatex\"></a>",
		ids:    []string{"neverValid"},
	},
	{
		name:   "invalid-7",
		source: "React.createElement(\"a\", { rel: \"alternate alternatex\" })",
		ids:    []string{"neverValid"},
	},
	{
		name:   "invalid-8",
		source: "React.createElement(\"a\", { rel: [\"alternate alternatex\"] })",
		ids:    []string{"neverValid"},
	},
	{
		name:   "invalid-9",
		source: "<html rel></html>",
		ids:    []string{"onlyMeaningfulFor"},
	},
	{
		name:   "invalid-10",
		source: "React.createElement(\"html\", { rel: 1 })",
		ids:    []string{"onlyMeaningfulFor"},
	},
	{
		name:   "invalid-11",
		source: "<a rel></a>",
		ids:    []string{"emptyIsMeaningless"},
	},
	{
		name:   "invalid-12",
		source: "React.createElement(\"a\", { rel: 1 })",
		ids:    []string{"neverValid"},
	},
	{
		name:   "invalid-13",
		source: "React.createElement(\"a\", { rel() { return 1; } })",
		ids:    []string{"noMethod"},
	},
	{
		name:   "invalid-14",
		source: "<span rel></span>",
		ids:    []string{"onlyMeaningfulFor"},
	},
	{
		name:   "invalid-15",
		source: "<a rel={null}></a>",
		ids:    []string{"onlyStrings"},
	},
	{
		name:   "invalid-16",
		source: "<a rel={5}></a>",
		ids:    []string{"onlyStrings"},
	},
	{
		name:   "invalid-17",
		source: "<a rel={true}></a>",
		ids:    []string{"onlyStrings"},
	},
	{
		name:   "invalid-18",
		source: "<a rel={{}}></a>",
		ids:    []string{"onlyStrings"},
	},
	{
		name:   "invalid-19",
		source: "<a rel={undefined}></a>",
		ids:    []string{"onlyStrings"},
	},
	{
		name:   "invalid-20",
		source: "<a rel=\"noreferrer noopener foobar\"></a>",
		ids:    []string{"neverValid"},
	},
	{
		name:   "invalid-21",
		source: "<a rel=\"noreferrer noopener   \"></a>",
		ids:    []string{"spaceDelimited"},
	},
	{
		name:   "invalid-22",
		source: "<a rel=\"noreferrer        noopener\"></a>",
		ids:    []string{"spaceDelimited"},
	},
	{
		name:   "invalid-23",
		source: "<a rel=\"noreferrer\u00a0\u00a0noopener\"></a>",
		ids:    []string{"spaceDelimited"},
	},
	{
		name:   "invalid-24",
		source: "<a rel={\"noreferrer noopener foobar\"}></a>",
		ids:    []string{"neverValid"},
	},
	{
		name:   "invalid-25",
		source: "React.createElement(\"a\", { rel: [\"noreferrer\", \"noopener\", \"foobar\" ] })",
		ids:    []string{"neverValid"},
	},
	{
		name:   "invalid-26",
		source: "<a rel={\"foobar noreferrer noopener\"}></a>",
		ids:    []string{"neverValid"},
	},
	{
		name:   "invalid-27",
		source: "<a rel={\"        noopener\"}></a>",
		ids:    []string{"spaceDelimited"},
	},
	{
		name:   "invalid-28",
		source: "<a rel={\"noopener        \"}></a>",
		ids:    []string{"spaceDelimited"},
	},
	{
		name:   "invalid-29",
		source: "<a rel={\"batgo noopener\"}></a>",
		ids:    []string{"neverValid"},
	},
	{
		name:   "invalid-30",
		source: "<a rel={\" noopener\"}></a>",
		ids:    []string{"spaceDelimited"},
	},
	{
		name:   "invalid-31",
		source: "<a rel=\"canonical\"></a>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-32",
		source: "<a rel=\"dns-prefetch\"></a>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-33",
		source: "<a rel=\"icon\"></a>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-34",
		source: "<link rel=\"shortcut\"></link>",
		ids:    []string{"notAlone"},
	},
	{
		name:   "invalid-35",
		source: "<link rel=\"shortcut  icon\"></link>",
		ids:    []string{"spaceDelimited"},
	},
	{
		name:   "invalid-36",
		source: "<a rel=\"manifest\"></a>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-37",
		source: "<a rel=\"modulepreload\"></a>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-38",
		source: "<a rel=\"pingback\"></a>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-39",
		source: "<a rel=\"preconnect\"></a>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-40",
		source: "<a rel=\"prefetch\"></a>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-41",
		source: "<a rel=\"preload\"></a>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-42",
		source: "<a rel=\"prerender\"></a>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-43",
		source: "<a rel=\"stylesheet\"></a>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-44",
		source: "<area rel=\"canonical\"></area>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-45",
		source: "<area rel=\"dns-prefetch\"></area>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-46",
		source: "<area rel=\"icon\"></area>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-47",
		source: "<area rel=\"manifest\"></area>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-48",
		source: "<area rel=\"modulepreload\"></area>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-49",
		source: "<area rel=\"pingback\"></area>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-50",
		source: "<area rel=\"preconnect\"></area>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-51",
		source: "<area rel=\"prefetch\"></area>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-52",
		source: "<area rel=\"preload\"></area>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-53",
		source: "<area rel=\"prerender\"></area>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-54",
		source: "<area rel=\"stylesheet\"></area>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-55",
		source: "<link rel=\"bookmark\"></link>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-56",
		source: "<link rel=\"external\"></link>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-57",
		source: "<link rel=\"nofollow\"></link>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-58",
		source: "<link rel=\"noopener\"></link>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-59",
		source: "<link rel=\"noreferrer\"></link>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-60",
		source: "<link rel=\"opener\"></link>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-61",
		source: "<link rel=\"tag\"></link>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-62",
		source: "<form rel=\"alternate\"></form>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-63",
		source: "<form rel=\"author\"></form>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-64",
		source: "<form rel=\"bookmark\"></form>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-65",
		source: "<form rel=\"canonical\"></form>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-66",
		source: "<form rel=\"dns-prefetch\"></form>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-67",
		source: "<form rel=\"icon\"></form>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-68",
		source: "<form rel=\"manifest\"></form>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-69",
		source: "<form rel=\"modulepreload\"></form>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-70",
		source: "<form rel=\"pingback\"></form>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-71",
		source: "<form rel=\"preconnect\"></form>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-72",
		source: "<form rel=\"prefetch\"></form>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-73",
		source: "<form rel=\"preload\"></form>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-74",
		source: "<form rel=\"prerender\"></form>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-75",
		source: "<form rel=\"stylesheet\"></form>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-76",
		source: "<form rel=\"tag\"></form>",
		ids:    []string{"notValidFor"},
	},
	{
		name:   "invalid-77",
		source: "<form rel=\"\"></form>",
		ids:    []string{"noEmpty"},
	},
}

// noInvalidHtmlAttributeSilentCases are the rows upstream is clean on.
var noInvalidHtmlAttributeSilentCases = []noInvalidHtmlAttributeCase{
	{
		name:   "valid-0",
		source: "<a rel=\"alternate\"></a>",
	},
	{
		name:   "valid-1",
		source: "React.createElement(\"a\", { rel: \"alternate\" })",
	},
	{
		name:   "valid-2",
		source: "React.createElement(\"a\", { rel: [\"alternate\"] })",
	},
	{
		name:   "valid-3",
		source: "<a rel=\"author\"></a>",
	},
	{
		name:   "valid-4",
		source: "React.createElement(\"a\", { rel: \"author\" })",
	},
	{
		name:   "valid-5",
		source: "React.createElement(\"a\", { rel: [\"author\"] })",
	},
	{
		name:   "valid-6",
		source: "<a rel=\"bookmark\"></a>",
	},
	{
		name:   "valid-7",
		source: "React.createElement(\"a\", { rel: \"bookmark\" })",
	},
	{
		name:   "valid-8",
		source: "React.createElement(\"a\", { rel: [\"bookmark\"] })",
	},
	{
		name:   "valid-9",
		source: "<a rel=\"external\"></a>",
	},
	{
		name:   "valid-10",
		source: "React.createElement(\"a\", { rel: \"external\" })",
	},
	{
		name:   "valid-11",
		source: "React.createElement(\"a\", { rel: [\"external\"] })",
	},
	{
		name:   "valid-12",
		source: "<a rel=\"help\"></a>",
	},
	{
		name:   "valid-13",
		source: "React.createElement(\"a\", { rel: \"help\" })",
	},
	{
		name:   "valid-14",
		source: "React.createElement(\"a\", { rel: [\"help\"] })",
	},
	{
		name:   "valid-15",
		source: "<a rel=\"license\"></a>",
	},
	{
		name:   "valid-16",
		source: "React.createElement(\"a\", { rel: \"license\" })",
	},
	{
		name:   "valid-17",
		source: "React.createElement(\"a\", { rel: [\"license\"] })",
	},
	{
		name:   "valid-18",
		source: "<a rel=\"next\"></a>",
	},
	{
		name:   "valid-19",
		source: "React.createElement(\"a\", { rel: \"next\" })",
	},
	{
		name:   "valid-20",
		source: "React.createElement(\"a\", { rel: [\"next\"] })",
	},
	{
		name:   "valid-21",
		source: "<a rel=\"nofollow\"></a>",
	},
	{
		name:   "valid-22",
		source: "React.createElement(\"a\", { rel: \"nofollow\" })",
	},
	{
		name:   "valid-23",
		source: "React.createElement(\"a\", { rel: [\"nofollow\"] })",
	},
	{
		name:   "valid-24",
		source: "<a rel=\"noopener\"></a>",
	},
	{
		name:   "valid-25",
		source: "React.createElement(\"a\", { rel: \"noopener\" })",
	},
	{
		name:   "valid-26",
		source: "React.createElement(\"a\", { rel: [\"noopener\"] })",
	},
	{
		name:   "valid-27",
		source: "<a rel=\"noreferrer\"></a>",
	},
	{
		name:   "valid-28",
		source: "React.createElement(\"a\", { rel: \"noreferrer\" })",
	},
	{
		name:   "valid-29",
		source: "React.createElement(\"a\", { rel: [\"noreferrer\"] })",
	},
	{
		name:   "valid-30",
		source: "<a rel=\"opener\"></a>",
	},
	{
		name:   "valid-31",
		source: "React.createElement(\"a\", { rel: \"opener\" })",
	},
	{
		name:   "valid-32",
		source: "React.createElement(\"a\", { rel: [\"opener\"] })",
	},
	{
		name:   "valid-33",
		source: "<a rel=\"prev\"></a>",
	},
	{
		name:   "valid-34",
		source: "React.createElement(\"a\", { rel: \"prev\" })",
	},
	{
		name:   "valid-35",
		source: "React.createElement(\"a\", { rel: [\"prev\"] })",
	},
	{
		name:   "valid-36",
		source: "<a rel=\"search\"></a>",
	},
	{
		name:   "valid-37",
		source: "React.createElement(\"a\", { rel: \"search\" })",
	},
	{
		name:   "valid-38",
		source: "React.createElement(\"a\", { rel: [\"search\"] })",
	},
	{
		name:   "valid-39",
		source: "<a rel=\"tag\"></a>",
	},
	{
		name:   "valid-40",
		source: "React.createElement(\"a\", { rel: \"tag\" })",
	},
	{
		name:   "valid-41",
		source: "React.createElement(\"a\", { rel: [\"tag\"] })",
	},
	{
		name:   "valid-42",
		source: "<area rel=\"alternate\"></area>",
	},
	{
		name:   "valid-43",
		source: "React.createElement(\"area\", { rel: \"alternate\" })",
	},
	{
		name:   "valid-44",
		source: "React.createElement(\"area\", { rel: [\"alternate\"] })",
	},
	{
		name:   "valid-45",
		source: "<area rel=\"author\"></area>",
	},
	{
		name:   "valid-46",
		source: "React.createElement(\"area\", { rel: \"author\" })",
	},
	{
		name:   "valid-47",
		source: "React.createElement(\"area\", { rel: [\"author\"] })",
	},
	{
		name:   "valid-48",
		source: "<area rel=\"bookmark\"></area>",
	},
	{
		name:   "valid-49",
		source: "React.createElement(\"area\", { rel: \"bookmark\" })",
	},
	{
		name:   "valid-50",
		source: "React.createElement(\"area\", { rel: [\"bookmark\"] })",
	},
	{
		name:   "valid-51",
		source: "<area rel=\"external\"></area>",
	},
	{
		name:   "valid-52",
		source: "React.createElement(\"area\", { rel: \"external\" })",
	},
	{
		name:   "valid-53",
		source: "React.createElement(\"area\", { rel: [\"external\"] })",
	},
	{
		name:   "valid-54",
		source: "<area rel=\"help\"></area>",
	},
	{
		name:   "valid-55",
		source: "React.createElement(\"area\", { rel: \"help\" })",
	},
	{
		name:   "valid-56",
		source: "React.createElement(\"area\", { rel: [\"help\"] })",
	},
	{
		name:   "valid-57",
		source: "<area rel=\"license\"></area>",
	},
	{
		name:   "valid-58",
		source: "React.createElement(\"area\", { rel: \"license\" })",
	},
	{
		name:   "valid-59",
		source: "React.createElement(\"area\", { rel: [\"license\"] })",
	},
	{
		name:   "valid-60",
		source: "<area rel=\"next\"></area>",
	},
	{
		name:   "valid-61",
		source: "React.createElement(\"area\", { rel: \"next\" })",
	},
	{
		name:   "valid-62",
		source: "React.createElement(\"area\", { rel: [\"next\"] })",
	},
	{
		name:   "valid-63",
		source: "<area rel=\"nofollow\"></area>",
	},
	{
		name:   "valid-64",
		source: "React.createElement(\"area\", { rel: \"nofollow\" })",
	},
	{
		name:   "valid-65",
		source: "React.createElement(\"area\", { rel: [\"nofollow\"] })",
	},
	{
		name:   "valid-66",
		source: "<area rel=\"noopener\"></area>",
	},
	{
		name:   "valid-67",
		source: "React.createElement(\"area\", { rel: \"noopener\" })",
	},
	{
		name:   "valid-68",
		source: "React.createElement(\"area\", { rel: [\"noopener\"] })",
	},
	{
		name:   "valid-69",
		source: "<area rel=\"noreferrer\"></area>",
	},
	{
		name:   "valid-70",
		source: "React.createElement(\"area\", { rel: \"noreferrer\" })",
	},
	{
		name:   "valid-71",
		source: "React.createElement(\"area\", { rel: [\"noreferrer\"] })",
	},
	{
		name:   "valid-72",
		source: "<area rel=\"opener\"></area>",
	},
	{
		name:   "valid-73",
		source: "React.createElement(\"area\", { rel: \"opener\" })",
	},
	{
		name:   "valid-74",
		source: "React.createElement(\"area\", { rel: [\"opener\"] })",
	},
	{
		name:   "valid-75",
		source: "<area rel=\"prev\"></area>",
	},
	{
		name:   "valid-76",
		source: "React.createElement(\"area\", { rel: \"prev\" })",
	},
	{
		name:   "valid-77",
		source: "React.createElement(\"area\", { rel: [\"prev\"] })",
	},
	{
		name:   "valid-78",
		source: "<area rel=\"search\"></area>",
	},
	{
		name:   "valid-79",
		source: "React.createElement(\"area\", { rel: \"search\" })",
	},
	{
		name:   "valid-80",
		source: "React.createElement(\"area\", { rel: [\"search\"] })",
	},
	{
		name:   "valid-81",
		source: "<area rel=\"tag\"></area>",
	},
	{
		name:   "valid-82",
		source: "React.createElement(\"area\", { rel: \"tag\" })",
	},
	{
		name:   "valid-83",
		source: "React.createElement(\"area\", { rel: [\"tag\"] })",
	},
	{
		name:   "valid-84",
		source: "<link rel=\"alternate\"></link>",
	},
	{
		name:   "valid-85",
		source: "React.createElement(\"link\", { rel: \"alternate\" })",
	},
	{
		name:   "valid-86",
		source: "React.createElement(\"link\", { rel: [\"alternate\"] })",
	},
	{
		name:   "valid-87",
		source: "<link rel=\"author\"></link>",
	},
	{
		name:   "valid-88",
		source: "React.createElement(\"link\", { rel: \"author\" })",
	},
	{
		name:   "valid-89",
		source: "React.createElement(\"link\", { rel: [\"author\"] })",
	},
	{
		name:   "valid-90",
		source: "<link rel=\"canonical\"></link>",
	},
	{
		name:   "valid-91",
		source: "React.createElement(\"link\", { rel: \"canonical\" })",
	},
	{
		name:   "valid-92",
		source: "React.createElement(\"link\", { rel: [\"canonical\"] })",
	},
	{
		name:   "valid-93",
		source: "<link rel=\"dns-prefetch\"></link>",
	},
	{
		name:   "valid-94",
		source: "React.createElement(\"link\", { rel: \"dns-prefetch\" })",
	},
	{
		name:   "valid-95",
		source: "React.createElement(\"link\", { rel: [\"dns-prefetch\"] })",
	},
	{
		name:   "valid-96",
		source: "<link rel=\"help\"></link>",
	},
	{
		name:   "valid-97",
		source: "React.createElement(\"link\", { rel: \"help\" })",
	},
	{
		name:   "valid-98",
		source: "React.createElement(\"link\", { rel: [\"help\"] })",
	},
	{
		name:   "valid-99",
		source: "<link rel=\"icon\"></link>",
	},
	{
		name:   "valid-100",
		source: "React.createElement(\"link\", { rel: \"icon\" })",
	},
	{
		name:   "valid-101",
		source: "React.createElement(\"link\", { rel: [\"icon\"] })",
	},
	{
		name:   "valid-102",
		source: "<link rel=\"shortcut icon\"></link>",
	},
	{
		name:   "valid-103",
		source: "React.createElement(\"link\", { rel: \"shortcut icon\" })",
	},
	{
		name:   "valid-104",
		source: "React.createElement(\"link\", { rel: [\"shortcut icon\"] })",
	},
	{
		name:   "valid-105",
		source: "<link rel=\"license\"></link>",
	},
	{
		name:   "valid-106",
		source: "React.createElement(\"link\", { rel: \"license\" })",
	},
	{
		name:   "valid-107",
		source: "React.createElement(\"link\", { rel: [\"license\"] })",
	},
	{
		name:   "valid-108",
		source: "<link rel=\"manifest\"></link>",
	},
	{
		name:   "valid-109",
		source: "React.createElement(\"link\", { rel: \"manifest\" })",
	},
	{
		name:   "valid-110",
		source: "React.createElement(\"link\", { rel: [\"manifest\"] })",
	},
	{
		name:   "valid-111",
		source: "<link rel=\"modulepreload\"></link>",
	},
	{
		name:   "valid-112",
		source: "React.createElement(\"link\", { rel: \"modulepreload\" })",
	},
	{
		name:   "valid-113",
		source: "React.createElement(\"link\", { rel: [\"modulepreload\"] })",
	},
	{
		name:   "valid-114",
		source: "<link rel=\"next\"></link>",
	},
	{
		name:   "valid-115",
		source: "React.createElement(\"link\", { rel: \"next\" })",
	},
	{
		name:   "valid-116",
		source: "React.createElement(\"link\", { rel: [\"next\"] })",
	},
	{
		name:   "valid-117",
		source: "<link rel=\"pingback\"></link>",
	},
	{
		name:   "valid-118",
		source: "React.createElement(\"link\", { rel: \"pingback\" })",
	},
	{
		name:   "valid-119",
		source: "React.createElement(\"link\", { rel: [\"pingback\"] })",
	},
	{
		name:   "valid-120",
		source: "<link rel=\"preconnect\"></link>",
	},
	{
		name:   "valid-121",
		source: "React.createElement(\"link\", { rel: \"preconnect\" })",
	},
	{
		name:   "valid-122",
		source: "React.createElement(\"link\", { rel: [\"preconnect\"] })",
	},
	{
		name:   "valid-123",
		source: "<link rel=\"prefetch\"></link>",
	},
	{
		name:   "valid-124",
		source: "React.createElement(\"link\", { rel: \"prefetch\" })",
	},
	{
		name:   "valid-125",
		source: "React.createElement(\"link\", { rel: [\"prefetch\"] })",
	},
	{
		name:   "valid-126",
		source: "<link rel=\"preload\"></link>",
	},
	{
		name:   "valid-127",
		source: "React.createElement(\"link\", { rel: \"preload\" })",
	},
	{
		name:   "valid-128",
		source: "React.createElement(\"link\", { rel: [\"preload\"] })",
	},
	{
		name:   "valid-129",
		source: "<link rel=\"prerender\"></link>",
	},
	{
		name:   "valid-130",
		source: "React.createElement(\"link\", { rel: \"prerender\" })",
	},
	{
		name:   "valid-131",
		source: "React.createElement(\"link\", { rel: [\"prerender\"] })",
	},
	{
		name:   "valid-132",
		source: "<link rel=\"prev\"></link>",
	},
	{
		name:   "valid-133",
		source: "React.createElement(\"link\", { rel: \"prev\" })",
	},
	{
		name:   "valid-134",
		source: "React.createElement(\"link\", { rel: [\"prev\"] })",
	},
	{
		name:   "valid-135",
		source: "<link rel=\"search\"></link>",
	},
	{
		name:   "valid-136",
		source: "React.createElement(\"link\", { rel: \"search\" })",
	},
	{
		name:   "valid-137",
		source: "React.createElement(\"link\", { rel: [\"search\"] })",
	},
	{
		name:   "valid-138",
		source: "<link rel=\"stylesheet\"></link>",
	},
	{
		name:   "valid-139",
		source: "React.createElement(\"link\", { rel: \"stylesheet\" })",
	},
	{
		name:   "valid-140",
		source: "React.createElement(\"link\", { rel: [\"stylesheet\"] })",
	},
	{
		name:   "valid-141",
		source: "<form rel=\"external\"></form>",
	},
	{
		name:   "valid-142",
		source: "React.createElement(\"form\", { rel: \"external\" })",
	},
	{
		name:   "valid-143",
		source: "React.createElement(\"form\", { rel: [\"external\"] })",
	},
	{
		name:   "valid-144",
		source: "<form rel=\"help\"></form>",
	},
	{
		name:   "valid-145",
		source: "React.createElement(\"form\", { rel: \"help\" })",
	},
	{
		name:   "valid-146",
		source: "React.createElement(\"form\", { rel: [\"help\"] })",
	},
	{
		name:   "valid-147",
		source: "<form rel=\"license\"></form>",
	},
	{
		name:   "valid-148",
		source: "React.createElement(\"form\", { rel: \"license\" })",
	},
	{
		name:   "valid-149",
		source: "React.createElement(\"form\", { rel: [\"license\"] })",
	},
	{
		name:   "valid-150",
		source: "<form rel=\"next\"></form>",
	},
	{
		name:   "valid-151",
		source: "React.createElement(\"form\", { rel: \"next\" })",
	},
	{
		name:   "valid-152",
		source: "React.createElement(\"form\", { rel: [\"next\"] })",
	},
	{
		name:   "valid-153",
		source: "<form rel=\"nofollow\"></form>",
	},
	{
		name:   "valid-154",
		source: "React.createElement(\"form\", { rel: \"nofollow\" })",
	},
	{
		name:   "valid-155",
		source: "React.createElement(\"form\", { rel: [\"nofollow\"] })",
	},
	{
		name:   "valid-156",
		source: "<form rel=\"noopener\"></form>",
	},
	{
		name:   "valid-157",
		source: "React.createElement(\"form\", { rel: \"noopener\" })",
	},
	{
		name:   "valid-158",
		source: "React.createElement(\"form\", { rel: [\"noopener\"] })",
	},
	{
		name:   "valid-159",
		source: "<form rel=\"noreferrer\"></form>",
	},
	{
		name:   "valid-160",
		source: "React.createElement(\"form\", { rel: \"noreferrer\" })",
	},
	{
		name:   "valid-161",
		source: "React.createElement(\"form\", { rel: [\"noreferrer\"] })",
	},
	{
		name:   "valid-162",
		source: "<form rel=\"opener\"></form>",
	},
	{
		name:   "valid-163",
		source: "React.createElement(\"form\", { rel: \"opener\" })",
	},
	{
		name:   "valid-164",
		source: "React.createElement(\"form\", { rel: [\"opener\"] })",
	},
	{
		name:   "valid-165",
		source: "<form rel=\"prev\"></form>",
	},
	{
		name:   "valid-166",
		source: "React.createElement(\"form\", { rel: \"prev\" })",
	},
	{
		name:   "valid-167",
		source: "React.createElement(\"form\", { rel: [\"prev\"] })",
	},
	{
		name:   "valid-168",
		source: "<form rel=\"search\"></form>",
	},
	{
		name:   "valid-169",
		source: "React.createElement(\"form\", { rel: \"search\" })",
	},
	{
		name:   "valid-170",
		source: "React.createElement(\"form\", { rel: [\"search\"] })",
	},
	{
		name:   "valid-171",
		source: "<form rel={callFoo()}></form>",
	},
	{
		name:   "valid-172",
		source: "React.createElement(\"form\", { rel: callFoo() })",
	},
	{
		name:   "valid-173",
		source: "React.createElement(\"form\", { rel: [callFoo()] })",
	},
	{
		name:   "valid-174",
		source: "<a rel={{a: \"noreferrer\"}[\"a\"]}></a>",
	},
	{
		name:   "valid-175",
		source: "<a rel={{a: \"noreferrer\"}[\"b\"]}></a>",
	},
	{
		name:   "valid-176",
		source: "<Foo rel></Foo>",
	},
	{
		name:   "valid-177",
		source: "React.createElement(\"Foo\", { rel: true })",
	},
	{
		name:   "valid-178",
		source: "\n        React.createElement('a', {\n          ...rest,\n          href: to,\n        })\n      ",
	},
	{
		name:   "valid-179",
		source: "<link rel=\"apple-touch-icon\" sizes=\"60x60\" href=\"apple-touch-icon-60x60.png\" />",
	},
	{
		name:   "valid-180",
		source: "<link rel=\"apple-touch-icon\" sizes=\"76x76\" href=\"apple-touch-icon-76x76.png\" />",
	},
	{
		name:   "valid-181",
		source: "<link rel=\"apple-touch-icon\" sizes=\"120x120\" href=\"apple-touch-icon-120x120.png\" />",
	},
	{
		name:   "valid-182",
		source: "<link rel=\"apple-touch-icon\" sizes=\"152x152\" href=\"apple-touch-icon-152x152.png\" />",
	},
	{
		name:   "valid-183",
		source: "<link rel=\"apple-touch-icon\" sizes=\"180x180\" href=\"apple-touch-icon-180x180.png\" />",
	},
	{
		name:   "valid-184",
		source: "<link rel=\"apple-touch-startup-image\" href=\"launch.png\" />",
	},
	{
		name:   "valid-185",
		source: "<link rel=\"apple-touch-startup-image\" href=\"iphone5.png\" media=\"(device-width: 320px) and (device-height: 568px) and (-webkit-device-pixel-ratio: 2)\" />",
	},
	{
		name:   "valid-186",
		source: "<link rel=\"mask-icon\" href=\"/safari-pinned-tab.svg\" color=\"#fff\" />",
	},
}

func TestNoInvalidHtmlAttributeFires(t *testing.T) {
	t.Parallel()
	for _, testCase := range noInvalidHtmlAttributeFiresCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoInvalidHtmlAttribute, noInvalidHtmlAttributeFile, testCase.source)
			rule_testing.ExpectFindings(t, result, testCase.ids...)
		})
	}
}

func TestNoInvalidHtmlAttributeStaysSilent(t *testing.T) {
	t.Parallel()
	for _, testCase := range noInvalidHtmlAttributeSilentCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoInvalidHtmlAttribute, noInvalidHtmlAttributeFile, testCase.source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoInvalidHtmlAttributeCorpusIsSingleFinding pins a property of the imported corpus that
// the tables above quietly rely on.
//
// Every one of upstream's 78 reporting cases asserts exactly one finding, so a rule that
// reported twice for one input would still satisfy a per-case id list built by zipping. If a
// future corpus import brings in a multi-finding case, this fails and says so rather than
// letting the table silently under-assert.
func TestNoInvalidHtmlAttributeCorpusIsSingleFinding(t *testing.T) {
	t.Parallel()
	if len(noInvalidHtmlAttributeFiresCases) == 0 {
		t.Fatal("no reporting cases, so this check proved nothing")
	}
	for _, testCase := range noInvalidHtmlAttributeFiresCases {
		if len(testCase.ids) != 1 {
			t.Errorf("%s asserts %d findings; the tables assume exactly one", testCase.name, len(testCase.ids))
		}
	}
}

// TestNoInvalidHtmlAttributeHasNoFileGate pins that this rule judges a `.ts` file.
//
// Three siblings in this package once declined every file not ending `.tsx`/`.jsx`, which was oxc
// residue rather than upstream behaviour. `eslint-plugin-react` gates this rule nowhere. The
// reporting source here is a `React.createElement` call rather than JSX, because that is the arm
// that is legal in a plain `.ts` file and therefore the one that can tell a working gate from an
// absent one.
func TestNoInvalidHtmlAttributeHasNoFileGate(t *testing.T) {
	t.Parallel()

	const reporting = "React.createElement(\"a\", { rel: \"alternatex\" });\n"

	for _, fileName := range []string{"/repository/source/Gate.tsx", "/repository/source/Gate.ts"} {
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoInvalidHtmlAttribute, fileName, reporting)
			rule_testing.ExpectFindings(t, result, "neverValid")
		})
	}
}

// TestNoInvalidHtmlAttributeReportsWhereItSays asserts spans, not just message ids.
//
// `ExpectFindings` checks ids and count and nothing else, so a rule whose defect is where it points
// passes a complete fixture pair while being wrong. This rule anchors on four different node kinds
// depending on the arm, which is exactly the situation where an id-only fixture set is blind.
func TestNoInvalidHtmlAttributeReportsWhereItSays(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
		id     string
		span   string
	}{
		{
			// The value arms anchor on the string literal, including its quotes.
			name:   "neverValid-anchors-on-the-literal",
			source: "<a rel=\"alternatex\"></a>;\n",
			id:     "neverValid",
			span:   "\"alternatex\"",
		},
		{
			name:   "notValidFor-anchors-on-the-literal",
			source: "<a rel=\"canonical\"></a>;\n",
			id:     "notValidFor",
			span:   "\"canonical\"",
		},
		{
			// A bare attribute anchors on the NAME, not the whole attribute.
			name:   "emptyIsMeaningless-anchors-on-the-name",
			source: "<a rel></a>;\n",
			id:     "emptyIsMeaningless",
			span:   "rel",
		},
		{
			// So does the not-meaningful-here arm.
			name:   "onlyMeaningfulFor-anchors-on-the-name",
			source: "<span rel></span>;\n",
			id:     "onlyMeaningfulFor",
			span:   "rel",
		},
		{
			// A non-string literal in braces anchors on the literal itself.
			name:   "onlyStrings-anchors-on-the-literal",
			source: "<a rel={5}></a>;\n",
			id:     "onlyStrings",
			span:   "5",
		},
		{
			// An object or `undefined` anchors on the whole expression container instead.
			name:   "onlyStrings-anchors-on-the-container-for-an-object",
			source: "<a rel={{}}></a>;\n",
			id:     "onlyStrings",
			span:   "{{}}",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoInvalidHtmlAttribute, noInvalidHtmlAttributeFile, testCase.source)
			rule_testing.ExpectFindings(t, result, testCase.id)
			if len(result.Diagnostics) != 1 {
				return
			}
			reported := testCase.source[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
			if reported != testCase.span {
				t.Errorf("finding spans %q, expected %q", reported, testCase.span)
			}
		})
	}
}

// TestNoInvalidHtmlAttributeRendersTheTagListInUpstreamOrder pins the one message whose text depends
// on an iteration order.
//
// `onlyMeaningfulFor` lists the tags an attribute is meaningful on, and upstream builds that list by
// iterating a JavaScript Set, which preserves insertion order. Go map iteration is randomised, so
// the tags live in a slice; sorting instead would render them alphabetically. Three corpus cases
// assert this text, but they assert it through a message ID, which cannot see the order at all.
func TestNoInvalidHtmlAttributeRendersTheTagListInUpstreamOrder(t *testing.T) {
	t.Parallel()

	const wanted = "<link>, <a>, <area>, <form>"
	if rendered := noInvalidHtmlAttributeRenderTagNames("rel"); rendered != wanted {
		t.Errorf("rendered %q, upstream renders %q", rendered, wanted)
	}
}

// TestNoInvalidHtmlAttributeSeparatesTheTwoValueArms pins an asymmetry that reads as a bug.
//
// The JSX arm splits a value on whitespace and judges each token, so
// `rel="alternate alternatex"` reports once for the bad token alone. The createElement arm looks the
// WHOLE string up in the table, so the same value reports once naming the entire string. That is
// upstream's, not a simplification here, and corpus cases invalid[4] through invalid[8] pin both
// halves; this test states the contrast in one place because reading it out of eight scattered rows
// is how a later reader "fixes" one arm to match the other.
func TestNoInvalidHtmlAttributeSeparatesTheTwoValueArms(t *testing.T) {
	t.Parallel()

	t.Run("jsx-splits-on-whitespace", func(t *testing.T) {
		t.Parallel()
		const source = "<a rel=\"alternate alternatex\"></a>;\n"
		result := rule_testing.Run(t, NoInvalidHtmlAttribute, noInvalidHtmlAttributeFile, source)
		rule_testing.ExpectFindings(t, result, "neverValid")
		if len(result.Diagnostics) == 1 &&
			!strings.Contains(result.Diagnostics[0].Message.Description, "`alternatex`") {
			t.Errorf("expected the message to name the bad token alone, got %q",
				result.Diagnostics[0].Message.Description)
		}
	})

	t.Run("createElement-looks-up-the-whole-string", func(t *testing.T) {
		t.Parallel()
		const source = "React.createElement(\"a\", { rel: \"alternate alternatex\" });\n"
		result := rule_testing.Run(t, NoInvalidHtmlAttribute, noInvalidHtmlAttributeFile, source)
		rule_testing.ExpectFindings(t, result, "neverValid")
		if len(result.Diagnostics) == 1 &&
			!strings.Contains(result.Diagnostics[0].Message.Description, "`alternate alternatex`") {
			t.Errorf("expected the message to name the whole string, got %q",
				result.Diagnostics[0].Message.Description)
		}
	})
}

// TestNoInvalidHtmlAttributeSurvivesAMemberItCannotName covers the shapes whose names cannot be read.
//
// `Node.Text()` panics on several kinds rather than returning empty, and the walk recovers per FILE
// rather than per rule, so one unguarded read costs every rule in this package its verdict on that
// file. Each shape below reaches a different name read: a computed property key, a namespaced JSX
// attribute, a member-expression tag name, and a spread with no name at all.
func TestNoInvalidHtmlAttributeSurvivesAMemberItCannotName(t *testing.T) {
	t.Parallel()

	sources := map[string]string{
		"computed-property-key":  "const k = \"rel\";\nReact.createElement(\"a\", { [k]: \"alternatex\" });\n",
		"member-expression-tag":  "<Foo.Bar rel=\"alternatex\"></Foo.Bar>;\n",
		"spread-attribute":       "const p = {};\n<a {...p}></a>;\n",
		"spread-in-create-props": "const p = {};\nReact.createElement(\"a\", { ...p });\n",
	}

	for name, source := range sources {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// The assertion that matters is that this returns a verdict at all rather than panicking.
			// All four are shapes the rule cannot read, so all four are clean.
			result := rule_testing.Run(t, NoInvalidHtmlAttribute, noInvalidHtmlAttributeFile, source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// The three tests below exist because a mutation sweep found three discriminations upstream's own
// corpus cannot see. Each was settled by driving the installed 7.37.5 build on inputs written for
// the question, with a control that fires, rather than by reading the reference implementation.

// TestNoInvalidHtmlAttributeChecksTheCreateElementReceiver covers the `React.` half of
// `isValidCreateElement`.
//
// Upstream's test is purely syntactic and reads `callee.object.name === 'React'`, so it needs no
// scope resolution and no type checker. Every one of the corpus's createElement cases spells the
// receiver `React`, so removing the receiver check survives all of them.
//
// Measured on the installed build with a control that fires: `Preact.createElement`,
// `h.createElement` and a bare `createElement` are all silent on a value that reports under
// `React.createElement`. This is deliberately NARROWER than this package's shared
// `isPragmaCreateElementCall`, which accepts the bare form; the difference is upstream's.
func TestNoInvalidHtmlAttributeChecksTheCreateElementReceiver(t *testing.T) {
	t.Parallel()

	t.Run("control-React-receiver-reports", func(t *testing.T) {
		t.Parallel()
		result := rule_testing.Run(t, NoInvalidHtmlAttribute, noInvalidHtmlAttributeFile,
			"React.createElement(\"a\", { rel: \"alternatex\" });\n")
		rule_testing.ExpectFindings(t, result, "neverValid")
	})

	for name, source := range map[string]string{
		"Preact-receiver": "Preact.createElement(\"a\", { rel: \"alternatex\" });\n",
		"lowercase-h":     "h.createElement(\"a\", { rel: \"alternatex\" });\n",
		"bare-call":       "createElement(\"a\", { rel: \"alternatex\" });\n",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoInvalidHtmlAttribute, noInvalidHtmlAttributeFile, source)
			rule_testing.ExpectClean(t, result)
		})
	}
}

// TestNoInvalidHtmlAttributeGatesTheCreateElementElementName covers a real defect this sweep found.
//
// Upstream's HTML-element gate is guarded by a STRING test:
//
//	if (typeof elemNameArg.value === 'string' && !HTML_ELEMENTS.has(elemNameArg.value))
//
// so a non-string literal skips the gate and falls through to the props check, where it fails
// against the attribute's permitted tags. An earlier version of this port required a string literal
// and went silent on `React.createElement(5, ...)`, which upstream reports.
//
// Dropping the gate entirely SURVIVED the whole 265-case corpus, so the survivor was investigated
// rather than written off. Driving the installed build on the three shapes the corpus does not
// write settled it, and the numeric one is where the two disagreed.
func TestNoInvalidHtmlAttributeGatesTheCreateElementElementName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
		id     string
	}{
		// The control: a literal HTML element, which reports on its value.
		{"control-literal-html-element", "React.createElement(\"a\", { rel: \"alternatex\" });\n", "neverValid"},
		// A non-string literal skips the HTML gate and is judged by the props check instead.
		{"numeric-element-skips-the-gate", "React.createElement(5, { rel: \"alternatex\" });\n", "onlyMeaningfulFor"},
		// A non-element string is gated out.
		{"non-html-string-is-gated-out", "React.createElement(\"Foo\", { rel: \"alternatex\" });\n", ""},
		// Neither an identifier nor a template is a Literal, so both decline before the gate.
		{"identifier-element-declines", "const tag = \"a\";\nReact.createElement(tag, { rel: \"alternatex\" });\n", ""},
		{"template-element-declines", "React.createElement(`a`, { rel: \"alternatex\" });\n", ""},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoInvalidHtmlAttribute, noInvalidHtmlAttributeFile, testCase.source)
			if testCase.id == "" {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.id)
		})
	}
}

// TestNoInvalidHtmlAttributeWhitespaceIsJavaScriptWhitespace pins the predicate that one corpus case
// caught and that two mutants could not.
//
// `invalid[23]` separates two valid tokens with NO-BREAK SPACEs, so nothing else reports and the
// whole finding is lost if the predicate is ASCII-only. An earlier version of this rule scanned
// bytes with an ASCII predicate and documented the gap as a deliberate narrowing; the corpus proved
// that reasoning wrong.
//
// Two mutants on this predicate SURVIVE and both are equivalent rather than uncaught: U+00A0 is in
// `unicode.Zs`, so removing it from the explicit list leaves the category fallback catching it, and
// removing only the fallback leaves the explicit entry. Removing BOTH is caught by 5 assertions,
// which is what establishes the predicate is load-bearing.
//
// Every code point below is written as an escape rather than as a literal character, because the
// whole file is scanned for non-ASCII bytes and because an editing tool has twice in this repository
// silently rewritten a non-ASCII byte in Go source.
func TestNoInvalidHtmlAttributeWhitespaceIsJavaScriptWhitespace(t *testing.T) {
	t.Parallel()

	// What JavaScript's `\s` matches that a naive ASCII predicate would miss.
	for name, character := range map[string]rune{
		"no-break-space":      '\u00a0',
		"byte-order-mark":     '\ufeff',
		"line-separator":      '\u2028',
		"paragraph-separator": '\u2029',
		"en-quad":             '\u2000',
		"ideographic-space":   '\u3000',
		"ogham-space-mark":    '\u1680',
		"narrow-no-break":     '\u202f',
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if !noInvalidHtmlAttributeIsSpaceRune(character) {
				t.Errorf("U+%04X is JavaScript whitespace and the predicate says it is not", character)
			}
		})
	}

	// The controls. Without these an always-true predicate would pass everything above.
	//
	// U+200B ZERO WIDTH SPACE is the sharp one: it looks like whitespace, is named like whitespace,
	// and is NOT matched by JavaScript's `\s`, because it is in category Cf rather than Zs.
	t.Run("control-these-are-not-javascript-whitespace", func(t *testing.T) {
		t.Parallel()
		for _, character := range []rune{'a', '-', '\u200b', '\u0085'} {
			if noInvalidHtmlAttributeIsSpaceRune(character) {
				t.Errorf("U+%04X is not JavaScript whitespace and the predicate says it is", character)
			}
		}
	})
}

// TestNoInvalidHtmlAttributeTreatsAJsxEscapeAsLiteralText records an EQUIVALENT mutant and pins the
// behaviour that makes it equivalent.
//
// Replacing this rule's raw-source read with the cooked `Text()` survives the whole corpus and every
// input written to break it. That is not a fixture gap: a JSX attribute value is NOT
// escape-processed, so `rel="a\tb"` holds a literal backslash and a literal `t`, and raw and cooked
// are the same string for every JSX attribute there is.
//
// Measured on the installed build with a control that fires. The backslash form reports
// `neverValid`, because the whole thing is one unrecognised token; the real-tab form reports
// `spaceDelimited`, because a tab really is whitespace. Both are asserted here so that a future
// change which starts unescaping JSX attribute values fails loudly rather than silently converging
// on the wrong arm.
func TestNoInvalidHtmlAttributeTreatsAJsxEscapeAsLiteralText(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source string
		id     string
	}{
		// The control: a real double space is whitespace and takes the spacing arm.
		{"control-real-double-space", "<a rel=\"noreferrer  noopener\"></a>;\n", "spaceDelimited"},
		// A backslash-t in JSX is two literal characters, so the whole value is one bad token.
		{"backslash-t-is-literal-text", "<a rel=\"noreferrer\\tnoopener\"></a>;\n", "neverValid"},
		// A real tab and a real newline are whitespace.
		{"real-tab-is-whitespace", "<a rel=\"noreferrer\tnoopener\"></a>;\n", "spaceDelimited"},
		{"real-newline-is-whitespace", "<a rel=\"noreferrer\nnoopener\"></a>;\n", "spaceDelimited"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, NoInvalidHtmlAttribute, noInvalidHtmlAttributeFile, testCase.source)
			rule_testing.ExpectFindings(t, result, testCase.id)
		})
	}
}

// TestNoInvalidHtmlAttributeOptions pins the option's three spellings and its refusals, measured
// against the installed build (7.37.5) on the same source: `["rel"]` and no option report twice,
// and `[]` reports nothing, because upstream keeps an empty array over its default.
func TestNoInvalidHtmlAttributeOptions(t *testing.T) {
	t.Parallel()

	const source = `<a rel="bogus" />; React.createElement("a", {rel: "bogus"});`
	cases := []struct {
		raw       string
		wantCount int
	}{
		{``, 2},
		{`["rel"]`, 2},
		{`[]`, 0},
	}
	for _, testCase := range cases {
		t.Run(testCase.raw, func(t *testing.T) {
			t.Parallel()
			decoded, err := DecodeNoInvalidHtmlAttributeOptions([]byte(testCase.raw))
			if err != nil {
				t.Fatalf("decoding %q: %v", testCase.raw, err)
			}
			result := rule_testing.RunWithOptions(t, NoInvalidHtmlAttribute, noInvalidHtmlAttributeFile, source, decoded)
			if len(result.Diagnostics) != testCase.wantCount {
				t.Errorf("got %d findings, want %d", len(result.Diagnostics), testCase.wantCount)
			}
		})
	}

	for _, raw := range []string{`["rel", "rel"]`, `["href"]`, `"rel"`, `{}`, `null`} {
		t.Run(raw+" is refused", func(t *testing.T) {
			t.Parallel()
			if _, err := DecodeNoInvalidHtmlAttributeOptions([]byte(raw)); err == nil {
				t.Errorf("%s decoded; upstream's schema refuses it", raw)
			}
		})
	}
}
