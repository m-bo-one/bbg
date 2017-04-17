from rest_framework import mixins, authentication, permissions
from rest_framework.viewsets import GenericViewSet

from .models import Tank
from .serializers import TankSerializer


class TankViewSet(mixins.CreateModelMixin,
                  GenericViewSet):

    serializer_class = TankSerializer
    authentication_classes = (authentication.TokenAuthentication,)
    permission_classes = (permissions.IsAuthenticated,)

    def get_queryset(self):
        return (Tank.objects
                .select_related('player')
                .prefetch_related('stats')
                .order_by('created_at'))


class UserTankViewSet(mixins.ListModelMixin,
                      GenericViewSet):

    serializer_class = TankSerializer
    authentication_classes = (authentication.TokenAuthentication,)
    permission_classes = (permissions.IsAuthenticated,)

    def get_queryset(self):
        return (Tank.objects
                .filter(player=self.request.user)
                .select_related('player')
                .prefetch_related('stats')
                .order_by('created_at'))
