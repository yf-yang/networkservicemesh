.PHONY: mon
mon: mon-server mon-client

.PHONY: mon-server
mon-server:
	CGO_ENABLED=0 GOOS=linux go build -ldflags '-extldflags "-static"' -o $(GOPATH)/bin/mon-server ./examples/skydive/server/monitor-server.go

.PHONY: mon-client
mon-client:
	CGO_ENABLED=0 GOOS=linux go build -ldflags '-extldflags "-static"' -o $(GOPATH)/bin/mon-client ./examples/skydive/client/monitor-client.go
