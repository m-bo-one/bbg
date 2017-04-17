import os
import asyncio
import logging

import django
from kafka.common import KafkaError
from aiokafka import AIOKafkaConsumer

from django.conf import settings

os.environ.setdefault("DJANGO_SETTINGS_MODULE", "bbg_client.settings")
django.setup()

logging.basicConfig(level=logging.INFO)


class BBGPusher(object):

    def __init__(self, topics=None):
        self._tasks = []
        self._loop = asyncio.get_event_loop()
        self._bootstrap_servers = "{host}:{port}".format(
            host=settings.KAFKA_SETTINGS['server1']['HOST'],
            port=settings.KAFKA_SETTINGS['server1']['PORT']
        )
        self._topics = topics
        self.consumer = AIOKafkaConsumer(
            *self._topics,
            loop=self._loop,
            bootstrap_servers=self._bootstrap_servers)
        self._loop.run_until_complete(self.consumer.start())

        self._tasks.append(self._loop.create_task(self.listener()))

    def run(self):
        try:
            self._loop.run_forever()
        finally:
            self._loop.run_until_complete(self.consumer.stop())
            for task in self._tasks:
                task.cancel()
            self._loop.close()

    async def listener(self):
        while True:
            try:
                msg = await self.consumer.getone()
                if msg.topic in self._topics:
                    f_key = 'handle_topic_%s' % msg.topic
                    logging.info("BBGPusher: start process handler %s...: ",
                                 f_key)
                    await getattr(self, f_key)(msg)
            except KafkaError as err:
                logging.error("BBGPusher: err while consuming message: ", err)
            except AttributeError as err:
                logging.error("BBGPusher: topic %s err: %s", msg.topic, err)
            except Exception as err:
                logging.error("BBGPusher: unexpected exception %s", err)

    async def handle_topic_tank_stat(self, msg):
        from core.models import Tank, Stat
        logging.info('BBGPusher: receive msg key: %s', msg.key)

        key = msg.key
        if key:
            key = key.decode('utf-8')
        if key.isdigit():
            key = int(key)

        if key and key in Stat.EVENT_CHOICES_LIST:
            try:
                value = msg.value.decode('utf-8')
                tank = Tank.objects.get(tkey=value)
            except Tank.DoesNotExist as err:
                logging.error("BBGPusher: tank model error %s.", err)
            else:
                tank.stats.create(event=key)
                logging.info("BBGPusher: stat %s added.",
                             dict(Stat.EVENT_CHOICES)[key])

        logging.info("BBGPusher: handler processing done.")


if __name__ == '__main__':
    pusher = BBGPusher(['tank_stat'])
    pusher.run()
