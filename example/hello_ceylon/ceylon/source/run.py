# !/usr/bin/python
import asyncio
import logging
import os

import aredis
import click

from importlib.machinery import SourceFileLoader


async def run_agent(source, agent, independent):
    logging.basicConfig(level=logging.DEBUG)
    client = aredis.StrictRedis()

    foo = SourceFileLoader("", f"{os.getcwd()}\\{source}").load_module()
    source_class = getattr(foo, agent)
    source_instance = source_class(config={
        "ab": "TEST"
    })

    async def response_stream(*args, **kwargs):
        print(kwargs)
        await client.publish(agent, kwargs)

    if independent:
        await source_instance.run_agent(request={
            "name": "From Ceylon"
        }, response=response_stream)

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
                await source_instance.run_agent(request={
                    "name": "From Ceylon Depend Agent",
                    "message": message
                }, response=response_stream)


@click.command()
@click.option("--source", prompt="Input Stream source", help="Please define input streams")
@click.option("--agent", prompt="Agent class name", help="Please define agent class")
@click.option("--independent/--dependent", default=False)
def run(source, agent, independent):
    asyncio.run(run_agent(source, agent, independent))


if __name__ == '__main__':
    run()
