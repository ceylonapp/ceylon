# !/usr/bin/python
import asyncio
import logging
import os

import aredis
import click

from importlib.machinery import SourceFileLoader


async def publish_input_stream(source, agent):
    logging.basicConfig(level=logging.DEBUG)
    client = aredis.StrictRedis()

    foo = SourceFileLoader("", f"{os.getcwd()}\\{source}").load_module()
    source_class = getattr(foo, agent)
    source_instance = source_class(config={
        "ab": "TEST"
    })

    async def input_stream(*args, **kwargs):
        print(kwargs)
        await client.publish(agent, kwargs)

    await source_instance.run_agent(request={
        "name": "From Ceylon"
    }, response=input_stream)


@click.command()
@click.option("--source", prompt="Input Stream source", help="Please define input streams")
@click.option("--agent", prompt="Agent class name", help="Please define agent class")
def run(source, agent):
    asyncio.run(publish_input_stream(source, agent))


if __name__ == '__main__':
    run()
