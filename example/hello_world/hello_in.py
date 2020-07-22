import asyncio
class HelloInput:

    def __init__(self,config):
        pass

    async def run_agent(self,request,response):
        while True:
            await response(data={"Hello"})
            await asyncio.sleep(1)