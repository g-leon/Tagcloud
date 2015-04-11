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

	duration            = flag.Int("t", 5, "Number of seconds before closing the stream")
	tagcloudSize        = flag.Int("n", 0, "Print top 'n' words and then the rest of the words")
	printToFileFlag     = flag.Bool("f", false, "Print the output to file in adition to terminal")
	stopPrintScreenFlag = flag.Bool("s", false, "Suppress printing the output to terminal")
	redisFlag           = flag.Bool("r", false, "Use Redis to count word frequency")
)

type JSONTag struct {
	Word  string `json:"word"`
	Count int    `json:"count"`
}

type JSONOtherTag struct {
	Other []JSONTag `json:"other"`
}

func getWordFreqFromRedis() {
	keys, err := redisClient.Keys("*")
	if err != nil {
		cleanupRedis()
		log.Fatalf("Redis KEYS error: %s\n", err)
	}

	for _, k := range keys {
		c, err := redisClient.Get(k)
		if err != nil {
			if err == redis.ErrNilReply {
				continue
			} else {
				cleanupRedis()
				log.Fatalf("Redis GET error on word '%v': %s\n", k, err.Error())
			}
		}

		wordcount[k], err = strconv.Atoi(c)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func printWordFreq() {
	words := make([]JSONTag, 0)

	if *redisFlag {
		getWordFreqFromRedis()
	}

	for k, v := range wordcount {
		words = append(words, JSONTag{k, v})
	}

	j, _ := json.MarshalIndent(words, "", "    ")
	if !*stopPrintScreenFlag {
		fmt.Println(string(j))
	}

	if *printToFileFlag {
		printToFile(j)
	}
}

func printTopWordsFreq() {
	output := make([]interface{}, 0)
	jot := &JSONOtherTag{}

	if *redisFlag {
		getWordFreqFromRedis()
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

	if len(jot.Other) > 0 {
		output = append(output, jot)
	}

	j, _ := json.MarshalIndent(output, "", "    ")
	if !*stopPrintScreenFlag {
		fmt.Println(string(j))
	}

	if *printToFileFlag {
		printToFile(j)
	}
}

func printToFile(o []byte) {
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

// Returns true if the string contains a letter
// A string with no letters is not considered a word
func isWord(s string) bool {
	if len(s) == 0 {
		return false
	}

	for _, c := range s {
		if c >= 'a' && c <= 'z' {
			return true
		}
	}

	return false
}

func countWordFreq(s string) {
	s = strings.ToLower(s)
	ws := strings.Split(s, " ")
	for _, w := range ws {
		w = strings.TrimSpace(w)
		if !stopword[w] && isWord(w) {
			if *redisFlag {
				_, err := redisClient.Incr(w)
				if err != nil {
					log.Fatalf("Redis INCR error on word '%v': %s\n", w, err)
				}
			} else {
				wordcount[w]++
			}
		}
	}
}

func cleanupRedis() {
	redisClient.FlushAll()
	redisClient.Close()
}

func loadStopwords() {
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

func init() {
	loadStopwords()
}

func main() {
	flag.Parse()

	// Create new Redis client (if -r flag was used)
	if *redisFlag {
		redisClient = redis.New()

		err := redisClient.Connect(redisHost, redisPort)
		if err != nil {
			log.Fatalf("Redis connection failed: %s\n", err)
		}

		defer cleanupRedis()
	}

	// Create new streaming API client
	twitterClient := streamingtwitter.NewClient()

	err := twitterClient.Authenticate(&streamingtwitter.ClientTokens{
		TokenFile: TOKEN_FILE_NAME,
	})
	if err != nil {
		log.Fatalf("Twitter connection failed: %s\n", err)
	}

	// Filter the stream by language
	args := &url.Values{}
	args.Add("language", "en")

	// Launch the stream
	tweets := make(chan *streamingtwitter.TwitterStatus)
	go twitterClient.Stream(tweets, streamingtwitter.Streams["Sample"], args)

	// Stream runs for <*duration> seconds
	timer := time.NewTimer(time.Second * time.Duration(*duration))

	// Streaming
stream:
	for {
		select {
		case status := <-tweets:
			countWordFreq(status.Text)
		case err := <-twitterClient.Errors:
			fmt.Printf("Twitter client error: %s\n", err)
		case <-twitterClient.Finished:
			return
		}

		select {
		case <-timer.C:
			break stream
		default:
		}
	}

	// Print results
	if *tagcloudSize == 0 {
		printWordFreq()
	} else {
		printTopWordsFreq()
	}
}
