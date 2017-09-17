cd bbg_client
pip install -r requirements.txt \
  && yarn && npm run exec_only \
  && python manage.py migrate \
  && python manage.py collectstatic --noinput
nohup python manage.py pusher > /tmp/pusher.log 2>&1 &
python manage.py runserver "$DJANGO_HOST:$DJANGO_PORT"