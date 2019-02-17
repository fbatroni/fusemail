#!/bin/sh

nohup go run main.go filecreator.go >./fileserver.log 2>&1 &
