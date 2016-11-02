package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/docker/docker/pkg/reexec"
)

func init() {
	reexec.Register("namespaced", namespaced)

	if reexec.Init() {
		os.Exit(0)
	}
}

func reexecInNamespace(args ...string) {
	reexecArgs := append([]string{"namespaced"}, args...)
	cmd := reexec.Command(reexecArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWNS,
	}

	if err := cmd.Run(); err != nil {
		fmt.Printf("undoo exec failed: %s\n", err)
		os.Exit(1)
	}
}

func namespaced() {
	unmount(os.Args[1], os.Args[2])

	cmd := exec.Command(os.Args[3], os.Args[4:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}

func main() {
	reexecInNamespace(os.Args[1:]...)
}

func unmount(mountsDir, layetToKeep string) {
	cmd := exec.Command("cat", "/proc/mounts")
	mounts, _ := cmd.CombinedOutput()
	for _, mount := range strings.Split(string(mounts), "\n") {
		if !strings.Contains(mount, mountsDir) || strings.Contains(mount, layetToKeep) {
			continue
		}
		mount = mount[strings.Index(mount, mountsDir):]
		mount = strings.Split(mount, " ")[0]

		syscall.Unmount(mount, 0)
	}
}
