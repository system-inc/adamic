package regexp

import "strings"

// These are the exact property/value spellings accepted by ECMAScript. Loose
// matching is intentionally not used by the language.
var generalCategories = set("C Cc Cf Cn Co Cs L LC Ll Lm Lo Lt Lu M Mc Me Mn N Nd Nl No P Pc Pd Pe Pf Pi Po Ps S Sc Sk Sm So Z Zl Zp Zs Other Control Format Unassigned Private_Use Surrogate Letter Cased_Letter Lowercase_Letter Modifier_Letter Other_Letter Titlecase_Letter Uppercase_Letter Mark Spacing_Mark Enclosing_Mark Nonspacing_Mark Number Decimal_Number Letter_Number Other_Number Punctuation Connector_Punctuation Dash_Punctuation Close_Punctuation Final_Punctuation Initial_Punctuation Other_Punctuation Open_Punctuation Symbol Currency_Symbol Modifier_Symbol Math_Symbol Other_Symbol Separator Line_Separator Paragraph_Separator Space_Separator")
var scripts = set("Adlam Ahom Anatolian_Hieroglyphs Arabic Armenian Avestan Balinese Bamum Bassa_Vah Batak Bengali Bhaiksuki Bopomofo Brahmi Braille Buginese Buhid Canadian_Aboriginal Carian Caucasian_Albanian Chakma Cham Cherokee Chorasmian Common Coptic Cuneiform Cypriot Cyrillic Deseret Devanagari Duployan Egyptian_Hieroglyphs Elbasan Elymaic Ethiopic Georgian Glagolitic Gothic Grantha Greek Gujarati Gunjala_Gondi Gurmukhi Han Hangul Hanifi_Rohingya Hanunoo Hatran Hebrew Hiragana Imperial_Aramaic Inherited Inscriptional_Pahlavi Inscriptional_Parthian Javanese Kaithi Kannada Katakana Kayah_Li Kharoshthi Khmer Khojki Khudawadi Lao Latin Lepcha Limbu Linear_A Linear_B Lisu Lycian Lydian Mahajani Makasar Malayalam Mandaic Manichaean Marchen Masaram_Gondi Medefaidrin Meetei_Mayek Mende_Kikakui Meroitic_Cursive Meroitic_Hieroglyphs Miao Modi Mongolian Mro Multani Myanmar Nabataean Nandinagari New_Tai_Lue Newa Nko Nushu Nyiakeng_Puachue_Hmong Ogham Ol_Chiki Old_Hungarian Old_Italic Old_North_Arabian Old_Permic Old_Persian Old_Sogdian Old_South_Arabian Old_Turkic Oriya Osage Osmanya Pahawh_Hmong Palmyrene Pau_Cin_Hau Phags_Pa Phoenician Psalter_Pahlavi Rejang Runic Samaritan Saurashtra Sharada Shavian Siddham SignWriting Sinhala Sogdian Sora_Sompeng Soyombo Sundanese Syloti_Nagri Syriac Tagalog Tagbanwa Tai_Le Tai_Tham Tai_Viet Takri Tamil Tangut Telugu Thaana Thai Tibetan Tifinagh Tirhuta Ugaritic Vai Wancho Warang_Citi Yezidi Yi Zanabazar_Square")
var binaryProperties = set("ASCII ASCII_Hex_Digit Alphabetic Any Assigned Bidi_Control Bidi_Mirrored Case_Ignorable Cased Changes_When_Casefolded Changes_When_Casemapped Changes_When_Lowercased Changes_When_NFKC_Casefolded Changes_When_Titlecased Changes_When_Uppercased Dash Default_Ignorable_Code_Point Deprecated Diacritic Emoji Emoji_Component Emoji_Modifier Emoji_Modifier_Base Emoji_Presentation Extended_Pictographic Extender Grapheme_Base Grapheme_Extend Hex_Digit IDS_Binary_Operator IDS_Trinary_Operator ID_Continue ID_Start Ideographic Join_Control Logical_Order_Exception Lowercase Math Noncharacter_Code_Point Pattern_Syntax Pattern_White_Space Quotation_Mark Radical Regional_Indicator Sentence_Terminal Soft_Dotted Terminal_Punctuation Unified_Ideograph Uppercase Variation_Selector White_Space XID_Continue XID_Start")

func validProperty(s string, unicodeSets bool) bool {
	if strings.Contains(s, "=") {
		p, v, ok := strings.Cut(s, "=")
		return ok && ((p == "General_Category" || p == "gc") && generalCategories[v] || (p == "Script" || p == "sc" || p == "Script_Extensions" || p == "scx") && (scripts[v] || scriptAliases[v]))
	}
	if generalCategories[s] || binaryProperties[s] {
		return true
	}
	return unicodeSets && (s == "Basic_Emoji" || s == "Emoji_Keycap_Sequence" || s == "RGI_Emoji_Flag_Sequence" || s == "RGI_Emoji_Modifier_Sequence" || s == "RGI_Emoji_Tag_Sequence" || s == "RGI_Emoji_ZWJ_Sequence" || s == "RGI_Emoji")
}

var scriptAliases = set("Latn Grek Cyrl Arab Armn Hebr Deva Beng Guru Gujr Orya Taml Telu Knda Mlym Sinh Thai Laoo Tibt Mymr Geor Hang Ethi Cher Cans Ogam Runr Khmr Mong Hira Kana Bopo Han Hani Yiii Ital Goth Dsrt Zyyy Zinh")

func set(words string) map[string]bool {
	m := make(map[string]bool)
	for _, s := range strings.Fields(words) {
		m[s] = true
	}
	return m
}
