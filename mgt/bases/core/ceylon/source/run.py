# !/usr/bin/python
import asyncio
import logging
import os
import sys
from importlib.machinery import SourceFileLoader

import aredis
import click

redis_host = os.environ.get('REDIS_HOST')
redis_port = os.environ.get('REDIS_PORT')
redis_db = os.environ.get('REDIS_DB')


async def run_agent(source, agent, read_params, init_params, path=None, expose=None):
    sys.path.append(os.path.abspath(os.getcwd()))  # append source path to system path
    print("system path ", sys.path)
    logging.basicConfig(level=logging.DEBUG)
    client = aredis.StrictRedis(host=redis_host, port=int(redis_port), db=int(redis_db))
    print("Agent Module:= ", f"{os.getcwd()}/{source}")
    print("Agent := ", agent)

    foo = SourceFileLoader("", f"{os.getcwd()}/{source}").load_module()

    source_class = getattr(foo, agent)
    source_instance = source_class(config=init_params)
    independent = True
    if hasattr(source_instance, '__dependents__'):
        independent = False

    async def response_stream(*args, **kwargs):
        await client.publish(agent, kwargs)

    if independent and not path and not expose:
        await source_instance.run_agent(request=read_params, response=response_stream)
    elif expose and path:

        from aiohttp import web
        from aiohttp.web_request import Request
        import asyncio

        async def handle(req: Request):
            json_body = await req.json()

            response_body = await source_instance.run_agent(request=json_body, response=response_stream)
            response_body = {} if not response_body else response_body
            return web.json_response({
                "status": "success",
                **response_body
            })

        app = web.Application()
        app.add_routes([web.post(f'{path}', handle)])
        # web.run_app(app, host="0.0.0.0", port=int(expose))

        runner = web.AppRunner(app)
        await runner.setup()
        site = web.TCPSite(runner, host="0.0.0.0", port=int(expose))
        print(f"Agent listing into 0.0.0.0:{expose}{path}")
        await site.start()
        # wait forever
        await asyncio.Event().wait()

        # uvicorn.run(app, host='0.0.0.0', port=int(expose), loop="auto")
    else:
        await client.flushdb()
        pub_sub = client.pubsub()
        subscribes = []
        for r in source_instance.__dependents__:
            subscribes.append(r.__name__)

        await pub_sub.subscribe(*subscribes)
        while True:
            message = await pub_sub.get_message()

            if message:
                if type(message["data"]) != int:
                    channel_name = message['channel'].decode("UTF-8")
                    message_body = message["data"].decode("UTF-8")

                    await source_instance.run_agent(request={
                        "name": "agent_framework",
                        "messages": {
                            channel_name: message_body
                        },
                        **read_params
                    }, response=response_stream)


@click.command()
@click.option("--source", default=None, help="Please define input streams")
@click.option("--agent", default=None, help="Please define agent class")
@click.option("--path", default=None, help="Please define agent class")
@click.option("--expose", default=None, help="Please define agent class")
@click.option("--override/--no-override", default=False, help="Please define agent class")
@click.option("--read-params", default={"name": "Agent Framework Reading"})
@click.option("--init-params", default={"name": "Agent Framework Init"})
def run(source, agent, path, expose, override, read_params, init_params):
    if override:
        source = os.environ.get('CEYLON_SOURCE')
        agent = os.environ.get('CEYLON_AGENT')
        path = os.environ.get('CEYLON_PATH')
        expose = os.environ.get('CEYLON_EXPOSE')

    source = f"{source}"
    agent = f"{agent}"

    asyncio.run(run_agent(source, agent, read_params, init_params, path, expose))


if __name__ == '__main__':
    run()
