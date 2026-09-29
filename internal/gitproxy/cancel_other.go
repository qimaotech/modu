//go:build !unix

package gitproxy

import "os/exec"

func configureCancellation(cmd *exec.Cmd) {}
