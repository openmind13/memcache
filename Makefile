build:
	go build -o main *.go && ./main

.DEFAULT_GOAL := build