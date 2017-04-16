import hashlib
import random
import uuid

from django.db import models
from django.contrib.auth.models import AbstractUser
from django.db.models.signals import post_save
from django.dispatch import receiver
from django.conf import settings
from django.utils.translation import ugettext as _
from django.utils import timezone

from django_redis import get_redis_connection
from rest_framework.authtoken.models import Token
from google.protobuf.message import Message

from protobufs import bbg1_pb2


class BBGUser(AbstractUser):

    tanks_limit = models.SmallIntegerField(default=3)

    @property
    def has_available_tank_slot(self):
        return self.tanks.count() < self.tanks_limit

    def tkeys(self):
        return [tank.tkey for tank in self.tanks.all()]

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
    def gun_damage(self):
        return self._rget().gun.damage

    @property
    def gun_bullets(self):
        return self._rget().gun.bullets

    @property
    def gun_distance(self):
        return self._rget().gun.distance

    @gun_distance.setter
    def gun_distance(self, value):
        self._rupdate(**{
            'gun': bbg1_pb2.TankGun(
                damage=self.gun_damage,
                bullets=self.gun_bullets,
                distance=float(value),
            )
        })

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

    def save(self):
        super(RTankProxy, self).save()
        if not self._rget():
            self._rcreate()

    def delete(self):
        super(RTankProxy, self).delete()
        self._rdel()

    def recreate(self):
        self._rdel()
        self._rcreate()

    def _rupdate(self, **kw):
        tank = self._rget()
        if tank:
            for k, v in kw.items():
                if isinstance(getattr(tank, k), Message):
                    getattr(tank, k).CopyFrom(v)
                else:
                    setattr(tank, k, v)
            self.redis.hset(self.thash, self.tkey, tank.SerializeToString())
            return tank

    def _rcreate(self):
        tank = bbg1_pb2.Tank(
            id=self.pk,
            x=int(settings.GAME_CONFIG['MAP']['width'] / 2),
            y=int(settings.GAME_CONFIG['MAP']['height'] / 2),
        )
        self.redis.hset(self.thash, self.tkey, tank.SerializeToString())
        return tank

    def _rget(self):
        buffer = self.redis.hget(self.thash, self.tkey)
        if buffer:
            tank = bbg1_pb2.Tank()
            tank.ParseFromString(buffer)
            return tank

    def _rdel(self):
        self.redis.hdel(self.thash, self.tkey)


def generate_rtk():
    return hashlib.md5(
        "{uid}:{salt}:{now}".format(uid=uuid.uuid4().hex,
                                    salt=random.random(),
                                    now=timezone.now().timestamp())
        .encode('utf-8')
    ).hexdigest()


class Tank(RTankProxy, models.Model):

    player = models.ForeignKey('BBGUser', related_name='tanks')
    tkey = models.CharField(max_length=32, unique=True, default=generate_rtk)

    name = models.CharField(max_length=16)
    lvl = models.BigIntegerField(default=1)
    kill_count = models.BigIntegerField(default=0)
    total_steps = models.BigIntegerField(default=0)

    created_at = models.DateTimeField(auto_now_add=True)

    class JSONAPIMeta:
        resource_name = "tanks"

    def save(self):
        is_created = True if self.pk else False
        if is_created and not self.player.has_available_tank_slot:
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
