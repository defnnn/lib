analytics_settings(False)
allow_k8s_contexts("k3d-control")

load("ext://uibutton", "cmd_button", "location")
load("ext://restart_process", "custom_build_with_restart")

default_registry("169.254.32.1:5000")

local_resource("vite",
    serve_cmd=[
        "bash", "-c",
        """
            pnpm install
            exec turbo dev
        """
    ],
    deps=[".vite-mode"]
)

local_resource("temporal",
    serve_cmd=[
        "bash", "-c",
        """
            pkill -9 temporalit[e] || true
            rm -f ~/.config/temporalite/db/default.db
            exec temporalite start --namespace default --ip 0.0.0.0
        """
    ]
)

cmd_button(
    name="client",
    text="Client",
    icon_name="login",
    argv=[
        "bash", "-c",
        """
            cd dist/infra/app && ./bin queue
        """,
    ],
    location=location.NAV,
)

# TODO when infra resource is updated, run infra-test
local_resource("infra-test",
    deps=[
            "cmd/%s/main.cue" % ("infra",),
        ],
    cmd=[
        "bash", "-c",
        """
            cd cmd/%s && ../../dist/%s/app/bin queue
        """ % ("infra","infra")
    ]
)

for app in ("api", "infra"):
    local_resource("%s-go" % (app,),
        "mkdir -p dist/%s/app; cp cmd/%s/*.cue dist/%s/app/; mkdir -p dist/%s/app && go build -o dist/%s/app/bin cmd/%s/%s.go; echo done" % (app,app,app,app,app,app,app),
        deps=[
            "cmd/%s/%s.go" % (app,app),
            "cmd/%s/schema/" % (app,)
        ])

    k8s_yaml("cmd/%s/%s.yaml" % (app,app))

    custom_build_with_restart(
        ref=app,
        command=(
            "c nix-docker-build %s .#go ${EXPECTED_REF}" % (app,)
        ),
        entrypoint="/app/bin",
        deps=[
            "dist/%s/app/bin" % (app,),
        ],
        live_update=[
            sync("dist/%s/app/bin" % (app,), "/app/"),
        ],
    )

    if app in ("infrax",):
        local_resource("%s-tf" % (app,),
            deps=[
                "dist/%s/app/bin" % (app,),
                "dist/%s/app/main.cue" % (app,),
            ],
            cmd=[
                "bash", "-c",
                """
                    set -exfu
                    export CDKTF_CONTEXT_JSON="$(jq -n '{excludeStackIdFromLogicalIds: "true", allowSepCharsInLogicalIds: "true"}')"
                    (cd dist/%s/app && rm -rf cdktf.out && echo ./bin)
                    mkdir -p cmd/%s/tf
                    (set +f; rsync -ia --no-times --checksum dist/%s/app/cdktf.out/stacks/. cmd/%s/tf/.)
                    set +x
                    for a in {1..10}; do echo; done
                    git diff cmd/%s/tf || true
                    echo done
                """ % (app,app,app,app,app,)
            ]
        )
        local_resource("%s-plan" % (app,),
            deps=[
                "cmd/%s/tf/workspaces/cdk.tf.json" % (app,)
            ],
            cmd=[
                "bash", "-c",
                """
                    set -exfu
                    (cd cmd/%s/tf/workspaces && make plan)
                """ % (app,)
            ]
        )
