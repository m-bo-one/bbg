import asyncio
import logging
import concurrent.futures

from django.core.management.base import BaseCommand, CommandError
from django.db import close_old_connections

from kafka.common import KafkaError
from aiokafka import AIOKafkaConsumer

from core.models import Tank, Stat


logging.basicConfig(level=logging.INFO,
                    format='[%(asctime)s %(levelname)s '
                           '%(name)s:%(filename)s:%(lineno)s] %(message)s')


class Command(BaseCommand):

    help = 'Kafka push service'

    def __init__(self, *args, **kwargs):
        super(Command, self).__init__(*args, **kwargs)
        self._tasks = []
        self._loop = asyncio.get_event_loop()
        self._topics = ['tank_stat']
        self.consumer = AIOKafkaConsumer(
            *self._topics,
            loop=self._loop,
            bootstrap_servers='localhost:9092')

        self._loop.run_until_complete(self.consumer.start())

        self._executor = concurrent.futures.ThreadPoolExecutor(max_workers=10)

        for task in [self.listener, self.cron]:
            self._tasks.append(self._loop.create_task(task()))

    def handle(self, *args, **options):
        try:
            self._loop.run_forever()
        except Exception as err:
            raise CommandError(err)
        finally:
            self._loop.run_until_complete(self.consumer.stop())
            [task.cancel() for task in self._tasks]
            self._loop.close()

    async def cron(self):
        while True:
            await asyncio.sleep(10)

    async def listener(self):
        while True:
            try:
                msg = await self.consumer.getone()
                if msg.topic in self._topics:
                    f_key = 'handle_topic_%s' % msg.topic
                    logging.info("BBGPusher: start process handler %s...: ",
                                 f_key)
                    await self._loop.run_in_executor(self._executor,
                                                     getattr(self, f_key),
                                                     msg)
            except KafkaError as err:
                logging.error("BBGPusher: err while consuming message: ", err)
            except AttributeError as err:
                logging.error("BBGPusher: topic %s err: %s", msg.topic, err)
            except Exception as err:
                logging.error("BBGPusher: unexpected exception %s", err)

    def handle_topic_tank_stat(self, msg):
        logging.info('BBGPusher: receive msg key: %s', msg.key)

        key = msg.key
        if key:
            key = key.decode('utf-8')
        if key.isdigit():
            key = int(key)

        if key and key in Stat.EVENT_CHOICES_LIST:
            close_old_connections()
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
