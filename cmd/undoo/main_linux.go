package main

import (
	"bufio"
	"fmt"
	"io"
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
		Setpgid:    true,
	}

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "undoo reexec failed: %s\n", err)
		os.Exit(1)
	}
}

func namespaced() {
	if err := unmount(os.Args[1], os.Args[2]); err != nil {
		fmt.Fprintf(os.Stderr, "undoo unmount %s %s failed: %s\n", os.Args[1], os.Args[2], err)
		os.Exit(2)
	}

	cmd := exec.Command(os.Args[3], os.Args[4:]...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "undoo cmd [%s, %#v] failed: %s\n", cmd.Path, cmd.Args, err)
		os.Exit(3)
	}
}

func main() {
	reexecInNamespace(os.Args[1:]...)
}

func unmount(mountsRoot, layerToKeep string) error {
	mountsFile, err := os.Open("/proc/mounts")
	if err != nil {
		return err
	}

	mountsReader := bufio.NewReader(mountsFile)
	for {
		lineBytes, _, err := mountsReader.ReadLine()
		if err == io.EOF {
			break
		}
		line := string(lineBytes)
		if strings.Contains(line, mountsRoot) && !strings.Contains(line, layerToKeep) {
			mount := strings.Split(line, " ")[1]

			if mount == mountsRoot {
				continue
			}

			err = syscall.Unmount(mount, 0)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
