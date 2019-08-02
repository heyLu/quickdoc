build: *.go
	go build .

install: build
	install quickdoc /usr/local/bin
	./generate-systemd-service.sh > quickdoc.service
	install -m 0644 quickdoc.service /lib/systemd/system
