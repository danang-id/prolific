build:
	@go build -o scripts/scripts scripts/main.go
	@./scripts/scripts build

deploy:
	@go build -o scripts/scripts scripts/main.go
	@./scripts/scripts deploy