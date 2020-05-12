# hoverdnsapi -- Hover DNS API

Hover DNS API is a tool to query Hover's DNS API.  I use it to pull out my domains in JSON and see if I've made a consistency mistake.

Also, when you move house, and your domains are tagged to your home address (even if hidden from Whois), Hover offers no API for bulk changes, so you need to make manual changes.  For each contact address (Admin, Tech, Billing, Owner).  For each domain.  They mostly work, but it's easy to miss one.

With a long-term goal somewhere like how Rancid used to commit changes to a SCM, I'm drawing from [Dan Krause's Example](https://gist.github.com/dankrause/5585907) to see how to read and parse DNS from Hover.


# How to use

It's Go (Golang if you need more letters).  Just import and roll forward.

I should apologize, I'm new to Go, so my Go might be of poor-quality.  Please PR me improvements if you see me doing a really silly thing.


# How to build

    $ git clone http://github.com/chickenandpork/hoverdnsapi hoverdnsapi
    $ cd hoverdnsapi && go test ./...

By itself, it's just the parsing structures which are the common code behind what I'm doing.

This builds with Go-1.10, I hope it won't break going forward.

## Automatic builds

```bash
go get github.com/githubnemo/CompileDaemon
cd hoverdnsapi && CompileDaemon \
	-build="go build ./cmd/hoverdns" -exclude-dir=.git -exclude=".*.swp" \
	-command="./hoverdns -a -m allanc@smallfoot.org --domains smallfoot.org info"


# Why?

I didn't find one, so I had to build it.  I hope the next person can leverage this to cruise on at peak efficiency and invent truly useful things.


# License

MIT.  Use as you like.  Everywhere.

A thanks would be cool, or kudos on https://www.linkedin.com/in/goldfish, but it's totally OK if you're too busy fighting truly criminal coding errors to feed my curiosity.

