Client:
--dependancies: tkinter, sys, time..
--build: pyinstaller
--default: exe file, network.config file
network.config as DST_IP, DST_PORT
--run via python client.py DST_IP DST_PORT
    --accepts hostnames and in case of invalid arguments usees network.config
Server:
--build: go build
--default: exe file, network.config file
--run: exe file, optional flag: -port=n

Protocol:
general format: /[commad]|[possible_param(s)]

S: /roulette-server: signals the client they are connected to the server supporting protocol

C:/nick|[name]: setting a player nickname, mandatory first command after connection is estabilished
S:/nick-set|[name]| - confirming nick setting.

C:/join|[roomName] - joins the room if it exists, otherwise it creates the new room
S:/room-join|[r.name]| - confirmation of successful room-join

C:/rooms - returns set of all rooms
S:/roomlist[roomNames]|....| - answer for '/rooms' command

C:/msg|[text] - relict from chat server, could be used to implement further functionality, broadcast of [text] in the room

C:/quit - closes connection with the client and server, GUI clientwise its bound to exit button

S:/game-ready - initiating command 3 seconds prior game start
C:/playermove|[yield/risk] - game move of roulette
S:/move-[risk/yield] - confirmation and broadcast of choice
S:/chance|n| - chance 1:n the roulette will kill the participant
S:/valid-turn - token for using /playermove|[yield/risk]
S:/state|[playerName],[playerScore] - updates game state to all client
S:/move-win|[player] - declares winner of game
S:/status-lost|[player] - declares that player lost the connection
S:/status-return|[player] - declares player succesfully reconnected and proceeding with the game
S:/reconnected-room - prepares player to join the room

Error messages if there is a CLI interaction:
/termination-cause-abuse - spam protection
/err-read-invalid-format - not using correct format (length < 20)
/err-nick-set - attempting to rename client
/err-unkown-command - using message totally out of scope of the protocol
/err-illegal-game-move - attempting to bypass game logic

Messages serve informative purpouse or CLI instrucion triggers
Game state updates(examplary):
    [nick] joined the room
    [nick] decided to risk
    [nick] decided to yield
    [nick] reached 1000 points and is a winner
    [nick] got shot

States:
1: Client-server connection: connection is estabilished, only acceptable approach is /nick [name] command from client

2: Client lobby: lobby is the only room without game instance

3: Pending room: room, where only one player (owner) is connected, waiting for other player to join

4: Game instance: instance of an active game
4.1: running game loop
4.2: finished game, waiting for all players to leave and terminating the room afterwards

1->1 scenario: client did not use the desired command or the name was occupied
1->2 scenario: client used '/nick [name]' command succesfully

2->3 scenario: player uses '/join [roomName]' command succesfully
3->2, 3->3 command: player usees '/join [roomName]' command for other room succesfully (can be lobby room)

3->4 scenario: second player joined the room and game has begun
4.1->4.2 scenario: one player won the game (also can occur after other player left without reconnecting) -> connected player won
1->4 scenario: reconnecting of player into the running instance of a game

1,2,3,4-> -1 scenario: client uses /quit command and terminates the connection (or the connection is terminated via other means)