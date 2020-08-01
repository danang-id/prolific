package main

import "prolific/application"

func main() {
	application.NewWithName("Prolific").
		RegisterRoutes().
		ListenAndServe()
}
