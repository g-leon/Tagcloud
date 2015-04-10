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

	"github.com/tg/gosortmap"
)

const STOPWORDS_FILE_NAME = "stopwords.txt"
const TOPWORDS_FILE_NAME = "topwords.txt"
const TOKEN_FILE_NAME = "tokens.json"

var (
	stopword  = make(map[string]bool)
	wordcount = make(map[string]int)

	duration     = flag.Int("t", 3, "Number of seconds before closing the stream")
	tagcloudSize = flag.Int("s", 0, "Prints top 's' words and then the rest of the words")
)

type JSONTag struct {
	Word  string `json:"word"`
	Count int    `json:"count"`
}

type JSONOutput struct {
	TopWords []JSONTag `json:"top"`
	Other    []JSONTag `json:"other"`
}

func PrintWordFreq() {
	words := make([]JSONTag, 0)

	for k, v := range wordcount {
		words = append(words, JSONTag{k, v})
	}

	j, _ := json.MarshalIndent(words, "", "    ")
	fmt.Println(string(j))
}

func PrintTopWordsFreq() {
	top := make([]JSONTag, 0)
	other := make([]JSONTag, 0)

	nr := 0
	for _, e := range sortmap.ByValDesc(wordcount) {
		if nr < *tagcloudSize {
			top = append(top, JSONTag{e.K.(string), e.V.(int)})
		} else {
			other = append(other, JSONTag{e.K.(string), e.V.(int)})
		}
		nr++
	}

	output := &JSONOutput{top, other}
	j, _ := json.MarshalIndent(output, "", "    ")
	fmt.Println(string(j))

	PrintToFile(j)
}

func PrintToFile(o []byte) {
	f, err := os.Create(TOPWORDS_FILE_NAME)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, err = f.Write(o)
	if err != nil {
		log.Fatal(err)
	}
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
		close(tweets)
	}()

	// Streaming
stream:
	for {
		select {
		case status := <-tweets:
			if status != nil {
				CountWords(status.Text)
			} else {
				break stream
			}
		case err := <-client.Errors:
			fmt.Printf("ERROR: '%s'\n", err)
		case <-client.Finished:
			return
		}
	}

	// Print results
	if *tagcloudSize == 0 {
		PrintWordFreq()
	} else {
		PrintTopWordsFreq()
	}
}
