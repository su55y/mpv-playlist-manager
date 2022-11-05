#!/bin/sh

SCRIPTPATH="$(cd -- "$(dirname "$0")" >/dev/null 2>&1 || exit 1 ; pwd -P)"

[ -f "$SCRIPTPATH/helper.sh" ] || {
    notify-send -i "rofi" -a "mpv control" "helper script not found"
    exit 1
}

theme() { cat <<EOF
configuration {
  font: "NotoSans Nerd Font 18";
}
inputbar {
  children: ["textbox-prompt-colon","entry","num-filtered-rows","textbox-num-sep","num-rows","case-indicator"];
}
textbox-prompt-colon {
  str: "";
  text-color: #502158;
}
element {
  border:  0;
  padding: 2px;
  children: ["element-icon","element-text"];
}
element-icon {
  size: 36px;
  border: 0px;
  padding: 0 5px;
}
/* hide element after clear */
element.selected.urgent {
  background-color: #00000000;
}
EOF
}

rofi -i -show "mpvctl" \
    -modi "mpvctl:$SCRIPTPATH/helper.sh" \
    -no-config \
    -no-custom \
    -theme-str "$(theme)"
