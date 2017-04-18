if [[ -z "$VIRTUAL_ENV" ]]; then
    echo "No VIRTUAL_ENV set"
    exit 1
else
    echo "VIRTUAL_ENV is set"
fi

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source $DIR/vars.sh


# GGOOOO

# get godep lib
go get -u github.com/tools/godep
# install plugins from godep dir
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
