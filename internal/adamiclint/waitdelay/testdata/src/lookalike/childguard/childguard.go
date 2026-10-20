// A package named childguard at another path: its Run and CombinedOutput are not the ones the rule knows.
package childguard

import "os/exec"

type Options struct{}

func Run(cmd *exec.Cmd, options Options) error { return cmd.Run() }

func CombinedOutput(cmd *exec.Cmd, options Options) ([]byte, error) { return nil, cmd.Run() }
