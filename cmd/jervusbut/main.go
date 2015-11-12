package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/ejamesc/jarvisbot"
	"github.com/kardianos/osext"
	"github.com/tucnak/telebot"
)

func main() {
	// Grab current executing directory
	// In most cases it's the folder in which the Go binary is located.
	pwd, err := osext.ExecutableFolder()
	if err != nil {
		log.Fatalf("error getting executable folder: %s", err)
	}
	configJSON, err := ioutil.ReadFile(path.Join(pwd, "config.json"))
	if err != nil {
		log.Fatalf("error reading config file! Boo: %s", err)
	}

	logger := log.New(os.Stdout, "[jarvis] ", 0)

	jb := jarvisbot.InitJarvis(configJSON, logger)
	defer jb.CloseDB()

	jb.AddFunction("/laugh", jb.SendLaugh)
	jb.AddFunction("/neverforget", jb.NeverForget)
	jb.AddFunction("/touch", jb.Touch)
	jb.AddFunction("/hanar", jb.Hanar)
	jb.AddFunction("/logic", jb.SendLogic)
	jb.AddFunction("/yank", jb.Yank)
	jb.AddFunction("/tellthatto", jb.TellThatTo)
	jb.AddFunction("/kanjiklub", jb.TellThatTo)
	jb.AddFunction("/ducks", jb.SendImage("quack quack motherfucker"))
	jb.AddFunction("/chickens", jb.SendImage("cluck cluck motherfucker"))

	// Additional jokes
	jb.AddFunction("/accent", jb.Accent)
	jb.AddFunction("/tangent", jb.Trigonometry)
	jb.AddFunction("/sine", jb.Trigonometry)
	jb.AddFunction("/cosine", jb.Trigonometry)
	jb.AddFunction("/echosat", jb.EchoSat)
	jb.AddFunction("/es", jb.EchoSat)
	jb.AddFunction("/indian", jb.Indian)

	// Core Kiwis
	jb.AddFunction("/stert", jb.Start)
	jb.AddFunction("/hilp", jb.Help)
	jb.AddFunction("/hillu", jb.SayHello)
	jb.AddFunction("/ichu", jb.Echo)
	jb.AddFunction("/i", jb.Echo)
	jb.AddFunction("/ichuset", jb.EchoSat)
	jb.AddFunction("/is", jb.EchoSat)
	jb.AddFunction("/clir", jb.Clear)
	jb.AddFunction("/surce", jb.Source)
	jb.AddFunction("/gugle", jb.GoogleSearch)
	jb.AddFunction("/guf", jb.GifSearch)
	jb.AddFunction("/yutube", jb.YoutubeSearch)
	jb.AddFunction("/urbenduct", jb.UrbanDictSearch)
	jb.AddFunction("/luc", jb.LocationSearch)
	jb.AddFunction("/pungsitup", jb.CollectPing)
	jb.AddFunction("/pung", jb.Ping)
	jb.AddFunction("/psu", jb.PSI)

	// Additional Kiwis
	jb.AddFunction("/legh", jb.SendLaugh)
	jb.AddFunction("/nivirfurgit", jb.NeverForget)
	jb.AddFunction("/tuch", jb.Touch)
	jb.AddFunction("/hener", jb.Hanar)
	jb.AddFunction("/lugic", jb.SendLogic)
	jb.AddFunction("/yenk", jb.Yank)
	jb.AddFunction("/tillthettu", jb.TellThatTo)
	jb.AddFunction("/kenjuklub", jb.TellThatTo)
	jb.AddFunction("/chuckins", jb.SendImage("cluck cluck motherfucker"))

	// Kiwiied Kiwis
	jb.AddFunction("/eccint", jb.Accent)
	jb.AddFunction("/tengint", jb.Trigonometry)
	jb.AddFunction("/sune", jb.Trigonometry)
	jb.AddFunction("/cusine", jb.Trigonometry)
	jb.AddFunction("/indun", jb.Indian)

	// Mangling
	jb.AddFunction("/englishplease", jb.NoMangle)
	jb.AddFunction("/relaxnow", jb.Mangle)

	jb.GoSafely(func() {
		logger.Println("Scheduling exchange rate update")
		for {
			time.Sleep(1 * time.Hour)
			jb.RetrieveAndSaveExchangeRates()
			logger.Printf("[%s] exchange rates updated!", time.Now().Format(time.RFC3339))
		}
	})

	messages := make(chan telebot.Message)
	jb.Listen(messages, 1*time.Second)

	for message := range messages {
		jb.Router(message)
	}
}
