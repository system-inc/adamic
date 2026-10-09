package javascript

import "github.com/system-inc/cohere/internal/format/doc"

// utilities/create-group-id-mapper.js.

// createGroupIdMapper is upstream's createGroupIdMapper: one group id per node, minted on first ask.
// Upstream keeps a module-level WeakMap per mapper; Go has no weak map, so the ids live on the
// format's settings, keyed on the description and then the node, and the returned function takes
// options to reach them.
func createGroupIdMapper(description string) func(options *Options, node Node) *doc.GroupID {
	return func(options *Options, node Node) *doc.GroupID {
		formatSettings := settingsOf(options)
		if formatSettings.groupIDs == nil {
			formatSettings.groupIDs = map[string]map[Node]*doc.GroupID{}
		}
		groupIds := formatSettings.groupIDs[description]
		if groupIds == nil {
			groupIds = map[Node]*doc.GroupID{}
			formatSettings.groupIDs[description] = groupIds
		}
		if groupID, present := groupIds[node]; present {
			return groupID
		}
		groupID := newGroupID(description)
		groupIds[node] = groupID
		return groupID
	}
}
