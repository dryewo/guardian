package main_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Undoo", func() {
	var (
		depotPath, mnt1, mnt2 string
	)

	It("executes the command line passed as args", func() {
		cmd := exec.Command(undooBinPath, "mountsRoot", "keep-id", "echo", "yabadabadoo")
		Expect(cmd.CombinedOutput()).To(ContainSubstring("yabadabadoo"))
	})

	It("creates a new mount namespace", func() {
		parentNsCmd := exec.Command("readlink", "/proc/self/ns/mnt")
		parentNsBytes, err := parentNsCmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred())

		cmd := exec.Command(undooBinPath, "mountsRoot", "keep-id", "readlink", "/proc/self/ns/mnt")
		childNsBytes, err := cmd.CombinedOutput()
		Expect(err).NotTo(HaveOccurred())

		Expect(childNsBytes).NotTo(Equal(parentNsBytes))
	})

	It("forwards any error message", func() {
		cmd := exec.Command(undooBinPath, "mountsRoot", "keep-id", "ls", "scooobydoo")
		out, err := cmd.CombinedOutput()
		Expect(err).To(HaveOccurred())
		Expect(string(out)).To(ContainSubstring("No such file or directory"))
	})

	Context("when there are mounts under depo path in the parent mount namespace", func() {
		BeforeEach(func() {
			var err error
			depotPath, err = ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())
			depotPath = filepath.Join(depotPath, "aufs")
			Expect(os.MkdirAll(depotPath, 0644)).To(Succeed())

			Expect(syscall.Mount(depotPath, depotPath, "", syscall.MS_BIND, "")).To(Succeed())

			mnt1 = filepath.Join(depotPath, "mnt1")
			Expect(os.MkdirAll(mnt1, 0644)).To(Succeed())
			syscall.Mount("tmpfs", mnt1, "tmpfs", 0, "")

			mnt2 = filepath.Join(depotPath, "mnt2")
			Expect(os.MkdirAll(mnt2, 0644)).To(Succeed())
			syscall.Mount("tmpfs", mnt2, "tmpfs", 0, "")
		})

		AfterEach(func() {
			syscall.Unmount(mnt1, 0)
			syscall.Unmount(mnt2, 0)
			syscall.Unmount(depotPath, 0)
			os.RemoveAll(depotPath)
		})

		It("unmounts all unneeded mounts from the child mount namespace", func() {
			cmd := exec.Command(undooBinPath, depotPath, "mnt2", "cat", "/proc/mounts")
			mountsBytes, err := cmd.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
			mounts := string(mountsBytes)

			Expect(mounts).NotTo(ContainSubstring("mnt1"))
			Expect(mounts).To(ContainSubstring("mnt2"))

			mountsBytes, err = exec.Command("cat", "/proc/self/mounts").CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
			mounts = string(mountsBytes)

			Expect(mounts).To(ContainSubstring("mnt1"))
			Expect(mounts).To(ContainSubstring("mnt2"))
		})
	})
})
