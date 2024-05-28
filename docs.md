---
name: Sonatype Nexus
description: Plugin to publish artifacts to Sonatype Nexus
author: Konstantin Demin
tags: [publish, Sonatype, Nexus]
containerImage: rockdrilla/woodpecker-sonatype-nexus
containerImageUrl: https://hub.docker.com/r/rockdrilla/woodpecker-sonatype-nexus
url: https://github.com/rockdrilla/woodpecker-sonatype-nexus
icon: https://www.sonatype.com/hubfs/2-2023-Product%20Logos/Repo%20Nav%20Icon%20updated.png
---

Woodpecker CI plugin to publish artifacts to Sonatype Nexus.

## Settings

| Name          | Required | Default value | Description                                                             |
|---------------|----------|---------------|-------------------------------------------------------------------------|
| `url`         | **yes**  | *none*        | Sonatype Nexus URL (e.g. `https://nexus.domain.com`)                    |
| `auth`        |  *no* \* | *none*        | HTTP Basic Authentication (plain-text, in form `{username}:{password}`) |
| `auth.base64` |  *no* \* | *none*        | HTTP Basic Authentication (base64-encoded)                              |
| `auth.header` |  *no* \* | *none*        | generic HTTP authentication header (in form `{Header}={Value}`)         |
| `upload`      | **yes**  | `[]`          | List of upload rules (JSON array, see below)                            |

**Notes:**

- At least one authentication setting **must** be provided.

  If there are more than one setting were specified then setting is selected in order of priority (from most to least):

  - `auth.header`
  - `auth.base64`
  - `auth`

- Setting names above are "short" variants.

  Full-qualified setting name looks like "`nexus.{short_name}`"
  and has higher priority if short variant is specified too.

- Dots in setting names are NOT mandatory.

  The one may replace dots ("`.`") with hyphens ("`-`") or underscores ("`_`").

### Upload settings

`upload` list consists of elements with following properties:

| Name          | Required | Default value | Description                                                                       |
|---------------|----------|---------------|-----------------------------------------------------------------------------------|
| `repository`  | **yes**  | *none*        | Repository name (of type "hosted")                                                |
| `paths`       | **yes**  | *none*        | List of files to upload (accepts [globs](https://pkg.go.dev/path/filepath#Match)) |

Additional (repository-specific) properties may be specified right with settings specified above.

## Example

```yaml
steps:
- name: publish
  image: rockdrilla/woodpecker-sonatype-nexus
  settings:
    url: https://nexus.domain.com
    auth.base64:
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

## Notes

- Preferred setting for HTTP Basic Authentication is `auth.base64` as there is minimal chance for breaking value during serialization/deserialization.

- Generic setting `auth.header` is provided for cases where authentication differs from HTTP Basic Authentication.

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

  Also, there is [fallback upload spec](https://github.com/rockdrilla/woodpecker-sonatype-nexus/blob/main/nexus/upload_spec/fallback.go):

  - if component/asset field does not specify `Optional: true` then this field is **required**.

## Known limitations

- No more than 32 assets may be uploaded at once (if destination repository type supports multiple upload).

  This is (merely) artificial limit for **single** upload - plugin will upload all listed files but via several calls.

  If you suppose that Sonatype Nexus is viable to receive more assets at once - feel free to contact me.
