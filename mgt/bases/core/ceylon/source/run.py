# !/usr/bin/python
import asyncio
import logging
import os

import aredis
import click

from importlib.machinery import SourceFileLoader


async def run_agent(source, agent, independent, read_params, init_params):
    logging.basicConfig(level=logging.DEBUG)
    redis_host = os.environ.get('REDIS_HOST')
    redis_port = os.environ.get('REDIS_PORT')
    redis_db = os.environ.get('REDIS_DB')
    client = aredis.StrictRedis(host=redis_host, port=int(redis_port),db=int(redis_db))

    foo = SourceFileLoader("", f"{os.getcwd()}/{source}").load_module()
    source_class = getattr(foo, agent)
    source_instance = source_class(config=init_params)

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
@click.option("--source", prompt="Input Stream source", help="Please define input streams")
@click.option("--agent", prompt="Agent class name", help="Please define agent class")
@click.option("--independent/--dependent", default=False)
@click.option("--read-params", default={"name": "Agent Framework Reading"})
@click.option("--init-params", default={"name": "Agent Framework Init"})
def run(source, agent, independent, read_params, init_params):
    asyncio.run(run_agent(source, agent, independent, read_params, init_params))


if __name__ == '__main__':
    run()
