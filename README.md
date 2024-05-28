# plugin-sonatype-nexus

[![build-status](https://ci.krd.sh/api/badges/5/status.svg)](https://ci.krd.sh/repos/5)
[![goreport](https://goreportcard.com/badge/git.krd.sh/krd/woodpecker-sonatype-nexus)](https://goreportcard.com/report/git.krd.sh/krd/woodpecker-sonatype-nexus)
[![docker-pulls](https://img.shields.io/docker/pulls/rockdrilla/woodpecker-sonatype-nexus)](https://hub.docker.com/r/rockdrilla/woodpecker-sonatype-nexus)
[![license](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

Woodpecker CI plugin/standalone executable to publish artifacts to Sonatype Nexus.

Example `.woodpecker.yml`:

```yaml
steps:
- name: publish
  image: rockdrilla/woodpecker-sonatype-nexus
  settings:
    url: https://nexus.domain.com
    auth_base64:
      # consult with #3406 for that syntax
      # ref: https://github.com/woodpecker-ci/woodpecker/pull/3406
      from_secret: nexus-auth-b64
    upload:
      - repository: project-apt
        paths:
          - dist/all/*.deb
          - dist/amd64/*.deb
      - repository: project-raw
        paths:
          - dist/raw/all-in-one.tar.xz
        # property from upload specification for "raw" repository
        directory: /build/
      - repository: project-r
        paths:
          - dist/r/*.tar.gz
        # property from upload specification for "r" repository
        pathId: /src/contrib/
```

Example `.gitlab-ci.yml`:

```yaml
publish R:
  stage: publish
  image: rockdrilla/woodpecker-sonatype-nexus
  variables:
    NEXUS_URL: https://nexus.domain.com
   #NEXUS_AUTH_BASE64 is stored as CI variable
    NEXUS_REPOSITORY: project-r
    NEXUS_PATHS: "dist/r/*.tar.gz"
    NEXUS_PROPERTIES: "pathId=/src/contrib/"
```

Example manual invocation (within `rockdrilla/woodpecker-sonatype-nexus` container):

```sh
# publish R
publish-nexus \
  --nexus.url          https://nexus.domain.com \
  --nexus.auth        'upload-user:super-$ecret-passw0rd' \
  --nexus.repository   project-r \
  --nexus.paths       'dist/r/*.tar.gz' \
  --nexus.properties  'pathId=/src/contrib/'
```

## Woodpecker CI plugin

Plugin documentation is provided in [separate document](./docs.md).

## Other CI systems / standalone executable

### Environment

| Environment variable | Required | Description                                                                                       |
|----------------------|----------|---------------------------------------------------------------------------------------------------|
| `NEXUS_URL`          | **yes**  | Sonatype Nexus URL (e.g. `https://nexus.domain.com`)                                              |
| `NEXUS_AUTH`         |  *no* \* | HTTP Basic Authentication (plain-text, in form `{username}:{password}`)                           |
| `NEXUS_AUTH_BASE64`  |  *no* \* | HTTP Basic Authentication (base64-encoded)                                                        |
| `NEXUS_AUTH_HEADER`  |  *no* \* | generic HTTP authentication header (in form `{Header}={Value}`)                                   |
| `NEXUS_REPOSITORY`   | **yes**  | Repository name (of type "hosted")                                                                |
| `NEXUS_PATHS`        | **yes**  | Comma-separated list of files to upload (accepts [globs](https://pkg.go.dev/path/filepath#Match)) |
| `NEXUS_PROPERTIES`   |  *no*    | Comma-separated list of additional repository-specific properties (in form `{key}={value}`)       |

### Command-line flags

| Flag                  | Required | Multiple times? | Description                                                                       |
|-----------------------|----------|-----------------|-----------------------------------------------------------------------------------|
| `--nexus.url`         | **yes**  |   *no*          | Sonatype Nexus URL (e.g. `https://nexus.domain.com`)                              |
| `--nexus.auth`        |  *no* \* |   *no*          | HTTP Basic Authentication (plain-text, in form `{username}:{password}`)           |
| `--nexus.auth.base64` |  *no* \* |   *no*          | HTTP Basic Authentication (base64-encoded)                                        |
| `--nexus.auth.header` |  *no* \* |   *no*          | generic HTTP authentication header (in form `{Header}={Value}`)                   |
| `--nexus.repository`  | **yes**  |   *no*          | Repository name (of type "hosted")                                                |
| `--nexus.paths`       | **yes**  | **yes**         | List of files to upload (accepts [globs](https://pkg.go.dev/path/filepath#Match)) |
| `--nexus.properties`  |  *no*    | **yes**         | Additional repository-specific properties (in form `{key}={value}`)               |

## Notes

- At least one authentication setting **must** be provided.

  If there are more than one setting were specified then setting is selected in order of priority (from most to least):

  - `NEXUS_AUTH_HEADER`
  - `NEXUS_AUTH_BASE64`
  - `NEXUS_AUTH`

- Preferred setting for HTTP Basic Authentication is `NEXUS_AUTH_BASE64` as there is minimal chance for breaking value during serialization/deserialization.

- Generic authentication setting `NEXUS_AUTH_HEADER` is provided for cases where authentication differs from HTTP Basic Authentication.

- The one may use [User Tokens](https://help.sonatype.com/en/user-tokens.html) for HTTP Basic Authentication.

  There is no need for special handling as tokens are conform to scheme:

  `{token name code}:{token pass code}`

- The one may consult with Sonatype Nexus REST API for repository-specific properties for component uploads.

  Sonatype Nexus REST API is available via:

  - Web UI  - `https://nexus.domain.com/#admin/system/api`
  - Swagger - `https://nexus.domain.com/service/rest/swagger.json`

  Points of interest are:

  - `/v1/formats/upload-specs`
  - `/v1/components` (with `POST` method)

  Also, there is [fallback upload spec](./nexus/upload_spec/fallback.go):

  - if component/asset field does not specify `Optional: true` then this field is **required**.

## Known limitations

- No more than 32 assets may be uploaded at once (if destination repository type supports multiple upload).

  This is (merely) artificial limit for **single** upload - plugin will upload all listed files but via several calls.

  If you suppose that Sonatype Nexus is viable to receive more assets at once - feel free to contact me.
