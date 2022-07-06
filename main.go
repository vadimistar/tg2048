package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const fieldSize = 4
const startSquareVal = 2

type field [][]int

func newField() field {
	f := make(field, fieldSize)
	for i := 0; i < fieldSize; i++ {
		f[i] = make([]int, fieldSize)
	}
	return f
}

func (f field) String() string {
	res := "```\n"
	for i := 0; i < fieldSize; i++ {
		for j := 0; j < fieldSize; j++ {
			if f[i][j] != 0 {
				res += fmt.Sprintf("[%4d]\t", f[i][j])
			} else {
				res += "[    ]\t"
			}
		}
		res += "\n"
	}
	res += "```\n"
	return res
}

func (f field) full() bool {
	for i := 0; i < fieldSize; i++ {
		for j := 0; j < fieldSize; j++ {
			if f[i][j] == 0 {
				return false
			}
		}
	}
	return true
}

type game struct {
	field
	score int
}

func newGame() (game, error) {
	g := game{
		field: newField(),
	}
	if err := g.prepare(); err != nil {
		return game{}, err
	}
	return g, nil
}

func (g game) withBestScore(score int) string {
	return fmt.Sprintf("Score: %d Best: %d\n%s", g.score, score, g.field)
}

type gameOver struct {
	score int
}

func (g gameOver) Error() string {
	return fmt.Sprintf("Game over. Your score is %d", g.score)
}

func (g game) checkOverlap(x *int, y *int) error {
	if g.field.full() {
		return gameOver{score: g.score}
	}
	for {
		if g.field[*x][*y] != 0 {
			if *x == fieldSize-1 {
				*y = (*y + 1) % fieldSize
			}
			*x = (*x + 1) % fieldSize
			continue
		}
		return nil
	}
}

var rngSrc = rand.NewSource(time.Now().UnixNano())
var rng = rand.New(rngSrc)

func (g *game) prepare() error {
	if _, _, e := g.addSquare(); e != nil {
		return e
	}
	if _, _, e := g.addSquare(); e != nil {
		return e
	}
	return nil
}

func (g *game) addSquare() (int, int, error) {
	x := rng.Intn(fieldSize)
	y := rng.Intn(fieldSize)
	if e := g.checkOverlap(&x, &y); e != nil {
		return 0, 0, e
	}
	g.field[x][y] = startSquareVal
	return x, y, nil
}

func (g *game) move(x int, y int, next func(int, int) (int, int), check func(int, int) bool, incr func(*int, *int)) {
	for ; check(x, y); incr(&x, &y) {
		nx, ny := next(x, y)
		if g.field[nx][ny] == 0 {
			g.field[nx][ny] = g.field[x][y]
			log.Printf("%d %d is moved from %d %d", nx, ny, x, y)
		} else if g.field[nx][ny] == g.field[x][y] {
			g.field[nx][ny] += g.field[x][y]
			g.score += g.field[nx][ny]
			log.Printf("%d %d combines with %d %d", nx, ny, x, y)
		} else {
			break
		}
		g.field[x][y] = 0
	}
}

func (g *game) moveLeft(x int, y int) {
	g.move(x, y, func(x, y int) (int, int) {
		return x, y - 1
	}, func(_, y int) bool {
		return y > 0
	}, func(_, y *int) {
		*y--
	})
}

func (g *game) moveUp(x int, y int) {
	g.move(x, y, func(x, y int) (int, int) {
		return x - 1, y
	}, func(x, _ int) bool {
		return x > 0
	}, func(x, _ *int) {
		*x--
	})
}

func (g *game) moveDown(x int, y int) {
	g.move(x, y, func(x, y int) (int, int) {
		return x + 1, y
	}, func(x, _ int) bool {
		return x < fieldSize-1
	}, func(x, _ *int) {
		*x++
	})
}

func (g *game) moveRight(x int, y int) {
	g.move(x, y, func(x, y int) (int, int) {
		return x, y + 1
	}, func(_, y int) bool {
		return y < fieldSize-1
	}, func(_, y *int) {
		*y++
	})
}

func (g *game) left() {
	for x := 0; x < fieldSize; x++ {
		for y := 0; y < fieldSize; y++ {
			if g.field[x][y] != 0 {
				g.moveLeft(x, y)
			}
		}
	}
}

func (g *game) up() {
	for x := 0; x < fieldSize; x++ {
		for y := 0; y < fieldSize; y++ {
			if g.field[x][y] != 0 {
				g.moveUp(x, y)
			}
		}
	}
}

func (g *game) down() {
	for x := fieldSize - 1; x >= 0; x-- {
		for y := 0; y < fieldSize; y++ {
			if g.field[x][y] != 0 {
				g.moveDown(x, y)
			}
		}
	}
}

func (g *game) right() {
	for x := 0; x < fieldSize; x++ {
		for y := fieldSize - 1; y >= 0; y-- {
			if g.field[x][y] != 0 {
				g.moveRight(x, y)
			}
		}
	}
}

const updateConfigTimeout = 30

var gameKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("⬅️️"),
		tgbotapi.NewKeyboardButton("️⬆️"),
		tgbotapi.NewKeyboardButton("️➡️️️"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton(" "),
		tgbotapi.NewKeyboardButton("⬇️"),
		tgbotapi.NewKeyboardButton(" "),
	),
)

type client struct {
	game
	lastMsgId int
	senderId  int64
	bestScore int
	started   bool
}

var clients = make([]client, 0)

func clientWithId(id int64) *client {
	for i := range clients {
		if clients[i].senderId == id {
			return &clients[i]
		}
	}
	return nil
}

func newClient(senderId int64) (client, error) {
	g, err := newGame()
	if err != nil {
		return client{}, err
	}
	return client{
		game:     g,
		senderId: senderId,
		started:  false,
	}, nil
}

func (c *client) restart() error {
	score := c.game.score
	if score > c.bestScore {
		c.bestScore = score
	}
	game, err := newGame()
	if err != nil {
		return err
	}
	c.game = game
	c.lastMsgId = 0
	c.started = false
	return nil
}

func sendField(c *client, b *tgbotapi.BotAPI, u *tgbotapi.Update) {
	msg := tgbotapi.NewMessage(u.Message.Chat.ID, c.game.withBestScore(c.bestScore))
	msg.ParseMode = "Markdown"
	sent, err := b.Send(msg)
	if err != nil {
		log.Fatalln(err)
	}
	c.lastMsgId = sent.MessageID
}

func removeClient(c *client) {
	index := 0
	for i := range clients {
		if &clients[i] == c {
			index = i
		}
	}
	clients = append(clients[:index], clients[index+1:]...)
}

func main() {
	log.Println("Token is", os.Getenv("TELEGRAM_API_TOKEN"))
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_API_TOKEN"))
	if err != nil {
		log.Fatalln("Can't create Telegram API:", err, "(API token might be invalid)")
	}
	bot.Debug = true
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = updateConfigTimeout
	updates := bot.GetUpdatesChan(updateConfig)
	for update := range updates {
		if update.Message == nil {
			continue
		}
		client := clientWithId(update.Message.From.ID)
		if client == nil {
			newClient, err := newClient(update.Message.From.ID)
			if err != nil {
				log.Fatalln(err)
			}
			clients = append(clients, newClient)
			client = &clients[len(clients)-1]
		}
		log.Printf("Current client: %v", client)
		if !client.started {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, `**tg2048**
2048 Clone Telegram Bot
by [vadimistar](https://github.com/vadimistar) *(Vadim Starostin)*

Press any button to start

/stop - stop the game and reset the best score
/restart - restart the current game`)
			msg.ParseMode = "Markdown"
			msg.ReplyMarkup = gameKeyboard
			if _, err := bot.Send(msg); err != nil {
				log.Fatalln(err)
			}
			sendField(client, bot, &update)
			client.started = true
			continue
		}
		if client.lastMsgId != 0 {
			bot.Send(tgbotapi.NewDeleteMessage(update.Message.Chat.ID,
				client.lastMsgId))
		}
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "stop":
				removeClient(client)
			case "restart":
				if err := client.restart(); err != nil {
					log.Fatalln(err)
				}
			}
		}
		input := update.Message.Text
		d := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID)
		bot.Send(d)
		switch input {
		case "⬅️️":
			client.game.left()
		case "️⬆️":
			client.game.up()
		case "️➡️️️":
			client.game.right()
		case "⬇️":
			client.game.down()
		default:
			continue
		}
		x, y, err := client.game.addSquare()
		if err != nil {
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
			client.restart()
			continue
		}
		log.Printf("square at %d %d was generated", x, y)
		sendField(client, bot, &update)
	}
}
