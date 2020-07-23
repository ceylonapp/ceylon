class UserInputAgent:

    def __init__(self, config=None):
        print("2nd process_agent_initiated", config)

    async def run_agent(self, request, response):
        print(request)
        await response(data={
            "system_result": request["test"]
        })
        return {
            "ark": "thanks"
        }
