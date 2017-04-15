#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source $DIR/vars.sh

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

# django morda
tmux send-keys -t 3 'workon '$ENV enter
tmux send-keys -t 3 'cd '$DJANGO_APP_DIR enter
tmux send-keys -t 3 './manage.py runserver' enter

# game ws
tmux send-keys -t 2 'cd '$GAME_APP_DIR enter
tmux send-keys -t 2 'fresh -c fresh.conf' enter

tmux select-pane -t 0

tmux attach
