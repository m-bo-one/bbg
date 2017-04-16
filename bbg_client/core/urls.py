from django.conf.urls import url, include

from rest_framework.routers import DefaultRouter

from .views import IndexView
from .api import TankViewSet

router = DefaultRouter()
router.register(r'tanks', TankViewSet, base_name='tanks')


urlpatterns = [
    url(r'^api/v1/', include(router.urls)),
    url(r'^$', IndexView.as_view(), name='index'),
    url(r'^test/', IndexView.as_view(template_name = "core/test.html"), name='tindex'),
]
