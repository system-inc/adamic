package javascript

import (
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

// The generated inventory is shared with native and checked against Node itself.
func nodeEnvironmentRuntime() string {
	keys := make([]string, len(ir.NodeEnvironmentPrototypeKeys))
	for index, key := range ir.NodeEnvironmentPrototypeKeys {
		keys[index] = quote(key)
	}
	return "const adamicEnvironmentPrototypeKeys = new Set([" + strings.Join(keys, ",") + "]);\n" +
		`const adamicHostEnvironment = (name, operation, assigned) => {
 if (operation === 'set') return adamicNodeProcess.env[name] = assigned;
 if (operation === 'delete') return delete adamicNodeProcess.env[name];
 const value = adamicNodeProcess.env[name];
 if ((value !== undefined && typeof value !== 'string') || (value === undefined && adamicEnvironmentPrototypeKeys.has(name))) panic(` + quote(ir.NodeEnvironmentInheritedPanic) + ` + name);
 return value;
};
`
}
