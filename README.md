# Tagcloud
A simple tagcloud service that counts the occurence of words from a stream of tweets.

## Quick Start
- go get github.com/g-leon/Tagcloud
- edit tokens.json to include your Twitter OAuth credentials
- go run main.go [OPTIONS]
- go test

## Command Line Options
```
-t  Number of seconds until the stream is closed (default is 5).
-n  Size of the tagcloud that contains most frequent N words (default is 0 - words are printed in random order). 
    Rest of the words will be aggregated in the "other" object of the json output.
-r  Use Redis to store and count the frequency of the words.
    Must have a Redis server started listening on the default port (6379).
-f  Print output to file in addition to terminal.
-s  Suppress printing the output to terminal.
```

