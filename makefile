DESTDIR = /usr/local

run:
	go run main.go
clean:
	rm -rf bin
build:
	go build -ldflags "-s -w" -o bin/Amnesia
install: build
	mkdir -p ${DESTDIR}/bin
	cp -f bin/Amnesia ${DESTDIR}/bin/Amnesia
	chmod 755 ${DESTDIR}/bin/Amnesia
uninstall:
	rm -f ${DESTDIR}/bin/Amnesia

all: build