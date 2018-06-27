#!/usr/bin/dumb-init /bin/sh

# check if a old fluent user exists and delete it
cat /etc/passwd | grep fluent
if [ $? -eq 0 ]; then
    deluser fluent
fi

exec "$@"
