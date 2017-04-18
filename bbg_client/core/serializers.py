from django.utils.translation import ugettext as _

from rest_framework_json_api import serializers

from .models import BBGUser, Tank


class BBGUserSerializer(serializers.ModelSerializer):

    tanks_limit = serializers.IntegerField(required=False)

    class Meta:
        model = BBGUser
        fields = ('username', 'tanks_limit')
        read_only_fields = ('username',)


class TankSerializer(serializers.ModelSerializer):

    player = serializers.HiddenField(default=serializers.CurrentUserDefault())

    class Meta:
        model = Tank
        fields = ('player', 'name', 'lvl', 'tkey', 'kda',
                  'kill_count', 'death_count', 'scores_count')
        read_only_fields = ('lvl', 'tkey', 'kill_count', 'death_count', 'kda',
                            'scores_count')

    def validate_player(self, player):
        if not player.has_available_tank_slot:
            raise serializers.ValidationError(
                _('No more available tanks for this user.'))
        return player
