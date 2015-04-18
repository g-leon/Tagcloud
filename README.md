# Tagcloud
A simple tagcloud service that counts the occurence of words from a stream of tweets.

## Quick Start
- go get github.com/g-leon/Tagcloud
- cd $GOPATH/src/github.com/g-leon/Tagcloud
- edit tokens.json to include your Twitter OAuth credentials
- go run main.go [OPTIONS]
- go test

## Command Line Options
```
-t  Number of seconds until the stream is closed (default is 5).
-n  Size of the tagcloud that contains most frequent N words. 
    Rest of the words will be aggregated in the "other" object of the json output.
    (default is 0 - words are printed in random order).
-r  Use Redis to store and count the frequency of the words.
    Must have a Redis server started listening on the default port (6379).
-f  Print output to file in addition to terminal.
-s  Suppress printing the output to terminal.
```

### Run Tagcloud in a Docker container 
- cd $GOPATH/src/github.com/g-leon/Tagcloud
- docker build -t tagcloud .
- docker run tagcloud [OPTIONS]

### Use a dockerized Redis server
- docker run -d -p 127.0.0.1:6379:6379 --name redis redis
- cd $GOPATH/src/github.com/g-leon/Tagcloud
- go run main.go [OPTIONS]

### Use a dockerized Tagcloud with a linked dockerized Redis server
- docker run -d -p 127.0.0.1:6379:6379 --name redis redis
- cd $GOPATH/src/github.com/g-leon/Tagcloud
- docker build -t tagcloud .
- docker run --link redis:redis tagcloud -r [OPTIONS]


