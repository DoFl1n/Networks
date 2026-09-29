import asyncio
import websockets

async def hello(websocket):
    async for message in websocket:
        print(f"client_says: {message}")
        reply = await asyncio.to_thread(input, "enter answer: ")
        await websocket.send(reply)	
        print(str(reply))

async def main():
    async with websockets.serve(hello, "127.0.0.1", 8008):
        await asyncio.Future()

if __name__ == "__main__":
    asyncio.run(main())
