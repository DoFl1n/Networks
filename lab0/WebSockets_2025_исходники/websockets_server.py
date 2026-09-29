#!/usr/bin/python3

import asyncio
import websockets

async def hello(websocket, path):
    recv_string = await websocket.recv()
    print("client say: "+str(recv_string))

    while(True):
        reply_string = input("enter answer to client: ")
        await websocket.send(reply_string)
        print(str(reply_string))

start_server = websockets.serve(hello, "127.0.0.1", 8008)
asyncio.get_event_loop().run_until_complete(start_server)
asyncio.get_event_loop().run_forever()
