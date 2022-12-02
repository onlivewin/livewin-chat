VERSION=1.0.0

.PHONY: build
# build
build:
	mkdir -p target/ && go build -ldflags="-s -w" -o ./target/ ./bin/...

.PHONY: noCgoBuild
noCgoBuild:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ./target/ ./bin/...

.PHONY: docker
docker:
	docker build -t wida/livewin-chat-broker:$(VERSION) -f docker/broker_Dockerfile  .
	docker build -t wida/livewin-chat-publisher-api:$(VERSION) -f docker/publisher_Dockerfile .
	docker build -t wida/livewin-chat-discovery:$(VERSION) -f docker/discovery_Dockerfile .
	
.PHONY: docker-build
docker-build: noCgoBuild docker


.PHONY: docker-push
docker-push: 
	docker push  wida/livewin-chat-broker:$(VERSION)
	docker push  wida/livewin-chat-publisher_api:$(VERSION)
	docker push  wida/livewin-chat-discovery:$(VERSION)


.PHONY: allone

allone: docker-build  docker-push

.PHONY: clean
clean:
	rm -rf target
	docker rmi -f wida/livewin-chat-broker:$(VERSION)
	docker rmi -f wida/livewin-chat-publisher_api:$(VERSION)
	docker rmi -f wida/livewin-chat-discovery:$(VERSION)

.PHONY: all
# generate all
all:
	make build
