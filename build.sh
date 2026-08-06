CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=1.0.0" -o bin/inx.exe ./cmd/inx
CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=1.0.0" -o bin/inx-plugin-example.exe ./cmd/inx-plugin-example
