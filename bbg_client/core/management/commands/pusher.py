import os
import asyncio
import logging
import uuid
import concurrent.futures

from django.core.management.base import BaseCommand, CommandError
from django.db import close_old_connections
from django.conf import settings

from kafka.common import KafkaError
import aioredis
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
            bootstrap_servers=settings.KAFKA_BROKER_URL)

        self._loop.run_until_complete(self.consumer.start())

        self._executor = concurrent.futures.ThreadPoolExecutor(max_workers=10)

        self._last_stat_id = None

        # call for init
        future = aioredis.create_redis(
            (os.getenv('REDIS_HOST', '127.0.0.1'), 6379),
            loop=self._loop)
        self.redis = self._loop.run_until_complete(future)

        self._loop.run_until_complete(self.clear_stats())

        for task in [self.listener, self.cron]:
            self._tasks.append(self._loop.create_task(task()))

    def handle(self, *args, **options):
        try:
            self._loop.run_forever()
        except Exception as err:
            raise CommandError(err)
        finally:
            self._loop.run_until_complete(self.consumer.stop())
            self.redis_pool.close()
            self._loop.run_until_complete(self.redis_pool.wait_closed())
            [task.cancel() for task in self._tasks]
            self._loop.close()

    def get_last_stats(self):
        filter_by = {}
        if self._last_stat_id:
            filter_by['id__gt'] = self._last_stat_id

        resp = {}

        close_old_connections()
        try:
            stats = Stat.objects \
                .filter(**filter_by) \
                .select_related('tank__name', 'tank__tkey') \
                .order_by('-created_at')
        except Exception as err:
            logging.error("BBGPusher: stats error: ", err)
            self._last_stat_id = None
        else:
            logging.info("BBGPusher: stats successfully received")
            self._last_stat_id = stats[0].pk
            for stat in stats:
                resp.setdefault(stat.tank.tkey, {
                    'scores': 0,
                    'name': stat.tank.name
                })
                resp[stat.tank.tkey]['scores'] += Stat.SCORE_MAP[stat.event]
        finally:
            return resp

    async def clear_stats(self):
        await self.redis.delete('bbg:scores')

    async def update_stats(self, stats):
        futures = []
        for tkey, data in stats.items():

            async def _callback(tkey, data):
                buff = await self.redis.hget('bbg:scores', tkey)
                stat = Stat.load_proto(buff)
                stat.id = uuid.uuid4().hex
                stat.scores += data['scores']
                stat.tankId = tkey
                stat.name = data['name']
                await self.redis.hset('bbg:scores', tkey,
                                      stat.SerializeToString())

            futures.append(_callback(tkey, data))

        await asyncio.wait(futures)

    async def cron(self):
        while True:
            await asyncio.sleep(0.5)
            stats = await self._loop.run_in_executor(self._executor,
                                                     self.get_last_stats)
            if stats:
                logging.info("BBGPusher: updating stats")
                asyncio.ensure_future(self.update_stats(stats))

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
