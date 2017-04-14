from django.conf.urls import url, include

from rest_framework.routers import DefaultRouter

from .views import IndexView
from .api import TankViewSet

router = DefaultRouter()
router.register(r'tanks', TankViewSet)


urlpatterns = [
    url(r'^api/v1/', include(router.urls)),
    url(r'^$', IndexView.as_view(), name='index'),
]
