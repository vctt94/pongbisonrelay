#!/usr/bin/env bash
set -Eeuo pipefail

###############################################################################
# Settings
###############################################################################
SESSION=pongclient_session1           # tmux session name

# bison-relay client
BRCLIENT_DIR=$HOME/projects/bisonrelay/brclient
CFG=$HOME/brclientdirs/dir1/brclient.conf
BRSERVER_PORT=12345                    # relays TCP port
BR_RPC_PORT=7777                       # client’s WS RPC port

# pong client
PONGCLIENT_DIR=$HOME/projects/BR/pongbisonrelay/cmd/pongclient
PONG_DATADIR=$HOME/.pongbot

###############################################################################
# Restart session if it already exists
###############################################################################
tmux kill-session -t "$SESSION" 2>/dev/null || true

###############################################################################
# Window 0 – brclient
###############################################################################
tmux new-session -d -s "$SESSION" -n brclient "
until nc -z localhost $BRSERVER_PORT; do
    echo 'waiting for brserver on :$BRSERVER_PORT'; sleep 3
done
cd \"$BRCLIENT_DIR\"
go build -o brclient
./brclient --cfg \"$CFG\"
"

###############################################################################
# Window 1 – pong client (interactive shell, pane stays open)
###############################################################################
tmux new-window  -t "$SESSION":1 -n pongclient "$SHELL"

tmux send-keys  -t "$SESSION":1 "
until nc -z localhost $BR_RPC_PORT; do
    echo 'waiting for WS on :$BR_RPC_PORT'; sleep 3
done
cd \"$PONGCLIENT_DIR\"
echo '--- pong client running (Ctrl-C to stop, ↑ to rerun) ---'
go build -o pongclient && ./pongclient
" C-m

###############################################################################
# Start attached on window 0 (Prefix-2 to jump to pong client)
###############################################################################
tmux select-window -t "$SESSION":0
tmux attach-session -t "$SESSION"
