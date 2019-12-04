PID_API      = /tmp/api_svc.pid
API_GO_FILES = ./api/cmd/main.go
API_SERVICE      = ./api_svc

PID_CONSUMER      = /tmp/consumer_svc.pid
CONSUMER_GO_FILES = ./consumer/cmd/main.go
CONSUMER_SERVICE      = ./consumer_svc

dependency:
	@go mod download

serve_api: restart_api
	@fswatch -o -e ".*" -i ".*/[^.]*\\.go$$" ./api ./internal | xargs -n1 -I{}  make restart_api || make kill_api

kill_api:
	@test -f $(PID_API) && kill `cat $(PID_API)` >> /dev/null 2>&1 || true

compile_api:
	@echo "Compiling API"
	@go build -race -o $(API_SERVICE) $(API_GO_FILES)

restart_api: kill_api compile_api
	@$(API_SERVICE) & echo $$! > $(PID_API)


serve_consumer: restart_consumer
	@fswatch -o -e ".*" -i ".*/[^.]*\\.go$$" ./consumer ./internal | xargs -n1 -I{}  make restart_consumer || make kill_consumer

kill_consumer:
	@test -f $(PID_CONSUMER) && kill `cat $(PID_CONSUMER)` >> /dev/null 2>&1 || true

compile_consumer:
	@go build -race -o $(CONSUMER_SERVICE) $(CONSUMER_GO_FILES)

restart_consumer: kill_consumer compile_consumer
	@$(CONSUMER_SERVICE) & echo $$! > $(PID_CONSUMER)

containers:
	@docker-compose build
	@docker-compose up

integration-test: containers dependency
	@go test -v ./...

.PHONY: serve_api restart_api kill_api compile_api serve_consumer restart_consumer kill_consumer compile_consumer integration-test containers dependency
