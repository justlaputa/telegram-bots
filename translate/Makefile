BINARY = translate
IMAGENAME = laputa/translate

include secrets

.PHONY: clean docker-image build deploy

$(BINARY):
	GOOS=linux GOARCH=amd64 go build -o $(BINARY)

clean:
	$(RM) $(BINARY)

docker-image: $(BINARY)
	docker build -t $(IMAGENAME) .

build: docker-image

deploy: build
	docker push $(IMAGENAME)
	hyper stop $(shell hyper ps -f "name=telegram-translate" -qa)
	hyper rm $(shell hyper ps -f "name=telegram-translate" -qa)
	hyper rmi laputa/translate:latest
	hyper run --size=s1 -d --restart=always --name telegram-translate \
		-e PINTEREST_TOKEN="$(PINTEREST_TOKEN)" \
		-e BOT_TOKEN="$(BOT_TOKEN)" \
		-e API_KEY="$(API_KEY)" \
		laputa/translate
