COMMAND=$1
shift

if [[ "${COMMAND}" != "run" && "${COMMAND}" != "provision" && "${COMMAND}" != "ssh" && "${COMMAND}" != "rm" ]]; then
        echo "No valid command provides. Valid commands  are 'run', 'provision', 'ssh' and 'rm'."
        exit 1
fi

