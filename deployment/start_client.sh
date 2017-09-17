cd bbg_client
yarn \
  && npm run exec_only \
  && pip install -r requirements.txt \
  && python manage.py migrate \
  && python manage.py collectstatic --noinput
nohup python manage.py pusher > /tmp/pusher.log 2>&1 &
python manage.py runserver "$DJANGO_HOST:$DJANGO_PORT"