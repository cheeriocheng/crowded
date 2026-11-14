package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"math"
	mrand "math/rand/v2"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

const (
	// The shared space every main screen draws. Screens scale it to fit.
	worldW = 1600.0
	worldH = 900.0

	// World units moved per pixel dragged on a phone.
	gain = 2.5
	// Largest single move accepted from a phone, in phone pixels.
	maxStep = 400.0

	// How long a disconnected phone keeps its place before it is removed.
	grace = 30 * time.Second
	// How often main screens receive the combined state.
	tick = time.Second / 30
)

// 7 colours and 6 polygons (3 to 8 sides) give 42 distinct identities.
var colors = []string{
	"#ff5a5f", "#ffb400", "#3ddc97", "#2ec4ff", "#7b6cff", "#ff6ad5", "#f5f5f5",
}

type player struct {
	token  string
	conns  int
	leftAt time.Time

	id    int
	color string
	sides int
	x, y  float64
}

type playerView struct {
	ID     int     `json:"id"`
	Color  string  `json:"color"`
	Sides  int     `json:"sides"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Online bool    `json:"online"`
}

type stateMsg struct {
	T       string       `json:"t"`
	W       float64      `json:"w"`
	H       float64      `json:"h"`
	Players []playerView `json:"players"`
}

type welcomeMsg struct {
	T     string `json:"t"`
	Token string `json:"token"`
	Color string `json:"color"`
	Sides int    `json:"sides"`
}

type hub struct {
	mu      sync.Mutex
	players map[string]*player
	screens map[*websocket.Conn]struct{}
	joined  int
}

func newHub() *hub {
	return &hub{
		players: map[string]*player{},
		screens: map[*websocket.Conn]struct{}{},
	}
}

// run sends the combined state to every main screen at a fixed rate.
func (h *hub) run() {
	t := time.NewTicker(tick)
	defer t.Stop()
	for range t.C {
		h.mu.Lock()
		for token, p := range h.players {
			if p.conns == 0 && time.Since(p.leftAt) > grace {
				delete(h.players, token)
				log.Printf("player %d removed", p.id)
			}
		}
		if len(h.screens) == 0 {
			h.mu.Unlock()
			continue
		}
		msg := stateMsg{T: "state", W: worldW, H: worldH, Players: make([]playerView, 0, len(h.players))}
		for _, p := range h.players {
			msg.Players = append(msg.Players, playerView{
				ID: p.id, Color: p.color, Sides: p.sides, X: p.x, Y: p.y, Online: p.conns > 0,
			})
		}
		screens := make([]*websocket.Conn, 0, len(h.screens))
		for c := range h.screens {
			screens = append(screens, c)
		}
		h.mu.Unlock()

		sort.Slice(msg.Players, func(i, j int) bool { return msg.Players[i].ID < msg.Players[j].ID })
		data, err := json.Marshal(msg)
		if err != nil {
			continue
		}
		for _, c := range screens {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			err := c.Write(ctx, websocket.MessageText, data)
			cancel()
			if err != nil {
				c.CloseNow()
				h.mu.Lock()
				delete(h.screens, c)
				h.mu.Unlock()
			}
		}
	}
}

func (h *hub) handleWS(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("role") == "screen" {
		h.serveScreen(w, r)
		return
	}
	h.servePhone(w, r)
}

func (h *hub) serveScreen(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer c.CloseNow()

	h.mu.Lock()
	h.screens[c] = struct{}{}
	h.mu.Unlock()
	log.Printf("main screen connected")

	// Screens only receive. Wait here until the connection closes.
	<-c.CloseRead(context.Background()).Done()

	h.mu.Lock()
	delete(h.screens, c)
	h.mu.Unlock()
	log.Printf("main screen disconnected")
}

func (h *hub) servePhone(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer c.CloseNow()

	p := h.join(r.URL.Query().Get("token"))
	defer h.leave(p)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Notice phones that vanish without saying goodbye.
	go func() {
		t := time.NewTicker(5 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				pctx, pcancel := context.WithTimeout(ctx, 5*time.Second)
				err := c.Ping(pctx)
				pcancel()
				if err != nil {
					c.CloseNow()
					return
				}
			}
		}
	}()

	welcome := welcomeMsg{T: "welcome", Token: p.token, Color: p.color, Sides: p.sides}
	if err := wsjson.Write(ctx, c, welcome); err != nil {
		return
	}

	for {
		var m struct {
			T  string  `json:"t"`
			DX float64 `json:"dx"`
			DY float64 `json:"dy"`
		}
		if err := wsjson.Read(ctx, c, &m); err != nil {
			return
		}
		if m.T == "move" {
			h.move(p, m.DX, m.DY)
		}
	}
}

// join returns the player for a token, creating one if the token is unknown.
func (h *hub) join(token string) *player {
	h.mu.Lock()
	defer h.mu.Unlock()

	if p, ok := h.players[token]; ok && token != "" {
		p.conns++
		log.Printf("player %d reconnected", p.id)
		return p
	}

	n := h.joined
	h.joined++
	p := &player{
		token: newToken(),
		conns: 1,
		id:    n + 1,
		color: colors[n%len(colors)],
		sides: 3 + n%6,
		x:     worldW * (0.2 + 0.6*mrand.Float64()),
		y:     worldH * (0.2 + 0.6*mrand.Float64()),
	}
	h.players[p.token] = p
	log.Printf("player %d joined", p.id)
	return p
}

func (h *hub) leave(p *player) {
	h.mu.Lock()
	defer h.mu.Unlock()
	p.conns--
	if p.conns == 0 {
		p.leftAt = time.Now()
		log.Printf("player %d disconnected", p.id)
	}
}

// move shifts a player by a drag measured in phone pixels.
func (h *hub) move(p *player, dx, dy float64) {
	dx = clamp(dx, -maxStep, maxStep)
	dy = clamp(dy, -maxStep, maxStep)
	h.mu.Lock()
	defer h.mu.Unlock()
	p.x = clamp(p.x+dx*gain, 0, worldW)
	p.y = clamp(p.y+dy*gain, 0, worldH)
}

func clamp(v, lo, hi float64) float64 {
	return math.Max(lo, math.Min(hi, v))
}

func newToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
