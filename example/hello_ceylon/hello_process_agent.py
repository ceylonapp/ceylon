from input_stream import HelloCeylonInputSourceAgent


class HelloProcessAgent:
    __dependents__ = [HelloCeylonInputSourceAgent]

    def __init__(self, config=None):
        print("process_agent_initiated", config)

    async def run_agent(self, request, response):
        print("request message ", request)
        await response(data={
            "response from process"
        })
