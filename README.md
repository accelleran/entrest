
# entrest

> **Note**: This is a fork of [lrstanley/entrest](https://github.com/lrstanley/entrest) maintained by Accelleran with additional features and improvements.
>
> [View changes from upstream](https://github.com/lrstanley/entrest/compare/master...accelleran:entrest:master)

<p align="center">
  <a href="https://github.com/accelleran/entrest/tags">
    <img title="Latest Semver Tag" src="https://img.shields.io/github/v/tag/accelleran/entrest?style=flat-square">
  </a>
  <a href="https://github.com/accelleran/entrest/commits/master">
    <img title="Last commit" src="https://img.shields.io/github/last-commit/accelleran/entrest?style=flat-square">
  </a>
  <a href="https://codecov.io/gh/accelleran/entrest">
    <img title="Code Coverage" src="https://img.shields.io/codecov/c/github/accelleran/entrest/master?style=flat-square">
  </a>
  <a href="https://pkg.go.dev/github.com/accelleran/entrest">
    <img title="Go Documentation" src="https://pkg.go.dev/badge/github.com/accelleran/entrest?style=flat-square">
  </a>
  <a href="https://goreportcard.com/report/github.com/accelleran/entrest">
    <img title="Go Report Card" src="https://goreportcard.com/badge/github.com/accelleran/entrest?style=flat-square">
  </a>
</p>

## :link: Table of Contents

  - [Features](#sparkles-features)
  - [Usage](#gear-usage)
  - [License](#balance_scale-license)

## :sparkles: Features

**entrest** is an [EntGo](https://entgo.io/) extension for generating compliant OpenAPI
specs and an HTTP handler implementation that matches that spec. It expands upon the
approach used by [entoas](https://github.com/ent/contrib/tree/master/entoas#entoas),
with additional functionality, and pairs the generated specification with a
fully-functional HTTP handler implementation.

- :sparkles: Generates OpenAPI specs for your EntGo schema.
- :sparkles: Generates a fully functional HTTP handler implementation that matches the OpenAPI spec.
- :sparkles: Supports automatic pagination (where applicable).
- :sparkles: Supports advanced filtering (using query parameters, `AND`/`OR` predicates, etc).
- :sparkles: Supports eager-loading edges, so you don't have to make additional calls unnecessarily.
- :sparkles: Supports various forms of sorting.
- :sparkles: And more!

---

## :gear: Usage

For general documentation, refer to the [upstream documentation](https://lrstanley.github.io/entrest/).

```console
go get -u github.com/accelleran/entrest@latest
```

---

## :balance_scale: License

MIT License

Copyright (c) 2024 Liam Stanley <liam@liam.sh>

This fork maintained by Accelleran includes modifications and additional features while preserving the original MIT license.

See [LICENSE](LICENSE) for full license text.
