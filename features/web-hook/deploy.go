package web_hook

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"prolific/config"
	"prolific/debug"
	"prolific/features/common"
	"strings"
)

func deploy(owner string, repository string, branch string) ([]common.ExecutableLog, error) {

	debug.Printf("Deployment Started for Branch %s [%s/%s]\n", branch, owner, repository)

	var executableLogs []common.ExecutableLog

	rootPath := filepath.Join(config.Get("Prolific", "Root_Path"))
	if _, err := os.Stat(rootPath); os.IsNotExist(err) || err != nil {
		if err != nil {
			err = errors.New("root path " + rootPath + " does not exist")
			debug.Printf("Deployment Finished with Error (Reason: %s)\n", err.Error())
			return executableLogs, err
		}
	}

	repoPath := filepath.Join(rootPath, branch, repository)
	if _, err := os.Stat(repoPath); os.IsNotExist(err) || err != nil {
		if err != nil {
			err = errors.New("repository path " + repoPath + " does not exist")
			debug.Printf("Deployment Finished with Error (Reason: %s)\n", err.Error())
			return executableLogs, err
		}
	}

	suExec, err := common.NewExecutable("su", repoPath)
	if err != nil {
		debug.Printf("Deployment Finished with Error (Reason: %s)\n", err.Error())
		return executableLogs, err
	}

	makeExec, err := common.NewExecutable("make", repoPath)
	if err != nil {
		debug.Printf("Deployment Finished with Error (Reason: %s)\n", err.Error())
		return executableLogs, err
	}

	err = checkDependencies(suExec, makeExec)
	if err != nil {
		debug.Printf("Deployment Finished with Error (Reason: %s)\n", err.Error())
		return executableLogs, err
	}

	user := config.GetWithDefault("Prolific", "User", "root")

	commandString := fmt.Sprintf("'git checkout %s'", branch)
	executableLog, err := execute(suExec, user, "-c", commandString)
	executableLogs = append(executableLogs, executableLog)
	if err != nil {
		debug.Printf("Deployment Finished with Error (Reason: %s)\n", err.Error())
		return executableLogs, err
	}

	commandString = fmt.Sprintf("'git pull'")
	executableLog, err = execute(suExec, user, "-c", commandString)
	executableLogs = append(executableLogs, executableLog)
	if err != nil {
		debug.Printf("Deployment Finished with Error (Reason: %s)\n", err.Error())
		return executableLogs, err
	}

	commandString = fmt.Sprintf("'make'")
	executableLog, err = execute(suExec, user, "-c", commandString)
	executableLogs = append(executableLogs, executableLog)
	if err != nil {
		debug.Printf("Deployment Finished with Error (Reason: %s)\n", err.Error())
		return executableLogs, err
	}

	executableLog, err = execute(makeExec, "deploy")
	executableLogs = append(executableLogs, executableLog)
	if err != nil {
		debug.Printf("Deployment Finished with Error (Reason: %s)\n", err.Error())
		return executableLogs, err
	}

	debug.Println("Deployment Finished Successfully")
	return executableLogs, nil

}

func checkDependencies(executables ...*common.Executable) error {
	for _, executable := range executables {
		if !executable.Exists() {
			return errors.New("dependency " + executable.Name + " not available")
		}
	}
	return nil
}

func execute(executable *common.Executable, args ...string) (common.ExecutableLog, error) {
	output, err := executable.Run(args...)
	message := ""
	if err != nil {
		message = err.Error()
	}
	executableLog := common.ExecutableLog{
		Name:		executable.Path,
		Args:		strings.Join(append([]string{ executable.Path }, args...), " "),
		WorkDir:	executable.WorkingDirectory,
		Output:		output,
		Error:		message,
	}
	return executableLog, err
}