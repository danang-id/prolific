package common

import (
	"os"
	"os/exec"
)

type Executable struct {
	Name 				string	`json:"name"`
	Path				string	`json:"path"`
	WorkingDirectory	string	`json:"working_directory"`
}

func (executable *Executable) Exists() bool {
	executablePath, err := exec.LookPath(executable.Name)
	if err != nil || executablePath == "" {
		return false
	}
	return true
}

func (executable *Executable) Run(args ...string) (string, error) {
	command := &exec.Cmd{
		Path:         executable.Path,
		Args:         append([]string{ executable.Path }, args...),
		Env:          os.Environ(),
		Dir:          executable.WorkingDirectory,
	}
	output, err := command.Output()
	if err != nil {
		return string(output), err
	}
	return string(output), nil
}

func NewExecutable(name string, workingDirectory string) (*Executable, error) {
	if workingDirectory == "" {
		workDir, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		workingDirectory = workDir
	}
	executablePath, err := exec.LookPath(name)
	if err != nil {
		return nil, err
	}
	return &Executable{Name: name, Path: executablePath, WorkingDirectory: workingDirectory}, nil
}

type ExecutableLog struct {
	Name	string	`json:"name"`
	Args	string	`json:"args"`
	WorkDir	string	`json:"work_dir"`
	Output	string	`json:"output"`
	Error	string	`json:"error,omitempty"`
}