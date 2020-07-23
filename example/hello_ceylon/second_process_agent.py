from hello_process_agent import HelloProcessAgent


class HelloSecondProcessAgent:
    __dependents__ = [HelloProcessAgent]

    def __init__(self, config=None):
        print("2nd process_agent_initiated", config)

    async def run_agent(self, request):
        await self.send_message(data={
            "response from Second process"
        })

    async def accept_message(self, agent, message):
        print("2nd processing := ", agent, message)
