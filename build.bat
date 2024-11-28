set GOARCH=amd64
rem set GOOS=darwin
set GOOS=linux
set CGO_ENABLED=0
rem go build -a -installsuffix cgo -o gapp .
go build -o ./ffimage-syncer