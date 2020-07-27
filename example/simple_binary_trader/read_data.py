import websockets
import json
import logging

logger = logging.getLogger('websockets')
logger.setLevel(logging.INFO)
logger.addHandler(logging.StreamHandler())


class BinaryDataReadingAgent:
    uri = "wss://ws.binaryws.com/websockets/v3?app_id=1089"

    last_ohlc_time = None

    def __init__(self, *args, **kwargs):
        pass

    async def run_agent(self, request):
        async with websockets.connect(self.uri) as ws:
            json_data = json.dumps({
                "ticks_history": "R_50",
                "adjust_start_time": 1,
                "count": 10,
                "end": "latest",
                "start": 1,
                "style": "candles",
                "granularity": 60,
                "subscribe": 1
            })
            await ws.send(json_data)

            async for message in ws:
                data = json.loads(message)
                if data["msg_type"] == "ohlc":
                    if not self.last_ohlc_time:
                        self.last_ohlc_time = data["ohlc"]["open_time"]

                    if self.last_ohlc_time != data["ohlc"]["open_time"]:
                        candle = data["ohlc"]
                        print(candle)
                        await self.send_message(data=candle)
                        self.last_ohlc_time = None

    async def process_message(self, message):
        pass
