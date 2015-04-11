package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/JustAdam/streamingtwitter"
	"github.com/tg/gosortmap"
	"menteslibres.net/gosexy/redis"
)

const STOPWORDS_FILE_NAME = "stopwords.txt"
const TOPWORDS_FILE_NAME = "topwords.json"
const TOKEN_FILE_NAME = "tokens.json"

var (
	stopword  = make(map[string]bool)
	wordcount = make(map[string]int)

	redisHost   = "127.0.0.1"
	redisPort   = uint(6379)
	redisClient *redis.Client

	duration     = flag.Int("t", 5, "Number of seconds before closing the stream")
	tagcloudSize = flag.Int("n", 0, "Print top 'n' words and then the rest of the words")
	printToFile  = flag.Bool("f", false, "Print the output to file in adition to terminal")
	redisFlag    = flag.Bool("r", false, "Use Redis to count word frequency")
)

type JSONTag struct {
	Word  string `json:"word"`
	Count int    `json:"count"`
}

type JSONOtherTag struct {
	Other []JSONTag `json:"other"`
}

func GetWordFreqFromRedis() {
	keys, err := redisClient.Keys("*")
	if err != nil {
		log.Fatal(err)
	}
	for _, k := range keys {
		c, err := redisClient.Get(k)
		if err != nil {
			log.Fatal(err)
		}
		wordcount[k], err = strconv.Atoi(c)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func PrintWordFreq() {
	words := make([]JSONTag, 0)

	if *redisFlag {
		GetWordFreqFromRedis()
	}

	for k, v := range wordcount {
		words = append(words, JSONTag{k, v})
	}

	j, _ := json.MarshalIndent(words, "", "    ")
	fmt.Println(string(j))

	if *printToFile {
		PrintToFile(j)
	}
}

func PrintTopWordsFreq() {
	output := make([]interface{}, 0)
	jot := &JSONOtherTag{}

	if *redisFlag {
		GetWordFreqFromRedis()
	}

	nr := 0
	for _, e := range sortmap.ByValDesc(wordcount) {
		if nr < *tagcloudSize {
			output = append(output, JSONTag{e.K.(string), e.V.(int)})
		} else {
			jot.Other = append(jot.Other, JSONTag{e.K.(string), e.V.(int)})
		}
		nr++
	}
	output = append(output, jot)

	j, _ := json.MarshalIndent(output, "", "    ")
	fmt.Println(string(j))

	if *printToFile {
		PrintToFile(j)
	}
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

func CountWordFreq(s string) {
	s = strings.ToLower(s)
	ws := strings.Split(s, " ")
	for _, w := range ws {
		if !stopword[w] {
			if *redisFlag {
				redisClient.Incr(w)
			} else {
				wordcount[w]++
			}

		}
	}
}

func init() {
	LoadStopwords()
}

func main() {
	flag.Parse()

	// Create new Redis client (if -r flag was used)
	if *redisFlag {
		redisClient = redis.New()

		err := redisClient.Connect(redisHost, redisPort)
		if err != nil {
			log.Fatalf("Redis connection failed: %s\n", err.Error())
		}

		// Cleanup Redis after the program has ended
		defer func() {
			redisClient.FlushAll()
			redisClient.Close()
		}()
	}

	// Create new streaming API client
	twitterClient := streamingtwitter.NewClient()

	err := twitterClient.Authenticate(&streamingtwitter.ClientTokens{
		TokenFile: TOKEN_FILE_NAME,
	})
	if err != nil {
		log.Fatalf("Twitter connection failed: %s\n", err.Error())
	}

	// Filter the stream by language
	args := &url.Values{}
	args.Add("language", "en")

	// Launch the stream
	tweets := make(chan *streamingtwitter.TwitterStatus)
	go twitterClient.Stream(tweets, streamingtwitter.Streams["Sample"], args)

	// Stream runs for <*duration> seconds
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
				CountWordFreq(status.Text)
			} else {
				break stream
			}
		case err := <-twitterClient.Errors:
			fmt.Printf("ERROR: '%s'\n", err)
		case <-twitterClient.Finished:
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
