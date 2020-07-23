import asyncio

from input_stream import HelloCeylonInputSourceAgent


class HelloProcessAgent:
    __dependents__ = [HelloCeylonInputSourceAgent]
    last_amount = 0

    def __init__(self, config=None):
        print("process_agent_initiated", config)

    async def run_agent(self, request):
        while True:
            await self.send_message(data={
                "name": "second input",
                "process_amount": self.last_amount + 50
            })
            await asyncio.sleep(2)

    async def accept_message(self, agent, message):
        print(agent, type(message["data"]["time"]))
        print(agent, type(message["data"]["amount"]))

        self.last_amount = message["data"]["amount"]
        # hello_input_message = message
        # print("processing := ", hello_input_message)
        # await self.send_message(data={
        #     f"response from process agent ${hello_input_message}"
        # })
