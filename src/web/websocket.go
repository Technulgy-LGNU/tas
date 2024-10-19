package web

import (
	"github.com/gofiber/websocket/v2"
	"log"
)

func (a *API) WebsocketConnection(c *websocket.Conn) {
	a.Clients[c] = true

	for {
		_, p, err := c.ReadMessage()
		log.Println(string(p))
		if err != nil {
			break
		}
	}
}

func (a *API) SendMessage(msg []byte) {
	for client := range a.Clients {
		err := client.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			if err := client.Close(); err != nil {
				log.SetFlags(log.LstdFlags | log.Lshortfile)
				log.Printf("Unable to close websocket connection: %d\n", err)
			}
			log.SetFlags(log.LstdFlags | log.Lshortfile)
			log.Printf("Error sending message: %d\n", err)
			log.Println("Deleting client ...")
			delete(a.Clients, client)
		}
	}
}
