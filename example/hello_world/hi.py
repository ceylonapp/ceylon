import asyncio
from hello_in import HelloInput
class HiProcess:

    __dependent__=[HelloInput]

    def __init__(self,config):
        pass

    async def run_agent(self,request,response):
        print("HI ",request)