cd bbg_client \
  && yarn \
  && npm run exec_only \
  && pip install -r requirements.txt \
  && python manage.py migrate \
  && python manage.py collectstatic --noinput \
  && python manage.py runserver "$DJANGO_HOST:$DJANGO_PORT"