import socket
import threading
from queue import Queue
from tkinter import *
import time
import sys

send_queue = Queue()
game_window = None
game_label = None
player_one_label = None
player_two_label = None
player_one_cross = None
player_two_cross = None
risk_button = None
yield_button = None
chance = ""
nickname = ""
room_name = ""
roulette_server_initiated = False


def read_str():
    while True:
        try:
            msg = sock.recv(1024).decode().strip()
            if not msg:
                print("Server disconnected")
                exit()
            print("Received:", msg)
            command_switch(msg)
        except:
            print("Error occurred - client")
            display_err("Connection lost. Please, relaunch client.")


def display_err(msg):
    err_window = Toplevel()
    res = Label(err_window, text=msg, font=("Helvetica", font_size))
    res.pack()
    time.sleep(5)
    root.destroy()
    sys.exit(1)


def send_str():
    while True:
        # Get the next command from the queue
        command = send_queue.get()
        # Send the command to the server
        sock.send(f"{command}\n".encode())
        send_queue.task_done()


def submit_name(name):
    global nickname
    nickname = name
    print("Submitting: /nick|" + name)
    send_queue.put("/nick|" + name)


def submit_room(name):
    global room_name
    room_name = name
    print("Submitting:/join|" + name)
    send_queue.put("/join|" + name)


def open_game_window():
    global game_window, game_label
    # Open a new window for the game
    game_window = Toplevel()
    game_window.protocol("WM_DELETE_WINDOW", close_game_window)
    game_window.title("Roulette Game: " + room_name)
    game_window.config(background="#3c3c3c")
    game_window.geometry("1080x540")
    game_window.resizable(height=False, width=False)
    game_label = Label(game_window, text="Welcome to the Roulette, waiting for an opponent..", font=('Helvetica', 32),
                       relief=RAISED, bd=5, fg='white', bg='#3c3c3c', padx=20, pady=15)
    game_label.pack()
    # Put the current menu window to sleep
    root.withdraw()


# noinspection PyUnresolvedReferences
def make_game(players_string):
    global player_one_label, player_two_label, player_one_cross, chance, player_two_cross, game_label, risk_button, \
        yield_button
    player_list = players_string.split("|")

    game_label.config(text="Game begins in 3 seconds...")

    player_one_label = Label(game_window, text=player_list[1] + ", 0", font=('Helvetica', 20), fg='white', bg='#3c3c3c')
    player_one_label.place(x=20, y=180)

    player_two_label = Label(game_window, text=player_list[2] + ", 0", font=('Helvetica', 20), fg='white', bg='#3c3c3c')
    player_two_label.place(x=895, y=180)

    player_one_cross = Label(game_window, text="+", font=('Helvetica', 200),
                             fg='#edbd6b', bg='#edbd6b')
    player_one_cross.place(x=20, y=220)
    player_two_cross = Label(game_window, text="+", font=('Helvetica', 200),
                             fg='#edbd6b', bg='#edbd6b')
    player_two_cross.place(x=895, y=220)

    risk_button = Button(game_window, text="Risk!", font=('Helvetica', 20, 'bold'), bg='green', fg='black',
                         command=submit_risk, state=DISABLED)
    risk_button.place(x=350, y=460)
    yield_button = Button(game_window, text="Yield!", font=('Helvetica', 20, 'bold'), bg='yellow', fg='black',
                          command=submit_yield, state=DISABLED)
    yield_button.place(x=650, y=460)

    chance = Label(game_window, text="Death chance: 1 : N", font=('Helvetica', 20), fg='white', bg='#3c3c3c')
    chance.place(x=420, y=250)
    time.sleep(2)
    game_label.config(text="Game started!")


def submit_yield():
    send_queue.put("/playermove|yield")


def submit_risk():
    send_queue.put("/playermove|risk")


# noinspection PyUnresolvedReferences
def close_game_window():
    # Close the game window
    game_window.destroy()
    # Make the menu window visible again
    root.deiconify()


# noinspection PyUnresolvedReferences
def command_switch(cmd):
    global roulette_server_initiated
    if "/nick-set" in cmd:
        nickname_submit_button.config(state=DISABLED)
        nickname_entry.config(state=DISABLED)
        room_entry.config(state=NORMAL)
        room_submit_button.config(state=NORMAL)
        reload_rooms_button.config(state=NORMAL)
    elif "/roulette-server" in cmd:
        roulette_server_initiated = True
    elif ("/room-join" in cmd) and ("lobby" not in cmd):
        open_game_window()
    elif "/roomlist" in cmd:
        update_room_list(cmd)
    elif "/game-ready" in cmd:
        make_game(cmd)
    elif "/chance|" in cmd:
        death_chance = cmd.split("|")
        chance.config(text="Death chance: 1 : " + death_chance[1])
    elif "/valid-turn" in cmd:
        if nickname in player_one_label.cget("text"):
            player_two_cross.config(fg='#edbd6b')
            player_one_cross.config(fg='red')
        else:
            player_one_cross.config(fg='#edbd6b')
            player_two_cross.config(fg='red')

        yield_button.config(state=NORMAL)
        risk_button.config(state=NORMAL)
    elif ("/move-risk" in cmd) or ("/move-yield" in cmd):
        if nickname in player_one_label.cget("text"):
            player_one_cross.config(fg='#edbd6b')
            player_two_cross.config(fg='red')
        else:
            player_two_cross.config(fg='#edbd6b')
            player_one_cross.config(fg='red')

        yield_button.config(state=DISABLED)
        risk_button.config(state=DISABLED)
    elif "/state|" in cmd:
        state = cmd.split("|")
        if state[1] in player_one_label.cget("text"):
            player_one_label.config(text=state[1] + ", " + state[2])
        else:
            player_two_label.config(text=state[1] + ", " + state[2])

        game_label.config(text=state[1] + " now has " + state[2] + " points.")
    elif "/move-win|" in cmd:
        state = cmd.split("|")
        if state[1] in player_one_label.cget("text"):
            player_two_cross.config(text=':(')
            player_one_cross.config(fg='green', bg='green')
            player_two_cross.config(fg='red')
        else:
            player_one_cross.config(text=':(')
            player_two_cross.config(fg='green', bg='green')
            player_one_cross.config(fg='red')

        game_label.config(text=state[1] + " has won! You can quit the window")
    elif "/status-lost|" in cmd:
        state = cmd.split("|")
        if state[1] in player_one_label.cget("text"):
            player_one_cross.config(bg='yellow')
        else:
            player_two_cross.config(bg='yellow')

        game_label.config(text=state[1] + " disconnected, waiting 20 seconds.")
    elif "/status-return|" in cmd:
        state = cmd.split("|")
        if state[1] in player_one_label.cget("text"):
            player_one_cross.config(bg='#edbd6b')
        else:
            player_two_cross.config(bg='#edbd6b')
        if nickname not in cmd:
            game_label.config(text=state[1] + " came back")
    elif "/reconnect-room" in cmd:
        open_game_window()
        make_game("/game-ready|TMP|TMP|")
        player_one_label.config(text=nickname)


def update_room_list(room_list_string):
    # split the string on '-'
    room_list = room_list_string.split("|")
    # remove the first element, which is '/room list'
    room_list.pop(0)
    # clear the current listbox
    rooms_listbox.delete(0, END)
    # insert the room names into the listbox
    for room in room_list:
        rooms_listbox.insert(END, room)


def go_double_clicked_room(event):
    selected_room = rooms_listbox.get(rooms_listbox.curselection())
    room_entry.config(text=selected_room)
    print("Double clicked room:", selected_room)
    submit_room(selected_room)


if len(sys.argv) == 3:
    try:
        host = sys.argv[1]
        port = int(sys.argv[2])
        sock = socket.socket()
        result = sock.connect_ex((host, port))
    except:
        print("Invalid port and/or IP, using default network.config file for connection config")
        with open("network.config", "r") as config:
            host = config.readline().strip()
            print(host)
            port = int(config.readline().strip())
            print(port)
            sock = socket.socket()
            result = sock.connect_ex((host, port))
else:
    with open("network.config", "r") as config:
        host = config.readline().strip()
        print(host)
        port = int(config.readline().strip())
        print(port)
        sock = socket.socket()
        result = sock.connect_ex((host, port))
font_size = "20"
if result != 0:
    print("Failed to connect to server. Error code:", result)
    sock.close()
else:
    print("Connected to server")

    receive_thread = threading.Thread(target=read_str)
    receive_thread.start()

    write_thread = threading.Thread(target=send_str)
    write_thread.start()
    print("Initiating")
    time.sleep(3)
    if not roulette_server_initiated:
        print("Connected server is not roulette game protocol server, terminating...")
        sys.exit(1)

    root = Tk()
    root.protocol("WM_DELETE_WINDOW", lambda: send_queue.put("/quit"))
    root.title("Roulette")

    nickname_label = Label(root, text="Enter your nickname:", font=("Helvetica", font_size))
    nickname_label.pack()

    nickname_entry = Entry(root, width=20, font=("Helvetica", font_size))
    nickname_entry.pack()

    nickname_submit_button = Button(root, text="Submit Nickname", font=("Helvetica", font_size),
                                    command=lambda: submit_name(nickname_entry.get()))
    nickname_submit_button.pack()

    room_label = Label(root, text="Enter room name:", font=("Helvetica", font_size))
    room_label.pack()

    room_entry = Entry(root, width=20, font=("Helvetica", font_size), state=DISABLED)
    room_entry.pack()

    room_submit_button = Button(root, text="Submit Room", font=("Helvetica", font_size), state=DISABLED,
                                command=lambda: submit_room(room_entry.get()))
    room_submit_button.pack()

    rooms_label = Label(root, text="Rooms:", font=("Helvetica", font_size))
    rooms_label.pack()

    # Increase the width of the listbox
    rooms_listbox = Listbox(root, width=30, font=("Helvetica", font_size))
    rooms_listbox.bind("<Double-Button-1>", go_double_clicked_room)
    rooms_listbox.pack()

    reload_rooms_button = Button(root, text="Reload rooms", font=("Helvetica", font_size),
                                 command=lambda: send_queue.put("/rooms"), state=DISABLED)

    reload_rooms_button.pack()

    root.mainloop()
