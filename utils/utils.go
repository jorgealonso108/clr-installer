// Copyright © 2018 Intel Corporation
//
// SPDX-License-Identifier: GPL-3.0-only

package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"unsafe"

	"github.com/clearlinux/clr-installer/errors"
)

// MkdirAll similar to go's standard os.MkdirAll() this function creates a directory
// named path, along with any necessary parents but also checks if path exists and
// takes no action if that's true.
func MkdirAll(path string, perm os.FileMode) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	if err := os.MkdirAll(path, perm); err != nil {
		return errors.Errorf("mkdir %s: %v", path, err)
	}

	return nil
}

// CopyFile copies src file to dest
func CopyFile(src string, dest string) error {
	var err error
	destDir := filepath.Dir(dest)

	if _, err = os.Stat(src); err != nil {
		if os.IsNotExist(err) {
			return errors.Errorf("no such file: %s", src)
		}
		return errors.Wrap(err)
	}

	if _, err = os.Stat(destDir); err != nil {
		if os.IsNotExist(err) {
			return errors.Errorf("no such dest directory: %s", destDir)
		}
		return errors.Wrap(err)
	}

	data, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	if err = ioutil.WriteFile(dest, data, 0644); err != nil {
		return err
	}

	return nil
}

// FileExists returns true if the file or directory exists
// else it returns false and the associated error
func FileExists(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return true, err
}

// VerifyRootUser returns an error if we're not running as root
func VerifyRootUser() string {
	// ProgName is the short name of this executable
	progName := path.Base(os.Args[0])

	user, err := user.Current()
	if err != nil {
		return fmt.Sprintf("%s MUST run as 'root' user to install! (user=%s)",
			progName, "UNKNOWN")
	}

	if user.Uid != "0" {
		return fmt.Sprintf("%s MUST run as 'root' user to install! (user=%s)",
			progName, user.Uid)
	}

	return ""
}

// IsClearLinux checks if the current OS is Clear by looking for Swupd
// Mostly used in Go Testing
func IsClearLinux() bool {
	is := false

	if runtime.GOOS == "linux" {
		clearFile := "/usr/bin/swupd"
		if _, err := os.Stat(clearFile); !os.IsNotExist(err) {
			is = true
		}
	}

	return is
}

// IsRoot checks if the current User is root (UID 0)
// Mostly used in Go Testing
func IsRoot() bool {
	is := false

	user, err := user.Current()
	if err == nil {
		if user.Uid == "0" {
			is = true
		}
	}

	return is
}

// StringSliceContains returns true if sl contains str, returns false otherwise
func StringSliceContains(sl []string, str string) bool {
	for _, curr := range sl {
		if curr == str {
			return true
		}
	}
	return false
}

// IntSliceContains returns true if is contains value, returns false otherwise
func IntSliceContains(is []int, value int) bool {
	for _, curr := range is {
		if curr == value {
			return true
		}
	}
	return false
}

// IsCheckCoverage returns true if CHECK_COVERAGE variable is set
func IsCheckCoverage() bool {
	return os.Getenv("CHECK_COVERAGE") != ""
}

// IsStdoutTTY returns true if the stdout is attached to a tty
func IsStdoutTTY() bool {
	var termios syscall.Termios

	fd := os.Stdout.Fd()
	ptr := uintptr(unsafe.Pointer(&termios))
	_, _, err := syscall.Syscall6(syscall.SYS_IOCTL, fd, syscall.TCGETS, ptr, 0, 0, 0)

	return err == 0
}

// ExpandVariables iterates over vars map and replace all the ocorrences of ${var} or
// $var in the str string
func ExpandVariables(vars map[string]string, str string) string {
	// iterate over available variables
	for k, v := range vars {
		// tries to replace both ${var} and $var forms
		for _, rep := range []string{fmt.Sprintf("$%s", k), fmt.Sprintf("${%s}", k)} {
			if strings.Contains(str, rep) {
				return strings.Replace(str, rep, v, -1)
			}
		}
	}

	// if no variables are expanded return the original string
	return str
}
