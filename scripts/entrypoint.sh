#!/bin/sh

export KEY_SEED=$(cat $KEY_SEED_FILE)

exec "$@"
