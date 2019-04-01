test : setup owl alerter_test.go
	go test
	./e2e.sh

owl : alerter.go
	go build

release : owl test
	strip owl

clean :
	rm -f owl

setup:
	@echo "setup"
	@go get github.com/golang/dep/cmd/dep
	dep ensure

.PHONY: clean test setup
