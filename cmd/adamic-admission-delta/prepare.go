package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Tools come from cloud/setup.sh's environment. Checkout-local declarations
// must instead be installed from this checkout's own package lock.
func prepareCheckout(tree string) error {
	api := filepath.Join(tree, "stage3", "api")
	if _, err := os.Stat(filepath.Join(api, "package-lock.json")); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	install := execute(api, 90*time.Second, "npm", "ci", "--ignore-scripts")
	if install.Exit != 0 || install.Error != "" {
		return fmt.Errorf("provision stage3/api: %s %s", install.Error, install.Stderr)
	}
	return verifyNodeTypes(api)
}
func verifyNodeTypes(api string) error {
	path := filepath.Join(api, "node_modules", "@types", "node", "package.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var pin struct {
		Name    string
		Version string
	}
	if err = json.Unmarshal(data, &pin); err != nil {
		return err
	}
	if pin.Name != "@types/node" || pin.Version != "25.3.3" {
		return fmt.Errorf("want @types/node 25.3.3, found %s %s", pin.Name, pin.Version)
	}
	_, err = os.Stat(filepath.Join(filepath.Dir(path), "index.d.ts"))
	return err
}
