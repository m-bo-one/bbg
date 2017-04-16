from django.db import models
from django.contrib.auth.models import AbstractUser
from django.db.models.signals import post_save
from django.dispatch import receiver
from django.conf import settings
from django.utils.translation import ugettext as _

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


class RTankProxy(object):

    @property
    def x(self):
        return self._rget().x

    @property
    def y(self):
        return self._rget().y

    @property
    def health(self):
        return self._rget().health

    @property
    def speed(self):
        return self._rget().speed

    @property
    def fire_rate(self):
        return self._rget().fireRate

    @property
    def width(self):
        return self._rget().width

    @property
    def height(self):
        return self._rget().height

    @property
    def damage(self):
        return self._rget().gun.damage

    @property
    def bullets(self):
        return self._rget().gun.bullets

    @property
    def direction(self):
        return self._rget().direction

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

    def save(self):
        super(RTankProxy, self).save()

        if not self._rget():
            self._rcreate()

    def _rcreate(self):
        tank = bbg1_pb2.Tank(
            id=self.pk,
            x=int(settings.GAME_CONFIG['MAP']['width'] / 2),
            y=int(settings.GAME_CONFIG['MAP']['height'] / 2)
        )
        buffer = tank.SerializeToString()
        self.redis.hset(self.thash, self.tkey, buffer)
        return tank

    def _rget(self):
        buffer = self.redis.hget(self.thash, self.tkey)
        if buffer:
            tank = bbg1_pb2.Tank()
            tank.ParseFromString(buffer)
            return tank


class Tank(RTankProxy, models.Model):

    player = models.ForeignKey('BBGUser', related_name='tanks')

    name = models.CharField(max_length=16)
    lvl = models.BigIntegerField(default=1)
    kill_count = models.BigIntegerField(default=0)
    total_steps = models.BigIntegerField(default=0)

    created_at = models.DateTimeField(auto_now_add=True)

    class JSONAPIMeta:
        resource_name = "tanks"

    def save(self):
        if self.pk and not self.player.has_available_tank_slot:
            raise Exception(_("No more available tanks for this user."))
        super(Tank, self).save()

    def game_connect(self):
        pass

    def game_disconnect(self):
        pass


@receiver(post_save, sender=BBGUser)
def create_auth_token(sender, instance=None, created=False, **kwargs):
    if created:
        Token.objects.create(user=instance)
