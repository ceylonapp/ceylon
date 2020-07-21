from hello_process_agent import HelloProcessAgent


class HelloSecondProcessAgent:
    __dependents__ = [HelloProcessAgent]

    def __init__(self, config=None):
        print("2nd process_agent_initiated", config)

    async def run_agent(self, request, response):
        hello_input_message = request["messages"][HelloProcessAgent.__name__]
        print("2nd processing := ",hello_input_message)
        await response(data={
            "response from Second process"
        })
