import sys
import time

from task_creator import TaskCreatorAgent


def get_size(obj, seen=None):
    """Recursively finds size of objects"""
    size = sys.getsizeof(obj)
    if seen is None:
        seen = set()
    obj_id = id(obj)
    if obj_id in seen:
        return 0
    # Important mark as seen *before* entering recursion to gracefully handle
    # self-referential objects
    seen.add(obj_id)
    if isinstance(obj, dict):
        size += sum([get_size(v, seen) for v in obj.values()])
        size += sum([get_size(k, seen) for k in obj.keys()])
    elif hasattr(obj, '__dict__'):
        size += get_size(obj.__dict__, seen)
    elif hasattr(obj, '__iter__') and not isinstance(obj, (str, bytes, bytearray)):
        size += sum([get_size(i, seen) for i in obj])
    return size


class AnalyserAgent:
    __dependents__ = [TaskCreatorAgent]

    def __init__(self, config=None):
        print("Task Creator", config)
        self.data_count = 0
        self.avg_speed = -1
        self.avg_data_size = -1

    async def accept_message(self, agent, message):
        self.data_count += 1
        sender_time = int(message["data"]["gen_time"])
        current_time = time.time_ns()

        if sender_time != current_time:
            speed = (1 / ((current_time - sender_time) / 1e9))
            self.avg_speed = (self.avg_speed + speed) / (2 if self.avg_speed > 0 else 1)

            msg_size = get_size(message["data"]) / 1024
            self.avg_data_size = (self.avg_data_size + msg_size) / (2 if self.avg_data_size > 0 else 1)

        if self.data_count >= 10000:
            print(f"{self.data_count} {self.avg_data_size:0.2f} kb ", f"{self.avg_speed:0.2f} Hz")
            self.avg_speed = -1
            self.avg_data_size = -1
            self.data_count = 0
        # else:
        #     print("Instant")

    async def run_agent(self, request):
        print("started")
