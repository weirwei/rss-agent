VERSION ?= $(shell cat ./VERSION)
docker-image-build:
	docker build -t rss-agent:${VERSION} --build-arg APP_NAME=rss-agent .

docker-run:
	docker run -p 8080:8080 -d rss-agent:${VERSION}