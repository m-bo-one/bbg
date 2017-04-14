from rest_framework import mixins, authentication, permissions
from rest_framework.viewsets import GenericViewSet

from .models import Tank
from .serializers import TankSerializer


class TankViewSet(mixins.CreateModelMixin,
                  GenericViewSet):

    queryset = Tank.objects.all()
    serializer_class = TankSerializer
    authentication_classes = (authentication.TokenAuthentication,)
    permission_classes = (permissions.IsAuthenticated,)
