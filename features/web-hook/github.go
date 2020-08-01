package web_hook

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"prolific/config"
	"prolific/debug"
	"prolific/features/common"
	"strings"
	"time"
)

const GitHubApiBaseUrl = "https://api.github.com"

type PayloadPullRequestBase struct {
	Ref	string	`json:"ref"`
}

type PayloadPullRequest struct {
	Base   PayloadPullRequestBase `json:"base"`
	Merged bool                   `json:"merged"`
	Number int                    `json:"number"`
}

type PayloadRepositoryOwner struct {
	Login	string	`json:"login"`
}

type PayloadRepository struct {
	Name  string                 `json:"name"`
	Owner PayloadRepositoryOwner `json:"owner"`
}

type GitHubWebHookPayload struct {
	Action      string             `json:"action"`
	PullRequest PayloadPullRequest `json:"pull_request"`
	Repository  PayloadRepository  `json:"repository"`
}

type GitHubPullCreateReviewPayload struct {
	Event			string	`json:"event"`
	Body			string	`json:"body"`
}


func processGitHub(webHookPayload GitHubWebHookPayload) {

	pullRequestNumber := webHookPayload.PullRequest.Number
	owner := webHookPayload.Repository.Owner.Login
	repository := webHookPayload.Repository.Name
	branch := webHookPayload.PullRequest.Base.Ref

	log := common.Log{
		Data: &common.LogData{
			Owner: owner,
			Repository: repository,
			Branch: branch,
			GitHubApiResponses: []map[string]interface{}{},
		},
	}

	if strings.ToUpper(webHookPayload.Action) == "CLOSED" && webHookPayload.PullRequest.Merged {

		var gitHubApiResponse map[string]interface{}

		comment := fmt.Sprintf("This PR #%d has been **approved to [%s] stage** of [%s/%s](https://github.com/%s/%s).\n",
			pullRequestNumber, branch, owner, repository, owner, repository)
		comment += "Prolific Deployment Tool will start the deployment process into the assigned server."
		r, err := createGitHubReview(webHookPayload, comment)
		if err != nil {
			debug.Println(err.Error())
		} else {
			data, err := ioutil.ReadAll(r.Body)
			if err != nil {
				debug.Println(err.Error())
			} else {
				err = json.Unmarshal(data, &gitHubApiResponse)
				if err != nil {
					debug.Println(err.Error())
				} else {
					log.Data.GitHubApiResponses = append(log.Data.GitHubApiResponses, gitHubApiResponse)
				}
			}
		}

		// Deployment Start
		start := time.Now()
		executablesLogs, err := deploy(owner, repository, branch)
		elapsed := time.Since(start)
		end := start.Add(elapsed)
		// Deployment Ended

		log.Success = err == nil
		log.StartedAt = start.Format(time.RFC1123)
		log.EndedAt = end.Format(time.RFC1123)
		log.TimeElapsed = elapsed.String()
		log.Data.ExecutableLogs = executablesLogs

		if err == nil {

			// Deployment Success
			comment = fmt.Sprintf("**SUCCESS**: [%s] stage of [%s/%s](https://github.com/%s/%s) has been deployed. \n\n",
				branch, owner, repository, owner, repository)
			comment += fmt.Sprintf("| _Key_ | _Value_ |\n|---|---|\n")
			comment += fmt.Sprintf("| Start Time | %s |\n", log.StartedAt)
			comment += fmt.Sprintf("| Finish Time | %s |\n", log.EndedAt)
			comment += fmt.Sprintf("| Elapsed Time | %s |\n", log.TimeElapsed)

		} else {

			hideErrorReason := config.GetWithDefault("github", "Hide_Error_Reason", "true") == "true"

			// Deployment Failed
			comment = fmt.Sprintf("**ERROR**: [%s] stage of [%s/%s](https://github.com/%s/%s) failed to be deployed. ",
				branch, owner, repository, owner, repository)
			comment += fmt.Sprintf("Manual review on the assigned server might be required.\n\n")
			if !hideErrorReason {
				comment += fmt.Sprintf("Reason: `%s`\n\n", err.Error())
			}
			comment += fmt.Sprintf("| _Key_ | _Value_ |\n|---|---|\n")
			comment += fmt.Sprintf("| Start Time | %s |\n", log.StartedAt)
			comment += fmt.Sprintf("| Finish Time | %s |\n", log.EndedAt)
			comment += fmt.Sprintf("| Elapsed Time | %s |\n", log.TimeElapsed)
			log.Error = err.Error()

		}

		r, err = createGitHubReview(webHookPayload, comment)
		if err != nil {
			debug.Println(err.Error())
		} else {
			data, err := ioutil.ReadAll(r.Body)
			if err != nil {
				debug.Println(err.Error())
			} else {
				err = json.Unmarshal(data, &gitHubApiResponse)
				if err != nil {
					debug.Println(err.Error())
				} else {
					log.Data.GitHubApiResponses = append(log.Data.GitHubApiResponses, gitHubApiResponse)
				}
			}
		}

		common.WriteLog(common.GitHubLogType, log)

	}

}

func github(writer http.ResponseWriter, request *http.Request) {

	response := common.CreateResponse()

	hubSignature := request.Header.Get("X-Hub-Signature")
	if hubSignature == "" {
		statusCode := http.StatusBadRequest
		response.SetError(common.CreateError(statusCode, "No signature provided."))
		common.SendResponseWithStatusCode(writer, response, statusCode)
		return
	}

	webHookBody, err := ioutil.ReadAll(request.Body)
	if err != nil {
		statusCode := http.StatusInternalServerError
		response.SetError(common.CreateError(statusCode, "Failed to read payload."))
		common.SendResponseWithStatusCode(writer, response, statusCode)
		return
	}

	webHookSecret := config.Get("github", "WebHook_Secret")
	hash := hmac.New(sha1.New, []byte(webHookSecret))
	hash.Write(webHookBody)
	bodySignature := fmt.Sprintf("sha1=%s", hex.EncodeToString(hash.Sum(nil)))

	if hubSignature != bodySignature {
		statusCode := http.StatusBadRequest
		response.SetError(common.CreateError(statusCode, "Invalid signature provided."))
		common.SendResponseWithStatusCode(writer, response, statusCode)
		debug.Printf("Hub Signature: %s", hubSignature)
		debug.Printf("Body Signature: %s", bodySignature)
		return
	}

	var webHookPayload GitHubWebHookPayload
	err = json.Unmarshal(webHookBody, &webHookPayload)
	if err != nil {
		statusCode := http.StatusInternalServerError
		response.SetError(common.CreateError(statusCode, "Failed to parse payload."))
		common.SendResponseWithStatusCode(writer, response, statusCode)
		return
	}

	owner := webHookPayload.Repository.Owner.Login
	repository := webHookPayload.Repository.Name
	branch := webHookPayload.PullRequest.Base.Ref

	if !isWatched(owner, OwnerElementKey) {
		reason := fmt.Sprintf("Owner %s is not being watched.", owner)
		response.SetError(common.CreateError(1001, reason))
		common.SendResponseWithStatusCode(writer, response, http.StatusOK)
		return
	}

	if !isWatched(repository, RepositoryElementKey) {
		reason := fmt.Sprintf("Repository %s is not being watched.", repository)
		response.SetError(common.CreateError(1002, reason))
		common.SendResponseWithStatusCode(writer, response, http.StatusOK)
		return
	}

	if !isWatched(branch, BranchElementKey) {
		reason := fmt.Sprintf("Branch %s is not being watched.", branch)
		response.SetError(common.CreateError(1003, reason))
		common.SendResponseWithStatusCode(writer, response, http.StatusOK)
		return
	}

	go processGitHub(webHookPayload)

	response.Message = "Event recorded."
	common.SendResponse(writer, response)

}
