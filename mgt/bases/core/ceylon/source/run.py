# !/usr/bin/python
import asyncio
import json
import logging
import os
import pickle
import sys
from datetime import datetime
from importlib.machinery import SourceFileLoader

import aredis
import click
import redis
from environs import Env

agent_name, source_name, stack_name = "", "", ""
for idx, arg in enumerate(sys.argv):
    if arg == "--source":
        source_name = sys.argv[idx + 1]
    elif arg == "--agent":
        agent_name = sys.argv[idx + 1]
    elif arg == "--stack":
        stack_name = sys.argv[idx + 1]

env = Env()
# Read .env into os.environ
if os.path.exists(".env"):
    env.read_env()
if os.path.exists(".env.ceylon"):
    env.read_env(".env.ceylon", recurse=False)
else:
    print("path not exits", os.path.abspath(".env.ceylon"))
# Initialize Redis ENV variables
stack_name_prefix = f"{stack_name}_" if stack_name != "" else ""
redis_host = os.environ.get(f'{stack_name_prefix}REDIS_HOST', '127.0.0.1')
redis_port = os.environ.get(f'{stack_name_prefix}REDIS_PORT', '6379')
redis_db = os.environ.get(f'{stack_name_prefix}REDIS_DB', '0')

client = aredis.StrictRedis(host=redis_host,
                            port=int(redis_port),
                            db=int(redis_db))


class SysLogger(object):
    def __init__(self, name="Agent"):
        self.name = name
        self.terminal = sys.stdout
        self.r = redis.Redis(host=redis_host, port=redis_port, db=redis_db)
        self.channel_name = f"{self.name}sys_log"
        self.terminal.write(f"Channel Name ::: {self.channel_name}")

    def write(self, message):
        self.terminal.write(message)
        self.r.publish(
            self.channel_name,
            json.dumps({
                "time": f"{datetime.now()}",
                "agent": self.name,
                "message": message
            }))

    def flush(self):
        pass


sys.stdout = SysLogger(name=stack_name_prefix)
print("init agent logger ", stack_name, source_name, agent_name)


async def process_message(source_instance, read_params):
    pub_sub = client.pubsub()
    subscribes = []
    if hasattr(source_instance, "__dependents__"):
        # print(f"All __dependents__ {source_instance.__dependents__}")
        for r in source_instance.__dependents__:
            subscribes.append(r if type(r) == str else r.__name__)

        await pub_sub.subscribe(*subscribes)
        while True:
            message = await pub_sub.get_message()
            if message:
                if type(message["data"]) != int:
                    channel_name = message['channel'].decode("UTF-8")
                    message_body = message["data"]
                    await source_instance.accept_message(
                        channel_name, pickle.loads(message_body))
    else:
        print("No __dependents__ found please check your agent again")


async def run_agent(source,
                    agent,
                    read_params,
                    init_params,
                    path=None,
                    expose=None,
                    type="http"):
    sys.path.append(os.path.abspath(
        os.getcwd()))  # append source path to system path
    logging.basicConfig(level=logging.DEBUG)

    async def response_stream(*args, **kwargs):
        await client.publish(agent, pickle.dumps(kwargs))

    def ws_response_stream(ws):
        async def _response_stream(*args, **kwargs):
            await client.publish(agent, pickle.dumps(kwargs))
            await ws.send_json(kwargs)

        return _response_stream

    foo = SourceFileLoader("", f"{os.getcwd()}/{source}").load_module()
    print(init_params)
    source_class = getattr(foo, agent)
    source_instance = source_class(config=init_params)
    source_class.send_message = response_stream

    task = asyncio.create_task(process_message(source_instance, read_params))

    if not path and not expose:
        async def run_agent_process():
            print("run_agent_process")
            await source_instance.run_agent(request=read_params)

        task_run_agent = asyncio.create_task(run_agent_process())
        result = await asyncio.gather(task_run_agent, task)
        print(result)

    elif expose and path:
        from aiohttp import web, WSMsgType
        from aiohttp.web_request import Request
        import aiohttp_cors

        async def handle(req: Request):
            json_body = await req.json()

            response_body = await source_instance.run_agent(request=json_body)
            response_body = {} if not response_body else response_body
            return web.json_response({"status": "success", **response_body})

        async def websocket_handler(request):

            ws = web.WebSocketResponse()
            await ws.prepare(request)
            task = None
            async for msg in ws:
                if msg.type == WSMsgType.TEXT:
                    if msg.data == 'close':
                        await ws.close()
                    else:
                        json_body = json.loads(msg.data)
                        source_instance.send_message = ws_response_stream(
                            ws=ws)
                        if task:
                            task.cancel()
                        task = asyncio.create_task(
                            source_instance.run_agent(request=json_body))

                        # response_body = await
                        # await ws.send_json(response_body)
                elif msg.type == WSMsgType.ERROR:
                    print('ws connection closed with exception %s' %
                          ws.exception())
            print('websocket connection closed')
            return ws

        app = web.Application()

        if type == "http":
            cors = aiohttp_cors.setup(app,
                                      defaults={
                                          "*":
                                              aiohttp_cors.ResourceOptions(
                                                  allow_credentials=True,
                                                  expose_headers="*",
                                                  allow_headers="*",
                                              )
                                      })
            app.add_routes([web.post(f'{path}', handle)])

            for route in list(app.router.routes()):
                cors.add(route)

        elif type == "ws":
            app.add_routes([web.get(f'{path}', websocket_handler)])

        runner = web.AppRunner(app)
        await runner.setup()
        site = web.TCPSite(runner, host="0.0.0.0", port=int(expose))
        print(f"Agent listing into 0.0.0.0:{expose}{path}")
        await site.start()
        await asyncio.Event().wait()

    await task
    print(f" {agent}  finished {task}")


@click.command()
@click.option("--stack", default=None, help="Please define Agent Stack name")
@click.option("--source", default=None, help="Please define input streams")
@click.option("--agent", default=None, help="Please define agent class")
@click.option("--path", default=None, help="Web expose path")
@click.option("--expose", default=None, help="Web expose port")
@click.option("--type",
              default="http",
              help="Type can be ws(WebSocket) or http")
@click.option("--override/--no-override",
              default=False,
              help="Override params by system variable ")
@click.option("--init-params", multiple=True, default=[("name", "agent_init")], type=click.Tuple([str, str]))
@click.option("--read-params", multiple=True, default=[("name", "agent_reading")], type=click.Tuple([str, str]))
def run(stack, source, agent, path, expose, type, override, read_params,
        init_params):
    if override:
        source = os.environ.get('CEYLON_SOURCE')
        agent = os.environ.get('CEYLON_AGENT')
        if os.environ.get("CEYLON_PATH") != "":
            path = os.environ.get('CEYLON_PATH')
        if os.environ.get("CEYLON_EXPOSE") != "":
            expose = os.environ.get('CEYLON_EXPOSE')
        if os.environ.get("CEYLON_TYPE") != "":
            type = os.environ.get('CEYLON_TYPE')

    source = f"{source}"
    agent = f"{agent}"

    if init_params:
        init_params = dict(init_params)

    if read_params:
        read_params = dict(read_params)

    print("Agent Stack ", stack_name, "source ", source, "agent", agent,
          "read_params", read_params, "init_params", init_params, "path", path,
          "expose", expose, "type", type)

    # loop = asyncio.get_event_loop()
    # loop.run_until_complete(
    # run_agent(source, agent, read_params, init_params, path, expose, type))
    asyncio.run(run_agent(source, agent, read_params, init_params, path, expose, type))


if __name__ == '__main__':
    run()
