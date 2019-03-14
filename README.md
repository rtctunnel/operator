# RTCTunnel Operator
The RTCTunnel Operator helps peer using [RTCTunnel](https://github.com/rtctunnel/rtctunnel) connect.

## Overview

The Operator is an HTTP server that exposes `/pub` and `/sub` endpoints for sending messages.

## Installation

Either use `go install` or you can also use the provided docker image at [quay.io/rtctunnel/operator](https://quay.io/rtctunnel/operator):

```bash
docker pull quay.io/rtctunnel/operator:v1.0.0
```

