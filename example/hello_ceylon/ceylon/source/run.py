# !/usr/bin/python
import asyncio
import os

import click


async def publish_input_stream(source):
    from importlib.machinery import SourceFileLoader
    foo = SourceFileLoader("", f"{os.getcwd()}\\{source}").load_module()
    source_class = getattr(foo, "HelloCeylonInputSource")
    source_instance = source_class(config={
        "ab": "TEST"
    })

    def input_stream(*args, **kwargs):
        print(kwargs)

    source_instance.run_agent(request={
        "name": "From Ceylon"
    }, response=input_stream)


@click.command()
@click.option("--source", prompt="Input Stream source", help="Please define input streams")
def run(source):
    asyncio.run(publish_input_stream(source))


if __name__ == '__main__':
    run()
