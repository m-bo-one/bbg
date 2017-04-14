from rest_framework import serializers

from .models import Tank


class TankSerializer(serializers.ModelSerializer):

    kill_count = serializers.IntegerField(required=False)
    total_steps = serializers.IntegerField(required=False)
    lvl = serializers.IntegerField(required=False)

    player = serializers.HiddenField(default=serializers.CurrentUserDefault())

    class Meta:
        model = Tank
        fields = ('player', 'name', 'lvl', 'kill_count', 'total_steps')
        read_only_fields = ('lvl', 'kill_count', 'total_steps')
