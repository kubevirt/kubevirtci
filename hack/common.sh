# Ensure proper terminal line wrapping with SSH
function _ssh_into_node() {
    if [[ $2 != "" ]]; then
        ${CRI_BIN} exec "$@"
    else
        ${CRI_BIN} exec -e COLUMNS="$COLUMNS" -e LINES="$LINES" -it "$1" bash
    fi
}
