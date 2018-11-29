.PHONY: mon
mon: mon-local-server mon-server mon-client

.PHONY: mon-local-server
mon-local-server:
	CGO_ENABLED=0 GOOS=linux go build -ldflags '-extldflags "-static"' -o $(GOPATH)/bin/mon-local-server ./examples/skydive/server/local-server.go

.PHONY: mon-server
mon-server:
	CGO_ENABLED=0 GOOS=linux go build -ldflags '-extldflags "-static"' -o $(GOPATH)/bin/mon-server ./examples/skydive/server/monitor-server.go

.PHONY: mon-client
mon-client:
	CGO_ENABLED=0 GOOS=linux go build -ldflags '-extldflags "-static"' -o $(GOPATH)/bin/mon-client ./examples/skydive/client/monitor-client.go