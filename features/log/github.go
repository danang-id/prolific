package log

import (
	"net/http"
	"prolific/config"
	"prolific/features/common"
	"strings"
)

func github(writer http.ResponseWriter, request *http.Request) {

	response := common.CreateResponse()

	authorizationHeader := request.Header.Get("Authorization")
	if authorizationHeader == "" {
		statusCode := http.StatusUnauthorized
		response.SetError(common.CreateError(statusCode, "No authorization provided."))
		common.SendResponseWithStatusCode(writer, response, statusCode)
		return
	}

	authorization := strings.Split(authorizationHeader, " ")
	if len(authorization) != 2 {
		statusCode := http.StatusUnauthorized
		response.SetError(common.CreateError(statusCode, "Authorization format invalid."))
		common.SendResponseWithStatusCode(writer, response, statusCode)
		return
	}

	if strings.ToLower(authorization[0]) != "token" {
		statusCode := http.StatusUnauthorized
		response.SetError(common.CreateError(statusCode, "Authorization type invalid."))
		common.SendResponseWithStatusCode(writer, response, statusCode)
		return
	}


	webHookSecret := config.Get("github", "Log_Access_Token")
	if authorization[1] != webHookSecret {
		statusCode := http.StatusUnauthorized
		response.SetError(common.CreateError(statusCode, "Authorization token invalid."))
		common.SendResponseWithStatusCode(writer, response, statusCode)
		return
	}

	response.Data = common.ReadLogs(common.GitHubLogType)
	common.SendResponse(writer, response)

}

