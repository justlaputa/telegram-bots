BINARY = translate
IMAGENAME = laputa/translate
SRC = $(shell find . -type f -name '*.go' -not -path "./vendor/*")

include secrets

.PHONY: clean docker-image build deploy

$(BINARY): $(SRC)
	GOOS=linux GOARCH=amd64 go build -o $(BINARY)

clean:
	$(RM) $(BINARY)

docker-image: $(BINARY)
	docker build -t $(IMAGENAME) .

build: docker-image

deploy: build
	docker push $(IMAGENAME)
	-ssh vultr docker stop telegram-translate
	-ssh vultr docker rm telegram-translate
	ssh vultr docker pull laputa/translate
	ssh vultr docker run -d --restart=always --name telegram-translate \
		-e GCSE_API_KEY="$(GCSE_API_KEY)" \
		-e GCSE_ID="$(GCSE_ID)" \
		-e BOT_TOKEN="$(BOT_TOKEN)" \
		-e API_KEY="$(API_KEY)" \
		-e SLACK_API_TOKEN=$(SLACK_API_TOKEN) \
		-e WUNDERGROUND_API_KEY=$(WUNDERGROUND_API_KEY) \
		laputa/translate
