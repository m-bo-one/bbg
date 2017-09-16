from django.views.generic import TemplateView
from django.conf import settings


class IndexView(TemplateView):

    template_name = "core/index.html"

    def get_context_data(self, **kwargs):
        ctx_data = super(IndexView, self).get_context_data(**kwargs)
        ctx_data['BBG_WS_URL'] = settings.BBG_WS_URL
        ctx_data['CDN_URL'] = settings.CDN_URL
        return ctx_data
