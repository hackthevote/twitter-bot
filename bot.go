package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"voteinfobot/bot"
)



const (
	T_CONSUMER_K       = "TWITTER_CONSUMER_KEY"
	T_CONSUMER_SEC_K   = "TWITTER_CONSUMER_SEC_KEY"
	T_ACCESS_TOKEN     = "TWITTER_ACCESS_TOKEN"
	T_ACCESS_TOKEN_SEC = "TWITTER_ACCESS_TOKEN_SECRET"
)

var logFile *string = flag.String("logfile", "bot.log", "path to logfile for the bot to use")

func main() {
	flag.Parse()

	f, err := os.OpenFile(*logFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("err opening logfile %s: %v", *logFile, err)
	}

	logger := log.New(f, "", log.LstdFlags)

	consumerKey := os.Getenv(T_CONSUMER_K)
	consumerSecretKey := os.Getenv(T_CONSUMER_SEC_K)
	accessToken := os.Getenv(T_ACCESS_TOKEN)
	accessTokenSecret := os.Getenv(T_ACCESS_TOKEN_SEC)

	missingEnvVars := ""
	for _, kv := range [][2]string{
		[2]string{consumerKey, T_CONSUMER_K},
		[2]string{consumerSecretKey, T_CONSUMER_SEC_K},
		[2]string{accessToken, T_ACCESS_TOKEN},
		[2]string{accessTokenSecret, T_ACCESS_TOKEN_SEC}} {

		if kv[0] == "" {
			logger.Printf("Err, env var %s missing - unable to start", kv[1])
			missingEnvVars += kv[1] + " "
		}
	}

	if missingEnvVars != "" {
		log.Fatalf("Err - missing environment vars: %s", missingEnvVars)
	}

	logC := make(chan string)
	go func() {
		for m := range logC {
			logger.Println(m)
		}
	}()

	botWg := &sync.WaitGroup{}
	handlers := []bot.Handler{
		bot.NewTwitterHandler(botWg, logC, consumerKey, consumerSecretKey, accessToken, accessTokenSecret),
	}
	botWg.Add(len(handlers))

	go func() {
		// Channel buffered for sig notify
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		<-c
		logger.Println("[BOT] Received SIGINT/SIGTERM. Shutting down bot.")
		for _, h := range handlers {
			go h.Stop()
		}
	}()

	logger.Println("[BOT] Bot initialised, starting up...")
	for _, h := range handlers {
		go h.Start()
	}

	botWg.Wait()
	close(logC)
	logger.Println("[BOT] All handlers stopped, bot shutdown complete.")
}
