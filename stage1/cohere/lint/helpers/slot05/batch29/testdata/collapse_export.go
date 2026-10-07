package tailwind

import (
	"strconv"
)

type AdamicHandler struct {
	Present bool   `json:"present"`
	OK      bool   `json:"ok"`
	Text    string `json:"text"`
}

func AdamicHandle(kind, suffix string, value *ParsedValue) AdamicHandler {
	h := bareValueHandler(BareValueKind(kind), suffix)
	if h == nil {
		return AdamicHandler{}
	}
	text, ok := h(value)
	return AdamicHandler{true, ok, text}
}
func AdamicFont(value string) bool { return isFontStretchPercentage(value) }
func AdamicParse(value string) (string, bool) {
	v, e := strconv.ParseFloat(value, 64)
	return strconv.FormatFloat(v, 'g', -1, 64), e == nil
}
func AdamicFunctional(theme []string, negative, defPresent bool, def, kind, suffix string, names []string) *FunctionalUtilityDescription {
	statics := []FrameworkStaticValue{}
	for _, name := range names {
		statics = append(statics, FrameworkStaticValue{Name: name})
	}
	return (FrameworkFunctionalUtility{ThemeKeys: theme, SupportsNegative: negative, DefaultValuePresent: defPresent, DefaultValue: def, BareValue: BareValueKind(kind), BareValueSuffix: suffix, StaticValues: statics}).Description()
}
func AdamicMulti(theme []string, negative, fractions, defPresent bool, def, kind, suffix string) *FunctionalUtilityDescription {
	return (FrameworkMultiDeclarationUtility{ThemeKeys: theme, SupportsNegative: negative, SupportsFractions: fractions, DefaultValuePresent: defPresent, DefaultValue: def, BareValue: BareValueKind(kind), BareValueSuffix: suffix}).Description()
}
