package handlers

import (
	"fmt"
	"github.com/CloudyKit/jet/v6"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sort"
)

var views = jet.NewSet(jet.NewOSFileSystemLoader("./html"), jet.InDevelopmentMode())

var upgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var wsChan = make(chan WsPayload)
var clients = make(map[WebSocketConnection]string)

func Home(w http.ResponseWriter, r *http.Request) {

	if err := renderPage(w, "home.htm", nil); err != nil {
		log.Println(err)
	}
}

type WebSocketConnection struct {
	*websocket.Conn
}
type WsJsonResponse struct {
	Action         string   `json:"action"`
	Message        string   `json:"message"`
	MessageType    string   `json:"message_type"`
	ConnectedUsers []string `json:"connected_users"`
}
type WsPayload struct {
	Action   string              `json:"action"`
	Message  string              `json:"message"`
	Username string              `json:"username"`
	Conn     WebSocketConnection `json:"-"`
}

func WsEndpoint(w http.ResponseWriter, r *http.Request) {
	wsConnection, err := upgradeConnection.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	var response WsJsonResponse
	response.Message = `<em><small>Connected to server</small></em>`

	conn := WebSocketConnection{Conn: wsConnection}
	clients[conn] = ""
	err = wsConnection.WriteJSON(response)
	if err != nil {
		log.Println(err)
	}

	go ListenForWs(&conn)
	log.Println("Client Connected to endpoint")

}
func ListenForWs(conn *WebSocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("error", fmt.Sprintf("%v", r))
		}
	}()

	var payload WsPayload

	for {
		err := conn.ReadJSON(&payload)
		if err != nil {
			log.Println(err)
		} else {
			payload.Conn = *conn
			wsChan <- payload
		}
	}
}

func ListenToWsChannel() {
	var response WsJsonResponse
	for {
		event := <-wsChan
		switch event.Action {
		case "username":
			clients[event.Conn] = event.Username
			users := getUserList()
			response.Action = "list_users"
			response.ConnectedUsers = users
			broadCastToAll(response)
		case "left":
			response.Action = "list_users"
			delete(clients, event.Conn)
			users := getUserList()
			response.ConnectedUsers = users
			broadCastToAll(response)
		case "broadcast":
			response.Action = "broadcast"
			response.Message = fmt.Sprintf("<strong>%s : </strong> %s", event.Username, event.Message)
			broadCastToAll(response)
		}
		//response.Action = "Got here"
		//response.Message = fmt.Sprintf("Some message and action was %s", event.Action)
		//broadCastToAll(response)
	}
}

func getUserList() []string {
	var userList []string
	for _, value := range clients {
		if value != "" {
			userList = append(userList, value)
		}

	}
	sort.Strings(userList)
	return userList
}

func broadCastToAll(response WsJsonResponse) {
	for client := range clients {
		err := client.WriteJSON(response)
		if err != nil {
			log.Println("web socket err")
			_ = client.Close()
			delete(clients, client)
		}
	}
}
func renderPage(w http.ResponseWriter, tmpl string, data jet.VarMap) error {
	view, err := views.GetTemplate(tmpl)
	if err != nil {
		log.Println(err)
		return err
	}
	if err = view.Execute(w, data, nil); err != nil {
		log.Println(err)
		return err
	}
	return nil
}
