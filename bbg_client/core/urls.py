from django.conf.urls import url, include

from rest_framework.routers import DefaultRouter

from .views import IndexView
from .api import TankViewSet, UserTankViewSet

router = DefaultRouter()
router.register(r'tanks', TankViewSet, base_name='tanks')
router.register(r'user/tanks', UserTankViewSet, base_name='user-tanks')


urlpatterns = [
    url(r'^api/v1/', include(router.urls)),
    url(r'^$', IndexView.as_view(), name='index'),
]
