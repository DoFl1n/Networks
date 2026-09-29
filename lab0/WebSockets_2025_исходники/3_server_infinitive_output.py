#!/usr/bin/python3
import datetime
import time
import asyncio
import websockets

async def hello(websocket, path):
    name = await websocket.recv()
    print("client say: "+str(name))
    counter=0
    while(True):
        #greeting = input("enter answer to client: ")
        counter = counter+1
        greeting = str(counter)+"->"+str(name) + " " + str(datetime.datetime.now())
        print("sleeping...")
        time.sleep(10) # 10 sec sleep
        print("... awaked and sending to websocket")
        await websocket.send(greeting)
        print("... just sent this:")
        print(str(greeting))

start_server = websockets.serve(hello, "151.248.113.144", 8000)
asyncio.get_event_loop().run_until_complete(start_server)
asyncio.get_event_loop().run_forever()
