package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	typesNotification "github.com/timmo001/letmeknow/server/types/notification"
	types "github.com/timmo001/letmeknow/server/types/websocket"
)

// TODO: Add user authentication, so only authenticated users can send messages
// TODO: Check if user is allowed to send messages to other users

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin
		return true
	},
}

type ConnectedClients []types.Client

var connectedClients ConnectedClients

func (cc ConnectedClients) Display() []string {
	var clients []string
	for _, client := range cc {
		clients = append(clients, client.Display())
	}
	return clients
}

func WebSocket(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	// Add client to connectedClients
	connectedClients = append(connectedClients,
		types.Client{
			Connection: c,
		},
	)
	log.Println("Client connected:", c.RemoteAddr())

	log.Println("Connected clients:", connectedClients.Display())

	for {
		mt, messageIn, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("Recv: %s", messageIn)

		// Parse JSON
		var request map[string]interface{}
		err = json.Unmarshal(messageIn, &request)
		if err != nil {
			log.Println("Error parsing JSON:", err)
			// Send error message
			errString := err.Error()
			resp := types.ResponseError{
				Type:    "error",
				Message: "Error parsing JSON",
				Error:   &errString,
			}
			message, err := json.Marshal(resp)
			if err != nil {
				log.Println("Error marshalling JSON:", err)
				break
			}
			c.WriteMessage(mt, message)
			break
		}

		// Validate JSON contains type
		if _, ok := request["type"]; !ok {
			log.Println("Error: JSON does not contain type")
			// Send error message
			resp := types.ResponseError{
				Type:    "error",
				Message: "Error: JSON does not contain type",
			}
			message, err := json.Marshal(resp)
			if err != nil {
				log.Println("Error marshalling JSON:", err)
				break
			}
			c.WriteMessage(mt, message)
			break
		}

		// If type is not "register" or "notification", send error message
		if request["type"] != "register" && request["type"] != "notification" {
			log.Println("Error: JSON type is not 'register' or 'notification'")
			// Send error message
			resp := types.ResponseError{
				Type:    "error",
				Message: "Error: JSON type is not 'register' or 'notification'",
			}
			message, err := json.Marshal(resp)
			if err != nil {
				log.Println("Error marshalling JSON:", err)
				break
			}
			c.WriteMessage(mt, message)
			break
		}

		// If type is "register", register client
		if request["type"] == "register" {
			// Validate JSON contains userID
			if _, ok := request["userID"]; !ok {
				log.Println("Error: JSON does not contain userID")
				// Send error message
				resp := types.ResponseError{
					Type:    "error",
					Message: "Error: JSON does not contain userID",
				}
				message, err := json.Marshal(resp)
				if err != nil {
					log.Println("Error marshalling JSON:", err)
					break
				}
				c.WriteMessage(mt, message)
				break
			}

			// Convert request to ClientRegistration
			clientRegistration := types.RequestRegister{
				UserID: request["userID"].(string),
			}

			// Set userID for client
			alreadyRegistered := false
			for i, client := range connectedClients {
				if client.Connection == c {
					// Check if userID is already registered
					if client.UserID != nil {
						log.Println("Error: Client already registered with userID:", *client.UserID)
						alreadyRegistered = true
						break
					}

					connectedClients[i].UserID = &clientRegistration.UserID
					break
				}
			}

			log.Println("Connected clients:", connectedClients.Display())

			// Send success message
			var resp types.ResponseSuccess
			if alreadyRegistered {
				resp = types.ResponseSuccess{
					Type:      "register",
					Succeeded: false,
					Message:   "Client already registered",
				}
			} else {
				resp = types.ResponseSuccess{
					Type:      "register",
					Succeeded: true,
					Message:   "Client registered",
				}
			}
			message, err := json.Marshal(resp)
			if err != nil {
				log.Println("Error marshalling JSON:", err)
				break
			}
			c.WriteMessage(mt, message)
			continue
		}

		// Request type is "notification"

		// Check if client is registered with a userID
		var clientRegistered bool = false
		for _, client := range connectedClients {
			if client.Connection == c && client.UserID != nil {
				clientRegistered = true
				break
			}
		}
		if !clientRegistered {
			log.Println("Error: Client not registered")
			// Send error message
			resp := types.ResponseError{
				Type:    "error",
				Message: "Error: Client not registered",
			}
			message, err := json.Marshal(resp)
			if err != nil {
				log.Println("Error marshalling JSON:", err)
				break
			}
			c.WriteMessage(mt, message)
			continue
		}

		// Validate JSON contains message
		if _, ok := request["data"]; !ok {
			log.Println("Error: JSON is not of type Notification")
			// Send error message
			resp := types.ResponseError{
				Type:    "error",
				Message: "Error: JSON is not of type Notification",
			}
			message, err := json.Marshal(resp)
			if err != nil {
				log.Println("Error marshalling JSON:", err)
				break
			}
			c.WriteMessage(mt, message)
			break
		}

		var title *string
		var subtitle *string
		var content *string

		requestData := request["data"].(map[string]interface{})
		log.Println("Request data:", requestData)
		if _, ok := requestData["title"]; ok {
			if t, ok := requestData["title"].(string); ok {
				title = &t
			}
		}
		if _, ok := requestData["subtitle"]; ok {
			if s, ok := requestData["subtitle"].(string); ok {
				subtitle = &s
			}
		}
		if _, ok := requestData["content"]; ok {
			if c, ok := requestData["content"].(string); ok {
				content = &c
			}
		}

		var image *typesNotification.Image
		if _, ok := requestData["image"]; ok {
			if i, ok := requestData["image"].(map[string]interface{}); ok {
				var url string

				if _, ok := i["url"]; ok {
					url = i["url"].(string)
				}

				image = &typesNotification.Image{
					URL: url,
				}
			}
		}

		// Convert request to Notification
		notification := types.RequestNotification{
			Data: typesNotification.Notification{
				Type:     "notification",
				Title:    title,
				Subtitle: subtitle,
				Content:  content,
				Image:    image,
			},
			Targets: []string{},
		}

		if _, ok := request["targets"]; ok {
			targets := request["targets"].([]interface{})
			for _, target := range targets {
				notification.Targets = append(notification.Targets, target.(string))
			}
		}

		// Prepare message to send to clients
		messageOut, err := json.Marshal(notification.Data)
		if err != nil {
			log.Println("Error marshalling JSON:", err)
			break
		}

		// Send message to all clients
		for _, client := range connectedClients {
			// Only send message to clients that are requested
			if notification.Targets != nil && len(notification.Targets) > 0 {
				found := false
				for _, target := range notification.Targets {
					if (strings.HasSuffix(target, "*") && strings.HasPrefix(*client.UserID, target[:len(target)-1])) || target == *client.UserID {
						found = true
						break
					}
					if target == *client.UserID {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			} else {
				// Don't send message to the client that sent it, if sending to all clients
				if client.Connection == c {
					continue
				}
			}

			err = client.Connection.WriteMessage(mt, messageOut)
			if err != nil {
				log.Println("Error writing message to client:", err)
				break
			}
		}

		// Send success message
		resp := types.ResponseSuccess{
			Type:      "notificationSent",
			Succeeded: true,
			Message:   "Message sent",
		}
		messageSuccess, err := json.Marshal(resp)
		if err != nil {
			log.Println("Error marshalling JSON:", err)
			break
		}

		// Send success message to client
		err = c.WriteMessage(mt, messageSuccess)
		if err != nil {
			log.Println("Error writing message:", err)
			break
		}
	}

	log.Println("Client disconnected:", c.RemoteAddr())

	// Remove client from connectedClients
	for i, client := range connectedClients {
		if client.Connection == c {
			connectedClients = append(connectedClients[:i], connectedClients[i+1:]...)
		}
	}

	log.Println("Connected clients:", connectedClients.Display())
}
