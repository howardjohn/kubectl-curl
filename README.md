# kubectl curl

`kubectl curl` is a plugin to make a curl call to the API server.

Why not `kubectl --raw`? It only supports some verbs.

## Install

`go install github.com/howardjohn/kubectl-curl@latest`

## Usage

Example usage:

```
$ kubectl curl /api
```

Note the URL MUST be first argument.