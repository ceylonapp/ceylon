import asyncio

from hello_process_agent import HelloProcessAgent


class UserInputAgent:
    __dependents__ = [HelloProcessAgent]

    processed_amount = 0

    def __init__(self, config=None):
        print("User input agent", config)

    async def accept_message(self, agent, message):
        self.processed_amount = message["data"]["process_amount"]

    async def run_agent(self, request):
        # print()
        while True:
            await self.send_message(data={
                "ab": "RUN AGENT",
                "amount": self.processed_amount + request["ad"]+50
            })
            await asyncio.sleep(2)
