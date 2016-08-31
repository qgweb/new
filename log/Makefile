GO := go

build:
	$(GO) build -o lserver logserver.go
	$(GO) build -o lclient logclient.go
clear:
	rm -f lserver lclient
scp:
	scp lclient lserver root@122.225.98.69:/data/qgbin/logserver