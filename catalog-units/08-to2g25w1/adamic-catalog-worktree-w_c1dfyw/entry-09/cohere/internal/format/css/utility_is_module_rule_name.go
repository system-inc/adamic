package css

// src/language-css/utilities/is-module-rule-name.js.

var moduleRuleNames = map[string]bool{"import": true, "use": true, "forward": true}

func isModuleRuleName(name string) bool {
	return moduleRuleNames[name]
}
