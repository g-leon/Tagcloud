FROM golang:onbuild
RUN go get github.com/g-leon/Tagcloud
ENTRYPOINT ["go", "run", "main.go"]
