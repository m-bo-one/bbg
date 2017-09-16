if [[ -z "$VIRTUAL_ENV" ]]; then
    echo "No VIRTUAL_ENV set"
    exit 1
else
    echo "VIRTUAL_ENV is set"
fi

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source $DIR/vars.sh

sudo useradd kafka -m

curl -sL https://deb.nodesource.com/setup_6.x | sudo -E bash -
sudo apt-get install -y build-essential \
    mysql-server \
    libmysqlclient-dev \
    python-dev \
    default-jre \
    zookeeperd \
    npm \
    default-jre \
    default-jdk \
    golang-goprotobuf-dev
sudo npm install -g yarn gulp

wget "http://mirror.cc.columbia.edu/pub/software/apache/kafka/0.10.1.0/kafka_2.10-0.10.1.0.tgz" -O /tmp/kafka.tgz
tar -xvzf /tmp/kafka.tgz -C /tmp
sudo mv /tmp/kafka_2.10-0.10.1.0 /usr/local/kafka
echo "delete.topic.enable = true" >> /usr/local/kafka/config/server.properites

source ./deployment/start_kafka.sh

# # GGOOOO

sudo add-apt-repository -y ppa:longsleep/golang-backports
sudo apt-get update
sudo apt-get install golang-go

# get godep lib
go get -u github.com/tools/godep
go get github.com/pilu/fresh
# install plugins from godep dir
cd $GAME_APP_DIR
$GOPATH/bin/godep restore


# db create
echo "CREATE DATABASE $PROJECT" | mysql -u root -p

# some noda
cd $DJANGO_APP_DIR
yarn install

# PROTOC 3

# Make sure you grab the latest version
curl -OL https://github.com/google/protobuf/releases/download/v3.2.0/protoc-3.2.0-linux-x86_64.zip

# Unzip
unzip protoc-3.2.0-linux-x86_64.zip -d protoc3

# Move protoc to /usr/local/bin/
sudo mv protoc3/bin/* /usr/local/bin/

# Move protoc3/include to /usr/local/include/
sudo mv protoc3/include/* /usr/local/include/

npm run exec_only

# Python3
mkvirtualenv -p python3 $PROJECT
pip install -r $DJANGO_APP_DIR/requirements.txt
python $DJANGO_APP_DIR/manage.py migrate

python $DJANGO_APP_DIR/manage.py collectstatic --noinput
