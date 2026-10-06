package policy

import "encoding/json"

// Export the resolved catalog, not a second interpretation of its source files.
func AdamicCatalog() []byte {
	type template struct {
		Text    string
		Options map[string]map[string]string
	}
	result := map[string]map[string]template{}
	for rule, messages := range Messages.templates {
		result[rule] = map[string]template{}
		for id, message := range messages {
			result[rule][id] = template{message.text, message.options}
		}
	}
	data, err := json.Marshal(result)
	if err != nil {
		panic(err)
	}
	return data
}
