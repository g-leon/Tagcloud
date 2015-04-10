package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/JustAdam/streamingtwitter"
)

const STOPWORDS_FILE_NAME = "stopwords.txt"
const TOKEN_FILE_NAME = "tokens.json"

var (
	stopword  = make(map[string]bool)
	wordcount = make(map[string]int)
)

type OutputElem struct {
	Word  string `json:"word"`
	Count int    `json:"count"`
}

func PrintOutput() {
	sl := make([]OutputElem, 0)

	for k, v := range wordcount {
		sl = append(sl, OutputElem{k, v})
	}

	j, _ := json.MarshalIndent(sl, "", "    ")
	fmt.Println(string(j))
}

func LoadStopwords() {
	file, err := os.Open(STOPWORDS_FILE_NAME)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		w := scanner.Text()
		stopword[w] = true
	}
}

func CountWords(s string) {
	s = strings.ToLower(s)
	ws := strings.Split(s, " ")
	for _, w := range ws {
		if !stopword[w] {
			wordcount[w]++
		}
	}
}

func init() {
	LoadStopwords()
}

func main() {
	duration := flag.Int("t", 3, "Number of seconds before closing the stream")
	flag.Parse()

	// Create new streaming API client
	client := streamingtwitter.NewClient()

	err := client.Authenticate(&streamingtwitter.ClientTokens{
		TokenFile: TOKEN_FILE_NAME,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Filter by language
	args := &url.Values{}
	args.Add("language", "en")

	// Launch the stream
	tweets := make(chan *streamingtwitter.TwitterStatus)
	go client.Stream(tweets, streamingtwitter.Streams["Sample"], args)

	// Stream runs for a number of seconds equal to *duration
	timer := time.NewTimer(time.Second * time.Duration(*duration))

	go func() {
		<-timer.C
		fmt.Println("exiting..")
		//os.Exit(1)
		close(tweets)
	}()

	// Streaming
loop:
	for {
		select {
		// Recieve tweets
		case status := <-tweets:
			if status != nil {
				//fmt.Println(status.Text)
				CountWords(status.Text)
			} else {
				break loop
			}
		// Any errors that occured
		case err := <-client.Errors:
			fmt.Printf("ERROR: '%s'\n", err)
			// Stream has finished
		case <-client.Finished:
			return
		}
	}

	PrintOutput()

}
