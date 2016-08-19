BINARY := kubenetes-slack
LDFLAGS := -ldflags="-s -w"

bin/kubenetes-slack:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/kubenetes-slack src/main/main.go

clean:
	rm -rf bin/*
