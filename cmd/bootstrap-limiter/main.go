//go:build linux
// +build linux

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/fatih/color"
	"github.com/spf13/pflag"

	"github.com/blessnetwork/b7s/executor/limits"
)

const (
	success = 0
	failure = 1

	// SUDO_USER - Set to the login name of the user who invoked sudo.
	// https://man7.org/linux/man-pages/man8/sudo.8.html
	sudoUserEnv = "SUDO_USER"

	rootCgroupFilePermissions = 0664
)

var (
	requiredFileNames = []string{
		"cgroup.procs",
		"cgroup.subtree_control",
	}
)

func main() {
	os.Exit(run())
}

func run() int {

	var (
		flagForce bool
	)

	pflag.BoolVarP(&flagForce, "force", "f", false, "set resource limits without asking for confirmation")

	pflag.Parse()

	log.Default().SetFlags(0)

	// We'll only target Linux, though potentially cgroups may exist on some other systems.
	if runtime.GOOS != "linux" {
		log.Printf("OS not supported")
		return failure
	}

	// If `--force` was not set - ask for user confirmation before continuing.
	if !flagForce && !haveConsent() {
		return failure
	}

	mountpoint := limits.DefaultMountpoint
	cgroupName := limits.DefaultCgroup

	currentUser, err := user.Current()
	if err != nil {
		log.Printf("could not get current user: %s", err)
		return failure
	}

	log.Printf("running as user: %v (uid: %v)", currentUser.Username, currentUser.Uid)

	// The owner-to-be of the cgroup directory.
	runningUser := os.Getenv(sudoUserEnv)
	if runningUser == "" {
		log.Printf("could not get user running sudo - trying with current user")
		runningUser = currentUser.Username
	} else {
		log.Printf("ownership for cgroup will be assigned to user '%v'", runningUser)
	}

	// Create directory on the default cgroup mountpoint.
	target := filepath.Join(mountpoint, cgroupName)
	err = os.MkdirAll(target, 0755)
	if err != nil {
		log.Printf("could not create directory: '%v': %s", target, err)
		return failure
	}

	log.Printf("cgroup %v created", cgroupName)

	runningUserInfo, err := user.Lookup(runningUser)
	if err != nil {
		log.Printf("could not lookup user ID: %s", err)
		return failure
	}

	// Convert ID to integer.
	id, err := strconv.ParseInt(runningUserInfo.Uid, 10, 32)
	if err != nil {
		log.Printf("could not parse user UID: %s", err)
		return failure
	}

	// Chown directory to be owned by the original user running sudo.
	err = chownRecursive(target, int(id), -1)
	if err != nil {
		log.Printf("could not set owner for the cgroup: %s", err)
		return failure
	}

	log.Printf("access to cgroup %v granted to user '%v'", cgroupName, runningUser)

	var requiredFiles []string
	for _, name := range requiredFileNames {
		path := filepath.Join(mountpoint, name)
		requiredFiles = append(requiredFiles, path)
	}

	fileGroup := -1
	// Set permissions for required files.
	for _, file := range requiredFiles {

		// Make sure the group has write permissions.
		err = os.Chmod(file, rootCgroupFilePermissions)
		if err != nil {
			log.Printf("could not set permissions to group writable for %v", file)
			return failure
		}

		log.Printf("set file permissions for %v to be group writable", file)

		// Get groupID if not already.
		if fileGroup == -1 {
			info, err := os.Stat(file)
			if err != nil {
				log.Printf("could not stat file: %v: %s", file, err)
				continue
			}

			stat, ok := info.Sys().(*syscall.Stat_t)
			if !ok {
				log.Printf("unexpected format for stat")
				continue
			}

			fileGroup = int(stat.Gid)
		}
	}

	group, _ := user.LookupGroupId(fmt.Sprint(fileGroup))

	// If we don't know which group it is, print a generic message.
	if group == nil {
		fileList := strings.Join(requiredFiles, ", ")
		color.Red("please ensure user '%v' is a member of a group with write access to files %s", runningUser, fileList)
		return success
	}

	// Get group name for gid.
	color.Red("please ensure user '%v' is a member of '%v' group in order to be able to set resource limits for Bless Functions\n", runningUser, group.Name)
	color.Red("e.g. by running sudo usermod -a -G <group> <user>")

	return success
}

func haveConsent() bool {

	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("This command will attempt to setup the system to set resource limits for Bless Functions. Would you like to continue? [yes/no]: ")

	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))

	return (answer == "yes" || answer == "y")
}

func chownRecursive(path string, uid int, gid int) error {
	return filepath.Walk(path, func(name string, _ os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		return os.Chown(name, uid, gid)
	})
}
