package web_hook

import (
	"prolific/config"
	"strings"
)

const (
	BranchElementKey		= "branches"
	OwnerElementKey			= "owners"
	RepositoryElementKey	= "repositories"
)

func isWatched(element string, elementKey string) bool {
	watched := strings.Split(config.Get("watch", elementKey), ";")
	for _, e := range watched {
		if e == element {
			return true
		}
	}
	return false
}