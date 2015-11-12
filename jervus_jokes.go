package jarvisbot

import (
	"fmt"
	"github.com/tucnak/telebot"
	"math/rand"
)

func (j *JarvisBot) EchoSat(msg *message) {
	if len(msg.Args) == 0 {
		so := &telebot.SendOptions{ReplyTo: *msg.Message, ReplyMarkup: telebot.ReplyMarkup{ForceReply: true, Selective: true}}
		j.SendMessage(msg.Chat, "/echosat Jarvis Parrot Mode \U0001F426\nWhat do you want me to parrot?\n\n", so)
	}
	response := ""
	for _, s := range msg.Args {
		response = response + s + " "
	}
	old_resp, resp := response, Kiwiify(response)
	for resp != old_resp {
		old_resp, resp = resp, Kiwiify(resp)
	}
	j.SendMessage(msg.Chat, resp, nil)
}

// Accent explains JervisChin's accent
func (j *JarvisBot) Accent(msg *message) {
	j.SendMessage(msg.Chat, "I know my accent isn't perfectly accurate, but we're working with character replacements here, not phonemes :<", nil)
}

// Tangent
func (j *JarvisBot) Trigonometry(msg *message) {
	j.SendMessage(msg.Chat, "Math? I can't do math!", nil)
}

func (j *JarvisBot) Indian(msg *message) {
	videos := []string{
		"https://www.youtube.com/watch?v=hjWd9a8Ck8U",
		"https://www.youtube.com/watch?v=WEldpMQGfYE",
		"https://www.youtube.com/watch?v=olF4kpkiWys",
		"https://www.youtube.com/watch?v=8tw7LIykvBw",
	}
	n := rand.Intn(len(videos))
	txt := fmt.Sprintf("Indian! \U0001F1EE\U0001F1F3\n%s", videos[n])
	j.SendMessage(msg.Chat, txt, nil)
}

func (j *JarvisBot) Mangle(msg *message) {
	if msg.Sender.ID == 56951899 {
		j.setShouldKiwiMangle(msg.Chat, true)
		j.SendMessage(msg.Chat, "Chur bro, I'll speak normally again!", nil)
	} else {
		j.SendMessage(msg.Chat, "Thanks, I'll prepare myself a bubble bath :D", nil)
	}
}

func (j *JarvisBot) NoMangle(msg *message) {
	if msg.Sender.ID == 56951899 {
		j.setShouldKiwiMangle(msg.Chat, false)
		j.SendMessage(msg.Chat, "*sigh* Okay, I'll try to speak proper English.", nil)
	} else {
		j.SendMessage(msg.Chat, "What do you mean? What did you think I was speaking all along? That's offensive, sir/madam! HELP! I'M BEING OPPRESSED!", nil)
	}
}
