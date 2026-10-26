package ir

const NodeEnvironmentInheritedPanic = "node:process.Process.env inherited key: "

func IsNodeEnvironmentPrototypeKey(name string) bool {
	for _, key := range NodeEnvironmentPrototypeKeys {
		if name == key {
			return true
		}
	}
	return false
}
