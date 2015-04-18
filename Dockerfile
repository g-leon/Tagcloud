FROM golang:onbuild
ENTRYPOINT ["go", "run", "main.go"]
