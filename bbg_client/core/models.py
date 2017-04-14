from django.db import models
from django.contrib.auth.models import AbstractUser
from django.db.models.signals import post_save
from django.dispatch import receiver

from django_redis import get_redis_connection
from rest_framework.authtoken.models import Token


class BBGUser(AbstractUser):

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
    def tank_key(self):
        return "bbg:users:%s:tanks" % self.player.pk

    def game_connect(self):
        pass

    def game_disconnect(self):
        pass

    @property
    def hot_data(self):
        return self.redis.hget(self.tank_key, self.pk)


@receiver(post_save, sender=BBGUser)
def create_auth_token(sender, instance=None, created=False, **kwargs):
    if created:
        Token.objects.create(user=instance)
