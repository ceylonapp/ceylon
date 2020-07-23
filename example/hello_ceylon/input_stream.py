import asyncio
import datetime


class HelloCeylonInputSourceAgent:

    def __init__(self, config=None):
        print("input_stream_initiated", config)

    async def run_agent(self, request):
        print("Init params ", request)
        name = request["name"]
        while True:
            await self.send_message(data={
                "name": name,
                "time": datetime.datetime.now().timestamp(),
                "amount": 100
            })
            await asyncio.sleep(1)
