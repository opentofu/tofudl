# Specification for TofuDL mirrors

Version: pending

## Abstract

This document describes the minimum requirements to host a mirror that is compatible with TofuDL. We also intend this document to be a guide for other OpenTofu downloading tools. 

## The API endpoint

The API endpoint is a JSON file, by default hosted at https://get.opentofu.org/tofu/api.json, containing all versions of OpenTofu and a list of all artifacts uploaded for each version. An example file would look as follows:

```json
{
  "versions": [
    {
      "id": "1.8.0",
      "files": [
        "file1.ext",
        "file2.ext",
        "..."
      ]
    }
  ]
}
```

A JSON schema document is included [in this repository](https://github.com/opentofu/tofudl/blob/main/api.schema.json). When using Go, you may want to use the `github.com/opentofu/tofudl.APIResponse` struct.

Note that versions are included *without* the `v` prefix in the version listing and *may* contain the suffixes `-alphaX`, `-betaX`, `-rcX`. The response **must** sort version in reverse order according to semantic versioning. The filenames in the response should *not* include a path.

Mirror implementations *may* restrict access to the API endpoint by means of the `Authorization` HTTP header and *should* use encrypted connections (`https://`).

## Download mirror

The files contained in the API endpoint response lead to a download mirror for the artifacts. The implementation *may* choose a URL structure freely and the URL *may* contain the version number. Client implementations *should* make the version URL configurable/templatable. The default URL template for the download mirror, expressed as a Go template, is: `https://github.com/opentofu/opentofu/releases/download/v{{ .Version }}/{{ .Artifact }}`

Mirror implementations *may* use HTTP redirects and *may* include authentication requirements by means of the `Authorization` HTTP header and *should* use encrypted connections (`https://`). 

### Downloading OpenTofu

The artifact containing OpenTofu is named `tofu_{{ .Version }}_{{ .Platform }}_{{ .Architecture }}.tar.gz`. While the official mirror contains more files, TofuDL implementations *should* not rely on other files being present to limit the scope of necessary mirroring. The archive will contain a file called `tofu`/`tofu.exe`, along with supplemental files, such as the license file.

The platform may contain the following values:

- `windows`
- `linux`
- `darwin` (MacOS)
- `freebsd`
- `openbsd`
- `solaris`

The architecture may contain the following values:

- `386`
- `amd64`
- `arm`
- `arm64`

Note: not all platform/architecture combinations lead to valid artifacts. Also note that future versions of OpenTofu may introduce more platforms or architectures.

### Verifying artifacts

Verifying the integrity of mirrored files should be performed for every download to ensure that no malicious binaries have been introduced to the mirror. The verification should be performed in the following two steps:

1. Verify the SHa256 checksum of the downloaded artifact against the file called `tofu_{{ .Version }}_SHA256SUMS`. This file has lines in the following format. The entries are separated by two spaces and end with a newline (`\n`), but no carriage return (`\r`). Implementations SHOULD nevertheless strip extra whitespace characters and disregard empty newlines.
   ```
   HEX-ENCODED-SHA256-CHECKSUM  FILENAME
   ```
   For example, consider the following line:
   ```
   f6f7c0a8cefde6e750d0755fb2de9f69272698fa6fd8cfe02f15d19911beef84  tofu_1.8.0_linux_arm.tar.gz
   ```
2. Once the checksum is verified, the `SHA256SUMS` file should be verified using a GPG key against the `tofu_{{ .Version }}_SHA256SUMS.gpgsig`. This file contains a non-armored OpenPGP/GnuPG signature with a corresponding signing key. The signing key defaults to the one found at https://get.opentofu.org/opentofu.asc, fingerprint `E3E6E43D84CB852EADB0051D0C0AF313E5FD9F80`. Implementations *should not* attempt to use a GnuPG keyserver to obtain this key and *should* allow for configurable signing keys for self-built binaries.

Note: TofuDL currently doesn't use Cosign/SigStore signatures because they are not available as lightweight libraries yet.
