#!/bin/bash
WORKSPACE_DIR=$GOPATH/src/github.com/DeV1doR
PROJ_DIR=$WORKSPACE_DIR/bbg
CLIENT_DIR=$PROJ_DIR/front
SERVER_DIR=$PROJ_DIR/server
PROJECT=bbg

tmux new-session -d -s $PROJECT
tmux attach-session $PROJECT

# Tmux config
tmux set -g mode-mouse on
tmux set -g mouse-select-pane on
tmux set -g mouse-resize-pane on
tmux set -g mouse-select-window on

# CopyPaste hotkeys for linux
tmux bind-key -n -t emacs-copy M-w copy-pipe "xclip -i -sel p -f | xclip -i -sel c "
tmux bind-key -n C-y run "xclip -o | tmux load-buffer - ; tmux paste-buffer"

tmux split-window -d -t 0 -v
tmux split-window -d -t 1 -h
tmux split-window -d -t 0 -h

# zmq pair
# tmux send-keys -t 1 'export ' PYTHONPATH=$WORKSPACE_DIR enter
# tmux send-keys -t 1 $ENV_DIR'/bin/python '$SERVER_DIR'/queue_server.py' enter

# client
tmux send-keys -t 3 'cd '$CLIENT_DIR enter
tmux send-keys -t 3 'npm start' enter

# websocket
tmux send-keys -t 2 'cd '$SERVER_DIR enter
tmux send-keys -t 2 'fresh' enter

tmux select-pane -t 0

tmux attach
