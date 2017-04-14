from django.db import models
from django.contrib.auth.models import AbstractUser


class BBGUser(AbstractUser):

    pass


class Tank(models.Model):

    player = models.ForeignKey('BBGUser', related_name='tanks')
    _rid = models.IntegerField()

    kill_count = models.BigIntegerField(default=0)
    total_steps = models.BigIntegerField(default=0)

    created_at = models.DateTimeField(auto_now_add=True)

    @property
    def rid(self):
        return str(self._rid)

    @rid.setter
    def rid(self, value):
        self._rid = int(value)
