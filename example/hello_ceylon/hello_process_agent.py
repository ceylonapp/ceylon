import sys
print(sys.path)
from input_stream import HelloCeylonInputSourceAgent


class HelloProcessAgent:
    __dependents__ = [HelloCeylonInputSourceAgent]

    def __init__(self, config=None):
        print("process_agent_initiated", config)

    async def run_agent(self, request, response):
        hello_input_message = request["messages"][HelloCeylonInputSourceAgent.__name__]
        print("processing := ",hello_input_message)
        await response(data={
            "response from process"
        })
