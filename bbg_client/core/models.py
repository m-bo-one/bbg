from django.db import models
from django.contrib.auth.models import AbstractUser
from django.db.models.signals import post_save
from django.dispatch import receiver
from django.conf import settings

from django_redis import get_redis_connection
from rest_framework.authtoken.models import Token

from protobufs import bbg1_pb2


class BBGUser(AbstractUser):

    tanks_limit = models.SmallIntegerField(default=3)

    @property
    def has_available_tank_slot(self):
        return self.tanks.count() < self.tanks_limit

    class JSONAPIMeta:
        resource_name = "users"


class Tank(models.Model):

    player = models.ForeignKey('BBGUser', related_name='tanks')

    name = models.CharField(max_length=16)
    lvl = models.BigIntegerField(default=1)
    kill_count = models.BigIntegerField(default=0)
    total_steps = models.BigIntegerField(default=0)

    created_at = models.DateTimeField(auto_now_add=True)

    class JSONAPIMeta:
        resource_name = "tanks"

    @property
    def redis(self):
        if not hasattr(self, '_redis'):
            self._redis = get_redis_connection("default")
        return self._redis

    @property
    def thash(self):
        return "bbg:tanks"

    @property
    def tkey(self):
        return "uid:%s:tank:%s" % (self.player.pk, self.pk)

    def game_connect(self):
        pass

    def game_disconnect(self):
        pass

    def save(self):
        super(Tank, self).save()
        if not self.rget():
            self.rcreate()

    def rcreate(self):
        tank = bbg1_pb2.Tank(
            id=self.pk,
            x=int(settings.GAME_CONFIG['MAP']['width'] / 2),
            y=int(settings.GAME_CONFIG['MAP']['height'] / 2),
            health=settings.GAME_CONFIG['TANK_DEFAULT']['health'],
            speed=settings.GAME_CONFIG['TANK_DEFAULT']['speed'],
            fireRate=settings.GAME_CONFIG['TANK_DEFAULT']['fire_rate'],
            width=settings.GAME_CONFIG['TANK_DEFAULT']['width'],
            height=settings.GAME_CONFIG['TANK_DEFAULT']['height'],
            gun=bbg1_pb2.TankGun(
                damage=settings.GAME_CONFIG['TANK_DEFAULT']['gun']['damage'],
                bullets=settings.GAME_CONFIG['TANK_DEFAULT']['gun']['bullets'],
            ),
            direction=bbg1_pb2.Direction.Value(
                settings.GAME_CONFIG['TANK_DEFAULT']['direction'])
        )
        buffer = tank.SerializeToString()
        self.redis.hset(self.thash, self.tkey, buffer)
        return tank

    def rget(self):
        buffer = self.redis.hget(self.thash, self.tkey)
        if buffer:
            tank = bbg1_pb2.Tank()
            tank.ParseFromString(buffer)
            return tank

    @property
    def x(self):
        return self.rget().x

    @property
    def y(self):
        return self.rget().y


@receiver(post_save, sender=BBGUser)
def create_auth_token(sender, instance=None, created=False, **kwargs):
    if created:
        Token.objects.create(user=instance)
