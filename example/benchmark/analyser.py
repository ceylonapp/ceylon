import sys
import time

from task_creator import TaskCreatorAgent


class AnalyserAgent:
    __dependents__ = [TaskCreatorAgent]

    def __init__(self, config=None):
        print("Task Creator", config)

    async def accept_message(self, agent, message):
        sender_time = int(message["data"]["gen_time"])
        current_time = time.time_ns()
        msg_size = sys.getsizeof(message["data"])
        if sender_time != current_time:
            print(f"Message {msg_size/1e3}kb Speed ", f"{(1 / ((current_time - sender_time) / 1e9)):2f}Hz")
        else:
            print("Instant")

    async def run_agent(self, request):
        print("started")
