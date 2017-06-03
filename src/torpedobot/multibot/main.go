package multibot


import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"torpedobot/memcache"

	"github.com/getsentry/raven-go"
)

var bot *TorpedoBot
var once sync.Once


type TorpedoBot struct {
	caches          map[string]*memcache.MemCacheType
	commandHandlers map[string]func(*TorpedoBotAPI, interface{}, string)
	Config          struct {
		FacebookIncomingAddr string
		GoogleWebAppKey string
		KikIncomingAddr string
		KikWebHook string
		LastFmKey string
		LastFmSecret string
		LineIncomingAddr string
		SkypeIncomingAddr string
		PinterestToken string
		RavenEnabled bool
	}
	logger *log.Logger
	throttle *memcache.MemCacheType
}


func (tb *TorpedoBot) PostMessage(channel interface{}, message string, api *TorpedoBotAPI, richmsgs ...RichMessage) {
	if len(richmsgs) > 0 {
		api.PostMessage(channel, message, richmsgs[0])
	} else {
		api.PostMessage(channel, message)
	}

}


func (tb *TorpedoBot) processChannelEvent(api *TorpedoBotAPI, channel interface{}, incoming_message string) {
	if strings.HasPrefix(incoming_message, api.CommandPrefix) && tb.NoSpam(channel, incoming_message) {
		command := strings.TrimPrefix(incoming_message, api.CommandPrefix)
		found := 0
		for handler := range tb.commandHandlers {
			if strings.HasPrefix(strings.Split(command, " ")[0], handler) {
				found += 1
				if tb.Config.RavenEnabled {
					raven.CapturePanicAndWait(func() {
						tb.commandHandlers[handler](api, channel, incoming_message)
					}, nil)
				} else {
					tb.commandHandlers[handler](api, channel, incoming_message)
				}
				break
			}
		}
		tb.logger.Printf("PROCESS! -> `%s`", command)
		if found == 0 {
			api.PostMessage(channel, fmt.Sprintf("Could not process your message: %s%s. Command unknown. Send %shelp for list of valid commands.", api.CommandPrefix, command, api.CommandPrefix))
		}
	}
}


func (tb *TorpedoBot) RunLoop() {
	for {
		time.Sleep(time.Second)
	}
}


func (tb *TorpedoBot) RunBotsCSV(method func(apiKey, cmd_prefix string), CSV, cmd_prefix string) {
	wrapped := func(a, b string) {}
	if tb.Config.RavenEnabled {
		wrapped = func(apiKey, cmd_prefix string) {
			// this should (!) capture bot protocol panic
			raven.CapturePanicAndWait(func() {
				method(apiKey, cmd_prefix)
			}, nil)
		}
	} else {
		wrapped = method
	}
	for _, key := range strings.Split(CSV, ",") {
		if key == "" {
			continue
		}
		go wrapped(key, cmd_prefix)
	}
}


func (tb *TorpedoBot) RegisterHandlers(handlers map[string]func(*TorpedoBotAPI, interface{}, string)) {
	tb.commandHandlers = handlers
	return
}


func (tb *TorpedoBot) GetCommandHandlers() (handlers map[string]func(*TorpedoBotAPI, interface{}, string)) {
	return tb.commandHandlers
}


func New(facebook_incoming_addr, google_webapp_key, skype_incoming_addr, kik_incoming_addr, kik_webhook_url, lastfm_key, lastfm_secret, line_incoming_addr, pinterest_token string) *TorpedoBot {
	once.Do(func() {
		bot = &TorpedoBot{}
		bot.logger = log.New(os.Stdout, "torpedo-bot: ", log.Lshortfile|log.LstdFlags)
		bot.caches = make(map[string]*memcache.MemCacheType)
		bot.Config.SkypeIncomingAddr = skype_incoming_addr
		bot.Config.FacebookIncomingAddr = facebook_incoming_addr
		bot.Config.GoogleWebAppKey = google_webapp_key
		bot.Config.KikIncomingAddr = kik_incoming_addr
		bot.Config.KikWebHook = kik_webhook_url
		bot.Config.LastFmKey = lastfm_key
		bot.Config.LastFmSecret = lastfm_secret
		bot.Config.LineIncomingAddr = line_incoming_addr
		bot.Config.PinterestToken = pinterest_token
		bot.throttle = memcache.New()
		env_dsn := os.Getenv("SENTRY_DSN")
		if env_dsn != "" {
			bot.logger.Print("Using Sentry error reporting...\n")
			raven.SetDSN(env_dsn)
			bot.Config.RavenEnabled = true
		}
	})
	return bot
}
