// file room.go
package main

import (
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type room struct {
	name     string
	clients  map[*client]struct{}
	roulette *roulette
}

type roulette struct {
	players []*client
	chamber int
	yields  int
}

func (r *room) addClient(c *client) {
	r.clients[c] = struct{}{}
	c.room = r
	c.conn.Write([]byte("/room-join|" + r.name + "|\n"))
	//r.broadcast(c.name + " has joined!")
	if r.name == "lobby" {
		return //no game instance in the lobby
	}

	if r.roulette == nil {
		r.roulette = &roulette{players: []*client{c}}
		r.clients[c] = struct{}{}
		c.room = r
		return
	}

	// If the roulette struct already exists and there's only one player
	if len(r.roulette.players) == 1 {
		r.roulette.players = append(r.roulette.players, c)
		r.clients[c] = struct{}{}
		c.room = r
		//Begin the game, prepapre basic score
		r.broadcast("/game-ready|" + r.roulette.players[0].name + "|" + r.roulette.players[1].name + "|")
		r.roulette.players[0].gameScore = 0
		r.roulette.players[1].gameScore = 0
		r.roulette.yields = 0
		r.roulette.chamber = 1
		r.broadcast("Starting in 3 seconds")
		duration := time.Second * 3
		time.Sleep(duration) //dramatic pause
		go r.playGame()
		return
	}

	// If there's already 2 players in the roulette
	if len(r.roulette.players) == 2 {
		// Send message to client that the room is full
		c.conn.Write([]byte("The room is full, please try a different room.\n"))
		return
	}
}

func (r *room) playGame() {
	currentPlayer := 0
	duration := time.Second * 1
	r.roulette.players[currentPlayer].isTurn = true
	for {
		if r.roulette.yields == 2 {
			r.broadcast("Reseting the chamber!\n")
			time.Sleep(duration)
			r.roulette.yields = 0
			r.roulette.chamber = 1
		}

		//chance 1 against n
		r.broadcast("/chance|" + strconv.Itoa(6-r.roulette.chamber) + "|")
		time.Sleep(duration)
		// Send message to current player asking for input
		r.roulette.players[currentPlayer].conn.Write([]byte("/valid-turn\n"))
		time.Sleep(duration)
		// Read input from current player
		select {
		case gameCommand := <-r.roulette.players[currentPlayer].command:
			parts := strings.Split(strings.TrimSpace(gameCommand), "|")
			if len(parts) < 2 {
				r.roulette.players[currentPlayer].conn.Write([]byte("Invalid command format. Use '/playermove|[yield/risk]'\n"))
				r.roulette.players[currentPlayer].conn.Write([]byte("/err-unkown-command'\n"))
				r.roulette.players[currentPlayer].badFaith++
				continue
			}
			if parts[0] == "/playermove" {
				if parts[1] != "yield" && parts[1] != "risk" {
					r.roulette.players[currentPlayer].conn.Write([]byte("Invalid command. Use '/playermove|[yield/risk]'\n"))
					r.roulette.players[currentPlayer].conn.Write([]byte("/err-unkown-command'\n"))
					r.roulette.players[currentPlayer].badFaith++
					continue
				}
				// Handle player input
				switch parts[1] {
				case "yield":
					r.roulette.yields++
					r.roulette.players[currentPlayer].conn.Write([]byte("/move-yield\n"))
					time.Sleep(duration)
					r.broadcast(r.roulette.players[currentPlayer].name + " chose to pussy out, next!")
					time.Sleep(duration)
				case "risk":
					roll := rand.Intn(6) + r.roulette.chamber
					r.roulette.players[currentPlayer].conn.Write([]byte("/move-risk\n"))
					if roll >= 6 {
						r.broadcast(r.roulette.players[currentPlayer].name + " got their brains on the ceiling!")
						time.Sleep(duration)
						r.roulette.players[currentPlayer].gameScore = -1
						//r.roulette.yields-- //Obsolete, game over
						//continue
					} else {
						r.broadcast(r.roulette.players[currentPlayer].name + " risked it and lived, next!")
						time.Sleep(duration)
						r.roulette.players[currentPlayer].gameScore += 200
						r.roulette.chamber++
						//blocks potential negative number of yields
						if r.roulette.yields > 0 {
							r.roulette.yields--
						}
					}
				}
			}
			r.broadcast("/state|" + r.roulette.players[currentPlayer].name + "|" + strconv.Itoa(r.roulette.players[currentPlayer].gameScore) + "|")
			time.Sleep(duration)
			//switch turn
			r.roulette.players[currentPlayer].isTurn = false
			r.roulette.players[(currentPlayer+1)%2].isTurn = true
			currentPlayer = (currentPlayer + 1) % 2
		}
		// Check for winner
		if r.gameOver() {
			return
		}
	}
}

func (r *room) gameOver() bool {
	if r.roulette.players[0].gameScore >= 1000 || r.roulette.players[1].gameScore < 0 {
		r.broadcast("/move-win|" + r.roulette.players[0].name)
		r.broadcast("Returning to lobby in 5 seconds...")
		duration := time.Second * 5
		time.Sleep(duration)
		if r.roulette.players[0] != nil {
			r.roulette.players[0].room.removeClient(r.roulette.players[0])
			r.roulette.players[0].room = rooms["lobby"]
			r.roulette.players[0].room.addClient(r.roulette.players[0])
		}
		if r.roulette.players[1] != nil {
			r.roulette.players[1].room.removeClient(r.roulette.players[1])
			r.roulette.players[1].room = rooms["lobby"]
			r.roulette.players[1].room.addClient(r.roulette.players[1])
		}

		return true
	} else if r.roulette.players[1].gameScore >= 1000 || r.roulette.players[0].gameScore < 0 {
		r.broadcast("/move-win|" + r.roulette.players[1].name)
		r.broadcast("Returning to lobby in 5 seconds...")
		duration := time.Second * 5
		time.Sleep(duration)
		if r.roulette.players[0] != nil {
			r.roulette.players[0].room.removeClient(r.roulette.players[0])
			r.roulette.players[0].room = rooms["lobby"]
			r.roulette.players[0].room.addClient(r.roulette.players[0])
		}
		if r.roulette.players[1] != nil {
			r.roulette.players[1].room.removeClient(r.roulette.players[1])
			r.roulette.players[1].room = rooms["lobby"]
			r.roulette.players[1].room.addClient(r.roulette.players[1])
		}
		return true
	}
	return false
}

func (r *room) removeClient(c *client) {
	if r == nil || c == nil {
		return
	}
	delete(r.clients, c)
	c.room = nil
	//r.broadcast(c.name + " has left the room.")
	if len(r.clients) == 0 && r.name != "lobby" {
		delete(rooms, r.name)
	}
}

func (r *room) broadcast(message string) {
	for c := range r.clients {
		c.conn.Write([]byte(message + "\n"))
	}
}
