if [[ -z "$VIRTUAL_ENV" ]]; then
    echo "No VIRTUAL_ENV set"
    exit 1
else
    echo "VIRTUAL_ENV is set"
fi

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source $DIR/vars.sh

sudo apt-get install libmysqlclient-dev python-dev mysql-server-5.6 default-jre zookeeperd

mkdir -p ~/Downloads
wget "http://mirror.cc.columbia.edu/pub/software/apache/kafka/0.10.1.0/kafka_2.10-0.10.1.0.tgz" -O ~/Downloads/kafka.tgz
mkdir -p ~/kafka && cd ~/kafka
tar -xvzf ~/Downloads/kafka.tgz --strip 1

# GGOOOO

# get godep lib
go get -u github.com/tools/godep
# install plugins from godep dir
cd $GAME_APP_DIR
godep restore


# db create
echo "CREATE DATABASE $PROJECT" | mysql -u root -p


# Python3

pip install -r $DJANGO_APP_DIR/requirements.txt
python $DJANGO_APP_DIR/manage.py migrate


# some noda
cd $DJANGO_APP_DIR
npm install
npm run exec_only

./manage.py collectstatic --noinput
