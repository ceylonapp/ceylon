import datetime
import time


class HelloCeylonInputSourceAgent:

    def __init__(self, config=None):
        print("input_stream_initiated", config)

    def run_agent(self, request, response):
        name = request["name"]
        while True:
            response(data={
                "name": name,
                "time": datetime.datetime.now()
            })
            time.sleep(1)
