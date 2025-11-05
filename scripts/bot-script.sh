#!/usr/bin/env bash
set -Eeuo pipefail

###############################################################################
# Settings
###############################################################################
SESSION=pokerbot_client

# client
BRCLIENT_DIR=$HOME/projects/bisonrelay/brclient
CFG=/home/pongbot/.brclient/brclient.conf
BRSERVER_PORT=12345
BR_RPC_PORT=7676

# bot
BOT_DIR=$HOME/projects/BR/pongbisonrelay/cmd/pongbot

###############################################################################
# (Re)start session
###############################################################################
tmux kill-session -t "$SESSION" 2>/dev/null || true
tmux new-session  -d -s "$SESSION" -n brclient \
  "until timeout 1 bash -c \"</dev/tcp/localhost/$BRSERVER_PORT\" 2>/dev/null; do
       echo 'waiting for brserver on :$BRSERVER_PORT'; sleep 3
   done
   cd \"$BRCLIENT_DIR\"
   go build -o brclient
   ./brclient --cfg \"$CFG\""

###############################################################################
# Window 1: interactive shell, then start the bot with send-keys
###############################################################################
tmux new-window -t "$SESSION":1 -n bot "$SHELL"

# build-and-run block – written as one multiline string to keep it readable
tmux send-keys -t "$SESSION":1 "
until timeout 1 bash -c \"</dev/tcp/localhost/$BR_RPC_PORT\" 2>/dev/null; do
    echo 'waiting for WS on :$BR_RPC_PORT'; sleep 3
done
cd \"$BOT_DIR\"
go build -o pongbot && ./pongbot
" C-m

###############################################################################
# Attach on the brclient pane
###############################################################################
tmux select-window -t "$SESSION":0
tmux attach-session -t "$SESSION"
