package web_hook

import (
	"fmt"
	"prolific/config"
)

func createComment(comment string) string {
	serverName := config.Get("server", "name")
	serverUrl := config.Get("server", "url")
	return fmt.Sprintf("**[Prolific Bot]**\n\n%s\n\nAssigned Server: [%s](%s)",
		comment, serverName, serverUrl)
}
