# caddy-lambda

<a href="https://github.com/greenpau/caddy-lambda/actions/" target="_blank"><img src="https://github.com/greenpau/caddy-lambda/workflows/build/badge.svg?branch=main"></a>
<a href="https://pkg.go.dev/github.com/greenpau/caddy-lambda" target="_blank"><img src="https://img.shields.io/badge/godoc-reference-blue.svg"></a>
<a href="https://caddy.community" target="_blank"><img src="https://img.shields.io/badge/community-forum-ff69b4.svg"></a>
<a href="https://caddyserver.com/docs/modules/http.handlers.lambda" target="_blank"><img src="https://img.shields.io/badge/caddydocs-trace-green.svg"></a>

Event-Based Function Execution (Lambda) Plugin for [Caddy v2](https://github.com/caddyserver/caddy).

<!-- begin-markdown-toc -->
## Table of Contents

* [Overview](#overview)
* [Getting Started](#getting-started)

<!-- end-markdown-toc -->

## Overview

The `caddy-lambda` triggers execution of a function when it is invoked. It is a terminal
plugin, i.e. the plugin writes response headers and body.

## Getting Started

```
localhost {
	route /api/* {
		lambda {
			name hello_world
			runtime python
			python_executable {$HOME}/path/to/venv/bin/python
			entrypoint assets/scripts/api/hello_world/app/index.py
			function handler
		}
	}
	route {
		respond "OK"
	}
}
```