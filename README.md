# oci-image-tool [![Build Status](https://travis-ci.org/opencontainers/image-tools.svg?branch=master)](https://travis-ci.org/opencontainers/image-tools)[![Go Report Card](https://goreportcard.com/badge/github.com/opencontainers/image-tools)](https://goreportcard.com/report/github.com/opencontainers/image-tools)

`oci-image-tool` is a collection of tools for working with the [OCI image format specification](https://github.com/opencontainers/image-spec).
To build from source code, image-tools requires Go 1.7.x or above.

## Project Status

**This project is no longer actively maintained, and has had major deficiencies
for several years. This project may be archived in the future, depending on
community feedback.**

The `oci-image-tools` repository has two classes of tools, and our
recommendation is different based on which tools you need:

* Conformance testing tools (`oci-image-tool validate`). Currently there is no
  replacement for this functionality in a different OCI project, so users
  should continue to use this tool -- but should be aware that it is unlikely
  to be updated to future versions of the OCI Image Specification.

* Image manipulation tools (`oci-image-tool create` and `oci-image-tool
  validate`). These tools have had many long-standing correctness and
  functionality issues (see [1][p1], [2][p2], [3][p3], [4][p4], [5][p5],
  [6][p6], [7][p7], [8][p8], [9][p9], [10][p10]), and we have not recommended
  their use for a fairly long time. Users should instead use [umoci][umoci] for
  these tasks, which is a production-grade reference implementation of the OCI
  Image Specification that is actively maintained and does not suffer from the
  same deficiencies as this project.

[umoci]: https://github.com/opencontainers/umoci
[p1]: https://github.com/opencontainers/image-tools/issues/17
[p2]: https://github.com/opencontainers/image-tools/pull/8
[p3]: https://github.com/opencontainers/image-tools/issues/59
[p4]: https://github.com/opencontainers/image-tools/issues/74
[p5]: https://github.com/opencontainers/image-tools/issues/76
[p6]: https://github.com/opencontainers/image-tools/issues/193
[p7]: https://github.com/opencontainers/image-tools/issues/218
[p8]: https://github.com/opencontainers/image-tools/issues/219
[p9]: https://github.com/opencontainers/image-tools/issues/223
[p10]: https://github.com/opencontainers/image-tools/issues/188


## Install

It is recommended that use `go get` to download a single command tools.

```
$ go get -d github.com/opencontainers/image-tools/cmd/oci-image-tool
$ cd $GOPATH/src/github.com/opencontainers/image-tools/
$ make all
$ sudo make install
```

## Uninstall

```
$ sudo make uninstall
```

## Example

### Obtaining an image

The following examples assume you have a [image-layout](https://github.com/opencontainers/image-spec/blob/v1.0.0-rc2/image-layout.md) tar archive at `busybox-oci`.
One way to acquire that image is with [skopeo](https://github.com/projectatomic/skopeo#installing):

```
$ skopeo copy docker://busybox oci:busybox-oci:latest
```

### oci-image-tool-create

More information about `oci-image-tool-create` can be found in its [man page](./man/oci-image-tool-create.1.md)

```
$ mkdir busybox-bundle
$ oci-image-tool create --ref latest busybox-oci busybox-bundle
$ cd busybox-bundle && sudo runc run busybox
```

### oci-image-tool-validate

More information about `oci-image-tool-validate` can be found in its [man page](./man/oci-image-tool-validate.1.md)

```
$ oci-image-tool validate --type imageLayout --ref latest busybox-oci
busybox-oci: OK
```

### oci-image-tool-unpack

More information about `oci-image-tool-unpack` can be found in its [man page](./man/oci-image-tool-unpack.1.md)

```
$ mkdir busybox-bundle
$ oci-image-tool unpack --ref latest busybox-oci busybox-bundle
$ tree busybox-bundle
busybox-bundle
├── bin
│   ├── [
│   ├── [[
│   ├── acpid
│   ├── addgroup
│   ├── add-shell
[...]
```

# Contributing

Development happens on GitHub. Issues are used for bugs and actionable items and longer discussions can happen on the [mailing list](#mailing-list).
The code is licensed under the Apache 2.0 license found in the `LICENSE` file of this repository.

## Code of Conduct

Participation in the OpenContainers community is governed by [OpenContainer's Code of Conduct][code-of-conduct].

## Security

If you find an issue, please follow the [security][security] protocol to report it.

## Discuss your design

The project welcomes submissions, but please let everyone know what you are working on.

Before undertaking a nontrivial change to this repository, send mail to the [mailing list](#mailing-list) to discuss what you plan to do.
This gives everyone a chance to validate the design, helps prevent duplication of effort, and ensures that the idea fits.
It also guarantees that the design is sound before code is written; a GitHub pull-request is not the place for high-level discussions.

Typos and grammatical errors can go straight to a pull-request.
When in doubt, start on the [mailing-list](#mailing-list).

## Weekly Call

The contributors and maintainers of all OCI projects have a weekly meeting Wednesdays at 2:00 PM (USA Pacific.)
Everyone is welcome to participate via [UberConference web][UberConference] or audio-only: 888-587-9088 or 860-706-8529 (no PIN needed.)
An initial agenda will be posted to the [mailing list](#mailing-list) earlier in the week, and everyone is welcome to propose additional topics or suggest other agenda alterations there.
Minutes are posted to the [mailing list](#mailing-list) and minutes from past calls are archived to the [wiki](https://github.com/opencontainers/runtime-spec/wiki) for those who are unable to join the call.

## Mailing List

You can subscribe and join the mailing list on [Google Groups](https://groups.google.com/a/opencontainers.org/forum/#!forum/dev).

## IRC

OCI discussion happens on #opencontainers on Freenode ([logs][irc-logs]).

## Git commit

### Sign your work

The sign-off is a simple line at the end of the explanation for the patch, which certifies that you wrote it or otherwise have the right to pass it on as an open-source patch.
The rules are pretty simple: if you can certify the below (from [developercertificate.org](http://developercertificate.org/)):

```
Developer Certificate of Origin
Version 1.1

Copyright (C) 2004, 2006 The Linux Foundation and its contributors.
660 York Street, Suite 102,
San Francisco, CA 94110 USA

Everyone is permitted to copy and distribute verbatim copies of this
license document, but changing it is not allowed.


Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
```

then you just add a line to every git commit message:

    Signed-off-by: Joe Smith <joe@gmail.com>

using your real name (sorry, no pseudonyms or anonymous contributions.)

You can add the sign off when creating the git commit via `git commit -s`.

### Commit Style

Simple house-keeping for clean git history.
Read more on [How to Write a Git Commit Message](http://chris.beams.io/posts/git-commit/) or the Discussion section of [`git-commit(1)`](http://git-scm.com/docs/git-commit).

1. Separate the subject from body with a blank line
2. Limit the subject line to 50 characters
3. Capitalize the subject line
4. Do not end the subject line with a period
5. Use the imperative mood in the subject line
6. Wrap the body at 72 characters
7. Use the body to explain what and why vs. how
  * If there was important/useful/essential conversation or information, copy or include a reference
8. When possible, one keyword to scope the change in the subject (i.e. "README: ...", "runtime: ...")


[security]: https://github.com/opencontainers/org/blob/master/security
[code-of-conduct]: https://github.com/opencontainers/org/blob/master/CODE_OF_CONDUCT.md
[UberConference]: https://www.uberconference.com/opencontainers
[irc-logs]: http://ircbot.wl.linuxfoundation.org/eavesdrop/%23opencontainers/
