package jarvisbot

//go:generate go-bindata -pkg $GOPACKAGE -o assets.go data/

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/kardianos/osext"
	"github.com/tucnak/telebot"
)

var exchange_rate_bucket_name = []byte("rates")
var group_usernames_bucket_name = []byte("groups")
var file_cache_bucket_name = []byte("file_ids")
var kiwi_mangle_bucket_name = []byte("kiwi_mangle")

// JarvisBot is the main struct. All response funcs bind to this.
type JarvisBot struct {
	Name          string // The name of the bot registered with Botfather
	bot           *telebot.Bot
	log           *log.Logger
	fmap          FuncMap
	db            *bolt.DB
	keys          config
	googleKeyChan chan string
}

// Configuration struct for setting up Jarvis
type config struct {
	Name               string          `json:"name"`
	TelegramAPIKey     string          `json:"telegram_api_key"`
	OpenExchangeAPIKey string          `json:"open_exchange_api_key"`
	GiphyAPIKey        string          `json:"giphy_api_key"`
	YoutubeAPIKey      string          `json:"youtube_api_key"`
	MapsAPIKey         string          `json:"maps_api_key"`
	CustomSearchAPIKey string          `json:"custom_search_api_key"`
	CustomSearchID     string          `json:"custom_search_id"`
	SearchKeys         []googleKeyPair `json:"custom_search_credentials"`
}

// Wrapper struct for a message
type message struct {
	Cmd  string
	Args []string
	*telebot.Message
}

// GetArgs prints out the arguments for the message in one string.
func (m message) GetArgString() string {
	argString := ""
	for _, s := range m.Args {
		argString = argString + s + " "
	}
	return strings.TrimSpace(argString)
}

// a Google key pair represents a custom search api key and a custom search id.
type googleKeyPair struct {
	SearchID string `json:"search_id"`
	APIKey   string `json:"api_key"`
}

// Turns the key pair into a string for sending via a channel
func (g googleKeyPair) toString() string {
	return fmt.Sprintf("%s %s", g.APIKey, g.SearchID)
}

// A FuncMap is a map of command strings to response functions.
// It is use for routing comamnds to responses.
type FuncMap map[string]ResponseFunc

// ResponseFunc is a handler for a bot command.
type ResponseFunc func(m *message)

// Initialise a JarvisBot.
// lg is optional.
func InitJarvis(configJSON []byte, lg *log.Logger) *JarvisBot {
	// We'll use random numbers throughout JarvisBot
	rand.Seed(time.Now().UTC().UnixNano())

	if lg == nil {
		lg = log.New(os.Stdout, "[jarvis] ", 0)
	}

	var cfg config
	err := json.Unmarshal(configJSON, &cfg)
	if err != nil {
		lg.Fatalf("cannot unmarshal config json: %s", err)
	}

	if cfg.TelegramAPIKey == "" {
		log.Fatalf("config.json exists but doesn't contain a Telegram API Key! Read https://core.telegram.org/bots#3-how-do-i-create-a-bot on how to get one!")
	}

	botName := cfg.Name
	if botName == "" {
		log.Fatalf("config.json exists but doesn't contain a bot name. Set your botname when registering with The Botfather.")
	}

	bot, err := telebot.NewBot(cfg.TelegramAPIKey)
	if err != nil {
		log.Fatalf("error creating new bot, dude %s", err)
	}

	keyChannel := make(chan string)
	j := &JarvisBot{Name: botName, bot: bot, log: lg, keys: cfg, googleKeyChan: keyChannel}

	j.fmap = j.getDefaultFuncMap()

	// Setup database
	// Get current executing folder
	pwd, err := osext.ExecutableFolder()
	if err != nil {
		lg.Fatalf("cannot retrieve present working directory: %s", err)
	}

	db, err := bolt.Open(path.Join(pwd, "jarvis.db"), 0600, nil)
	if err != nil {
		lg.Fatal("unable to open bolt db: %s", err)
	}
	j.db = db
	createAllBuckets(db)

	// Ensure temp directory is created.
	// This is used to store media temporarily.
	tmpDirPath := filepath.Join(pwd, tempDir)
	if _, err := os.Stat(tmpDirPath); os.IsNotExist(err) {
		j.log.Printf("[%s] creating temporary directory", time.Now().Format(time.RFC3339))
		mkErr := os.Mkdir(tmpDirPath, 0775)
		if mkErr != nil {
			j.log.Printf("[%s] error creating temporary directory\n%s", time.Now().Format(time.RFC3339), err)
		}
	}

	// We loop through all the googleKeys and shove them into a channel
	// This is a fucking hack. I know. LOL.
	j.GoSafely(func() {
		googleKeys := j.keys.SearchKeys
		if len(googleKeys) < 1 {
			return
		}

		count := 0
		for {
			j.googleKeyChan <- googleKeys[count].toString()
			count++
			if count >= len(googleKeys) {
				count = 0
			}
		}
	})

	return j
}

// Listen exposes the telebot Listen API.
func (j *JarvisBot) Listen(subscription chan telebot.Message, timeout time.Duration) {
	j.bot.Listen(subscription, timeout)
}

// Get the built-in, default FuncMap.
func (j *JarvisBot) getDefaultFuncMap() FuncMap {
	return FuncMap{
		"/start":     j.Start,
		"/help":      j.Help,
		"/hello":     j.SayHello,
		"/echo":      j.Echo,
		"/e":         j.Echo,
		"/xchg":      j.Exchange,
		"/x":         j.Exchange,
		"/clear":     j.Clear,
		"/c":         j.Clear,
		"/img":       j.ImageSearch,
		"/psi":       j.PSI,
		"/source":    j.Source,
		"/google":    j.GoogleSearch,
		"/g":         j.GoogleSearch,
		"/gif":       j.GifSearch,
		"/youtube":   j.YoutubeSearch,
		"/yt":        j.YoutubeSearch,
		"/urbandict": j.UrbanDictSearch,
		"/ud":        j.UrbanDictSearch,
		"/loc":       j.LocationSearch,
		"/pingsetup": j.CollectPing,
		"/ping":      j.Ping,
	}
}

// Add a response function to the FuncMap
func (j *JarvisBot) AddFunction(command string, resp ResponseFunc) error {
	if !strings.HasPrefix(command, "/") {
		return fmt.Errorf("not a valid command string - it should be of the format /something")
	}
	j.fmap[command] = resp
	return nil
}

// Route received Telegram messages to the appropriate response functions.
func (j *JarvisBot) Router(msg telebot.Message) {
	// If the chat is a group chat, we save the username for the ping function.
	if msg.Chat.IsGroupChat() {
		j.GoSafely(func() { j.saveUsernameSafely(&msg.Chat, &msg.Sender) })
	}
	// Don't respond to forwarded commands
	if msg.IsForwarded() {
		return
	}
	jmsg := j.parseMessage(&msg)
	if jmsg.Cmd != "" {
		j.log.Printf("[%s][id: %d] command: %s, args: %s", time.Now().Format(time.RFC3339), jmsg.ID, jmsg.Cmd, jmsg.GetArgString())
	}
	execFn := j.fmap[jmsg.Cmd]

	if execFn != nil {
		j.GoSafely(func() { execFn(jmsg) })
	}
}

// CloseDB closes the connection with the bolt db.
func (j *JarvisBot) CloseDB() {
	j.db.Close()
}

// GoSafely is a utility wrapper to recover and log panics in goroutines.
// If we use naked goroutines, a panic in any one of them crashes
// the whole program. Using GoSafely prevents this.
func (j *JarvisBot) GoSafely(fn func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				stack := make([]byte, 1024*8)
				stack = stack[:runtime.Stack(stack, false)]

				j.log.Printf("PANIC: %s\n%s", err, stack)
			}
		}()

		fn()
	}()
}

// Ensure all buckets needed by jarvisbot are created.
func createAllBuckets(db *bolt.DB) error {
	// Check all buckets have been created
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(exchange_rate_bucket_name)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(group_usernames_bucket_name)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(file_cache_bucket_name)
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists(kiwi_mangle_bucket_name)
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

// Helper to parse incoming messages and return JarvisBot messages
func (j *JarvisBot) parseMessage(msg *telebot.Message) *message {
	cmd := ""
	args := []string{}

	if msg.IsReply() {
		// We use a hack. All reply-to messages have the command it's replying to as the
		// part of the message.
		r := regexp.MustCompile(`\/\w*`)
		res := r.FindString(msg.ReplyTo.Text)
		for k, _ := range j.fmap {
			if res == k {
				cmd = k
				args = strings.Split(msg.Text, " ")
				break
			}
		}
	} else if msg.Text != "" {
		msgTokens := strings.Fields(msg.Text)
		cmd, args = strings.ToLower(msgTokens[0]), msgTokens[1:]
		// Deal with commands of the form command@JarvisBot, which appear in
		// group chats.
		if strings.Contains(cmd, "@") {
			c := strings.Split(cmd, "@")
			cmd = c[0]
		}
	}

	return &message{Cmd: cmd, Args: args, Message: msg}
}

func (j *JarvisBot) SendMessage(recipient telebot.Recipient, msg string, options *telebot.SendOptions) {
	if shouldMangle, err := j.shouldKiwiMangle(recipient); shouldMangle {
		if err != nil {
			j.log.Printf("shouldKiwiMangle error: %s", err)
		}
		msg = Kiwiify(msg)
	}
	j.bot.SendMessage(recipient, msg, options)
}

func (j *JarvisBot) setShouldKiwiMangle(recipient telebot.Recipient, should bool) error {
	err := j.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(kiwi_mangle_bucket_name)
		err := b.Put([]byte(recipient.Destination()), []byte(strconv.FormatBool(should)))
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func (j *JarvisBot) shouldKiwiMangle(recipient telebot.Recipient) (bool, error) {
	should := true

	err := j.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(kiwi_mangle_bucket_name)
		v := b.Get([]byte(recipient.Destination()))
		var err error
		should, err = strconv.ParseBool(string(v[:]))
		return err
	})
	if err != nil {
		return true, err
	}

	return should, nil
}
