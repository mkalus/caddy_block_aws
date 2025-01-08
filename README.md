# Caddy Block AWS Cloud

Automatically block all calls from AWS servers in your Caddy server.

I created this [Caddy Server](https://caddyserver.com/) module to automatically get rid of bot calls from the AWS cloud.
Unfortunately, it looks like AWS is used by many crawler bots to crawl your site. In case of
[Memorial Archives](https://memorial-archives.international/), I wanted to get rid of those calls, because they are
simply unwanted.

The module loads the official AWS ip range file from https://ip-ranges.amazonaws.com/ip-ranges.json and parses it. AWS
contains over 8000 ip ranges, so an efficient ip matching is required. I use
[Ryo Namiki's ipfilter](https://github.com/paralleltree/ipfilter) for this, since it implements an efficient binary tree
search.

## TODOs

There are still some todos to implement/check:

* [ ] Periodic update of the data: Right now, the AWS ip list only loaded once. It should be updated once in a while.
* [x] Caching? Check if it is faster to cache ips once they are checked in the binary tree (especially on misses).
* [ ] Configurable settings (for cache, omission of ipv4/ipv6)

## Requirements

* Go
* xcaddy: `go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest`


## Usage

The module implements a simple directive `blockaws` to be included in your caddy file. Since this is a third party
directive, you have to tell Caddy when to add the directive using the global `order` setting.

Here is a simple example (also found in [Caddyfile](./Caddyfile)):

```
{
	order blockaws after header
	auto_https off
}

http://localhost:2015 {
	blockaws

	respond "Hello, world!"
}
```

Compile your custom Caddy server using:

```shell
xcaddy build --with github.com/mkalus/caddy_block_aws
```

And run it:

```shell
./caddy run
```
You should see a log entry http.handlers.blockaws  Loaded AWS IP ranges` - this shows that the directive has been loaded
correctly.

## Compiling from Source

You can clone this repository and run:

```bash
xcaddy run
```

You should see a log entry http.handlers.blockaws  Loaded AWS IP ranges` - this shows that the directive has been loaded
correctly.

You can test with:

```bash
curl -v localhost:2015
```

If you try this from an AWS server, your request *should* be blocked. Building from a local repo is done via:

```bash
xcaddy build --with github.com/mkalus/caddy_block_aws=.
```
