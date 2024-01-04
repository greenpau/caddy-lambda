# caddy-lambda

<a href="https://github.com/greenpau/caddy-lambda/actions/" target="_blank"><img src="https://github.com/greenpau/caddy-lambda/actions/workflows/build.yml/badge.svg"></a>
<a href="https://pkg.go.dev/github.com/greenpau/caddy-lambda" target="_blank"><img src="https://img.shields.io/badge/godoc-reference-blue.svg"></a>
<a href="https://caddy.community" target="_blank"><img src="https://img.shields.io/badge/community-forum-ff69b4.svg"></a>
<a href="https://caddyserver.com/docs/modules/http.handlers.lambda" target="_blank"><img src="https://img.shields.io/badge/caddydocs-lambda-green.svg"></a>

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

The `Caddyfile` config follows:

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

The `assets/scripts/api/hello_world/app/index.py` follows:

```py
import json

def handler(event: dict) -> dict:
    print(f"event: {event}")
    response = {
        "body": json.dumps({"message": "hello world!"}),
        "status_code": 200,
    }
    return response
```

The `response` dictionary is mandatory for a handler. he `status_code` and `body` are
mandatory fields of the `response`. The plugin writes `status_code` and `body` back to
the requestor.