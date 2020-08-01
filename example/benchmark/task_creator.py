import sys
import time

payload = ''.join([f"01010101010" for r in range(int(1e5))])
print(sys.getsizeof(payload)/1e3)


class TaskCreatorAgent:
    def __init__(self, config=None):
        print("Task Creator", config)

    async def accept_message(self, agent, message):
        pass

    async def run_agent(self, request):
        while True:
            await self.send_message(data={
                "payload": payload,
                "gen_time": time.time_ns()
            })
