analytics_settings(False)
allow_k8s_contexts("pod")

load("ext://uibutton", "cmd_button", "location")
load("ext://restart_process", "custom_build_with_restart")

default_registry("169.254.32.1:5000")

for app in ["defn", "defm"]:
    local_resource("pants-go-%s" % (app,), "p package cmd/%s::" % (app), deps=["cmd/%s" % (app,)])

    k8s_yaml("cmd/%s/%s.yaml" % (app,app))

    custom_build_with_restart(
        ref=app,
        command=(
            "earthly --push --remote-cache=${EXPECTED_REGISTRY}/${EXPECTED_IMAGE}-cache +%s --image=${EXPECTED_REF}" % (app,)
        ),
        entrypoint="/app/bin",
        deps=["dist/cmd.%s" % (app,)],
        live_update=[
            sync("dist/cmd.%s/bin" % (app,), "/app/bin"),
        ],
    )
