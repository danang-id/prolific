package web_hook

import (
	"encoding/json"
	"fmt"
	"net/http"
	"prolific/config"
	"prolific/debug"
	"strings"
	"time"
)

// createGitHubReview: ["POST /repos/{owner}/{repo}/pulls/{pull_number}/reviews"]
func createGitHubReview(webHookPayload GitHubWebHookPayload, comment string)( *http.Response, error) {

	pullRequestNumber := webHookPayload.PullRequest.Number
	owner := webHookPayload.Repository.Owner.Login
	repository := webHookPayload.Repository.Name

	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d/reviews",
		GitHubApiBaseUrl,
		owner,
		repository,
		pullRequestNumber)
	gitHubPersonalAccessToken := config.Get("github", "Personal_Access_Token")

	reviewPayload := GitHubPullCreateReviewPayload{
		Event: "COMMENT",
		Body:  createComment(comment),
	}

	body, err := json.Marshal(reviewPayload)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout:       15 * time.Second,
	}

	request, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/vnd.github.v3+json")
	request.Header.Set("Authorization", fmt.Sprintf("Token %s", gitHubPersonalAccessToken))

	debug.Printf("Created Github Review on PR #%d [%s/%s]\n", pullRequestNumber, owner, repository)

	return client.Do(request)
}

