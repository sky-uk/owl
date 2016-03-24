test : owl alerter_test.go
	go test
	./e2e.sh

owl : alerter.go
	go build

release : owl test
	strip owl

clean :
	rm -f owl

.PHONY: clean test
