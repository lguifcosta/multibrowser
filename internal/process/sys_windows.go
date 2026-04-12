//go:build windows

package process

import (
	"os/exec"
)

func setSysProcAttr(cmd *exec.Cmd) {
	// Windows doesn't have Setpgid.
	// We can leave this empty or add Windows-specific attributes if needed.
}
