VERSION --shell-out-anywhere --use-chmod --use-host-command --earthly-version-arg --use-copy-link 0.6

IMPORT github.com/defn/cloud/lib:master AS lib

FROM lib+platform

warm:

build:
    FROM +warm
    COPY src src
    #RUN ~/bin/e poetry build
    SAVE ARTIFACT dist/* AS LOCAL dist/

publish:
    FROM +warm
    COPY dist dist
    #RUN --push --secret POETRY_PYPI_TOKEN_PYPI ~/bin/e poetry publish

get:
    FROM +warm
    COPY cdktf.json .
    RUN ~/bin/e cdktf get
    RUN --no-cache find .gen -ls
    SAVE ARTIFACT .gen/boundary/* AS LOCAL provider/defn_cdktf_provider_boundary/
    SAVE ARTIFACT .gen/vault/* AS LOCAL provider/defn_cdktf_provider_vault/
    SAVE ARTIFACT .gen/cloudflare/* AS LOCAL provider/defn_cdktf_provider_cloudflare/
    SAVE ARTIFACT .gen/buildkite/* AS LOCAL provider/defn_cdktf_provider_buildkite/
