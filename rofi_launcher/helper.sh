#!/bin/sh

# activate hotkeys
printf "\000use-hot-keys\037true\n"

request() {
    curl "http://localhost:5000/rofi$1" --user-agent 'rofi' 2>/dev/null
}

case $ROFI_RETV in
    # print list on start
    0) request "";;
    # select line
    1)
        request "?index=$ROFI_INFO"
    ;;
esac
