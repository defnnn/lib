include('/home/ubuntu/Tiltfile')

load("ext://uibutton", "cmd_button", "location")

local_resource(
    "python",
    serve_cmd="""
        cd; cd work/cloud;
        (python -mvenv .v);
        . .v/bin/activate;
        p export src::;
        code --install-extension ms-python.python || true;
        code --install-extension bungcip.better-toml || true;
        p --loop fmt lint check package ::;
    """,
    allow_parallel=True,
    labels=["automation"],
)

cmd_button(
    name="make login",
    text="make login",
    icon_name="login",
    argv=[
        "bash",
        "-c",
        """
            cd; cd work/cloud;
            make login;
        """
    ],
    location="nav",
)

cmd_button(
    name="make config-test",
    text="make config-test",
    icon_name="login",
    argv=[
        "bash",
        "-c",
        """
            cd; cd work/cloud;
            make config-test;
        """
    ],
    location="nav",
)
