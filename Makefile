install:
	export CGO_ENABLED=1
	rm -f $(GOPATH)/bin/booking-svc
	go build -o $(GOPATH)/bin/booking-svc .

run: install
	booking-svc service run