#!/bin/bash

env GOOS=linux GOARCH=amd64 go build
ssh do pkill cats
scp cats do:~/bots

ssh do "PORT=8100 BOT_TOKEN=$BOT_TOKEN nohup ~/bots/cats >> ~/bots/cats.log 2>&1 &"
