# !/usr/bin/python
import asyncio
import json
import logging
import os
import sys
from importlib.machinery import SourceFileLoader

import aredis
import click

redis_host = os.environ.get('REDIS_HOST', '127.0.0.1')
redis_port = os.environ.get('REDIS_PORT', '6379')
redis_db = os.environ.get('REDIS_DB', '0')

client = aredis.StrictRedis(host=redis_host, port=int(redis_port), db=int(redis_db))


async def process_message(source_instance, read_params):
    # await client.flushdb()
    pub_sub = client.pubsub()
    subscribes = []
    if hasattr(source_instance, "__dependents__"):
        for r in source_instance.__dependents__:
            subscribes.append(r.__name__)

        await pub_sub.subscribe(*subscribes)
        while True:
            message = await pub_sub.get_message()

            if message:
                if type(message["data"]) != int:
                    channel_name = message['channel'].decode("UTF-8")
                    message_body = message["data"].decode("UTF-8")
                    message_body = message_body.replace("'", '"')
                    await source_instance.accept_message(channel_name, json.loads(message_body))


async def run_agent(source, agent, read_params, init_params, path=None, expose=None, type="http"):
    sys.path.append(os.path.abspath(os.getcwd()))  # append source path to system path
    logging.basicConfig(level=logging.DEBUG)

    async def response_stream(*args, **kwargs):
        await client.publish(agent, kwargs)

    def ws_response_stream(ws):
        async def _response_stream(*args, **kwargs):
            print(ws.status)
            await ws.send_json(kwargs)
            await client.publish(agent, kwargs)

        return _response_stream

    foo = SourceFileLoader("", f"{os.getcwd()}/{source}").load_module()

    source_class = getattr(foo, agent)
    source_instance = source_class(config=init_params)
    source_class.send_message = response_stream

    asyncio.create_task(process_message(source_instance, read_params))

    if not path and not expose:
        await source_instance.run_agent(request=read_params)
    elif expose and path:
        from aiohttp import web, WSMsgType
        from aiohttp.web_request import Request

        async def handle(req: Request):
            json_body = await req.json()

            response_body = await source_instance.run_agent(request=json_body)
            response_body = {} if not response_body else response_body
            return web.json_response({
                "status": "success",
                **response_body
            })

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
                        source_instance.send_message = ws_response_stream(ws=ws)
                        if task:
                            task.cancel()
                        task = asyncio.create_task(source_instance.run_agent(request=json_body))

                        # response_body = await
                        # await ws.send_json(response_body)
                elif msg.type == WSMsgType.ERROR:
                    print('ws connection closed with exception %s' %
                          ws.exception())

            print('websocket connection closed')
            return ws

        app = web.Application()

        if type == "http":
            app.add_routes([web.post(f'{path}', handle)])
        elif type == "ws":
            app.add_routes([web.get(f'{path}', websocket_handler)])
        # web.run_app(app, host="0.0.0.0", port=int(expose))

        runner = web.AppRunner(app)
        await runner.setup()
        site = web.TCPSite(runner, host="0.0.0.0", port=int(expose))
        print(f"Agent listing into 0.0.0.0:{expose}{path}")
        await site.start()
        # wait forever
        await asyncio.Event().wait()


@click.command()
@click.option("--source", default=None, help="Please define input streams")
@click.option("--agent", default=None, help="Please define agent class")
@click.option("--path", default=None, help="Please define agent class")
@click.option("--expose", default=None, help="Please define agent class")
@click.option("--type", default="http", help="Please define agent class")
@click.option("--override/--no-override", default=False, help="Please define agent class")
@click.option("--read-params", default={"name": "Agent Framework Reading"})
@click.option("--init-params", default={"name": "Agent Framework Init"})
def run(source, agent, path, expose, type, override, read_params, init_params):
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
    print("source ", source, "agent", agent, "read_params", read_params, "init_params", init_params, "path", path,
          "expose", expose, "type", type)

    asyncio.run(run_agent(source, agent, read_params, init_params, path, expose, type))


if __name__ == '__main__':
    run()
