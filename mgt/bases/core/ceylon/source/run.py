# !/usr/bin/python
import asyncio
import logging
import os
import sys
from importlib.machinery import SourceFileLoader

import aredis
import click


async def run_agent(source, agent, read_params, init_params):
    sys.path.append(os.path.abspath(os.getcwd()))  # append source path to system path
    print("system path ", sys.path)
    logging.basicConfig(level=logging.DEBUG)
    redis_host = os.environ.get('REDIS_HOST')
    redis_port = os.environ.get('REDIS_PORT')
    redis_db = os.environ.get('REDIS_DB')
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

    if independent:
        await source_instance.run_agent(request=read_params, response=response_stream)
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
@click.option("--override/--no-override", default=False, help="Please define agent class")
@click.option("--read-params", default={"name": "Agent Framework Reading"})
@click.option("--init-params", default={"name": "Agent Framework Init"})
def run(source, agent, override, read_params, init_params):
    if override:
        source = os.environ.get('CEYLON_SOURCE')
    if override:
        agent = os.environ.get('CEYLON_AGENT')

    source = f"{source}"
    agent = f"{agent}"

    asyncio.run(run_agent(source, agent, read_params, init_params))


if __name__ == '__main__':
    run()
