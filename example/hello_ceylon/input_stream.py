import asyncio
import datetime


class HelloCeylonInputSourceAgent:

    def __init__(self, config=None):
        print("input_stream_initiated", config)

    async def run_agent(self, request, response):
        print("Init params ", request)
        name = request["name"]
        while True:
            await response(data={
                "name": name,
                "time": datetime.datetime.now()
            })
            await asyncio.sleep(1)
