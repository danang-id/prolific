package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	PackageName = "prolific"
)

func main() {
	if len(os.Args) <= 1 {
		exit(1001, "No command provided.")
	}

	switch strings.ToLower(os.Args[1]) {
	case "build":
		build()
		break
	case "deploy":
		deploy()
		break
	default:
		exit(1002, "Command " + os.Args[1] + " is unknown.")
	}
}

func build() {
	executablePath, err := exec.LookPath("go")
	if err != nil {
		exit(2002, err.Error())
	}

	command := &exec.Cmd{
		Path:         executablePath,
		Args:         append([]string{ executablePath },"build", "-o", getBinPath(), PackageName),
		Env:          os.Environ(),
		Dir:          getWorkDir(),
	}

	output, err := command.Output()
	if err != nil {
		exit(2003, err.Error())
	}

	if string(output) != "" {
		fmt.Println(output)
	} else {
		fmt.Println(PackageName, "compiled successfully.")
	}
}

func deploy() {
	switch runtime.GOOS {
	case "linux":
		deployLinux()
		break
	default:
		err := errors.New("deployment on " + runtime.GOOS + " platform is not supported at the moment")
		exit(1003, err.Error())
	}
}

func deployLinux() {
	serviceTemplateFile, err := os.Open(getServiceTemplatePath())
	if err != nil {
		exit(3001, err.Error())
	}
	defer func() {
		if err := serviceTemplateFile.Close(); err != nil {
			exit(3101, err.Error())
		}
	}()

	scanner := bufio.NewScanner(serviceTemplateFile)
	scanner.Split(bufio.ScanLines)
	var service []string

	for scanner.Scan() {
		service = append(service, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		exit(3002, err.Error())
	}

	for index, line := range service {
		if strings.HasPrefix(line, "Description=") {
			service[index] = "Description=" + PackageName
		}
		if strings.HasPrefix(line, "ConditionPathExists=") {
			service[index] = "ConditionPathExists=" + getBinPath()
		}
		if strings.HasPrefix(line, "WorkingDirectory=") {
			service[index] = "WorkingDirectory=" + getWorkDir()
		}
		if strings.HasPrefix(line, "ExecStart=") {
			service[index] = "ExecStart=" + getBinPath()
		}
		if strings.HasPrefix(line, "ExecStartPre=/bin/mkdir -p /var/log/") {
			service[index] = "ExecStartPre=/bin/mkdir -p /var/log/" + PackageName
		}
		if strings.HasPrefix(line, "ExecStartPre=/bin/chown syslog:adm /var/log/") {
			service[index] = "ExecStartPre=/bin/chown syslog:adm /var/log/" + PackageName
		}
		if strings.HasPrefix(line, "SyslogIdentifier=") {
			service[index] = "SyslogIdentifier=" + PackageName
		}
	}

	serviceFilePath := filepath.Join("etc", "systemd", "system", fmt.Sprintf("%s.service", PackageName))
	if err != nil {
		exit(3003, err.Error())
	}

	serviceFile, err := os.OpenFile(serviceFilePath, os.O_CREATE | os.O_RDWR, 0755)
	if err != nil {
		exit(3004, err.Error())
	}
	defer func() {
		if err := serviceFile.Close(); err != nil {
			exit(3102, err.Error())
		}
	}()

	_, err = fmt.Fprint(serviceFile, strings.Join(service, "\n"))
	if err != nil {
		exit(3005, err.Error())
	}

	err = serviceFile.Sync()
	if err != nil {
		exit(3006, err.Error())
	}

	executablePath, err := exec.LookPath("systemctl")
	if err != nil {
		exit(3007, err.Error())
	}

	command := &exec.Cmd{
		Path: executablePath,
		Args: append([]string{executablePath},"enable", PackageName + ".service"),
		Env:  os.Environ(),
		Dir:  getWorkDir(),
	}

	output, err := command.Output()
	if err != nil {
		exit(3008, err.Error())
	}

	if string(output) != "" {
		fmt.Println(output)
	}

	command = &exec.Cmd{
		Path: executablePath,
		Args: append([]string{executablePath},"restart", PackageName + ".service"),
		Env:  os.Environ(),
		Dir:  getWorkDir(),
	}

	output, err = command.Output()
	if err != nil {
		exit(3009, err.Error())
	}

	if string(output) != "" {
		fmt.Println(output)
	}

}

func exit(code int, message string) {
	fmt.Println(message)
	os.Exit(code)
}

func getBinPath() string {
	return filepath.Join(getWorkDir(), "bin", "release", fmt.Sprintf("%s_release", PackageName))
}

func getServiceTemplatePath() string {
	return filepath.Join(getWorkDir(), "scripts", "template.service")
}

func getWorkDir() string {
	workDir, err := os.Getwd()
	if err != nil {
		exit(2001, err.Error())
	}
	return workDir
}