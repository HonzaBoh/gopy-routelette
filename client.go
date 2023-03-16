// File: client.go
package main

import (
	"bufio"
	"net"
	"strconv"
	"strings"
	"time"
)

type client struct {
	conn      net.Conn
	name      string
	room      *room
	isNickSet bool
	gameScore int //i dont like it but prevents keeping score in two identical arrays
	command   chan string
	isTurn    bool
	badFaith  int
}

func (c *client) read() {
	defer c.conn.Close()
	defer func() {
		if c.room != nil {
			if c.room.name != "lobby" {
				c.room.broadcast("/status-lost|" + c.name + "|")
			}
			if c.room.roulette != nil {
				c.room.broadcast(c.name + " has 20 seconds to reconnect")
				duration := time.Second * 20
				time.Sleep(duration)
				var index int
				for i, searched := range c.room.roulette.players {
					if searched == c {
						index = i
						break
					}
				}
				if len(c.room.clients) == 2 { //there is no new reconnected player detected
					c.room.roulette.players[index].gameScore = -1
					c.handleCommand("/playermove|yield")
				}
			}
			c.name = "DeadClient-" + c.name
			c.room.removeClient(c)
		}
	}()
	for {
		if c.badFaith > 30 {
			c.conn.Write([]byte("Terminating for badFaith index of spam > 30\n"))
			c.conn.Write([]byte("/termination-cause-abuse\n"))
			c.conn.Close()
		}
		message, err := bufio.NewReader(c.conn).ReadString('\n')
		if err != nil {
			c.conn.Write([]byte("/err-read-invalid-format\n"))
			c.badFaith++
			break
		}
		if len(message) > 20 {
			c.conn.Write([]byte("Message is too long, please keep it under 20 characters.\n"))
			c.conn.Write([]byte("/err-invalid-format\n"))
			c.badFaith++
			continue
		}
		if strings.HasPrefix(message, "/") {
			c.handleCommand(strings.TrimRight(message, "\r\n"))
		} else if c.room != nil && c.isNickSet {
			if strings.HasPrefix(message, "/msg ") {
				c.room.broadcast(c.name + ": " + strings.TrimPrefix(message, "/msg "))
			} else {
				c.conn.Write([]byte("Wrong message format\n"))
				c.conn.Write([]byte("/err-invalid-format\n"))
				c.badFaith++
			}
		} else if !c.isNickSet {
			c.conn.Write([]byte("Please set nickname first. Use command '/nick yourNick'\n"))
			c.conn.Write([]byte("/err-nick-set'\n"))
			c.badFaith++
		} else {
			c.conn.Write([]byte("Wrong message format\n"))
			c.conn.Write([]byte("/err-invalid-format\n"))
			c.badFaith++
		}
	}
}

func (c *client) handleCommand(message string) {
	parts := strings.Split(message, "|")
	switch parts[0] {
	case "/nick":
		if c.isNickSet {
			c.conn.Write([]byte("You already set your name " + c.name + "!\n"))
			c.conn.Write([]byte("/err-nick-change\n"))
			c.badFaith++
		} else if len(parts) == 2 {
			c.name = parts[1]
			if findClientByNickname(c.name) == nil {
				c.isNickSet = true
				c.conn.Write([]byte("/nick-set|" + c.name + "|\n"))
				if c.room == nil {
					c.room = rooms["lobby"]
					c.room.addClient(c)
				}
			} else {
				c.conn.Write([]byte("/nick-set|" + c.name + "|\n"))
				c.restoreFrom(findClientByNickname(c.name))
			}
		}
	case "/rooms":
		if c.isNickSet {
			var roomList string
			roomList += "/roomlist"
			for _, r := range rooms {
				if r.name != "lobby" && len(r.clients) < 2 {
					roomList += "|" + r.name
				}
			}
			roomList += "|\n"
			c.conn.Write([]byte(roomList))
		} else {
			c.conn.Write([]byte("/err-nick-set'\n"))
			c.conn.Write([]byte("Please set nickname first. Use command '/nick|yourNick'\n"))
			c.badFaith++
		}
	case "/join":
		if c.isNickSet {
			if len(parts) == 2 {
				if r, ok := rooms[parts[1]]; ok {
					c.room.removeClient(c)
					r.addClient(c)
				} else {
					r := &room{
						name:    parts[1],
						clients: make(map[*client]struct{}),
					}
					if c.room != nil {
						c.room.removeClient(c)
					}
					r.addClient(c)
					rooms[r.name] = r
				}
			}
		} else {
			c.conn.Write([]byte("Please set nickname first. Use command '/nick|yourNick'\n"))
			c.conn.Write([]byte("/err-nick-set'\n"))
			c.badFaith++
		}
	case "/msg":
		if c.isNickSet {
			if len(parts) > 1 {
				c.room.broadcast(c.name + ": " + strings.Join(parts[1:], " "))
			}
		} else {
			c.conn.Write([]byte("Please set nickname first. Use command '/nick|yourNick'\n"))
			c.conn.Write([]byte("/err-nick-set'\n"))
			c.badFaith++
		}
	case "/quit":
		c.conn.Close()
	case "/playermove":
		if c.isTurn {
			c.command <- message
		} else {
			c.conn.Write([]byte("It is not your turn, cheater\n"))
			c.conn.Write([]byte("/err-illegal-game-move'\n"))
			c.conn.Write([]byte("/err-nick-set'\n"))
		}
	default:
		if c.isNickSet {
			c.conn.Write([]byte("Unknown command\n"))
			c.conn.Write([]byte("/err-unkown-command'\n"))
			c.badFaith++
		} else {
			c.conn.Write([]byte("Please set nickname first. Use command '/nick|yourNick'\n"))
			c.conn.Write([]byte("/err-nick-set'\n"))
			c.badFaith++
		}
	}
}

func findClientByNickname(nickname string) *client {
	for _, r := range rooms {
		for c := range r.clients {
			if c.name == nickname {
				return c
			}
		}
	}
	return nil
}

func (newClient *client) restoreFrom(c *client) {
	newClient.name = c.name
	newClient.isNickSet = c.isNickSet
	c.room.clients[newClient] = struct{}{}
	newClient.room = c.room
	newClient.gameScore = c.gameScore
	newClient.isTurn = c.isTurn
	newClient.command = c.command
	if newClient.room.roulette != nil {
		if newClient.room.roulette.players != nil {
			newClient.room.roulette.replaceClient(c, newClient)
		}
	}
	duration := time.Second * 1
	newClient.conn.Write([]byte("you have reconnected room " + newClient.room.name + "\n"))
	time.Sleep(duration)
	newClient.conn.Write([]byte("/reconnect-room\n"))
	time.Sleep(duration * 3)
	if newClient.room.roulette.players[0].name == newClient.name && newClient.room.roulette.players[0].isTurn {
		newClient.conn.Write([]byte("/valid-turn\n"))
	} else if newClient.room.roulette.players[1].name == newClient.name && newClient.room.roulette.players[1].isTurn {
		newClient.conn.Write([]byte("/valid-turn\n"))
	} else {
		newClient.conn.Write([]byte("/move-yield\n"))
	}
	time.Sleep(duration)
	newClient.room.broadcast("/status-return|" + c.name + "|")
	time.Sleep(duration)
	newClient.conn.Write([]byte("/state|" + newClient.room.roulette.players[0].name + "|" + strconv.Itoa(newClient.room.roulette.players[0].gameScore) + "|\n"))
	time.Sleep(duration)
	newClient.conn.Write([]byte("/state|" + newClient.room.roulette.players[1].name + "|" + strconv.Itoa(newClient.room.roulette.players[1].gameScore) + "|\n"))
	time.Sleep(duration)
	newClient.conn.Write([]byte("/chance|" + strconv.Itoa(6-newClient.room.roulette.chamber) + "|"))
	time.Sleep(duration)
	c.conn.Close()
}

func (r *roulette) replaceClient(oldClient, newClient *client) {
	// Find the index of the old client in the players slice
	var index int
	for i, c := range r.players {
		if c == oldClient {
			index = i
			break
		}
	}
	// Replace the old client with the new one
	r.players[index] = newClient
}
