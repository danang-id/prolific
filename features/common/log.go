package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"prolific/debug"
)

var GitHubLogType LogType = "GitHub"
var logDirPath = filepath.Join("logs")

type Log struct {
	Success		bool   `json:"success"`
	StartedAt	string   `json:"started_at"`
	EndedAt		string `json:"ended_at"`
	TimeElapsed	string `json:"time_elapsed"`
	Error		string   `json:"error,omitempty"`
	Data		*LogData  `json:"data,omitempty"`
}

type Logs []Log

type LogData struct {
	Owner				string                    `json:"owner"`
	Repository			string                   `json:"repository"`
	Branch				string                   `json:"branch"`
	GitHubApiResponses	[]map[string]interface{} `json:"github_api_responses"`
	ExecutableLogs		[]ExecutableLog          `json:"executable_logs"`
}

type LogType string

func ReadLogs(logType LogType) Logs {
	logFilePath := filepath.Join(logDirPath, fmt.Sprintf("%s.json", logType))
	if _, err := os.Stat(logDirPath); os.IsNotExist(err) {
		err = os.MkdirAll(logDirPath, os.ModePerm)
		if err != nil {
			debug.Println(err.Error())
		}
	}
	if _, err := os.Stat(logFilePath); err != nil {
		debug.Println(err.Error())
		return Logs{}
	}
	content, err := ioutil.ReadFile(logFilePath)
	if err != nil {
		debug.Println(err.Error())
		return Logs{}
	}
	var logs Logs
	err = json.Unmarshal(content, &logs)
	if err != nil {
		debug.Println(err.Error())
		return Logs{}
	}
	return logs
}

func WriteLog(logType LogType, log Log) {
	logFilePath := filepath.Join(logDirPath, fmt.Sprintf("%s.json", logType))
	logs := ReadLogs(logType)
	logs = append(logs, log)
	content, err := json.MarshalIndent(logs, "", "\t")
	if err != nil {
		debug.Println(err.Error())
	}
	err = ioutil.WriteFile(logFilePath, content, 0755)
	if err != nil {
		debug.Println(err.Error())
	}
}