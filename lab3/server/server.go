package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/SlyMarbo/rss"
	"github.com/gorilla/websocket"
)

type Client struct {
	conn   *websocket.Conn
	send   chan []byte
	cancel context.CancelFunc
}

func rss_listener(url string, wg *sync.WaitGroup, ch chan<- rss_lenta) {
	defer wg.Done()

	for {
		rssObj, err := rss.Fetch(url)
		if err != nil {
			fmt.Println("Error:", err)
			time.Sleep(10 * time.Second)
			continue
		}
		rss_lenta1 := rss_lenta{rss: rssObj, url: url}
		ch <- rss_lenta1
		time.Sleep(10 * time.Second)
	}
}

func newsProcessor(ch <-chan rss_lenta, hub *Hub) {
	seen := make(map[string]bool)

	for lenta := range ch {
		for _, item := range lenta.rss.Items {
			if !seen[item.Link] {
				seen[item.Link] = true
				htmlstring := fmt.Sprintf("<a href=\"%v\">%v</a><br><br>Описание: %v<br><br>", item.Link, item.Title, item.Summary)
				htmlstring = strings.ReplaceAll(htmlstring, "\n", " ")
				htmlstring = strings.ReplaceAll(htmlstring, "\r", " ")
				hub.Broadcast(lenta.url, []byte(htmlstring))
			}
		}
	}
}

type rss_lenta struct {
	rss *rss.Feed
	url string
}

type Hub struct {
	mu      sync.RWMutex
	topics  map[string]map[*Client]bool
	history map[string][][]byte
}

func (h *Hub) AddClient(topic string, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.topics[topic] == nil {
		h.topics[topic] = make(map[*Client]bool)
	}
	h.topics[topic][c] = true

	for _, msg := range h.history[topic] {
		c.send <- msg
	}
}

func (h *Hub) Broadcast(topic string, message []byte) {
	h.mu.Lock()

	h.history[topic] = append(h.history[topic], message)

	const maxHistory = 10
	if len(h.history[topic]) > maxHistory {
		h.history[topic] = h.history[topic][len(h.history[topic])-maxHistory:]
	}

	var slowClients []*Client
	for client := range h.topics[topic] {
		select {
		case client.send <- message:
		default:
			slowClients = append(slowClients, client)
		}
	}

	for _, c := range slowClients {
		delete(h.topics[topic], c)
		close(c.send)
	}
	h.mu.Unlock()
}

func mockNewsGenerator(topic string, label string, hub *Hub) {
	adjectives := []string{"Срочная", "Неожиданная", "Важная", "Техническая", "Сенсационная", "Аналитическая"}
	nouns := []string{"новость", "сводка", "информация", "статья", "разработка", "статистика"}
	verbs := []string{"появилась", "опубликована", "обновлена", "проанализирована", "проверена"}

	counter := 1
	for {
		time.Sleep(time.Duration(2+rand.Intn(3)) * time.Second)

		adj := adjectives[rand.Intn(len(adjectives))]
		noun := nouns[rand.Intn(len(nouns))]
		verb := verbs[rand.Intn(len(verbs))]
		timestamp := time.Now().Format("15:04:05")

		snippet := fmt.Sprintf(
			"<strong>[%s #%d] %s %s</strong><br>Статус: успешно %s в %s.<br><br>",
			label, counter, adj, noun, verb, timestamp,
		)

		hub.Broadcast(topic, []byte(snippet))
		counter++
	}
}

func sseHandler(h *Hub, topic string, w http.ResponseWriter, r *http.Request) {
	client := &Client{
		send: make(chan []byte, 16),
	}
	h.AddClient(topic, client)
	defer h.RemoveClient(topic, client)

	client.ServeSSE(w, r)
}

func (c *Client) ServeSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	for {
		select {
		case htmlSnippet, ok := <-c.send:
			if !ok {
				return
			}

			fmt.Fprintf(w, "data: %s\n\n", htmlSnippet)
			flusher.Flush()

		case <-r.Context().Done():
			return
		}
	}
}

func (h *Hub) RemoveClient(topic string, c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.topics[topic] == nil {
		return
	}

	if _, ok := h.topics[topic][c]; ok {
		delete(h.topics[topic], c)
		close(c.send)
	}
}

func main() {
	hub := Hub{
		topics:  make(map[string]map[*Client]bool),
		history: make(map[string][][]byte),
	}

	ch := make(chan rss_lenta, 8)
	var wg sync.WaitGroup

	go rss_listener("https://news.mail.ru/rss/90/", &wg, ch)
	go rss_listener("https://lenta.ru/rss", &wg, ch)
	go newsProcessor(ch, &hub)

	go mockNewsGenerator("mock", "Live Demo", &hub)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "index.html")
	})

	http.HandleFunc("/feed", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "feed.html")
	})

	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		topic := r.URL.Query().Get("topic")
		sseHandler(&hub, topic, w, r)
	})

	http.HandleFunc("/dashboard", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "dashboard.html")
	})

	fmt.Println("Сервер запущен на http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Ошибка сервера:", err)
	}
}
