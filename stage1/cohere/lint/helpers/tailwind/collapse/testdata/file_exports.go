package tailwind

// Oracle-only access to actual loadFile and the theme value it installs.
func AdamicLoadFileValue(path string) (string, bool, error) {
	collector := &stylesheetCollector{
		theme:    NewTheme(),
		resolve:  NodeStylesheetResolver(""),
		visiting: make(map[string]bool),
	}
	if err := collector.loadFile(path); err != nil {
		return "", false, err
	}
	value, found := collector.theme.Get([]string{"--color-probe"})
	return value, found, nil
}
