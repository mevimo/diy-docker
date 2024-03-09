package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func main() {
	debugLog("Sysargs: %v", os.Args)
	command := os.Args[3]

	tmpDir, err := os.MkdirTemp("", "sandbox")
	if err != nil {
		debugLog("Error making tmpdir: %s", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	src, err := os.Open(command)
	check(err)
	srcinfo, err := src.Stat()
	debugLog("srcinfo stat: %v", srcinfo.Mode())

	err = os.MkdirAll(tmpDir+filepath.Dir(command), 0700)
	check(err)
	dest, err := os.Create(tmpDir + command)
	check(err)
	// defer dest.Close()
	debugLog("dest create OK")
	_, err = io.Copy(dest, src)
	check(err)
	debugLog("copy OK")
	check(dest.Chmod(srcinfo.Mode()))
	debugLog("perms OK")

	destinfo, err := dest.Stat()
	debugLog("destinfo stat: %v", destinfo.Mode())

	dest.Close()

	check(chroot(tmpDir))
	debugLog("Chroot OK")
	check(os.Chdir("/"))
	check(os.Mkdir("/dev", 0755))
	f, err := os.Create("/dev/null")
	check(err)
	defer f.Close()
	debugLog("dev/null OK")
	devinfo, err := f.Stat()
	check(err)
	debugLog("dev/null stat: %v", devinfo.Mode())

	dest, _ = os.Open(command)
	destinfo, _ = dest.Stat()
	debugLog("destinfo stat: %v", destinfo.Size())

	args := os.Args[4:len(os.Args)]
	debugLog("Running Command: %s %s", command, args)
	cmd := exec.Command(command, args...)
	// cmd := exec.Command("echo fuckyou")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}
	_ = cmd.Run()
	// out, err := cmd.CombinedOutput()
	// check(err)
	// debugLog(string(out))

	os.Exit(cmd.ProcessState.ExitCode())
	// if runErr != nil {
	// 	exitError, ok := runErr.(*exec.ExitError)
	// 	if ok {
	// 		os.Exit(exitError.ExitCode())
	// 	}
	// }
}

func chroot(jailDir string) error {
	if err := syscall.Chroot(jailDir); err != nil {
		debugLog("Error chrooting: %s", err)
		return err
	}
	if err := os.Chdir("/"); err != nil {
		debugLog("Error chdir: %s", err)
		return err
	}
	return nil
}

func debugLog(format string, a ...interface{}) {
	if os.Getenv("DEBUG") != "" {
		fmt.Printf("[DEBUG] "+format+"\n", a...)
	}
}

func check(err error) {
	if err != nil {
		debugLog("Error: %s", err)
		os.Exit(1)
	}
}
