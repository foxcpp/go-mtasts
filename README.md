go-mtasts
===========

[![Reference](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/foxcpp/go-mtasts)
[![stability-unstable](https://img.shields.io/badge/stability-unstable-yellow.svg?style=flat-square)](https://github.com/emersion/stability-badges#unstable)

MTA-STS policy processing library for Go. Extracted from [maddy] code base.

See example_test.go for usage example.

Related standards
-------------------

- SMTP MTA Strict Transport Security (MTA-STS)
  [RFC 8461](https://tools.ietf.org/html/rfc8461)

Notes
-------

- Absence of direct "download policy" and similar methods is intentional.
  Caching is critical for MTA-STS security.

[maddy]: https://github.com/foxcpp/maddy/go-mtasts
