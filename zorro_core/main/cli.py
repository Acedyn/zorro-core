import argparse

from db import db_manager

HANDLERS_MAPPING = {
    "db": {}
    # "action": handlers.action_handler,
    # "command": handlers.command_handler,
    # "launch": handlers.launch_handler,
}


def main():
    """
    Execute actions and commands on a resolved context
    """

    # Root parser: global parameter used to define the context and others
    parser = argparse.ArgumentParser(description=main.__doc__)
    parser.add_argument(
        "-e",
        "--entities",
        help="list of the entity IDs to select",
        default=[],
        nargs="*",
    )
    parser.add_argument(
        "--db-url",
        "-d",
        help="The url of the sqlite database",
        type=str,
        required=False,
    )

    subparsers = parser.add_subparsers(
        help="The action to perform under the given context", dest="subcommand"
    )

    # DB parser: manage the database
    db_parser = subparsers.add_parser(
        "db",
        help="Manage the local database",
    )
    db_parser.add_argument(
        "operation",
        help="The operation to perform on the database",
        choices=["migrate", "reset", "create-admin"],
    )

    # Tool parser: global parameters to actions, commands, events...
    tool_parser = argparse.ArgumentParser(add_help=False)
    tool_parser.add_argument(
        "-lp",
        "--list-parameters",
        help="Print the available parameters",
        default=False,
        action="store_true",
    )
    tool_parser.add_argument(
        "-p",
        "--parameter",
        help="Set the parameters value with <parameter> = <value>",
        action="append",
        default=[],
    )

    # Action parser: execute an action
    action_parser = subparsers.add_parser(
        "action",
        help="Execute the given action in the selected context",
        parents=[tool_parser],
    )
    action_parser.add_argument(
        "action_name",
        help="The name of the action to perform under the context",
    )

    # Command parser: execute a command
    command_parser = subparsers.add_parser(
        "command",
        help="Execute the given command in the selected context",
        parents=[tool_parser],
    )
    command_parser.add_argument(
        "command_name",
        help="The name of the command to execute under the context",
    )

    # # Widget parser: popup a widget
    # widget_parser = subparsers.add_parser(
    #     "widget",
    #     help="Pop up a widget in the selected context",
    # )

    # # Event parser: listen for events
    # event_parser = subparsers.add_parser(
    #     "listen",
    #     help="Listen to events in the selected context",
    # )

    # Launch parser: launch a client
    launcher_parser = subparsers.add_parser(
        "launch",
        help="Launch the given program as a client in the context",
        parents=[tool_parser],
    )
    launcher_parser.add_argument(
        "program",
        help="The name of the program to start",
        type=str,
    )
    launcher_parser.add_argument(
        "--file",
        "-f",
        help="The file to open within the client",
        type=str,
        required=False,
    )

    args = vars(parser.parse_args())
    subcommand = args.pop("subcommand", None)
    print(subcommand, args)


if __name__ == "__main__":
    main()
