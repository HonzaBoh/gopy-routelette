import socket
import threading



sock = socket.socket()
sock.connect(("localhost", 10000))

def recv_str():
    while True:
        try:
            print("Received:", msg := sock.recv(1024).decode().strip())
        except:
            print("Error occured - client")
            sock.close()
            break
         

def send_str():
    while True:
        msg = input('')
        print("Sent out:", msg)
        sock.send(f"{msg}\n".encode())
    
receive_thread = threading.Thread(target=recv_str)
receive_thread.start()

write_thread = threading.Thread(target=send_str)
write_thread.start()