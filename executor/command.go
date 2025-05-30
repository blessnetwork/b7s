package executor

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/blessnetwork/b7s/models/execute"
)

// createCmd will create the command to be executed, prepare working directory, environment, standard input and all else.
func (e *Executor) createCmd(paths requestPaths, req execute.Request) *exec.Cmd {

	// Prepare command to be executed.
	exePath := filepath.Join(e.cfg.RuntimeDir, e.cfg.ExecutableName)

	cfg := req.Config.Runtime
	cfg.Input = paths.input
	cfg.FSRoot = paths.fsRoot
	cfg.DriversRootPath = e.cfg.DriversRootPath

	// Prepare CLI arguments.
	// Append the input argument first first.
	var args []string
	args = append(args, cfg.Input)

	// Append the arguments for the runtime.
	runtimeFlags := runtimeFlags(cfg, req.Config.Permissions)
	args = append(args, runtimeFlags...)

	// Separate runtime arguments from the function arguments.
	args = append(args, "--")
	args = append(args, req.Arguments...)

	// Function arguments.
	cmd := exec.Command(exePath, args...)
	cmd.Dir = paths.workdir

	// Setup stdin of the command.
	var stdin io.Reader
	if req.Config.Stdin != nil {
		stdin = strings.NewReader(*req.Config.Stdin)
	}
	cmd.Stdin = stdin

	// Setup environment.
	// First, pass through our environment variables.
	cmd.Env = os.Environ()

	// Second, set the variables set in the execution request.
	names := make([]string, 0, len(req.Config.Environment))
	for _, env := range req.Config.Environment {
		e := fmt.Sprintf("%s=%s", env.Name, env.Value)
		cmd.Env = append(cmd.Env, e)

		names = append(names, env.Name)
	}

	// Third and final - set the `BLS_LIST_VARS` variable with
	// the list of names of the variables from the execution request.
	blsList := strings.Join(names, ";")
	blsEnv := fmt.Sprintf("%s=%s", blsListEnvName, blsList)
	cmd.Env = append(cmd.Env, blsEnv)

	return cmd
}
