package chat

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Message представляет сообщение чата с флагом прочтения.
type Message struct {
	ChatID    string `json:"chat_id"`
	Sender    string `json:"sender"`
	Text      string `json:"text"`
	Timestamp int64  `json:"timestamp"`
	Read      bool   `json:"read"`
}

// StatusMessage представляет статус чата с онлайн-пользователями.
type StatusMessage struct {
	Type        string   `json:"type"`
	OnlineUsers []string `json:"onlineUsers"`
}

// ChatState представляет состояние чата.
type ChatState struct {
	Messages    []Message `json:"messages"`
	OnlineUsers []string  `json:"onlineUsers"`
}

// ReadStatusUpdate представляет обновление статуса прочтения для сообщения.
type ReadStatusUpdate struct {
	Update    string `json:"update"` // Ожидается значение "read"
	ChatID    string `json:"chat_id"`
	Timestamp int64  `json:"timestamp"`
	Read      bool   `json:"read"`
}

// ChatHub управляет чатами и рассылкой сообщений.
type ChatHub struct {
	chats     map[string]map[*websocket.Conn]string // Храним имя пользователя, а не только client_id.
	messages  map[string][]Message                    // История сообщений по chatID.
	broadcast chan Message                            // Канал для новых сообщений.
	mu        sync.Mutex                              // Синхронизация доступа.
}

// NewChatHub создаёт новый экземпляр ChatHub.
func NewChatHub() *ChatHub {
	return &ChatHub{
		chats:     make(map[string]map[*websocket.Conn]string),
		messages:  make(map[string][]Message),
		broadcast: make(chan Message),
	}
}

// Run обрабатывает входящие сообщения и рассылает их всем клиентам.
func (hub *ChatHub) Run() {
	for msg := range hub.broadcast {
		hub.mu.Lock()
		clients, exists := hub.chats[msg.ChatID]
		if !exists {
			hub.mu.Unlock()
			continue
		}

		// Если время не установлено, устанавливаем его.
		if msg.Timestamp == 0 {
			msg.Timestamp = time.Now().UnixMilli()
		}
		// Если в чате подключено два и более клиента (например, клиент и админ),
		// считаем, что сообщение прочитано.
		if len(clients) >= 2 {
			msg.Read = true
		} else {
			msg.Read = false
		}
		// Сохраняем сообщение в истории.
		hub.messages[msg.ChatID] = append(hub.messages[msg.ChatID], msg)

		// Отправляем сообщение всем клиентам.
		data, err := json.Marshal(msg)
		if err != nil {
			log.Printf("Ошибка маршаллинга сообщения: %v", err)
			hub.mu.Unlock()
			continue
		}
		for client := range clients {
			if err := client.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("Ошибка отправки клиенту в чате %s: %v", msg.ChatID, err)
				client.Close()
				delete(clients, client)
			}
		}

		hub.broadcastStatus(msg.ChatID)
		hub.mu.Unlock()
	}
}

// broadcastStatus отправляет статус (список онлайн-пользователей) в указанный чат.
func (hub *ChatHub) broadcastStatus(chatID string) {
	clients, exists := hub.chats[chatID]
	if !exists {
		return
	}
	onlineUsers := make([]string, 0, len(clients))
	for _, username := range clients {
		onlineUsers = append(onlineUsers, username)
	}
	statusMsg := StatusMessage{
		Type:        "status",
		OnlineUsers: onlineUsers,
	}
	data, err := json.Marshal(statusMsg)
	if err != nil {
		log.Printf("Ошибка маршаллинга статуса: %v", err)
		return
	}
	for client := range clients {
		if err := client.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("Ошибка отправки статуса в чате %s: %v", chatID, err)
			client.Close()
			delete(clients, client)
		}
	}
}

// sendChatState отправляет текущее состояние чата (сообщения и онлайн-пользователи) новому клиенту.
func (hub *ChatHub) sendChatState(conn *websocket.Conn, chatID string) {
	clients := hub.chats[chatID]
	onlineUsers := make([]string, 0, len(clients))
	for _, username := range clients {
		onlineUsers = append(onlineUsers, username)
	}
	state := ChatState{
		Messages:    hub.messages[chatID],
		OnlineUsers: onlineUsers,
	}
	data, err := json.Marshal(state)
	if err != nil {
		log.Printf("Ошибка маршаллинга состояния чата %s: %v", chatID, err)
		return
	}
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Printf("Ошибка отправки состояния чата %s: %v", chatID, err)
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // Для продакшена добавьте проверку Origin.
}

// ChatHandler обрабатывает подключение WebSocket и обмен сообщениями.
func (hub *ChatHub) ChatHandler(w http.ResponseWriter, r *http.Request) {
    // Извлекаем client_id и username
    clientID := r.URL.Query().Get("client_id")
    username := r.URL.Query().Get("username")
    if clientID == "" {
        http.Error(w, "client_id обязателен", http.StatusBadRequest)
        return
    }
    // Если username не передан, используем clientID
    if username == "" {
        username = clientID
    }
    // chat_id теперь необязателен при подключении
    chatID := r.URL.Query().Get("chat_id")
    if chatID == "" {
        chatID = "global" // Используем временный chatID, если не указан
    }

    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("Ошибка апгрейда до WS: %v", err)
        return
    }

    hub.mu.Lock()
    if _, exists := hub.chats[chatID]; !exists {
        hub.chats[chatID] = make(map[*websocket.Conn]string)
        hub.messages[chatID] = []Message{}
    }
    hub.chats[chatID][ws] = username
    hub.mu.Unlock()

    log.Printf("Клиент %s (%s) подключился к чату %s", clientID, username, chatID)
    hub.sendChatState(ws, chatID)
    hub.broadcastStatus(chatID)

    go hub.pingLoop(ws, clientID, chatID)

    for {
        _, data, err := ws.ReadMessage()
        if err != nil {
            log.Printf("Клиент %s отключился из чата %s: %v", clientID, chatID, err)
            hub.mu.Lock()
            delete(hub.chats[chatID], ws)
            if len(hub.chats[chatID]) == 0 {
                delete(hub.chats, chatID)
                delete(hub.messages, chatID)
            }
            hub.broadcastStatus(chatID)
            hub.mu.Unlock()
            break
        }

        // Обработка обновления статуса прочтения
        var potentialUpdate struct {
            Update string `json:"update"`
        }
        if err := json.Unmarshal(data, &potentialUpdate); err == nil && potentialUpdate.Update == "read" {
            var update ReadStatusUpdate
            if err := json.Unmarshal(data, &update); err != nil {
                log.Printf("Ошибка разбора обновления статуса: %v", err)
                continue
            }
            hub.mu.Lock()
            msgs := hub.messages[update.ChatID]
            for i, m := range msgs {
                if m.Timestamp == update.Timestamp {
                    hub.messages[update.ChatID][i].Read = update.Read
                    updatedMsg, err := json.Marshal(hub.messages[update.ChatID][i])
                    if err != nil {
                        log.Printf("Ошибка маршаллинга обновленного сообщения: %v", err)
                        continue
                    }
                    for client := range hub.chats[update.ChatID] {
                        if err := client.WriteMessage(websocket.TextMessage, updatedMsg); err != nil {
                            log.Printf("Ошибка отправки обновленного статуса: %v", err)
                            client.Close()
                            delete(hub.chats[update.ChatID], client)
                        }
                    }
                }
            }
            hub.mu.Unlock()
            continue
        }

        // Обработка текстового сообщения
        var msg Message
        if err := json.Unmarshal(data, &msg); err != nil {
            log.Printf("Ошибка разбора JSON от %s: %v", clientID, err)
            continue
        }

        // Проверяем, что chat_id указан в сообщении
        if msg.ChatID == "" {
            log.Printf("Ошибка: chat_id не указан в сообщении от %s", clientID)
            continue
        }

        hub.mu.Lock()
        senderName := hub.chats[chatID][ws]
        hub.mu.Unlock()
        msg.Sender = senderName
        if msg.Timestamp == 0 {
            msg.Timestamp = time.Now().UnixMilli()
        }

        log.Printf("Получено от %s в чате %s: %s", clientID, msg.ChatID, msg.Text)
        hub.broadcast <- msg
    }
}

// pingLoop отправляет пинг-сообщения для поддержания соединения.
func (hub *ChatHub) pingLoop(conn *websocket.Conn, clientID, chatID string) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		hub.mu.Lock()
		_, exists := hub.chats[chatID][conn]
		hub.mu.Unlock()
		if !exists {
			return
		}
		if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(10*time.Second)); err != nil {
			log.Printf("Ошибка отправки пинга клиенту %s в чате %s: %v", clientID, chatID, err)
			return
		}
	}
}
