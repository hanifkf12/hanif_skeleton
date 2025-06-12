package app

import (
	"fmt"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"log"
)

type test func(c *websocket.Conn)

func (app *App) SetupSocket() {
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}

		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/:id", websocket.New(func(c *websocket.Conn) {
		// c.Locals is added to the *websocket.Conn
		log.Println(c.Locals("allowed"))  // true
		log.Println(c.Params("id"))       // 123
		log.Println(c.Query("v"))         // 1.0
		log.Println(c.Cookies("session")) // ""

		id := c.Params("id")
		// websocket.Conn bindings https://pkg.go.dev/github.com/fasthttp/websocket?tab=doc#pkg-index
		var (
			mt  int
			msg []byte
			err error
		)
		for {
			if mt, msg, err = c.ReadMessage(); err != nil {
				log.Println("read:", err)
				break
			}
			log.Printf("recv: %s", msg)
			log.Printf("RT: %d", mt)
			log.Printf("ID: %s", id)

			if err = c.WriteMessage(mt, []byte(fmt.Sprintf("FROM SERVER %s", msg))); err != nil {
				log.Println("write:", err)
				break
			}
		}

	}))

}
