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

func newGame() game {
	return game{
		field: newField(),
	}
}

func (g game) String() string {
	return fmt.Sprintf("Score: %d\n%s", g.score, g.field)
}

type gameOver struct{}

func (g gameOver) Error() string {
	return "Game over"
}

func (f field) checkOverlap(x *int, y *int) error {
	if f.full() {
		return gameOver{}
	}
	for {
		if f[*x][*y] != 0 {
			if *x == len(f)-1 {
				*y = (*y + 1) % len(f)
			}
			*x = (*x + 1) % len(f)
			continue
		}
		return nil
	}
}

var rngSrc = rand.NewSource(time.Now().UnixNano())
var rng = rand.New(rngSrc)

func (f field) prepare() error {
	if _, _, e := f.addSquare(); e != nil {
		return e
	}
	if _, _, e := f.addSquare(); e != nil {
		return e
	}
	return nil
}

func (f field) addSquare() (int, int, error) {
	x := rng.Intn(fieldSize)
	y := rng.Intn(fieldSize)
	if e := f.checkOverlap(&x, &y); e != nil {
		return 0, 0, e
	}
	f[x][y] = startSquareVal
	return x, y, nil
}

func (g *game) handle(b byte) bool {
	switch b {
	case 'a':
		g.left()
	case 'w':
		g.up()
	case 's':
		g.down()
	case 'd':
		g.right()
	default:
		return false
	}
	return true
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
	game := newGame()
	if err := game.prepare(); err != nil {
		log.Fatalln(err)
	}
	lastMsgId := 0
	for update := range updates {
		if update.Message == nil {
			continue
		}
		if lastMsgId == 0 {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, `**tg2048**
2048 Clone Telegram Bot
by [vadimistar](https://github.com/vadimistar) *(Vadim Starostin)*`)
			msg.ParseMode = "Markdown"
			msg.ReplyMarkup = gameKeyboard
			if _, err := bot.Send(msg); err != nil {
				log.Fatalln(err)
			}
		} else {
			bot.Send(tgbotapi.NewDeleteMessage(update.Message.Chat.ID,
				lastMsgId))
		}
		input := update.Message.Text
		d := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID)
		bot.Send(d)
		switch input {
		case "⬅️️":
			game.left()
		case "️⬆️":
			game.up()
		case "️➡️️️":
			game.right()
		case "⬇️":
			game.down()
		default:
			continue
		}
		x, y, err := game.addSquare()
		if err != nil {
			log.Fatalln(err)
		}
		log.Printf("square at %d %d was generated", x, y)
		fieldMsg := tgbotapi.NewMessage(update.Message.Chat.ID, game.String())
		fieldMsg.ParseMode = "Markdown"
		sentFieldMsg, err := bot.Send(fieldMsg)
		if err != nil {
			log.Fatalln(err)
		}
		lastMsgId = sentFieldMsg.MessageID
		//switch update.Message.Command() {
		// case "start":
		// 	msg.ReplyMarkup = gameKeyboard
		// default:
		// 	msg.Text = update.Message.Command()
		// }
		// if _, err := bot.Send(msg); err != nil {
		// 	log.Fatalln(err)
		// }
		// if game.handle(update.Message.Text[0]) {
		// 	msg := tgbotapi.NewMessage()
		// }
		// msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		// msg.ReplyToMessageID = update.Message.MessageID
		// if _, err := bot.Send(msg); err != nil {
		// 	log.Fatalln(err)
		// }
	}
	// b := make([]byte, 256)
	// for {
	// 	fmt.Println("Score: ", score)
	// 	showField()
	// 	os.Stdin.Read(b)
	// 	if handle(b[0]) {
	// 		log.Println(string(b[0]), "was handled")
	// 		x, y, err := addSquare()
	// 		if err != nil {
	// 			panic(err)
	// 		}
	// 		log.Printf("square at %d %d was generated", x, y)
	// 	}
	// }
}
