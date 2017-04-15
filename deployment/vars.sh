#!/bin/bash
PROJECT=bbg

WORKSPACE_DIR=$GOPATH/src/github.com/DeV1doR
PROJ_DIR=$WORKSPACE_DIR/bbg

DJANGO_APP_DIR=$PROJ_DIR/bbg_client
GAME_APP_DIR=$PROJ_DIR/bbg_server
ENV_DIR=~/.virtualenvs/$PROJECT
ENV=$PROJECT